package releasesvc

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver"
)

type releaseSvc struct {
	*service.SvcBase
	releaseRepo            idbaccess.IReleaseRepo
	releaseHistoryRepo     idbaccess.IReleaseHistoryRepo
	agentConfigRepo        idbaccess.IDataAgentConfigRepo
	releaseCategoryRelRepo idbaccess.IReleaseCategoryRelRepo
	releasePermissionRepo  idbaccess.IReleasePermissionRepo

	categoryRepo idbaccess.ICategoryRepo

	umHttp iumacc.UmHttpAcc

	authZHttp iauthzacc.AuthZHttpAcc

	pmsSvc iv3portdriver.IPermissionSvc
}

var _ iv3portdriver.IReleaseSvc = &releaseSvc{}

func NewReleaseService(dto *NewReleaseSvcDto) iv3portdriver.IReleaseSvc {
	releaseSvcImpl := &releaseSvc{
		SvcBase:                dto.SvcBase,
		releaseRepo:            dto.ReleaseRepo,
		releaseHistoryRepo:     dto.ReleaseHistoryRepo,
		agentConfigRepo:        dto.AgentConfigRepo,
		releaseCategoryRelRepo: dto.ReleaseCategoryRepo,
		releasePermissionRepo:  dto.ReleasePermissionRepo,
		categoryRepo:           dto.CategoryRepo,
		umHttp:                 dto.UmHttp,
		authZHttp:              dto.AuthZHttp,
		pmsSvc:                 dto.PmsSvc,
	}

	return releaseSvcImpl
}

type NewReleaseSvcDto struct {
	SvcBase               *service.SvcBase
	ReleaseRepo           idbaccess.IReleaseRepo
	ReleaseHistoryRepo    idbaccess.IReleaseHistoryRepo
	AgentConfigRepo       idbaccess.IDataAgentConfigRepo
	ReleaseCategoryRepo   idbaccess.IReleaseCategoryRelRepo
	ReleasePermissionRepo idbaccess.IReleasePermissionRepo

	CategoryRepo idbaccess.ICategoryRepo

	UmHttp iumacc.UmHttpAcc

	Logger    icmp.Logger
	AuthZHttp iauthzacc.AuthZHttpAcc

	PmsSvc iv3portdriver.IPermissionSvc
}
