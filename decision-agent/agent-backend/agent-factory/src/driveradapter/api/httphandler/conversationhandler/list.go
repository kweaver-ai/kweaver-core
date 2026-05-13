package conversationhandler

import (
	"fmt"
	"net/http"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// List 获取对话列表
// @Summary      获取对话列表
// @Description  获取指定应用的对话列表，支持分页
// @Tags         对话管理
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "应用 Key"
// @Param        page     query     int     false "页码，默认1"  default(1)
// @Param        size     query     int     false "每页数量，默认10"  default(10)
// @Success      200       {string}  string  "成功"
// @Failure      400     {object}  swagger.APIError  "请求参数错误"
// @Failure      500     {object}  swagger.APIError  "服务器内部错误"
// @Router       /v1/app/{app_key}/conversation [get]
// @Security     BearerAuth
func (h *conversationHTTPHandler) List(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	var req conversationreq.ListReq

	if err := c.ShouldBind(&req); err != nil {
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, httpErr)

		return
	}

	req.AgentAPPKey = c.Param("app_key")
	user := chelper.GetVisitorFromCtx(ctx)
	req.UserId = user.ID

	list, total, err := h.conversationSvc.List(ctx, req)
	if err != nil {
		h.logger.Errorf("list conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err)
		otellog.LogError(c.Request.Context(), fmt.Sprintf("list conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.ConversationGetListFailed).WithErrorDetails(
			"list conversation failed:" + errors.Cause(err).Error(),
		)
		// 返回错误
		rest.ReplyError(c, httpErr)

		return
	}

	rt := map[string]interface{}{
		"entries": list,
		"total":   total,
	}
	// 返回成功
	rest.ReplyOK(c, http.StatusOK, rt)
}
