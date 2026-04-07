package permissionhandler

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/httphandler/apiv3common"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
)

type permissionHandler struct {
	logger        icmp.Logger
	permissionSvc iv3portdriver.IPermissionSvc
}

func (h *permissionHandler) RegPubRouter(router *gin.RouterGroup) {
	// 权限相关路由
	router.POST("/agent-permission/execute", h.CheckUsePermission)
	// router.POST("/agent-permission/is-custom-space-member", h.CheckIsCustomSpaceMember)

	router.GET("/agent-permission/management/user-status", h.GetUserStatus)
}

func (h *permissionHandler) RegPriRouter(router *gin.RouterGroup) {
	g := apiv3common.GetPrivateRouterGroupWithAccountTypes(router, cenum.AccountTypeUser, cenum.AccountTypeApp)

	// 私有路由注册
	g.POST("/agent-permission/execute", h.CheckUsePermission)
	// g.POST("/agent-permission/is-custom-space-member", h.CheckIsCustomSpaceMember)

	g.GET("/agent-permission/management/user-status", h.GetUserStatus)
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewPermissionHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &permissionHandler{
			logger:        logger.GetLogger(),
			permissionSvc: dainject.NewPermissionSvc(),
		}
	})

	return _handler
}
