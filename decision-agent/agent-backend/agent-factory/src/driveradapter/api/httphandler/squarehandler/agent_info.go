package squarehandler

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      智能体详情（已发布或未发布）
// @Description  - 智能体详情 - 包含已发布和未发布
// @Tags         已发布,agent-internal
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        version  path      string  true  "version"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-market/agent/{agent_id}/version/{version} [get]
func (h *squareHandler) AgentInfo(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	// 1. get req
	iReq, exists := c.Get(agentInfoReqCtxKey)
	if !exists {
		err := capierr.New400Err(c, "[AgentInfo]: agentInfoReqCtxKey不存在")
		_ = c.Error(err)
		c.Abort()

		return
	}

	req, ok := iReq.(*squarereq.AgentInfoReq)
	if !ok {
		err := capierr.New400Err(c, "[AgentInfo]: agentInfoReqCtxKey类型错误")
		_ = c.Error(err)
		c.Abort()

		return
	}

	// 2. 获取 agent 信息
	agentInfo, err := h.squareSvc.GetAgentInfo(ctx, req)
	if err != nil {
		h.logger.Errorf("GetPublishAgentList error cause: %v, err trace: %+v\n", errors.Cause(err), err)

		_ = c.Error(err)

		return
	}

	// 3. 返回成功
	rest.ReplyOK(c, http.StatusOK, agentInfo)
}
