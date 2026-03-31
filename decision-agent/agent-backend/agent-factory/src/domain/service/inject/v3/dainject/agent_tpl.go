package dainject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/categorysvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/tplsvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/bddbacc/bdagenttpldbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/categoryacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/daconfdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/daconftpldbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/productdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/publishedtpldbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/chttpinject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/rediscmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	daTplSvcOnce sync.Once
	daTplSvcImpl iv3portdriver.IDataAgentTplSvc
)

// NewDaTplSvc 创建模板服务实例
func NewDaTplSvc() iv3portdriver.IDataAgentTplSvc {
	daTplSvcOnce.Do(func() {
		dto := &tplsvc.NewDaTplSvcDto{
			RedisCmp:          rediscmp.NewRedisCmp(),
			SvcBase:           service.NewSvcBase(),
			AgentTplRepo:      daconftpldbacc.NewDataAgentTplRepo(),
			AgentConfRepo:     daconfdbacc.NewDataAgentRepo(),
			Logger:            logger.GetLogger(),
			UmHttp:            chttpinject.NewUmHttpAcc(),
			CategorySvc:       categorysvc.NewCategorySvc(),
			ProductRepo:       productdbacc.NewProductRepo(),
			PmsSvc:            NewPermissionSvc(),
			PublishedTplRepo:  publishedtpldbacc.NewPublishedTplRepo(),
			CategoryRepo:      categoryacc.NewCategoryRepo(),
			BizDomainHttp:     chttpinject.NewBizDomainHttpAcc(),
			BdAgentTplRelRepo: bdagenttpldbacc.NewBizDomainAgentTplRelRepo(),
		}

		daTplSvcImpl = tplsvc.NewDataAgentTplService(dto)
	})

	return daTplSvcImpl
}
