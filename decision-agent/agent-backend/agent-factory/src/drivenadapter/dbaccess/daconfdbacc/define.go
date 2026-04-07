package daconfdbacc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	agentRepoOnce sync.Once
	agentRepoImpl idbaccess.IDataAgentConfigRepo
)

type DAConfigRepo struct {
	idbaccess.IDBAccBaseRepo

	db *sqlx.DB

	logger icmp.Logger
}

var _ idbaccess.IDataAgentConfigRepo = &DAConfigRepo{}

func NewDataAgentRepo() idbaccess.IDataAgentConfigRepo {
	agentRepoOnce.Do(func() {
		agentRepoImpl = &DAConfigRepo{
			db:             global.GDB,
			logger:         logger.GetLogger(),
			IDBAccBaseRepo: dbaccess.NewDBAccBase(),
		}
	})

	return agentRepoImpl
}
