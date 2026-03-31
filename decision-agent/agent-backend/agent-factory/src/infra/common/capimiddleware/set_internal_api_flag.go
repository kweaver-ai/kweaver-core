package capimiddleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func SetInternalAPIFlag() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxKey := cenum.InternalAPIFlagCtxKey.String()
		c.Set(ctxKey, true)

		// 设置request context
		_ctx := context.WithValue(c.Request.Context(), ctxKey, true) //nolint:staticcheck // SA1029
		cutil.UpdateGinReqCtx(c, _ctx)

		c.Next()
	}
}

func IsInternalAPI(c *gin.Context) bool {
	return chelper.IsInternalAPIFromCtx(c.Request.Context())
}
