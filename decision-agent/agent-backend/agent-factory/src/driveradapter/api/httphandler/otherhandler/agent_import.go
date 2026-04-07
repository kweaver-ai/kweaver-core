package otherhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// ImportAgent 导入agent数据
// @Summary      导入数据
// @Description  - 从导出的JSON文件中导入data agent数据 - 支持批量导入多个agent - agent标识（key）不可与已存在的重复 - 如果重复，需要先删除已存在的agent，再导入 - 重复时，会返回重复的agent列表（包括agent_id、agent_key和agent_name） - 只能导入到\"我的个人空间\"中
// @Tags         其他
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "导入结果"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-inout/import [post]
func (o *otherHTTPHandler) ImportAgent(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取请求参数
	var req agentinoutreq.ImportReq

	if err := c.ShouldBind(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.IMPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 1.1 检查文件是否存在
	if req.File == nil {
		err := capierr.New400Err(c, "未上传文件")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.IMPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &err.BaseError)
		}

		_ = c.Error(err)

		return
	}

	// 2. 调用服务层
	resp, err := o.agentInOutSvc.Import(c, &req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.IMPORT, audit.TransforOperator(*visitor),
				auditconstant.GenerateAgentAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.IMPORT, audit.TransforOperator(*visitor),
			auditconstant.GenerateAgentAuditObject("", ""), "")
	}
	// 3. 返回响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
