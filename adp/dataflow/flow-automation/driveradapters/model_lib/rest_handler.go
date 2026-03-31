// Package modellib 模型库接口
package modellib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/gin-gonic/gin"
)

type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once                 sync.Once
	rh                   RESTHandler
	ocrSchema            = "base/ocr.json"
	uieCreateSchema      = "base/uie_create.json"
	uieExtractSchema     = "base/uie_extract.json"
	audioTransferSchema  = "base/audiotransfer.json"
	createTagsRuleSchema = "base/tagsrule_create.json"
	updateTagsRuleSchema = "base/tagsrule_update.json"
	listTagsRuleSchema   = "base/tagsrule_list.json"
	extractTagsSchema    = "base/tagsrule_extract.json"
)

type restHandler struct {
	modelLib    mgnt.ModelLibHandler
	hydraAdmin  drivenadapters.HydraAdmin
	speechModel drivenadapters.SpeechModel
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			modelLib:    mgnt.NewModelLib(),
			hydraAdmin:  drivenadapters.NewHydraAdmin(),
			speechModel: drivenadapters.NewSpeechModel(),
		}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/ocr/task", middleware.TokenAuth(), h.createOCRTask)
	engine.POST("/convert/task", middleware.TokenAuth(), h.createAudioTransferTask)
	engine.GET("/convert/result/:taskId", middleware.TokenAuth(), h.getAudioTransferTaskResult)
	engine.POST("/uie/infer", middleware.TokenAuth(), h.entityExtract)
	engine.POST("/uie/train", middleware.TokenAuth(), h.startTrainModule)
	engine.POST("/uie/training-file", middleware.TokenAuth(), h.trainingFile)
	engine.POST("/tags/rule", middleware.TokenAuth(), h.createTagsRule)
	engine.POST("/tags/extract-by-rule", middleware.TokenAuth(), h.extractTagsByRule)
	engine.PUT("/models/:id", middleware.TokenAuth(), h.updateModelInfo)
	engine.GET("/models/:id", middleware.TokenAuth(), h.getModelInfo)
	engine.GET("/models", middleware.TokenAuth(), h.listTagsRule)
	engine.DELETE("/models/:id", middleware.TokenAuth(), h.deleteTagsRuleByID)
}

func (h *restHandler) createOCRTask(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	params := map[string]interface{}{}
	err := common.BindAndValid(c.Request.Body, ocrSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	ctx := context.WithValue(c.Request.Context(), common.Authorization, userInfo.TokenID) //nolint
	res, err := h.modelLib.RecognizeText(ctx, params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *restHandler) createAudioTransferTask(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	params := map[string]interface{}{}
	err := common.BindAndValid(c.Request.Body, audioTransferSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	res, err := h.modelLib.AudioTransfer(c.Request.Context(), params["docid"].(string), userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *restHandler) getAudioTransferTaskResult(c *gin.Context) {
	taskID := c.Param("taskId")
	code, res, err := h.speechModel.GetAudioTransferResult(c.Request.Context(), taskID)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	if code == http.StatusAccepted {
		c.Status(http.StatusAccepted)
	}
	c.JSON(http.StatusOK, res)
}

func (h *restHandler) entityExtract(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	params := map[string]interface{}{}
	err := common.BindAndValid(c.Request.Body, uieExtractSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	docID := params["docid"].(string)
	trainID := params["id"].(string)
	res, err := h.modelLib.EntityExtract(c.Request.Context(), trainID, docID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *restHandler) startTrainModule(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	params := map[string]interface{}{}
	err := common.BindAndValid(c.Request.Body, uieCreateSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	trainID := params["id"].(string)
	status, err := h.modelLib.TrainModule(c.Request.Context(), trainID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     trainID,
		"status": status,
	})
}

func (h *restHandler) createTagsRule(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	var params mgnt.TagRulesMol
	err := common.BindAndValid(c.Request.Body, createTagsRuleSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	res, err := h.modelLib.CreateTagsRule(c.Request.Context(), userInfo, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"id": fmt.Sprintf("%v", res)})
}

func (h *restHandler) getModelInfo(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	id := c.Param("id")

	res, err := h.modelLib.GetModelInfoByID(c.Request.Context(), id, userInfo.UserID)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *restHandler) updateModelInfo(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	id := c.Param("id")
	var params mgnt.TagRulesMol
	err := common.BindAndValid(c.Request.Body, updateTagsRuleSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	err = h.modelLib.UpdateModelInfo(c.Request.Context(), id, userInfo, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) deleteTagsRuleByID(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	id := c.Param("id")

	err := h.modelLib.DeleteModelInfoByID(c.Request.Context(), id, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *restHandler) listTagsRule(c *gin.Context) {
	var (
		page   = c.DefaultQuery("page", "0")
		limit  = c.DefaultQuery("limit", "50")
		status = c.DefaultQuery("status", "-1")
	)

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

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

	statusInt, err := strconv.ParseInt(status, 0, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"status": []string{"limit: Invalid type, expected integer"}}))
	}

	var query = map[string]interface{}{
		"page":   pageInt,
		"limit":  limitInt,
		"status": statusInt,
	}

	queryByte, _ := json.Marshal(query)
	err = common.JSONSchemaValid(queryByte, listTagsRuleSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	res, err := h.modelLib.ListModelInfo(c.Request.Context(), userInfo.UserID, statusInt, pageInt, limitInt)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *restHandler) extractTagsByRule(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)
	var params mgnt.TagExtractionParams
	err := common.BindAndValid(c.Request.Body, extractTagsSchema, &params)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	res, err := h.modelLib.ExtractTagsByRule(c.Request.Context(), &params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"tags": res})
}

func (h *restHandler) trainingFile(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	if fileHeader.Size > 1024*1024*1024 {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.FileSizeExceed, "", map[string]interface{}{
			"name":  fileHeader.Filename,
			"limit": 1024 * 1024 * 1024,
		}))
		return
	}

	trainID, lines, schema, err := h.modelLib.UploadTrainFile(c.Request.Context(), file, fileHeader)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":     trainID,
		"count":  lines,
		"schema": schema,
	})
}
