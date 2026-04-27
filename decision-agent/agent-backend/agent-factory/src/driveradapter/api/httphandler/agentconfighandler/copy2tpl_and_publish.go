package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Copy2TplAndPublish 复制Agent为模板并发布
// @Summary      复制agent为模板并发布
// @Description  复制agent为模板并发布
// @Tags         agent,模板
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        request  body      object  true  "请求体"
// @Success      201  {object}  object  "操作成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id}/copy2tpl-and-publish [post]
func (h *daConfHTTPHandler) Copy2TplAndPublish(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(ctx)
	}
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	if agentID == "" {
		err := capierr.New400Err(c, "agent_id不能为空")
		_ = c.Error(err)

		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY_PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(agentID, ""), &err.BaseError)
		}

		return
	}

	req := agenttplreq.NewPublishReq()
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(ctx, chelper.ErrMsg(err, req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY_PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(agentID, ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	resp, auditLogInfo, err := h.daConfSvc.Copy2TplAndPublish(ctx, agentID, req)
	if err != nil {
		httpErr := rest.NewHTTPError(c, http.StatusInternalServerError, apierr.AgentFactory_InternalError).WithErrorDetails(err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY_PUBLISH, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(agentID, auditLogInfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.COPY_PUBLISH, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject(agentID, auditLogInfo.Name), "")
	}

	c.JSON(http.StatusOK, resp)
}
