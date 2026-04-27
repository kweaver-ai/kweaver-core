package v3agentconfighandler

import "github.com/gin-gonic/gin"

// CreateReact 创建 react agent
// @Summary      创建react agent
// @Description  创建一个新的 React 模式 Agent 配置。请求体与普通创建接口一致，但 config.mode 必须为 react。
// @Tags         agent,agent-internal
// @Accept       json
// @Produce      json
// @Param        request  body      swagger.AgentConfigCreateReq  true  "React Agent 配置"
// @Success      201      {object}  swagger.AgentConfigCreateRes  "成功"
// @Failure      400      {object}  swagger.APIError   "请求参数错误"
// @Failure      401      {object}  swagger.APIError   "未授权"
// @Failure      403      {object}  swagger.APIError   "禁止访问"
// @Failure      500      {object}  swagger.APIError   "服务器内部错误"
// @Router       /v3/agent/react [post]
// @Security     BearerAuth
func (h *daConfHTTPHandler) CreateReact(c *gin.Context) {
	h.create(c, validateCreateReactReq)
}
