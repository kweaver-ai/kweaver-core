package mcpinstance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type toolManager struct {
	executor interfaces.IMCPToolExecutor
	logger   interfaces.Logger
}

func newToolManager(executor interfaces.IMCPToolExecutor, logger interfaces.Logger) *toolManager {
	return &toolManager{
		executor: executor,
		logger:   logger,
	}
}

// RegisterTools 注册工具
func (tm *toolManager) RegisterTools(tools []*interfaces.MCPToolDeployConfig, mcpServer *server.MCPServer) error {
	for _, tool := range tools {
		rawSchema := tool.InputSchema
		if len(rawSchema) == 0 {
			rawSchema = nil
		}
		toolID := tool.ToolID
		handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return tm.toolHandler(ctx, request, toolID)
		}
		mcpServer.AddTool(mcp.NewToolWithRawSchema(tool.Name, tool.Description, rawSchema), handler)
	}
	return nil
}

// toolHandler 工具处理函数
func (tm *toolManager) toolHandler(ctx context.Context, request mcp.CallToolRequest, mcpToolID string) (*mcp.CallToolResult, error) {
	requestJSON := utils.ObjectToJSON(request)
	if mcpToolID == "" {
		errMsg := fmt.Sprintf("[mcpinstance.tool_manager#toolHandler] unrecognized tool, request: %s", requestJSON)
		tm.logger.Error(errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}

	if tm.executor == nil {
		errMsg := fmt.Sprintf("[mcpinstance.tool_manager#toolHandler] tool executor not configured, mcpToolID: %s", mcpToolID)
		tm.logger.Error(errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}

	var toolArgs interfaces.MCPToolArguments
	if err := request.BindArguments(&toolArgs); err != nil {
		errMsg := fmt.Sprintf("[mcpinstance.tool_manager#toolHandler] bad request, mcpToolID: %s, request: %s, error: %v", mcpToolID, requestJSON, err)
		tm.logger.Error(errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}

	params := interfaces.HTTPRequestParams{
		Headers:     toolArgs.Headers,
		Body:        toolArgs.Body,
		QueryParams: toolArgs.QueryParams,
		PathParams:  make(map[string]string),
	}
	for k, v := range toolArgs.PathParams {
		params.PathParams[k] = toString(v)
	}

	resp, err := tm.executor.ExecuteTool(ctx, mcpToolID, params)
	if err != nil {
		errMsg := fmt.Sprintf("[mcpinstance.tool_manager#toolHandler] execute tool failed, mcpToolID: %s, request: %s, error: %v", mcpToolID, requestJSON, err)
		tm.logger.Error(errMsg)
		return mcp.NewToolResultError(errMsg), nil
	}

	b, _ := json.Marshal(resp)
	return mcp.NewToolResultText(string(b)), nil
}

func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprintf("%v", v)
	}
}
