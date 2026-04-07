package dainject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/publishedsvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/daconftpldbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/productdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/publishedtpldbacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/chttpinject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

var (
	publishedSvcOnce sync.Once
	publishedSvcImpl iv3portdriver.IPublishedSvc
)

// NewPublishedSvc .
func NewPublishedSvc() iv3portdriver.IPublishedSvc {
	publishedSvcOnce.Do(func() {
		dto := &publishedsvc.NewPublishedSvcDto{
			SvcBase:          service.NewSvcBase(),
			AgentTplRepo:     daconftpldbacc.NewDataAgentTplRepo(),
			PublishedTplRepo: publishedtpldbacc.NewPublishedTplRepo(),
			ProductRepo:      productdbacc.NewProductRepo(),
			UmHttp:           chttpinject.NewUmHttpAcc(),
			AuthZHttp:        chttpinject.NewAuthZHttpAcc(),
			PubedAgentRepo:   pubedagentdbacc.NewPubedAgentRepo(),
			PmsSvc:           NewPermissionSvc(),
			BizDomainHttp:    chttpinject.NewBizDomainHttpAcc(),
		}

		publishedSvcImpl = publishedsvc.NewPublishedService(dto)
	})

	return publishedSvcImpl
}
