package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil/crest"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      编辑agent
// @Description  编辑agent
// @Tags         agent,agent-internal
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Param        request  body      object  false  "请求体"
// @Success      204  {object}  object  "请求成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id} [put]
func (h *daConfHTTPHandler) Update(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取id
	id := c.Param("agent_id")
	if id == "" {
		err := capierr.New400Err(c, "id is empty")
		rest.ReplyError(c, err)

		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, ""), &err.BaseError)
		}

		return
	}

	// 2. 获取请求参数
	var req agentconfigreq.UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 2.1 设置is_private字段
	setIsPrivate2Req(c, &req)

	// 3. 验证请求参数
	if err := req.ReqCheckWithCtx(c); err != nil {
		httpError, ok := crest.GetRestHttpErr(err)
		if !ok {
			httpError = capierr.New400Err(c, err.Error())
		}

		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, ""), &httpError.BaseError)
		}

		_ = c.Error(httpError)

		return
	}

	// 3.1 custom check
	if err := req.CustomCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	if cenvhelper.IsLocalDev(cenvhelper.RunScenario_Aaron_Local_Dev) {
		req.ProductKey = "dip"
		if req.Name == "open_plan_mode_v1" && !req.Config.IsDolphinMode.Bool() {
			req.Config.PlanMode = daconfvalobj.NewPlanMode(true)
		} else {
			req.Config.PlanMode = daconfvalobj.NewPlanMode(false)
		}
	}

	// 4. 更新
	auditLogInfo, err := h.daConfSvc.Update(c, &req, id)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(id, auditLogInfo.OldName), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject(id, auditLogInfo.OldName), "")
	}

	c.Status(http.StatusNoContent)
}
