package othersvc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

var (
	otherSvcOnce sync.Once
	otherSvcImpl iv3portdriver.IOtherSvc
)

type otherSvc struct {
	*service.SvcBase
	agentConfRepo idbaccess.IDataAgentConfigRepo
}

type NewOtherSvcDto struct {
	SvcBase       *service.SvcBase
	AgentConfRepo idbaccess.IDataAgentConfigRepo
}

var _ iv3portdriver.IOtherSvc = &otherSvc{}

func NewOtherService(dto *NewOtherSvcDto) iv3portdriver.IOtherSvc {
	otherSvcOnce.Do(func() {
		otherSvcImpl = &otherSvc{
			SvcBase:       dto.SvcBase,
			agentConfRepo: dto.AgentConfRepo,
		}
	})

	return otherSvcImpl
}
