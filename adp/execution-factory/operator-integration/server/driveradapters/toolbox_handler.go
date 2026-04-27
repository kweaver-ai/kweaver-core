package driveradapters

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/toolbox"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
)

// ToolBoxRestHandler 工具箱rest接口
type ToolBoxRestHandler interface {
	// RegisterPrivate 注册内部API
	RegisterPrivate(engine *gin.RouterGroup)

	// RegisterPublic 注册外部API
	RegisterPublic(engine *gin.RouterGroup)
}

type toolboxRestHandler struct {
	ToolBoxHandler        toolbox.ToolBoxHandler
	Logger                interfaces.Logger
	businessDomainService interfaces.IBusinessDomainService
}

var (
	tOnce    sync.Once
	tHandler ToolBoxRestHandler
)

func NewToolBoxRestHandler() ToolBoxRestHandler {
	tOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		tHandler = &toolboxRestHandler{
			ToolBoxHandler:        toolbox.NewToolBoxHandler(),
			Logger:                confLoader.GetLogger(),
			businessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return tHandler
}

// RegisterPrivate 注册内部API
func (r *toolboxRestHandler) RegisterPrivate(engine *gin.RouterGroup) {
	/*工具箱相关接口*/
	// 查询工具箱信息
	engine.GET("/tool-box/list", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.QueryToolBoxPage)
	engine.GET("/tool-box/:box_id", r.ToolBoxHandler.QueryToolBox)
	engine.GET("/tool-box/:box_id/tool/:tool_id", r.ToolBoxHandler.QueryTool)
	engine.GET("/tool-box/:box_id/tools/list", r.ToolBoxHandler.QueryBoxToolPage)
	engine.POST("/tool-box/:box_id/proxy/:tool_id", middlewareProxyRequest(), r.ToolBoxHandler.ExecuteTool)
	// 内置工具注册
	engine.POST("/tool-box/intcomp", middlewareBusinessDomain(true, true, r.businessDomainService), r.ToolBoxHandler.CreateInternalToolBox)
}

// RegisterPublic 注册外部API
func (r *toolboxRestHandler) RegisterPublic(engine *gin.RouterGroup) {
	engine.POST("/tool-box", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.CreateToolBox)
	engine.POST("/tool-box/:box_id", r.ToolBoxHandler.UpdateToolBox)
	engine.GET("/tool-box/:box_id", r.ToolBoxHandler.QueryToolBox)
	engine.DELETE("/tool-box/:box_id", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.DeleteToolBox)
	engine.GET("/tool-box/list", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.QueryToolBoxPage)
	// 工具
	engine.POST("/tool-box/:box_id/tool", r.ToolBoxHandler.CreateTool)
	engine.POST("/tool-box/:box_id/tool/:tool_id", r.ToolBoxHandler.UpdateTool)
	engine.GET("/tool-box/:box_id/tool/:tool_id", r.ToolBoxHandler.QueryTool)
	engine.POST("/tool-box/:box_id/tools/batch-delete", r.ToolBoxHandler.DeleteBoxTool)
	engine.GET("/tool-box/:box_id/tools/list", r.ToolBoxHandler.QueryBoxToolPage)
	engine.POST("/tool-box/:box_id/tools/status", r.ToolBoxHandler.UpdateToolStatus)
	engine.POST("/tool-box/:box_id/tool/:tool_id/debug", middlewareProxyRequest(), r.ToolBoxHandler.DebugTool)
	engine.POST("/tool-box/:box_id/proxy/:tool_id", middlewareProxyRequest(), r.ToolBoxHandler.ExecuteTool)
	engine.POST("/tool-box/:box_id/status", r.ToolBoxHandler.UpdateToolBoxStatus)

	// 算子转换成工具
	engine.POST("/operator/convert/tool", r.ToolBoxHandler.OperatorToTool)
	// 内置工具注册
	engine.POST("/tool-box/intcomp", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.CreateInternalToolBox)
	// 批量获取已发布工具箱信息
	engine.GET("/tool-box/market/:box_id/:fields", r.ToolBoxHandler.GetReleaseToolBoxInfo)

	/*工具箱市场界面*/
	engine.GET("/tool-box/market", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.QueryMarketToolBoxPage)
	engine.GET("/tool-box/market/:box_id", r.ToolBoxHandler.QueryMarketToolBox)
	engine.GET("/tool-box/market/tools", middlewareBusinessDomain(true, false, r.businessDomainService), r.ToolBoxHandler.GetMarketToolList)
}
