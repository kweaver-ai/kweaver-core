package openapidoc

import (
	"fmt"
	"strconv"
	"strings"
)

// sanitizeOpenAPIMap 从文档根节点开始清洗不兼容或历史遗留的字段形态。
func sanitizeOpenAPIMap(node map[string]any) {
	sanitizeOpenAPIMapWithContext(node, false, false)
}

// sanitizeOpenAPIMapWithContext 根据当前是否处于 schema/example 上下文，递归修正字段格式。
func sanitizeOpenAPIMapWithContext(node map[string]any, inSchema bool, inExample bool) {
	for key, value := range node {
		if !inExample && key == "required" {
			if _, isBool := value.(bool); isBool {
				delete(node, key)
				continue
			}
		}

		if !inExample && key == "description" {
			if descriptionLines, ok := value.([]any); ok {
				node[key] = joinStringList(descriptionLines)
				continue
			}
		}

		if inSchema && !inExample && key == "type" {
			if _, isList := value.([]any); isList {
				delete(node, key)
				continue
			}

			if normalizedType, keep := normalizeSchemaType(value); keep {
				node[key] = normalizedType
			} else {
				delete(node, key)
				continue
			}
		}

		sanitizeOpenAPIValue(value, inSchema || isSchemaContextKey(key), inExample || key == "example" || key == "examples")
	}

	if inSchema && !inExample {
		if schemaType, ok := node["type"].(string); ok {
			if schemaType == "array" {
				if _, hasItems := node["items"]; !hasItems {
					node["items"] = map[string]any{}
				}
			}

			if sanitizedExample, exists := node["example"]; exists {
				if normalizedExample, keep := sanitizeExampleForSchemaType(schemaType, sanitizedExample); keep {
					node["example"] = normalizedExample
				} else {
					delete(node, "example")
				}
			}
		}
	}
}

// sanitizeOpenAPIValue 递归处理 map 或数组里的嵌套值。
func sanitizeOpenAPIValue(value any, inSchema bool, inExample bool) {
	switch child := value.(type) {
	case map[string]any:
		sanitizeOpenAPIMapWithContext(child, inSchema, inExample)
	case []any:
		for _, item := range child {
			sanitizeOpenAPIValue(item, inSchema, inExample)
		}
	}
}

// isSchemaContextKey 判断某个 key 是否意味着其子节点处于 schema 语境中。
func isSchemaContextKey(key string) bool {
	switch key {
	case "schema", "schemas", "properties", "items", "additionalProperties", "allOf", "anyOf", "oneOf", "not":
		return true
	default:
		return false
	}
}

// normalizeSchemaType 将历史或非标准类型名映射为 OpenAPI 可接受的 schema type。
func normalizeSchemaType(value any) (string, bool) {
	schemaType, ok := value.(string)
	if !ok {
		return "", false
	}

	switch strings.TrimSpace(schemaType) {
	case "string", "number", "integer", "boolean", "array", "object":
		return schemaType, true
	case "int":
		return "integer", true
	case "float", "double":
		return "number", true
	case "any", "null":
		return "", false
	default:
		return schemaType, true
	}
}

// joinStringList 将多行描述数组拼成单个字符串描述。
func joinStringList(items []any) string {
	lines := make([]string, 0, len(items))

	for _, item := range items {
		line := strings.TrimSpace(fmt.Sprint(item))
		if line == "" {
			continue
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// sanitizeExampleForSchemaType 根据 schema type 校正 example 的值类型。
func sanitizeExampleForSchemaType(schemaType string, example any) (any, bool) {
	switch schemaType {
	case "string":
		return fmt.Sprint(example), true
	case "integer":
		switch v := example.(type) {
		case int:
			return v, true
		case int64:
			return v, true
		case float64:
			if v == float64(int64(v)) {
				return int64(v), true
			}
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err == nil {
				return parsed, true
			}
		}

		return nil, false
	case "number":
		switch v := example.(type) {
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		case float64:
			return v, true
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err == nil {
				return parsed, true
			}
		}

		return nil, false
	case "boolean":
		switch v := example.(type) {
		case bool:
			return v, true
		case string:
			parsed, err := strconv.ParseBool(strings.TrimSpace(v))
			if err == nil {
				return parsed, true
			}
		}

		return nil, false
	case "array":
		if values, ok := example.([]any); ok {
			return values, true
		}

		return nil, false
	case "object":
		if valueMap, ok := example.(map[string]any); ok {
			return valueMap, true
		}

		return nil, false
	default:
		return example, true
	}
}
