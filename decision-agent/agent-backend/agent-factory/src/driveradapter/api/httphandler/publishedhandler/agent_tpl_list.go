package publishedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// PubedTplList 已发布模板列表
// @Summary      已发布模板列表
// @Description  已发布模板列表
// @Tags         已发布,模板
// @Accept       json
// @Produce      json
// @Param        pagination_marker_str  query      string  false  "- 分页marker（用于获取下一页数据） - base64编码的json字符串 - base64编码前的json格式： ``` { \"last_pubed_tpl_id\": 111 } ```"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/published/agent-tpl [get]
func (h *publishedHandler) PubedTplList(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	// 构建请求参数
	req := pubedreq.PubedTplListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		_ = c.Error(httpErr)

		return
	}

	err := req.LoadMarkerStr()
	if err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	// 调用service层获取已发布模板列表
	resp, err := h.publishedSvc.GetPubedTplList(ctx, &req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	// 返回成功
	rest.ReplyOK(c, http.StatusOK, resp)
}
