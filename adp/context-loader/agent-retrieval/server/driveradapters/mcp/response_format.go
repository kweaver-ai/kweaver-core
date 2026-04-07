// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

// GetResponseFormatFromRequest 从 MCP CallToolRequest 的 arguments 中解析 response_format，未传时默认 toon
func GetResponseFormatFromRequest(req mcp.CallToolRequest) (rest.ResponseFormat, error) {
	s := req.GetString("response_format", "toon")
	return rest.ParseResponseFormat(s)
}

// BuildMCPToolResult 根据 response_format 统一构造 MCP Tool 返回结果（文本为 JSON 或 TOON，structuredContent 仍为原对象）
func BuildMCPToolResult(resp interface{}, format rest.ResponseFormat) (*mcp.CallToolResult, error) {
	var textContent string
	if format == rest.FormatTOON {
		_, bodyBytes, err := rest.MarshalResponse(rest.FormatTOON, resp)
		if err != nil {
			return nil, err
		}
		textContent = string(bodyBytes)
	} else {
		textContent = utils.ObjectToJSON(resp)
	}
	return mcp.NewToolResultStructured(resp, textContent), nil
}
