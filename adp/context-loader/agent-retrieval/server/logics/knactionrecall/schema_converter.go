// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knactionrecall

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const (
	// MaxSchemaDepth 最大 $ref 引用递归深度，用于防止过深嵌套和循环引用导致的无限递归
	// 作用：
	// 1. 限制非循环的深层嵌套引用（如 A -> B -> C -> D）
	// 2. 作为循环引用的第二道防线（循环引用检测优先触发）
	// 建议值：2-3 层
	// - 2 层：适合简单场景（如树形结构）
	// - 3 层：适合复杂场景（多层嵌套）
	MaxSchemaDepth = 3
)

// ==================== 通用 Schema 引用解析器 ====================
// NOTE: 统一 OpenAPI (#/components/schemas/) 和 MCP (#/$defs/) 的 $ref 解析逻辑
// 两者的核心递归逻辑、循环检测、深度控制、剪枝策略完全一致，
// 仅引用路径格式和查找位置不同，通过 RefResolver 函数参数化实现复用。

// RefResolver 引用解析器函数类型
// 作用：根据 $ref 路径查找并返回被引用的 Schema 定义
// OpenAPI: 从 apiSpec["components"]["schemas"] 查找
// MCP: 从 inputSchema["$defs"] 查找
type RefResolver func(refPath string) (map[string]any, error)

// resolveSchemaWithResolver 通用 Schema 解析函数（支持 $ref 引用、循环检测和深度控制）
// 参数：
//   - ctx: 上下文
//   - schema: 待解析的 Schema
//   - refResolver: 引用解析器（根据 $ref 路径查找定义）
//   - visitedRefs: 已访问的引用路径（用于循环检测）
//   - currentDepth: 当前递归深度
//
// 返回：解析后的 Schema（$ref 已内联）
func (s *knActionRecallServiceImpl) resolveSchemaWithResolver(
	ctx context.Context,
	schema map[string]any,
	refResolver RefResolver,
	visitedRefs map[string]bool,
	currentDepth int,
) (map[string]any, error) {
	if schema == nil {
		return map[string]any{"type": "string"}, nil
	}

	// 复制 schema 以避免修改原 map
	resolved := make(map[string]any)
	for k, v := range schema {
		resolved[k] = v
	}

	// 1. 处理 $ref 引用
	if refPath, ok := resolved["$ref"].(string); ok {
		// 1.1 循环引用检测
		if visitedRefs[refPath] {
			s.logger.WithContext(ctx).Debugf("[SchemaResolver] Circular reference detected for %s, pruning", refPath)
			refSchema, err := refResolver(refPath)
			if err != nil {
				return map[string]any{"type": "object"}, nil
			}
			return s.pruneSchema(refSchema), nil
		}

		// 1.2 深度限制检测
		if currentDepth >= MaxSchemaDepth {
			s.logger.WithContext(ctx).Debugf("[SchemaResolver] Max depth reached for %s (depth: %d), pruning", refPath, currentDepth)
			refSchema, err := refResolver(refPath)
			if err != nil {
				return map[string]any{"type": "object"}, nil
			}
			return s.pruneSchema(refSchema), nil
		}

		// 1.3 标记为已访问
		visitedRefs[refPath] = true
		defer func() { delete(visitedRefs, refPath) }()

		// 1.4 获取引用定义并递归解析（深度 +1）
		refSchema, err := refResolver(refPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve $ref %s: %w", refPath, err)
		}
		return s.resolveSchemaWithResolver(ctx, refSchema, refResolver, visitedRefs, currentDepth+1)
	}

	// 2. 处理 properties（递归解析每个属性，深度不变）
	if props, ok := resolved["properties"].(map[string]any); ok {
		newProps := make(map[string]any)
		for propName, propDef := range props {
			if propMap, ok := propDef.(map[string]any); ok {
				resolvedProp, err := s.resolveSchemaWithResolver(ctx, propMap, refResolver, visitedRefs, currentDepth)
				if err != nil {
					s.logger.WithContext(ctx).Warnf("[SchemaResolver] Failed to resolve property %s: %v", propName, err)
					newProps[propName] = propDef // 降级：保留原值
				} else {
					newProps[propName] = resolvedProp
				}
			} else {
				newProps[propName] = propDef
			}
		}
		resolved["properties"] = newProps
	}

	// 3. 处理 array items（递归解析，深度不变）
	if resolved["type"] == "array" {
		if items, ok := resolved["items"].(map[string]any); ok {
			resolvedItems, err := s.resolveSchemaWithResolver(ctx, items, refResolver, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[SchemaResolver] Failed to resolve array items: %v", err)
			} else {
				resolved["items"] = resolvedItems
			}
		}
	}

	return resolved, nil
}

