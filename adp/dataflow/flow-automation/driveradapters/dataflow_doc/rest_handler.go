// Package dataflow_doc provides REST handlers for Dataflow document trigger APIs
package dataflow_doc

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

// RESTHandler defines the interface for dataflow-doc REST APIs
type RESTHandler interface {
	// RegisterAPIv2 registers v2 version APIs
	RegisterAPIv2(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

var (
	triggerSchema  = "dataflow-doc/trigger.json"
	completeSchema = "dataflow-doc/complete.json"
)

// restHandler implements RESTHandler
type restHandler struct {
	config   *common.Config
	hydra    drivenadapters.HydraPublic
	userMgnt drivenadapters.UserManagement
	mgnt     mgnt.MgntHandler
}

// NewRESTHandler creates a new RESTHandler instance
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			config:   common.NewConfig(),
			hydra:    drivenadapters.NewHydraPublic(),
			userMgnt: drivenadapters.NewUserManagement(),
			mgnt:     mgnt.NewMgnt(),
		}
	})

	return rh
}

// RegisterAPIv2 registers v2 version APIs
func (h *restHandler) RegisterAPIv2(engine *gin.RouterGroup) {
	engine.POST("/dataflow-doc/trigger/:dagId", middleware.TokenAuth(), h.trigger)
	engine.POST("/dataflow-doc/complete", middleware.TokenAuth(), h.complete)

	// 文件访问接口
	engine.GET("/dataflow-doc/files", middleware.TokenAuth(), h.listFiles)
	engine.GET("/dataflow-doc/files/:file_id", middleware.TokenAuth(), h.getFile)
	engine.GET("/dataflow-doc/files/:file_id/download", middleware.TokenAuth(), h.downloadFile)
}

// TriggerMode defines the trigger mode type
type TriggerMode string

const (
	TriggerModeForm   TriggerMode = "form"   // 表单直接上传
	TriggerModeLocal  TriggerMode = "local"  // 先触发后上传
	TriggerModeRemote TriggerMode = "remote" // URL下载触发
)

// TriggerRequest represents the JSON trigger request body
type TriggerRequest struct {
	SourceFrom  string                 `json:"source_from"`  // local or remote
	Name        string                 `json:"name"`         // 文件名
	Size        int64                  `json:"size"`         // 文件大小
	ContentType string                 `json:"content_type"` // MIME类型
	URL         string                 `json:"url"`          // 源文件URL(仅remote模式)
	Data        map[string]interface{} `json:"data"`         // 触发器扩展字段
}

// TriggerResponse represents the trigger response
type TriggerResponse struct {
	DagID         string                        `json:"dag_id"`
	DagInstanceID string                        `json:"dag_instance_id,omitempty"`
	FileID        string                        `json:"file_id"`
	DocID         string                        `json:"docid"`
	Status        string                        `json:"status"` // ready, pending, processing
	Name          string                        `json:"name"`
	Size          int64                         `json:"size,omitempty"`
	UploadReq     *drivenadapters.UploadRequest `json:"upload_req,omitempty"`
}

// CompleteRequest represents the complete request body
type CompleteRequest struct {
	FileID string `json:"file_id"` // 支持纯ID或dfs://格式
	Etag   string `json:"etag"`
	Size   int64  `json:"size"`
}

// CompleteResponse represents the complete response
type CompleteResponse struct {
	FileID    string `json:"file_id"`
	DocID     string `json:"docid"`
	Status    string `json:"status"`
	Continued bool   `json:"continued"`
}

