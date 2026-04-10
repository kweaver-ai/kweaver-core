package dainject

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	v3agentconfigsvc "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/agentconfigsvc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/bddbacc/bdagentdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/bddbacc/bdagenttpldbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/daconfdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/daconftpldbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/productdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/releaseacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/chttpinject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/httpinject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/usermanagementacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/mqaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/cmpopenai"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/rediscmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	daConfSvcOnce sync.Once
	daConfSvcImpl iv3portdriver.IDataAgentConfigSvc
)

// NewDaConfSvc
func NewDaConfSvc() iv3portdriver.IDataAgentConfigSvc {
	daConfSvcOnce.Do(func() {
		mfConf := global.GConfig.ModelFactory
		baseURL := getModelApiUrlPrefix(mfConf)
		openAICmp := cmpopenai.NewOpenAICmp(mfConf.LLM.APIKey, baseURL, mfConf.LLM.DefaultModelName, true)

		dto := &v3agentconfigsvc.NewDaConfSvcDto{
			RedisCmp:          rediscmp.NewRedisCmp(),
			SvcBase:           service.NewSvcBase(),
			AgentConfRepo:     daconfdbacc.NewDataAgentRepo(),
			AgentTplRepo:      daconftpldbacc.NewDataAgentTplRepo(),
			ReleaseRepo:       releaseacc.NewReleaseRepo(),
			PubedAgentRepo:    pubedagentdbacc.NewPubedAgentRepo(),
			Logger:            logger.GetLogger(),
			OpenAICmp:         openAICmp,
			UmHttp:            usermanagementacc.NewClient(),
			ProductRepo:       productdbacc.NewProductRepo(),
			Um2Http:           chttpinject.NewUmHttpAcc(),
			TplSvc:            NewDaTplSvc(),
			ModelApiAcc:       httpinject.NewModelApiAcc(),
			MqAccess:          mqaccess.NewMqAccess(),
			PmsSvc:            NewPermissionSvc(),
			AuthZHttp:         chttpinject.NewAuthZHttpAcc(),
			BizDomainHttp:     chttpinject.NewBizDomainHttpAcc(),
			BdAgentRelRepo:    bdagentdbacc.NewBizDomainAgentRelRepo(),
			BdAgentTplRelRepo: bdagenttpldbacc.NewBizDomainAgentTplRelRepo(),
		}

		daConfSvcImpl = v3agentconfigsvc.NewDataAgentConfigService(dto)
	})

	return daConfSvcImpl
}
