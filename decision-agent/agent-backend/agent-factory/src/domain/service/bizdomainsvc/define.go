package bizdomainsvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
)

type BizDomainSvc struct {
	*service.SvcBase
	logger        icmp.Logger
	bizDomainHttp ibizdomainacc.BizDomainHttpAcc
}

type NewBizDomainSvcDto struct {
	SvcBase       *service.SvcBase
	Logger        icmp.Logger
	BizDomainHttp ibizdomainacc.BizDomainHttpAcc
}

func NewBizDomainService(dto *NewBizDomainSvcDto) *BizDomainSvc {
	impl := &BizDomainSvc{
		SvcBase:       dto.SvcBase,
		logger:        dto.Logger,
		bizDomainHttp: dto.BizDomainHttp,
	}

	return impl
}
