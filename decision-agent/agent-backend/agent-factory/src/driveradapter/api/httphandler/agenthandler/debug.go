package agenthandler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req/chatopt"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      调试
// @Description  调试
// @Tags         对话
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "app_key"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/app/{app_key}/debug/completion [post]
func (h *agentHTTPHandler) Debug(c *gin.Context) {
	reqStartTime := cutil.GetCurrentMSTimestamp()
	// 1. app_key
	agentAPPKey := c.Param("app_key")
	if agentAPPKey == "" {
		err := capierr.New400Err(c, "[Debug] app key is empty")
		otellog.LogError(c, "[Debug] app key is empty", err)
		h.logger.Errorf("[Debug] app key is empty")
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	req := agentreq.DebugReq{
		Stream:          true,
		IncStream:       true,
		ExecutorVersion: "v2",
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, fmt.Sprintf("[Debug] should bind json err: %v", err))
		otellog.LogError(c, fmt.Sprintf("[Debug] should bind json err: %v", err), err)
		h.logger.Errorf("[Debug] should bind json err: %v", err)
		rest.ReplyError(c, httpErr)

		return
	}

	if req.ExecutorVersion == "" {
		req.ExecutorVersion = "v2"
	}

	ctx := c.Request.Context()
	req.AgentAPPKey = agentAPPKey

	// NOTE: 获取用户ID
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		httpErr := capierr.New401Err(c, "[Debug] user not found")
		otellog.LogError(c, "[Debug] user not found", nil)
		h.logger.Errorf("[Debug] user not found")
		rest.ReplyError(c, httpErr)

		return
	}

	req.UserID = user.ID
	req.Token = strings.TrimPrefix(user.TokenID, "Bearer ")

	// 如果AgentRunID为空，则生成新的ID
	if req.AgentRunID == "" {
		req.AgentRunID = cutil.UlidMake()
	}

	// NOTE: 目前Debug 和chat 内部实现逻辑一致，先复用
	chatReq := &agentreq.ChatReq{
		AgentAPPKey:               req.AgentAPPKey,
		AgentID:                   req.AgentID,
		AgentVersion:              req.AgentVersion,
		ConversationID:            req.ConversationID,
		Query:                     req.Input.Query,
		History:                   req.Input.History,
		CustomQuerys:              req.Input.CustomQuerys,
		SelectedFiles:             req.SelectedFiles,
		AgentRunID:                req.AgentRunID,
		ResumeInterruptInfo:       req.ResumeInterruptInfo,
		InterruptedAssistantMsgID: req.InterruptedAssistantMsgID,
		ChatMode:                  req.ChatMode,
		Stream:                    req.Stream,
		IncStream:                 req.IncStream,
		InternalParam: agentreq.InternalParam{
			UserID:       req.UserID,
			Token:        req.Token,
			ReqStartTime: reqStartTime,
		},
		ExecutorVersion: req.ExecutorVersion,
		ChatOption: chatopt.ChatOption{
			EnableDependencyCache:        req.ChatOption.EnableDependencyCache,
			IsNeedDocRetrivalPostProcess: req.ChatOption.IsNeedDocRetrivalPostProcess,
			IsNeedHistory:                req.ChatOption.IsNeedHistory,
			IsNeedProgress:               req.ChatOption.IsNeedProgress,
		},
	}
	chatReq.XAccountID = user.ID
	chatReq.XAccountType.LoadFromMDLVisitorType(user.Type)
	chatReq.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)
	// // 3. 调用服务
	// if user.Type == rest.VisitorType_App {
	// 	chatReq.VisitorType = constant.Business
	// } else if user.Type == rest.VisitorType_RealName {
	// 	chatReq.VisitorType = constant.RealName
	// } else {
	// 	chatReq.VisitorType = constant.Anonymous
	// }
	chatReq.CallType = constant.DebugChat
	oteltrace.SetConversationID(ctx, chatReq.ConversationID)

	channel, err := h.agentSvc.Chat(ctx, chatReq)
	oteltrace.SetConversationID(ctx, chatReq.ConversationID)

	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Debug] chat error: %v", err.Error()), err)
		h.logger.Errorf("[Debug] chat error: %v", err.Error())
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
			ttftFlag := false

			defer close(done)

			for data := range channel {
				// NOTE: 遇到错误时，不能break，否则会关闭channel导致对话结束
				if !ttftFlag {
					ttftFlag = true

					h.logger.Infof("[Debug] ttft: %d ms", cutil.GetCurrentMSTimestamp()-reqStartTime)
				}

				_, _ = c.Writer.Write(data)
				c.Writer.Flush()
			}
		}()
		<-done
	} else {
		var res any
		for data := range channel {
			if err := sonic.Unmarshal(data, &res); err != nil {
				rest.ReplyError(c, err)
				return
			}
			// fmt.Println(res)
		}

		if res == nil {
			h.logger.Errorf("[Debug] chat failed: res is nil")
			c.JSON(http.StatusInternalServerError, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).
				WithErrorDetails("[Debug] chat failed: res is nil").BaseError)

			return
		}

		resultMap := res.(map[string]any)
		if _, ok := resultMap["BaseError"]; ok {
			// *rest.HTTPError
			c.JSON(http.StatusInternalServerError, resultMap["BaseError"])
		} else {
			c.JSON(http.StatusOK, resultMap)
		}
	}
}
