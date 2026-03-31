// Package mgnt automation driveradapters
package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/policy"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/state"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RuncodeReq 代码执行参数
type RuncodeReq struct {
	Code         string           `json:"code"`
	InputParams  []map[string]any `json:"input_params"`
	OutputParams []map[string]any `json:"output_params"`
}

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
	// 注册v2版本API
	RegisterAPIv2(engine *gin.RouterGroup)
	// 注册私有API
	RegisterPrivateAPI(engine *gin.RouterGroup)
}

var (
	once                     sync.Once
	rh                       RESTHandler
	createSchema             = "base/create.json"
	updateSchema             = "base/update.json"
	listDagSchema            = "base/listtask.json"
	listDocumentDagSchema    = "base/list-document-dag.json"
	cancletaskSchema         = "base/cancletask.json"
	dagRunListSchema         = "base/runtasklist.json"
	runcodeSchema            = "base/runcode.json"
	formInstanceSchema       = "base/forminstance.json"
	runWithDocSchema         = "base/run-with-doc.json"
	operateFlowSchema        = "base/operate-flow.json"
	batchGetDagSchema        = "base/batchlistdag.json"
	listTaskInstanceV2Schema = "base/list-task-instance-v2.json"
	singleDeBugSchema        = "base/single_debug.json"
	fullDeBugSchema          = "base/full_debug.json"
)

type restHandler struct {
	mgnt       mgnt.MgntHandler
	hydra      drivenadapters.HydraPublic
	hydraAdmin drivenadapters.HydraAdmin
	coderunner drivenadapters.CodeRunner
	userMgnt   drivenadapters.UserManagement
	config     *common.Config
	policy     policy.Handler
	efast      drivenadapters.Efast
	uniquery   drivenadapters.UniqueryDriven
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			mgnt:       mgnt.NewMgnt(),
			policy:     policy.NewPolicy(),
			hydra:      drivenadapters.NewHydraPublic(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			coderunner: drivenadapters.NewCodeRunner(),
			config:     common.NewConfig(),
			userMgnt:   drivenadapters.NewUserManagement(),
			efast:      drivenadapters.NewEfast(),
			uniquery:   drivenadapters.NewUniquery(),
		}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/dag", middleware.TokenAuth(), middleware.CheckIsApp(), middleware.CheckBizDomainID(), h.create)
	engine.PUT("/dag/:dagId", middleware.TokenAuth(), h.update)
	engine.GET("/dag/:dagId", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.getDagByID)
	engine.DELETE("/dag/:dagId", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.deleteDagByID)
	engine.GET("/dags", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.listDags(mgnt.ListDagsConfig{IsShared: false}))
	engine.GET("/shared-dags", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.listDags(mgnt.ListDagsConfig{IsShared: true}))
	engine.GET("/document-dags", middleware.TokenAuth(), h.listDocumentDags)
	engine.GET("/related-dags", middleware.TokenAuth(), h.listRelatedDags)
	engine.POST("/run-instance/:dagId", middleware.TokenAuth(), h.runInstance)
	engine.POST("/run-instance-form/:dagId", middleware.TokenAuth(), h.runInstanceWithForm)
	engine.POST("/public-api/:dagId", h.runPublicAPI)
	engine.POST("/run-instance-with-doc/:dagId", middleware.TokenAuth(), h.runInstanceWithDoc)
	engine.PUT("/run-instance/:instanceId", middleware.TokenAuth(), h.cancleRunningInstance)
	engine.GET("/dag/:dagId/results", middleware.TokenAuth(), h.dagInsRunList)
	engine.GET("/dag/:dagId/result/:resultId", middleware.TokenAuth(), h.listTaskInstance)
	engine.GET("/dag/suggestname/:name", middleware.TokenAuth(), h.getSuggestDagName)
	engine.GET("/actions", middleware.TokenAuth(), h.getActions)
	engine.POST("/pycode/run-by-params", middleware.TokenAuth(), h.runCode)
	engine.POST("/task/:instanceId", middleware.TokenAuth(), h.continueBlockedInstance)
	engine.PUT("/task/:taskId/results", middleware.TokenAuth(), h.updateTaskResults)
	engine.GET("/models/:id/dags", middleware.TokenAuth(), h.listModelBindDags)
	engine.GET("/task-instance/:taskId/trigger-config", middleware.TokenAuth(), h.getDagTriggerConfig)
	engine.POST("/agent/:agentKey", middleware.TokenAuth(), h.callAgent)
	engine.GET("/agents", middleware.TokenAuth(), h.getAgents)
	engine.GET("/dag/:dagId/count", middleware.TokenAuth(), h.getDagInstanceCount)
	engine.PUT("/dag-instance/:dagInsId/retry", middleware.TokenAuth(), h.retryDagInstance)
	engine.POST("/dags/single-debug", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.singleDebug)
	engine.POST("/dags/full-debug", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.fullDebug)
	engine.GET("/dags/single-debug/result", middleware.TokenAuth(), h.debugDagsResult)
}

