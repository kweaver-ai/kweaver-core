package tplhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/ginhelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Copy 复制智能体模板
// @Summary      复制智能体模板
// @Description  复制智能体模板
// @Tags         模板
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "id"
// @Success      201  {object}  object  "复制成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-tpl/{id}/copy [post]
func (h *daTplHTTPHandler) Copy(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(ctx)
	}

	tplID, err := ginhelper.GetParmIDInt64(c)
	if err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentTemplateAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	res, auditloginfo, err := h.daTplSvc.Copy(ctx, tplID)
	if err != nil {
		httpErr := rest.NewHTTPError(c, http.StatusInternalServerError, apierr.AgentFactory_InternalError).WithErrorDetails(err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentTemplateAuditObject("", auditloginfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentTemplateAuditObject("", auditloginfo.Name), "")
	}

	rest.ReplyOK(c, http.StatusCreated, res)
}
