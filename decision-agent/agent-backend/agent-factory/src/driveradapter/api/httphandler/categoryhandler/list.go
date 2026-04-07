package categoryhandler

import (
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

// @Summary      获取智能体分类
// @Description  获取智能体分类
// @Tags         发布相关
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "获取成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/category [get]
func (h *categoryHandler) List(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	rt, err := h.categorySvc.List(ctx)
	if err != nil {
		h.logger.Errorf("list category failed, err: %v", err)
		httpErr := capierr.New500Err(ctx, err.Error())

		// 返回错误
		rest.ReplyError(c, httpErr)

		return
	}
	// 返回成功
	rest.ReplyOK(c, http.StatusOK, rt)
}
