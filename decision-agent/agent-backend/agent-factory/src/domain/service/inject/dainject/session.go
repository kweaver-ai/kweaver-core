package dainject

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/sessionsvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/httpinject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/redisaccess/sessionredisacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	sessionSvcOnce sync.Once
	sessionSvcImpl iportdriver.ISessionSvc
)

func NewSessionSvc() iportdriver.ISessionSvc {
	sessionSvcOnce.Do(func() {
		dto := &sessionsvc.NewSessionSvcDto{
			Logger:          logger.GetLogger(),
			SessionRedis:    sessionredisacc.NewSessionRedisAcc(),
			AgentExecutorV1: httpinject.NewAgentExecutorV1HttpAcc(),
		}
		sessionSvcImpl = sessionsvc.NewSessionService(dto)
	})

	return sessionSvcImpl
}
