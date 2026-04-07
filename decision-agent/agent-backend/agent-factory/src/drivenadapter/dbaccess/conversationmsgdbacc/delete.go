package conversationmsgdbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"go.opentelemetry.io/otel/attribute"
)

// Delete implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) Delete(ctx context.Context, id string) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("msgID", id))

	po := &dapo.ConversationMsgPO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", id).Delete()

	return
}

// DeleteByConversationID implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) DeleteByConversationID(ctx context.Context, tx *sql.Tx, conversationID string) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetConversationID(ctx, conversationID)

	po := &dapo.ConversationMsgPO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	if tx != nil {
		dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_conversation_id", conversationID).Update(map[string]interface{}{"f_is_deleted": 1})

	return
}

// DeleteByAPPKey implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) DeleteByAPPKey(ctx context.Context, tx *sql.Tx, appKey string) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("appKey", appKey))

	po := &dapo.ConversationMsgPO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	if tx != nil {
		dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_agent_app_key", appKey).Update(map[string]interface{}{"f_is_deleted": 1})

	return
}
