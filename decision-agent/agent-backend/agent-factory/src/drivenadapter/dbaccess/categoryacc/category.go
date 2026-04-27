package categoryacc

import (
	"context"
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	categoryRepoOnce sync.Once
	categoryRepoImpl idbaccess.ICategoryRepo
)

type categoryRepo struct {
	*drivenadapter.RepoBase

	db     *sqlx.DB
	logger icmp.Logger
}

// GetByReleaseId implements idbaccess.ICategoryRepo.
func (repo *categoryRepo) GetByReleaseId(ctx context.Context, releaaseId string) (rt []*dapo.CategoryPO, err error) {
	return nil, nil
}

// DeleteByAgentId implements idbaccess.CategoryRepo.

// List implements idbaccess.CategoryRepo.
func (repo *categoryRepo) List(ctx context.Context, req interface{}) (rt []*dapo.CategoryPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	po := &dapo.CategoryPO{}
	sr.FromPo(po)

	list := make([]dapo.CategoryPO, 0)

	err = sr.Find(&list)
	if err != nil {
		return nil, err
	}

	rt = cutil.SliceToPtrSlice(list)

	return rt, err
}

var _ idbaccess.ICategoryRepo = &categoryRepo{}

func NewCategoryRepo() idbaccess.ICategoryRepo {
	categoryRepoOnce.Do(func() {
		categoryRepoImpl = &categoryRepo{
			db:       global.GDB,
			logger:   logger.GetLogger(),
			RepoBase: drivenadapter.NewRepoBase(),
		}
	})

	return categoryRepoImpl
}
