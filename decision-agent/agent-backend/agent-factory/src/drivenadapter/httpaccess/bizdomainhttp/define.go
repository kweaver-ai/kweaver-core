package bizdomainhttp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
)

type bizDomainHttpAcc struct {
	logger         icmp.Logger
	privateBaseURL string
}

var _ ibizdomainacc.BizDomainHttpAcc = &bizDomainHttpAcc{}

func NewBizDomainHttpAcc(
	logger icmp.Logger,
) ibizdomainacc.BizDomainHttpAcc {
	// 从配置中获取业务域服务的地址
	bizDomainConf := cglobal.GConfig.BizDomain.PrivateSvc

	privateBaseURL := cutil.GetHTTPAccess(bizDomainConf.Host, bizDomainConf.Port, bizDomainConf.Protocol)

	bizDomainImpl := &bizDomainHttpAcc{
		logger:         logger,
		privateBaseURL: privateBaseURL,
	}

	return bizDomainImpl
}
