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

// @Summary      获取指定run的详情信息
// @Description  获取指定run的详情信息，包含run的完整信息
// @Tags         可观测性
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        conversation_id  path      string  true  "conversation_id"
// @Param        session_id  path      string  true  "session_id"
// @Param        run_id  path      string  true  "run_id"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功返回run详情"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/observability/agent/{agent_id}/conversation/{conversation_id}/session/{session_id}/run/{run_id}/detail [post]
func (h *observabilityHTTPHandler) RunDetail(c *gin.Context) {
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	conversationID := c.Param("conversation_id")
	sessionID := c.Param("session_id")
	runID := c.Param("run_id")

	if agentID == "" {
		h.logger.Errorf("[RunDetail] agent_id is required")
		err := capierr.New400Err(c, "[RunDetail] agent_id is required")
		otellog.LogError(c.Request.Context(), "[RunDetail] agent_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if conversationID == "" {
		h.logger.Errorf("[RunDetail] conversation_id is required")
		err := capierr.New400Err(c, "[RunDetail] conversation_id is required")
		otellog.LogError(c.Request.Context(), "[RunDetail] conversation_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if sessionID == "" {
		h.logger.Errorf("[RunDetail] session_id is required")
		err := capierr.New400Err(c, "[RunDetail] session_id is required")
		otellog.LogError(c.Request.Context(), "[RunDetail] session_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if runID == "" {
		h.logger.Errorf("[RunDetail] run_id is required")
		err := capierr.New400Err(c, "[RunDetail] run_id is required")
		otellog.LogError(c.Request.Context(), "[RunDetail] run_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var runDetailReq observabilityreq.RunDetailReq
	if err := c.ShouldBindJSON(&runDetailReq); err != nil {
		h.logger.Errorf("[RunDetail] should bind json err: %v", err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[RunDetail] should bind json err: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[RunDetail] should bind json err: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	runDetailReq.AgentID = agentID
	runDetailReq.ConversationID = conversationID
	runDetailReq.SessionID = sessionID
	runDetailReq.RunID = runID

	// 4. 参数验证
	if runDetailReq.StartTime == 0 || runDetailReq.EndTime == 0 {
		err := capierr.New400Err(c, "[RunDetail] start_time and end_time are required")
		h.logger.Errorf("[RunDetail] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[RunDetail] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if runDetailReq.StartTime > runDetailReq.EndTime {
		err := capierr.New400Err(c, "[RunDetail] start_time cannot be greater than end_time")
		h.logger.Errorf("[RunDetail] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[RunDetail] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 5. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		err := capierr.New404Err(c, "[RunDetail] user not found")
		otellog.LogError(c.Request.Context(), "[RunDetail] user not found", err)
		h.logger.Errorf("[RunDetail] user not found: %v", err)
		rest.ReplyError(c, err)

		return
	}

	// 6. 设置用户信息到请求对象中
	runDetailReq.XAccountID = user.ID
	runDetailReq.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 7. 调用服务
	// h.logger.Infof("[RunDetail] query run detail, agent_id: %s, conversation_id: %s, session_id: %s, run_id: %s, version: %s, time_range: [%d, %d]",
	// 	runDetailReq.AgentID, runDetailReq.ConversationID, runDetailReq.SessionID, runDetailReq.RunID, runDetailReq.AgentVersion, runDetailReq.StartTime, runDetailReq.EndTime)

	// 调用可观测性服务获取Run详情
	resp, httpErr := h.observabilitySvc.RunDetail(c.Request.Context(), &runDetailReq)
	if httpErr != nil {
		h.logger.Errorf("[RunDetail] call observability service error: %v", httpErr.Error())
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[RunDetail] call observability service error: %v", httpErr.Error()), httpErr)
		oteltrace.EndSpan(c.Request.Context(), httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, resp)
}