package drivenadapters

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	mcpClientName      = "agent-operator-integration"
	mcpClientVersion   = "1.0.0"
	defaultConnTimeout = 30 * time.Second
)

// MCPClient 定义 MCP 客户端
type MCPClient struct {
	MCPCoreConfigInfo *interfaces.MCPCoreConfigInfo
	client            *client.Client
	serverInitInfo    *mcp.InitializeResult
}

// NewMCPClient 创建 MCP 客户端
func NewMCPClient(ctx context.Context, mcpCoreInfo *interfaces.MCPCoreConfigInfo) (interfaces.MCPClient, error) {
	mcpClient := &MCPClient{
		MCPCoreConfigInfo: mcpCoreInfo,
	}
	if err := mcpClient.initClient(ctx); err != nil {
		return nil, errors.NewHTTPError(ctx, http.StatusGatewayTimeout,
			errors.ErrExtMCPServerNotAccessible,
			fmt.Sprintf("mcp server %s is not accessible, please check if the MCP server is running, error: %v", mcpCoreInfo.URL, err))
	}
	return mcpClient, nil
}

// NewInProcessMCPClient 创建进程内 MCP 客户端
func NewInProcessMCPClient(ctx context.Context, server *server.MCPServer) (interfaces.MCPClient, error) {
	mcpClient := &MCPClient{
		MCPCoreConfigInfo: &interfaces.MCPCoreConfigInfo{
			Mode: interfaces.MCPModeStream, // In-process behaves like a stream
			URL:  "in-process",
		},
	}
	cli, err := client.NewInProcessClient(server)
	if err != nil {
		return nil, err
	}
	mcpClient.client = cli

	if err := mcpClient.performHandshake(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize in-process client: %w", err)
	}

	return mcpClient, nil
}

// ListTools 列出工具
func (m *MCPClient) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	serverCapabilities := m.client.GetServerCapabilities()
	if serverCapabilities.Tools == nil {
		return &mcp.ListToolsResult{
			Tools: []mcp.Tool{},
		}, nil
	}

	result, err := m.client.ListTools(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("[mcpClient.ListTools] failed to list tools:\n %v", err)
	}
	return result, nil
}

// CallTool 调用工具
func (m *MCPClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := m.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}
	return result, nil
}

func (m *MCPClient) initClient(ctx context.Context) error {
	mode := m.MCPCoreConfigInfo.Mode
	var cli *client.Client
	var err error

	switch mode {
	case interfaces.MCPModeSSE:
		httpClient := m.createHTTPClient()
		cli, err = client.NewSSEMCPClient(m.MCPCoreConfigInfo.URL, client.WithHeaders(m.MCPCoreConfigInfo.Headers), client.WithHTTPClient(httpClient))
	case interfaces.MCPModeStream:
		httpClient := m.createHTTPClient()
		cli, err = client.NewStreamableHttpClient(m.MCPCoreConfigInfo.URL, transport.WithHTTPHeaders(m.MCPCoreConfigInfo.Headers), transport.WithHTTPBasicClient(httpClient))
	default:
		return errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtMCPModeNotSupported,
			fmt.Sprintf("MCP mode %s is not supported", mode), mode)
	}

	if err != nil {
		err = fmt.Errorf("[mcp.initClient] failed to create %s MCP client:\n %v", mode, err)
		return err
	}

	m.client = cli
	return m.performHandshake(ctx)
}

var (
	httpClient     *http.Client
	httpClientOnce sync.Once
)

func (m *MCPClient) createHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		connTimeout := time.Duration(config.NewConfigLoader().MCPConfig.ConnTimeout) * time.Second
		if connTimeout <= 0 {
			connTimeout = defaultConnTimeout
		}
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				DialContext: (&net.Dialer{
					Timeout:   connTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 50,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	})
	return httpClient
}

// performHandshake 执行 MCP 握手
func (m *MCPClient) performHandshake(ctx context.Context) error {
	// 1. Start connection
	if err := m.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	// 2. Initialize
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.Capabilities = mcp.ClientCapabilities{}
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    mcpClientName,
		Version: mcpClientVersion,
	}

	initResp, err := m.client.Initialize(ctx, initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}
	m.serverInitInfo = initResp
	// 3. Ping
	if err := m.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping MCP client: %w", err)
	}
	return nil
}

// GetInitInfo 获取初始化信息
func (m *MCPClient) GetInitInfo(ctx context.Context) *mcp.InitializeResult {
	return m.serverInitInfo
}

func (m *MCPClient) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}
