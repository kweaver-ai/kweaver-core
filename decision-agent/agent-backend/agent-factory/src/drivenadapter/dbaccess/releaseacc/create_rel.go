package releaseacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Create implements release.IReleaseCategoryRelRepo.
func (repo *releaseCategoryRelRepo) Create(ctx context.Context, tx *sql.Tx, po *dapo.ReleaseCategoryRelPO) (err error) {
	po.ID = cutil.UlidMake()

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)
	_, err = sr.InsertStruct(po)

	return
}

// 批量创建
func (repo *releaseCategoryRelRepo) BatchCreate(ctx context.Context, tx *sql.Tx, pos []*dapo.ReleaseCategoryRelPO) (err error) {
	if len(pos) == 0 {
		return
	}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(&dapo.ReleaseCategoryRelPO{})
	_, err = sr.InsertStructs(pos)

	return
}
