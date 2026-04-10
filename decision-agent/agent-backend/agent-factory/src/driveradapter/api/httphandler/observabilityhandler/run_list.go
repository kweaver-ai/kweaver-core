package observabilityhandler

import (
	"fmt"
	"net/http"

	observabilityreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

// @Summary      获取指定session下的run列表
// @Description  获取指定对话的session下的run列表，返回每个run的详情列表
// @Tags         可观测性
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        conversation_id  path      string  true  "conversation_id"
// @Param        session_id  path      string  true  "session_id"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功返回run列表"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/observability/agent/{agent_id}/conversation/{conversation_id}/session/{session_id}/run [post]
func (h *observabilityHTTPHandler) RunList(c *gin.Context) {
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	conversationID := c.Param("conversation_id")
	sessionID := c.Param("session_id")

	if agentID == "" {
		h.logger.Errorf("[RunList] agent_id is required")

		err := capierr.New400Err(c, "[RunList] agent_id is required")
		otellog.LogError(c.Request.Context(), "[RunList] agent_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if conversationID == "" {
		h.logger.Errorf("[RunList] conversation_id is required")

		err := capierr.New400Err(c, "[RunList] conversation_id is required")
		otellog.LogError(c.Request.Context(), "[RunList] conversation_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if sessionID == "" {
		h.logger.Errorf("[RunList] session_id is required")

		err := capierr.New400Err(c, "[RunList] session_id is required")
		otellog.LogError(c.Request.Context(), "[RunList] session_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var runListReq observabilityreq.RunListReq
	if err := c.ShouldBindJSON(&runListReq); err != nil {
		h.logger.Errorf("[RunList] should bind json err: %v", err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[RunList] should bind json err: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[RunList] should bind json err: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	runListReq.AgentID = agentID
	runListReq.ConversationID = conversationID
	runListReq.SessionID = sessionID

	// 4. 参数验证
	if runListReq.StartTime == 0 || runListReq.EndTime == 0 {
		err := capierr.New400Err(c, "[RunList] start_time and end_time are required")
		h.logger.Errorf("[RunList] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[RunList] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if runListReq.StartTime > runListReq.EndTime {
		err := capierr.New400Err(c, "[RunList] start_time cannot be greater than end_time")
		h.logger.Errorf("[RunList] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[RunList] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if runListReq.Size <= 0 {
		runListReq.Size = 10
	}

	if runListReq.Page <= 0 {
		runListReq.Page = 1
	}

	// 5. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		err := capierr.New404Err(c, "[RunList] user not found")
		otellog.LogError(c.Request.Context(), "[RunList] user not found", err)
		h.logger.Errorf("[RunList] user not found: %v", err)
		rest.ReplyError(c, err)

		return
	}

	// 6. 设置用户信息到请求对象中
	runListReq.XAccountID = user.ID
	runListReq.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 7. 调用服务
	// h.logger.Infof("[RunList] query run list, agent_id: %s, conversation_id: %s, session_id: %s, version: %s, time_range: [%d, %d], page: %d, size: %d",
	// 	runListReq.AgentID, runListReq.ConversationID, runListReq.SessionID, runListReq.AgentVersion, runListReq.StartTime, runListReq.EndTime, runListReq.Page, runListReq.Size)

	// 调用可观测性服务获取Run列表
	resp, httpErr := h.observabilitySvc.RunList(c.Request.Context(), &runListReq)
	if httpErr != nil {
		h.logger.Errorf("[RunList] call observability service error: %v", httpErr.Error())
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[RunList] call observability service error: %v", httpErr.Error()), httpErr)
		oteltrace.EndSpan(c.Request.Context(), httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, resp)
}
