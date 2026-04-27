package publishedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/ginhelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// PubedTplDetail 已发布模板详情
// @Summary      已发布模板详情
// @Description  已发布模板详情
// @Tags         已发布
// @Accept       json
// @Produce      json
// @Param        tpl_id  path      string  true  "tpl_id"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/published/agent-tpl/{tpl_id} [get]
func (h *publishedHandler) PubedTplDetail(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	publishedTplID, err := ginhelper.GetParmInt64(c, "tpl_id")
	if err != nil {
		_ = c.Error(err)
		return
	}

	detail, err := h.publishedSvc.PubedTplDetail(ctx, publishedTplID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, detail)
}
