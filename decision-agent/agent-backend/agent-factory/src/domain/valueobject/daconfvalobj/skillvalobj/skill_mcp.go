package skillvalobj

import (
	"github.com/pkg/errors"
)

// Tool 表示工具配置
type SkillMCP struct {
	MCPServerID string `json:"mcp_server_id" binding:"required"` // MCP Server ID
}

// ValObjCheck 验证工具配置
func (p *SkillMCP) ValObjCheck() (err error) {
	// 检查MCPServerID是否为空
	if p.MCPServerID == "" {
		err = errors.New("[Tool]: mcp_server_id is required")
		return
	}

	return
}
