package deployer

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/mark3labs/mcp-go/server"
)

var (
	httpDeployerOnce sync.Once
	httpDeployer     Deployer
)

// streamableHTTPDeployer 可流式传输的 HTTP 部署器
type streamableHTTPDeployer struct{}

// newStreamableHTTPDeployer 创建可流式传输的 HTTP 部署器
func newStreamableHTTPDeployer() Deployer {
	httpDeployerOnce.Do(func() {
		httpDeployer = &streamableHTTPDeployer{}
	})
	return httpDeployer
}

// Deploy 部署 MCP 实例
func (d *streamableHTTPDeployer) Deploy(ctx context.Context, instance *interfaces.MCPServerInstance) error {
	streamPath := fmt.Sprintf("/app/%s/%d/stream", instance.Config.MCPID, instance.Config.Version)
	streamServer := server.NewStreamableHTTPServer(instance.MCPServer, server.WithEndpointPath(streamPath))
	instance.StreamServer = streamServer
	instance.StreamRoutePath = streamPath
	return nil
}

// Undeploy 卸载
func (d *streamableHTTPDeployer) Undeploy(ctx context.Context, instance *interfaces.MCPServerInstance) error {
	if instance.StreamServer != nil {
		return instance.StreamServer.Shutdown(ctx)
	}
	return nil
}