func (h *restHandler) RegisterAPIv2(engine *gin.RouterGroup) {
	engine.GET("/dag/:dagId/results", middleware.TokenAuth(), h.listDagInstanceV2)
	engine.POST("/run-instance-form/:dagId", middleware.TokenAuth(), h.runInstanceWithFormV2)
	engine.POST("/dags/:fields", middleware.TokenAuth(), h.batchListDag)
	engine.GET("/dag/:dagId/result/:resultId", middleware.TokenAuth(), h.listTaskInstanceV2)
	engine.GET("/dags", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.listDagsWithPerm)
}

// 注册私有API
func (h *restHandler) RegisterPrivateAPI(engine *gin.RouterGroup) {
	engine.PUT("/dags/activate", h.activateDag)
	engine.PUT("/dags/deactivate", h.deactivateDag)
	engine.DELETE("/dags", h.deleteDag)
	engine.GET("/history-dags", h.historyData)
}

// create create a dag
func (h *restHandler) create(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, createSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.CreateDagReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dagID string
	param.BizDomainID = c.GetString("bizDomainID")
	dagID, err = h.mgnt.CreateDag(c.Request.Context(), &param, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Writer.Header().Set("Location", "/api/automation/v1/dag/"+dagID)
	c.JSON(http.StatusCreated, gin.H{
		"id": dagID,
	})
}

func (h *restHandler) update(c *gin.Context) { // nolint
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, updateSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.OptionalUpdateDagReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.mgnt.UpdateDag(c.Request.Context(), dagID, &param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *restHandler) getDagByID(c *gin.Context) {
	dagID := c.Param("dagId")
	if strings.EqualFold(c.Request.URL.Path, "/api/automation/v1/dag/suggestname") {
		h.getSuggestDagName(c)
		return
	}
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	versionID := c.Query("version")
	bizDomainID := c.GetString("bizDomainID")

	dagInfo, err := h.mgnt.GetDagByID(c.Request.Context(), dagID, versionID, bizDomainID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, dagInfo)
}

func (h *restHandler) deleteDagByID(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	bizDomainID := c.GetString("bizDomainID")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.mgnt.DeleteDagByID(context.Background(), dagID, bizDomainID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *restHandler) listDags(config mgnt.ListDagsConfig) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		userInfo := user.(*drivenadapters.UserInfo)
		config.Scope = c.Query("scope")

		var (
			page         = c.DefaultQuery("page", "0")
			limit        = c.DefaultQuery("limit", "20")
			sortBy       = c.DefaultQuery("sortby", "updated_at")
			order        = c.DefaultQuery("order", "desc")
			keyword      = c.Query("keyword")
			triggerType  = c.Query("trigger_type")
			triggerTypes = c.Query("trigger_types")
			flowType     = c.Query("type")
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

		var query = map[string]interface{}{
			"page":         pageInt,
			"limit":        limitInt,
			"sortby":       sortBy,
			"order":        order,
			"keyword":      keyword,
			"trigger_type": triggerType,
			"type":         flowType,
		}

		if triggerTypes != "" {
			query["trigger_types"] = strings.Split(triggerTypes, ",")
		}

		queryByte, _ := json.Marshal(query)
		err = common.JSONSchemaValid(queryByte, listDagSchema)
		if err != nil {
			errors.ReplyError(c, err)
			return
		}

		var params mgnt.QueryParams

		err = json.Unmarshal(queryByte, &params)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
			return
		}

		err = h.checkSwitchStatus()
		if err != nil {
			errors.ReplyError(c, err)
			return
		}

		params.BizDomainID = c.GetString("bizDomainID")
		dags, total, err := h.mgnt.ListDag(c.Request.Context(), params, userInfo, &config)
		if err != nil {
			errors.ReplyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"page":  pageInt,
			"limit": limitInt,
			"dags":  dags,
		})
	}
}

