package testhandler

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/httphandler/apiv3common"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/httphandler/testhandler/bizdomain"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
)

type testHTTPHandler struct {
	bizDomainHandler *bizdomain.BizDomainTestHandler
}

func (t *testHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
}

func (t *testHTTPHandler) RegPriRouter(router *gin.RouterGroup) {
	g := apiv3common.GetPrivateRouterGroup(router)

	// 私有路由注册
	// 委托给bizdomain handler注册路由
	t.bizDomainHandler.RegisterRoutes(g)
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewTestHTTPHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &testHTTPHandler{
			bizDomainHandler: bizdomain.NewBizDomainTestHandler(),
		}
	})

	return _handler
}
