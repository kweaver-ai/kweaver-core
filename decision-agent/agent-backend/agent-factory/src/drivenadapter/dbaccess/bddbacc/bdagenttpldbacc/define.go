package bdagenttpldbacc

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
	bizDomainAgentTplRelRepoOnce sync.Once
	bizDomainAgentTplRelRepoImpl idbaccess.IBizDomainAgentTplRelRepo
)

// BizDomainAgentTplRelRepo 业务域与agent模板关联表操作实现
type BizDomainAgentTplRelRepo struct {
	idbaccess.IDBAccBaseRepo

	db     *sqlx.DB
	logger icmp.Logger
}

var _ idbaccess.IBizDomainAgentTplRelRepo = &BizDomainAgentTplRelRepo{}

func NewBizDomainAgentTplRelRepo() idbaccess.IBizDomainAgentTplRelRepo {
	bizDomainAgentTplRelRepoOnce.Do(func() {
		bizDomainAgentTplRelRepoImpl = &BizDomainAgentTplRelRepo{
			db:             global.GDB,
			logger:         logger.GetLogger(),
			IDBAccBaseRepo: dbaccess.NewDBAccBase(),
		}
	})

	return bizDomainAgentTplRelRepoImpl
}
