package conversationdbacc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"go.opentelemetry.io/otel/attribute"

	"github.com/pkg/errors"
)

// List implements idbaccess.IConversationRepo.
func (repo *ConversationRepo) List(ctx context.Context, req conversationreq.ListReq) (rt []*dapo.ConversationPO, count int64, err error) {
	_, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("agentAPPKey", req.AgentAPPKey))
	span.SetAttributes(attribute.String("userId", req.UserId))

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	srCount := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	po := &dapo.ConversationPO{}
	sr.FromPo(po)
	srCount.FromPo(po)

	sr.WhereEqual("f_agent_app_key", req.AgentAPPKey).WhereEqual("f_create_by", req.UserId).WhereEqual("f_is_deleted", 0)
	srCount.WhereEqual("f_agent_app_key", req.AgentAPPKey).WhereEqual("f_create_by", req.UserId).WhereEqual("f_is_deleted", 0)

	if req.Title != "" {
		sr.Like("f_titile", req.Title)
		srCount.Like("f_titile", req.Title)
	}

	poList := make([]dapo.ConversationPO, 0)

	sr.Order("f_update_time DESC")
	sr.Offset((req.Page - 1) * req.Size)
	sr.Limit(req.Size)

	err = sr.Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get agent list")
		return
	}

	rt = cutil.SliceToPtrSlice(poList)

	count, err = srCount.Count()
	if err != nil {
		err = errors.Wrapf(err, "get agent count")
		return
	}

	return
}
