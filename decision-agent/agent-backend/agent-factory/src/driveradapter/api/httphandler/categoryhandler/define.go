package categoryhandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/categorysvc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type categoryHandler struct {
	logger      icmp.Logger
	categorySvc iv3portdriver.ICategorySvc
}

func (h *categoryHandler) RegPubRouter(router *gin.RouterGroup) {
	router.GET("/category", h.List)
}

func (a *categoryHandler) RegPriRouter(router *gin.RouterGroup) {
	// 私有路由注册
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewCategoryHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &categoryHandler{
			logger:      logger.GetLogger(),
			categorySvc: categorysvc.NewCategorySvc(),
		}
	})

	return _handler
}
