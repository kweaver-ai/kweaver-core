package httpserver

import (
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/httphandler/personalspacehandler"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
)

// httpServer HTTP服务器结构体
type httpServer struct {
	// HTTP 服务器实例
	httpSrv *http.Server

	// ========== 健康检查 ==========
	httpHealthHandler ihandlerportdriver.IHTTPHealthRouter

	// ========== Management侧 (V3) ==========
	// agent 配置
	v3AgentConfigHandler ihandlerportdriver.IHTTPRouter
	// agent 模板
	v3AgentTplHandler ihandlerportdriver.IHTTPRouter
	// 产品相关接口
	productHandler ihandlerportdriver.IHTTPRouter
	// 分类相关接口
	categoryHandler ihandlerportdriver.IHTTPRouter
	// 发布相关接口
	releaseHandler ihandlerportdriver.IHTTPRouter
	// agent 广场相关接口
	squareHandler ihandlerportdriver.IHTTPRouter
	// 权限相关接口
	permissionHandler ihandlerportdriver.IHTTPRouter
	// 个人空间相关接口
	personalSpaceHandler *personalspacehandler.PersonalSpaceHTTPHandler
	// 发布相关接口
	publishedHandler ihandlerportdriver.IHTTPRouter
	// other
	otherHandler ihandlerportdriver.IHTTPRouter
	// test
	testHandler ihandlerportdriver.IHTTPRouter
	// anyshare 文档库代理接口
	anysharedsHandler ihandlerportdriver.IHTTPRouter

	// ========== Run侧 (V1) ==========
	// agent 对话
	agentHandler ihandlerportdriver.IHTTPRouter
	// conversation 会话管理
	conversationHandler ihandlerportdriver.IHTTPRouter
	// temparea 临时区域
	// observability 可观测性
	observabilityHandler ihandlerportdriver.IHTTPRouter
	// session 会话
	sessionHandler ihandlerportdriver.IHTTPRouter
}
