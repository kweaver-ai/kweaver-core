package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/mark3labs/mcp-go/mcp"
)

type CallToolRequest struct {
	*ListToolsRequest
	ToolName string         `json:"tool_name"` // 工具名称
	Params   map[string]any `json:"params"`    // 工具参数
}

type CallToolResponse struct {
	interfaces.MCPProxyCallToolResponse
}

type ListToolsRequest struct {
	CreationType interfaces.MCPCreationType
	MCPID        string
	Version      int
	MCPCoreInfo  *interfaces.MCPCoreConfigInfo
}

type ListToolsResponse struct {
	interfaces.MCPProxyToolListResponse
	ServerInitInfo *mcp.InitializeResult
}

// GetMCPTools 获取MCP工具列表
func (s *mcpServiceImpl) GetMCPTools(ctx context.Context, req *interfaces.MCPProxyToolListRequest) (resp *interfaces.MCPProxyToolListResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"mcp_id":  req.MCPID,
		"user_id": req.UserID,
	})
	// 如果是公开接口，检查公开访问或者查看权限，内部接口暂时不校验
	if common.IsPublicAPIFromCtx(ctx) {
		var accessor *interfaces.AuthAccessor
		accessor, err = s.AuthService.GetAccessor(ctx, "")
		if err != nil {
			return
		}
		var authorized bool
		authorized, err = s.AuthService.OperationCheckAny(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP, interfaces.AuthOperationTypeView, interfaces.AuthOperationTypePublicAccess)
		if err != nil {
			return
		}
		if !authorized {
			err = infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonOperationForbidden, nil)
			return
		}
	}

	serverConfig, err := s.DBMCPServerConfig.SelectByID(ctx, nil, req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp server config by id error: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("select mcp server config by id error: %v", err))
		return
	}

	if serverConfig == nil {
		err = infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "mcp server config not found")
		return
	}

	listToolsReq := &ListToolsRequest{
		CreationType: interfaces.MCPCreationType(serverConfig.CreationType),
		MCPID:        req.MCPID,
		Version:      serverConfig.Version,
		MCPCoreInfo: &interfaces.MCPCoreConfigInfo{
			Mode:    interfaces.MCPMode(serverConfig.Mode),
			URL:     serverConfig.URL,
			Headers: utils.JSONToObject[map[string]string](serverConfig.Headers),
		},
	}

	listToolsResp, err := s.listTools(ctx, listToolsReq)
	if err != nil {
		return
	}

	resp = &interfaces.MCPProxyToolListResponse{
		Tools: listToolsResp.Tools,
	}
	return resp, nil
}

// CallMCPTool 调用MCP工具
func (s *mcpServiceImpl) CallMCPTool(ctx context.Context, req *interfaces.MCPProxyCallToolRequest) (resp *interfaces.MCPProxyCallToolResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"mcp_id":  req.MCPID,
		"user_id": req.UserID,
	})
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckExecutePermission(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}
	serverConfig, err := s.DBMCPServerConfig.SelectByID(ctx, nil, req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp server config by id error: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("select mcp server config by id error: %v", err))
		return
	}

	if serverConfig == nil {
		err = infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "mcp server config not found")
		return
	}

	callToolReq := &CallToolRequest{
		ListToolsRequest: &ListToolsRequest{
			CreationType: interfaces.MCPCreationType(serverConfig.CreationType),
			MCPID:        req.MCPID,
			Version:      serverConfig.Version,
			MCPCoreInfo: &interfaces.MCPCoreConfigInfo{
				Mode:    interfaces.MCPMode(serverConfig.Mode),
				URL:     serverConfig.URL,
				Headers: utils.JSONToObject[map[string]string](serverConfig.Headers),
			},
		},
		ToolName: req.ToolName,
		Params:   req.Parameters,
	}

	callToolResult, err := s.callTool(ctx, callToolReq)
	if err != nil {
		return
	}
	// 异步记录审计日志
	go func() {
		accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.logger.WithContext(ctx).Errorf("get account auth context from ctx error")
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationExecute,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectMCP,
				Name: req.ToolName,
				ID:   req.MCPID,
			},
		})
	}()

	resp = &interfaces.MCPProxyCallToolResponse{
		Content: callToolResult.Content,
		IsError: callToolResult.IsError,
	}
	return resp, nil
}

