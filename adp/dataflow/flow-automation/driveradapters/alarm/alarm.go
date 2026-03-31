package alarm

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/alarm"
	"github.com/gin-gonic/gin"
)

type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once        sync.Once
	rh          RESTHandler
	alarmSchema = "base/alarmrule.json"
)

type restHandler struct {
	hydraAdmin     drivenadapters.HydraAdmin
	userManagement drivenadapters.UserManagement
	alarm          alarm.Alarm
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			hydraAdmin:     drivenadapters.NewHydraAdmin(),
			userManagement: drivenadapters.NewUserManagement(),
			alarm:          alarm.NewAlarm(),
		}
	})

	return rh
}

// RegisterAPI implements RESTHandler.
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/alarm-rule", middleware.TokenAuth(), middleware.CheckAdmin(), h.createAlarmRule)
	engine.PUT("/alarm-rules/:id", middleware.TokenAuth(), middleware.CheckAdmin(), h.updateAlarmRule)
	engine.GET("/alarm-rules", middleware.TokenAuth(), middleware.CheckAdmin(), h.listAlarmRule)
	engine.GET("/alarm-rules/:id", middleware.TokenAuth(), middleware.CheckAdmin(), h.getAlarmRule)
}

func (h *restHandler) createAlarmRule(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)
	err := common.JSONSchemaValid(data, alarmSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param alarm.AlarmTask
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	ruleID, err := h.alarm.ModifyAlarmRule(c.Request.Context(), &param)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rule_id": ruleID})
}

func (h *restHandler) updateAlarmRule(c *gin.Context) {
	id := c.Param("id")
	data, _ := io.ReadAll(c.Request.Body)
	err := common.JSONSchemaValid(data, alarmSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param alarm.AlarmTask
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	param.RuleID = strings.TrimSpace(id)
	_, err = h.alarm.ModifyAlarmRule(c.Request.Context(), &param)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) listAlarmRule(c *gin.Context) {
	var (
		page  = c.DefaultQuery("page", "0")
		limit = c.DefaultQuery("limit", "20")
	)
	pageInt, err := strconv.ParseInt(page, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"page: Invalid type, expected integer"}}))
		return
	}

	limitInt, err := strconv.ParseInt(limit, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"limit: Invalid type, expected integer"}}))
		return
	}

	alarmTasks, err := h.alarm.ListAlarmRule(c.Request.Context(), pageInt*limitInt, limitInt)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": alarmTasks})
}

func (h *restHandler) getAlarmRule(c *gin.Context) {
	id := c.Param("id")
	alarmTask, err := h.alarm.GetAlarmRule(c.Request.Context(), id)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, alarmTask)
}
