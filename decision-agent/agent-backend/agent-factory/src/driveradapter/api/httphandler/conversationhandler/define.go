package conversationhandler

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

type conversationHTTPHandler struct {
	conversationSvc iportdriver.IConversationSvc
	logger          icmp.Logger
}

func (h *conversationHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	router.GET("/app/:app_key/conversation", h.List)                   // 获取会话列表
	router.GET("/app/:app_key/conversation/:id", h.Detail)             // 获取会话详情
	router.PUT("/app/:app_key/conversation/:id", h.Update)             // 更新会话
	router.DELETE("/app/:app_key/conversation/:id", h.Delete)          // 删除会话
	router.DELETE("/app/:app_key/conversation", h.DeleteByAPPKey)      // 删除指定agent应用下所有会话
	router.POST("/app/:app_key/conversation", h.Init)                  // 初始化会话
	router.PUT("/app/:app_key/conversation/:id/mark_read", h.MarkRead) // 删除指定agent应用下所有会话
}

func (h *conversationHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	router.Use(capimiddleware.SetInternalAPIFlag())
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewConversationHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &conversationHTTPHandler{
			conversationSvc: dainject.NewConversationSvc(),
			logger:          logger.GetLogger(),
		}
	})

	return _handler
}
