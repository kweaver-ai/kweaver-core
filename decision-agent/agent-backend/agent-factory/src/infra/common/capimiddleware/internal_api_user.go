package capimiddleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// SetInternalAPIUserInfo 设置内部api用户信息
// 逻辑：
// 1. 从header中获取account-id和account-type
// 2. 如果account-type不为空，并且account-type不在supportAccountTypes中，返回401
// 3. 如果account-id为空，不设置visitor信息到context，继续执行
// 4. 设置visitor信息到context
// supportAccountTypes: 支持的账户类型，默认为[cenum.AccountTypeUser]
func SetInternalAPIUserInfo(isCheckAccountType bool, supportAccountTypes ...cenum.AccountType) gin.HandlerFunc {
	// 如果没有传入supportAccountTypes，默认为[cenum.AccountTypeUser]
	if len(supportAccountTypes) == 0 {
		supportAccountTypes = []cenum.AccountType{cenum.AccountTypeUser}
	}

	return func(c *gin.Context) {
		// log.Println("in SetInternalAPIUserInfo...")
		// 1. 从header中获取account-id和account-type
		uid, isExist, err := chelper.GetAccountIDFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 1.1 如果account-id为空，不设置visitor信息到context，继续执行（一些场景下不需要account-id等）
		if !isExist {
			c.Next()
			return
		}

		// 2. 从header中获取account-type
		accountType, isExist, err := chelper.GetAccountTypeFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 2.1 如果account-type不为空，检查是否在supportAccountTypes中
		if isExist {
			if isCheckAccountType {
				isSupported := false

				for _, supportedType := range supportAccountTypes {
					if accountType == supportedType {
						isSupported = true
						break
					}
				}

				if !isSupported {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid account type"})
					return
				}
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account type is required when account-id is not empty"})
			return
		}

		if cenvhelper.IsAaronLocalDev() {
			uid = "mock_id"
		}

		// 3. 设置visitor信息到context
		visitor := rest.Visitor{
			ID:   uid,
			Type: accountType.ToMDLVisitorType(),
		}

		ctxKey := cenum.VisitUserInfoCtxKey.String()
		c.Set(ctxKey, &visitor)

		// 设置request context
		_ctx := context.WithValue(c.Request.Context(), ctxKey, &visitor) //nolint:staticcheck // SA1029
		cutil.UpdateGinReqCtx(c, _ctx)

		// 4. 继续执行
		c.Next()
	}
}
