package apimiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func IsUserType(t rest.VisitorType) bool {
	return t == rest.VisitorType_User || t == rest.VisitorType_RealName
}

func VisitorTypeCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// /api/agent-factory/v3/agent-permission/execute 这个接口支持应用账号
		if c.Request.URL.Path == "/api/agent-factory/v3/agent-permission/execute" {
			c.Next()
			return
		}

		user := chelper.GetVisitorFromCtx(c)

		if user != nil && !IsUserType(user.Type) {
			httpError := capierr.New403Err(c, "[visitor type is not user] 当前服务的外部接口仅支持实名用户访问，暂不支持应用账号等访问。如有相关需求，请联系我们")
			rest.ReplyError(c, httpError)
			c.Abort()

			return
		}

		c.Next()
	}
}
