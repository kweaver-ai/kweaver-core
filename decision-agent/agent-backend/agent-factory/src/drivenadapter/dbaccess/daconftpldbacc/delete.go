package daconftpldbacc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigTplRepo) Delete(ctx context.Context, tx *sql.Tx, id int64) (err error) {
	po := &dapo.DataAgentTplPo{}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = errors.New("[DAConfigTplRepo][Delete]: uid is empty")
		return
	}

	sr.FromPo(po)

	_, err = sr.WhereEqual("f_id", id).
		SetUpdateFields([]string{
			"f_deleted_at",
			"f_deleted_by",
		}).
		UpdateByStruct(struct {
			DeletedAt int64  `db:"f_deleted_at"`
			DeletedBy string `db:"f_deleted_by"`
		}{
			DeletedAt: cutil.GetCurrentMSTimestamp(),
			DeletedBy: uid,
		})

	return
}
