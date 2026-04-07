package conversationdbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// DeleteByAppKey implements idbaccess.IConversationRepo.
func (repo *ConversationRepo) Delete(ctx context.Context, tx *sql.Tx, id string) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetConversationID(ctx, id)

	po := &dapo.ConversationPO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", id).Update(map[string]interface{}{"f_is_deleted": 1})

	return
}

func (repo *ConversationRepo) DeleteByAPPKey(ctx context.Context, tx *sql.Tx, appKey string) (err error) {
	po := &dapo.ConversationPO{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_agent_app_key", appKey).Update(map[string]interface{}{"f_is_deleted": 1})

	return
}
