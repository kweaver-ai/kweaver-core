package releasehandler

import (
	"net/http"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// @Summary      发布智能体
// @Description  可以通过提交已有的 agent_id 发布智能体，或者提交 agent_config 创建并发布智能体
// @Tags         发布相关,发布相关-internal
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        request  body      releasereq.UpdatePublishInfoReq  true  "请求体"
// @Success      201  {object}  object  "发布成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id}/publish [post]
func (h *releaseHandler) Publish(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	var err error

	req := releasereq.NewPublishReq()

	setIsPrivate2Req(c, req)

	req.UserID, err = chelper.GetUserIDFromGinContext(c)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	if req.UserID == "" {
		err = errors.New("[releaseHandler.Publish]user_id is empty")

		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	req.AgentID = c.Param("agent_id")

	err = c.ShouldBind(req)
	if err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, req))
		// todo error log
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	req.AgentID = c.Param("agent_id")

	if err = req.ReqCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(req.AgentID, ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	resp, auditloginfo, err := h.releaseSvc.Publish(ctx, req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(auditloginfo.ID, auditloginfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.PUBLISH, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject(auditloginfo.ID, auditloginfo.Name), "")
	}

	rest.ReplyOK(c, http.StatusCreated, resp)
}
