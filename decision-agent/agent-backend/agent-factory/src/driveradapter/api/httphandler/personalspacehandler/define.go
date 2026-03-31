package personalspacehandler

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	personalSpaceHandlerOnce sync.Once
	personalSpaceHandlerImpl *PersonalSpaceHTTPHandler
)

// PersonalSpaceHTTPHandler 个人空间HTTP处理器
type PersonalSpaceHTTPHandler struct {
	personalSpaceService iv3portdriver.IPersonalSpaceService
	logger               icmp.Logger
}

// GetPersonalSpaceHTTPHandler 获取个人空间HTTP处理器实例
func GetPersonalSpaceHTTPHandler() *PersonalSpaceHTTPHandler {
	personalSpaceHandlerOnce.Do(func() {
		personalSpaceHandlerImpl = &PersonalSpaceHTTPHandler{
			personalSpaceService: dainject.NewPersonalSpaceSvc(),
			logger:               logger.GetLogger(),
		}
	})

	return personalSpaceHandlerImpl
}

// RegPubRouter 注册公共路由
func (h *PersonalSpaceHTTPHandler) RegPubRouter(router *gin.RouterGroup) {
	// 个人空间Agent模板列表
	router.GET("/personal-space/agent-tpl-list", h.AgentTplList)

	// 个人空间Agent列表
	router.GET("/personal-space/agent-list", h.AgentList)
}
