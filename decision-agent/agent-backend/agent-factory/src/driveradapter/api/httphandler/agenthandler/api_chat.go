package agenthandler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	// "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper" // reserved for local dev debug
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"

	// "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr/chelper"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// NOTE: API调用，除url不同，其余与外部调用相同，只是token变为长期有效
// @Summary      APIChat
// @Description  APIChat
// @Tags         对话,对话-internal
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/api/chat/completion [post]
func (h *agentHTTPHandler) APIChat(c *gin.Context) {
	reqStartTime := cutil.GetCurrentMSTimestamp()
	agentAPPKey := c.Param("app_key")

	// 2. 获取请求参数
	var req agentreq.ChatReq = agentreq.ChatReq{
		AgentVersion:    "latest",
		ExecutorVersion: "v2",
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, fmt.Sprintf("[APIChat] should bind json err: %v", err))
		otellog.LogError(c, fmt.Sprintf("[APIChat] should bind json err: %v", err), err)
		h.logger.Errorf("[APIChat] should bind json err: %v", err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 本地调试：可取消注释以关闭 IncStream
	// if cenvhelper.IsLocalDev(cenvhelper.RunScenario_Aaron_Local_Dev) {
	// 	req.IncStream = false
	// }

	// NOTE: 获取用户ID
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		httpErr := capierr.New401Err(c, "[APIChat] user not found")
		otellog.LogError(c, "[APIChat] user not found", nil)
		h.logger.Errorf("[APIChat] user not found")
		rest.ReplyError(c, httpErr)

		return
	}

	req.ReqStartTime = reqStartTime
	req.UserID = user.ID
	req.XAccountID = user.ID
	req.XAccountType.LoadFromMDLVisitorType(user.Type)
	req.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)

	req.Token = strings.TrimPrefix(user.TokenID, "Bearer ")
	// if user.Type == rest.VisitorType_App {
	// 	req.VisitorType = constant.Business
	// } else if user.Type == rest.VisitorType_RealName {
	// 	req.VisitorType = constant.RealName
	// } else {
	// 	req.VisitorType = constant.Anonymous
	// }
	req.CallType = constant.APIChat
	if req.AgentKey != "" {
		req.AgentID = req.AgentKey
		if agentAPPKey == "" {
			agentAPPKey = req.AgentKey
		}
		req.AgentAPPKey = agentAPPKey
	} else {
		httpErr := capierr.New400Err(c, "[APIChat] agent_key is required")

		h.logger.Errorf("[APIChat] agent_key is required")
		rest.ReplyError(c, httpErr)

		return
	}

	if req.AgentVersion == "" {
		req.AgentVersion = "latest"
	}

	if req.ExecutorVersion == "" {
		req.ExecutorVersion = "v2"
	}

	ctx := c.Request.Context()
	oteltrace.SetConversationID(ctx, req.ConversationID)

	// 3. 调用服务
	channel, err := h.agentSvc.Chat(ctx, &req)
	oteltrace.SetConversationID(ctx, req.ConversationID)

	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[APIChat] chat error: %v", err.Error()), err)
		h.logger.Errorf("[APIChat] chat error: %v", err.Error())
		rest.ReplyError(c, err)

		return
	}

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		done := make(chan struct{})

		go func() {
			defer close(done)

			for data := range channel {
				_, err = c.Writer.Write(data)
				c.Writer.Flush()
			}
		}()
		<-done
	} else {
		// res := agentresp.ChatResp{}
		var res any
		for data := range channel {
			if err := sonic.Unmarshal(data, &res); err != nil {
				rest.ReplyError(c, err)
				return
			}
		}

		if res == nil {
			h.logger.Errorf("[APIChat] chat failed: res is nil")
			c.JSON(http.StatusInternalServerError, rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError, apierr.AgentAPP_InternalError).
				WithErrorDetails("[APIChat] chat failed: res is nil").BaseError)

			return
		}

		resultMap := res.(map[string]any)
		if _, ok := resultMap["BaseError"]; ok {
			// *rest.HTTPError
			c.JSON(http.StatusInternalServerError, resultMap["BaseError"])
		} else {
			// 如果res是agentresp.ChatResp{}，则返回200
			c.JSON(http.StatusOK, resultMap)
		}
	}
}
