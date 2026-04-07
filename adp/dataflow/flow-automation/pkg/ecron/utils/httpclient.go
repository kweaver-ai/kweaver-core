package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils/rest"

	jsoniter "github.com/json-iterator/go"
)

//go:generate mockgen -package mock -source ../utils/httpclient.go -destination ../mock/mock_httpclient.go

// HTTPClient HTTP客户端服务接口
type HTTPClient interface {
	Get(url string, headers map[string]string, respParam interface{}) (err error)
	Post(url string, headers map[string]string, reqParam interface{}, respParam interface{}) (err error)
	PostV2(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Put(url string, headers map[string]string, reqParam interface{}, respParam interface{}) (err error)
	Delete(url string, headers map[string]string, respParam interface{}) (err error)
}

// NewHTTPClient 创建HTTP客户端服务句柄
func NewHTTPClient() HTTPClient {
	return &HTTPCli{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 60 * time.Second,
		},
	}
}

// HTTPCli HTTP客户端结构
type HTTPCli struct {
	client *http.Client
}

// Get http client get
func (c *HTTPCli) Get(url string, headers map[string]string, respParam interface{}) (err error) {
	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return
	}
	return c.httpDo(req, headers, &respParam)
}

// Post http client post
func (c *HTTPCli) Post(url string, headers map[string]string, reqParam interface{}, respParam interface{}) (err error) {
	reqBody := c.prepareBody(headers, reqParam)
	req, err := http.NewRequest("POST", url, reqBody)
	if nil != err {
		return
	}
	return c.httpDo(req, headers, &respParam)
}

// Post http client post
func (c *HTTPCli) PostV2(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	reqBody := c.prepareBody(headers, reqParam)

	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDoV2(req, headers)
	return
}

// Put http client put
func (c *HTTPCli) Put(url string, headers map[string]string, reqParam interface{}, respParam interface{}) (err error) {
	reqBody := c.prepareBody(headers, reqParam)
	req, err := http.NewRequest("PUT", url, reqBody)
	if nil != err {
		return
	}
	return c.httpDo(req, headers, &respParam)
}

// Delete http client delete
func (c *HTTPCli) Delete(url string, headers map[string]string, respParam interface{}) (err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if nil != err {
		return
	}
	return c.httpDo(req, headers, &respParam)
}

func (c *HTTPCli) httpDo(req *http.Request, headers map[string]string, respParam interface{}) (err error) {
	if nil == c.client {
		return errors.New(common.ErrHTTPClientUnavailable)
	}

	c.addHeaders(req, headers)

	resp, err := c.client.Do(req)
	if nil != err {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if (resp.StatusCode < 200) || (resp.StatusCode >= 300) {
		ecError := &common.ECronError{}
		jsonErr := jsoniter.Unmarshal(body, &ecError)
		if nil != jsonErr {
			ecError.Cause = string(body)
		}
		err = ecError
		return
	}

	if (nil != respParam) && 0 != len(body) {
		_ = jsoniter.Unmarshal(body, &respParam)
	}
	return
}

func (c *HTTPCli) httpDoV2(req *http.Request, headers map[string]string) (respCode int, respParam interface{}, err error) {
	if c.client == nil {
		return 0, nil, errors.New("http client is unavailable")
	}

	c.addHeaders(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			NewLogger().Errorln(closeErr)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = &rest.ExHTTPError{
			Body:   body,
			Status: respCode,
		}
		return
	}

	return respCode, body, err
}

func (c *HTTPCli) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
}

func (c *HTTPCli) prepareBody(headers map[string]string, reqParam interface{}) (body io.Reader) {
	contentType := "application/json"
	if nil != headers {
		if v, ok := headers["Content-Type"]; ok {
			contentType = v
		}
	}

	switch contentType {
	case "application/x-www-form-urlencoded":
		{
			req := reqParam.(map[string]interface{})
			if nil != req {
				reader := make([]string, 0)
				for k, v := range req {
					reader = append(reader, fmt.Sprintf("%v=%v", k, v))
				}
				return strings.NewReader(strings.Join(reader, "&"))
			}
		}
	}

	reqBody, err := jsoniter.Marshal(reqParam)
	if nil != err {
		return
	}
	return bytes.NewReader(reqBody)
}
