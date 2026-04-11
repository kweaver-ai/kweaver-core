package conversationmsgdbacc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"go.opentelemetry.io/otel/attribute"
)

// Update implements idbaccess.IConversationMsgRepo.
func (repo *ConversationMsgRepo) Update(ctx context.Context, po *dapo.ConversationMsgPO) (err error) {
	_, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("msgID", po.ID))

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", po.ID).
		SetUpdateFields([]string{
			"f_content",
			"f_content_type",
			"f_status",
			"f_ext",
			"f_update_time",
			"f_update_by",
		}).
		UpdateByStruct(po)

	return
}