func (h *restHandler) runInstance(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	params := &mgnt.RunDagParams{
		ID:                  dagID,
		VersionID:           c.Query("version_id"),
		ParentDagInstanceID: c.GetHeader("X-Parent-Execution-ID"),
		CallBack:            c.GetHeader("X-Callback-URL"),
	}

	err = h.mgnt.RunInstance(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *restHandler) runInstanceWithForm(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, formInstanceSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.RunInstanceWithFormReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	params := &mgnt.RunDagParams{
		ID:                  dagID,
		VersionID:           c.Query("version_id"),
		ParentDagInstanceID: c.GetHeader("X-Parent-Execution-ID"),
		CallBack:            c.GetHeader("X-Callback-URL"),
		FormData:            param.Data,
	}

	dagInsID, err := h.mgnt.RunFormInstance(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Writer.Header().Set("Location", fmt.Sprintf("/api/automation/v2/dag/%s/result/%s", dagID, dagInsID))
	c.Status(http.StatusOK)
}

func (h *restHandler) runPublicAPI(c *gin.Context) {
	dagID := c.Param("dagId")

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, formInstanceSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.RunInstanceWithFormReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	params := &mgnt.RunDagParams{
		ID:        dagID,
		VersionID: c.Query("version_id"),
		FormData:  param.Data,
	}

	_, err = h.mgnt.RunFormInstance(c.Request.Context(), params, nil)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *restHandler) cancleRunningInstance(c *gin.Context) { //nolint
	instanceID := c.Param("instanceId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, cancletaskSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.DagInsStatusReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.mgnt.CancelRunningInstance(c.Request.Context(), instanceID, &param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *restHandler) dagInsRunList(c *gin.Context) {
	dagID := c.Param("dagId")
	if dagID == "" {
		detail := map[string]interface{}{"params": []string{0: "dagId should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	var (
		page                          = c.DefaultQuery("page", "0")
		limit                         = c.DefaultQuery("limit", "20")
		sortBy                        = c.DefaultQuery("sortby", "started_at")
		order                         = c.DefaultQuery("order", "desc")
		statusType, isTypeNotNull     = c.GetQuery("type")
		startTime, isStartTimeNotNull = c.GetQuery("start_time")
		endTime, isEndTimeNotNull     = c.GetQuery("end_time")
		name, isNameNotNull           = c.GetQuery("name")
		_type                         []string
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

	var query = map[string]interface{}{
		"page":   pageInt,
		"limit":  limitInt,
		"sortby": sortBy,
		"order":  order,
	}

	if isTypeNotNull {
		_type = strings.Split(statusType, ",")
		query["type"] = _type
	}

	if isStartTimeNotNull {
		startTimeInt, err := strconv.ParseInt(startTime, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"start_time: Invalid type, expected integer"}}))
			return
		}
		query["start_time"] = startTimeInt
	}

	if isEndTimeNotNull {
		endTimeInt, err := strconv.ParseInt(endTime, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"end_time: Invalid type, expected integer"}}))
			return
		}
		query["end_time"] = endTimeInt
	}

	name = strings.TrimSpace(name)
	if isNameNotNull && len(name) > 0 {
		name := regexp.QuoteMeta(name)
		query["name"] = name
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, dagRunListSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	dagInsRunList, total, err := h.mgnt.ListDagInstance(c.Request.Context(), dagID, query, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageInt,
		"limit":    limitInt,
		"results":  dagInsRunList.DagInstanceRunInfo,
		"progress": dagInsRunList.Progress,
	})
}

