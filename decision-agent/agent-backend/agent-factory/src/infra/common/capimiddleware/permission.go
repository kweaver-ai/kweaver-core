package capimiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/squaresvc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// CheckPms 检查 Agent 使用权限中间件
// 使用示例：见`CheckAgentUsePmsDemo`
func CheckPms(req *CheckPmsReq, clb func(c *gin.Context, hasPms bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := req.ReqCheck(); err != nil {
			err = errors.Wrap(err, "capimiddleware: [CheckPms]: req check error")
			httpErr := capierr.New400Err(c, err.Error())
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		if !req.IsAgentUseCheck() {
			panic("capimiddleware: [CheckPms]: 目前只支持检查Agent使用权限")
		}

		if clb == nil {
			panic("capimiddleware: [CheckPms]: clb is nil")
		}

		// NOTE: 通过AgentKey查询AgentID
		agentID, err := squaresvc.NewSquareService().CheckAndGetID(c.Request.Context(), req.ResourceID)
		if err != nil {
			httpErr := capierr.New400Err(c, "[capimiddleware][CheckPms] CheckAndGetID failed")
			rest.ReplyError(c, httpErr)
			c.Abort()

			return
		}

		_req := &cpmsreq.CheckAgentRunReq{
			AgentID:      agentID,
			UserID:       req.UserID,
			AppAccountID: req.AppAccountID,
		}

		resp, err := dainject.NewPermissionSvc().CheckUsePermission(c, _req)
		if err != nil {
			err = errors.Wrap(err, "middleware: [CheckPms]: CheckUsePermission error")
			rest.ReplyError(c, err)
			c.Abort()

			return
		}

		hasPms := resp != nil && resp.IsAllowed
		clb(c, hasPms)
	}
}
