package sessionhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/session/sessionreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/ctype"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Manage 管理对话session
// @Summary      管理对话session
// @Description  管理指定的对话会话状态
// @Tags         对话session管理
// @Accept       json
// @Produce      json
// @Param        conversation_id  path      string                 true  "会话 ID"
// @Param        request         body      sessionreq.ManageReq  true  "管理请求"
// @Success      200       {string}  string  "成功"
// @Failure      400      {object}  swagger.APIError  "请求参数错误"
// @Failure      500      {object}  swagger.APIError  "服务器内部错误"
// @Router       /v1/conversation/session/{conversation_id} [put]
// @Security     BearerAuth
func (h *sessionHTTPHandler) Manage(c *gin.Context) {
	// 1. 获取请求参数
	var req sessionreq.ManageReq

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, httpErr)

		return
	}

	// 获取path参数
	req.ConversationID = c.Param("conversation_id")
	if req.ConversationID == "" {
		httpErr := capierr.New400Err(c, "conversation_id不能为空")
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 获取visitor信息
	visitor := chelper.GetVisitorFromCtx(c)
	if visitor == nil {
		httpErr := capierr.New401Err(c, "[Manage] visitor not found")
		rest.ReplyError(c, httpErr)

		return
	}

	visitorInfo := &ctype.VisitorInfo{
		XBusinessDomainID: cenum.BizDomainID(chelper.GetBizDomainIDFromCtx(c)),
	}
	visitorInfo.XAccountID = visitor.ID
	visitorInfo.XAccountType.LoadFromMDLVisitorType(visitor.Type)

	// 3. 调用service
	resp, err := h.sessionSvc.Manage(c.Request.Context(), req, visitorInfo)
	if err != nil {
		errMsg := fmt.Sprintf("[sessionHTTPHandler][Manage] failed to manage session: %v", err)

		h.logger.Errorf(errMsg)

		httpErr := capierr.New500Err(c, errMsg)

		rest.ReplyError(c, httpErr)

		return
	}

	// 4. 返回结果
	rest.ReplyOK(c, http.StatusOK, resp)
}