func (s *mcpServiceImpl) callTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
	mcpClient, err := s.getMCPClient(ctx, req.ListToolsRequest)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get mcp client error: %v", err)
		return nil, err
	}
	defer func() {
		if e := mcpClient.Close(); e != nil {
			s.logger.WithContext(ctx).Errorf("close mcp client error: %v", e)
		}
	}()

	callToolRequest := mcp.CallToolRequest{}
	callToolRequest.Params.Name = req.ToolName
	callToolRequest.Params.Arguments = req.Params

	result, err := mcpClient.CallTool(ctx, callToolRequest)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("call mcp tool error: %v", err)
		return nil, infraerrors.NewHTTPError(ctx, http.StatusGatewayTimeout, infraerrors.ErrExtMCPCallToolFailed, err.Error())
	}

	return &CallToolResponse{
		MCPProxyCallToolResponse: interfaces.MCPProxyCallToolResponse{
			Content: result.Content,
			IsError: result.IsError,
		},
	}, nil
}

func (s *mcpServiceImpl) getMCPClient(ctx context.Context, req *ListToolsRequest) (mcpClient interfaces.MCPClient, err error) {
	if req.CreationType != interfaces.MCPCreationTypeToolImported {
		mcpClient, err = drivenadapters.NewMCPClient(ctx, req.MCPCoreInfo)
		return mcpClient, err
	}
	var instance *interfaces.MCPServerInstance
	instance, err = s.MCPInstanceService.GetMCPInstance(ctx, req.MCPID, req.Version)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get mcp instance error: %v", err)
		return nil, err
	}
	mcpClient, err = drivenadapters.NewInProcessMCPClient(ctx, instance.MCPServer)
	return mcpClient, err
}

func (s *mcpServiceImpl) listTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error) {
	mcpClient, err := s.getMCPClient(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get mcp client error: %v", err)
		return nil, err
	}
	defer func() {
		if e := mcpClient.Close(); e != nil {
			s.logger.WithContext(ctx).Errorf("close mcp client error: %v", e)
		}
	}()

	initInfo := mcpClient.GetInitInfo(ctx)

	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, infraerrors.NewHTTPError(ctx, http.StatusGatewayTimeout, infraerrors.ErrExtMCPListToolsFailed, err.Error())
	}

	return &ListToolsResponse{
		MCPProxyToolListResponse: interfaces.MCPProxyToolListResponse{
			Tools: tools.Tools,
		},
		ServerInitInfo: initInfo,
	}, nil
}

// ExecuteTool 执行MCP工具
func (s *mcpServiceImpl) ExecuteTool(ctx context.Context, mcpToolID string, params interfaces.HTTPRequestParams) (*interfaces.HTTPResponse, error) {
	// 获取MCP工具配置信息
	tool, err := s.DBMCPTool.SelectByMCPToolID(ctx, nil, mcpToolID)
	if err != nil {
		s.logger.Warnf("select mcp tool failed, err: %v", err)
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select mcp tool failed")
	}

	if tool == nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "mcp tool not found")
	}

	// 调用工具服务执行工具
	executeToolReq := &interfaces.ExecuteToolReq{
		BoxID:             tool.BoxID,
		ToolID:            tool.ToolID,
		HTTPRequestParams: params,
	}
	executeToolResp, err := s.ToolService.ExecuteToolCore(ctx, executeToolReq)
	if err != nil {
		return nil, err
	}
	return executeToolResp, nil
}
