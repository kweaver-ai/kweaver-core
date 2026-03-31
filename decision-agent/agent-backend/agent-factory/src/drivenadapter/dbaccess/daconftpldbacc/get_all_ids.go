package daconftpldbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetAllIDs 获取所有未删除的agent模板ID列表
func (repo *DAConfigTplRepo) GetAllIDs(ctx context.Context) (ids []int64, err error) {
	po := &dapo.DataAgentTplPo{}
	poList := make([]dapo.DataAgentTplPo, 0)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.Select([]string{"f_id"}).
		WhereEqual("f_deleted_at", 0).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrap(err, "get all agent tpl ids")
	}

	ids = make([]int64, 0, len(poList))
	for _, p := range poList {
		ids = append(ids, p.ID)
	}

	return
}
