package releaseacc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// DelByReleaseId implements idbaccess.ReleasePermissionRepo.
func (repo *releasePermissionRepo) DelByReleaseID(ctx context.Context, tx *sql.Tx, releaseID string) (err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	if tx != nil {
		sr = dbhelper2.TxSr(tx, repo.logger)
	}

	po := &dapo.ReleasePermissionPO{}
	sr.FromPo(po)

	_, err = sr.WhereEqual("f_release_id", releaseID).Delete()
	if err != nil {
		return errors.Wrapf(err, "delete permission by release id %s", releaseID)
	}

	return nil
}
