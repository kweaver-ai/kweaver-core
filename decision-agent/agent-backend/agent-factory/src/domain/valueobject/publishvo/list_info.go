package publishvo

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"

type ListPublishInfo struct {
	dapo.PublishedToBeStruct
}

func NewListPublishInfo() *ListPublishInfo {
	return &ListPublishInfo{}
}
