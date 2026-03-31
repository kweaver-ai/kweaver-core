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
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

// Chat Agent 对话接口
// @Summary      对话
// @Description  与 Agent 进行对话交互，支持流式响应
// @Tags         对话,对话-internal
// @Accept       json
// @Produce      json
// @Param        app_key  path      string        true  "应用 Key"
// @Param        request  body      swagger.ChatReq  true  "对话请求"
// @Success      200       {object}  swagger.ChatResp  "成功"
// @Failure      400       {object}  swagger.APIError   "请求参数错误"
// @Failure      500       {object}  swagger.APIError   "服务器内部错误"
// @Router       /v1/app/{app_key}/chat/completion [post]
// @Security     BearerAuth
func (h *agentHTTPHandler) Chat(c *gin.Context) {
	reqStartTime := cutil.GetCurrentMSTimestamp()
	// 1. app_key
	agentAPPKey := c.Param("app_key")
	if agentAPPKey == "" {
		err := capierr.New400Err(c, "[Chat] app key is empty")
		h.logger.Errorf("[Chat] app key is empty: %v", err)
		otellog.LogError(c, "[Chat] app key is empty", err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var req agentreq.ChatReq = agentreq.ChatReq{
		Stream:       true,
		IncStream:    true,
		AgentVersion: "latest",
		// ConfirmPlan:     true,
		ExecutorVersion: "v2",
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[Chat] should bind json err: %v", err)
		otellog.LogError(c, fmt.Sprintf("[Chat] should bind json err: %v", err), err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[Chat] should bind json err: %v", err))
		rest.ReplyError(c, httpErr)

		return
	}

	// 本地调试：可取消注释以关闭 IncStream
	// if cenvhelper.IsLocalDev(cenvhelper.RunScenario_Aaron_Local_Dev) {
	// 	req.IncStream = false
	// }

	req.AgentAPPKey = agentAPPKey
	req.ReqStartTime = reqStartTime
	// NOTE: 获取用户ID
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		httpErr := capierr.New404Err(c, "[Chat] user not found")
		otellog.LogError(c, "[Chat] user not found", nil)
		h.logger.Errorf("[Chat] user not found: %v", httpErr)
		rest.ReplyError(c, httpErr)

		return
	}

	// // 检查应用账号，应用账号应使用API Chat接口
	// if user.Type == rest.VisitorType_App {
	// 	errMsg := "应用账号应该使用API Chat接口"
	// 	h.logger.Errorf("[Chat] %s", errMsg)
	// 	o11y.Error(c, fmt.Sprintf("[Chat] %s", errMsg))
	// 	httpErr := capierr.New400Err(c, fmt.Sprintf("[Chat] %s", errMsg))
	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	req.UserID = user.ID
	req.XAccountID = user.ID
	req.XAccountType.LoadFromMDLVisitorType(user.Type)
	req.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)
	req.Token = strings.TrimPrefix(user.TokenID, "Bearer ")

	// 如果AgentRunID为空，则生成新的ID
	if req.AgentRunID == "" {
		req.AgentRunID = cutil.UlidMake()
	}

	// if user.Type == rest.VisitorType_App {
	// 	req.VisitorType = constant.Business
	// } else if user.Type == rest.VisitorType_RealName {
	// 	req.VisitorType = constant.RealName
	// } else {
	// 	req.VisitorType = constant.Anonymous
	// }
	if req.ExecutorVersion == "" {
		req.ExecutorVersion = "v2"
	}

	req.CallType = constant.Chat
	ctx := c.Request.Context()
	oteltrace.SetConversationID(ctx, req.ConversationID)

	// 3. 调用服务
	channel, err := h.agentSvc.Chat(ctx, &req)
	oteltrace.SetConversationID(ctx, req.ConversationID)

	if err != nil {
		otellog.LogError(ctx, fmt.Sprintf("[Chat] chat failed: %v", err.Error()), err)
		h.logger.Errorf("[Chat] chat failed: %v", err.Error())
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
			ttft := false

			defer close(done)
			// NOTE: 清空channel，直到channel关闭，再退出；
			drainFunc := func() {
				for range channel {
				}
			}

			for {
				select {
				case data, ok := <-channel:
					// NOTE: 如果channel关闭，则退出
					if !ok {
						h.logger.Debugf("[Chat] channel closed")
						return
					}
					// NOTE: 往SSE中写入数据
					if !ttft {
						ttft = true

						h.logger.Infof("[Chat] ttft: %d ms", cutil.GetCurrentMSTimestamp()-reqStartTime)
					}

					_, err := c.Writer.Write(data)
					if err != nil {
						h.logger.Errorf("[Chat] write data err: %v", err)
						otellog.LogError(ctx, fmt.Sprintf("[Chat] write data err: %v", err), err)
						// NOTE:如果出错，清空channel，直到channel关闭，再退出；
						// NOTE: 如果channel未关闭直接退出，会导致管道阻塞，对话Process无法继续
						drainFunc()

						return
					} else {
						// NOTE: 如果写入成功，则刷新缓冲区
						c.Writer.Flush()
					}
				case <-c.Writer.CloseNotify():
					// NOTE: 如果SSE连接关闭，则清空channel，直到channel关闭，再退出；
					h.logger.Debugf("[Chat] SSE connection closed")
					drainFunc()

					return
				case <-c.Request.Context().Done():
					// NOTE: 如果请求上下文关闭，则清空channel，直到channel关闭，再退出；
					h.logger.Debugf("[Chat] request context done")
					drainFunc()

					return
				}
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
			h.logger.Errorf("[Chat] chat failed: res is nil")
			c.JSON(http.StatusInternalServerError, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).
				WithErrorDetails("[Chat] chat failed: res is nil").BaseError)

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
