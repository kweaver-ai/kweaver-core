package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      删除agent
// @Description  删除agent
// @Tags         agent,agent-internal
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Success      204  {object}  object  "请求成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id} [delete]
func (h *daConfHTTPHandler) Delete(c *gin.Context) {
	// 判断是否是私有API
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取id
	id := c.Param("agent_id")
	if id == "" {
		err := capierr.New400Err(c, "id is empty")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, ""), &err.BaseError)
		}

		rest.ReplyError(c, err)

		return
	}

	// 2. 获取ownerUid
	uid := chelper.GetUserIDFromCtx(c)
	if !isPrivate && uid == "" {
		err := capierr.New400Err(c, "uid is empty")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, ""), &err.BaseError)
		}

		rest.ReplyError(c, err)

		return
	}

	// 3. 删除
	auditLogInfo, err := h.daConfSvc.Delete(c, id, uid, isPrivate)
	if err != nil {
		httpErr := rest.NewHTTPError(c, http.StatusInternalServerError, apierr.AgentFactory_InternalError).WithErrorDetails(err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, auditLogInfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewWarnLog(audit.OPERATION, auditconstant.DELETE, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject(id, auditLogInfo.Name), audit.SUCCESS, "")
	}

	// 3. 返回结果
	c.Status(http.StatusNoContent)
}
