package pubedagentdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
)

func (repo *pubedAgentRepo) GetPubedPoMapByXx(ctx context.Context, arg *padbarg.GetPaPoListByXxArg) (ret *padbret.GetPaPoMapByXxRet, err error) {
	ret = padbret.NewGetPaPoMapByXxRet()

	listRet, err := repo.GetPubedListByXx(ctx, arg)
	if err != nil {
		return
	}

	for _, joinPo := range listRet.JoinPos {
		ret.JoinPosID2PoMap[joinPo.ID] = joinPo
		ret.JoinPosKey2PoMap[joinPo.Key] = joinPo
	}

	return
}
