package conversationhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// @Summary      标记对话已读
// @Description  标记对话已读
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
// @Router       /v1/app/{app_key}/conversation/{id}/mark_read [put]
func (h *conversationHTTPHandler) MarkRead(c *gin.Context) {
	ctx := rest.GetLanguageCtx(c)

	// 1. 获取id
	id := c.Param("id")
	if id == "" {
		h.logger.Errorf("[MarkRead] id is empty")
		otellog.LogError(c.Request.Context(), "[MarkRead] id is empty", nil)
		oteltrace.EndSpan(c.Request.Context(), nil)
		err := capierr.New400Err(c, "id is empty")
		rest.ReplyError(c, err)

		return
	}

	req := conversationreq.MarkReadReq{}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[MarkRead] should bind json error: %v", errors.Cause(err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[MarkRead] should bind json error: %v", errors.Cause(err)), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		err = capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, err)

		return
	}

	// 2. 验证请求参数
	if err := req.ReqCheck(); err != nil {
		h.logger.Errorf("[MarkRead] req check error: %v", errors.Cause(err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[MarkRead] req check error: %v", errors.Cause(err)), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		err = capierr.New400Err(c, err.Error())
		rest.ReplyError(c, err)

		return
	}

	// 3. 标记已读
	err := h.conversationSvc.MarkRead(ctx, id, req.LastestReadIdx)
	if err != nil {
		h.logger.Errorf("mark read conversation failed, cause: %v, err trace: %+v\n", errors.Cause(err), err)
		otellog.LogError(c.Request.Context(), fmt.Sprintf("mark read conversation failed, cause: %v, err trace: %+v\n", errors.Cause(err), err), err)
		oteltrace.EndSpan(c.Request.Context(), err)

		// 返回错误
		rest.ReplyError(c, err)

		return
	}
	// 4. 返回结果
	c.JSON(http.StatusNoContent, "")
}
