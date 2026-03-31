package agenthandler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// @Summary      获取Agent Api文档
// @Description  获取Agent Api文档
// @Tags         对话
// @Accept       json
// @Produce      json
// @Param        app_key  path      string  true  "app_key"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功"
// @Security     BearerAuth
// @Router       /v1/app/{app_key}/api/doc [post]
func (h *agentHTTPHandler) GetAPIDoc(c *gin.Context) {
	var req agentreq.GetAPIDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		rest.ReplyError(c, err)
		return
	}

	appKey := c.Param("app_key")
	if appKey == "" {
		rest.ReplyError(c, capierr.New400Err(c, "app_key is required"))
		h.logger.Errorf("[GetAPIDoc] app_key is required")
		o11y.Error(c, "[GetAPIDoc] app_key is required")

		return
	}

	ctx := context.WithValue(c.Request.Context(), constant.AppKey, appKey)

	doc, err := h.agentSvc.GetAPIDoc(ctx, &req)
	if err != nil {
		h.logger.Errorf("[GetAPIDoc] error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		o11y.Error(c, fmt.Sprintf("[GetAPIDoc] error cause: %v, err trace: %+v\n", errors.Cause(err), err))
		rest.ReplyError(c, err)

		return
	}

	rest.ReplyOK(c, http.StatusOK, doc)
}
