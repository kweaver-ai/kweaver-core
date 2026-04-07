package openapidoc

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	pkgerrors "github.com/pkg/errors"
)

// ValidateOpenAPI 使用 kin-openapi 对文档做最终结构校验，并放宽部分历史兼容字段。
func ValidateOpenAPI(ctx context.Context, doc *openapi3.T) error {
	if doc == nil {
		return pkgerrors.New("openapi document is nil")
	}

	return doc.Validate(
		ctx,
		openapi3.DisableExamplesValidation(),
		openapi3.AllowExtraSiblingFields(
			"content",
			"description",
			"summary",
			"properties",
			"required",
			"type",
			"items",
			"example",
			"default",
			"nullable",
			"enum",
			"format",
			"allOf",
			"oneOf",
			"anyOf",
		),
	)
}

// NormalizePathParameters 确保路径参数被声明为 required，并补齐路径中缺失的 path 参数定义。
func NormalizePathParameters(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}

	if doc.Components != nil {
		for _, parameterRef := range doc.Components.Parameters {
			normalizePathParameterRef(parameterRef)
		}
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Value(path)
		normalizeParameters(pathItem.Parameters)
		pathLevelParams := collectPathParameterNames(pathItem.Parameters)

		for _, operation := range pathItem.Operations() {
			normalizeParameters(operation.Parameters)
		}

		for _, name := range extractPathParameterNames(path) {
			if _, exists := pathLevelParams[name]; exists {
				continue
			}

			pathItem.Parameters = append(pathItem.Parameters, &openapi3.ParameterRef{
				Value: openapi3.NewPathParameter(name).WithSchema(openapi3.NewStringSchema()),
			})
			pathLevelParams[name] = struct{}{}
		}
	}
}

// NormalizeOperationIDs 为重复的 operationId 自动追加 method/path 后缀，避免冲突。
func NormalizeOperationIDs(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}

	counts := map[string]int{}

	for _, path := range doc.Paths.InMatchingOrder() {
		for _, pair := range orderedOperations(doc.Paths.Value(path)) {
			if pair.operation.OperationID == "" {
				continue
			}

			counts[pair.operation.OperationID]++
		}
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		for _, pair := range orderedOperations(doc.Paths.Value(path)) {
			if pair.operation.OperationID == "" || counts[pair.operation.OperationID] <= 1 {
				continue
			}

			pair.operation.OperationID = fmt.Sprintf(
				"%s_%s_%s",
				pair.operation.OperationID,
				strings.ToLower(pair.method),
				sanitizePathForOperationID(path),
			)
		}
	}
}

// NormalizeSecurity 统一历史 BearerAuth 命名，并清理与全局安全配置重复的接口级声明。
func NormalizeSecurity(doc *openapi3.T) {
	if doc == nil {
		return
	}

	if doc.Components == nil {
		components := openapi3.NewComponents()
		doc.Components = &components
	}

	if doc.Components.SecuritySchemes == nil {
		doc.Components.SecuritySchemes = openapi3.SecuritySchemes{}
	}

	if legacyScheme, ok := doc.Components.SecuritySchemes[bearerSecurityName]; ok {
		if _, exists := doc.Components.SecuritySchemes[apiKeySecurityName]; !exists {
			doc.Components.SecuritySchemes[apiKeySecurityName] = legacyScheme
		}

		delete(doc.Components.SecuritySchemes, bearerSecurityName)
	}

	doc.Security = normalizeSecurityRequirementsRef(doc.Security)

	if doc.Paths == nil {
		return
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Value(path)
		for _, pair := range orderedOperations(pathItem) {
			if pair.operation.Security == nil {
				continue
			}

			normalized := normalizeSecurityRequirementsRef(*pair.operation.Security)
			if securityRequirementsEqual(normalized, doc.Security) {
				pair.operation.Security = nil
				continue
			}

			pair.operation.Security = &normalized
		}
	}
}

// normalizeParameters 归一化参数列表中的每一个参数定义。
func normalizeParameters(parameters openapi3.Parameters) {
	for _, parameterRef := range parameters {
		normalizePathParameterRef(parameterRef)
	}
}

// normalizePathParameterRef 将 path 参数强制标记为 required，符合 OpenAPI 规范。
func normalizePathParameterRef(parameterRef *openapi3.ParameterRef) {
	if parameterRef == nil || parameterRef.Value == nil {
		return
	}

	if strings.EqualFold(parameterRef.Value.In, "path") {
		parameterRef.Value.Required = true
	}
}

// collectPathParameterNames 收集一组参数里已经声明过的 path 参数名。
func collectPathParameterNames(parameters openapi3.Parameters) map[string]struct{} {
	names := map[string]struct{}{}

	for _, parameterRef := range parameters {
		if parameterRef == nil || parameterRef.Value == nil {
			continue
		}

		if !strings.EqualFold(parameterRef.Value.In, "path") {
			continue
		}

		names[parameterRef.Value.Name] = struct{}{}
	}

	return names
}

// extractPathParameterNames 从 URL 模板中提取 {param} 形式的路径参数名。
func extractPathParameterNames(path string) []string {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	names := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		names = append(names, match[1])
	}

	return names
}

// normalizeSecurityRequirements 将全局或接口级安全声明压平成稳定的字符串集合，便于比较。
func normalizeSecurityRequirements(global openapi3.SecurityRequirements, override *openapi3.SecurityRequirements) []string {
	requirements := global
	if override != nil {
		requirements = *override
	}

	values := make([]string, 0, len(requirements))

	for _, requirement := range requirements {
		keys := make([]string, 0, len(requirement))
		for key := range requirement {
			keys = append(keys, key)
		}

		sort.Strings(keys)
		values = append(values, strings.Join(keys, "+"))
	}

	return normalizeStringSlice(values)
}

// normalizeSecurityRequirementsRef 归一化安全声明中的 scheme 名称，并保证空 scope 序列化为数组而不是 null。
func normalizeSecurityRequirementsRef(requirements openapi3.SecurityRequirements) openapi3.SecurityRequirements {
	if requirements == nil {
		return nil
	}

	normalized := make(openapi3.SecurityRequirements, 0, len(requirements))

	for _, requirement := range requirements {
		if requirement == nil {
			normalized = append(normalized, openapi3.SecurityRequirement{})
			continue
		}

		next := openapi3.SecurityRequirement{}

		for key, scopes := range requirement {
			targetKey := key
			if key == bearerSecurityName {
				targetKey = apiKeySecurityName
			}

			next[targetKey] = cloneSecurityScopes(scopes)
		}

		normalized = append(normalized, next)
	}

	return normalized
}

// cloneSecurityScopes 复制 scope 列表，并将空值显式表示为空数组。
func cloneSecurityScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return []string{}
	}

	return append([]string{}, scopes...)
}

// securityRequirementsEqual 比较两组安全声明在名称层面是否等价。
func securityRequirementsEqual(left openapi3.SecurityRequirements, right openapi3.SecurityRequirements) bool {
	return stringSlicesEqual(normalizeSecurityRequirementNames(left), normalizeSecurityRequirementNames(right))
}

// normalizeSecurityRequirementNames 提取每个 security requirement 的 scheme 名称并做排序归一化。
func normalizeSecurityRequirementNames(requirements openapi3.SecurityRequirements) []string {
	values := make([]string, 0, len(requirements))

	for _, requirement := range requirements {
		keys := make([]string, 0, len(requirement))
		for key := range requirement {
			keys = append(keys, key)
		}

		sort.Strings(keys)
		values = append(values, strings.Join(keys, "+"))
	}

	return normalizeStringSlice(values)
}
