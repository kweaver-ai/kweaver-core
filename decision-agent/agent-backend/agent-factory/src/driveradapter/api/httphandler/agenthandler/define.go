package agenthandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/dainject"
	apimiddleware "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/apimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type agentHTTPHandler struct {
	agentSvc iportdriver.IAgent
	logger   icmp.Logger
}

func (h *agentHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	router.POST("/app/:app_key/chat/resume", h.ResumeChat)
	router.POST("/app/:app_key/chat/termination", h.TerminateChat)

	permissionRouter := router.Group("",
		apimiddleware.CheckAgentUsePms(),
		// apimiddleware.CheckSpaceMember(),
	)
	permissionRouter.POST("/app/:app_key/chat/completion", h.Chat)
	permissionRouter.POST("/app/:app_key/debug/completion", h.Debug)
	permissionRouter.POST("/app/:app_key/api/chat/completion", h.APIChat)
	permissionRouter.POST("/api/chat/completion", h.APIChat)
	permissionRouter.POST("/app/:app_key/api/doc", h.GetAPIDoc)
	// permissionRouter.POST("/conversation/session/init", h.ConversationSessionInit)
}

func (h *agentHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	// router.POST("/app/:app_key/chat/completion", h.InternalChat)
	// router.POST("/app/:app_key/chat/resume", h.ResumeChat)
	// router.POST("/app/:app_key/debug/completion", h.Debug)
	// router.POST("/app/:app_key/chat/termination", h.TerminateChat)
	// router.POST("/app/:app_key/api/chat/completion", h.APIChat)
	// router.POST("/app/:app_key/api/doc", h.GetAPIDoc)
	permissionRouter := router.Group("",
		apimiddleware.CheckAgentUsePmsInternal(),
		// apimiddleware.CheckSpaceMemberInternal(),
	)
	permissionRouter.POST("/app/:app_key/chat/completion", h.InternalChat)
	permissionRouter.POST("/app/:app_key/api/chat/completion", h.InternalAPIChat)
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewAgentHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &agentHTTPHandler{
			agentSvc: dainject.NewAgentSvc(),
			logger:   logger.GetLogger(),
		}
	})

	return _handler
}
