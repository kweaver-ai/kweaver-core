package isandboxhtpp

import (
	"context"

	sandboxdto "github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
)

// ISandboxPlatform Sandbox Platform 接口
// 根据 OpenAPI 3.1.0 规范定义
type ISandboxPlatform interface {
	// CreateSession 创建 Sandbox Session
	CreateSession(ctx context.Context, req sandboxdto.CreateSessionReq) (*sandboxdto.CreateSessionResp, error)
	// GetSession 获取 Sandbox Session 信息
	GetSession(ctx context.Context, sessionID string) (*sandboxdto.GetSessionResp, error)
	// DeleteSession 终止 Sandbox Session
	DeleteSession(ctx context.Context, sessionID string) error
	// ListFiles 列出 Session workspace 下的所有文件
	// limit: 最大返回文件数 (1-10000)，默认 1000
	ListFiles(ctx context.Context, sessionID string, limit int) ([]string, error)
}
