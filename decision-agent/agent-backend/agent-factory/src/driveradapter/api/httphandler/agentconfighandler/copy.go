package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Copy 复制Agent
// @Summary      复制agent
// @Description  复制agent
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        agent_id  path      string  true  "agent_id"
// @Success      201  {object}  object  "复制成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/{agent_id}/copy [post]
func (h *daConfHTTPHandler) Copy(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	if agentID == "" {
		err := capierr.New400Err(c, "agent_id不能为空")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(agentID, ""), &err.BaseError)
		}

		_ = c.Error(err)

		return
	}

	// 2. 获取请求体参数
	var req agentconfigreq.CopyReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
			if !isPrivate {
				audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
					auditconstant.GenerateAgentAuditObject(agentID, ""), &httpErr.BaseError)
			}

			_ = c.Error(httpErr)

			return
		}
	}

	// 3. 参数校验
	if err := req.ReqCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(agentID, ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 4. 调用服务层
	res, auditLogInfo, err := h.daConfSvc.Copy(c, agentID, &req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject(auditLogInfo.ID, auditLogInfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.COPY, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject(auditLogInfo.ID, auditLogInfo.Name), "")
	}

	// 5. 返回结果
	c.JSON(http.StatusCreated, res)
}
