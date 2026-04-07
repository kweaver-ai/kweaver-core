package conversationmsgdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Create implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) Create(ctx context.Context, po *dapo.ConversationMsgPO) (id string, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetConversationID(ctx, po.ConversationID)
	po.ID = cutil.UlidMake()
	po.CreateTime = cutil.GetCurrentMSTimestamp()
	po.UpdateTime = po.CreateTime
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	sr.FromPo(po)
	_, err = sr.InsertStruct(po)

	return po.ID, err
}
