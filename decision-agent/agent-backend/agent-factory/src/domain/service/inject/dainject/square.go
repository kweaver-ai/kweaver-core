package dainject

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/squaresvc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver"
)

var (
	squareSvcOnce sync.Once
	squareSvcImpl iv3portdriver.ISquareSvc
)

func NewSquareSvc() iv3portdriver.ISquareSvc {
	squareSvcOnce.Do(func() {
		squareSvcImpl = squaresvc.NewSquareService()
	})

	return squareSvcImpl
}
