package httphelper

import (
	"context"
	"io"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func (c *httpClient) Do(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	// _, span := c.arTrace.AddClientTrace(ctx)
	defer func() {
		// c.arTraceRecord2(span, req, resp, nil, err)
	}()

	// 注入 Trace 上下文
	// observability.InjectTraceHeader(ctx, req.Header)
	resp, err = c.client.Client.Do(req)

	return
}

func (c *httpClient) DoExpect2xx(ctx context.Context, req *http.Request) (str string, err error) {
	resp, err := c.Do(ctx, req)
	if err != nil {
		return
	}

	err = c.errExpect2xxStd(resp)
	if err != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if cutil.IsHttpErr(resp) {
		debugResLog(debugResLogger{
			Err:      err,
			RespBody: body,
		})
	}

	str = string(body)

	return
}
