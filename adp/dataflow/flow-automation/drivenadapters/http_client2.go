package drivenadapters

import (
	"context"
	"encoding/json"
	"strings"

	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
)

//go:generate mockgen -package mock_httpclient -source ../drivenadapters/http_client2.go -destination ../tests/mock_httpclient/http_client2_mock.go

type HTTPClient2 interface {
	Get(ctx context.Context, url string, headers map[string]string, respParam any) (respCode int, err error)
	Post(ctx context.Context, url string, headers map[string]string, reqParam any, respParam any) (respCode int, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqParam any, respParam any) (respCode int, err error)
	Delete(ctx context.Context, url string, headers map[string]string, respParam any) (respCode int, err error)
	Request(ctx context.Context, url, method string, headers map[string]string, reqParam *[]byte, respParam any) (respCode int, err error)
}

type httpClient2 struct {
	otelHttpClient otelHttp.HTTPClient
}

func NewHTTPClient2() HTTPClient2 {
	return &httpClient2{
		NewOtelHTTPClient(),
	}
}

func (c *httpClient2) Get(ctx context.Context, url string, headers map[string]string, respParam any) (respCode int, err error) {
	respCode, bytes, err := c.otelHttpClient.Request(ctx, url, "GET", headers, &[]byte{})
	if err != nil {
		return respCode, err
	}

	if respParam != nil {
		err = json.Unmarshal(bytes, respParam)
		if err != nil {
			return 0, err
		}
	}
	return respCode, err
}

func (c *httpClient2) Post(ctx context.Context, url string, headers map[string]string, reqParam any, respParam any) (respCode int, err error) {
	var reqBytes []byte
	if reqParam != nil {
		reqBytes, err = json.Marshal(reqParam)
		if err != nil {
			return 0, err
		}
	}

	respCode, bytes, err := c.otelHttpClient.Request(ctx, url, "POST", headers, &reqBytes)
	if err != nil {
		return respCode, err
	}

	if respParam != nil {
		err = json.Unmarshal(bytes, respParam)
		if err != nil {
			return 0, err
		}
	}
	return respCode, err
}

func (c *httpClient2) Put(ctx context.Context, url string, headers map[string]string, reqParam any, respParam any) (respCode int, err error) {
	var reqBytes []byte
	if reqParam != nil {
		reqBytes, err = json.Marshal(reqParam)
		if err != nil {
			return 0, err
		}
	}

	respCode, bytes, err := c.otelHttpClient.Request(ctx, url, "PUT", headers, &reqBytes)
	if err != nil {
		return respCode, err
	}

	if respParam != nil {
		err = json.Unmarshal(bytes, respParam)
		if err != nil {
			return 0, err
		}
	}
	return respCode, err
}

func (c *httpClient2) Delete(ctx context.Context, url string, headers map[string]string, respParam any) (respCode int, err error) {
	respCode, bytes, err := c.otelHttpClient.Request(ctx, url, "DELETE", headers, &[]byte{})
	if err != nil {
		return respCode, err
	}

	if respParam != nil {
		err = json.Unmarshal(bytes, respParam)
		if err != nil {
			return 0, err
		}
	}
	return respCode, err
}

func (c *httpClient2) Request(ctx context.Context, url, method string, headers map[string]string, reqParam *[]byte, respParam any) (respCode int, err error) {
	respCode, bytes, err := c.otelHttpClient.Request(ctx, url, method, headers, reqParam)
	if err != nil {
		return respCode, err
	}

	if respParam != nil {
		err = json.Unmarshal(bytes, respParam)
		if err != nil {
			return 0, err
		}
	}
	return respCode, err
}

func BearerToken(token string) string {
	if strings.ToLower(token[0:6]) == "bearer" {
		return token
	}

	return "Bearer " + token
}
