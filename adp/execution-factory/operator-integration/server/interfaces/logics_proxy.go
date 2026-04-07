package interfaces

import (
	"context"
	"io"
)

// ProxyHandler 代理处理器
//
//go:generate mockgen -source=logics_proxy.go -destination=../mocks/logics_proxy.go -package=mocks
type ProxyHandler interface {
	HandlerRequest(ctx context.Context, req *HTTPRequest) (resp *HTTPResponse, err error)
}

// IOutboxMessageEvent 消息事件管理
type IOutboxMessageEvent interface {
	Publish(ctx context.Context, req *OutboxMessageReq) (err error)
}

// Forwarder 转发器接口
type Forwarder interface {
	Forward(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)
	ForwardStream(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)
}

// StreamProcessor 流式处理器接口
type StreamProcessor interface {
	ProcessSSE(ctx context.Context, reader io.Reader, writer io.Writer) error
	ProcessHTTPStream(ctx context.Context, reader io.Reader, writer io.Writer) error
}

// FunctionProxyExecuteCodeReq 函数代理执行代码请求
type FunctionProxyExecuteCodeReq struct {
	Code            string            `json:"code" validate:"required"`                                      // 执行代码
	Event           map[string]any    `json:"event" validate:"required"`                                     // 事件
	Language        string            `json:"language" default:"python"`                                     // 执行语言
	Timeout         int               `json:"timeout,omitempty"`                                             // 超时时间，单位秒
	Dependencies    []*DependencyInfo `json:"dependencies,omitempty"`                                        // 依赖资源
	DependenciesURL string            `json:"dependencies_url,omitempty" default:"https://pypi.org/simple/"` // 安装源URL
}
