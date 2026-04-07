package agenthandler

import (
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func (h *agentHTTPHandler) InternalChat(c *gin.Context) {
	reqStartTime := cutil.GetCurrentMSTimestamp()
	// 1. app_key
	agentAPPKey := c.Param("app_key")
	if agentAPPKey == "" {
		h.logger.Errorf("[InternalChat] app key is empty")

		err := capierr.New400Err(c, "[InternalChat] app key is empty")
		otellog.LogError(c, "[InternalChat] app key is empty", err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var req agentreq.ChatReq = agentreq.ChatReq{
		ExecutorVersion: "v2",
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		otellog.LogError(c, fmt.Sprintf("[InternalChat] should bind json err: %v", err), err)
		rest.ReplyError(c, err)

		return
	}

	req.AgentAPPKey = agentAPPKey
	if req.ExecutorVersion == "" {
		req.ExecutorVersion = "v2"
	}
	// NOTE: 内部接口调用，从header中获取userID
	ctx := c.Request.Context()
	req.XAccountID = c.Request.Header.Get("x-account-id")
	req.XAccountType = cenum.AccountType(c.Request.Header.Get("x-account-type"))
	req.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)
	req.UserID = c.Request.Header.Get("x-user")

	// // 检查应用账号，应用账号应使用API Chat接口
	// if req.XAccountType == "app" {
	// 	errMsg := "应用账号应该使用API Chat接口"
	// 	h.logger.Errorf("[InternalChat] %s", errMsg)
	// 	o11y.Error(c, fmt.Sprintf("[InternalChat] %s", errMsg))
	// 	httpErr := capierr.New400Err(c, fmt.Sprintf("[InternalChat] %s", errMsg))
	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	// 3. 调用服务
	req.CallType = constant.InternalChat
	req.ReqStartTime = reqStartTime
	oteltrace.SetConversationID(ctx, req.ConversationID)

	channel, err := h.agentSvc.Chat(ctx, &req)
	oteltrace.SetConversationID(ctx, req.ConversationID)

	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[InternalChat] chat error: %v", err.Error()), err)
		h.logger.Errorf("[InternalChat] chat error: %v", err.Error())
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
				_, _ = c.Writer.Write(data)
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
			// fmt.Println(res)
		}

		if res == nil {
			h.logger.Errorf("[InternalChat] chat failed: res is nil")
			c.JSON(http.StatusInternalServerError, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).
				WithErrorDetails("[InternalChat] chat failed: res is nil").BaseError)

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
