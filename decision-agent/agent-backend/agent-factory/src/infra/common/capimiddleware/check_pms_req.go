package capimiddleware

import (
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

// CheckPmsReq 权限检查请求（从 afhttpdto.CheckPmsReq 迁移而来）
type CheckPmsReq struct {
	ResourceType cdaenum.ResourceType
	ResourceID   string

	Operator cdapmsenum.Operator

	UserID       string `json:"user_id"`
	AppAccountID string `json:"app_account_id"`
}

func (r *CheckPmsReq) IsAgentUseCheck() bool {
	if r.ResourceType == cdaenum.ResourceTypeDataAgent && r.Operator == cdapmsenum.AgentUse {
		return true
	}

	return false
}

func (r *CheckPmsReq) ReqCheck() (err error) {
	if err = r.ResourceType.EnumCheck(); err != nil {
		return
	}

	if err = r.Operator.EnumCheck(); err != nil {
		return
	}

	if r.UserID == "" && r.AppAccountID == "" {
		err = errors.New("[CheckPmsReq][ReqCheck]: req.UserID and req.AppAccountID are both empty")
		return
	}

	if r.ResourceID == "" {
		err = errors.New("[CheckPmsReq][ReqCheck]: req.ResourceID is empty")
		return
	}

	return
}

// NewCheckAgentUsePmsReq 创建检查 agent 使用权限的请求
func NewCheckAgentUsePmsReq(agentID string, userID string, appAccountID string) *CheckPmsReq {
	return &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		ResourceID:   agentID,
		Operator:     cdapmsenum.AgentUse,
		UserID:       userID,
		AppAccountID: appAccountID,
	}
}
