package publishedhandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type publishedHandler struct {
	logger       icmp.Logger
	publishedSvc iv3portdriver.IPublishedSvc
}

func (h *publishedHandler) RegPubRouter(router *gin.RouterGroup) {
	// router.GET("/published/agent", h.PublishedAgentList)
	router.POST("/published/agent", h.PublishedAgentList)

	router.POST("/published/agent-info-list", h.PubedAgentInfoList)

	router.GET("/published/agent-tpl", h.PubedTplList)
	router.GET("/published/agent-tpl/:tpl_id", h.PubedTplDetail)
}

func (h *publishedHandler) RegPriRouter(router *gin.RouterGroup) {
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewPublishedHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &publishedHandler{
			logger:       logger.GetLogger(),
			publishedSvc: dainject.NewPublishedSvc(),
		}
	})

	return _handler
}
