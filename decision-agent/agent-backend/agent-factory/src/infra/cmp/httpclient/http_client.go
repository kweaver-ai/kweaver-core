// Package httpclient httpclient 客户端基础实现
package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

//go:generate mockgen -package mock -source ./http_client.go -destination ./mock/mock_http_client.go

// HTTPClient HTTP客户端服务接口
type HTTPClient interface {
	Get(url string, headers map[string]string) (respCode int, respParam []byte, err error)
	Post(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam []byte, err error)
	Put(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam []byte, err error)
	Delete(url string, headers map[string]string) (respParam []byte, err error)
	HttpForward(r *http.Request, url string, method string) (respCode int, body []byte, contentType string, err error)
	StreamPost(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (chan string, chan error, error)
}

var (
	rawOnce   sync.Once
	rawClient *http.Client
	httpOnce  sync.Once
	client    HTTPClient
)

// httpClient HTTP客户端结构
type httpClient struct {
	client *http.Client
}

// NewRawHTTPClient 创建原生HTTP客户端对象
func NewRawHTTPClient() *http.Client {
	rawOnce.Do(func() {
		rawClient = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 100 * time.Second,
		}
	})

	return rawClient
}

// NewHTTPClient 创建HTTP客户端对象
func NewHTTPClient() HTTPClient {
	httpOnce.Do(func() {
		client = &httpClient{
			client: NewRawHTTPClient(),
		}
	})

	return client
}

// NewHTTPClientEx 创建HTTP客户端对象, 自定义超时时间
func NewHTTPClientEx(timeout time.Duration) HTTPClient {
	return &httpClient{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: timeout * time.Second,
		},
	}
}

// Get http client get
func (c *httpClient) Get(url string, headers map[string]string) (respCode int, respParam []byte, err error) {
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return
	}

	respCode, respParam, _, err = c.httpDo(req, headers)
	if err != nil {
		return
	}

	return
}

// Post http client post
func (c *httpClient) Post(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam []byte, err error) {
	var reqBody []byte
	if v, ok := reqParam.([]byte); ok {
		reqBody = v
	} else {
		reqBody, err = jsoniter.Marshal(reqParam)
		if err != nil {
			return
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, _, err = c.httpDo(req, headers)

	return
}

// Post http client post
func (c *httpClient) StreamPost(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (
	chan string, chan error, error,
) {
	messages := make(chan string)
	errs := make(chan error)

	go func() {
		defer close(messages)
		defer close(errs)

		var reqBody []byte
		if v, ok := reqParam.([]byte); ok {
			reqBody = v
		} else {
			var err error

			reqBody, err = jsoniter.Marshal(reqParam)
			if err != nil {
				errs <- err
				return
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
		// req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			errs <- err
			return
		}
		// 应用传入的Header
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		// req.Header.Set("Content-Type", "text/event-stream")
		// req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("X-Accel-Buffering", "no")
		req.Header.Set("Connection-Type", "keep-alive")

		resp, err := c.client.Do(req)
		if err != nil {
			errs <- err
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				errs <- err
				return
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			// 如果HTTP响应状态码不是200，读取响应体并将其作为错误处理
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				err = readErr
				errs <- err
			} else {
				err = fmt.Errorf("HTTP request failed with status: %s, body: %s", resp.Status, body)
				errs <- err
			}

			return
		}

		reader := bufio.NewReader(resp.Body)

		var currentEvent []byte

		for {
			line, isPrefix, err := reader.ReadLine()
			if err != nil {
				if err.Error() == "unexpected EOF" {
					// 为空直接结束不报错
					err = nil
					return
				}
				errs <- err

				return
			}

			if !isPrefix {
				currentEvent = append(currentEvent, line...)

				message := string(currentEvent)
				if len(message) > 0 {
					messages <- message // 直接转发接收到的消息
				}

				currentEvent = nil
			} else {
				currentEvent = append(currentEvent, line...)
			}
		}
	}()

	return messages, errs, nil
}

// Put http client put
func (c *httpClient) Put(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam []byte, err error) {
	reqBody, err := jsoniter.Marshal(reqParam)
	if err != nil {
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, _, err = c.httpDo(req, headers)

	return
}

// Delete http client delete
func (c *httpClient) Delete(url string, headers map[string]string) (respParam []byte, err error) {
	req, err := http.NewRequest("DELETE", url, http.NoBody)
	if err != nil {
		return
	}

	_, respParam, _, err = c.httpDo(req, headers)

	return
}

// HttpForward 转发request，需要保证request里的body可以读取
func (c *httpClient) HttpForward(r *http.Request, url, method string) (respCode int, body []byte, contentType string, err error) {
	req, err := http.NewRequest(method, url, r.Body)
	if err != nil {
		return
	}

	req.Header = r.Header
	req.Form = r.Form
	respCode, body, contentType, err = c.httpDo(req, nil)

	return
}

func (c *httpClient) httpDo(req *http.Request, headers map[string]string) (respCode int, body []byte, contentType string, err error) {
	if c.client == nil {
		return 0, nil, "", errors.New("http client is unavailable")
	}

	c.addHeaders(req, headers)

	start := time.Now()
	resp, err := c.client.Do(req)
	// 计算请求耗时
	duration := time.Since(start)
	// 检查是否超过1秒
	if duration > time.Second {
		// 只记录URL的scheme和host，避免泄露完整URL中的敏感参数
		sanitizedURL := fmt.Sprintf("%s://%s", req.URL.Scheme, req.URL.Host)
		fmt.Printf("[WARNING]: url:%s \n 请求耗时 %v 超过1秒\n ", sanitizedURL, duration)
	}

	if err != nil {
		return
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			// 这里可能有问题
			err = closeErr
			return
		}
	}()

	body, err = io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")

	return
}

func (c *httpClient) addHeaders(req *http.Request, headers map[string]string) {
	if len(headers) == 0 {
		return
	}

	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
}
