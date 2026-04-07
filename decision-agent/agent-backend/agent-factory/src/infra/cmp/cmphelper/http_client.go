package cmphelper

import (
	"net/http"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"

	"github.com/gogf/gf/v2/frame/g"
)

func GetClientWithTimeout(timeout time.Duration, opts ...httphelper.Option) (c icmp.IHttpClient) {
	tran := httphelper.GetDefaultTp()

	cutil.SetTpTlsInsecureSkipVerify(tran)

	client := &http.Client{
		Transport: tran,
		Timeout:   timeout,
	}

	gClient := g.Client()
	gClient.Client = *client
	opt := httphelper.WithClient(gClient)

	// 注意这个顺序，先设置client，再设置其他option
	opts = append([]httphelper.Option{opt}, opts...)

	c = httphelper.NewHTTPClient(opts...)

	return
}

func GetClient(opts ...httphelper.Option) (c icmp.IHttpClient) {
	c = GetClientWithTimeout(httphelper.DefaultTimeout, opts...)

	return
}
