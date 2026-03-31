package categoryacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *categoryRepo) GetIDNameMap(ctx context.Context, ids []string) (m map[string]string, err error) {
	m = make(map[string]string, len(ids))

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	po := &dapo.CategoryPO{}
	sr.FromPo(po)

	poList := make([]dapo.CategoryPO, 0)

	err = sr.Select([]string{"f_id", "f_name"}).
		In("f_id", ids).
		Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get by ids %v", ids)
		return
	}

	for _, _po := range poList {
		m[_po.ID] = _po.Name
	}

	return
}
