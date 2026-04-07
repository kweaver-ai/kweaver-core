package tplutils

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// SafeGet 安全地获取嵌套map中的值
func SafeGet(data interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil
		}

		v := reflect.ValueOf(current)
		if v.Kind() == reflect.Map {
			value := v.MapIndex(reflect.ValueOf(part))
			if !value.IsValid() {
				return nil
			}

			current = value.Interface()
		} else {
			return nil
		}
	}

	return current
}

// SafeRenderTemplate 安全的模板渲染，变量不存在时保留占位符
func SafeRenderTemplate(templateStr string, data map[string]interface{}) (string, error) {
	// 创建模板函数
	funcMap := template.FuncMap{
		"safe": func(path string) string {
			value := SafeGet(data, path)
			if value == nil {
				// 返回原始占位符
				return fmt.Sprintf("{{.%s}}", path)
			}
			// 如果是空字符串，直接返回
			if str, ok := value.(string); ok && str == "" {
				return ""
			}
			// 如果是nil，返回占位符
			if value == nil {
				return fmt.Sprintf("{{.%s}}", path)
			}

			jsonStr, err := cutil.JSON().MarshalToString(value)
			if err != nil {
				panic(fmt.Sprintf("序列化失败: %v", err))
			}

			// 去除首尾的引号
			jsonStr = strings.Trim(jsonStr, "\"")
			return jsonStr
		},
	}

	// 预处理模板字符串，将 {{.xxx}} 转换为 {{safe "xxx"}}
	processedTemplate := templateStr
	// 匹配 {{.xxx}} 格式
	parts := strings.Split(processedTemplate, "{{.")
	if len(parts) > 1 {
		var result []string
		result = append(result, parts[0])

		for _, part := range parts[1:] {
			if idx := strings.Index(part, "}}"); idx != -1 {
				path := part[:idx]
				rest := part[idx+2:]
				result = append(result, fmt.Sprintf(`{{safe "%s"}}%s`, path, rest))
			} else {
				result = append(result, part)
			}
		}

		processedTemplate = strings.Join(result, "")
	}

	// 创建并解析模板
	tmpl, err := template.New("safe_template").
		Funcs(funcMap).
		Parse(processedTemplate)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %v", err)
	}

	// 执行模板
	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("执行模板失败: %v", err)
	}

	return buf.String(), nil
}
