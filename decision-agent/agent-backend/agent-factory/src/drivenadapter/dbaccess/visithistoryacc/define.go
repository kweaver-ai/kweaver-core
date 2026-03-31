package visithistoryacc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	visitHistoryRepoOnce sync.Once
	visitHistoryRepoImpl idbaccess.IVisitHistoryRepo
)

type visitHistoryRepo struct {
	*drivenadapter.RepoBase

	db     *sqlx.DB
	logger icmp.Logger
}

var _ idbaccess.IVisitHistoryRepo = &visitHistoryRepo{}

func NewVisitHistoryRepo() idbaccess.IVisitHistoryRepo {
	visitHistoryRepoOnce.Do(func() {
		visitHistoryRepoImpl = &visitHistoryRepo{
			db:       global.GDB,
			logger:   logger.GetLogger(),
			RepoBase: drivenadapter.NewRepoBase(),
		}
	})

	return visitHistoryRepoImpl
}
