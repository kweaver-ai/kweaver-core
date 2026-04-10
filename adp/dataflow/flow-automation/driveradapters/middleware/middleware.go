// Package middleware validate auth
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

var (
	hydra          drivenadapters.HydraAdmin
	userManagement drivenadapters.UserManagement
	logger         commonLog.Logger
)

func SetMiddleware() {
	hydra = drivenadapters.NewHydraAdmin()
	userManagement = drivenadapters.NewUserManagement()
	logger = commonLog.NewLogger()
}

func SetMiddlewareMock(hydraMock drivenadapters.HydraAdmin, userManagementMock drivenadapters.UserManagement) {
	hydra = hydraMock
	userManagement = userManagementMock
}

// TokenAuth 验证token
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userInfo = &drivenadapters.UserInfo{}
		userInfo.TokenID = c.GetHeader("Authorization")
		userInfo.UserAgent = c.GetHeader("User-Agent")
		userInfo.Mac = c.GetHeader("X-Request-MAC")

		res, err := hydra.Introspect(c.Request.Context(), strings.TrimPrefix(userInfo.TokenID, "Bearer "))
		if err != nil {
			var code = errors.InternalError
			c.JSON(http.StatusUnauthorized, errors.NewIError(code, "", map[string]string{"hydra": err.Error()}))
			traceLog.WithContext(c.Request.Context()).Warnf(err.Error())
			c.Abort()
			return
		}

		if !res.Active {
			var code = errors.UnAuthorization

			c.JSON(http.StatusUnauthorized, errors.NewIError(code, "", map[string]string{"auth": "token expired"}))
			err = errors.NewIError(code, "", map[string]string{"auth": "token expired"})
			c.Abort()
			return
		}
		userInfo.UdID = res.UdID
		userInfo.LoginIP = res.LoginIP
		userInfo.UserID = res.UserID
		userInfo.ClientType = res.ClientType
		userInfo.AccountType = common.User.ToString()
		userInfo.ExpiresIn = res.ExpiresIn

		// 应用账户调用ClientID与UserID相同
		if res.ClientID == res.UserID {
			userInfo.AccountType = common.APP.ToString()
			userInfo.VisitorType = common.APP.ToString()
			c.Set("user", userInfo)
			c.Next()
			return
		}

		switch res.VisitorType {
		case "realname":
			userInfo.VisitorType = common.AuthenticatedUserType
		case "anonymous":
			userInfo.VisitorType = common.AnonymousUserType
			c.Set("user", userInfo)
			c.Next()
			return
		}

		userDetail, err := userManagement.GetUserInfo(userInfo.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", map[string]string{"err": err.Error()}))
			traceLog.WithContext(c.Request.Context()).Warnf(err.Error())
			c.Abort()
			return
		}
		userInfo.UserName = userDetail.UserName
		userInfo.Roles = userDetail.Roles
		userInfo.IsKnowledge = userDetail.IsKnowledge
		userInfo.DepartmentPaths = userDetail.DepartmentPaths
		userInfo.DepartmentNames = userDetail.DepartmentNames

		c.Set("user", userInfo)
		c.Next()
	}
}

// CheckAdmin 验证token
func CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", map[string]string{"err": "userinfo not exist"}))
			c.Abort()
			return
		}
		userInfo := user.(*drivenadapters.UserInfo)
		var isAdmin = utils.IsAdminRole(userInfo.Roles)
		if !isAdmin {
			c.JSON(http.StatusForbidden, errors.NewIError(errors.NoPermission, "", map[string]string{"user": userInfo.UserID}))
			logger.Warnf("[TokenAuth] user %s is not a system administrator", userInfo.UserID)
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckAdminOrKnowledgeManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", map[string]string{"err": "userinfo not exist"}))
			c.Abort()
			return
		}
		userInfo := user.(*drivenadapters.UserInfo)
		var isAdmin = utils.IsAdminRole(userInfo.Roles)

		if !(isAdmin || userInfo.IsKnowledge) {
			c.JSON(http.StatusForbidden, errors.NewIError(errors.NoPermission, "", map[string]string{"user": userInfo.UserID}))
			logger.Warnf("[TokenAuth] user %s is not a system administrator or knowledge manager", userInfo.UserID)
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckExecutorWhiteList(appstore drivenadapters.Appstore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", map[string]string{"err": "userinfo not exist"}))
			c.Abort()
			return
		}
		userInfo := user.(*drivenadapters.UserInfo)

		result, err := appstore.GetWhiteListStatus(c.Request.Context(), common.ExecutorWhiteListKey, userInfo.TokenID)

		if err != nil {
			c.JSON(http.StatusForbidden, errors.NewIError(errors.Forbidden, "", []interface{}{err.Error()}))
			c.Abort()
			return
		}

		if result["enable"] != true {
			c.JSON(http.StatusForbidden, errors.NewIError(errors.Forbidden, "", []interface{}{}))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckBizDomainID checks if the request header contains a valid X-Business-Domain
// It will return a 400 error if the header is empty or missing.
func CheckBizDomainID() gin.HandlerFunc {
	return func(c *gin.Context) {
		bizDomainID := c.Request.Header.Get("X-Business-Domain")
		isBizDomainEnabled := common.NewConfig().IsBusinessDomainEnabled()
		if strings.TrimSpace(bizDomainID) == "" && isBizDomainEnabled {
			c.JSON(http.StatusBadRequest, ierr.NewPublicRestError(c.Request.Context(), ierr.PErrorBadRequest, ierr.PErrorBadRequest, "X-Business-Domain Header is required"))
			c.Abort()
			return
		}

		c.Set("bizDomainID", bizDomainID)
		c.Next()
	}
}

// CheckIsApp 检查用户是否为应用账户，如果为应用账户禁止创建数据流和工作流
func CheckIsApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", map[string]string{"err": "userinfo not exist"}))
			c.Abort()
			return
		}
		userInfo := user.(*drivenadapters.UserInfo)
		if userInfo.AccountType == common.APP.ToString() {
			c.JSON(http.StatusForbidden, errors.NewIError(errors.NoPermission, "", map[string]string{"user": userInfo.UserID, "type": common.APP.ToString()}))
			c.Abort()
			return
		}
		c.Next()
	}
}
