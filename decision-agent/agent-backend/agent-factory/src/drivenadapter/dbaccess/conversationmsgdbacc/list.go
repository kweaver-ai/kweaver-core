package conversationmsgdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation_message/conversationmsgreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/pkg/errors"
)

// List implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) List(ctx context.Context, req conversationmsgreq.ListReq) (rt []*dapo.ConversationMsgPO, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, nil)
	oteltrace.SetConversationID(ctx, req.ConversationID)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	po := &dapo.ConversationMsgPO{}
	sr.FromPo(po)

	sr.WhereEqual("f_conversation_id", req.ConversationID).WhereEqual("f_is_deleted", 0)

	poList := make([]dapo.ConversationMsgPO, 0)

	sr.Order("f_index ASC")

	err = sr.Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get conversation message list")
		return
	}

	rt = cutil.SliceToPtrSlice(poList)

	return
}
