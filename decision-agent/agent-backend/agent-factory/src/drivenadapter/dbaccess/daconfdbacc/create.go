package daconfdbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigRepo) Create(ctx context.Context, tx *sql.Tx, id string, po *dapo.DataAgentPo) (err error) {
	po.ID = id

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)
	_, err = sr.InsertStruct(po)

	return
}

func (repo *DAConfigRepo) CreateBatch(ctx context.Context, tx *sql.Tx, pos []*dapo.DataAgentPo) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(&dapo.DataAgentPo{})
	_, err = sr.InsertStructs(pos)

	return
}
