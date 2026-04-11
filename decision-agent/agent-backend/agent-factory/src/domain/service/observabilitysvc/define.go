package observabilitysvc

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver"
)

// NewObservabilitySvcDto 可观测性服务构造参数
type NewObservabilitySvcDto struct {
	Logger    icmp.Logger
	Uniquery  iuniqueryhttp.IUniquery
	SquareSvc iv3portdriver.ISquareSvc
}

type observabilitySvc struct {
	logger    icmp.Logger
	uniquery  iuniqueryhttp.IUniquery
	squareSvc iv3portdriver.ISquareSvc
}

var _ iportdriver.IObservability = &observabilitySvc{}

func NewObservabilitySvc(dto *NewObservabilitySvcDto) iportdriver.IObservability {
	return &observabilitySvc{
		logger:    dto.Logger,
		uniquery:  dto.Uniquery,
		squareSvc: dto.SquareSvc,
	}
}
