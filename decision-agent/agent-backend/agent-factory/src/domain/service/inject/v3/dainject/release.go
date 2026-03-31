package dainject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/releasesvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/categoryacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/daconfdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/releaseacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/chttpinject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	releaseSvcOnce sync.Once
	releaseSvcImpl iv3portdriver.IReleaseSvc
)

// NewDaConfSvc
func NewReleaseSvc() iv3portdriver.IReleaseSvc {
	releaseSvcOnce.Do(func() {
		dto := &releasesvc.NewReleaseSvcDto{
			SvcBase:               service.NewSvcBase(),
			ReleaseRepo:           releaseacc.NewReleaseRepo(),
			ReleaseHistoryRepo:    releaseacc.NewReleaseHistoryRepo(),
			ReleaseCategoryRepo:   releaseacc.NewReleaseCategoryRelRepo(),
			ReleasePermissionRepo: releaseacc.NewReleasePermissionRepo(),
			AgentConfigRepo:       daconfdbacc.NewDataAgentRepo(),
			Logger:                logger.GetLogger(),
			CategoryRepo:          categoryacc.NewCategoryRepo(),
			UmHttp:                chttpinject.NewUmHttpAcc(),
			AuthZHttp:             chttpinject.NewAuthZHttpAcc(),
			PmsSvc:                NewPermissionSvc(),
		}

		releaseSvcImpl = releasesvc.NewReleaseService(dto)
	})

	return releaseSvcImpl
}
