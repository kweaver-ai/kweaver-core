package squarehandler

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"

	"github.com/gin-gonic/gin"
)

var agentInfoReqCtxKey = "agentInfoReqCtxKey"

// 1. 从path中获取 agent id 和 version
// 2. agent_id 可能为 id 或 key，当agent_id查询不到时会根据key来查询，然后获取其对应的id，赋值给agent_id
func (h *squareHandler) agentInfoGetReqMiddleware(c *gin.Context) {
	// 1. 获取 agent id 和 version
	agentID := c.Param("agent_id")
	version := c.Param("version")

	if agentID == "" || version == "" {
		err := capierr.New400Err(c, "agent_id和version不能为空")
		_ = c.Error(err)
		c.Abort()

		return
	}

	// 2. 检查 agent_id 是否存在，不存在则根据key来查询，然后获取其对应的id，赋值给agent_id
	agentID, err := h.squareSvc.CheckAndGetID(c, agentID)
	if err != nil {
		_ = c.Error(err)
		c.Abort()

		return
	}

	// 3. 获取用户ID
	userID := chelper.GetUserIDFromCtx(c)

	// 4. 构造请求参数
	agentInfoReq := squarereq.AgentInfoReq{
		AgentID:      agentID,
		AgentVersion: version,
		IsVisit:      cutil.StringToBool(c.Query("is_visit")),
		UserID:       userID,
	}

	// 5. 设置请求参数
	c.Set(agentInfoReqCtxKey, &agentInfoReq)

	// 6. 继续执行
	c.Next()
}

func (h *squareHandler) agentInfoAgentUsePmsCheck(c *gin.Context) {
	iReq, exists := c.Get(agentInfoReqCtxKey)
	if !exists {
		_ = c.Error(capierr.New400Err(c, "[agentInfoAgentUsePmsCheck]: agentInfoReqCtxKey不存在"))
		c.Abort()

		return
	}

	req, ok := iReq.(*squarereq.AgentInfoReq)
	if !ok {
		_ = c.Error(capierr.New400Err(c, "[agentInfoAgentUsePmsCheck]: agentInfoReqCtxKey类型错误"))
		c.Abort()

		return
	}

	// 检查 Agent 使用权限
	pmsReq := &cpmsreq.CheckAgentRunReq{
		AgentID:      req.AgentID,
		UserID:       req.UserID,
		AppAccountID: "",
	}

	resp, pmsErr := dainject.NewPermissionSvc().CheckUsePermission(c, pmsReq)
	if pmsErr != nil {
		_ = c.Error(pmsErr)
		c.Abort()

		return
	}

	if resp == nil || !resp.IsAllowed {
		_err := capierr.NewCustom403Err(c, apierr.AgentFactoryPermissionForbidden, "[agentInfoAgentUsePmsCheck]: 无当前Agent使用权限")
		_ = c.Error(_err)
		c.Abort()

		return
	}

	// 继续执行
	c.Next()
}
