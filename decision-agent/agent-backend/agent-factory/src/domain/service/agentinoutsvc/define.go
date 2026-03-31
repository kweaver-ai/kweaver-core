package agentinoutsvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

type agentInOutSvc struct {
	*service.SvcBase
	logger         icmp.Logger
	agentConfRepo  idbaccess.IDataAgentConfigRepo
	pmsSvc         iv3portdriver.IPermissionSvc
	bizDomainHttp  ibizdomainacc.BizDomainHttpAcc
	bdAgentRelRepo idbaccess.IBizDomainAgentRelRepo
}

var _ iv3portdriver.IAgentInOutSvc = &agentInOutSvc{}

type NewAgentInOutSvcDto struct {
	SvcBase        *service.SvcBase
	Logger         icmp.Logger
	AgentConfRepo  idbaccess.IDataAgentConfigRepo
	PmsSvc         iv3portdriver.IPermissionSvc
	BizDomainHttp  ibizdomainacc.BizDomainHttpAcc
	BdAgentRelRepo idbaccess.IBizDomainAgentRelRepo
}

func NewAgentInOutService(dto *NewAgentInOutSvcDto) iv3portdriver.IAgentInOutSvc {
	impl := &agentInOutSvc{
		SvcBase:        dto.SvcBase,
		logger:         dto.Logger,
		agentConfRepo:  dto.AgentConfRepo,
		pmsSvc:         dto.PmsSvc,
		bizDomainHttp:  dto.BizDomainHttp,
		bdAgentRelRepo: dto.BdAgentRelRepo,
	}

	return impl
}