// convertSchemaToFunctionCall 将 OpenAPI Schema 转换为 OpenAI Function Call Schema
// 改进：保持分层结构（header/path/query/body），而不是扁平化
//
//nolint:unparam // 保持接口一致性，error 返回用于后续扩展
func (s *knActionRecallServiceImpl) convertSchemaToFunctionCall(ctx context.Context, apiSpec map[string]any) (map[string]any, error) {
	// 使用分层结构：header/path/query/body
	properties := map[string]any{
		"header": map[string]any{
			"type":        "object",
			"description": "HTTP Header 参数",
			"properties":  make(map[string]any),
		},
		"path": map[string]any{
			"type":        "object",
			"description": "URL Path 参数",
			"properties":  make(map[string]any),
		},
		"query": map[string]any{
			"type":        "object",
			"description": "URL Query 参数",
			"properties":  make(map[string]any),
		},
		"body": map[string]any{
			"type":        "object",
			"description": "Request Body 参数",
			"properties":  make(map[string]any),
		},
	}

	// 各位置的必填参数
	requiredByLocation := map[string][]string{
		"header": {},
		"path":   {},
		"query":  {},
		"body":   {},
	}

	// 用于循环引用检测的访问记录
	visitedRefs := make(map[string]bool)

	// 1. 处理 parameters (path/query/header)
	if params, ok := apiSpec["parameters"].([]any); ok {
		for _, paramItem := range params {
			param, ok := paramItem.(map[string]any)
			if !ok {
				continue
			}

			paramName, _ := param["name"].(string)
			if paramName == "" {
				continue
			}

			paramLocation, _ := param["in"].(string) // path/query/header
			if paramLocation == "" {
				continue
			}

			// 解析参数 schema（支持 $ref，支持深度控制）
			paramSchema, err := s.resolveSchema(ctx, param["schema"], apiSpec, visitedRefs, 0)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertSchema] Failed to resolve param schema for %s: %v", paramName, err)
				continue
			}

			// 构建参数定义
			propDef := s.buildPropertyDefinition(paramSchema, param["description"])

			// 根据位置放入对应的 properties
			if locationProps, ok := properties[paramLocation].(map[string]any); ok {
				if props, ok := locationProps["properties"].(map[string]any); ok {
					props[paramName] = propDef
				}
			}

			// 收集必填参数
			if isRequired, ok := param["required"].(bool); ok && isRequired {
				requiredByLocation[paramLocation] = append(requiredByLocation[paramLocation], paramName)
			}
		}
	}

	// 2. 处理 request_body (body 参数)
	if requestBody, ok := apiSpec["request_body"].(map[string]any); ok {
		if content, ok := requestBody["content"].(map[string]any); ok {
			if appJSON, ok := content["application/json"].(map[string]any); ok {
				if schema, ok := appJSON["schema"].(map[string]any); ok {
					// 解析 body schema（支持 $ref，支持深度控制）
					bodySchema, err := s.resolveSchema(ctx, schema, apiSpec, visitedRefs, 0)
					if err != nil {
						s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertSchema] Failed to resolve body schema: %v", err)
						// 添加一个通用的 body 参数作为兜底
						if bodyProps, ok := properties["body"].(map[string]any); ok {
							if props, ok := bodyProps["properties"].(map[string]any); ok {
								props["request_body"] = map[string]any{
									"type":        "object",
									"description": "请求体参数",
								}
							}
						}
					} else {
						// 展开 body schema 的 properties
						if bodyProps, ok := properties["body"].(map[string]any); ok {
							if props, ok := bodyProps["properties"].(map[string]any); ok {
								s.mergeSchemaProperties(ctx, props, bodySchema, apiSpec, visitedRefs, 0)
							}
							// 合并必填项
							if bodyRequired, ok := bodySchema["required"].([]any); ok {
								for _, req := range bodyRequired {
									if reqStr, ok := req.(string); ok {
										requiredByLocation["body"] = append(requiredByLocation["body"], reqStr)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 3. 设置各位置的 required 字段
	for location, required := range requiredByLocation {
		if len(required) > 0 {
			if locationProps, ok := properties[location].(map[string]any); ok {
				locationProps["required"] = required
			}
		}
	}

	// 4. 清理空的 location（如果没有参数，移除该位置）
	result := map[string]any{
		"type":       "object",
		"properties": make(map[string]any),
	}

	resultProps := result["properties"].(map[string]any)
	for location, locationProps := range properties {
		if props, ok := locationProps.(map[string]any)["properties"].(map[string]any); ok {
			if len(props) > 0 {
				resultProps[location] = locationProps
			}
		}
	}

	// 如果没有任何参数，至少返回一个空结构
	if len(resultProps) == 0 {
		resultProps["body"] = map[string]any{
			"type":        "object",
			"description": "Request Body 参数",
			"properties":  make(map[string]any),
		}
	}

	return result, nil
}

// resolveSchema 解析 schema，支持 $ref 引用、循环引用检测和深度控制
// 采用深度限制剪枝策略：
// - 每次解析 $ref 时，深度 +1
// - 解析 properties 中的属性时，深度不变（同一层级）
// - 达到最大深度时，执行剪枝（保留类型和原始描述，移除 properties）
// currentDepth: 当前递归深度，用于控制循环引用的展开深度
func (s *knActionRecallServiceImpl) resolveSchema(
	ctx context.Context,
	schema any,
	apiSpec map[string]any,
	visitedRefs map[string]bool,
	currentDepth int,
) (map[string]any, error) {
	if schema == nil {
		return map[string]any{"type": "string"}, nil
	}

	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return map[string]any{"type": "string"}, nil
	}

	// 如果直接有 type 且没有 $ref，直接返回
	if _, hasType := schemaMap["type"]; hasType && schemaMap["$ref"] == nil {
		// 如果有 properties，需要递归处理（深度不变）
		if props, ok := schemaMap["properties"].(map[string]any); ok {
			resolvedProps := make(map[string]any)
			for propName, propDef := range props {
				resolvedProp, err := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, currentDepth)
				if err != nil {
					s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve property %s: %v", propName, err)
					continue
				}
				resolvedProps[propName] = resolvedProp
			}
			schemaMap["properties"] = resolvedProps
		}
		// 处理 array.items（深度不变）
		if schemaMap["type"] == "array" {
			if items, ok := schemaMap["items"].(map[string]any); ok {
				resolvedItems, err := s.resolveSchema(ctx, items, apiSpec, visitedRefs, currentDepth)
				if err != nil {
					s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve array items: %v", err)
				} else {
					schemaMap["items"] = resolvedItems
				}
			}
		}
		return schemaMap, nil
	}

	// 处理 $ref 引用
	if refPath, ok := schemaMap["$ref"].(string); ok {
		// 检查循环引用（必须在深度检查之前，避免无限递归）
		if visitedRefs[refPath] {
			// 检测到循环引用，执行剪枝
			s.logger.WithContext(ctx).Debugf("[KnActionRecall#resolveSchema] Circular reference detected for %s (depth: %d), pruning", refPath, currentDepth)
			// 获取被引用的 schema 基本信息，然后剪枝
			referencedSchema, err := s.getReferencedSchema(refPath, apiSpec)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to get referenced schema for pruning: %v", err)
				return map[string]any{"type": "object"}, nil
			}
			return s.pruneSchema(referencedSchema), nil
		}

		// 检查是否达到最大深度
		if currentDepth >= MaxSchemaDepth {
			s.logger.WithContext(ctx).Debugf("[KnActionRecall#resolveSchema] Max depth reached for %s (depth: %d), pruning", refPath, currentDepth)
			// 获取被引用的 schema 基本信息，然后剪枝
			referencedSchema, err := s.getReferencedSchema(refPath, apiSpec)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to get referenced schema for pruning: %v", err)
				return map[string]any{"type": "object"}, nil
			}
			return s.pruneSchema(referencedSchema), nil
		}

		// 标记为已访问
		wasVisited := visitedRefs[refPath]
		visitedRefs[refPath] = true
		defer func() {
			// 递归返回时，如果这是第一次访问，清理标记
			if !wasVisited {
				delete(visitedRefs, refPath)
			}
		}()

		// 解析 $ref 路径（深度 +1）
		resolvedSchema, err := s.resolveDollarRef(ctx, refPath, apiSpec, visitedRefs, currentDepth+1)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve $ref %s: %w", refPath, err)
		}

		return resolvedSchema, nil
	}

	// 如果有 properties，递归处理（深度不变，同一层级）
	if props, ok := schemaMap["properties"].(map[string]any); ok {
		resolvedProps := make(map[string]any)
		for propName, propDef := range props {
			resolvedProp, err := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve property %s: %v", propName, err)
				continue
			}
			resolvedProps[propName] = resolvedProp
		}
		schemaMap["properties"] = resolvedProps
	}

	// 处理 array.items（深度不变）
	if schemaMap["type"] == "array" {
		if items, ok := schemaMap["items"].(map[string]any); ok {
			resolvedItems, err := s.resolveSchema(ctx, items, apiSpec, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#resolveSchema] Failed to resolve array items: %v", err)
			} else {
				schemaMap["items"] = resolvedItems
			}
		}
	}

	return schemaMap, nil
}

// getReferencedSchema 获取被引用的 schema 定义（不解析，只获取基本信息）
func (s *knActionRecallServiceImpl) getReferencedSchema(refPath string, apiSpec map[string]any) (map[string]any, error) {
	// 解析 $ref 路径格式：#/components/schemas/SchemaName
	if !strings.HasPrefix(refPath, "#/components/schemas/") {
		return nil, fmt.Errorf("unsupported $ref path format: %s (only #/components/schemas/* is supported)", refPath)
	}

	schemaName := strings.TrimPrefix(refPath, "#/components/schemas/")
	if schemaName == "" {
		return nil, fmt.Errorf("empty schema name in $ref: %s", refPath)
	}

	// 从 components.schemas 中查找
	components, ok := apiSpec["components"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("components not found in api_spec")
	}

	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("components.schemas not found in api_spec")
	}

	schema, ok := schemas[schemaName].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema %s not found in components.schemas", schemaName)
	}

	return schema, nil
}

