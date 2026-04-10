package tplhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/ginhelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// GetPublishInfo 获取模板发布信息
// @Summary      获取模板发布信息
// @Description  获取模板发布信息
// @Tags         模板
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "id"
// @Success      200  {object}  object  "获取成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-tpl/{id}/publish-info [get]
func (h *daTplHTTPHandler) GetPublishInfo(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	tplID, err := ginhelper.GetParmIDInt64(c)
	if err != nil {
		_ = c.Error(err)

		return
	}

	res, err := h.daTplSvc.GetPublishInfo(ctx, tplID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, res)
}
