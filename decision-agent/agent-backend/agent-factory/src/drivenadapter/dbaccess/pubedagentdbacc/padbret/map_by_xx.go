package padbret

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"

type GetPaPoMapByXxRet struct {
	JoinPosID2PoMap  map[string]*dapo.PublishedJoinPo
	JoinPosKey2PoMap map[string]*dapo.PublishedJoinPo
}

func NewGetPaPoMapByXxRet() *GetPaPoMapByXxRet {
	return &GetPaPoMapByXxRet{
		JoinPosID2PoMap:  make(map[string]*dapo.PublishedJoinPo),
		JoinPosKey2PoMap: make(map[string]*dapo.PublishedJoinPo),
	}
}
