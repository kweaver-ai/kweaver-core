package daconfdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *DAConfigRepo) GetByID(ctx context.Context, id string) (po *dapo.DataAgentPo, err error) {
	po = &dapo.DataAgentPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_id", id).
		FindOne(po)

	return
}

func (repo *DAConfigRepo) GetByKey(ctx context.Context, key string) (po *dapo.DataAgentPo, err error) {
	po = &dapo.DataAgentPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_key", key).
		FindOne(po)

	return
}

func (repo *DAConfigRepo) GetByKeys(ctx context.Context, keys []string) (pos []*dapo.DataAgentPo, err error) {
	po := &dapo.DataAgentPo{}
	poList := make([]dapo.DataAgentPo, 0)
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		In("f_key", keys).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by keys %v", keys)
	}

	pos = cutil.SliceToPtrSlice(poList)

	return
}

func (repo *DAConfigRepo) GetByIDS(ctx context.Context, ids []string) (res []*dapo.DataAgentPo, err error) {
	po := &dapo.DataAgentPo{}
	poList := make([]dapo.DataAgentPo, 0)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		In("f_id", ids).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by ids %v", ids)
	}

	res = cutil.SliceToPtrSlice(poList)

	return
}

func (repo *DAConfigRepo) GetMapByIDs(ctx context.Context, ids []string) (res map[string]*dapo.DataAgentPo, err error) {
	res = make(map[string]*dapo.DataAgentPo)

	if len(ids) == 0 {
		return
	}

	pos, err := repo.GetByIDS(ctx, ids)
	if err != nil {
		err = errors.Wrap(err, "[DAConfigRepo][GetMapByIDs] error")
		return
	}

	for _, po := range pos {
		res[po.ID] = po
	}

	return
}

// GetIDNameMapByID 根据id获取name map
func (repo *DAConfigRepo) GetIDNameMapByID(ctx context.Context, ids []string) (res map[string]string, err error) {
	res = make(map[string]string)

	if len(ids) == 0 {
		return
	}

	po := &dapo.DataAgentPo{}
	poList := make([]dapo.DataAgentPo, 0)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.Select([]string{"f_id", "f_name"}).
		WhereEqual("f_deleted_at", 0).
		In("f_id", ids).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by ids %v", ids)
	}

	for _, po := range poList {
		res[po.ID] = po.Name
	}

	return
}
