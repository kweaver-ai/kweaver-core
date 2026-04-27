package deployer

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/mark3labs/mcp-go/server"
)

var (
	sseDeployerOnce     sync.Once
	sseDeployerInstance Deployer
)

type sseDeployer struct {
	prefixURL string
}

// newSSEDeployer 创建 SSE 部署器
func newSSEDeployer() Deployer {
	sseDeployerOnce.Do(func() {
		sseDeployerInstance = &sseDeployer{
			prefixURL: interfaces.AOPInternalV1Prefix,
		}
	})
	return sseDeployerInstance
}

// Deploy 安装
func (d *sseDeployer) Deploy(ctx context.Context, instance *interfaces.MCPServerInstance) error {
	ssePath := fmt.Sprintf("%s/mcp/app/%s/%d/sse", d.prefixURL, instance.Config.MCPID, instance.Config.Version)
	messagePath := fmt.Sprintf("%s/mcp/app/%s/%d/message", d.prefixURL, instance.Config.MCPID, instance.Config.Version)
	sseServer := server.NewSSEServer(
		instance.MCPServer,
		server.WithSSEEndpoint(ssePath),
		server.WithMessageEndpoint(messagePath),
	)
	instance.SSEServer = sseServer
	instance.SSERoutePath = ssePath
	instance.MessageRoutePath = messagePath
	return nil
}

// Undeploy 卸载
func (d *sseDeployer) Undeploy(ctx context.Context, instance *interfaces.MCPServerInstance) error {
	if instance.SSEServer != nil {
		return instance.SSEServer.Shutdown(ctx)
	}
	return nil
}
