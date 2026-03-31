package deployer

import (
	"context"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// Deployer is the interface for deploying and undeploying MCP server instances.
type Deployer interface {
	Deploy(ctx context.Context, instance *interfaces.MCPServerInstance) error
	Undeploy(ctx context.Context, instance *interfaces.MCPServerInstance) error
}

// DeployerType is the type of deployer.
type DeployerType string

const (
	SSEDeployerType    DeployerType = "sse"    // SSEDeployerType is the deployer type for SSE server.
	StreamDeployerType DeployerType = "stream" // StreamDeployerType is the deployer type for stream server.
)

// GetDeployer returns the deployer for the given deployer type.
func GetDeployer(deployerType DeployerType) Deployer {
	switch deployerType {
	case SSEDeployerType:
		return newSSEDeployer()
	case StreamDeployerType:
		return newStreamableHTTPDeployer()
	default:
		return nil
	}
}
