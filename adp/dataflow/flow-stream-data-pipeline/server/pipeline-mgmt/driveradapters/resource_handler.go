package driveradapters

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	serrors "flow-stream-data-pipeline/errors"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

// 分页获取指标模型资源列表
func (r *restHandler) ListResources(c *gin.Context) {
	logger.Debug("ListResources Start")

	// 获取分页参数
	resourceType := c.Query("resource_type")
	switch resourceType {
	case interfaces.RESOURCE_TYPE_PIPELINE:
		r.ListPipelineSources(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			serrors.StreamDataPipeline_UnSupported_ResourceType)

		rest.ReplyError(c, httpErr)
	}

}
