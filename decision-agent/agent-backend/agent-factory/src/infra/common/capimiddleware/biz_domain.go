package capimiddleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
)

// BizDomainMiddleware 业务域中间件
// isUseDefault: 是否使用默认业务域（当请求中未携带业务域ID时）
func HandleBizDomain(isUseDefault bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GConfig.IsBizDomainDisabled() {
			c.Next()
			return
		}

		// 1. 从请求上下文中获取业务域ID
		bizDomainID, isExist, err := chelper.GetBizDomainIDFromGinHeader(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 1.1 验证业务域ID是否存在，如果不存在且不使用默认值，则返回错误
		if !isExist && !isUseDefault {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "biz domain id is required"})
			return
		}

		// 1.2 如果业务域ID不存在且使用默认值，则设置为默认业务域
		if !isExist && isUseDefault {
			bizDomainID = cenum.BizDomainPublic.ToString()
		}

		// 2. 将业务域ID设置到上下文中
		ctxKey := cenum.BizDomainIDCtxKey.String()
		c.Set(ctxKey, bizDomainID)

		// 3. 设置request context
		_ctx := context.WithValue(c.Request.Context(), ctxKey, bizDomainID) //nolint:staticcheck // SA1029 使用 cenum 字符串作为 ctx key，后续统一改造
		cutil.UpdateGinReqCtx(c, _ctx)

		// 4. 继续处理请求
		c.Next()
	}
}
