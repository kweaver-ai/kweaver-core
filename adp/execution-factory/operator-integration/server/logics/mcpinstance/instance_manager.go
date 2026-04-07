package mcpinstance

import (
	"context"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/mcpinstance/deployer"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"github.com/mark3labs/mcp-go/server"
)

// InstanceManager 实例管理器
type InstanceManager struct {
	httpDeployer deployer.Deployer
	sseDeployer  deployer.Deployer
	toolManager  *toolManager
}

func newInstanceManager(executor interfaces.IMCPToolExecutor, logger interfaces.Logger) *InstanceManager {
	return &InstanceManager{
		httpDeployer: deployer.GetDeployer(deployer.StreamDeployerType),
		sseDeployer:  deployer.GetDeployer(deployer.SSEDeployerType),
		toolManager:  newToolManager(executor, logger),
	}
}

// Build 创建 MCP 实例
func (m *InstanceManager) Build(ctx context.Context, cfg *interfaces.MCPRuntimeConfig) (*interfaces.MCPServerInstance, error) {
	now := time.Now()
	instance := &interfaces.MCPServerInstance{
		Config:    cfg,
		CreatedAt: &now,
	}

	mcpServer := server.NewMCPServer(cfg.Name, utils.GenerateMCPServerVersion(cfg.Version), server.WithInstructions(cfg.Instructions))
	instance.MCPServer = mcpServer

	if err := m.toolManager.RegisterTools(cfg.Tools, mcpServer); err != nil {
		return nil, err
	}
	if err := m.httpDeployer.Deploy(ctx, instance); err != nil {
		return nil, err
	}
	if err := m.sseDeployer.Deploy(ctx, instance); err != nil {
		return nil, err
	}
	return instance, nil
}

func (m *InstanceManager) Shutdown(ctx context.Context, instance *interfaces.MCPServerInstance) error {
	if instance == nil {
		return nil
	}
	if err := m.httpDeployer.Undeploy(ctx, instance); err != nil {
		return err
	}
	if err := m.sseDeployer.Undeploy(ctx, instance); err != nil {
		return err
	}
	return nil
}
