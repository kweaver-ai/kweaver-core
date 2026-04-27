package producthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/product/productreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Update 编辑产品
// @Summary      编辑产品
// @Description  更新产品信息
// @Tags         产品
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "产品 ID"
// @Param        product  body      productreq.UpdateReq     true  "产品信息"
// @Success      204
// @Failure      400  {object}  swagger.APIError           "请求参数错误"
// @Failure      401  {object}  swagger.APIError           "未授权"
// @Failure      403  {object}  swagger.APIError           "禁止访问"
// @Failure      500  {object}  swagger.APIError           "服务器内部错误"
// @Router       /v3/product/{id} [put]
// @Security     BearerAuth
func (h *productHTTPHandler) Update(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取path参数
	id := c.Param("id")
	if id == "" {
		err := capierr.New400Err(c, "id is empty")
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", ""), &err.BaseError)
		}

		_ = c.Error(err)

		return
	}

	// 2. 获取请求体
	var req productreq.UpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 3. 校验请求体
	if err := req.CustomCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 4. 调用服务层更新
	auditloginfo, err := h.productService.Update(c, &req, cutil.MustParseInt64(id))
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject(auditloginfo.ID, auditloginfo.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.UPDATE, audit.TransforOperator(*visitor),
			auditconstant.GenerateProductAuditObject(auditloginfo.ID, auditloginfo.Name), "")
	}

	c.Status(http.StatusNoContent)
}
