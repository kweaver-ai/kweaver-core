package conversationhistoryacc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *conversationHistoryRepo) GetLatestVisitAgentIds(ctx context.Context, userID string) (rt []*dapo.ConversationHistoryLatestVisitAgentPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	list := make([]dapo.ConversationHistoryLatestVisitAgentPO, 0)
	po := &dapo.ConversationHistoryLatestVisitAgentPO{}
	sr.FromPo(po)

	rawSql := "SELECT f_bot_id, MAX(f_modified_at) as last_modified_at FROM tb_conversation_history_v2 " +
		"WHERE f_user_id=? AND f_deleted=0 GROUP BY f_bot_id ORDER BY last_modified_at DESC"
	sr.Raw(rawSql, userID)

	err = sr.Find(&list)
	fmt.Println(list)

	if err != nil {
		return nil, errors.Wrapf(err, "get latest visit agent")
	}

	rt = cutil.SliceToPtrSlice(list)

	return rt, err
}
