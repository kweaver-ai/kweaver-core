package v3agentconfighandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// @Summary      获取SELF_CONFIG字段结构
// @Description  - 获取SELF_CONFIG字段结构定义 - SELF_CONFIG字段对应agent详情接口的config字段（减去了少量字段） - 用于前端动态展示SELF_CONFIG字段树（供用户选择字段）
// @Tags         agent
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "成功获取字段结构"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/agent-self-config-fields [get]
func (h *daConfHTTPHandler) SelfConfig(c *gin.Context) {
	// 1. 创建自配置字段对象
	sf := agentconfigresp.NewSelfConfigField()

	// 2. 从内嵌JSON加载配置字段
	err := sf.LoadFromJSONStr()
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 3. 返回结果
	c.JSON(http.StatusOK, sf)
}
