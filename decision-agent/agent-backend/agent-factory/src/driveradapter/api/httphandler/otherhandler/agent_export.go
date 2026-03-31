package otherhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// ExportAgent 导出agent数据
// @Summary      导出数据
// @Description  - 批量导出指定的data agent数据 - 支持导出多个agent - 返回JSON格式的导出文件 - 只能导出\"我的个人空间\"中的agent - 如果提供的agent_id有不存在的，会返回404错误
// @Tags         其他
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "导出成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-inout/export [post]
func (o *otherHTTPHandler) ExportAgent(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取请求参数
	var req agentinoutreq.ExportReq

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.EXPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 1.1 校验请求参数
	if err := req.CustomCheckAndDedupl(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.EXPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 2. 调用服务层
	resp, filename, err := o.agentInOutSvc.Export(c, &req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.EXPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	// 3. 序列化响应数据
	jsonData, err := cutil.JSON().MarshalToString(resp)
	if err != nil {
		httpErr := capierr.New500Err(c, "序列化导出数据失败")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.EXPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.EXPORT, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject("", ""), "")
	}
	// 4. 设置响应头并返回文件
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.String(http.StatusOK, jsonData)
}

// ExportAgentGet GET 方法导出（临时测试用）
func (o *otherHTTPHandler) ExportAgentGet(c *gin.Context) {
	// 1. 获取请求参数
	req := &agentinoutreq.ExportReq{
		AgentIDs: c.QueryArray("agent_ids"),
	}

	// 1.1 校验请求参数
	if err := req.CustomCheckAndDedupl(); err != nil {
		err = capierr.New400Err(c, err.Error())
		_ = c.Error(err)

		return
	}

	// 2. 调用服务层
	resp, filename, err := o.agentInOutSvc.Export(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// 3. 序列化响应数据
	jsonData, err := cutil.JSON().MarshalToString(resp)
	if err != nil {
		err = capierr.New500Err(c, "序列化导出数据失败")
		_ = c.Error(err)

		return
	}

	// 4. 设置响应头并返回文件
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.String(http.StatusOK, jsonData)
}
