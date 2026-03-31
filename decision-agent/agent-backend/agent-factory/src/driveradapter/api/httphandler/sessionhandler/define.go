package sessionhandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type sessionHTTPHandler struct {
	sessionSvc iportdriver.ISessionSvc
	logger     icmp.Logger
}

func (h *sessionHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	router.PUT("/conversation/session/:conversation_id", h.Manage) // 管理对话session
}

func (h *sessionHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	router.Use(capimiddleware.SetInternalAPIFlag())
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewSessionHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &sessionHTTPHandler{
			sessionSvc: dainject.NewSessionSvc(),
			logger:     logger.GetLogger(),
		}
	})

	return _handler
}
