package agenthandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      终止对话
// @Description  终止对话
// @Tags         对话
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "app_key"
// @Param        request  body      object  true  "请求体"
// @Success      204  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/app/{app_key}/chat/termination [post]
func (h *agentHTTPHandler) TerminateChat(c *gin.Context) {
	var req agentreq.TerminateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[TerminateChat] should bind json error: %v", err)
		otellog.LogError(c, fmt.Sprintf("[TerminateChat] should bind json error: %v", err), err)
		rest.ReplyError(c, err)

		return
	}

	oteltrace.SetConversationID(c.Request.Context(), req.ConversationID)

	if req.ConversationID == "" {
		h.logger.Errorf("[TerminateChat] conversation_id is required")
		otellog.LogError(c, "[TerminateChat] conversation_id is required", nil)
		rest.ReplyError(c, capierr.New400Err(c, "conversation_id is required"))

		return
	}

	err := h.agentSvc.TerminateChat(c.Request.Context(), req.ConversationID, req.AgentRunID, req.InterruptedAssistantMessageID)
	if err != nil {
		h.logger.Errorf("[TerminateChat] terminate chat error: %v", err)
		otellog.LogError(c, fmt.Sprintf("[TerminateChat] terminate chat error: %v", err), err)
		rest.ReplyError(c, err)

		return
	}

	rest.ReplyOK(c, http.StatusNoContent, nil)
}
