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

// @Summary      获取指定session的详情信息
// @Description  获取指定session的详情信息，包含多个指标
// @Tags         可观测性
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        conversation_id  path      string  true  "conversation_id"
// @Param        session_id  path      string  true  "session_id"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功返回session详情"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/observability/agent/{agent_id}/conversation/{conversation_id}/session/{session_id}/detail [post]
func (h *observabilityHTTPHandler) SessionDetail(c *gin.Context) {
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	conversationID := c.Param("conversation_id")
	sessionID := c.Param("session_id")

	if agentID == "" {
		h.logger.Errorf("[SessionDetail] agent_id is required")
		err := capierr.New400Err(c, "[SessionDetail] agent_id is required")
		otellog.LogError(c.Request.Context(), "[SessionDetail] agent_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if conversationID == "" {
		h.logger.Errorf("[SessionDetail] conversation_id is required")
		err := capierr.New400Err(c, "[SessionDetail] conversation_id is required")
		otellog.LogError(c.Request.Context(), "[SessionDetail] conversation_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if sessionID == "" {
		h.logger.Errorf("[SessionDetail] session_id is required")
		err := capierr.New400Err(c, "[SessionDetail] session_id is required")
		otellog.LogError(c.Request.Context(), "[SessionDetail] session_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var sessionDetailReq observabilityreq.SessionDetailReq
	if err := c.ShouldBindJSON(&sessionDetailReq); err != nil {
		h.logger.Errorf("[SessionDetail] should bind json err: %v", err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[SessionDetail] should bind json err: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[SessionDetail] should bind json err: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	sessionDetailReq.AgentID = agentID
	sessionDetailReq.ConversationID = conversationID
	sessionDetailReq.SessionID = sessionID

	// 4. 参数验证
	if sessionDetailReq.StartTime == 0 || sessionDetailReq.EndTime == 0 {
		err := capierr.New400Err(c, "[SessionDetail] start_time and end_time are required")
		h.logger.Errorf("[SessionDetail] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[SessionDetail] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if sessionDetailReq.StartTime > sessionDetailReq.EndTime {
		err := capierr.New400Err(c, "[SessionDetail] start_time cannot be greater than end_time")
		h.logger.Errorf("[SessionDetail] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[SessionDetail] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 5. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		err := capierr.New404Err(c, "[SessionDetail] user not found")
		otellog.LogError(c.Request.Context(), "[SessionDetail] user not found", err)
		h.logger.Errorf("[SessionDetail] user not found: %v", err)
		rest.ReplyError(c, err)

		return
	}

	// 6. 设置用户信息到请求对象中
	sessionDetailReq.XAccountID = user.ID
	sessionDetailReq.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 7. 调用服务
	// h.logger.Infof("[SessionDetail] query session detail, agent_id: %s, conversation_id: %s, session_id: %s, version: %s, time_range: [%d, %d]",
	// 	sessionDetailReq.AgentID, sessionDetailReq.ConversationID, sessionDetailReq.SessionID, sessionDetailReq.AgentVersion, sessionDetailReq.StartTime, sessionDetailReq.EndTime)

	// 调用可观测性服务获取Session详情
	resp, httpErr := h.observabilitySvc.SessionDetail(c.Request.Context(), &sessionDetailReq)
	if httpErr != nil {
		h.logger.Errorf("[SessionDetail] call observability service error: %v", httpErr.Error())
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[SessionDetail] call observability service error: %v", httpErr.Error()), httpErr)
		oteltrace.EndSpan(c.Request.Context(), httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, resp)
}