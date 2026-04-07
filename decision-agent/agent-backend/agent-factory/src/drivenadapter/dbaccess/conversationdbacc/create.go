package conversationdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *ConversationRepo) Create(ctx context.Context, po *dapo.ConversationPO) (rt *dapo.ConversationPO, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()

	po.ID = cutil.UlidMake()
	oteltrace.SetConversationID(ctx, po.ID)
	po.CreateTime = cutil.GetCurrentMSTimestamp()
	po.UpdateTime = po.CreateTime
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	sr.FromPo(po)
	_, err = sr.InsertStruct(po)

	return po, err
}
