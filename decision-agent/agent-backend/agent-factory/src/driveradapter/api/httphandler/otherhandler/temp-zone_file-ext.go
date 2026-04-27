package otherhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
)

// @Summary      获取文件扩展名映射
// @Description  获取文件扩展名映射
// @Tags         其他
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "获取成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/temp-zone/file-ext-map [get]
func (o *otherHTTPHandler) TempZoneFileExt(ctx *gin.Context) {
	fileExtMap := cdaenum.GetFileExtMap()

	ctx.JSON(http.StatusOK, fileExtMap)
}
