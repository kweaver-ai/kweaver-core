package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
)

//go:generate mockgen -package mock -source httpserver_interface.go -destination ../mock/mock_httpserver.go

// HTTPServerStarter 定义 HTTP 服务器启动接口
type HTTPServerStarter interface {
	Start(svr common.ServerInfo, handlers map[string][]map[string]func(c *gin.Context)) error
}

// DefaultHTTPServerStarter 默认的 HTTP 服务器启动器实现
type DefaultHTTPServerStarter struct{}

// NewHTTPServerStarter 创建默认的 HTTP 服务器启动器
func NewHTTPServerStarter() HTTPServerStarter {
	return &DefaultHTTPServerStarter{}
}

// Start 实现 HTTPServerStarter 接口
func (s *DefaultHTTPServerStarter) Start(svr common.ServerInfo, handlers map[string][]map[string]func(c *gin.Context)) error {
	return NewHTTPServer(svr, handlers)
}
