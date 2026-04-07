package httphelper

import (
	"context"
	"log"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
)

func (c *httpClient) Delete(ctx context.Context, url string) (resp *gclient.Response, err error) {
	return c.client.Retry(3, time.Second*RetryInterval).Delete(ctx, url)
}

func (c *httpClient) DeleteExpect2xx(ctx context.Context, url string, data ...interface{}) (resp string, err error) {
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

	r, requestErr = c.client.Retry(3, time.Second*RetryInterval).Delete(ctx, url, data...)

	if requestErr != nil {
		// todo logger 从外面传进来
		log.Printf("[DeleteExpect2xx] request error: %v\n", requestErr)
		return
	}

	defer func(r *gclient.Response) {
		_ = r.Close()
	}(r)

	err = c.errExpect2xx(r)
	if err != nil {
		_resp := r.ReadAllString()
		// 记录请求日志（错误情况）
		logGClientRequest(ctx, "DELETE", url, data, r, []byte(_resp), startTime)

		return
	}

	resp = r.ReadAllString()

	// 记录请求日志
	logGClientRequest(ctx, "DELETE", url, data, r, []byte(resp), startTime)

	return
}

func (c *httpClient) DeleteExpect2xxWithQueryParams(ctx context.Context, url string, queryData interface{}) (resp string, err error) {
	// 将查询参数作为URL的一部分传递
	url += "?" + c.buildQueryParams(queryData)

	resp, err = c.DeleteExpect2xx(ctx, url)

	return
}