func (h *restHandler) listTaskInstance(c *gin.Context) {
	dagID := c.Param("dagId")
	if dagID == "" {
		detail := map[string]interface{}{"params": []string{0: "dagId should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}

	resultID := c.Param("resultId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	taskInsRunList, _, err := h.mgnt.ListTaskInstance(c.Request.Context(), dagID, resultID, 0, -1, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, taskInsRunList)
}

func (h *restHandler) listTaskInstanceV2(c *gin.Context) {
	dagID := c.Param("dagId")
	if dagID == "" {
		detail := map[string]interface{}{"params": []string{0: "dagId should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}

	resultID := c.Param("resultId")
	page := c.DefaultQuery("page", "0")
	limit := c.DefaultQuery("limit", "20")

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

	var query = map[string]interface{}{
		"page":  pageInt,
		"limit": limitInt,
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, listTaskInstanceV2Schema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	taskInsRunList, total, err := h.mgnt.ListTaskInstance(c.Request.Context(), dagID, resultID, pageInt, limitInt, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"limit":   limitInt,
		"page":    pageInt,
		"total":   total,
		"results": taskInsRunList,
	})
}

func (h *restHandler) getSuggestDagName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		detail := map[string]interface{}{"params": []string{0: "name should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	suggestNmae, err := h.mgnt.GetSuggestDagName(c.Request.Context(), name, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name": suggestNmae,
	})
}

func (h *restHandler) checkSwitchStatus() error {
	status, err := h.policy.CheckStatus()
	if err != nil {
		return err
	}

	if !status {
		return errors.NewIError(errors.Forbidden, errors.ServiceDisabled, map[string]interface{}{"enable": false})
	}

	return nil
}

func (h *restHandler) getActions(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	actions, err := h.mgnt.ListActions(c.Request.Context(), userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, actions)
}

func (h *restHandler) runCode(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, runcodeSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param RuncodeReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	ctx := context.WithValue(c.Request.Context(), common.Authorization, userInfo.TokenID)
	actions, err := h.coderunner.RunPyCode(ctx, param.Code, param.InputParams, param.OutputParams)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, actions)
}

func (h *restHandler) continueBlockedInstance(c *gin.Context) {
	instanceID := c.Param("instanceId")

	err := h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.mgnt.ContinueBlockInstances(c.Request.Context(), []string{instanceID}, map[string]interface{}{"result": "pass"}, entity.TaskInstanceStatusSuccess)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *restHandler) updateTaskResults(c *gin.Context) {
	taskID := c.Param("taskId")

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)
	var param = make(map[string]interface{}, 0)

	err := json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.mgnt.UpdateTaskResults(c.Request.Context(), taskID, param, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) runInstanceWithDoc(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, runWithDocSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params mgnt.RunWithDocParams
	err = json.Unmarshal(data, &params)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.mgnt.RunInstanceWithDoc(c.Request.Context(), dagID, params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *restHandler) listDocumentDags(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	var (
		page    = c.DefaultQuery("page", "0")
		limit   = c.DefaultQuery("limit", "20")
		sortBy  = c.DefaultQuery("sortby", "updated_at")
		order   = c.DefaultQuery("order", "desc")
		keyword = c.Query("keyword")
		docid   = c.Query("docid")
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

	var query = map[string]interface{}{
		"page":    pageInt,
		"limit":   limitInt,
		"sortby":  sortBy,
		"order":   order,
		"keyword": keyword,
		"docid":   docid,
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, listDocumentDagSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	attrs, err := h.efast.GetDocMsg(c.Request.Context(), docid)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	ancestorIDs := utils.GetParentDocIDs(docid)
	var parentID string
	if len(ancestorIDs) > 0 {
		parentID = ancestorIDs[len(ancestorIDs)-1]
	}

	accessors, err := h.userMgnt.GetUserAccessorIDs(userInfo.UserID)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var (
		operatorFilter bson.M
		accessorFilter bson.M
		docIDFilters   bson.A
	)

	accessorFilter = bson.M{
		"$or": bson.A{
			bson.M{
				"accessors.id": bson.M{
					"$in": accessors,
				},
			},
			bson.M{
				"userid": userInfo.UserID,
			},
		},
	}

	docIDFilters = bson.A{
		bson.M{
			"steps.0.parameters.docid": parentID,
		},
		bson.M{
			"steps.0.parameters.inherit": true,
			"steps.0.parameters.docid": bson.M{
				"$in": ancestorIDs,
			},
		},
		bson.M{
			"steps.0.parameters.docids": bson.M{
				"$elemMatch": bson.M{
					"$eq": parentID,
				},
			},
		},
		bson.M{
			"steps.0.parameters.inherit": true,
			"steps.0.parameters.docids": bson.M{
				"$elemMatch": bson.M{
					"$in": ancestorIDs,
				},
			},
		},
	}

	if attrs.Size == -1 {
		operatorFilter = bson.M{
			"steps.0.operator": common.OpAnyShareSelectedFolderTrigger,
		}

		docIDFilters = append(
			docIDFilters,
			bson.M{
				"steps.0.parameters.docid": docid,
			},
			bson.M{
				"steps.0.parameters.docids": bson.M{
					"$elemMatch": bson.M{
						"$eq": docid,
					},
				},
			})
	} else {
		operatorFilter = bson.M{
			"steps.0.operator": common.OpAnyShareSelectedFileTrigger,
		}
	}

	var filter = bson.M{
		"$and": bson.A{
			accessorFilter,
			operatorFilter,
			bson.M{
				"$or": docIDFilters,
			},
		},
	}

	if keyword != "" {
		filter = bson.M{
			"$and": bson.A{
				bson.M{
					"name": bson.M{
						"$regex": keyword,
					},
				},
				filter,
			},
		}
	}

	var opt = options.FindOptions{}

	if limitInt > 0 {
		opt.Limit = &limitInt
		offset := pageInt * limitInt
		opt.Skip = &offset
	}

	if sortBy != "" {

		var sortKey = common.UpdatedAt
		var sortOrder = -1

		switch sortBy {
		case common.Name:
			sortKey = common.Name
		case common.Created_At:
			sortKey = common.CreatedAt
		}

		if strings.ToLower(order) == common.ASC {
			sortOrder = 1
		}
		opt.Sort = bson.D{{
			Key:   sortKey,
			Value: sortOrder,
		}}
	}

	dags, total, err := h.mgnt.ListDagByFields(c.Request.Context(), filter, opt)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var dagsWithCreator []interface{}

	for _, dag := range dags {
		dagsWithCreator = append(dagsWithCreator, map[string]interface{}{
			"id":         dag.ID,
			"title":      dag.Title,
			"actions":    dag.Actions,
			"created_at": dag.CreatedAt,
			"updated_at": dag.UpdatedAt,
			"status":     dag.Status,
			"creator":    dag.Creator,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  pageInt,
		"limit": limitInt,
		"dags":  dagsWithCreator,
	})
}

func (h *restHandler) listRelatedDags(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	var (
		page    = c.DefaultQuery("page", "0")
		limit   = c.DefaultQuery("limit", "20")
		sortBy  = c.DefaultQuery("sortby", "updated_at")
		order   = c.DefaultQuery("order", "desc")
		keyword = c.Query("keyword")
		docid   = c.Query("docid")
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

	var query = map[string]interface{}{
		"page":    pageInt,
		"limit":   limitInt,
		"sortby":  sortBy,
		"order":   order,
		"keyword": keyword,
		"docid":   docid,
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, listDocumentDagSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	attrs, err := h.efast.GetDocMsg(c.Request.Context(), docid)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	ancestorIDs := utils.GetParentDocIDs(docid)
	var parentID string
	if len(ancestorIDs) > 0 {
		parentID = ancestorIDs[len(ancestorIDs)-1]
	} else {
		ancestorIDs = make([]string, 0)
	}

	accessors, err := h.userMgnt.GetUserAccessorIDs(userInfo.UserID)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	selectedItemTrigger := common.OpAnyShareSelectedFileTrigger
	listItemDataSource := []string{common.AnyshareDataListFiles}
	specifyItemDataSource := common.AnyshareDataSpecifyFiles

	if attrs.Size == -1 {
		selectedItemTrigger = common.OpAnyShareSelectedFolderTrigger
		listItemDataSource = append(listItemDataSource, common.AnyshareDataListFolders)
		specifyItemDataSource = common.AnyshareDataSpecifyFolders
	}

	listItemDataSourceFilter := bson.A{
		// 触发目录包含当前文件/文件夹
		bson.M{
			"steps.0.datasource.parameters.docids": bson.M{
				"$elemMatch": bson.M{
					"$eq": docid,
				},
			},
		},
		bson.M{
			"steps.0.datasource.parameters.docid": docid,
		},
	}

	speciafyItemDataSourceFilters := bson.A{
		bson.M{
			"steps.0.datasource.parameters.docids": bson.M{
				"$elemMatch": bson.M{
					"$eq": docid,
				},
			},
		},
		bson.M{
			"steps.0.datasource.parameters.docid": docid,
		},
	}

	selectedItemTriggerFilters := bson.A{
		bson.M{
			"steps.0.parameters.docid": docid,
		},
		bson.M{
			"steps.0.parameters.docids": bson.M{
				"$elemMatch": bson.M{
					"$eq": docid,
				},
			},
		},
	}

	if len(ancestorIDs) > 0 {
		listItemDataSourceFilter = append(listItemDataSourceFilter,
			// 继承且触发目录包含文件任意祖先
			bson.M{
				"steps.0.datasource.parameters.docids": bson.M{
					"$elemMatch": bson.M{
						"$in": ancestorIDs,
					},
				},
				"steps.0.datasource.parameters.depth": -1,
			},
			bson.M{
				"steps.0.datasource.parameters.docid": bson.M{
					"$in": ancestorIDs,
				},
				"steps.0.datasource.parameters.depth": -1,
			},
			bson.M{
				"steps.0.datasource.parameters.docids": bson.M{
					"$elemMatch": bson.M{
						"$eq": parentID,
					},
				},
			},
			bson.M{
				"steps.0.datasource.parameters.docid": parentID,
			},
		)

		selectedItemTriggerFilters = append(selectedItemTriggerFilters,
			// 当前选中项在触发范围
			bson.M{
				"$and": bson.A{
					bson.M{
						"steps.0.operator": selectedItemTrigger,
					},
					bson.M{
						"$or": bson.A{
							bson.M{
								"steps.0.parameters.docid": parentID,
							},
							bson.M{
								"steps.0.parameters.inherit": true,
								"steps.0.parameters.docid": bson.M{
									"$in": ancestorIDs,
								},
							},
							bson.M{
								"steps.0.parameters.docids": bson.M{
									"$elemMatch": bson.M{
										"$eq": parentID,
									},
								},
							},
							bson.M{
								"steps.0.parameters.inherit": true,
								"steps.0.parameters.docids": bson.M{
									"$elemMatch": bson.M{
										"$in": ancestorIDs,
									},
								},
							},
						},
					},
				},
			},
		)
	}

	filter := bson.M{
		"$or": bson.A{
			// 手动、定时触发源包含当前文件
			bson.M{
				"$and": bson.A{
					bson.M{
						// 当前用户创建
						"userid": userInfo.UserID,
						// 触发器为手动、定时
						"steps.0.operator": bson.M{
							"$in": []string{
								common.MannualTrigger,
								common.CronTrigger,
								common.CronWeekTrigger,
								common.CronMonthTrigger,
								common.CronCustomTrigger,
							},
						},
					},
					bson.M{
						"$or": bson.A{
							bson.M{
								"$and": bson.A{
									bson.M{
										// 数据源为指定文件夹下的文件/文件夹
										"steps.0.datasource.operator": bson.M{
											"$in": listItemDataSource,
										},
									},
									bson.M{
										"$or": listItemDataSourceFilter,
									},
								},
							},
							bson.M{
								"$and": bson.A{
									bson.M{
										"steps.0.datasource.operator": bson.M{
											"$in": []string{
												specifyItemDataSource,
											},
										},
									},
									bson.M{
										"$or": speciafyItemDataSourceFilters,
									},
								},
							},
						},
					},
				},
			},

			// 右键触发，创建者或具有执行权限
			bson.M{
				"$and": bson.A{
					bson.M{
						"$or": bson.A{
							bson.M{
								"status": "normal",
								"accessors.id": bson.M{
									"$in": accessors,
								},
							},
							bson.M{
								"userid": userInfo.UserID,
							},
						},
					},

					bson.M{
						"$or": selectedItemTriggerFilters,
					},
				},
			},

			// 事件触发
			bson.M{
				"$and": bson.A{
					bson.M{
						"userid": userInfo.UserID,
						"steps.0.operator": bson.M{
							"$in": bson.A{
								common.AnyshareFileUploadTrigger,
								common.AnyshareFileCopyTrigger,
								common.AnyshareFileMoveTrigger,
								common.AnyshareFileDeleteTrigger,
								common.AnyshareFolderCreateTrigger,
								common.AnyshareFolderCopyTrigger,
								common.AnyshareFolderMoveTrigger,
								common.AnyshareFolderRemoveTrigger,
							},
						},
					},
					bson.M{
						"$or": bson.A{
							bson.M{
								"steps.0.parameters.docids": bson.M{
									"$elemMatch": bson.M{
										"$eq": docid,
									},
								},
							},
							bson.M{
								"steps.0.parameters.docid": docid,
							},
						},
					},
				},
			},
		},
	}

	if keyword != "" {
		filter = bson.M{
			"$and": bson.A{
				bson.M{
					"name": bson.M{
						"$regex": keyword,
					},
				},
				filter,
			},
		}
	}

	var opt = options.FindOptions{}

	if limitInt > 0 {
		opt.Limit = &limitInt
		offset := pageInt * limitInt
		opt.Skip = &offset
	}

	if sortBy != "" {

		var sortKey = common.UpdatedAt
		var sortOrder = -1

		switch sortBy {
		case common.Name:
			sortKey = common.Name
		case common.Created_At:
			sortKey = common.CreatedAt
		}

		if strings.ToLower(order) == common.ASC {
			sortOrder = 1
		}
		opt.Sort = bson.D{{
			Key:   sortKey,
			Value: sortOrder,
		}}
	}

	dags, total, err := h.mgnt.ListDagByFields(c.Request.Context(), filter, opt)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	dagsWithCreator := make([]interface{}, 0)

	for _, dag := range dags {
		dagsWithCreator = append(dagsWithCreator, map[string]interface{}{
			"id":           dag.ID,
			"title":        dag.Title,
			"trigger_step": dag.TriggerStep,
			"actions":      dag.Actions,
			"created_at":   dag.CreatedAt,
			"updated_at":   dag.UpdatedAt,
			"status":       dag.Status,
			"creator":      dag.Creator,
			"is_owner":     dag.UserID == userInfo.UserID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  pageInt,
		"limit": limitInt,
		"dags":  dagsWithCreator,
	})
}

func (h *restHandler) listModelBindDags(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	id := c.Param("id")

	dags, err := h.mgnt.ListModelBindDags(c.Request.Context(), id, userInfo.UserID)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dags": dags,
	})
}

func (h *restHandler) getDagTriggerConfig(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	id := c.Param("taskId")

	typeBy := c.DefaultQuery("type", common.CreateFlowByClient)

	trigger, err := h.mgnt.GetDagTriggerConfig(c.Request.Context(), id, typeBy, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, trigger)
}

func (h *restHandler) callAgent(c *gin.Context) {
	data, _ := io.ReadAll(c.Request.Body)

	var inputs map[string]interface{}

	err := json.Unmarshal(data, &inputs)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	agentKey := c.Param("agentKey")
	stream := c.DefaultQuery("stream", "false")
	options := &drivenadapters.CallAgentOptions{
		Stream: stream == "true",
	}

	res, ch, err := h.mgnt.CallAgent(c.Request.Context(), agentKey, inputs, options, strings.TrimPrefix(userInfo.TokenID, "Bearer "))

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	if options.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		for msg := range ch {
			data, _ := json.Marshal(msg)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		}
	} else {
		c.JSON(http.StatusOK, res)
	}
}

func (h *restHandler) getAgents(c *gin.Context) {
	agents, err := h.mgnt.GetAgents(c.Request.Context())
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, agents)
}

func (h *restHandler) getDagInstanceCount(c *gin.Context) {
	id := c.Param("dagId")
	if id == "" {
		detail := map[string]interface{}{"params": []string{0: "dagId should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	_type := strings.TrimSpace(c.Query("type"))
	startTime, isStartTimeNotNull := c.GetQuery("start_time")
	endTime, isEndTimeNotNull := c.GetQuery("end_time")

	var query = map[string]interface{}{}
	if _type != "" {
		query["type"] = strings.Split(_type, ",")
	}

	if isStartTimeNotNull {
		startTimeInt, err := strconv.ParseInt(startTime, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"start_time: Invalid type, expected integer"}}))
			return
		}
		query["start_time"] = startTimeInt
	}

	if isEndTimeNotNull {
		endTimeInt, err := strconv.ParseInt(endTime, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"end_time: Invalid type, expected integer"}}))
			return
		}
		query["end_time"] = endTimeInt
	}

	count, err := h.mgnt.GetDagInstanceCount(c.Request.Context(), id, query, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

func (h *restHandler) listDagInstanceV2(c *gin.Context) {
	dagID := c.Param("dagId")
	if dagID == "" {
		detail := map[string]interface{}{"params": []string{0: "dagId should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	var (
		page                          = c.DefaultQuery("page", "0")
		limit                         = c.DefaultQuery("limit", "20")
		sortBy                        = c.DefaultQuery("sortby", "started_at")
		order                         = c.DefaultQuery("order", "desc")
		statusType, isTypeNotNull     = c.GetQuery("type")
		startTime, isStartTimeNotNull = c.GetQuery("start_time")
		endTime, isEndTimeNotNull     = c.GetQuery("end_time")
		name, isNameNotNull           = c.GetQuery("name")
		_type                         []string
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

	var query = map[string]interface{}{
		"page":   pageInt,
		"limit":  limitInt,
		"sortby": sortBy,
		"order":  order,
	}

	if isTypeNotNull {
		_type = strings.Split(statusType, ",")
		query["type"] = _type
	}

	if isStartTimeNotNull {
		startTimeInt, err := strconv.ParseInt(startTime, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"start_time: Invalid type, expected integer"}}))
			return
		}
		query["start_time"] = startTimeInt
	}

	if isEndTimeNotNull {
		endTimeInt, err := strconv.ParseInt(endTime, 0, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"params": []string{"end_time: Invalid type, expected integer"}}))
			return
		}
		query["end_time"] = endTimeInt
	}

	name = strings.TrimSpace(name)
	if isNameNotNull && len(name) > 0 {
		name := regexp.QuoteMeta(name)
		query["name"] = name
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, dagRunListSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	dagInsList, total, err := h.mgnt.ListDagInstanceV2(c.Request.Context(), dagID, query, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":    pageInt,
		"limit":   limitInt,
		"results": dagInsList,
		"total":   total,
	})
}

func (h *restHandler) retryDagInstance(c *gin.Context) {
	dagInsID := c.Param("dagInsId")
	if dagInsID == "" {
		detail := map[string]interface{}{"params": []string{0: "dagInsId should not empty"}}
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", detail))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.mgnt.RetryDagInstance(c.Request.Context(), dagInsID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *restHandler) runInstanceWithFormV2(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	successCallback := c.Query("success-callback")
	errorCallback := c.Query("error-callback")
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, formInstanceSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.RunInstanceWithFormReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	dagIns, vmIns, err := h.mgnt.RunFormInstanceV2(c.Request.Context(), dagID, param.Data, successCallback, errorCallback, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	if vmIns == nil {
		c.JSON(http.StatusAccepted, map[string]interface{}{"id": dagIns.ID})
		return
	}

	vmState, ret, vmErr := vmIns.Result()

	switch vmState {
	case state.Done:
		c.JSON(http.StatusOK, ret)
	case state.Error:
		c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", vmErr))
	case state.Wait:
		c.JSON(http.StatusAccepted, map[string]interface{}{"id": dagIns.ID})
	default:
		c.Status(http.StatusNoContent)
	}
}

// FlowOperationParams 流程参数
type FlowOperationParams struct {
	IDs    []string `json:"ids"`
	UserID string   `json:"userid"`
}

// handleFlowOperation handles common flow operations (activate, deactivate, delete)
func (h *restHandler) handleFlowOperation(c *gin.Context, operation func(context.Context, string, *mgnt.OptionalUpdateDagReq, *drivenadapters.UserInfo) error) {
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, operateFlowSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params FlowOperationParams
	err = json.Unmarshal(data, &params)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	userInfo, err := h.userMgnt.GetUserInfo(params.UserID)
	if err != nil {
		log.Warnf("[handleFlowOperation] GetUserInfo failed, details: %s, id: %s", err.Error(), params.UserID)
		errors.ReplyError(c, err)
		return
	}

	for _, id := range params.IDs {
		if err := operation(c.Request.Context(), id, nil, &userInfo); err != nil {
			log.Warnf("[handleFlowOperation] Operation failed, details: %s, id: %s", err.Error(), id)
			continue
		}
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) activateDag(c *gin.Context) {
	h.handleFlowOperation(c, func(ctx context.Context, id string, _ *mgnt.OptionalUpdateDagReq, userInfo *drivenadapters.UserInfo) error {
		status := string(entity.DagStatusNormal)
		return h.mgnt.UpdateDag(ctx, id, &mgnt.OptionalUpdateDagReq{
			Status: &status,
		}, userInfo)
	})
}

func (h *restHandler) deactivateDag(c *gin.Context) {
	h.handleFlowOperation(c, func(ctx context.Context, id string, _ *mgnt.OptionalUpdateDagReq, userInfo *drivenadapters.UserInfo) error {
		status := string(entity.DagStatusStopped)
		return h.mgnt.UpdateDag(ctx, id, &mgnt.OptionalUpdateDagReq{
			Status: &status,
		}, userInfo)
	})
}

func (h *restHandler) deleteDag(c *gin.Context) {
	h.handleFlowOperation(c, func(ctx context.Context, id string, _ *mgnt.OptionalUpdateDagReq, userInfo *drivenadapters.UserInfo) error {
		bizDomainID := c.GetString("bizDomainID")
		return h.mgnt.DeleteDagByID(ctx, id, bizDomainID, userInfo)
	})
}

func (h *restHandler) batchListDag(c *gin.Context) {
	fields := c.Param("fields")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)
	err := common.JSONSchemaValid(data, batchGetDagSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.DagInfoOptionReq
	err = json.Unmarshal(data, &param)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	dagInfo, err := h.mgnt.BatchGetDag(c.Request.Context(), param.DagIDs, fields, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, dagInfo)
}

func (h *restHandler) listDagsWithPerm(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	var (
		page         = c.DefaultQuery("page", "0")
		limit        = c.DefaultQuery("limit", "20")
		sortBy       = c.DefaultQuery("sortby", "updated_at")
		order        = c.DefaultQuery("order", "desc")
		keyword      = c.Query("keyword")
		triggerType  = c.Query("trigger_type")
		triggerTypes = c.Query("trigger_types")
		flowType     = c.Query("type")
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

	var query = map[string]interface{}{
		"page":         pageInt,
		"limit":        limitInt,
		"sortby":       sortBy,
		"order":        order,
		"keyword":      keyword,
		"trigger_type": triggerType,
		"type":         flowType,
	}

	if triggerTypes != "" {
		query["trigger_types"] = strings.Split(triggerTypes, ",")
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, listDagSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params mgnt.QueryParams

	err = json.Unmarshal(queryByte, &params)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.checkSwitchStatus()
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	params.BizDomainID = c.GetString("bizDomainID")
	dags, total, err := h.mgnt.ListDagV2(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  pageInt,
		"limit": limitInt,
		"dags":  dags,
	})
}

func (h *restHandler) singleDebug(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)
	err := common.JSONSchemaValid(data, singleDeBugSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.SingleDeBugReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	param.BizDomainID = c.GetString("bizDomainID")
	id, err := h.mgnt.SingleDeBug(c.Request.Context(), param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *restHandler) fullDebug(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)
	err := common.JSONSchemaValid(data, fullDeBugSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.FullDeBugReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	param.BizDomainID = c.GetString("bizDomainID")
	id, instID, err := h.mgnt.FullDebug(c.Request.Context(), param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"inst_id": instID,
	})
}

func (h *restHandler) debugDagsResult(c *gin.Context) {
	id := c.Query("id")

	status, contents, err := h.mgnt.SingleDeBugResult(c.Request.Context(), id)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	// 如果是异步执行，处于阻塞状态则返回202
	if status != entity.TaskInstanceStatusSuccess && status != entity.TaskInstanceStatusFailed {
		c.Status(http.StatusAccepted)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"result": contents,
	})
}

func (h *restHandler) historyData(c *gin.Context) {
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

	res, err := h.mgnt.ListHistoryData(c.Request.Context(), pageInt, limitInt)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
