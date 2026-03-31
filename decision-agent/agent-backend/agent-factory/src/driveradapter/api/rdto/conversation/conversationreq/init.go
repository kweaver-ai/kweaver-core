package conversationreq

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"

// UpdateReq 表示更新agent的请求
type InitReq struct {
	AgentAPPKey       string            `json:"-"`                // agent app key
	Title             string            `json:"title"`            // conversation 标题
	UserID            string            `json:"-"`                // 用户ID
	TempareaId        string            `json:"temparea_id"`      // 临时区域ID
	VisitorType       string            `json:"visitor_type"`     // 访客类型  user / app
	AgentID           string            `json:"agent_id"`         // agent id
	AgentVersion      string            `json:"agent_version"`    // agent version
	ExecutorVersion   string            `json:"executor_version"` // executor version v1 或 v2 默认v2
	XAccountID        string            `json:"-"`
	XAccountType      cenum.AccountType `json:"-"`
	XBusinessDomainID string            `json:"-"`
}

func (p *InitReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Title.required": `"title"不能为空`,
	}
}

func (p *InitReq) ReqCheck() (err error) {
	return nil
}
