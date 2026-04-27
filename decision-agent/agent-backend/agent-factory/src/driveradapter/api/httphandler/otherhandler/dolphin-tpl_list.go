package otherhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/other/otherreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// @Summary      获取dolphin模板列表
// @Description  获取dolphin模板列表
// @Tags         其他
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "获取成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent/dolphin-tpl/list [post]
func (o *otherHTTPHandler) DolphinTplList(c *gin.Context) {
	// 1. 获取请求参数
	var req otherreq.DolphinTplListReq

	if err := c.ShouldBind(&req); err != nil {
		err = capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, err)

		return
	}

	// 1.1 config配置处理（如设置默认值等）
	err := agentconfigreq.HandleConfig(req.Config)
	if err != nil {
		err = errors.Wrap(err, "[DolphinTplList]: HandleConfig failed")
		_ = c.Error(err)

		return
	}

	//// 1.1 验证请求参数
	// if err := req.Config.ValObjCheckWithCtx(c, false); err != nil {
	//	err = capierr.New400Err(c, err.Error())
	//	rest.ReplyError(c, err)
	//
	//	return
	//}

	// 2. 调用服务层
	res, err := o.otherService.DolphinTplList(c, &req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
