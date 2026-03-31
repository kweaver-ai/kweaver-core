package conversationhandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// List 获取对话列表
// @Summary      获取对话列表
// @Description  获取指定应用的会话列表，支持分页
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

	req := conversationreq.ListReq{}
	// TODO:
	req.AgentAPPKey = c.Param("app_key")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Errorf("GetPublishAgentList error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		o11y.Error(c, fmt.Sprintf("GetPublishAgentList error cause: %v, err trace: %+v\n", errors.Cause(err), err))
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, httpErr)

		return
	}

	req.Page = page

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		h.logger.Errorf("GetPublishAgentList error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		o11y.Error(c, fmt.Sprintf("GetPublishAgentList error cause: %v, err trace: %+v\n", errors.Cause(err), err))
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		rest.ReplyError(c, httpErr)

		return
	}

	req.Size = size
	user := chelper.GetVisitorFromCtx(ctx)
	req.UserId = user.ID

	list, total, err := h.conversationSvc.List(ctx, req)
	if err != nil {
		h.logger.Errorf("list conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err)
		o11y.Error(c, fmt.Sprintf("list conversation failed cause: %v, err trace: %+v\n", errors.Cause(err), err))
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