// trigger handles POST /api/automation/v2/dataflow-doc/trigger/:dagId
// Supports three sources:
// 1. multipart/form-data - 表单直接上传 (source_from=form)
// 2. JSON with source_from=local - 先触发后上传
// 3. JSON with source_from=remote - URL下载触发
func (h *restHandler) trigger(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	contentType := c.GetHeader("Content-Type")

	var params *mgnt.TriggerDataflowDocParams
	var err error

	// 根据Content-Type判断来源
	if contentType != "" && len(contentType) >= 9 && contentType[:9] == "multipart" {
		// multipart/form-data 来源 (source_from=form)
		params, err = h.parseMultipartTrigger(c, dagID, userInfo)
	} else {
		// JSON 来源 (source_from=local 或 source_from=remote)
		params, err = h.parseJSONTrigger(c, dagID, userInfo)
	}

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	// 调用应用服务层
	result, err := h.mgnt.TriggerDataflowDoc(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	// 构建响应
	response := &TriggerResponse{
		DagID:         result.DagID,
		DagInstanceID: result.DagInstanceID,
		FileID:        result.FileID,
		DocID:         result.DocID,
		Status:        result.Status,
		Name:          result.Name,
		Size:          result.Size,
	}

	if result.UploadReq != nil {
		response.UploadReq = result.UploadReq
	}

	c.JSON(http.StatusOK, response)
}

// parseMultipartTrigger parses multipart/form-data trigger request
func (h *restHandler) parseMultipartTrigger(c *gin.Context, dagID string, userInfo *drivenadapters.UserInfo) (*mgnt.TriggerDataflowDocParams, error) {
	// 获取上传的文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		return nil, errors.NewIError(errors.InvalidParameter, "", []interface{}{"file is required"})
	}
	defer file.Close()

	// 获取可选的name字段
	name := c.PostForm("name")
	if name == "" {
		name = fileHeader.Filename
	}

	// 获取可选的data字段
	var data map[string]interface{}
	dataStr := c.PostForm("data")
	if dataStr != "" {
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return nil, errors.NewIError(errors.InvalidParameter, "", []interface{}{"invalid data format"})
		}
	}

	// 立即将文件缓存到临时文件，避免 multipart.File 在 defer 后被关闭
	tempFile, _, err := utils.BufferToTempFile(file, "dataflow-upload")
	if err != nil {
		return nil, errors.NewIError(errors.InternalError, "", []interface{}{"failed to cache uploaded file"})
	}
	// 注意：tempFile 由调用者（mgnt层）在使用完毕后关闭

	return &mgnt.TriggerDataflowDocParams{
		DagID:       dagID,
		SourceFrom:  string(TriggerModeForm),
		Name:        name,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Data:        data,
		File:        tempFile,
	}, nil
}

// parseJSONTrigger parses JSON trigger request
func (h *restHandler) parseJSONTrigger(c *gin.Context, dagID string, userInfo *drivenadapters.UserInfo) (*mgnt.TriggerDataflowDocParams, error) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()})
	}

	// JSON schema 校验
	if err := common.JSONSchemaValid(data, triggerSchema); err != nil {
		return nil, err
	}

	var req TriggerRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()})
	}

	return &mgnt.TriggerDataflowDocParams{
		DagID:       dagID,
		SourceFrom:  req.SourceFrom,
		Name:        req.Name,
		Size:        req.Size,
		ContentType: req.ContentType,
		URL:         req.URL,
		Data:        req.Data,
	}, nil
}

// complete handles POST /api/automation/v2/dataflow-doc/complete
// 用于"先触发后上传"来源(source_from=local)，客户端上传完成后调用
func (h *restHandler) complete(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	// JSON schema 校验
	if err := common.JSONSchemaValid(data, completeSchema); err != nil {
		errors.ReplyError(c, err)
		return
	}

	var req CompleteRequest
	if err := json.Unmarshal(data, &req); err != nil {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	// 调用应用服务层
	result, err := h.mgnt.CompleteDataflowDocUpload(c.Request.Context(), &mgnt.CompleteDataflowDocUploadParams{
		FileID: req.FileID,
		Etag:   req.Etag,
		Size:   req.Size,
	}, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, &CompleteResponse{
		FileID:    result.FileID,
		DocID:     result.DocID,
		Status:    result.Status,
		Continued: result.Continued,
	})
}
