package padbret

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"

type GetPaPoListByXxRet struct {
	JoinPos []*dapo.PublishedJoinPo
}

func NewGetPaPoListByXxRet() *GetPaPoListByXxRet {
	return &GetPaPoListByXxRet{
		JoinPos: make([]*dapo.PublishedJoinPo, 0),
	}
}
