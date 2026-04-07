package dainject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/othersvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/daconfdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

var (
	otherSvcOnce sync.Once
	otherSvcImpl iv3portdriver.IOtherSvc
)

// NewOtherSvc 创建 Other 服务实例
func NewOtherSvc() iv3portdriver.IOtherSvc {
	otherSvcOnce.Do(func() {
		dto := &othersvc.NewOtherSvcDto{
			SvcBase:       service.NewSvcBase(),
			AgentConfRepo: daconfdbacc.NewDataAgentRepo(),
		}
		otherSvcImpl = othersvc.NewOtherService(dto)
	})

	return otherSvcImpl
}
