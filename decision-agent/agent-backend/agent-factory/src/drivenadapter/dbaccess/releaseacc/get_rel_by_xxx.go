package releaseacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *releaseCategoryRelRepo) GetByReleaseID(ctx context.Context, releaseID string) (poList []*dapo.ReleaseCategoryRelPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	po := &dapo.ReleaseCategoryRelPO{}
	sr.FromPo(po)

	pos := make([]dapo.ReleaseCategoryRelPO, 0)

	err = sr.WhereEqual("f_release_id", releaseID).Find(&pos)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "get category by release id %s", releaseID)
	}

	poList = cutil.SliceToPtrSlice(pos)

	return
}

// GetByCategoryId implements idbaccess.IReleaseCategoryRelRepo.
func (repo *releaseCategoryRelRepo) GetByCategoryID(ctx context.Context, categoryID string) (rt []*dapo.ReleaseCategoryRelPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	po := &dapo.ReleaseCategoryRelPO{}
	sr.FromPo(po)

	poList := make([]dapo.ReleaseCategoryRelPO, 0)

	err = sr.WhereEqual("f_category_id", categoryID).Find(&poList)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			return
		}

		err = errors.Wrapf(err, "get category by category id %s", categoryID)

		return
	}

	rt = cutil.SliceToPtrSlice(poList)

	return
}
