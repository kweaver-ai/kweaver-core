package daconfdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetByIDsAndCreatedBy 根据ID列表和创建者获取agent
func (repo *DAConfigRepo) GetByIDsAndCreatedBy(ctx context.Context, ids []string, createdBy string) (res []*dapo.DataAgentPo, err error) {
	if len(ids) == 0 {
		return make([]*dapo.DataAgentPo, 0), nil
	}

	// 去重ID列表
	ids = cutil.DeduplGeneric(ids)

	po := &dapo.DataAgentPo{}
	poList := make([]dapo.DataAgentPo, 0)
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_created_by", createdBy).
		In("f_id", ids).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "[DAConfigRepo][GetByIDsAndCreatedBy]: get by ids %v and created_by %s", ids, createdBy)
	}

	res = cutil.SliceToPtrSlice(poList)

	return
}
