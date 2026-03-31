package dataflow

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

var (
	createFlowParamsSchema = "dataflow/create.json"
	updateFlowParamsSchema = "dataflow/update.json"
)

type DataFlowHandler struct {
	config     *common.Config
	hydra      drivenadapters.HydraPublic
	hydraAdmin drivenadapters.HydraAdmin
	userMgnt   drivenadapters.UserManagement
	mgnt       mgnt.MgntHandler
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &DataFlowHandler{
			config:     common.NewConfig(),
			hydra:      drivenadapters.NewHydraPublic(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			userMgnt:   drivenadapters.NewUserManagement(),
			mgnt:       mgnt.NewMgnt(),
		}
	})

	return rh
}

// RegisterAPI 注册API
func (h *DataFlowHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/flow", middleware.TokenAuth(), middleware.CheckIsApp(), middleware.CheckBizDomainID(), h.create)
	engine.PUT("/flow/:id", middleware.TokenAuth(), h.update)
	engine.DELETE("/flow/:id", middleware.TokenAuth(), middleware.CheckBizDomainID(), h.delete)

	engine.POST("/:id/files/upload", middleware.TokenAuth(), h.uploadFile)
	engine.GET("/:id/files", middleware.TokenAuth(), h.listFiles)
	engine.DELETE("/:id/files", middleware.TokenAuth(), h.deleteFile)
	engine.GET("/:id/files/download", middleware.TokenAuth(), h.downloadFile)
}

func (h *DataFlowHandler) create(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, createFlowParamsSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.CreateDataFlowReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	var dagID string
	param.BizDomainID = c.GetString("bizDomainID")
	dagID, err = h.mgnt.CreateDataFlow(c.Request.Context(), &param, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Writer.Header().Set("Location", "/api/automation/v1/data-flow/"+dagID)
	c.JSON(http.StatusCreated, gin.H{
		"id": dagID,
	})
}

func (h *DataFlowHandler) update(c *gin.Context) {
	dagID := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, updateFlowParamsSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var param mgnt.UpdateDataFlowReq
	err = json.Unmarshal(data, &param)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.mgnt.UpdateDataFlow(c.Request.Context(), dagID, &param, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DataFlowHandler) delete(c *gin.Context) {
	dagID := c.Param("id")
	user, _ := c.Get("user")
	bizDomainID := c.GetString("bizDomainID")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.mgnt.DeleteDataFlow(c.Request.Context(), dagID, bizDomainID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DataFlowHandler) uploadFile(c *gin.Context) {
	dagID := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{"file is required"}))
		return
	}

	item, err := h.mgnt.UploadS3File(c.Request.Context(), dagID, fileHeader, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *DataFlowHandler) listFiles(c *gin.Context) {
	dagID := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	items, err := h.mgnt.ListS3Files(c.Request.Context(), dagID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": items,
	})
}

func (h *DataFlowHandler) deleteFile(c *gin.Context) {
	dagID := c.Param("id")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{"key is required"}))
		return
	}
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.mgnt.DeleteS3File(c.Request.Context(), dagID, key, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *DataFlowHandler) downloadFile(c *gin.Context) {
	dagID := c.Param("id")
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{"key is required"}))
		return
	}
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	url, err := h.mgnt.GetS3FileDownloadURL(c.Request.Context(), dagID, key, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	// 302 重定向到S3预签名地址
	c.Redirect(http.StatusFound, url)
}