// pruneSchema 剪枝函数：当达到最大深度时，保留类型和原始描述，移除 properties
// 核心策略：不添加循环引用说明，节省 token
func (s *knActionRecallServiceImpl) pruneSchema(schema map[string]any) map[string]any {
	result := make(map[string]any)

	// 保留类型信息
	if schemaType, ok := schema["type"].(string); ok && schemaType != "" {
		result["type"] = schemaType
	} else {
		result["type"] = "object" // 默认类型
	}

	// 保留原始 description（如果存在，不修改，不添加循环引用说明）
	if desc, ok := schema["description"].(string); ok && desc != "" {
		result["description"] = desc
	}

	// 如果是 array，保留 items 结构但不展开 properties
	if result["type"] == "array" {
		if items, ok := schema["items"].(map[string]any); ok {
			// 递归剪枝 items
			result["items"] = s.pruneSchema(items)
		}
	}

	// 不包含 properties（避免继续递归）
	// 不添加循环引用说明（节省 token）

	return result
}

// resolveDollarRef 解析 $ref 引用（完整实现，支持循环引用检测和深度控制）
func (s *knActionRecallServiceImpl) resolveDollarRef(
	ctx context.Context,
	refPath string,
	apiSpec map[string]any,
	visitedRefs map[string]bool,
	currentDepth int,
) (map[string]any, error) {
	// 获取被引用的 schema
	schema, err := s.getReferencedSchema(refPath, apiSpec)
	if err != nil {
		return nil, err
	}

	// 递归解析（可能包含嵌套的 $ref，传递深度信息）
	return s.resolveSchema(ctx, schema, apiSpec, visitedRefs, currentDepth)
}

