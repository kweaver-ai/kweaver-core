package config

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/config"
	"github.com/gin-gonic/gin"
)

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	// RegisterPrivateAPI 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
	// RegisterPrivateAPI 注册内部API
	RegisterPrivateAPI(engine *gin.RouterGroup)
}

var (
	once         sync.Once
	rh           RESTHandler
	updateSchema = "base/config.json"
)

type restHandler struct {
	config     config.ConfigHandler
	hydraAdmin drivenadapters.HydraAdmin
}

// NewRESTHandler 创建RESTHandler
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			config:     config.NewConfig(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
		}
	})
	return rh
}

func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/configs", middleware.TokenAuth(), h.listConfigs)
}

func (h *restHandler) RegisterPrivateAPI(engine *gin.RouterGroup) {
	engine.PUT("/config", h.updateConfigs)
}

func (h *restHandler) listConfigs(c *gin.Context) {
	key := c.Query("key")

	res, err := h.config.ListConfigs(c.Request.Context(), key)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": res,
	})
}

func (h *restHandler) updateConfigs(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, updateSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param config.ConfigReq
	err = json.Unmarshal(data, &param)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.config.UpdateConfig(c.Request.Context(), &param)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
