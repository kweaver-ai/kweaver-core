package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"github.com/pkg/errors"
)

// Forwarder HTTP请求转发器接口
type Forwarder interface {
	Forward(ctx context.Context, req *interfaces.HTTPRequest) (*interfaces.HTTPResponse, error)
	ForwardStream(ctx context.Context, req *interfaces.HTTPRequest) (*interfaces.HTTPResponse, error)
}

// forwarder HTTP请求转发器
type forwarder struct {
	pool            *clientPool
	streamProcessor *StreamProcessor
	logger          interfaces.Logger
}

var (
	forwarderOnce sync.Once
	f             Forwarder
)

// NewForwarder 创建一个新的HTTP请求转发器
func NewForwarder() Forwarder {
	forwarderOnce.Do(func() {
		logger := config.NewConfigLoader().GetLogger()
		f = &forwarder{
			pool:            NewClientPool(),
			streamProcessor: NewStreamProcessor(logger),
			logger:          logger,
		}
	})
	return f
}

// HTTPStreamForward 处理HTTP流式请求
func (f *forwarder) ForwardStream(ctx context.Context, req *interfaces.HTTPRequest) (*interfaces.HTTPResponse, error) {
	startTime := time.Now()
	// 验证请求参数
	streamingMode, ok := common.GetStreamingModeFromCtx(ctx)
	if !ok {
		streamingMode = interfaces.StreamingModeHTTP
	}
	// 获取响应写入器
	headerWriter, ok := common.GetResponseWriterFromCtx(ctx)
	if !ok {
		headerWriter.WriteHeader(http.StatusInternalServerError)
		err := fmt.Errorf("response writer not found in context")
		f.logger.WithContext(ctx).Warnf("failed to forward stream, err: %v", err)
		err = myErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	httpReq, err := f.buildRequest(req)
	if err != nil {
		headerWriter.WriteHeader(http.StatusInternalServerError)
		f.logger.WithContext(ctx).Warnf("build request failed, err: %v", err)
		err = myErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	// 创建不带超时的客户端用于流式请求
	streamClient := f.pool.GetStreamClient(streamingMode, req.Timeout)

	// 新添加：为流式请求设置必要的请求头
	prepareStreamRequest(streamingMode, httpReq)
	now := time.Now()
	f.logger.Debugf("do stream request, streamingMode: %v, timeout: %v", streamingMode, req.Timeout)
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		// 流式服务端，设置ResponseHeaderTimeout参数为10s, 10s内无请求头返回，认为是超时
		// 如果服务端不支持流式请求，但是客户端使用流式代理，也认为是超时
		// 超时错误示例: net/http: timeout awaiting response headers"
		headerWriter.WriteHeader(http.StatusRequestTimeout)
		if strings.Contains(err.Error(), "timeout awaiting response headers") {
			err = errors.Wrapf(err, "The server may not support streaming requests, or the server response timed out with no response headers received within 10 seconds")
			err = myErr.DefaultHTTPError(ctx, http.StatusRequestTimeout, err.Error())
		} else {
			// 请求转发失败，返回服务器不可用，请检查是否可用，或稍后重试
			err = errors.Wrapf(err, "Request forwarding failed, please check if the request is correct, or try again later")
			err = myErr.NewHTTPError(ctx, http.StatusServiceUnavailable, myErr.ErrExtProxyForwardFailed, err.Error())
		}
		f.logger.WithContext(ctx).Warnf("do request failed, err: %v, cost: %v", err, time.Since(now))
		return nil, err
	}
	f.logger.Debugf("do stream request success, streamingMode: %v, timeout: %v, cost: %v", streamingMode, req.Timeout, time.Since(now))
	defer func() {
		_ = resp.Body.Close()
	}()
	// 复制响应头到原始响应
	var isSSE bool
	headers := make(map[string]any)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
			if key == "Content-Type" && strings.HasPrefix(values[0], "text/event-stream") {
				isSSE = true
			}
			headerWriter.Header().Set(key, values[0])
		}
	}
	preprocessResponseHeaders(streamingMode, headerWriter)
	// 确保设置正确的状态码
	headerWriter.WriteHeader(resp.StatusCode)
	// 添加这行代码以确保响应头被立即发送
	if flusher, ok := headerWriter.(http.Flusher); ok {
		flusher.Flush()
	}
	// 根据流式模式处理
	switch streamingMode {
	case interfaces.StreamingModeSSE:
		err = f.streamProcessor.ProcessSSE(ctx, resp.Body, headerWriter, isSSE)
	case interfaces.StreamingModeHTTP:
		err = f.streamProcessor.ProcessHTTPStream(ctx, resp.Body, headerWriter)
	default:
		_, err = io.Copy(headerWriter, resp.Body)
	}
	if err != nil {
		err = fmt.Errorf("failed to forward stream: %v", err)
		f.logger.WithContext(ctx).Warnf("failed to forward stream, err: %v", err)
		err = myErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	return &interfaces.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Duration:   time.Since(startTime).Milliseconds(),
	}, nil
}