// buildPropertyDefinition 构建属性定义
func (s *knActionRecallServiceImpl) buildPropertyDefinition(schema map[string]any, description any) map[string]any {
	propDef := make(map[string]any)

	// 类型
	if propType, ok := schema["type"].(string); ok && propType != "" {
		propDef["type"] = propType
	} else {
		propDef["type"] = "string" // 默认类型
	}

	// 描述（优先使用参数级别的 description，其次使用 schema 中的 description）
	if desc, ok := description.(string); ok && desc != "" {
		propDef["description"] = desc
	} else if desc, ok := schema["description"].(string); ok && desc != "" {
		propDef["description"] = desc
	}

	// 枚举
	if enum, ok := schema["enum"].([]any); ok {
		propDef["enum"] = enum
	}

	// 如果 schema 有 properties，保留嵌套结构
	if props, ok := schema["properties"].(map[string]any); ok {
		propDef["properties"] = props
		propDef["type"] = "object"
	}

	// 如果 schema 是 array，保留 items 结构
	if schema["type"] == "array" {
		if items, ok := schema["items"].(map[string]any); ok {
			propDef["items"] = items
		}
	}

	return propDef
}

// mergeSchemaProperties 合并 schema 的 properties 到目标 properties
func (s *knActionRecallServiceImpl) mergeSchemaProperties(
	ctx context.Context,
	targetProps, schema, apiSpec map[string]any,
	visitedRefs map[string]bool,
	currentDepth int,
) {
	if props, ok := schema["properties"].(map[string]any); ok {
		for propName, propDef := range props {
			resolvedProp, err := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, currentDepth)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#mergeSchemaProperties] Failed to resolve property %s: %v", propName, err)
				continue
			}
			targetProps[propName] = s.buildPropertyDefinition(resolvedProp, nil)
		}
	}
}

