package permissionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// GetUserStatus 获取用户拥有的管理权限状态
// @Summary      获取用户拥有的管理权限状态
// @Description  - 获取用户拥有的管理权限状态 - 这里暂时不针对某个具体的资源实例，而是针对资源类型 - 管理权限和具体的资源实例无关
// @Tags         权限,权限-internal
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-permission/management/user-status [get]
func (h *permissionHandler) GetUserStatus(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)

	// 调用service层获取用户权限状态
	resp, err := h.permissionSvc.GetUserStatus(ctx)
	if err != nil {
		h.logger.Errorf("GetUserStatus error cause: %v, err trace: %+v\n", errors.Cause(err), err)
		_ = c.Error(err)

		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, resp)
}
