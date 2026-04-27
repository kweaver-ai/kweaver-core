package bizdomain

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/bizdomainsvc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/inject/v3/dainject"
)

// BizDomainTestHandler 业务域测试handler
type BizDomainTestHandler struct {
	bizDomainSvc *bizdomainsvc.BizDomainSvc
}

var (
	handlerOnce sync.Once
	_handler    *BizDomainTestHandler
)

// NewBizDomainTestHandler 创建业务域测试handler
func NewBizDomainTestHandler() *BizDomainTestHandler {
	handlerOnce.Do(func() {
		_handler = &BizDomainTestHandler{
			bizDomainSvc: dainject.NewBizDomainSvc(),
		}
	})

	return _handler
}

// RegisterRoutes 注册路由
func (h *BizDomainTestHandler) RegisterRoutes(router *gin.RouterGroup) {
	// https://{host}:{port}/api/agent-factory/internal/v3/test/bizdomain/query-resource-associations
	router.POST("/test/bizdomain/query-resource-associations", h.QueryResourceAssociationsTestHandler)
}
