package daconftpldbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *DAConfigTplRepo) GetByCategoryID(ctx context.Context, categoryID string) (res []*dapo.DataAgentTplPo, err error) {
	po := &dapo.DataAgentTplPo{}
	poList := make([]dapo.DataAgentTplPo, 0)
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_category_id", categoryID).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by category id %s", categoryID)
	}

	res = cutil.SliceToPtrSlice(poList)

	return
}

func (repo *DAConfigTplRepo) GetMapByIDs(ctx context.Context, ids []int64) (res map[int64]*dapo.DataAgentTplPo, err error) {
	res = make(map[int64]*dapo.DataAgentTplPo)

	if len(ids) == 0 {
		return
	}

	pos, err := repo.GetByIDS(ctx, ids)
	if err != nil {
		err = errors.Wrap(err, "[DAConfigTplRepo][GetMapByIDs] error")
		return
	}

	for _, po := range pos {
		res[po.ID] = po
	}

	return
}
