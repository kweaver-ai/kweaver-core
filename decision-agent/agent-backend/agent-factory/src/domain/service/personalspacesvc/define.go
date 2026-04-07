package personalspacesvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

// PersonalSpaceService 个人空间服务实现
type PersonalSpaceService struct {
	agentTplRepo      idbaccess.IDataAgentTplRepo
	agentConfigRepo   idbaccess.IDataAgentConfigRepo
	personalSpaceRepo idbaccess.IPersonalSpaceRepo
	releaseRepo       idbaccess.IReleaseRepo
	pubedAgentRepo    idbaccess.IPubedAgentRepo
	umHttp            iumacc.UmHttpAcc

	pmsSvc iv3portdriver.IPermissionSvc

	bizDomainHttp ibizdomainacc.BizDomainHttpAcc

	*service.SvcBase
}

var _ iv3portdriver.IPersonalSpaceService = &PersonalSpaceService{}

type NewPersonalSpaceSvcDto struct {
	SvcBase *service.SvcBase

	AgentTplRepo      idbaccess.IDataAgentTplRepo
	AgentConfigRepo   idbaccess.IDataAgentConfigRepo
	PersonalSpaceRepo idbaccess.IPersonalSpaceRepo
	ReleaseRepo       idbaccess.IReleaseRepo
	PubedAgentRepo    idbaccess.IPubedAgentRepo
	UmHttp            iumacc.UmHttpAcc

	PmsSvc iv3portdriver.IPermissionSvc

	BizDomainHttp ibizdomainacc.BizDomainHttpAcc
}

// NewPersonalSpaceService 创建个人空间服务实例
func NewPersonalSpaceService(dto *NewPersonalSpaceSvcDto) iv3portdriver.IPersonalSpaceService {
	personalSpaceServiceImpl := &PersonalSpaceService{
		agentTplRepo:      dto.AgentTplRepo,
		agentConfigRepo:   dto.AgentConfigRepo,
		personalSpaceRepo: dto.PersonalSpaceRepo,
		releaseRepo:       dto.ReleaseRepo,
		pubedAgentRepo:    dto.PubedAgentRepo,
		umHttp:            dto.UmHttp,
		pmsSvc:            dto.PmsSvc,
		SvcBase:           dto.SvcBase,
		bizDomainHttp:     dto.BizDomainHttp,
	}

	return personalSpaceServiceImpl
}
