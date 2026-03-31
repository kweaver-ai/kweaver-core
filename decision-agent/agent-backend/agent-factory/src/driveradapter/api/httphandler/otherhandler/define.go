package otherhandler

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

type otherHTTPHandler struct {
	otherService  iv3portdriver.IOtherSvc
	agentInOutSvc iv3portdriver.IAgentInOutSvc
}

func (o *otherHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	// https://{host}:{port}/api/agent-factory/v3/agent/dolphin-tpl/list
	router.POST("/agent/dolphin-tpl/list", o.DolphinTplList) // 获取product列表

	// https://{host}:{port}/api/agent-factory/v3/agent/temp-zone/file-ext-map
	router.GET("/agent/temp-zone/file-ext-map", o.TempZoneFileExt) // 获取temp-zone文件扩展名map

	// agent导入导出相关接口
	router.POST("/agent-inout/export", o.ExportAgent)   // 导出agent
	router.GET("/agent-inout/export", o.ExportAgentGet) // 导出agent FOR TEST

	router.POST("/agent-inout/import", o.ImportAgent) // 导入agent

	// 策略相关接口
	router.GET("/tool-result-process-strategy/category", o.CategoryList)                       // 获取策略分类列表
	router.GET("/tool-result-process-strategy/category/:category_id/strategy", o.StrategyList) // 获取策略列表
}

func (o *otherHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	// 私有路由注册
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewOtherHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &otherHTTPHandler{
			otherService:  dainject.NewOtherSvc(),
			agentInOutSvc: dainject.NewAgentInOutSvc(),
		}
	})

	return _handler
}
