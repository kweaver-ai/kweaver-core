package pubedeo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

type PublishedTplListEo struct {
	dapo.PublishedTplPo

	CreatedByName   string `json:"created_by_name"`
	UpdatedByName   string `json:"updated_by_name"`
	PublishedByName string `json:"published_by_name"`
}
