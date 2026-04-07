package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil/crest"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Create 创建agent
// @Summary      创建agent
// @Description  创建一个新的 Agent 配置
// @Tags         agent,agent-internal
// @Accept       json
// @Produce      json
// @Param        request  body      agentconfigreq.CreateReq  true  "Agent 配置"
// @Success      201      {object}  agentconfigresp.DetailRes  "成功"
// @Failure      400      {object}  swagger.APIError   "请求参数错误"
// @Failure      401      {object}  swagger.APIError   "未授权"
// @Failure      403      {object}  swagger.APIError   "禁止访问"
// @Failure      500      {object}  swagger.APIError   "服务器内部错误"
// @Router       /v3/agent [post]
// @Security     BearerAuth
func (h *daConfHTTPHandler) Create(c *gin.Context) {
	// 1. 获取请求参数
	var req agentconfigreq.CreateReq

	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", req.Name), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}
	// 1.1 设置is_private字段
	setIsPrivate2Req(c, req.UpdateReq)

	// 2. 验证请求参数
	if err := req.ReqCheckWithCtx(c); err != nil {
		httpError, ok := crest.GetRestHttpErr(err)
		if !ok {
			httpError = capierr.New400Err(c, err.Error())
		}

		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", req.Name), &httpError.BaseError)
		}

		_ = c.Error(httpError)

		return
	}

	// 3. 创建
	id, err := h.daConfSvc.Create(c, &req)
	if err != nil {
		httpErr := rest.NewHTTPError(c, http.StatusInternalServerError, apierr.AgentFactory_InternalError).WithErrorDetails(err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", req.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	// 4. 返回结果
	res := &agentconfigresp.CreateRes{
		ID:      id,
		Version: daconstant.AgentVersionUnpublished,
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject("", req.Name), "")
	}

	c.JSON(http.StatusCreated, res)
}
