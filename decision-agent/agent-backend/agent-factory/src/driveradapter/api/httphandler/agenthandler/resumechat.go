package agenthandler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	agentsvc "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/agentrunsvc"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// ResumeChat 恢复对话
// @Summary      对话恢复
// @Description  恢复之前的对话会话
// @Tags         对话
// @Accept       json
// @Produce      json
// @Param        app_key  path      string                   true  "应用 Key"
// @Param        request  body      swagger.ResumeReq      true  "恢复请求"
// @Success      200
// @Failure      400     {object}  swagger.APIError "请求参数错误"
// @Failure      500     {object}  swagger.APIError "服务器内部错误"
// @Router       /v1/app/{app_key}/chat/resume [post]
// @Security     BearerAuth
func (h *agentHTTPHandler) ResumeChat(c *gin.Context) {
	req := &agentreq.ResumeReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		otellog.LogError(c, fmt.Sprintf("[ResumeChat] should bind json error: %v", err), err)
		h.logger.Errorf("[ResumeChat] should bind json error: %v", err)
		rest.ReplyError(c, capierr.New400Err(c, err.Error()))

		return
	}

	oteltrace.SetConversationID(c.Request.Context(), req.ConversationID)

	channel, err := h.agentSvc.ResumeChat(c.Request.Context(), req.ConversationID)
	if err != nil {
		otellog.LogError(c, fmt.Sprintf("[ResumeChat] resume chat error: %v", err), err)
		h.logger.Errorf("[ResumeChat] resume chat error cause: %v,err trace: %+v\n", errors.Cause(err), err)
		httpErr := rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError, apierr.AgentAPP_Agent_ResumeFailed).WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	defer func() {
		// NOTE: 恢复会话结束后，关闭信号,或者报错中断之后，也要把信号关闭；
		session, ok := agentsvc.SessionMap.Load(req.ConversationID)
		if ok {
			session.(*agentsvc.Session).SetIsResuming(false)
		}
	}()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for data := range channel {
			_, err = c.Writer.Write(data)
			if err != nil {
				h.logger.Errorf("write stream data err: %v", err)
				break
			}

			c.Writer.Flush()

			if strings.HasPrefix(string(data), constant.DataEventEndStr) {
				break
			}
		}
	}()
	<-done
}
