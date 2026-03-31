package daconftpldbacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (repo *DAConfigTplRepo) ExistsByName(ctx context.Context, name string) (exists bool, err error) {
	po := &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	count, err := sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_name", name).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *DAConfigTplRepo) ExistsByKey(ctx context.Context, key string) (exists bool, err error) {
	po := &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	count, err := sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_key", key).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *DAConfigTplRepo) ExistsByID(ctx context.Context, id int64) (exists bool, err error) {
	po := &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	count, err := sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_id", id).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *DAConfigTplRepo) ExistsByNameExcludeID(ctx context.Context, name string, id int64) (exists bool, err error) {
	po := &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	count, err := sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_name", name).
		WhereNotEqual("f_id", id).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *DAConfigTplRepo) ExistsByKeyExcludeID(ctx context.Context, key string, id int64) (exists bool, err error) {
	po := &dapo.DataAgentTplPo{}
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po)

	count, err := sr.WhereEqual("f_deleted_at", 0).
		WhereEqual("f_key", key).
		WhereNotEqual("f_id", id).
		Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
