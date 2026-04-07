package conversationhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// Init 初始化对话
// @Summary      初始化对话
// @Description  创建一个新的会话会话
// @Tags         对话管理
// @Accept       json
// @Produce      json
// @Param        app_key  path      string                     true  "应用 Key"
// @Param        request  body      conversationreq.InitReq  true  "初始化请求"
// @Success      200       {string}  string  "成功"
// @Failure      400      {object}  swagger.APIError   "请求参数错误"
// @Failure      500      {object}  swagger.APIError   "服务器内部错误"
// @Router       /v1/app/{app_key}/conversation [post]
// @Security     BearerAuth
func (h *conversationHTTPHandler) Init(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)
	user := chelper.GetVisitorFromCtx(c)

	agentAPPKey := c.Param("app_key")
	if agentAPPKey == "" {
		h.logger.Errorf("[Init] agent_app_key is empty")
		otellog.LogError(c.Request.Context(), "[Init] agent_app_key is empty", nil)
		oteltrace.EndSpan(c.Request.Context(), nil)

		httpErr := capierr.New400Err(ctx, "agent_app_key is empty")
		rest.ReplyError(c, httpErr)

		return
	}

	// 1. 获取请求参数
	var req conversationreq.InitReq

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[Init] should bind json error: %v", errors.Cause(err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[Init] should bind json error: %v", errors.Cause(err)), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 验证请求参数
	if err := req.ReqCheck(); err != nil {
		h.logger.Errorf("[Init] req check error: %v", errors.Cause(err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[Init] req check error: %v", errors.Cause(err)), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		httpErr := capierr.New400Err(c, err.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	req.UserID = user.ID
	req.XAccountID = user.ID
	req.XAccountType.LoadFromMDLVisitorType(user.Type)
	req.XBusinessDomainID = chelper.GetBizDomainIDFromCtx(c)
	req.AgentAPPKey = agentAPPKey
	// visitor := chelper.GetVisitorFromCtx(c)
	// if visitor != nil {
	// 	if visitor.Type == rest.VisitorType_App {
	// 		req.VisitorType = "app"
	// 	} else if visitor.Type == rest.VisitorType_RealName {
	// 		req.VisitorType = "user"
	// 	} else {
	// 		req.VisitorType = "anonymous"
	// 	}
	// }
	// NOTE: 截取前50个字符
	if req.Title != "" {
		runes := []rune(req.Title)
		if len(runes) < 50 {
			req.Title = string(runes)
		} else {
			req.Title = string(runes[:50])
		}
	}

	if req.ExecutorVersion == "" {
		req.ExecutorVersion = "v2"
	}

	rt, err := h.conversationSvc.Init(ctx, req)
	if err != nil {
		h.logger.Errorf("init conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err)
		otellog.LogError(c.Request.Context(), fmt.Sprintf("init conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		httpErr := rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError,
			apierr.ConversationInitFailed).WithErrorDetails(fmt.Sprintf("get conversation detail failed %s", err.Error()))

		// 返回错误
		rest.ReplyError(c, httpErr)

		return
	}

	rest.ReplyOK(c, http.StatusOK, rt)
}
