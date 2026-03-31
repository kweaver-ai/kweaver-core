package httphelper

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

func (c *httpClient) Post(ctx context.Context, url string, data interface{}) (resp *gclient.Response, err error) {
	// ctx, span := c.arTrace.AddClientTrace(ctx)
	// defer func() {
	// 	c.arTraceRecord(span, resp, nil, err)
	// }()
	debugReqLog(debugReqLogger{
		URL:    url,
		Data:   data,
		Method: http.MethodPost,
	})

	resp, err = c.client.Retry(3, time.Second*RetryInterval).Post(ctx, url, data)

	return
}

func (c *httpClient) PostJSONExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error) {
	c.setContentType(cenum.HTTPHctJSON)
	resp, err = c.PostExpect2xx(ctx, url, data)

	return
}

func (c *httpClient) PostJSONExpect2xxByte(ctx context.Context, url string, data interface{}) (resp []byte, err error) {
	c.setContentType(cenum.HTTPHctJSON)
	resp, err = c.PostExpect2xxByte(ctx, url, data)

	return
}

func (c *httpClient) PostFormExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error) {
	c.setContentType(cenum.HTTPHctForm)
	resp, err = c.PostExpect2xx(ctx, url, data)

	return
}

func (c *httpClient) PostExpect2xx(ctx context.Context, url string, data interface{}) (resp string, err error) {
	respBytes, err := c.PostExpect2xxByte(ctx, url, data)
	if err != nil {
		return
	}

	resp = string(respBytes)

	return
}

func (c *httpClient) PostExpect2xxByte(ctx context.Context, url string, data interface{}) (resp []byte, err error) {
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
		Method: http.MethodPost,
	})

	r, requestErr = c.client.Retry(3, time.Second*RetryInterval).
		Post(ctx, url, data)

	if requestErr != nil {
		// todo logger 从外面传进来
		log.Printf("[PostExpect2xx] request error: %v\n", requestErr)
		return
	}

	defer func(r *gclient.Response) {
		_ = r.Close()
	}(r)

	err = c.errExpect2xx(r)
	if err != nil {
		_resp := r.ReadAll()
		// 记录请求日志（错误情况）
		logGClientRequest(ctx, http.MethodPost, url, data, r, _resp, startTime)

		return
	}

	resp = r.ReadAll()

	// 记录请求日志
	logGClientRequest(ctx, http.MethodPost, url, data, r, resp, startTime)

	return
}
