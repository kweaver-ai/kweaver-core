package driveradapters

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/mcp"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
)

type MCPRestHandler interface {
	// RegisterPrivate 注册内部API
	RegisterPrivate(engine *gin.RouterGroup)

	// RegisterPublic 注册外部API
	RegisterPublic(engine *gin.RouterGroup)
}

type mcpRestHandler struct {
	MCPPublicHandler      mcp.MCPPublicHandler
	MCPPrivateHandler     mcp.MCPPrivateHandler
	businessDomainService interfaces.IBusinessDomainService
}

var (
	mcpRestHandlerOnce sync.Once
	mHandler           MCPRestHandler
)

func NewMCPRestHandler() MCPRestHandler {
	mcpRestHandlerOnce.Do(func() {
		mHandler = &mcpRestHandler{
			MCPPublicHandler:      mcp.NewMCPHandler(),
			MCPPrivateHandler:     mcp.NewMCPHandler(),
			businessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return mHandler
}

func (r *mcpRestHandler) RegisterPrivate(engine *gin.RouterGroup) {
	mcpGroup := engine.Group("/mcp")

	// MCP 代理相关接口
	mcpProxyGroup := mcpGroup.Group("/proxy")
	// 获取指定MCP Server的工具列表 GET /api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_id}/tools
	mcpProxyGroup.GET("/:mcp_id/tools", r.MCPPrivateHandler.GetMCPTools)
	// 调用指定MCP Server的工具 POST /api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_id}/tool/call
	mcpProxyGroup.POST("/:mcp_id/tool/call", r.MCPPrivateHandler.CallMCPTool)

	// MCP 内置相关接口
	mcpGroup.POST("/intcomp/register", middlewareBusinessDomain(true, true, r.businessDomainService), r.MCPPrivateHandler.RegisterBuiltinMCPServerPrivate)
	mcpGroup.POST("/intcomp/unregister/:mcp_id", middlewareBusinessDomain(true, true, r.businessDomainService), r.MCPPrivateHandler.UnregisterBuiltinMCPServerPrivate)
}

func (r *mcpRestHandler) RegisterPublic(engine *gin.RouterGroup) {
	// MCP 相关接口
	mcpGroup := engine.Group("/mcp")

	// MCP 管理相关接口
	// MCP服务解析 POST /api/agent-operator-integration/v1/mcp/parse/sse
	mcpGroup.POST("/parse/sse", r.MCPPublicHandler.ParseSSE)
	// 添加MCP Server配置 POST /api/agent-operator-integration/v1/mcp
	mcpGroup.POST("/", middlewareBusinessDomain(true, false, r.businessDomainService), r.MCPPublicHandler.AddMCPServer)
	// 删除MCP Server配置 POST /api/agent-operator-integration/v1/mcp/delete
	mcpGroup.DELETE("/:mcp_id", middlewareBusinessDomain(true, false, r.businessDomainService), r.MCPPublicHandler.DeleteMCPServer)
	// 获取MCP Server配置列表 GET /api/agent-operator-integration/v1/mcp/list
	mcpGroup.GET("/list", middlewareBusinessDomain(true, false, r.businessDomainService), r.MCPPublicHandler.QueryMCPServerPage)
	// 获取MCP Server配置详情 GET /api/agent-operator-integration/v1/mcp/{mcp_id}
	mcpGroup.GET("/:mcp_id", r.MCPPublicHandler.QueryMCPServerDetail)
	// 编辑MCP Server配置 POST /api/agent-operator-integration/v1/mcp/{mcp_id}
	mcpGroup.PUT("/:mcp_id", r.MCPPublicHandler.UpdateMCPServer)
	// 更新MCP Server状态 POST /api/agent-operator-integration/v1/mcp/{mcp_id}/status
	mcpGroup.POST("/:mcp_id/status", r.MCPPublicHandler.UpdateMCPServerStatus)
	// MCP工具调试 POST /api/agent-operator-integration/v1/mcp/{mcp_id}/tool/{tool_name}/debug
	mcpGroup.POST("/:mcp_id/tool/:tool_name/debug", r.MCPPublicHandler.DebugTool)

	// MCP服务市场相关接口
	mcpGroup.GET("/market/list", middlewareBusinessDomain(true, false, r.businessDomainService), r.MCPPublicHandler.QueryMCPServerMarketList)
	// 批量查询MCP服务市场详情 GET /api/agent-operator-integration/v1/mcp/market/{mcp_ids}/{fields}
	mcpGroup.GET("/market/batch/:mcp_ids/:fields", middlewareBusinessDomain(true, false, r.businessDomainService), r.MCPPublicHandler.QueryMCPServerMarketBatch)
	mcpGroup.GET("/market/:mcp_id", r.MCPPublicHandler.QueryMCPServerMarketDetail)

	// MCP 代理相关接口
	// 获取指定MCP Server的工具列表 GET /api/agent-operator-integration/v1/mcp/proxy/{mcp_id}/tools
	mcpGroup.GET("/proxy/:mcp_id/tools", r.MCPPublicHandler.GetMCPTools)
	// 调用指定MCP Server的工具 POST /api/agent-operator-integration/v1/mcp/proxy/{mcp_id}/tool/call
	mcpGroup.POST("/proxy/:mcp_id/tool/call", r.MCPPublicHandler.CallMCPTool)

	// MCP endpoint 相关接口
	// Streamable Http Endpoint
	mcpGroup.Any("/app/:mcp_id/mcp", r.MCPPublicHandler.HandleStreamingHttp)
	// SSE Endpoint
	mcpGroup.GET("/app/:mcp_id/sse", r.MCPPublicHandler.HandleServerSentEvents)
	// message endpoint
	mcpGroup.POST("/app/:mcp_id/message", r.MCPPublicHandler.HandleSSEMessage)
}
