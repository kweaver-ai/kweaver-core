package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

//go:generate mockgen -package mock_httpclient -source oauth2_client.go -destination ../mock/mock_httpclient/oauth2_client_mock.go

type OAuth2Client interface {
	// Get(url string) (resp *http.Response, err error)
	// Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
	// Do(*http.Request) (resp *http.Response, err error)
	Get(url string, headers map[string]string) (respCode int, respParam interface{}, err error)
	Request(url, method string, headers map[string]string, reqParam *[]byte) (respCode int, body []byte, err error)
}

type oAuth2Client struct {
	client *http.Client
}

type OAuth2ClientConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
	TokenURL     string
}

var (
	oauthRawOnce   sync.Once
	oauthRawClient *http.Client
	oauthHttpOnce  sync.Once
	oauthClient    OAuth2Client
)

// NewOAuthRawHttpClient 初始化OAuth2Client的http.Client
func NewOAuthRawHttpClient(config OAuth2ClientConfig) *http.Client {
	oauthRawOnce.Do(func() {
		conf := &clientcredentials.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       config.Scopes,
			TokenURL:     config.TokenURL,
		}
		oauthRawClient = conf.Client(context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 10 * time.Second,
		}))
	})

	return oauthRawClient
}

func NewOAuth2Client(config OAuth2ClientConfig) OAuth2Client {
	oauthHttpOnce.Do(func() {
		oauthClient = &oAuth2Client{
			client: NewOAuthRawHttpClient(config),
		}
	})

	return oauthClient
}

// Get http client request
func (c *oAuth2Client) Request(url, method string, headers map[string]string, reqParam *[]byte) (respCode int, body []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(*reqParam))
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
func (c *oAuth2Client) Get(url string, headers map[string]string) (respCode int, respParam interface{}, err error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

func (c *oAuth2Client) httpDo(req *http.Request, headers map[string]string) (respCode int, respParam interface{}, err error) {
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

func (c *oAuth2Client) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
}
