package conversationhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// Detail 获取对话详情
// @Summary      获取对话详情
// @Description  根据会话 ID 获取会话详细信息
// @Tags         对话管理
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "应用 Key"
// @Param        id       path      string  true  "会话 ID"
// @Success      200       {string}  string  "成功"
// @Failure      404     {object}  swagger.APIError  "会话不存在"
// @Failure      500     {object}  swagger.APIError  "服务器内部错误"
// @Router       /v1/app/{app_key}/conversation/{id} [get]
// @Security     BearerAuth
func (h *conversationHTTPHandler) Detail(c *gin.Context) {
	// 1. 获取id
	id := c.Param("id")
	if id == "" {
		h.logger.Errorf("[Detail] id is empty")
		otellog.LogError(c.Request.Context(), "[Detail] id is empty", nil)
		oteltrace.EndSpan(c.Request.Context(), nil)
		err := capierr.New400Err(c, "id is empty")
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取详情
	res, err := h.conversationSvc.Detail(c, id)
	if err != nil {
		h.logger.Errorf("get conversation detail failed, cause: %v, err trace: %+v\n", errors.Cause(err), err)
		otellog.LogError(c.Request.Context(), fmt.Sprintf("get conversation detail failed, cause: %v, err trace: %+v\n", errors.Cause(err), err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		httpErr := rest.NewHTTPError(c.Request.Context(), http.StatusInternalServerError,
			apierr.ConversationDetailFailed).WithErrorDetails(fmt.Sprintf("get conversation detail failed %s", err.Error()))
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 返回结果
	c.JSON(http.StatusOK, res)
}
