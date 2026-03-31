package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/executor"
	"github.com/gin-gonic/gin"
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

const (
	CreateExecutorSchema            = "executors/create-executor.json"
	UpdateExecutorSchema            = "executors/update-executor.json"
	CreateExecutorActionSchema      = "executors/create-executor-action.json"
	UpdateExecutorActionSchema      = "executors/update-executor-action.json"
	CheckCreateExecutorSchema       = "executors/check-create-executor.json"
	CheckUpdateExecutorSchema       = "executors/check-update-executor.json"
	CheckCreateExecutorActionSchema = "executors/check-create-executor-action.json"
	CheckUpdateExecutorActionSchema = "executors/check-update-executor-action.json"
	ImportAgentsSchema              = "executors/import-agents.json"
)

type restHandler struct {
	hydraAdmin drivenadapters.HydraAdmin
	executor   executor.ExecutorHandler
	appstore   drivenadapters.Appstore
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			executor:   executor.NewExecutorHandler(),
			appstore:   drivenadapters.NewAppStore(),
		}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/executors", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.createExecutor)
	engine.GET("/executors", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.getExecutors)
	engine.GET("/executors/:id", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.getExecutor)
	engine.PUT("/executors/:id", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.updateExecutor)
	engine.DELETE("/executors/:id", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.deleteExecutor)
	engine.POST("/executors/:id/actions", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.createExecutorAction)
	engine.PUT("/executors/:id/actions/:actionId", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.updateExecutorAction)
	engine.DELETE("/executors/:id/actions/:actionId", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.deleteExecutorAction)
	engine.GET("/accessable-executors", middleware.TokenAuth(), h.getAccessableExecutors)
	engine.POST("/import-agents", middleware.TokenAuth(), middleware.CheckAdmin(), h.importAgents)

	engine.POST("/check/executors", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.checkCreateExecutor)
	engine.PUT("/check/executors/:id", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.checkUpdateExecutor)
	engine.POST("/check/executors/:id/actions", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.checkCreateExecutorAction)
	engine.PUT("/check/executors/:id/actions/:actionId", middleware.TokenAuth(), middleware.CheckExecutorWhiteList(h.appstore), h.checkUpdateExecutorAction)
}

func (h *restHandler) getParamUint(c *gin.Context, key string) (uint64, error) {
	param := c.Param(key)

	if key == "" {
		return 0, errors.NewIError(errors.InvalidParameter, "", []interface{}{fmt.Sprintf("param %s cannot be empty", key)})
	}

	value, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return 0, errors.NewIError(errors.InvalidParameter, "", []interface{}{fmt.Sprintf("param %s is not a number", key)})
	}

	return value, nil
}

// 创建自定义节点
func (h *restHandler) checkCreateExecutor(c *gin.Context) {

	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err = common.JSONSchemaValid(data, CheckCreateExecutorSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	result, err := h.executor.CheckCreateExecutor(c.Request.Context(), dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// 创建自定义节点
func (h *restHandler) createExecutor(c *gin.Context) {

	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err = common.JSONSchemaValid(data, CreateExecutorSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	executorId, err := h.executor.CreateExecutor(c.Request.Context(), dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Writer.Header().Set("Location", fmt.Sprintf("/api/automation/v1/executors/%d", executorId))

	c.JSON(http.StatusCreated, gin.H{
		"id": fmt.Sprintf("%d", executorId),
	})
}

func (h *restHandler) checkUpdateExecutor(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	data, _ := io.ReadAll(c.Request.Body)

	err = common.JSONSchemaValid(data, UpdateExecutorSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	result, err := h.executor.CheckUpdateExecutor(c.Request.Context(), id, dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// 更新自定义节点
func (h *restHandler) updateExecutor(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	data, _ := io.ReadAll(c.Request.Body)

	err = common.JSONSchemaValid(data, UpdateExecutorSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.executor.UpdateExecutor(c.Request.Context(), id, dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// 获取自定义节点
func (h *restHandler) getExecutor(c *gin.Context) {

	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	executor, err := h.executor.GetExecutor(c.Request.Context(), id, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, executor)
}

// 获取自定义节点
func (h *restHandler) getExecutors(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	executors, err := h.executor.GetExecutors(c.Request.Context(), userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, executors)
}

// 删除自定义节点
func (h *restHandler) deleteExecutor(c *gin.Context) {

	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.executor.DeleteExecutor(c.Request.Context(), id, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) checkCreateExecutorAction(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	data, _ := io.ReadAll(c.Request.Body)
	err = common.JSONSchemaValid(data, CheckCreateExecutorActionSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorActionDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	result, err := h.executor.CheckCreateExecutorAction(c.Request.Context(), id, dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// 新建动作
func (h *restHandler) createExecutorAction(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	data, _ := io.ReadAll(c.Request.Body)
	err = common.JSONSchemaValid(data, CreateExecutorActionSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorActionDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	actionID, err := h.executor.CreateExecutorAction(c.Request.Context(), id, dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Writer.Header().Set("Location", fmt.Sprintf("/api/automation/v1/executors/%d/actions/%d", id, actionID))

	c.JSON(http.StatusCreated, gin.H{
		"id": fmt.Sprintf("%d", actionID),
	})
}

// 更新动作
func (h *restHandler) checkUpdateExecutorAction(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	actionID, err := h.getParamUint(c, "actionId")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	data, _ := io.ReadAll(c.Request.Body)
	err = common.JSONSchemaValid(data, CheckUpdateExecutorActionSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorActionDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	result, err := h.executor.CheckUpdateExecutorAction(c.Request.Context(), id, actionID, dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// 更新动作
func (h *restHandler) updateExecutorAction(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	actionID, err := h.getParamUint(c, "actionId")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	data, _ := io.ReadAll(c.Request.Body)
	err = common.JSONSchemaValid(data, UpdateExecutorActionSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ExecutorActionDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.executor.UpdateExecutorAction(c.Request.Context(), id, actionID, dto, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// 删除动作
func (h *restHandler) deleteExecutorAction(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	id, err := h.getParamUint(c, "id")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	actionID, err := h.getParamUint(c, "actionId")

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.executor.DeleteExecutorAction(c.Request.Context(), id, actionID, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) getAccessableExecutors(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	executors, err := h.executor.GetAccessableExecutors(c.Request.Context(), userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, executors)
}

func (h *restHandler) importAgents(c *gin.Context) {
	var err error
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err = common.JSONSchemaValid(data, ImportAgentsSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dto executor.ImportAgentsDto
	err = json.Unmarshal(data, &dto)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	result, err := h.executor.ImportAgents(c.Request.Context(), userInfo, &dto)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
