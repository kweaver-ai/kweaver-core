package conversationhandler

import (
	"fmt"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// @Summary      删除对话
// @Description  删除对话
// @Tags         对话管理
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "app_key"
// @Param        id  path      string  true  "id"
// @Success      204  {object}  object  "删除成功"
// @Failure      404  {object}  object  "对话不存在"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/app/{app_key}/conversation/{id} [delete]
func (h *conversationHTTPHandler) Delete(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	id := c.Param("id")
	if id == "" {
		h.logger.Errorf("[Delete] id is empty")
		otellog.LogError(c.Request.Context(), "[Delete] id is empty", nil)
		oteltrace.EndSpan(c.Request.Context(), nil)
		httpErr := capierr.New400Err(c, "id is empty")
		rest.ReplyError(c, httpErr)

		return
	}

	err := h.conversationSvc.Delete(ctx, id)
	if err != nil {
		h.logger.Errorf("delete conversation failed, cause: %v, err trace: %+v\n", errors.Cause(err), err)
		otellog.LogError(c.Request.Context(), fmt.Sprintf("delete conversation failed, cause: %v, err trace: %+v\n", errors.Cause(err), err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		// 返回错误
		rest.ReplyError(c, err)

		return
	}

	rest.ReplyOK(c, http.StatusNoContent, "")
}

// @Summary      删除所有对话
// @Description  删除所有对话
// @Tags         对话管理
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "app_key"
// @Success      204  {object}  object  "所有对话删除成功"
// @Failure      500  {object}  object  "服务器内部错误"
// @Security     BearerAuth
// @Router       /v1/app/{app_key}/conversation [delete]
func (h *conversationHTTPHandler) DeleteByAPPKey(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	appKey := c.Param("app_key")

	if appKey == "" {
		h.logger.Errorf("[DeleteByAPPKey] appKey is empty")
		otellog.LogError(c.Request.Context(), "[DeleteByAPPKey] appKey is empty", nil)
		oteltrace.EndSpan(c.Request.Context(), nil)
		err := capierr.New400Err(c, "appKey is empty")
		rest.ReplyError(c, err)

		return
	}

	err := h.conversationSvc.DeleteByAppKey(ctx, appKey)
	if err != nil {
		h.logger.Errorf("delete conversation failed, cause: %v, err trace: %+v\n", errors.Cause(err), err)
		otellog.LogError(c.Request.Context(), fmt.Sprintf("delete conversation failed, cause: %v, err trace: %+v\n", errors.Cause(err), err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		httpErr := rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError, apierr.ConversationDeleteFailed).WithErrorDetails(fmt.Sprintf("delete conversation failed: %s", err.Error()))
		rest.ReplyError(c, httpErr)

		return
	}

	rest.ReplyOK(c, http.StatusNoContent, "")
}
