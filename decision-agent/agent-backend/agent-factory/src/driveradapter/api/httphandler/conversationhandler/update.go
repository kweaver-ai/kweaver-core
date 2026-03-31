package conversationhandler

import (
	"fmt"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// @Summary      编辑对话(修改标题)
// @Description  编辑对话(修改标题)
// @Tags         对话管理
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "app_key"
// @Param        id  path      string  true  "id"
// @Param        request  body      object  true  "请求体"
// @Success      204  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/app/{app_key}/conversation/{id} [put]
func (h *conversationHTTPHandler) Update(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)
	// 1. 获取请求参数
	var req conversationreq.UpdateReq

	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[Update] should bind json error: %v", errors.Cause(err))
		o11y.Error(c, fmt.Sprintf("[Update] should bind json error: %v", errors.Cause(err)))
		err = capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, err)

		return
	}

	// 2. 验证请求参数
	if err := req.ReqCheck(); err != nil {
		h.logger.Errorf("[Update] req check error: %v", errors.Cause(err))
		o11y.Error(c, fmt.Sprintf("[Update] req check error: %v", errors.Cause(err)))
		err = capierr.New400Err(c, err.Error())
		rest.ReplyError(c, err)

		return
	}

	err := h.conversationSvc.Update(ctx, req)
	if err != nil {
		h.logger.Errorf("update conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err)
		o11y.Error(c, fmt.Sprintf("update conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err))
		// 返回错误
		rest.ReplyError(c, err)

		return
	}

	rest.ReplyOK(c, http.StatusNoContent, "")
}
