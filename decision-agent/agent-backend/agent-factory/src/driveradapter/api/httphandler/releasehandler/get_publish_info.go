package releasehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// GetPublishInfo 获取发布信息
// @Summary      获取已发布智能体的发布信息
// @Description  - 获取已发布智能体的发布信息 - 如果提供的agent_id不是已发布智能体，会返回404错误
// @Tags         发布相关,发布相关-internal
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Success      200  {object}  object  "获取成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id}/publish-info [get]
func (h *releaseHandler) GetPublishInfo(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	if agentID == "" {
		err := errors.New("agent_id is required")
		httpErr := capierr.New400Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	// 2. 调用服务层
	resp, err := h.releaseSvc.GetPublishInfo(ctx, agentID)
	if err != nil {
		h.logger.Errorf("GetPublishInfo failed, agentID: %s, error cause: %v, err trace: %+v\n", agentID, errors.Cause(err), err)
		_ = c.Error(err)

		return
	}

	// 3. 返回成功响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
