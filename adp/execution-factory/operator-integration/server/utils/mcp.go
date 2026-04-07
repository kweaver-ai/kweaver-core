package utils

import "fmt"

// GenerateMCPServerVersion 生成MCP Server版本
func GenerateMCPServerVersion(mcpVersion int) string {
	return fmt.Sprintf("%d.0.0", mcpVersion)
}