// mapFixedParams 映射固定参数到 header/path/query/body
func (s *knActionRecallServiceImpl) mapFixedParams(
	_ context.Context,
	parameters, apiSpec map[string]any,
) interfaces.KnFixedParams {
	fixedParams := interfaces.KnFixedParams{
		Header: make(map[string]any),
		Path:   make(map[string]any),
		Query:  make(map[string]any),
		Body:   make(map[string]any),
	}

	// 建立参数名到位置的映射表
	paramLocationMap := make(map[string]string)
	if params, ok := apiSpec["parameters"].([]any); ok {
		for _, paramItem := range params {
			if param, ok := paramItem.(map[string]any); ok {
				if name, ok := param["name"].(string); ok {
					if in, ok := param["in"].(string); ok {
						paramLocationMap[name] = in
					}
				}
			}
		}
	}

	// 根据映射表分类参数
	for key, value := range parameters {
		location := paramLocationMap[key]
		switch location {
		case "header":
			fixedParams.Header[key] = value
		case "path":
			fixedParams.Path[key] = value
		case "query":
			fixedParams.Query[key] = value
		case "body":
			fixedParams.Body[key] = value
		default:
			// 未找到映射，使用命名规则判断
			if isHeaderParam(key) {
				fixedParams.Header[key] = value
			} else {
				// 默认放入 body
				fixedParams.Body[key] = value
			}
		}
	}

	return fixedParams
}

