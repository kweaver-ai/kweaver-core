package daconfdbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigRepo) getUpdateFields() []string {
	return []string{
		"f_name",
		"f_profile",
		"f_avatar_type",
		"f_avatar",
		"f_config",
		"f_updated_at",
		"f_updated_by",
		"f_is_built_in",
		"f_status",
		"f_product_key",
	}
}

func (repo *DAConfigRepo) Update(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentPo) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	updateFields := repo.getUpdateFields()

	_, err = sr.WhereEqual("f_id", po.ID).
		SetUpdateFields(updateFields).
		UpdateByStruct(po)

	return
}

func (repo *DAConfigRepo) UpdateByKey(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentPo) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	updateFields := repo.getUpdateFields()

	_, err = sr.WhereEqual("f_key", po.Key).
		SetUpdateFields(updateFields).
		UpdateByStruct(po)

	return
}
