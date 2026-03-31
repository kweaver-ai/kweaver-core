// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bytedance/sonic"

	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/telemetry"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

// httpClient HTTP客户端结构
type httpClient struct {
	client *http.Client
	logger interfaces.Logger
}

// HTTPClientOptions 配置信息
type HTTPClientOptions struct {
	TimeOut int
}

// NewRawHTTPClient 创建原生HTTP客户端对象
func NewRawHTTPClient() *http.Client {
	opts := HTTPClientOptions{
		TimeOut: 600, //nolint:mnd
	}
	return NewRawHTTPClientWithOptions(opts)
}

// NewHTTPClientWithOptions 根据配置创建HTTP客户端对象
func NewHTTPClientWithOptions(opts HTTPClientOptions) interfaces.HTTPClient {
	client := &httpClient{
		client: NewRawHTTPClientWithOptions(opts),
		logger: logger.DefaultLogger(),
	}

	return client
}

// NewRawHTTPClientWithOptions 根据配置创建原生HTTP客户端对象
func NewRawHTTPClientWithOptions(opts HTTPClientOptions) *http.Client {
	rawClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost:   100,              //nolint:mnd
			MaxIdleConns:          100,              //nolint:mnd
			IdleConnTimeout:       90 * time.Second, //nolint:mnd
			TLSHandshakeTimeout:   10 * time.Second, //nolint:mnd
			ExpectContinueTimeout: 30 * time.Second, //nolint:mnd
			DisableKeepAlives:     false,
		},
		Timeout: time.Duration(opts.TimeOut) * time.Second,
	}

	return rawClient
}

func NewHTTPClientWithRawClient(rawClient *http.Client) *httpClient {
	client := &httpClient{
		client: rawClient,
	}

	return client
}

// NewHTTPClient 创建HTTP客户端对象
func NewHTTPClient() interfaces.HTTPClient {
	client := &httpClient{
		client: NewRawHTTPClient(),
		logger: logger.DefaultLogger(),
	}

	return client
}

// Get, 返回序列化对象
func (c *httpClient) Get(ctx context.Context, rawURL string, queryValues url.Values, headers map[string]string) (respCode int, respData interface{}, err error) {
	url, err := c.generateURL(rawURL, queryValues)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}

	return c.httpDo(ctx, http.MethodGet, url.String(), headers, nil)
}

// Get, 返回text
func (c *httpClient) GetNoUnmarshal(ctx context.Context, rawURL string, queryValues url.Values, headers map[string]string) (respCode int, respBody []byte, err error) {
	url, err := c.generateURL(rawURL, queryValues)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}

	return c.httpDoNoUnmarshal(ctx, http.MethodGet, url.String(), headers, nil)
}

// Post, 传入序列化对象，返回序列化对象
func (c *httpClient) Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodPost, url, headers, reqParam)
}

// Post, 传入序列化对象，返回text
func (c *httpClient) PostNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodPost, url, headers, reqParam)
}

// Put, 传入序列化对象，返回序列化对象
func (c *httpClient) Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodPut, url, headers, reqParam)
}

// Put, 传入序列化对象，返回text
func (c *httpClient) PutNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodPut, url, headers, reqParam)
}

// Delete, 返回序列化对象
func (c *httpClient) Delete(ctx context.Context, url string, headers map[string]string) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodDelete, url, headers, nil)
}

// Delete, 传入序列化对象，返回text
func (c *httpClient) DeleteNoUnmarshal(ctx context.Context, url string, headers map[string]string) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodDelete, url, headers, nil)
}

// Patch, 传入序列化对象，返回序列化对象
func (c *httpClient) Patch(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodPatch, url, headers, reqParam)
}

// Patch, 传入序列化对象，返回text
func (c *httpClient) PatchNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodPatch, url, headers, reqParam)
}

// 反序列化返回内容
func (c *httpClient) httpDo(ctx context.Context, mtehod, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	respCode, respBody, err := c.httpDoNoUnmarshal(ctx, mtehod, url, headers, reqParam)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}
	if len(respBody) == 0 {
		return
	}
	err = sonic.Unmarshal(respBody, &respData)
	if err != nil {
		c.logger.Error(err.Error())
		respData = string(respBody)
	}
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		respStr := utils.ObjectToJSON(respData)
		c.logger.Errorf("Exception(http do error, method: %s, url: %s, headers: %v, reqParam: %v, respCode: %d, error: %s)",
			mtehod, url, utils.ObjectToJSON(headers), utils.ObjectToJSON(reqParam), respCode, respStr)
		// 调用外部服务异常
		err = infraErr.NewHTTPError(ctx, respCode, infraErr.ErrExtCommonExternalServerError,
			fmt.Sprintf("Exception(http do error, method: %s, url: %s,  http status: %d, error: %s)", mtehod, url, respCode, respStr))
		return
	}
	return
}

// 返回原始respBody, 不进行反序列化
func (c *httpClient) httpDoNoUnmarshal(ctx context.Context, mtehod, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	if c.client == nil {
		return 0, nil, errors.New("http client is unavailable")
	}

	req, err := c.generateReq(ctx, mtehod, url, headers, reqParam)
	if err != nil {
		c.logger.Error(err.Error())
		return 0, nil, err
	}

	resp, err := telemetry.HTTPRequest(ctx, req, func(req *http.Request) (rsp *http.Response, err error) {
		return c.client.Do(req)
	})
	if err != nil {
		c.logger.Error(err.Error())
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			c.logger.Error(closeErr.Error())
		}
	}()
	respBody, err = io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	return
}

func (c *httpClient) generateURL(rawURL string, queryValues url.Values) (*url.URL, error) {
	uri, err := url.Parse(rawURL)
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	if queryValues != nil {
		values := uri.Query()
		for k, v := range values {
			queryValues[k] = v
		}
		uri.RawQuery = queryValues.Encode()
	}

	return uri, err
}

func (c *httpClient) generateReq(ctx context.Context, httpMethod, url string,
	headers map[string]string, reqParam interface{}) (req *http.Request, err error) {
	if reqParam != nil {
		var reader *bytes.Reader
		if v, ok := reqParam.([]byte); ok {
			reader = bytes.NewReader(v)
		} else {
			reqData, err := sonic.Marshal(reqParam)
			if err != nil {
				c.logger.Error(err.Error())
				return nil, err
			}
			reader = bytes.NewReader(reqData)
		}
		req, err = http.NewRequestWithContext(ctx, httpMethod, url, reader)
	} else {
		req, err = http.NewRequestWithContext(ctx, httpMethod, url, http.NoBody)
	}

	if err != nil {
		c.logger.Error(err.Error())
		return
	}

	for k, v := range headers {
		if v != "" {
			req.Header.Add(k, v)
		}
	}
	return
}
