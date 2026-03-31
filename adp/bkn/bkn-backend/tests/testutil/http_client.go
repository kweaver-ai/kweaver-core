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
	"mime/multipart"
	"net/http"
	"time"
)

// HTTPClient 封装HTTP请求的测试客户端
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Headers map[string]string // 包含X-Account-ID等公共头
}

// HTTPResponse HTTP响应封装
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header    // 响应头
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
			"Content-Type":      "application/json",
			"X-Account-ID":      "test-user-001",
			"X-Account-Type":    "user",
			"x-business-domain": "test-domain",
		},
	}
}

// CheckHealth 检查服务健康状态
func (c *HTTPClient) CheckHealth() error {
	resp, err := c.Client.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	defer resp.Body.Close()

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

// POSTMultipart 发送 multipart/form-data POST 请求（用于文件上传）
func (c *HTTPClient) POSTMultipart(path string, fileFieldName string, fileContent []byte, fileName string, extraParams map[string]string) HTTPResponse {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// 添加文件
	part, err := writer.CreateFormFile(fileFieldName, fileName)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "create_form_error", ErrorDetails: err.Error()},
		}
	}
	_, err = part.Write(fileContent)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "write_file_error", ErrorDetails: err.Error()},
		}
	}

	// 添加额外参数
	for key, val := range extraParams {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "close_writer_error", ErrorDetails: err.Error()},
		}
	}

	// 构建完整URL
	url := c.BaseURL + path

	// 创建请求
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "create_request_error", ErrorDetails: err.Error()},
		}
	}

	// 设置 headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "request_error", ErrorDetails: err.Error()},
		}
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{
			StatusCode: resp.StatusCode,
			Error:      &ErrorResponse{ErrorCode: "read_body_error", ErrorDetails: err.Error()},
		}
	}

	// 解析响应
	result := HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		RawBody:    respBody,
	}

	// 尝试解析 JSON
	if len(respBody) > 0 {
		var jsonBody map[string]any
		if err := json.Unmarshal(respBody, &jsonBody); err == nil {
			result.Body = jsonBody
		}
	}

	return result
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
			Error:      &ErrorResponse{ErrorCode: "create_request_error", ErrorDetails: err.Error()},
		}
	}

	// 设置 headers
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return HTTPResponse{
			StatusCode: 0,
			Error:      &ErrorResponse{ErrorCode: "request_error", ErrorDetails: err.Error()},
		}
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{
			StatusCode: resp.StatusCode,
			Error:      &ErrorResponse{ErrorCode: "read_body_error", ErrorDetails: err.Error()},
		}
	}

	// 解析响应
	result := HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		RawBody:    respBody,
	}

	// 尝试解析 JSON
	if len(respBody) > 0 {
		var jsonBody map[string]any
		if err := json.Unmarshal(respBody, &jsonBody); err == nil {
			result.Body = jsonBody
		} else {
			// 如果不是 JSON，保留原始内容
			result.Body = map[string]any{"raw": string(respBody)}
		}
	}

	return result
}
