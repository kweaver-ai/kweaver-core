package dataflow_doc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
)

// listFiles handles GET /api/automation/v2/dataflow-doc/files
func (h *restHandler) listFiles(c *gin.Context) {
	dagInstanceID := c.Query("dag_instance_id")
	if dagInstanceID == "" {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"dag_instance_id": "required"}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	files, err := h.mgnt.ListFlowFiles(c.Request.Context(), dagInstanceID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// getFile handles GET /api/automation/v2/dataflow-doc/files/:file_id
func (h *restHandler) getFile(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"file_id": "required"}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	file, err := h.mgnt.GetFlowFile(c.Request.Context(), fileID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, file)
}

// downloadFile handles GET /api/automation/v2/dataflow-doc/files/:file_id/download
func (h *restHandler) downloadFile(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		errors.ReplyError(c, errors.NewIError(errors.InvalidParameter, "", map[string]interface{}{"file_id": "required"}))
		return
	}

	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	downloadInfo, err := h.mgnt.GetFlowFileDownloadURL(c.Request.Context(), fileID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, downloadInfo)
}
