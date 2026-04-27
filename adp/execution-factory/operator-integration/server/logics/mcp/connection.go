package mcp

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

const (
	// 对外MCP Server Stream URL
	externalMCPStreamURI = "/api/agent-operator-integration/v1/mcp/app/:mcp_id/mcp"
	// 对外MCP Server SSE URL
	externalMCPSSEURI = "/api/agent-operator-integration/v1/mcp/app/:mcp_id/sse"
	// internalMCPStreamURL 内部MCP流式URL, 参数为mcpID
	internalMCPStreamURI = "/api/agent-operator-integration/internal-v1/mcp/app/%s/%d/stream"
	// internalMCPSSEURL 内部MCP SSE URL, 参数为mcpID
	internalMCPSSEURI = "/api/agent-operator-integration/internal-v1/mcp/app/%s/%d/sse"
)

// generateExternalConnectionInfo 生成对外MCP Server连接信息
func (s *mcpServiceImpl) generateExternalConnectionInfo(mcpID string,
	mode interfaces.MCPMode,
	creationType interfaces.MCPCreationType,
) (connectionInfo *interfaces.MCPConnectionInfo) {
	connectionInfo = &interfaces.MCPConnectionInfo{}
	switch creationType {
	case interfaces.MCPCreationTypeToolImported:
		// 生成SSE URL
		connectionInfo.SSEURL = strings.NewReplacer(
			":mcp_id", mcpID,
		).Replace(externalMCPSSEURI)

		// 生成Stream URL
		connectionInfo.StreamURL = strings.NewReplacer(
			":mcp_id", mcpID,
		).Replace(externalMCPStreamURI)
	case interfaces.MCPCreationTypeCustom:
		// 如果Mode为stream, 则使用stream url
		if mode == interfaces.MCPModeStream {
			connectionInfo.StreamURL = strings.NewReplacer(
				":mcp_id", mcpID,
			).Replace(externalMCPStreamURI)
		}
		// 如果mode为sse, 则使用sse url
		if mode == interfaces.MCPModeSSE {
			connectionInfo.SSEURL = strings.NewReplacer(
				":mcp_id", mcpID,
			).Replace(externalMCPSSEURI)
		}
	default:
		// 如果Mode为stream, 则使用stream url
		if mode == interfaces.MCPModeStream {
			connectionInfo.StreamURL = strings.NewReplacer(
				":mcp_id", mcpID,
			).Replace(externalMCPStreamURI)
		}
		// 如果mode为sse, 则使用sse url
		if mode == interfaces.MCPModeSSE {
			connectionInfo.SSEURL = strings.NewReplacer(
				":mcp_id", mcpID,
			).Replace(externalMCPSSEURI)
		}
	}
	return
}

// generateInternalConnectionInfo 生成内部MCP Server连接信息
func (s *mcpServiceImpl) generateInternalMCPURL(mcpID string,
	mcpVersion int,
	mode interfaces.MCPMode,
) (url string) {
	baseURL := strings.TrimRight(interfaces.AOIServerURL, "/")
	switch mode {
	case interfaces.MCPModeStream:
		url = fmt.Sprintf("%s%s", baseURL, fmt.Sprintf(internalMCPStreamURI, mcpID, mcpVersion))
	case interfaces.MCPModeSSE:
		url = fmt.Sprintf("%s%s", baseURL, fmt.Sprintf(internalMCPSSEURI, mcpID, mcpVersion))
	case interfaces.MCPModeStdioNpx, interfaces.MCPModeStdioUv:
	}
	return
}
