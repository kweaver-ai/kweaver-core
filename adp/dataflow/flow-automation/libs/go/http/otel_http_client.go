package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:generate mockgen -package mock_httpclient -source otel_http_client.go -destination ../mock/mock_httpclient/otel_http_client_mock.go

// HTTPClient HTTP客户端服务接口
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (respCode int, respParam interface{}, err error)
	Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Delete(ctx context.Context, url string, headers map[string]string) (respParam interface{}, err error)
	Head(ctx context.Context, url string, headers map[string]string) (respCode int, header *http.Header, err error)
	Request(ctx context.Context, url, method string, headers map[string]string, reqParam *[]byte) (respCode int, body []byte, err error)
	OSSClient(ctx context.Context, url, method string, headers map[string]string, reqParam *[]byte) (respHeaders http.Header, body []byte, err error)
}

var (
	otelRawOnce   sync.Once
	otelRawClient *http.Client
	otelHttpOnce  sync.Once
	otelClient    HTTPClient
)

// httpClient HTTP客户端结构
type httpClient struct {
	client *http.Client
}

// NewOtelHttpClient 创建带trace的http请求示例
func NewOtelRawHttpClient() *http.Client {
	otelRawOnce.Do(func() {
		otelRawClient = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: otelhttp.NewTransport(
				&http.Transport{
					TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
					MaxIdleConnsPerHost:   100,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ResponseHeaderTimeout: 30 * 60 * time.Second,
				},
			),
			Timeout: 30 * 60 * time.Second,
		}
	})
	return otelRawClient
}

// NewHTTPClient 创建HTTP客户端对象
func NewOtelHttpClient() HTTPClient {
	otelHttpOnce.Do(func() {
		otelClient = &httpClient{
			client: NewOtelRawHttpClient(),
		}
	})

	return otelClient
}

// Get http client request
func (c *httpClient) Request(ctx context.Context, url, method string, headers map[string]string, reqParam *[]byte) (respCode int, body []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(*reqParam))
	if err != nil {
		return
	}
	c.addHeaders(req, headers)
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.NewLogger().Errorln(closeErr)
		}
	}()
	body, err = io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return
	}

	return
}

// Get http client get
func (c *httpClient) Get(ctx context.Context, url string, headers map[string]string) (respCode int, respParam interface{}, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(ctx, req, headers)
	return
}

// Post http client post
func (c *httpClient) Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	var reqBody []byte
	if v, ok := reqParam.([]byte); ok {
		reqBody = v
	} else {
		reqBody, err = jsoniter.Marshal(reqParam)
		if err != nil {
			return
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(ctx, req, headers)
	return
}

// Put http client put
func (c *httpClient) Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	reqBody, err := jsoniter.Marshal(reqParam)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(ctx, req, headers)
	return
}

// Delete http client delete
func (c *httpClient) Delete(ctx context.Context, url string, headers map[string]string) (respParam interface{}, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return
	}

	_, respParam, err = c.httpDo(ctx, req, headers)
	return
}

// Get http client request
func (c *httpClient) OSSClient(ctx context.Context, url, method string, headers map[string]string, reqParam *[]byte) (respHeaders http.Header, body []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(*reqParam))
	if err != nil {
		return
	}
	c.addHeaders(req, headers)
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.NewLogger().Errorln(closeErr)
		}
	}()
	body, err = io.ReadAll(resp.Body)
	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return
	}
	respHeaders = resp.Header
	return
}

func (c *httpClient) httpDo(ctx context.Context, req *http.Request, headers map[string]string) (respCode int, respParam interface{}, err error) {
	if c.client == nil {
		return 0, nil, fmt.Errorf("http client is unavailable")
	}

	c.addHeaders(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.NewLogger().Errorln(closeErr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = errors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return
	}

	if len(body) != 0 {
		err = jsoniter.Unmarshal(body, &respParam)
	}

	if err != nil && strings.Contains(err.Error(), "Read: unexpected value") {
		err = nil
		respParam = body
	}

	return
}

// Head http client request
func (c *httpClient) Head(ctx context.Context, url string, headers map[string]string) (respCode int, header *http.Header, err error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return
	}
	c.addHeaders(req, headers)
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.NewLogger().Errorln(closeErr)
		}
	}()

	header = &resp.Header

	return
}

func (c *httpClient) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
}

// ExHTTPErrorParser parse http error
func ExHTTPErrorParser(err error) (map[string]interface{}, error) {
	httpError, ok := err.(errors.ExHTTPError)
	var httpErrorBody map[string]interface{}

	if !ok {
		return httpErrorBody, err
	}

	parseErr := json.Unmarshal([]byte(httpError.Body), &httpErrorBody)

	if parseErr != nil {
		return httpErrorBody, err
	}

	return httpErrorBody, nil
}