// Forward 转发HTTP请求
func (f *forwarder) Forward(ctx context.Context, req *interfaces.HTTPRequest) (*interfaces.HTTPResponse, error) {
	startTime := time.Now()

	// 获取HTTP客户端
	client := f.pool.GetClient(req.Timeout)

	// 构建HTTP请求
	httpReq, err := f.buildRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return &interfaces.HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Error:      err.Error(),
			Duration:   time.Since(startTime).Milliseconds(),
		}, nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 处理响应
	return f.processResponse(resp, startTime)
}

// buildRequest 根据请求参数构建HTTP请求
func (f *forwarder) buildRequest(req *interfaces.HTTPRequest) (*http.Request, error) {
	// 处理URL和路径参数
	requestURL := req.URL
	if len(req.PathParams) > 0 {
		for key, value := range req.PathParams {
			requestURL = strings.Replace(requestURL, fmt.Sprintf("{%s}", key), value, -1)
			requestURL = strings.Replace(requestURL, fmt.Sprintf(":{%s}", key), value, -1)
			requestURL = strings.Replace(requestURL, fmt.Sprintf(":%s", key), value, -1)
		}
	}
	// 处理查询参数
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(requestURL)
		if err != nil {
			return nil, err
		}

		q := parsedURL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, fmt.Sprintf("%v", value))
		}

		parsedURL.RawQuery = q.Encode()
		requestURL = parsedURL.String()
	}

	// 处理请求体
	var reqBody io.Reader
	var contentType string

	if req.Body != nil {
		// 检查Content-Type头
		contentType = ""
		if req.Headers != nil {
			for k, v := range req.Headers {
				if strings.EqualFold(k, "content-type") {
					contentType = fmt.Sprintf("%v", v)
					break
				}
			}
		}
		// 根据Content-Type处理请求体
		switch {
		case strings.Contains(contentType, "application/json"):
			// JSON格式
			jsonData, err := json.Marshal(req.Body)
			if err != nil {
				return nil, err
			}
			reqBody = bytes.NewBuffer(jsonData)
		case strings.Contains(contentType, "application/x-www-form-urlencoded"):
			// 表单格式
			formData := url.Values{}

			// 尝试将body转换为map
			if bodyMap, ok := req.Body.(map[string]interface{}); ok {
				for key, value := range bodyMap {
					formData.Add(key, fmt.Sprintf("%v", value))
				}
			}
			reqBody = strings.NewReader(formData.Encode())
		case strings.Contains(contentType, "multipart/form-data"):
			// 多部分表单格式
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// 尝试将body转换为map
			if bodyMap, ok := req.Body.(map[string]interface{}); ok {
				for key, value := range bodyMap {
					fw, err := writer.CreateFormField(key)
					if err != nil {
						return nil, err
					}
					_, err = fw.Write(utils.ObjectToByte(value))
					if err != nil {
						return nil, err
					}
				}
			}

			contentType = writer.FormDataContentType()
			_ = writer.Close()
			reqBody = body
		case strings.Contains(contentType, "text/plain"):
			// 文本格式
			reqBody = strings.NewReader(fmt.Sprintf("%v", req.Body))
		case strings.Contains(contentType, "text/event-stream"):
			// SSE格式
			reqBody = strings.NewReader(fmt.Sprintf("%v", req.Body))
		case strings.Contains(contentType, "application/stream+json"), // HTTP Streaming格式
			strings.Contains(contentType, "application/x-ndjson"),      // NDJSON格式
			strings.Contains(contentType, "application/x-json-stream"): // HTTP Streaming格式
			// HTTP Streaming格式
			reqBody = strings.NewReader(fmt.Sprintf("%v", req.Body))
		default:
			jsonData, err := json.Marshal(req.Body)
			if err != nil {
				return nil, err
			}
			reqBody = bytes.NewBuffer(jsonData)
			if contentType == "" {
				contentType = "application/json"
			}
		}
	} else {
		reqBody = http.NoBody
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest(req.Method, requestURL, reqBody)
	if err != nil {
		err = fmt.Errorf("failed to create request: %v", err)
		return nil, err
	}

	// 设置请求头
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	// 如果Content-Type未在请求头中设置，但我们有确定的类型，则设置它
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	return httpReq, nil
}