// isHeaderParam 判断是否为 header 参数（基于命名规则）
func isHeaderParam(key string) bool {
	// 常见的 header 参数名称模式
	headerPatterns := []string{
		"x-", "X-",
		"authorization", "Authorization",
		"content-type", "Content-Type",
	}

	for _, pattern := range headerPatterns {
		if len(key) >= len(pattern) && key[:len(pattern)] == pattern {
			return true
		}
	}

	return false
}

// convertMCPSchemaToFunctionCall 将 MCP JSON Schema 转换为 OpenAI Function Call Schema
// NOTE: 使用通用的 resolveSchemaWithResolver 函数，通过 RefResolver 参数化 $defs 查找逻辑
func (s *knActionRecallServiceImpl) convertMCPSchemaToFunctionCall(ctx context.Context, inputSchema map[string]any) (map[string]any, error) {
	// OpenAI function call schema 期望根节点有 type=object 和 properties
	// MCP schema 通常已经是 JSON Schema，但可能包含 $defs
	// 我们需要解析 $defs，并确保根结构符合 OpenAI 要求

	visitedRefs := make(map[string]bool)

	// 提取 rootDefs ($defs)
	rootDefs := make(map[string]any)
	if defs, ok := inputSchema["$defs"].(map[string]any); ok {
		rootDefs = defs
	}

	// 构建 MCP 专用的引用解析器
	// MCP 引用格式: #/$defs/SchemaName
	mcpRefResolver := func(refPath string) (map[string]any, error) {
		prefix := "#/$defs/"
		if !strings.HasPrefix(refPath, prefix) {
			return nil, fmt.Errorf("unsupported MCP $ref path format: %s (only #/$defs/* is supported)", refPath)
		}
		name := strings.TrimPrefix(refPath, prefix)
		if def, ok := rootDefs[name].(map[string]any); ok {
			return def, nil
		}
		return nil, fmt.Errorf("MCP schema definition not found: %s", name)
	}

	// 使用通用解析器解析 schema
	resolvedSchema, err := s.resolveSchemaWithResolver(ctx, inputSchema, mcpRefResolver, visitedRefs, 0)
	if err != nil {
		return nil, err
	}

	// NOTE: 为第一层 body 参数添加默认描述
	// 当 body 参数存在但缺少 description 时（例如通过 $ref 引用的 schema 没有 description），
	// 自动添加默认描述 "Request Body参数"
	if props, ok := resolvedSchema["properties"].(map[string]any); ok {
		if bodyProp, ok := props["body"].(map[string]any); ok {
			if _, hasDesc := bodyProp["description"]; !hasDesc {
				bodyProp["description"] = "Request Body参数"
			}
		}
	}

	// 确保有 type=object (通常 MCP schema 根就是 object，但为了保险)
	if _, ok := resolvedSchema["type"]; !ok {
		resolvedSchema["type"] = "object"
	}

	// 移除 $defs (因为已经解析并内联了)
	delete(resolvedSchema, "$defs")

	return resolvedSchema, nil
}

// ==================== Action Driver Schema 转换方法 ====================

