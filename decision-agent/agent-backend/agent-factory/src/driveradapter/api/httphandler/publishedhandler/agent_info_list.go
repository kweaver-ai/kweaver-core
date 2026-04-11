package publishedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      已发布智能体信息列表
// @Description  已发布智能体信息列表
// @Tags         已发布
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/published/agent-info-list [post]
func (h *publishedHandler) PubedAgentInfoList(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	// 1. 构建请求参数
	req := &pubedreq.PAInfoListReq{}
	if err := c.ShouldBind(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, req))
		_ = c.Error(httpErr)

		return
	}

	// 1.1 校验请求参数
	if err := req.ReqCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	// 1.2 设置默认值
	req.HlDefaultVal()

	// 2. 调用service层获取已发布智能体列表
	resp, err := h.publishedSvc.GetPubedAgentInfoList(ctx, req)
	if err != nil {
		_ = c.Error(err)

		return
	}

	// 3. 返回成功
	rest.ReplyOK(c, http.StatusOK, resp)
}
