package tplhandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/ginhelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      删除agent模板
// @Description  删除agent模板
// @Tags         模板
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "id"
// @Success      204  {object}  object  "请求成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-tpl/{id} [delete]
func (h *daTplHTTPHandler) Delete(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(ctx)
	}

	id, err := ginhelper.GetParmIDInt64(c)
	if err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentTemplateAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 获取ownerUid
	uid := chelper.GetUserIDFromCtx(c)
	if !isPrivate && uid == "" {
		httpErr := capierr.New400Err(c, "uid is empty")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentTemplateAuditObject(strconv.FormatInt(id, 10), ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	auditloginfo, err := h.daTplSvc.Delete(ctx, id, uid, isPrivate)
	if err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentTemplateAuditObject(strconv.FormatInt(id, 10), auditloginfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewWarnLog(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentTemplateAuditObject(strconv.FormatInt(id, 10), auditloginfo.Name), audit.SUCCESS, "")
	}

	c.Status(http.StatusNoContent)
}