// convertToolSchemaToActionDriver 将 Tool OpenAPI Schema 去壳转换为行动驱动 dynamic_params
// 去除 header/path/query/body 外壳，将所有参数合并为扁平的 dynamic_params.properties
// 若同名字段来自不同 location，返回错误
func (s *knActionRecallServiceImpl) convertToolSchemaToActionDriver(ctx context.Context, apiSpec map[string]any) (map[string]any, error) {
	// 合并后的 dynamic_params properties 和 required
	dynamicProperties := make(map[string]any)
	dynamicRequired := []string{}

	// 记录参数名到 location 的映射，用于冲突检测
	paramLocationMap := make(map[string]string)

	// 用于循环引用检测的访问记录
	visitedRefs := make(map[string]bool)

	// 1. 处理 parameters (path/query/header)
	if params, ok := apiSpec["parameters"].([]any); ok {
		for _, paramItem := range params {
			param, ok := paramItem.(map[string]any)
			if !ok {
				continue
			}

			paramName, _ := param["name"].(string)
			if paramName == "" {
				continue
			}

			paramLocation, _ := param["in"].(string)
			if paramLocation == "" {
				continue
			}

			// 冲突检测：同名字段来自不同 location
			if existingLocation, exists := paramLocationMap[paramName]; exists {
				if existingLocation != paramLocation {
					errMsg := fmt.Sprintf("参数名 '%s' 在不同位置重复出现 (已有: %s, 当前: %s)，无法生成行动驱动动态工具",
						paramName, existingLocation, paramLocation)
					s.logger.WithContext(ctx).Errorf("[KnActionRecall#convertToolSchemaToActionDriver] %s", errMsg)
					return nil, fmt.Errorf("%s", errMsg)
				}
			}
			paramLocationMap[paramName] = paramLocation

			// 解析参数 schema
			paramSchema, err := s.resolveSchema(ctx, param["schema"], apiSpec, visitedRefs, 0)
			if err != nil {
				s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertToolSchemaToActionDriver] Failed to resolve param schema for %s: %v", paramName, err)
				continue
			}

			// 构建参数定义，直接放入 dynamic_params.properties
			propDef := s.buildPropertyDefinition(paramSchema, param["description"])
			dynamicProperties[paramName] = propDef

			// 收集必填参数
			if isRequired, ok := param["required"].(bool); ok && isRequired {
				dynamicRequired = append(dynamicRequired, paramName)
			}
		}
	}

	// 2. 处理 request_body (body 参数) — 去掉 body 外壳，展开到 dynamic_params
	if requestBody, ok := apiSpec["request_body"].(map[string]any); ok {
		if content, ok := requestBody["content"].(map[string]any); ok {
			if appJSON, ok := content["application/json"].(map[string]any); ok {
				if schema, ok := appJSON["schema"].(map[string]any); ok {
					bodySchema, err := s.resolveSchema(ctx, schema, apiSpec, visitedRefs, 0)
					if err != nil {
						s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertToolSchemaToActionDriver] Failed to resolve body schema: %v", err)
					} else {
						// 展开 body schema 的 properties 到 dynamic_params
						if bodyProps, ok := bodySchema["properties"].(map[string]any); ok {
							for propName, propDef := range bodyProps {
								// 冲突检测
								if existingLocation, exists := paramLocationMap[propName]; exists {
									errMsg := fmt.Sprintf("参数名 '%s' 在不同位置重复出现 (已有: %s, 当前: body)，无法生成行动驱动动态工具",
										propName, existingLocation)
									s.logger.WithContext(ctx).Errorf("[KnActionRecall#convertToolSchemaToActionDriver] %s", errMsg)
									return nil, fmt.Errorf("%s", errMsg)
								}
								paramLocationMap[propName] = "body"

								resolvedProp, resolveErr := s.resolveSchema(ctx, propDef, apiSpec, visitedRefs, 0)
								if resolveErr != nil {
									s.logger.WithContext(ctx).Warnf("[KnActionRecall#convertToolSchemaToActionDriver] Failed to resolve body property %s: %v", propName, resolveErr)
									continue
								}
								dynamicProperties[propName] = s.buildPropertyDefinition(resolvedProp, nil)
							}
						}

						// 合并 body required — 仅添加实际存在于 dynamicProperties 中的 key
						if bodyRequired, ok := bodySchema["required"].([]any); ok {
							for _, req := range bodyRequired {
								if reqStr, ok := req.(string); ok {
									if _, exists := dynamicProperties[reqStr]; exists {
										dynamicRequired = append(dynamicRequired, reqStr)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 3. 构造 dynamic_params schema
	dynamicParamsSchema := map[string]any{
		"type":        "object",
		"description": "行动执行动态参数",
		"properties":  dynamicProperties,
	}
	if len(dynamicRequired) > 0 {
		dynamicParamsSchema["required"] = dynamicRequired
	}

	// 4. 包装顶层行动驱动结构
	return s.wrapActionDriverParameters(dynamicParamsSchema), nil
}

// convertMCPSchemaToActionDriver 将 MCP Schema 转换为行动驱动请求结构
// 将解析后的 MCP input_schema 直接作为 dynamic_params 的 schema
func (s *knActionRecallServiceImpl) convertMCPSchemaToActionDriver(ctx context.Context, inputSchema map[string]any) (map[string]any, error) {
	visitedRefs := make(map[string]bool)

	// 提取 $defs
	rootDefs := make(map[string]any)
	if defs, ok := inputSchema["$defs"].(map[string]any); ok {
		rootDefs = defs
	}

	// 构建 MCP 专用的引用解析器
	mcpRefResolver := func(refPath string) (map[string]any, error) {
		prefix := "#/$defs/"
		if !strings.HasPrefix(refPath, prefix) {
			return nil, fmt.Errorf("unsupported MCP $ref path format: %s (only #/$defs/* is supported)", refPath)
		}
		name := strings.TrimPrefix(refPath, prefix)
		if def, ok := rootDefs[name].(map[string]any); ok {
			return def, nil
		}
		return nil, fmt.Errorf("MCP schema definition not found: %s", name)
	}

	// 使用通用解析器解析 schema
	resolvedSchema, err := s.resolveSchemaWithResolver(ctx, inputSchema, mcpRefResolver, visitedRefs, 0)
	if err != nil {
		return nil, err
	}

	// 确保有 type=object
	if _, ok := resolvedSchema["type"]; !ok {
		resolvedSchema["type"] = "object"
	}

	// 移除 $defs
	delete(resolvedSchema, "$defs")

	// 构造 dynamic_params schema：使用解析后的 MCP schema 作为 dynamic_params
	dynamicParamsSchema := map[string]any{
		"type":        "object",
		"description": "行动执行动态参数",
	}
	if props, ok := resolvedSchema["properties"].(map[string]any); ok {
		dynamicParamsSchema["properties"] = props
	} else {
		dynamicParamsSchema["properties"] = make(map[string]any)
	}
	if required, ok := resolvedSchema["required"]; ok {
		dynamicParamsSchema["required"] = required
	}

	// 包装顶层行动驱动结构
	return s.wrapActionDriverParameters(dynamicParamsSchema), nil
}

// wrapActionDriverParameters 统一包装顶层行动驱动请求参数结构
// 最外层固定为 dynamic_params + _instance_identities
func (s *knActionRecallServiceImpl) wrapActionDriverParameters(dynamicParamsSchema map[string]any) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dynamic_params": dynamicParamsSchema,
			"_instance_identities": map[string]any{
				"type":        "array",
				"description": "目标实例列表；为空时由行动驱动按条件扫描。如果需要指定实例，请根据上下文提取实例的动态属性作为键值对填入（例如 [{\"id\": \"123\"}, {\"name\": \"test_instance\"}]）",
				"items": map[string]any{
					"type": "object",
					"description": "实例标识对象，包含动态的属性键值对",
					"additionalProperties": map[string]any{},
				},
			},
		},
	}
}
