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
		o11y.Error(c, "[RunList] agent_id is required")
		httpErr := capierr.New400Err(c, "[RunList] agent_id is required")
		rest.ReplyError(c, httpErr)

		return
	}

	if conversationID == "" {
		h.logger.Errorf("[RunList] conversation_id is required")
		o11y.Error(c, "[RunList] conversation_id is required")
		httpErr := capierr.New400Err(c, "[RunList] conversation_id is required")
		rest.ReplyError(c, httpErr)

		return
	}

	if sessionID == "" {
		h.logger.Errorf("[RunList] session_id is required")
		o11y.Error(c, "[RunList] session_id is required")
		httpErr := capierr.New400Err(c, "[RunList] session_id is required")
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 获取请求参数
	var req observabilityreq.RunListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[RunList] should bind json err: %v", err)
		o11y.Error(c, fmt.Sprintf("[RunList] should bind json err: %v", err))
		httpErr := capierr.New400Err(c, fmt.Sprintf("[RunList] should bind json err: %v", err))
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	req.AgentID = agentID
	req.ConversationID = conversationID
	req.SessionID = sessionID

	// 4. 参数验证
	if req.StartTime == 0 || req.EndTime == 0 {
		err := capierr.New400Err(c, "[RunList] start_time and end_time are required")
		h.logger.Errorf("[RunList] time range is invalid: %v", err)
		o11y.Error(c, "[RunList] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	if req.StartTime > req.EndTime {
		err := capierr.New400Err(c, "[RunList] start_time cannot be greater than end_time")
		h.logger.Errorf("[RunList] time range is invalid: %v", err)
		o11y.Error(c, "[RunList] time range is invalid")
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
		httpErr := capierr.New404Err(c, "[RunList] user not found")
		o11y.Error(c, "[RunList] user not found")
		h.logger.Errorf("[RunList] user not found: %v", httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	// 6. 设置用户信息到请求对象中
	req.XAccountID = user.ID
	req.XAccountType.LoadFromMDLVisitorType(user.Type)

	// 7. 调用服务
	// h.logger.Infof("[RunList] query run list, agent_id: %s, conversation_id: %s, session_id: %s, version: %s, time_range: [%d, %d], page: %d, size: %d",
	// 	req.AgentID, req.ConversationID, req.SessionID, req.AgentVersion, req.StartTime, req.EndTime, req.Page, req.Size)

	// 调用可观测性服务获取Run列表
	resp, httpErr := h.observabilitySvc.RunList(c.Request.Context(), &req)
	if httpErr != nil {
		h.logger.Errorf("[RunList] call observability service error: %v", httpErr.Error())
		o11y.Error(c, fmt.Sprintf("[RunList] call observability service error: %v", httpErr.Error()))
		rest.ReplyError(c, httpErr)

		return
	}

	c.JSON(http.StatusOK, resp)
}
