package httphelper

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

func (c *httpClient) Put(ctx context.Context, url string, data interface{}) (resp *gclient.Response, err error) {
	debugReqLog(debugReqLogger{
		URL:    url,
		Data:   data,
		Method: http.MethodPut,
	})

	return c.client.Retry(3, time.Second*RetryInterval).Put(ctx, url, data)
}

func (c *httpClient) PutJSONExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error) {
	c.setContentType(cenum.HTTPHctJSON)
	return c.PutExpect2xx(ctx, url, data)
}

func (c *httpClient) PutExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error) {
	var (
		r          *gclient.Response
		requestErr error
	)

	startTime := time.Now()

	// ctx, span := c.arTrace.AddClientTrace(ctx)
	defer func() {
		// c.arTraceRecord(span, r, err, requestErr)
		if requestErr != nil {
			err = requestErr
		}
	}()

	debugReqLog(debugReqLogger{
		URL:    url,
		Data:   data,
		Method: http.MethodPut,
	})

	r, requestErr = c.client.Retry(3, time.Second*RetryInterval).Put(ctx, url, data)

	if requestErr != nil {
		// todo logger 从外面传进来
		log.Printf("[PutExpect2xx] request error: %v\n", requestErr)
		return
	}

	defer func(r *gclient.Response) {
		_ = r.Close()
	}(r)

	err = c.errExpect2xx(r)
	if err != nil {
		_resp := r.ReadAllString()
		// 记录请求日志（错误情况）
		logGClientRequest(ctx, http.MethodPut, url, data, r, []byte(_resp), startTime)

		return
	}

	resp = r.ReadAllString()

	// 记录请求日志
	logGClientRequest(ctx, http.MethodPut, url, data, r, []byte(resp), startTime)

	return
}
