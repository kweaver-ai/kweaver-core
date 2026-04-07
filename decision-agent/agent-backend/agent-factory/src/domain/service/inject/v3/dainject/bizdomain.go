package dainject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/bizdomainsvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/chttpinject"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	bizDomainSvcOnce sync.Once
	bizDomainSvcImpl *bizdomainsvc.BizDomainSvc
)

// NewBizDomainSvc 创建业务域服务实例
func NewBizDomainSvc() *bizdomainsvc.BizDomainSvc {
	bizDomainSvcOnce.Do(func() {
		dto := &bizdomainsvc.NewBizDomainSvcDto{
			SvcBase:       service.NewSvcBase(),
			Logger:        logger.GetLogger(),
			BizDomainHttp: chttpinject.NewBizDomainHttpAcc(),
		}

		bizDomainSvcImpl = bizdomainsvc.NewBizDomainService(dto)
	})

	return bizDomainSvcImpl
}
