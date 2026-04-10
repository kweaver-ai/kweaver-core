package conversationmsgdbacc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"go.opentelemetry.io/otel/attribute"

	"github.com/pkg/errors"
)

func (repo *ConversationMsgRepo) GetRecentMessages(ctx context.Context, conversationID string, limit int) (rt []*dapo.ConversationMsgPO, err error) {
	_, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("conversationID", conversationID))
	span.SetAttributes(attribute.Int("limit", limit))

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	po := &dapo.ConversationMsgPO{}
	sr.FromPo(po)

	sr.WhereEqual("f_conversation_id", conversationID).
		WhereEqual("f_is_deleted", 0).
		Order("f_index DESC").
		Limit(limit)

	poList := make([]dapo.ConversationMsgPO, 0)

	err = sr.Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get recent conversation messages")
		return
	}

	if len(poList) == 0 {
		return
	}

	reversePoList := make([]dapo.ConversationMsgPO, len(poList))
	for i, msg := range poList {
		reversePoList[len(poList)-1-i] = msg
	}

	rt = cutil.SliceToPtrSlice(reversePoList)

	return
}
