package daconftpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *DAConfigTplRepo) GetByID(ctx context.Context, id int64) (po *dapo.DataAgentTplPo, err error) {
	po = &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_id", id).
		FindOne(po)

	return
}

func (repo *DAConfigTplRepo) GetByKey(ctx context.Context, key string) (po *dapo.DataAgentTplPo, err error) {
	po = &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)
	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_key", key).
		FindOne(po)

	return
}

func (repo *DAConfigTplRepo) GetByKeys(ctx context.Context, keys []string) (res []*dapo.DataAgentTplPo, err error) {
	po := &dapo.DataAgentTplPo{}
	poList := make([]dapo.DataAgentTplPo, 0)
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		In("f_key", keys).
		Find(&poList)
	if err != nil {
		return nil, errors.Wrapf(err, "get by keys %v", keys)
	}

	res = cutil.SliceToPtrSlice(poList)

	return
}

func (repo *DAConfigTplRepo) GetByIDS(ctx context.Context, ids []int64) (res []*dapo.DataAgentTplPo, err error) {
	po := &dapo.DataAgentTplPo{}
	poList := make([]dapo.DataAgentTplPo, 0)

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

func (repo *DAConfigTplRepo) GetByIDWithTx(ctx context.Context, tx *sql.Tx, id int64) (po *dapo.DataAgentTplPo, err error) {
	po = &dapo.DataAgentTplPo{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_id", id).
		FindOne(po)
	if err != nil {
		return nil, errors.Wrapf(err, "[GetByIDWithTx]: get by id %d", id)
	}

	return
}

func (repo *DAConfigTplRepo) GetByKeyWithTx(ctx context.Context, tx *sql.Tx, key string) (po *dapo.DataAgentTplPo, err error) {
	po = &dapo.DataAgentTplPo{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	err = sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_key", key).
		FindOne(po)
	if err != nil {
		return nil, errors.Wrapf(err, "[GetByKeyWithTx]: get by key %s", key)
	}

	return
}
