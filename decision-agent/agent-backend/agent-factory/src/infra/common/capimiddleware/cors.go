package capimiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
)

// Cors 跨域
// 【注意】这个目前仅用于在开发和测试时用
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cenvhelper.IsLocalDev(cenvhelper.RunScenario_Aaron_Local_Dev) {
			return
		}

		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")

		if cenvhelper.IsLocalDev(cenvhelper.RunScenario_Aaron_Local_Dev) {
			c.Header("Access-Control-Allow-Headers", "*")
		}

		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}
