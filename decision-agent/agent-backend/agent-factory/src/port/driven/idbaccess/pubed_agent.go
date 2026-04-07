package idbaccess

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/pubed_agent.go github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess IPubedAgentRepo
type IPubedAgentRepo interface {
	IDBAccBaseRepo

	GetPubedList(ctx context.Context, req *pubedreq.PubedAgentListReq) (rt []*dapo.PublishedJoinPo, err error)

	GetPubedListByXx(ctx context.Context, arg *padbarg.GetPaPoListByXxArg) (ret *padbret.GetPaPoListByXxRet, err error)

	GetPubedPoMapByXx(ctx context.Context, arg *padbarg.GetPaPoListByXxArg) (ret *padbret.GetPaPoMapByXxRet, err error)
}
