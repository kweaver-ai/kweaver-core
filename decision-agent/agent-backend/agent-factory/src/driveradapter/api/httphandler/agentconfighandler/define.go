package v3agentconfighandler

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/httphandler/apiv3common"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type daConfHTTPHandler struct {
	daConfSvc iv3portdriver.IDataAgentConfigSvc
	logger    icmp.Logger
}

func (h *daConfHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	router.POST("/agent", h.Create)            // 新建agent
	router.POST("/agent/react", h.CreateReact) // 新建react agent
	router.PUT("/agent/:agent_id", h.Update)   // 编辑agent
	router.GET("/agent/:agent_id", h.Detail)   // 获取agent详情
	// router.GET("/agent", h.AgentList)        // agent列表

	router.GET("/agent/by-key/:key", h.DetailByKey) // 获取agent详情 by key
	router.DELETE("/agent/:agent_id", h.Delete)     // 删除agent

	router.POST("/agent/ai-autogen", h.AIAutogenContent)

	// 复制相关接口
	router.POST("/agent/:agent_id/copy", h.Copy)                               // 复制Agent
	router.POST("/agent/:agent_id/copy2tpl", h.Copy2Tpl)                       // 复制Agent为模板
	router.POST("/agent/:agent_id/copy2tpl-and-publish", h.Copy2TplAndPublish) // 复制Agent为模板并发布

	// 获取内置头像列表
	router.GET("/agent/avatar/built-in", h.GetBuiltInAvatarList)
	// 获取内置头像
	router.GET("/agent/avatar/built-in/:avatar_id", h.GetBuiltInAvatar)

	// 获取SELF_CONFIG字段结构
	router.GET("/agent-self-config-fields", h.SelfConfig)
}

func (h *daConfHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	g := apiv3common.GetPrivateRouterGroup(router)

	// 私有路由注册
	g.POST("/agent", h.Create)             // 新建agent
	g.PUT("/agent/:agent_id", h.Update)    // 编辑agent
	g.DELETE("/agent/:agent_id", h.Delete) // 删除agent

	g.GET("/agent/:agent_id", h.Detail)        // 获取agent详情
	g.GET("/agent/by-key/:key", h.DetailByKey) // 获取agent详情 by key
	g.POST("/agent-fields", h.BatchFields)     // 批量获取agent指定字段
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewDAConfHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &daConfHTTPHandler{
			daConfSvc: dainject.NewDaConfSvc(),
			logger:    logger.GetLogger(),
		}
	})

	return _handler
}
