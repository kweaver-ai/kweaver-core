package icmp

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/gclient"
)

// IHttpClient http client interface
//
//go:generate mockgen -destination=./cmpmock/httpClient_mock.go -package=cmpmock -source=./http_client.go
type IHttpClient interface {
	// -- get start --
	Get(ctx context.Context, url string, queryData ...interface{}) (resp *gclient.Response, err error)

	GetExpect2xx(ctx context.Context, url string, queryData ...interface{}) (resp string, err error)

	GetExpect2xxByte(ctx context.Context, url string, queryData ...interface{}) (resp []byte, err error)
	// -- get end --

	// -- post start --
	Post(ctx context.Context, url string, data interface{}) (resp *gclient.Response, err error)

	PostJSONExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error)

	PostFormExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error)

	PostExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error)

	PostJSONExpect2xxByte(ctx context.Context, url string, data interface{}) (resp []byte, err error)

	PostExpect2xxByte(ctx context.Context, url string, data interface{}) (resp []byte, err error)

	// -- post end --

	// -- put start --
	Put(ctx context.Context, url string, data interface{}) (resp *gclient.Response, err error)

	PutJSONExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error)

	PutExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error)

	// -- put end --

	// -- delete start --
	Delete(ctx context.Context, url string) (resp *gclient.Response, err error)

	DeleteExpect2xx(ctx context.Context, url string, data ...interface{}) (resp string, err error)

	DeleteExpect2xxWithQueryParams(ctx context.Context, url string, queryData interface{}) (resp string, err error)

	// -- delete end --

	Do(ctx context.Context, req *http.Request) (resp *http.Response, err error)

	DoExpect2xx(ctx context.Context, req *http.Request) (str string, err error)

	// other
	GetClient() *gclient.Client

	// Stream()
}
