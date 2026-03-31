package chttpinject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	bizDomainOnce sync.Once
	bizDomainImpl ibizdomainacc.BizDomainHttpAcc
)

func NewBizDomainHttpAcc() ibizdomainacc.BizDomainHttpAcc {
	bizDomainOnce.Do(func() {
		if global.GConfig.SwitchFields.Mock.MockBizDomain {
			bizDomainImpl = bizdomainhttp.NewMockBizDomainHttpAcc(
				logger.GetLogger(),
			)
		} else {
			bizDomainImpl = bizdomainhttp.NewBizDomainHttpAcc(
				logger.GetLogger(),
			)
		}
	})

	return bizDomainImpl
}
