package capimiddleware

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		rest.SetLang(cglobal.GConfig.GetDefaultLanguage())

		// 1. 获取language
		language := rest.GetBCP47(c.GetHeader(rest.XLangHeader))

		ctxKey := cenum.VisitLangCtxKey.String()

		// 2. 设置到gin context中
		c.Set(ctxKey, language)

		// 3. 设置request context
		_ctx := context.WithValue(c.Request.Context(), ctxKey, language) //nolint:staticcheck // SA1029
		// 满足mdl中的language ctx设置
		_ctx = context.WithValue(_ctx, rest.XLangKey, language)
		cutil.UpdateGinReqCtx(c, _ctx)

		// 4. 执行后续操作
		c.Next()
	}
}
