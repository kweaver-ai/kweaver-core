package bdagentdbacc

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
	bizDomainAgentRelRepoOnce sync.Once
	bizDomainAgentRelRepoImpl idbaccess.IBizDomainAgentRelRepo
)

// BizDomainAgentRelRepo 业务域与agent关联表操作实现
type BizDomainAgentRelRepo struct {
	idbaccess.IDBAccBaseRepo

	db     *sqlx.DB
	logger icmp.Logger
}

var _ idbaccess.IBizDomainAgentRelRepo = &BizDomainAgentRelRepo{}

func NewBizDomainAgentRelRepo() idbaccess.IBizDomainAgentRelRepo {
	bizDomainAgentRelRepoOnce.Do(func() {
		bizDomainAgentRelRepoImpl = &BizDomainAgentRelRepo{
			db:             global.GDB,
			logger:         logger.GetLogger(),
			IDBAccBaseRepo: dbaccess.NewDBAccBase(),
		}
	})

	return bizDomainAgentRelRepoImpl
}
