package producthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/auditconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Create 创建产品
// @Summary      创建产品
// @Description  创建一个新的产品
// @Tags         产品
// @Accept       json
// @Produce      json
// @Param        product  body      productreq.CreateReq  true  "产品信息"
// @Success      201     {object}  productresp.CreateRes  "创建成功"
// @Failure      400     {object}  swagger.APIError         "请求参数错误"
// @Failure      401     {object}  swagger.APIError         "未授权"
// @Failure      403     {object}  swagger.APIError         "禁止访问"
// @Failure      500     {object}  swagger.APIError         "服务器内部错误"
// @Router       /v3/product [post]
// @Security     BearerAuth
func (h *productHTTPHandler) Create(c *gin.Context) {
	isPrivate := capimiddleware.IsInternalAPI(c)

	var visitor *rest.Visitor

	if !isPrivate {
		visitor = chelper.GetVisitorFromCtx(c.Request.Context())
	}
	// 1. 获取请求体
	var req productreq.CreateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", ""), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 2. 校验请求体
	if err := req.CustomCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", req.Name), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 3. 调用服务层创建
	key, err := h.productService.Create(c, &req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", req.Name), &httpErr.BaseError)
		}

		_ = c.Error(err)

		return
	}

	// 4. 获取详情
	ret, err := h.productService.GetByKey(c, key)
	if err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		if !isPrivate {
			audit.NewWarnLogWithError(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
				auditconstant.GenerateProductAuditObject("", req.Name), &httpErr.BaseError)
		}

		_ = c.Error(httpErr)

		return
	}

	// 5. 返回结果
	res := &productresp.CreateRes{
		Key: key,
		ID:  ret.ID,
	}

	if !isPrivate {
		audit.NewInfoLog(audit.OPERATION, auditconstant.CREATE, audit.TransforOperator(*visitor),
			auditconstant.GenerateProductAuditObject("", req.Name), "")
	}

	c.JSON(http.StatusCreated, res)
}
