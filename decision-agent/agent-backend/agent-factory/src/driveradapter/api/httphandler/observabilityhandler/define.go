package observabilityhandler

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

type observabilityHTTPHandler struct {
	observabilitySvc iportdriver.IObservability
	conversationSvc  iportdriver.IConversationSvc
	logger           icmp.Logger
}

func (h *observabilityHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	permissionRouter := router.Group("",

		apimiddleware.CheckAgentUsePms(),
		// apimiddleware.CheckSpaceMember(),
	)
	// permissionRouter.POST("/observability/analytics/query", h.AnalyticsQuery)

	// Agent 可观测信息查询
	permissionRouter.POST("/observability/agent/:agent_id/detail", h.AgentDetail)

	// 对话列表查询
	// NOTE: 在获取Agent下的会话列表这个接口中，路径参数的agent_id,实际上前端传入的是agent_key的值，
	// NOTE：前端在chat接口的agent_id字段传入值是agent_key; 在debug接口的agent_id字段传入值是agent_id;
	// NOTE：因此产生的对话列表中，会话ID实际值为agent_id的是debug产生的对话，会话ID实际值为agent_key的是chat产生的对话
	// NOTE: 这个接口获取的是chat产生的对话列表,因此前端传入的是agent_key的值，
	permissionRouter.POST("/observability/agent/:agent_id/conversation", h.ConversationList)

	// Session 列表查询
	permissionRouter.POST("/observability/agent/:agent_id/conversation/:conversation_id/session", h.SessionList)

	// Session 详情查询
	permissionRouter.POST("/observability/agent/:agent_id/conversation/:conversation_id/session/:session_id/detail", h.SessionDetail)

	// Run 列表查询
	permissionRouter.POST("/observability/agent/:agent_id/conversation/:conversation_id/session/:session_id/run", h.RunList)

	// Run 详情查询
	permissionRouter.POST("/observability/agent/:agent_id/conversation/:conversation_id/session/:session_id/run/:run_id/detail", h.RunDetail)
}

func (h *observabilityHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	permissionRouter := router.Group("") // apimiddleware.CheckSpaceMemberInternal(),

	permissionRouter.POST("/observability/analytics/query", h.AnalyticsQuery)
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewObservabilityHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &observabilityHTTPHandler{
			observabilitySvc: dainject.NewObservabilitySvc(),
			conversationSvc:  dainject.NewConversationSvc(),
			logger:           logger.GetLogger(),
		}
	})

	return _handler
}
