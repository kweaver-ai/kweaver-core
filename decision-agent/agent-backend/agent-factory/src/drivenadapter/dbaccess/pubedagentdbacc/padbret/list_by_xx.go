package padbret

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"

type GetPaPoListByXxRet struct {
	JoinPos []*dapo.PublishedJoinPo
}

func NewGetPaPoListByXxRet() *GetPaPoListByXxRet {
	return &GetPaPoListByXxRet{
		JoinPos: make([]*dapo.PublishedJoinPo, 0),
	}
}
