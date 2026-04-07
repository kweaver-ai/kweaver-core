package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Result 转换结果
type Result struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// SimpleConverter 处理简化OpenAPI格式的转换器
type SimpleConverter struct{}

// SimpleOpenAPI 简化OpenAPI格式
type SimpleOpenAPI struct {
	Parameters   []Parameter  `json:"parameters,omitempty"`
	RequestBody  *RequestBody `json:"request_body,omitempty"`
	Responses    []Response   `json:"responses,omitempty"`
	Components   *Components  `json:"components,omitempty"`
	Callbacks    any          `json:"callbacks,omitempty"`
	Security     any          `json:"security,omitempty"`
	Tags         any          `json:"tags,omitempty"`
	ExternalDocs any          `json:"external_docs,omitempty"`
}

// Parameter 参数定义
type Parameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required"`
	Schema      map[string]any `json:"schema"`
}

// RequestBody 请求体定义
type RequestBody struct {
	Description string             `json:"description,omitempty"`
	Content     map[string]Content `json:"content"`
	Required    bool               `json:"required"`
}

// Content 内容定义
type Content struct {
	Schema   map[string]any            `json:"schema"`
	Examples map[string]map[string]any `json:"examples,omitempty"`
}

// Response 响应定义
type Response struct {
	StatusCode  string             `json:"status_code"`
	Description string             `json:"description"`
	Content     map[string]Content `json:"content"`
}

// Components 组件定义
type Components struct {
	Schemas map[string]map[string]any `json:"schemas"`
}

// NewSimpleConverter 创建简化转换器实例
func NewSimpleConverter() *SimpleConverter {
	return &SimpleConverter{}
}

// ConvertFromBytes 从字节数组转换简化OpenAPI格式
func (c *SimpleConverter) ConvertFromBytes(data []byte) *Result {
	var simpleOpenAPI SimpleOpenAPI
	if err := json.Unmarshal(data, &simpleOpenAPI); err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("解析JSON失败: %v", err),
		}
	}

	return c.convertSimpleOpenAPI(&simpleOpenAPI)
}

// ConvertFromString 从字符串转换简化OpenAPI格式
func (c *SimpleConverter) ConvertFromString(jsonStr string) *Result {
	return c.ConvertFromBytes([]byte(jsonStr))
}

// convertSimpleOpenAPI 转换简化OpenAPI格式
func (c *SimpleConverter) convertSimpleOpenAPI(simple *SimpleOpenAPI) *Result {
	// 构建properties
	properties := map[string]any{}

	headers := c.extractParameters(simple.Parameters, "header", "HTTP 请求头参数")
	if headers != nil {
		properties["header"] = headers
	}

	query := c.extractParameters(simple.Parameters, "query", "URL 查询字符串参数")
	if query != nil {
		properties["query"] = query
	}

	path := c.extractParameters(simple.Parameters, "path", "URL 路径参数")
	if path != nil {
		properties["path"] = path
	}

	body := c.extractRequestBody(simple.RequestBody)
	if body != nil {
		properties["body"] = body
	}

	// 构建标准JSON Schema格式
	result := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	// 添加$defs（而不是components）
	if simple.Components != nil && len(simple.Components.Schemas) > 0 {
		// 转换components.schemas为$defs格式
		defs := make(map[string]any)
		for name, schema := range simple.Components.Schemas {
			// 递归处理schema中的$ref引用
			defs[name] = c.processSchemaRefs(schema)
		}
		result["$defs"] = defs
	}

	return &Result{
		Success: true,
		Data:    result,
	}
}

