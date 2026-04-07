package httprequesthelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Formatter 日志格式化器
type Formatter struct {
	prettyJSON  bool
	maxBodySize int
}

// NewFormatter 创建格式化器
func NewFormatter(prettyJSON bool, maxBodySize int) *Formatter {
	return &Formatter{
		prettyJSON:  prettyJSON,
		maxBodySize: maxBodySize,
	}
}

// Format 格式化日志记录
func (f *Formatter) Format(record *LogRecord) string {
	var buf bytes.Buffer

	// 分隔线
	buf.WriteString("\n")
	buf.WriteString(strings.Repeat("=", 80))
	buf.WriteString("\n")

	// 请求信息
	buf.WriteString(fmt.Sprintf("📤 请求开始: %s\n", record.Request.Timestamp.Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("📍 %s %s\n", record.Request.Method, record.Request.URL))

	// 请求头
	if len(record.Request.Headers) > 0 {
		buf.WriteString("📋 请求头:\n")
		buf.WriteString(f.formatHeaders(record.Request.Headers))
	}

	// 请求体
	if record.Request.Body != "" {
		buf.WriteString("📦 请求体:\n")
		buf.WriteString(f.formatBody(record.Request.Body))
		buf.WriteString("\n")
	}

	// CURL命令
	curlCmd := f.generateCURL(record)
	buf.WriteString(fmt.Sprintf("🔧 CURL: %s\n", curlCmd))

	// 分隔线
	buf.WriteString(strings.Repeat("-", 40))
	buf.WriteString("\n")

	// 响应信息
	buf.WriteString(fmt.Sprintf("📥 响应: HTTP %d (%.2fms)\n", record.Response.StatusCode, record.Response.DurationMs))

	// 响应头
	if len(record.Response.Headers) > 0 {
		buf.WriteString("📋 响应头:\n")
		buf.WriteString(f.formatHeaders(record.Response.Headers))
	}

	// 响应体
	if record.Response.Body != "" {
		buf.WriteString("📦 响应体:\n")
		buf.WriteString(f.formatBody(record.Response.Body))
		buf.WriteString("\n")
	}

	// 结束分隔线
	buf.WriteString(strings.Repeat("=", 80))
	buf.WriteString("\n")

	return buf.String()
}

// formatHeaders 格式化请求头
func (f *Formatter) formatHeaders(headers map[string]string) string {
	var buf bytes.Buffer
	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("   %s: %s\n", k, v))
	}

	return buf.String()
}

// formatBody 格式化body内容
func (f *Formatter) formatBody(body string) string {
	// 截断过长的body
	if f.maxBodySize > 0 && len(body) > f.maxBodySize {
		body = body[:f.maxBodySize] + fmt.Sprintf("...[truncated, total %d bytes]", len(body))
	}

	// 尝试格式化JSON
	if f.prettyJSON && isJSON(body) {
		formatted, err := formatJSON(body)
		if err == nil {
			return formatted
		}
	}

	return body
}

// generateCURL 生成CURL命令
func (f *Formatter) generateCURL(record *LogRecord) string {
	var parts []string
	parts = append(parts, "curl")

	// 方法
	if record.Request.Method != "GET" {
		parts = append(parts, "-X", record.Request.Method)
	}

	// 请求头
	for k, v := range record.Request.Headers {
		// 跳过一些常见的由代理添加的头部
		lk := strings.ToLower(k)
		if lk == "host" || lk == "connection" || lk == "content-length" {
			continue
		}

		parts = append(parts, "-H", fmt.Sprintf("'%s: %s'", k, v))
	}

	// 请求体
	if record.Request.Body != "" && (record.Request.Method == "POST" || record.Request.Method == "PUT" || record.Request.Method == "PATCH") {
		body := record.Request.Body
		// 转义单引号
		body = strings.ReplaceAll(body, "'", "'\\''")
		parts = append(parts, "-d", fmt.Sprintf("'%s'", body))
	}

	// URL
	parts = append(parts, fmt.Sprintf("'%s'", record.Request.URL))

	return strings.Join(parts, " ")
}

// isJSON 检查字符串是否为JSON格式
func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

// formatJSON 格式化JSON字符串
func formatJSON(s string) (string, error) {
	var buf bytes.Buffer

	err := json.Indent(&buf, []byte(s), "", "  ")
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
