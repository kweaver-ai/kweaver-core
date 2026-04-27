package publishedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// PublishedAgentList 已发布智能体列表
// @Summary      已发布智能体列表
// @Description  获取已发布的智能体列表，支持分页和筛选
// @Tags         已发布
// @Accept       json
// @Produce      json
// @Param        request  body      pubedreq.PubedAgentListReq  true  "查询参数"
// @Success      200       {string}  string  "成功"
// @Failure      400      {object}  swagger.APIError  "请求参数错误"
// @Failure      401      {object}  swagger.APIError  "未授权"
// @Failure      403      {object}  swagger.APIError  "禁止访问"
// @Failure      500      {object}  swagger.APIError  "服务器内部错误"
// @Router       /v3/published/agent [post]
// @Security     BearerAuth
func (h *publishedHandler) PublishedAgentList(c *gin.Context) {
	// 构建请求参数
	req := &pubedreq.PubedAgentListReq{}
	if err := c.ShouldBind(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, req))
		_ = c.Error(httpErr)

		return
	}

	if err := req.CustomCheck(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	// 如果未全局禁用业务域且 business_domain_ids 为空，则回填当前上下文中的业务域
	if len(req.BusinessDomainIDs) == 0 && !global.GConfig.IsBizDomainDisabled() {
		bdID := chelper.GetBizDomainIDFromCtx(c)
		req.BusinessDomainIDs = []string{bdID}
	}

	if err := req.LoadMarkerStr(); err != nil {
		httpErr := capierr.New400Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	req.IDs = cutil.RemoveEmptyStrFromSlice(req.IDs)

	// 调用service层获取已发布智能体列表
	resp, err := h.publishedSvc.GetPublishedAgentList(c, req)
	if err != nil {
		httpErr := capierr.New500Err(c, err.Error())
		_ = c.Error(httpErr)

		return
	}

	// 返回成功
	rest.ReplyOK(c, http.StatusOK, resp)
}
