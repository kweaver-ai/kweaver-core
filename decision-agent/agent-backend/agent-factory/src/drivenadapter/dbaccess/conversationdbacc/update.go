package conversationdbacc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

// Update implements idbaccess.IConversationRepo.
func (repo *ConversationRepo) Update(ctx context.Context, po *dapo.ConversationPO) (err error) {
	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	oteltrace.SetConversationID(ctx, po.ID)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", po.ID).WhereEqual("f_is_deleted", 0).
		SetUpdateFields([]string{
			"f_title",
			"f_message_index",
			"f_read_message_index",
			"f_ext",
			"f_update_time",
			"f_update_by",
		}).
		UpdateByStruct(po)

	return
}
