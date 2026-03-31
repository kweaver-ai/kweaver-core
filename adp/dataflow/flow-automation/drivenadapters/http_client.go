package drivenadapters

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	_errors "errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
)

//go:generate mockgen -package mock_httpclient -source ../drivenadapters/http_client.go -destination ../tests/mock_httpclient/http_client_mock.go

// HTTPClient HTTP客户端服务接口
type HTTPClient interface {
	Get(url string, headers map[string]string) (respParam interface{}, err error)
	Post(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Put(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Delete(url string, headers map[string]string) (respParam interface{}, err error)
	Request(url, method string, headers map[string]string, reqParam *[]byte) (respCode int, body []byte, err error)
}

var (
	rawOnce                 sync.Once
	rawClient               *http.Client
	httpOnce                sync.Once
	otelHttpOnce            sync.Once
	otelHttpClient          otelHttp.HTTPClient
	otelRawHttpOnce         sync.Once
	otelRawHttpClient       *http.Client
	otelOauthHttpClientOnce sync.Once
	otelOauthHttpClient     otelHttp.OAuth2Client
	client                  HTTPClient
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
				TLSHandshakeTimeout:   30 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: 5 * 60 * time.Second,
			},
			Timeout: 5 * 60 * time.Second,
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

// NewOtelHTTPClient 创建otelhttp客户端
func NewOtelHTTPClient() otelHttp.HTTPClient {
	otelHttpOnce.Do(func() {
		otelHttpClient = otelHttp.NewOtelHttpClient()
	})

	return otelHttpClient
}

// NewOtelRawHTTPClient 创建otelhttp客户端
func NewOtelRawHTTPClient() *http.Client {
	otelRawHttpOnce.Do(func() {
		otelRawHttpClient = otelHttp.NewOtelRawHttpClient()
	})

	return otelRawHttpClient
}

// NewOauthOtelHTTPClient 创建otelhttp客户端
func NewOauthOtelHTTPClient() otelHttp.OAuth2Client {
	otelOauthHttpClientOnce.Do(func() {
		config := common.NewConfig()
		otelOauthHttpClient = otelHttp.NewOAuth2Client(otelHttp.OAuth2ClientConfig{
			ClientID:     config.OAuth.ClientID,
			ClientSecret: config.OAuth.ClientSecret,
			Scopes:       []string{"all"},
			TokenURL:     fmt.Sprintf("http://%s:%v/oauth2/token", config.OAuth.PublicHost, config.OAuth.PublicPort),
		})
	})

	return otelOauthHttpClient
}

// Get http client request
func (c *httpClient) Request(url, method string, headers map[string]string, reqParam *[]byte) (respCode int, body []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(*reqParam))
	c.addHeaders(req, headers)
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			commonLog.NewLogger().Errorln(closeErr)
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
func (c *httpClient) Get(url string, headers map[string]string) (respParam interface{}, err error) {
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return
	}

	_, respParam, err = c.httpDo(req, headers)
	return
}

// Post http client post
func (c *httpClient) Post(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
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

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Put http client put
func (c *httpClient) Put(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	reqBody, err := jsoniter.Marshal(reqParam)
	if err != nil {
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Delete http client delete
func (c *httpClient) Delete(url string, headers map[string]string) (respParam interface{}, err error) {
	req, err := http.NewRequest("DELETE", url, http.NoBody)
	if err != nil {
		return
	}

	_, respParam, err = c.httpDo(req, headers)
	return
}

func (c *httpClient) httpDo(req *http.Request, headers map[string]string) (respCode int, respParam interface{}, err error) {
	if c.client == nil {
		return 0, nil, _errors.New("http client is unavailable")
	}

	c.addHeaders(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			commonLog.NewLogger().Errorln(closeErr)
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
