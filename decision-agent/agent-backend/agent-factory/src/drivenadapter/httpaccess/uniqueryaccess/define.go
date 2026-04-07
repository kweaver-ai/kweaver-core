package uniqueryaccess

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type uniqueryHttpAcc struct {
	logger         icmp.Logger
	client         rest.HTTPClient
	uniqueryConf   *conf.UniqueryConf
	privateAddress string
}

var _ iuniqueryhttp.IUniquery = &uniqueryHttpAcc{}

func NewUniqueryHttpAcc(logger icmp.Logger, uniqueryConf *conf.UniqueryConf, client rest.HTTPClient) iuniqueryhttp.IUniquery {
	impl := &uniqueryHttpAcc{
		logger:         logger,
		client:         client,
		uniqueryConf:   uniqueryConf,
		privateAddress: cutil.GetHTTPAccess(uniqueryConf.PrivateSvc.Host, uniqueryConf.PrivateSvc.Port, uniqueryConf.PrivateSvc.Protocol),
	}

	return impl
}