// extractParameters 提取指定类型的参数
// inType: 参数类型 (header/query/path)
// description: 该类型参数组的描述
func (c *SimpleConverter) extractParameters(params []Parameter, inType, description string) map[string]any {
	props := map[string]any{}
	required := []string{}

	for _, param := range params {
		if param.In == inType {
			schemaObj := c.convertSchema(param.Schema)
			if param.Description != "" {
				schemaObj["description"] = param.Description
			}
			props[param.Name] = schemaObj
			if param.Required {
				required = append(required, param.Name)
			}
		}
	}

	if len(props) == 0 {
		return nil
	}

	obj := map[string]any{
		"type":        "object",
		"description": description,
		"properties":  props,
	}

	if len(required) > 0 {
		obj["required"] = required
	}

	return obj
}

// extractRequestBody 提取请求体
func (c *SimpleConverter) extractRequestBody(reqBody *RequestBody) map[string]any {
	if reqBody == nil {
		return nil
	}

	// 获取JSON内容
	if content, ok := reqBody.Content["application/json"]; ok {
		schema := c.convertSchema(content.Schema)
		// 如果没有 description，补充默认描述
		if _, hasDesc := schema["description"]; !hasDesc {
			schema["description"] = "Request Body 参数"
		}
		return schema
	}

	return nil
}

// convertSchema 转换Schema
func (c *SimpleConverter) convertSchema(schema map[string]any) map[string]any {
	if schema == nil {
		return map[string]any{
			"type": "object",
		}
	}
	// 处理schema中的$ref引用
	return c.processSchemaRefs(schema)
}

// processSchemaRefs 递归处理schema中的$ref引用，将components/schemas/改为$defs/
func (c *SimpleConverter) processSchemaRefs(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	result := make(map[string]any)
	for key, value := range schema {
		switch v := value.(type) {
		case string:
			// 如果是$ref字段，转换引用路径
			if key == "$ref" {
				// 将 #/components/schemas/ 替换为 #/$defs/
				refStr := c.convertRefPath(v)
				result[key] = refStr
			} else {
				result[key] = v
			}
		case map[string]any:
			// 递归处理嵌套的map
			result[key] = c.processSchemaRefs(v)
		case []any:
			// 处理数组中的每个元素
			processedArray := make([]any, len(v))
			for i, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					processedArray[i] = c.processSchemaRefs(itemMap)
				} else {
					processedArray[i] = item
				}
			}
			result[key] = processedArray
		default:
			result[key] = v
		}
	}
	return result
}

// convertRefPath 转换引用路径
func (c *SimpleConverter) convertRefPath(refPath string) string {
	// 将 #/components/schemas/ 替换为 #/$defs/
	if refPath != "" && refPath[0] == '#' {
		// 替换 components/schemas 为 $defs
		refPath = strings.ReplaceAll(refPath, "/components/schemas/", "/$defs/")
	}
	return refPath
}

// ToJSONString 将转换结果转换为JSON字符串
func (c *SimpleConverter) ToJSONString(result *Result) (string, error) {
	if !result.Success {
		return "", fmt.Errorf("转换失败: %s", result.Error)
	}

	out, err := json.MarshalIndent(result.Data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	return string(out), nil
}

// GetSchemaInfo 获取Schema信息
func (c *SimpleConverter) GetSchemaInfo(result *Result) map[string]any {
	if !result.Success {
		return map[string]any{
			"error": result.Error,
		}
	}

	info := map[string]any{
		"type":        result.Data["type"],
		"description": result.Data["description"],
	}

	if properties, ok := result.Data["properties"].(map[string]any); ok {
		info["property_count"] = len(properties)
		info["properties"] = make([]string, 0, len(properties))
		for k := range properties {
			info["properties"] = append(info["properties"].([]string), k)
		}
	}

	if defs, ok := result.Data["$defs"].(map[string]any); ok {
		info["schema_count"] = len(defs)
		schemaList := make([]string, 0, len(defs))
		for k := range defs {
			schemaList = append(schemaList, k)
		}
		info["schemas"] = schemaList
	} else {
		// 如果没有$defs，设置默认值
		info["schema_count"] = 0
		info["schemas"] = []string{}
	}

	return info
}
