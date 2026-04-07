package daconfdbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigRepo) ExistsByNameExcludeID(ctx context.Context, name, id string) (exists bool, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(&dapo.DataAgentPo{})
	exists, err = sr.
		WhereEqual("f_deleted_at", 0).
		WhereEqual("f_name", name).
		WhereNotEqual("f_id", id).
		Exists()

	return
}
