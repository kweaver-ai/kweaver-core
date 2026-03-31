package producthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// Detail 获取产品详情
// @Summary      获取产品详情
// @Description  根据产品 ID 获取产品详细信息
// @Tags         产品
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "产品 ID"
// @Success      200  {object}  productresp.DetailRes  "成功"
// @Failure      400  {object}  swagger.APIError          "请求参数错误"
// @Failure      401  {object}  swagger.APIError          "未授权"
// @Failure      403  {object}  swagger.APIError          "禁止访问"
// @Failure      500  {object}  swagger.APIError          "服务器内部错误"
// @Router       /v3/product/{id} [get]
// @Security     BearerAuth
func (h *productHTTPHandler) Detail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		err := capierr.New400Err(c, "id is empty")
		rest.ReplyError(c, err)

		return
	}

	res, err := h.productService.Detail(c, cutil.MustParseInt64(id))
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
