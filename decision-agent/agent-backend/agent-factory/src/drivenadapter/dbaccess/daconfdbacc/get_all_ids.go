package daconfdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetAllIDs 获取所有未删除的agent ID列表
func (repo *DAConfigRepo) GetAllIDs(ctx context.Context) (ids []string, err error) {
	po := &dapo.DataAgentPo{}
	poList := make([]dapo.DataAgentPo, 0)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.Select([]string{"f_id"}).
		WhereEqual("f_deleted_at", 0).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrap(err, "get all agent ids")
	}

	ids = make([]string, 0, len(poList))
	for _, p := range poList {
		ids = append(ids, p.ID)
	}

	return
}
