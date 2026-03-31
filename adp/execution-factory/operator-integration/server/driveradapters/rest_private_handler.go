// Package driveradapters 定义驱动适配器
// @file rest_private_handler.go
// @description: 定义rest私有接口适配器
package driveradapters

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
)

type restPrivateHandler struct {
	OperatorRestHandler   OperatorRestHandler
	ToolBoxRestHandler    ToolBoxRestHandler
	MCPRestHandler        MCPRestHandler
	UpgradeHandler        common.UpgradeHandler
	UnifiedProxyHandler   common.UnifiedProxyHandler
	Logger                interfaces.Logger
	SkillRestHandler      SkillRestHandler
	businessDomainService interfaces.IBusinessDomainService
	Hydra                 interfaces.Hydra
}

// NewRestPrivateHandler 创建restHandler实例
func NewRestPrivateHandler() interfaces.HTTPRouterInterface {
	return &restPrivateHandler{
		OperatorRestHandler:   NewOperatorRestHandler(),
		ToolBoxRestHandler:    NewToolBoxRestHandler(),
		MCPRestHandler:        NewMCPRestHandler(),
		UpgradeHandler:        common.NewUpgradeHandler(),
		UnifiedProxyHandler:   common.NewUnifiedProxyHandler(),
		Logger:                config.NewConfigLoader().GetLogger(),
		SkillRestHandler:      NewSkillRestHandler(),
		businessDomainService: business_domain.NewBusinessDomainService(),
		Hydra:                 drivenadapters.NewHydra(),
	}
}

// RegisterRouter 内部接口注册路由
func (r *restPrivateHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareHeaderAuthContext(r.Hydra))
	engine.Use(mws...)
	// 算子接口
	r.OperatorRestHandler.RegisterPrivate(engine)
	// 工具箱接口
	r.ToolBoxRestHandler.RegisterPrivate(engine)
	// MCP 相关接口
	r.MCPRestHandler.RegisterPrivate(engine)
	// 技能接口
	r.SkillRestHandler.RegisterPrivate(engine)

	// 临时升级接口 - 仅在从旧版本升级到5.0.0.3时使用
	engine.GET("/upgrade/v5003/migrate-history", r.UpgradeHandler.MigrateHistoryData)
	// 函数沙箱执行
	engine.POST("/function/exec/:version", middlewareBusinessDomain(true, false, r.businessDomainService), r.UnifiedProxyHandler.FunctionExecuteProxy)
}
