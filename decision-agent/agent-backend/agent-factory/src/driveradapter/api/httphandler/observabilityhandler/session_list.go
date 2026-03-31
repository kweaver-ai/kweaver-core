package observabilityhandler

import (
	"fmt"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
)

// @Summary      获取指定对话的session列表
// @Description  获取指定对话的session列表，包含session的详细信息
// @Tags         可观测性
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        conversation_id  path      string  true  "conversation_id"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功返回session列表"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/observability/agent/{agent_id}/conversation/{conversation_id}/session [post]
func (h *observabilityHTTPHandler) SessionList(c *gin.Context) {
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	conversationID := c.Param("conversation_id")

	if agentID == "" {
		h.logger.Errorf("[SessionList] agent_id is required")
		o11y.Error(c, "[SessionList] agent_id is required")
		httpErr := capierr.New400Err(c, "[SessionList] agent_id is required")
		rest.ReplyError(c, httpErr)

		return
	}

	if conversationID == "" {
		h.logger.Errorf("[SessionList] conversation_id is required")
		o11y.Error(c, "[SessionList] conversation_id is required")
		httpErr := capierr.New400Err(c, "[SessionList] conversation_id is required")
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 获取请求参数
	var req observabilityreq.SessionListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[SessionList] should bind json err: %v", err)
		o11y.Error(c, fmt.Sprintf("[SessionList] should bind json err: %v", err))
		httpErr := capierr.New400Err(c, fmt.Sprintf("[SessionList] should bind json err: %v", err))
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	req.AgentID = agentID
	req.ConversationID = conversationID

	// 4. 参数验证
	if req.StartTime == 0 || req.EndTime == 0 {
		err := capierr.New400Err(c, "[SessionList] start_time and end_time are required")
		h.logger.Errorf("[SessionList] time range is invalid: %v", err)
		o11y.Error(c, "[SessionList] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	if req.StartTime > req.EndTime {
		err := capierr.New400Err(c, "[SessionList] start_time cannot be greater than end_time")
		h.logger.Errorf("[SessionList] time range is invalid: %v", err)
		o11y.Error(c, "[SessionList] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	if req.Size <= 0 {
		req.Size = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	// 5. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		httpErr := capierr.New404Err(c, "[SessionList] user not found")
		o11y.Error(c, "[SessionList] user not found")
		h.logger.Errorf("[SessionList] user not found: %v", httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	// 6. 设置用户信息到请求对象中
	req.XAccountID = user.ID
	req.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 7. 调用服务
	// h.logger.Infof("[SessionList] query session list, agent_id: %s, conversation_id: %s, version: %s, time_range: [%d, %d], page: %d, size: %d",
	// 	req.AgentID, req.ConversationID, req.AgentVersion, req.StartTime, req.EndTime, req.Page, req.Size)

	// 调用可观测性服务获取Session列表
	resp, httpErr := h.observabilitySvc.SessionList(c.Request.Context(), &req)
	if httpErr != nil {
		h.logger.Errorf("[SessionList] call observability service error: %v", httpErr.Error())
		o11y.Error(c, fmt.Sprintf("[SessionList] call observability service error: %v", httpErr.Error()))
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, resp)
}
