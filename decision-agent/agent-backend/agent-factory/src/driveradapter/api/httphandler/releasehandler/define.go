package releasehandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/httphandler/apiv3common"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type releaseHandler struct {
	releaseSvc iv3portdriver.IReleaseSvc
	logger     icmp.Logger
}

func (h *releaseHandler) RegPubRouter(router *gin.RouterGroup) {
	router.POST("/agent/:agent_id/publish", h.Publish)
	router.PUT("/agent/:agent_id/unpublish", h.UnPublish)
	router.GET("/agent/:agent_id/release-history", h.HistoryList)
	router.GET("/agent/:agent_id/release-history/:history_id", h.HistoryInfo)

	// 发布信息相关接口
	router.GET("/agent/:agent_id/publish-info", h.GetPublishInfo)
	router.PUT("/agent/:agent_id/publish-info", h.UpdatePublishInfo)
}

func (h *releaseHandler) RegPriRouter(router *gin.RouterGroup) {
	g := apiv3common.GetPrivateRouterGroup(router)

	// 私有路由注册
	g.POST("/agent/:agent_id/publish", h.Publish) // 发布 Agent

	g.GET("/agent/:agent_id/publish-info", h.GetPublishInfo)
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewReleaseHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &releaseHandler{
			logger:     logger.GetLogger(),
			releaseSvc: dainject.NewReleaseSvc(),
		}
	})

	return _handler
}
