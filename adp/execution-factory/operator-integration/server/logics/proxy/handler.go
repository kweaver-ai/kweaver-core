// Package proxy provides a simple HTTP proxy server that can be used to forward requests to a target server.
package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

var (
	syncOnce sync.Once
	s        *ProxyServer
)

// const (
// 	defaultCleanupInterval = 5 * time.Minute
// )

// ProxyServer 代理服务器
type ProxyServer struct {
	Forwarder Forwarder // 转发器
	// Pool      *clientPool // 客户端池
}

// NewProxyServer 创建一个新的代理服务器
func NewProxyServer() interfaces.ProxyHandler {
	syncOnce.Do(func() {
		s = &ProxyServer{
			Forwarder: NewForwarder(),
		}
	})
	return s
}

// HandlerRequest 处理请求
func (s *ProxyServer) HandlerRequest(ctx context.Context, req *interfaces.HTTPRequest) (resp *interfaces.HTTPResponse, err error) {
	// 从上下文获取执行模式，请求头优先
	executionMode := common.GetExecutionModeFromCtx(ctx)
	if executionMode != "" {
		req.ExecutionMode = executionMode
	}
	switch req.ExecutionMode {
	case interfaces.ExecutionModeSync:
		// 验证请求参数
		resp, err = s.Forwarder.Forward(ctx, req)
	case interfaces.ExecutionModeStream:
		// 验证请求参数
		resp, err = s.Forwarder.ForwardStream(ctx, req)
	case interfaces.ExecutionModeAsync:
		// 暂时不支持异步模式
		err = fmt.Errorf("async execution mode is not supported")
	default:
		// 验证请求参数
		resp, err = s.Forwarder.Forward(ctx, req)
	}
	return resp, err
}
