package observabilityhandler

import (
	"fmt"
	"net/http"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

// @Summary      查询指定Agent的可观测信息
// @Description  查询指定Agent的可观测信息，返回结果包含Agent的配置信息以及可观测性的指标，支持时间过滤
// @Tags         可观测性
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功返回Agent可观测信息"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/observability/agent/{agent_id}/detail [post]
func (h *observabilityHTTPHandler) AgentDetail(c *gin.Context) {
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	if agentID == "" {
		h.logger.Errorf("[AgentDetail] agent_id is required")
		o11y.Error(c, "[AgentDetail] agent_id is required")
		httpErr := capierr.New400Err(c, "[AgentDetail] agent_id is required")
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 获取请求参数
	var req observabilityreq.AgentDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[AgentDetail] should bind json err: %v", err)
		o11y.Error(c, fmt.Sprintf("[AgentDetail] should bind json err: %v", err))
		httpErr := capierr.New400Err(c, fmt.Sprintf("[AgentDetail] should bind json err: %v", err))
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 参数验证
	if req.StartTime == 0 || req.EndTime == 0 {
		err := capierr.New400Err(c, "[AgentDetail] start_time and end_time are required")
		h.logger.Errorf("[AgentDetail] time range is invalid: %v", err)
		o11y.Error(c, "[AgentDetail] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	if req.StartTime > req.EndTime {
		err := capierr.New400Err(c, "[AgentDetail] start_time cannot be greater than end_time")
		h.logger.Errorf("[AgentDetail] time range is invalid: %v", err)
		o11y.Error(c, "[AgentDetail] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	// 4. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		httpErr := capierr.New404Err(c, "[AgentDetail] user not found")
		o11y.Error(c, "[AgentDetail] user not found")
		h.logger.Errorf("[AgentDetail] user not found: %v", httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	// 5. 设置用户信息到请求对象中
	req.XAccountID = user.ID
	req.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 6. 设置AgentID到请求对象中
	req.AgentID = agentID

	// 7. 调用服务
	ctx := c.Request.Context()
	data, httpErr := h.observabilitySvc.AgentDetail(ctx, &req)

	if httpErr != nil {
		o11y.Error(ctx, fmt.Sprintf("[AgentDetail] agent detail query failed: %v", httpErr.Error()))
		h.logger.Errorf("[AgentDetail] agent detail query failed: %v", httpErr.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, data)
}
