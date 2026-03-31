package conversationdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/attribute"

	"github.com/pkg/errors"
)

// ListByAgentID implements idbaccess.IConversationRepo.
func (repo *ConversationRepo) ListByAgentID(ctx context.Context, agentID, title string, page, size int) (rt []*dapo.ConversationPO, count int64, err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, nil)
	o11y.SetAttributes(ctx, attribute.String("agentID", agentID))

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	po := &dapo.ConversationPO{}
	sr.FromPo(po)

	// 根据agentID过滤，并且只查询未删除的会话
	sr.WhereEqual("f_agent_app_key", agentID).WhereEqual("f_is_deleted", 0)
	// 只查询有消息的会话
	sr.WhereNotEqual("f_message_index", 0)

	// 如果title不为空，则进行模糊查询
	if title != "" {
		sr.Like("f_title", title)
	}

	count = 0
	// 按更新时间倒序排列
	sr.Order("f_update_time DESC")

	// 设置分页
	if page > 0 && size > 0 {
		offset := (page - 1) * size
		sr.Limit(size).Offset(offset)
	}

	poList := make([]dapo.ConversationPO, 0)

	err = sr.Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get conversation list by agentID: %s", agentID)
		return
	}

	rt = cutil.SliceToPtrSlice(poList)
	// 获取总数
	countSr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	countSr.FromPo(po)
	countSr.WhereEqual("f_agent_app_key", agentID).WhereEqual("f_is_deleted", 0)
	// 如果title不为空，则进行模糊查询
	if title != "" {
		countSr.Like("f_title", title)
	}

	countSr.WhereNotEqual("f_message_index", 0)

	count, err = countSr.Count()
	if err != nil {
		err = errors.Wrapf(err, "count conversation by agentID: %s", agentID)
		return
	}

	return
}
