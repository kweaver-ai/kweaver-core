package producthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// List 获取产品列表
// @Summary      获取产品列表
// @Description  获取产品列表，支持分页
// @Tags         产品
// @Accept       json
// @Produce      json
// @Param        page     query     int  false  "页码，默认1"  default(1)
// @Param        page_size query     int  false  "每页数量，默认10"  default(10)
// @Success      200  {array}   productresp.ListRes  "成功"
// @Failure      400  {object}  swagger.APIError        "请求参数错误"
// @Failure      401  {object}  swagger.APIError        "未授权"
// @Failure      403  {object}  swagger.APIError        "禁止访问"
// @Failure      500  {object}  swagger.APIError        "服务器内部错误"
// @Router       /v3/product [get]
// @Security     BearerAuth
func (h *productHTTPHandler) List(c *gin.Context) {
	// 1. 获取请求参数
	var req productreq.ListReq

	if err := c.ShouldBind(&req); err != nil {
		err = capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, err)

		return
	}

	// 2. 调用服务层
	res, err := h.productService.List(c, req.GetOffset(), req.GetLimit())
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
