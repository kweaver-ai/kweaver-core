package permissionsvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

type permissionSvc struct {
	*service.SvcBase
	agentConfRepo         idbaccess.IDataAgentConfigRepo
	releaseRepo           idbaccess.IReleaseRepo
	releasePermissionRepo idbaccess.IReleasePermissionRepo
	umHttp                iumacc.UmHttpAcc
	authZHttp             iauthzacc.AuthZHttpAcc
}

var _ iv3portdriver.IPermissionSvc = &permissionSvc{}

type NewPermissionSvcDto struct {
	SvcBase               *service.SvcBase
	AgentConfigRepo       idbaccess.IDataAgentConfigRepo
	ReleaseRepo           idbaccess.IReleaseRepo
	ReleasePermissionRepo idbaccess.IReleasePermissionRepo
	UmHttp                iumacc.UmHttpAcc
	AuthZHttp             iauthzacc.AuthZHttpAcc
}

func NewPermissionService(dto *NewPermissionSvcDto) iv3portdriver.IPermissionSvc {
	permissionSvcImpl := &permissionSvc{
		SvcBase:               dto.SvcBase,
		agentConfRepo:         dto.AgentConfigRepo,
		releaseRepo:           dto.ReleaseRepo,
		releasePermissionRepo: dto.ReleasePermissionRepo,
		umHttp:                dto.UmHttp,
		authZHttp:             dto.AuthZHttp,
	}

	return permissionSvcImpl
}
