package releaseacc

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	releaseRepoOnce sync.Once
	releaseRepoImpl idbaccess.IReleaseRepo
)

type releaseRepo struct {
	idbaccess.IDBAccBaseRepo

	db     *sqlx.DB
	logger icmp.Logger
}

var _ idbaccess.IReleaseRepo = &releaseRepo{}

func NewReleaseRepo() idbaccess.IReleaseRepo {
	releaseRepoOnce.Do(func() {
		releaseRepoImpl = &releaseRepo{
			db:             global.GDB,
			logger:         logger.GetLogger(),
			IDBAccBaseRepo: dbaccess.NewDBAccBase(),
		}
	})

	return releaseRepoImpl
}
