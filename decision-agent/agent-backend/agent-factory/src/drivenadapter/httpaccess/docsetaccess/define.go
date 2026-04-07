package docsetaccess

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/idocsethttp"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type docsetHttpAcc struct {
	logger         icmp.Logger
	client         rest.HTTPClient
	docsetConf     *conf.DocsetConf
	privateAddress string
}

var _ idocsethttp.IDocset = &docsetHttpAcc{}

func NewDocsetHttpAcc(logger icmp.Logger, docsetConf *conf.DocsetConf, httpClient rest.HTTPClient) idocsethttp.IDocset {
	impl := &docsetHttpAcc{
		logger:         logger,
		client:         httpClient,
		docsetConf:     docsetConf,
		privateAddress: cutil.GetHTTPAccess(docsetConf.PrivateSvc.Host, docsetConf.PrivateSvc.Port, docsetConf.PrivateSvc.Protocol),
	}

	return impl
}
