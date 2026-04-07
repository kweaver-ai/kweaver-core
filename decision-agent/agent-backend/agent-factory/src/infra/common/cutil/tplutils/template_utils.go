package tplutils

import (
	"bytes"
	"html/template"
)

// RenderTemplate 渲染模板字符串，返回渲染后的结果
// templateStr: 模板字符串
// data: 模板参数，使用map[string]interface{}支持任意类型的数据
// 返回渲染后的字符串和可能的错误
func RenderTemplate(templateStr string, data map[string]interface{}) (string, error) {
	// 解析模板字符串
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return "", err
	}

	// 创建buffer来存储输出
	var buf bytes.Buffer

	// 执行模板渲染
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	// 返回渲染结果
	return buf.String(), nil
}
