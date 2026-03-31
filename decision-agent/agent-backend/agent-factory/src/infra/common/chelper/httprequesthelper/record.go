package httprequesthelper

import (
	"net/http"
	"time"
)

// RequestRecord 请求记录
type RequestRecord struct {
	// Timestamp 请求时间戳
	Timestamp time.Time `json:"timestamp"`

	// Method HTTP方法
	Method string `json:"method"`

	// URL 请求URL
	URL string `json:"url"`

	// Headers 请求头
	Headers map[string]string `json:"headers,omitempty"`

	// Body 请求体
	Body string `json:"body,omitempty"`
}

// ResponseRecord 响应记录
type ResponseRecord struct {
	// StatusCode HTTP状态码
	StatusCode int `json:"status_code"`

	// Headers 响应头
	Headers map[string]string `json:"headers,omitempty"`

	// Body 响应体
	Body string `json:"body,omitempty"`

	// Duration 请求耗时（毫秒）
	DurationMs float64 `json:"duration_ms"`
}

// LogRecord 完整的日志记录
type LogRecord struct {
	Request  RequestRecord  `json:"request"`
	Response ResponseRecord `json:"response"`
}

// NewRequestRecord 从http.Request创建请求记录
func NewRequestRecord(req *http.Request, body string, includeHeaders bool) RequestRecord {
	// 构建完整的 URL
	fullURL := req.URL.String()

	if req.URL.Host == "" && req.Host != "" {
		// 如果 URL 中没有 Host，使用 req.Host 构建完整 URL
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}

		fullURL = scheme + "://" + req.Host + req.URL.RequestURI()
	}

	record := RequestRecord{
		Timestamp: time.Now(),
		Method:    req.Method,
		URL:       fullURL,
		Body:      body,
	}

	if includeHeaders {
		record.Headers = make(map[string]string)

		for k, v := range req.Header {
			if len(v) > 0 {
				record.Headers[k] = v[0]
			}
		}
	}

	return record
}

// NewResponseRecord 从http.Response创建响应记录
func NewResponseRecord(statusCode int, headers http.Header, body string, duration time.Duration, includeHeaders bool) ResponseRecord {
	record := ResponseRecord{
		StatusCode: statusCode,
		Body:       body,
		DurationMs: float64(duration.Milliseconds()),
	}

	if includeHeaders && headers != nil {
		record.Headers = make(map[string]string)

		for k, v := range headers {
			if len(v) > 0 {
				record.Headers[k] = v[0]
			}
		}
	}

	return record
}
