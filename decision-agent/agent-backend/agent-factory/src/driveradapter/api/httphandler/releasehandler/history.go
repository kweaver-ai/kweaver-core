package releasehandler

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

// @Summary      获取发布历史记录列表
// @Description  获取发布历史记录列表
// @Tags         已发布
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id}/release-history [get]
func (h *releaseHandler) HistoryList(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	agentID := c.Param("agent_id")
	if agentID == "" {
		h.logger.Errorf("agent id is empty")

		httpErr := capierr.New400Err(c, errors.New("agent id is empty"))

		rest.ReplyError(c, httpErr)

		return
	}

	historyList, _, err := h.releaseSvc.GetPublishHistoryList(ctx, agentID)
	if err != nil {
		h.logger.Errorf("GetPublishHistoryList error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		httpErr := capierr.New500Err(c, err.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	rt := map[string]interface{}{
		"entries": historyList,
		// "total":   total,
	}
	// 返回成功
	rest.ReplyOK(c, http.StatusOK, rt)
}

func (h *releaseHandler) HistoryInfo(c *gin.Context) {
	// 返回成功
	rest.ReplyOK(c, http.StatusOK, "ok")
}
