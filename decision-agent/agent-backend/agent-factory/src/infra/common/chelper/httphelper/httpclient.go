package httphelper

import (
	"net/http"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/pkg/errors"
)

type Option func(c *httpClient)

type httpClient struct {
	token  string
	client *gclient.Client
	// arTrace api.Tracer
}

var _ icmp.IHttpClient = &httpClient{}

func NewHTTPClient(opts ...Option) icmp.IHttpClient {
	gClient := GetNewGClientWithDefaultStdClient()
	// 自动注入 Trace 处理
	gClient.Transport = otelhttp.NewTransport(gClient.Transport)
	c := &httpClient{
		token:  "",
		client: gClient,

		// arTrace: arTrace,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithToken(token string) Option {
	return func(c *httpClient) {
		if token == "" {
			return
		}

		// 如果以“Bearer ”开头,去除
		token = strings.TrimPrefix(token, "Bearer ")

		c.token = token
		c.client.SetHeader("Authorization", "Bearer "+token)
	}
}

func WithHeader(k, v string) Option {
	return func(c *httpClient) {
		c.client.SetHeader(k, v)
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(c *httpClient) {
		for k, v := range headers {
			c.client.SetHeader(k, v)
		}
	}
}

func WithClient(client *gclient.Client) Option {
	return func(c *httpClient) {
		c.client = client
	}
}

func (c *httpClient) errExpect2xx(r *gclient.Response) (err error) {
	//nolint:nestif
	if cutil.IsHttpErr(r.Response) {
		resp := &CommonRespError{}

		body := r.ReadAll()
		err = cutil.JSON().Unmarshal(body, &resp)

		if err != nil {
			prefixLen := 3000
			_url := r.Request.URL

			if len(body) > prefixLen {
				err = errors.Errorf("httpClient failed(not 2xx), url: [%s], http code: [%v], body[0:%d] is [%q]",
					_url, r.StatusCode, prefixLen, string(body)[:prefixLen])
			} else {
				err = errors.Errorf("httpClient failed(not 2xx), url: [%s], http code: [%v], body is [%q]", _url, r.StatusCode, string(body))
			}
		} else {
			err = errors.Wrap(resp, "httpClient failed(not 2xx), response")
		}

		debugResLog(debugResLogger{
			Err:      err,
			RespBody: body,
		})
	}

	return
}

func (c *httpClient) errExpect2xxStd(r *http.Response) (err error) {
	if cutil.IsHttpErr(r) {
		err = errors.Errorf("httpClient failed(not 2xx), url: [%s], http code: [%v]", r.Request.URL, r.StatusCode)
	}

	return
}

func (c *httpClient) setContentType(contentType string) {
	c.client.SetHeader(cenum.HTTPHct, contentType)
}

func (c *httpClient) GetClient() *gclient.Client {
	return c.client
}
