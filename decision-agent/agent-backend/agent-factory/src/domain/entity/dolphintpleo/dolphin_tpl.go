package dolphintpleo

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
)

type DolphinTplEo struct {
	Key   cdaenum.DolphinTplKey `json:"key"`
	Name  string                `json:"name"`
	Value string                `json:"value"`
}
