package observabilityhandler

import (
	"fmt"
	"net/http"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
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
		err := capierr.New400Err(c, "[SessionList] agent_id is required")
		otellog.LogError(c.Request.Context(), "[SessionList] agent_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if conversationID == "" {
		h.logger.Errorf("[SessionList] conversation_id is required")
		err := capierr.New400Err(c, "[SessionList] conversation_id is required")
		otellog.LogError(c.Request.Context(), "[SessionList] conversation_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var sessionListReq observabilityreq.SessionListReq
	if err := c.ShouldBindJSON(&sessionListReq); err != nil {
		h.logger.Errorf("[SessionList] should bind json err: %v", err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[SessionList] should bind json err: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[SessionList] should bind json err: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	sessionListReq.AgentID = agentID
	sessionListReq.ConversationID = conversationID

	// 4. 参数验证
	if sessionListReq.StartTime == 0 || sessionListReq.EndTime == 0 {
		err := capierr.New400Err(c, "[SessionList] start_time and end_time are required")
		h.logger.Errorf("[SessionList] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[SessionList] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if sessionListReq.StartTime > sessionListReq.EndTime {
		err := capierr.New400Err(c, "[SessionList] start_time cannot be greater than end_time")
		h.logger.Errorf("[SessionList] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[SessionList] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if sessionListReq.Size <= 0 {
		sessionListReq.Size = 10
	}

	if sessionListReq.Page <= 0 {
		sessionListReq.Page = 1
	}

	// 5. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		err := capierr.New404Err(c, "[SessionList] user not found")
		otellog.LogError(c.Request.Context(), "[SessionList] user not found", err)
		h.logger.Errorf("[SessionList] user not found: %v", err)
		rest.ReplyError(c, err)

		return
	}

	// 6. 设置用户信息到请求对象中
	sessionListReq.XAccountID = user.ID
	sessionListReq.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 7. 调用服务
	// h.logger.Infof("[SessionList] query session list, agent_id: %s, conversation_id: %s, version: %s, time_range: [%d, %d], page: %d, size: %d",
	// 	sessionListReq.AgentID, sessionListReq.ConversationID, sessionListReq.AgentVersion, sessionListReq.StartTime, sessionListReq.EndTime, sessionListReq.Page, sessionListReq.Size)

	// 调用可观测性服务获取Session列表
	resp, httpErr := h.observabilitySvc.SessionList(c.Request.Context(), &sessionListReq)
	if httpErr != nil {
		h.logger.Errorf("[SessionList] call observability service error: %v", httpErr.Error())
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[SessionList] call observability service error: %v", httpErr.Error()), httpErr)
		oteltrace.EndSpan(c.Request.Context(), httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, resp)
}