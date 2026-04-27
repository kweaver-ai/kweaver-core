package permissionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// CheckUsePermission 检查非个人空间下的某个agent是否有运行权限
// @Summary      检查某个agent是否有执行（使用）权限
// @Description  - 检查某个agent是否有执行（使用）权限 - 此接口有`内部接口`和`外部接口`
// @Tags         权限,权限-internal
// @Accept       json
// @Produce      json
// @Param        request  body      object  false  "请求体"
// @Success      200  {object}  object  "请求成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-permission/execute [post]
func (h *permissionHandler) CheckUsePermission(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	// 1. 获取请求参数
	var req cpmsreq.CheckAgentRunReq
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Errorf("CheckUsePermission bind json error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		httpErr := capierr.New400Err(c, chelper.ErrMsg(err, &req))
		_ = c.Error(httpErr)

		return
	}

	// 2. 调用service层检查权限
	resp, err := h.permissionSvc.CheckUsePermission(ctx, &req)
	if err != nil {
		h.logger.Errorf("CheckUsePermission error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		_ = c.Error(err)

		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, resp)
}
