package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/admin"
	"github.com/gin-gonic/gin"
)

type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once           sync.Once
	rh             RESTHandler
	addAdminSchema = "base/admin.json"
)

type restHandler struct {
	hydraAdmin     drivenadapters.HydraAdmin
	userManagement drivenadapters.UserManagement
	adminHandler   admin.AdminHandler
}

// CheckAuthReq CheckAuthReq struct
type CheckAuthReq struct {
	RedirectURI string `json:"redirect_uri"`
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			hydraAdmin:     drivenadapters.NewHydraAdmin(),
			userManagement: drivenadapters.NewUserManagement(),
			adminHandler:   admin.NewAdmin(),
		}
	})

	return rh
}

// RegisterAPI implements RESTHandler.
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/admins", middleware.TokenAuth(), middleware.CheckAdmin(), h.listAdmins)
	engine.POST("/admin", middleware.TokenAuth(), middleware.CheckAdmin(), h.addAdmin)
	engine.DELETE("/admins/:id", middleware.TokenAuth(), middleware.CheckAdmin(), h.deleteAdmin)
	engine.GET("/admins/:user_id/is-admin", middleware.TokenAuth(), h.isAdmin)
}

func (h *restHandler) listAdmins(c *gin.Context) {
	admins, err := h.adminHandler.ListAdmins(c.Request.Context())
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"admins": admins})
}

func (h *restHandler) addAdmin(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)
	err := common.JSONSchemaValid(data, addAdminSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param admin.AdminReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.adminHandler.CreateAdmin(c.Request.Context(), &param)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) deleteAdmin(c *gin.Context) {
	ID := c.Param("id")
	err := h.adminHandler.DeleteAdmin(c.Request.Context(), ID)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) isAdmin(c *gin.Context) {
	userID := c.Param("user_id")
	isAdmin, err := h.adminHandler.IsAdmin(c.Request.Context(), userID)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_admin": isAdmin})
}
