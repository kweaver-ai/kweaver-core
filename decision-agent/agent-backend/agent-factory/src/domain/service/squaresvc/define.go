package squaresvc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/daconfdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/releaseacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/visithistoryacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/chttpinject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/usermanagementacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

var (
	squareSvcOnce sync.Once
	squareSvcImpl iv3portdriver.ISquareSvc
)

type squareSvc struct {
	*service.SvcBase
	agentConfRepo            idbaccess.IDataAgentConfigRepo
	releaseRepo              idbaccess.IReleaseRepo
	releaseHistoryRepo       idbaccess.IReleaseHistoryRepo
	usermanagementHttpClient iusermanagementacc.UserMgnt
	visitHistoryRepo         idbaccess.IVisitHistoryRepo

	umHttp iumacc.UmHttpAcc
}

var _ iv3portdriver.ISquareSvc = &squareSvc{}

func NewSquareService() iv3portdriver.ISquareSvc {
	squareSvcOnce.Do(func() {
		squareSvcImpl = &squareSvc{
			SvcBase:                  service.NewSvcBase(),
			releaseRepo:              releaseacc.NewReleaseRepo(),
			releaseHistoryRepo:       releaseacc.NewReleaseHistoryRepo(),
			agentConfRepo:            daconfdbacc.NewDataAgentRepo(),
			usermanagementHttpClient: usermanagementacc.NewClient(),
			visitHistoryRepo:         visithistoryacc.NewVisitHistoryRepo(),

			umHttp: chttpinject.NewUmHttpAcc(),
		}
	})

	return squareSvcImpl
}
