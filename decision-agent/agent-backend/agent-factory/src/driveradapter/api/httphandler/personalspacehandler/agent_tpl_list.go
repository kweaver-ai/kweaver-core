package personalspacehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// AgentTplList 获取个人空间Agent模板列表
// @Summary      个人空间（开发）下Agent模板列表
// @Description  个人空间（开发）下Agent模板列表
// @Tags         个人空间（开发）,模板
// @Accept       json
// @Produce      json
// @Param        pagination_marker_str  query      string  false  "- 分页marker（用于获取下一页数据） - base64编码的json字符串 - base64编码前的json格式： ``` { \"updated_at\": 111, \"last_tpl_id\": 222 } ```"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/personal-space/agent-tpl-list [get]
func (h *PersonalSpaceHTTPHandler) AgentTplList(c *gin.Context) {
	// 1. 获取请求参数
	var req personalspacereq.AgentTplListReq

	if err := c.ShouldBindQuery(&req); err != nil {
		err = capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, err)

		return
	}

	// 2. 自定义参数校验
	if err := req.CustomCheck(); err != nil {
		err = capierr.New400Err(c, err.Error())
		rest.ReplyError(c, err)

		return
	}

	// 2.1 加载 marker
	if err := req.LoadMarkerStr(); err != nil {
		err = capierr.New400Err(c, err.Error())
		rest.ReplyError(c, err)

		return
	}

	// 3. 调用服务层
	res, err := h.personalSpaceService.AgentTplList(c, &req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 4. 返回响应
	rest.ReplyOK(c, http.StatusOK, res)
}
