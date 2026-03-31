package daconftpldbacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigTplRepo) Update(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentTplPo) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", po.ID).
		SetUpdateFields([]string{
			"f_name",
			"f_profile",

			"f_avatar_type",
			"f_avatar",

			"f_config",

			"f_updated_at",
			"f_updated_by",
			// "f_is_built_in",

			"f_status",

			"f_is_last_one",
		}).
		UpdateByStruct(po)

	return
}
