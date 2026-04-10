package v3agentconfigsvc

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/imodelfactoryacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/imqaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver"
)

type dataAgentConfigSvc struct {
	*service.SvcBase
	logger         icmp.Logger
	agentConfRepo  idbaccess.IDataAgentConfigRepo
	agentTplRepo   idbaccess.IDataAgentTplRepo
	releaseRepo    idbaccess.IReleaseRepo
	pubedAgentRepo idbaccess.IPubedAgentRepo
	redisCmp       icmp.RedisCmp
	OpenAICmp      icmp.IOpenAI

	umHttp iusermanagementacc.UserMgnt

	productRepo idbaccess.IProductRepo

	um2Http iumacc.UmHttpAcc

	tplSvc iv3portdriver.IDataAgentTplSvc

	modelFactoryAcc imodelfactoryacc.IModelApiAcc
	mqAccess        imqaccess.IMqAccess

	pmsSvc iv3portdriver.IPermissionSvc

	authZHttp iauthzacc.AuthZHttpAcc

	bizDomainHttp     ibizdomainacc.BizDomainHttpAcc
	bdAgentRelRepo    idbaccess.IBizDomainAgentRelRepo
	bdAgentTplRelRepo idbaccess.IBizDomainAgentTplRelRepo
}

var _ iv3portdriver.IDataAgentConfigSvc = &dataAgentConfigSvc{}

type NewDaConfSvcDto struct {
	RedisCmp       icmp.RedisCmp
	SvcBase        *service.SvcBase
	AgentConfRepo  idbaccess.IDataAgentConfigRepo
	AgentTplRepo   idbaccess.IDataAgentTplRepo
	ReleaseRepo    idbaccess.IReleaseRepo
	PubedAgentRepo idbaccess.IPubedAgentRepo
	Logger         icmp.Logger
	OpenAICmp      icmp.IOpenAI
	UmHttp         iusermanagementacc.UserMgnt
	ProductRepo    idbaccess.IProductRepo
	Um2Http        iumacc.UmHttpAcc
	ModelApiAcc    imodelfactoryacc.IModelApiAcc

	TplSvc   iv3portdriver.IDataAgentTplSvc
	MqAccess imqaccess.IMqAccess

	PmsSvc iv3portdriver.IPermissionSvc

	AuthZHttp iauthzacc.AuthZHttpAcc

	BizDomainHttp     ibizdomainacc.BizDomainHttpAcc
	BdAgentRelRepo    idbaccess.IBizDomainAgentRelRepo
	BdAgentTplRelRepo idbaccess.IBizDomainAgentTplRelRepo
}

func NewDataAgentConfigService(dto *NewDaConfSvcDto) iv3portdriver.IDataAgentConfigSvc {
	impl := &dataAgentConfigSvc{
		redisCmp:          dto.RedisCmp,
		SvcBase:           dto.SvcBase,
		agentConfRepo:     dto.AgentConfRepo,
		agentTplRepo:      dto.AgentTplRepo,
		releaseRepo:       dto.ReleaseRepo,
		pubedAgentRepo:    dto.PubedAgentRepo,
		logger:            dto.Logger,
		OpenAICmp:         dto.OpenAICmp,
		umHttp:            dto.UmHttp,
		productRepo:       dto.ProductRepo,
		um2Http:           dto.Um2Http,
		tplSvc:            dto.TplSvc,
		modelFactoryAcc:   dto.ModelApiAcc,
		mqAccess:          dto.MqAccess,
		pmsSvc:            dto.PmsSvc,
		authZHttp:         dto.AuthZHttp,
		bizDomainHttp:     dto.BizDomainHttp,
		bdAgentRelRepo:    dto.BdAgentRelRepo,
		bdAgentTplRelRepo: dto.BdAgentTplRelRepo,
	}

	return impl
}
