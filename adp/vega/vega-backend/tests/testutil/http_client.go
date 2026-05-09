// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTPClient 封装HTTP请求的测试客户端
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Headers map[string]string // 包含X-Account-ID等公共头
	mu      sync.RWMutex      // 保护Headers的并发访问
}

// HTTPResponse HTTP响应封装
type HTTPResponse struct {
	StatusCode int
	Body       map[string]any // 成功响应的JSON body
	Error      *ErrorResponse // 错误响应
	RawBody    []byte         // 原始响应体
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	ErrorCode    string `json:"error_code"`
	ErrorDetails string `json:"error_details"`
}

// NewHTTPClient 创建新的HTTP测试客户端
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 120 * time.Second},
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"X-Account-ID":   "test-user-001",
			"X-Account-Type": "user",
		},
	}
}

// CheckHealth 检查服务健康状态
func (c *HTTPClient) CheckHealth() error {
	resp, err := c.Client.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("健康检查返回状态码 %d", resp.StatusCode)
	}
	return nil
}

// POST 发送POST请求
func (c *HTTPClient) POST(path string, payload any) HTTPResponse {
	return c.doRequest("POST", path, payload)
}

// GET 发送GET请求
func (c *HTTPClient) GET(path string) HTTPResponse {
	return c.doRequest("GET", path, nil)
}

// PUT 发送PUT请求
func (c *HTTPClient) PUT(path string, payload any) HTTPResponse {
	return c.doRequest("PUT", path, payload)
}

// DELETE 发送DELETE请求
func (c *HTTPClient) DELETE(path string) HTTPResponse {
	return c.doRequest("DELETE", path, nil)
}

// doRequest 执行HTTP请求的内部方法
func (c *HTTPClient) doRequest(method, path string, payload any) HTTPResponse {
	var bodyReader io.Reader
	if payload != nil {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return HTTPResponse{
				StatusCode: 0,
				Error:      &ErrorResponse{ErrorCode: "marshal_error", ErrorDetails: err.Error()},
			}
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	// 构建完整URL
	url := c.BaseURL + path

	// 创建请求
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "request_error", ErrorDetails: err.Error()},
		}
	}

	// 设置公共头
	c.mu.RLock()
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	c.mu.RUnlock()

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "network_error", ErrorDetails: err.Error()},
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return c.parseResponse(resp)
}

// parseResponse 解析HTTP响应
func (c *HTTPClient) parseResponse(resp *http.Response) HTTPResponse {
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{
			StatusCode: resp.StatusCode,
			Error:      &ErrorResponse{ErrorCode: "read_error", ErrorDetails: err.Error()},
		}
	}

	result := HTTPResponse{
		StatusCode: resp.StatusCode,
		RawBody:    rawBody,
	}

	// 如果响应体为空，直接返回
	if len(rawBody) == 0 {
		return result
	}

	// 尝试解析为JSON
	var bodyMap map[string]any
	if err := json.Unmarshal(rawBody, &bodyMap); err != nil {
		// 无法解析为JSON，保留原始body
		result.Error = &ErrorResponse{ErrorCode: "parse_error", ErrorDetails: err.Error()}
		return result
	}

	// 检查是否是错误响应（包含error_code字段）
	if errorCode, ok := bodyMap["error_code"].(string); ok {
		errorDetails := ""
		if details, exists := bodyMap["error_details"].(string); exists {
			errorDetails = details
		}
		result.Error = &ErrorResponse{
			ErrorCode:    errorCode,
			ErrorDetails: errorDetails,
		}
	} else {
		// 正常响应
		result.Body = bodyMap
	}

	return result
}

// SetHeader 设置自定义请求头
func (c *HTTPClient) SetHeader(key, value string) {
	c.mu.Lock()
	c.Headers[key] = value
	c.mu.Unlock()
}

// RemoveHeader 移除请求头
func (c *HTTPClient) RemoveHeader(key string) {
	c.mu.Lock()
	delete(c.Headers, key)
	c.mu.Unlock()
}
