package releaseacc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Create implements release.ReleasePermissionRepo.
func (repo *releasePermissionRepo) Create(ctx context.Context, tx *sql.Tx, po *dapo.ReleasePermissionPO) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)
	result, err := sr.InsertStruct(po)
	fmt.Println(result)

	return
}

func (repo *releasePermissionRepo) BatchCreate(ctx context.Context, tx *sql.Tx, pos []*dapo.ReleasePermissionPO) (err error) {
	if len(pos) == 0 {
		return
	}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(&dapo.ReleasePermissionPO{})
	_, err = sr.InsertStructs(pos)

	return
}
