// Package policy set policy
package policy

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/policy"
	"github.com/gin-gonic/gin"
)

type policyReq struct {
	// Status 开关状态
	Status bool `json:"status"`
}

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once             sync.Once
	rh               RESTHandler
	policySchema     = "base/policy.json"
	permPolicySchema = "base/minperm_check.json"
)

type restHandler struct {
	hydra          drivenadapters.HydraAdmin
	userManagement drivenadapters.UserManagement
	policy         policy.Handler
	permPolicy     perm.PermPolicyHandler
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			hydra:          drivenadapters.NewHydraAdmin(),
			userManagement: drivenadapters.NewUserManagement(),
			policy:         policy.NewPolicy(),
			permPolicy:     perm.NewPermPolicy(),
		}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/switch", middleware.TokenAuth(), h.getServiceSwitch)
	engine.PUT("/switch", middleware.TokenAuth(), middleware.CheckAdmin(), h.setServiceSwitch)
	engine.POST("/permissions/check", middleware.TokenAuth(), h.minPermList)
}

func (h *restHandler) getServiceSwitch(c *gin.Context) {
	_, isexist := c.Get("user")
	if !isexist {
		c.JSON(http.StatusUnauthorized, errors.NewIError(errors.UnAuthorization, "", map[string]string{"auth": "user info not found"}))
		return
	}

	status, err := h.policy.CheckStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"enable": status})
}

func (h *restHandler) setServiceSwitch(c *gin.Context) {
	_, isexist := c.Get("user")
	if !isexist {
		c.JSON(http.StatusUnauthorized, errors.NewIError(errors.UnAuthorization, "", map[string]string{"auth": "user info not found"}))
		return
	}

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, policySchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	var param policyReq
	err = json.Unmarshal(data, &param)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]string{"param": err.Error()}))
		return
	}
	err = h.policy.SetStatus(param.Status)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *restHandler) minPermList(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, permPolicySchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param perm.MinPermCheckReq
	err = json.Unmarshal(data, &param)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]string{"param": err.Error()}))
		return
	}

	permList, err := h.permPolicy.MinPermList(c.Request.Context(), userInfo.UserID, userInfo.AccountType, param.ResourceIDs)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"perms": permList})
}
