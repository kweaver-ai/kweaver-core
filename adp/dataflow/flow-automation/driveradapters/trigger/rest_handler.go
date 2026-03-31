// Package trigger trigger driveradapters
package trigger

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const kcinfoSchema = "base/kcinfo.json"
const pythonCodeCallback = "pythoncode"

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)

	// RegisterPrivateAPI 注册内部API
	RegisterPrivateAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

type restHandler struct {
	mgnt mgnt.MgntHandler
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			mgnt: mgnt.NewMgnt(),
		}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	// engine.GET("/trigger", h.trigger)
	// engine.POST("/trigger/kc-userinfo", h.kcInfoTrigger)
	engine.POST("/trigger/cron/:id", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.cronTriggerManual)
}

// 注册内部API
func (h *restHandler) RegisterPrivateAPI(engine *gin.RouterGroup) {
	engine.POST("/trigger/kc-userinfo", h.kcInfoTrigger)
	engine.POST("/trigger/continue/:id", h.continueTrigger)
	engine.POST("/trigger/cron/:id", h.cronTrigger)
}

func (h *restHandler) trigger(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(http.StatusOK, "ready")
}

func (h *restHandler) kcInfoTrigger(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, kcinfoSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	var param common.UserInfoMsg
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.mgnt.HandleKCUserInfoEvent(context.Background(), &param)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *restHandler) continueTrigger(c *gin.Context) {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tracelog := traceLog.WithContext(ctx)
	taskID := c.Param("id")
	data, _ := io.ReadAll(c.Request.Body)

	var param interface{}
	var paramMap map[string]interface{}
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	debugMode := os.Getenv("DEBUG")
	if debugMode == "true" {
		tracelog.Infof("callback %s, data: %v", taskID, param)
	}

	status := entity.TaskInstanceStatusSuccess
	if _paramMap, ok := param.(map[string]interface{}); ok {
		paramMap = _paramMap
		if _paramMap["code"] != nil && _paramMap["code"] != "" {
			status = entity.TaskInstanceStatusFailed
		}
	} else {
		paramMap = map[string]interface{}{"data": param}
	}

	if paramMap["type"] == pythonCodeCallback {
		if status == entity.TaskInstanceStatusFailed {
			status = entity.TaskInstanceStatusRetrying
		}
		if _paramMap, ok := paramMap["res"].(map[string]interface{}); ok {
			paramMap = _paramMap
		}
	}

	go func() {
		err = h.mgnt.ContinueBlockInstances(context.Background(), []string{taskID}, paramMap, status)

		if err != nil {
			tracelog.Errorln(err)
			errors.ReplyError(c, err)
			return
		}
	}()

	c.Status(http.StatusOK)
}

func (h *restHandler) cronTrigger(c *gin.Context) {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tracelog := traceLog.WithContext(ctx)
	taskID := c.Param("id")
	webhook := c.Request.Header.Get("webhook")
	tracelog.Infof("[cronTrigger] RunCronInstance start, webhook: %s, taskID: %s", webhook, taskID)
	data, _ := io.ReadAll(c.Request.Body)

	var param interface{}
	err = json.Unmarshal(data, &param)

	if err != nil {
		tracelog.Warnf("[cronTrigger] check param failed, err: %s", err.Error())
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	go func(taskID, webhook string) {
		err = h.mgnt.RunCronInstance(context.Background(), taskID, webhook)
		if err != nil {
			tracelog.Warnf("[cronTrigger] RunCronInstance failed, err: %s", err.Error())
		}
	}(taskID, webhook)

	c.Status(http.StatusOK)
}

func (h *restHandler) cronTriggerManual(c *gin.Context) {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	tracelog := traceLog.WithContext(ctx)
	dagID := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	data, _ := io.ReadAll(c.Request.Body)

	var param interface{}
	err = json.Unmarshal(data, &param)

	if err != nil {
		tracelog.Warnf("[cronTrigger] check param failed, err: %s", err.Error())
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	dag, err := h.mgnt.GetDagByID(context.Background(), dagID, "", c.GetString("bizDomainID"), userInfo)

	if err != nil {
		tracelog.Warnf("[cronTrigger] GetDagByID failed, err: %s", err.Error())
		c.JSON(http.StatusNotFound, errors.NewIError(errors.TaskNotFound, "", []interface{}{err.Error()}))
		return
	}

	if dag.ID == "" {
		tracelog.Warnf("[cronTrigger] DagID not match, dagID: %s, dag.ID: %s", dagID, dag.ID)
		c.JSON(http.StatusNotFound, errors.NewIError(errors.TaskNotFound, "", []interface{}{"DagID not found"}))
		return
	}

	go func(taskID, webhook string) {
		err = h.mgnt.RunCronInstance(context.Background(), taskID, webhook)
		if err != nil {
			tracelog.Warnf("[cronTrigger] RunCronInstance failed, err: %s", err.Error())
		}
	}(dagID, "")

	c.Status(http.StatusOK)
}
