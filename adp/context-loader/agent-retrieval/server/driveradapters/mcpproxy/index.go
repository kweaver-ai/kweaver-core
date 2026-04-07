// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mcpproxy provides HTTP proxy handler for MCP tool invocations.
package mcpproxy

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type MCPProxyHandler interface {
	CallMCPTool(c *gin.Context)
}

type mcpProxyHandler struct {
	operatorIntegration interfaces.DrivenOperatorIntegration
}

func NewMCPProxyHandler() MCPProxyHandler {
	return &mcpProxyHandler{
		operatorIntegration: drivenadapters.NewOperatorIntegrationClient(),
	}
}

// CallMCPTool 代理调用 MCP 工具
func (h *mcpProxyHandler) CallMCPTool(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. 获取路径参数
	mcpID := c.Param("mcp_id")
	toolName := c.Param("tool_name")

	if mcpID == "" || toolName == "" {
		rest.ReplyError(c, infraErr.DefaultHTTPError(ctx, http.StatusBadRequest, "mcp_id and tool_name are required"))
		return
	}

	// 2. 解析请求体（扁平化参数）
	var parameters map[string]interface{}
	if err := c.ShouldBindJSON(&parameters); err != nil {
		// 允许空参数 {}
		if err.Error() == "EOF" {
			parameters = make(map[string]interface{})
		} else {
			rest.ReplyError(c, infraErr.DefaultHTTPError(ctx, http.StatusBadRequest, "invalid request body"))
			return
		}
	}

	// 3. 调用 OperatorIntegration
	req := &interfaces.CallMCPToolRequest{
		McpID:      mcpID,
		ToolName:   toolName,
		Parameters: parameters,
	}

	resp, err := h.operatorIntegration.CallMCPTool(ctx, req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 4. 返回结果
	rest.ReplyOK(c, http.StatusOK, resp)
}
