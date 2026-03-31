package dolphintpleo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

type DolphinTplEo struct {
	Key   cdaenum.DolphinTplKey `json:"key"`
	Name  string                `json:"name"`
	Value string                `json:"value"`
}
