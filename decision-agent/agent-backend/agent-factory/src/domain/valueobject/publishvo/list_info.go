package publishvo

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"

type ListPublishInfo struct {
	dapo.PublishedToBeStruct
}

func NewListPublishInfo() *ListPublishInfo {
	return &ListPublishInfo{}
}