// processResponse 处理HTTP响应
func (f *forwarder) processResponse(resp *http.Response, startTime time.Time) (*interfaces.HTTPResponse, error) {
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应头
	headers := make(map[string]any)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 尝试解析JSON响应
	var responseBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(body, &responseBody); err != nil {
			// 如果解析失败，使用原始响应体
			responseBody = string(body)
		}
	} else {
		// 非JSON响应，使用字符串
		responseBody = string(body)
	}

	return &interfaces.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       responseBody,
		Duration:   time.Since(startTime).Milliseconds(),
	}, nil
}

func prepareStreamRequest(streamingMode interfaces.StreamingMode, req *http.Request) {
	// 设置流类型特定请求头
	switch streamingMode {
	case interfaces.StreamingModeSSE:
		// 兼容处理Accept头: 第一个必须为text/event-stream，使用*/*作为兜底
		acceptValue := req.Header.Get("Accept")
		acceptValues := strings.Split(acceptValue, ",")
		if len(acceptValues) == 0 || acceptValues[0] != "text/event-stream" {
			acceptValues = append([]string{"text/event-stream"}, acceptValues...)
		}
		acceptValues = append(acceptValues, "*/*")
		// 去重
		acceptValues = utils.UniqueStrings(acceptValues)
		req.Header.Set("Accept", strings.Join(acceptValues, ", "))
		req.Header.Set("Cache-Control", "no-cache") // 禁用缓存
		req.Header.Set("Connection", "keep-alive")  // 保持连接
	case interfaces.StreamingModeHTTP:
		req.Header.Set("Transfer-Encoding", "chunked") // 分块传输
		req.Header.Set("Connection", "Upgrade")        // 升级连接
	}
}

// 预处理响应头
func preprocessResponseHeaders(streamingMode interfaces.StreamingMode, headerWriter http.ResponseWriter) {
	// 根据流式模式处理
	switch streamingMode {
	case interfaces.StreamingModeSSE:
		headerWriter.Header().Set("Content-Type", "text/event-stream")
		headerWriter.Header().Set("Cache-Control", "no-cache")
		headerWriter.Header().Set("Connection", "keep-alive")
		// 移除可能存在的Content-Length头部
		headerWriter.Header().Del("Content-Length")
	case interfaces.StreamingModeHTTP:
		headerWriter.Header().Set("Transfer-Encoding", "chunked")
		headerWriter.Header().Set("X-Content-Type-Options", "nosniff")
		// 移除可能存在的Content-Length头部
		headerWriter.Header().Del("Content-Length")
	}
}
