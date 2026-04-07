package conversationmsgdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// GetByID implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) GetByID(ctx context.Context, id string) (po *dapo.ConversationMsgPO, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("msgID", id))

	po = &dapo.ConversationMsgPO{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_id", id).FindOne(po)

	return
}

func (repo *ConversationMsgRepo) GetMaxIndexByID(ctx context.Context, id string) (maxIndex int, err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetConversationID(ctx, id)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	po := &dapo.ConversationMsgPO{}
	sr.FromPo(po)

	err = sr.WhereEqual("f_conversation_id", id).Order("f_index DESC").Limit(1).FindOne(po)
	if err != nil {
		return 0, errors.Wrapf(err, "get max index by id")
	}

	return po.Index, nil
}
