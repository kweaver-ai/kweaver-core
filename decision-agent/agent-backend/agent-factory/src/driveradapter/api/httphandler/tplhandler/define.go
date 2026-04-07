package tplhandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type daTplHTTPHandler struct {
	daTplSvc iv3portdriver.IDataAgentTplSvc
	logger   icmp.Logger
}

func (h *daTplHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	// router.POST("/agent-tpl", h.Create)    // 新建模板
	router.PUT("/agent-tpl/:id", h.Update) // 编辑模板
	router.GET("/agent-tpl/:id", h.Detail) // 获取模板详情

	router.GET("/agent-tpl/by-key/:key", h.DetailByKey) // 获取模板详情 by key
	router.DELETE("/agent-tpl/:id", h.Delete)           // 删除模板
	router.POST("/agent-tpl/:id/publish", h.Publish)    // 发布模板
	router.PUT("/agent-tpl/:id/unpublish", h.Unpublish) // 取消发布模板

	// 新增的3个接口
	router.GET("/agent-tpl/:id/publish-info", h.GetPublishInfo)    // 获取模板发布信息
	router.PUT("/agent-tpl/:id/publish-info", h.UpdatePublishInfo) // 更新模板发布信息
	router.POST("/agent-tpl/:id/copy", h.Copy)                     // 复制智能体模板
}

func (h *daTplHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	// 后面如果有，则打开注释，基于g来注册私有接口
	// g := apiv3common.GetPrivateRouterGroup(router)
	// g.GET()
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewDATplHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &daTplHTTPHandler{
			daTplSvc: dainject.NewDaTplSvc(),
			logger:   logger.GetLogger(),
		}
	})

	return _handler
}
