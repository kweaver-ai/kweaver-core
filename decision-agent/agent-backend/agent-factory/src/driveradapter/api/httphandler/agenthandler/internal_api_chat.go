package agenthandler

import (
	"fmt"
	"net/http"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"

	// "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr/chelper"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// NOTE: API调用，除url不同，其余与外部调用相同，只是token变为长期有效
func (h *agentHTTPHandler) InternalAPIChat(c *gin.Context) {
	reqStartTime := cutil.GetCurrentMSTimestamp()
	// 1. app_key
	agentAPPKey := c.Param("app_key")
	if agentAPPKey == "" {
		httpErr := capierr.New400Err(c, "[InternalAPIChat] app key is empty")
		otellog.LogError(c, "[InternalAPIChat] app key is empty", httpErr)
		h.logger.Errorf("[InternalAPIChat] app key is empty")
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 获取请求参数
	var req agentreq.ChatReq = agentreq.ChatReq{
		AgentAPPKey:     agentAPPKey,
		AgentVersion:    "latest",
		ExecutorVersion: "v2",
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, fmt.Sprintf("[InternalAPIChat] should bind json err: %v", err))
		otellog.LogError(c, fmt.Sprintf("[InternalAPIChat] should bind json err: %v", err), err)
		h.logger.Errorf("[InternalAPIChat] should bind json err: %v", err)
		rest.ReplyError(c, httpErr)

		return
	}

	// NOTE: 获取用户ID
	req.XAccountID = c.Request.Header.Get("x-account-id")
	req.XAccountType = cenum.AccountType(c.Request.Header.Get("x-account-type"))
	req.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)
	req.UserID = c.Request.Header.Get("x-user")
	// visitorType := c.Request.Header.Get("x-visitor-type")
	// if visitorType == "app" {
	// 	req.VisitorType = constant.Business
	// } else { //user
	// 	req.VisitorType = constant.RealName
	// }
	req.CallType = constant.APIChat
	req.ReqStartTime = reqStartTime
	// NOTE: APIChat接口请求时，agentID 实际值为agentKey
	if req.AgentKey != "" {
		req.AgentID = req.AgentKey
	} else {
		httpErr := capierr.New400Err(c, "[InternalAPIChat] agent_key is required")

		h.logger.Errorf("[InternalAPIChat] agent_key is required")
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
		otellog.LogError(ctx, fmt.Sprintf("[InternalAPIChat] chat error: %v", err.Error()), err)
		h.logger.Errorf("[InternalAPIChat] chat error: %v", err.Error())
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
			h.logger.Errorf("[InternalAPIChat] chat failed: res is nil")
			c.JSON(http.StatusInternalServerError, rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError, apierr.AgentAPP_InternalError).
				WithErrorDetails("[InternalAPIChat] chat failed: res is nil").BaseError)

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
