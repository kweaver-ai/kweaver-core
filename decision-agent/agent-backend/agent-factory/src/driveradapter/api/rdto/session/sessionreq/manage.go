package sessionreq

type SessionManageActionType string

const (
	SessionManageActionGetInfoOrCreate         SessionManageActionType = "get_info_or_create"
	SessionManageActionRecoverLifetimeOrCreate SessionManageActionType = "recover_lifetime_or_create"
)

type ManageReq struct {
	ConversationID string                  `json:"-"`
	Action         SessionManageActionType `json:"action" binding:"required"`
	AgentID        string                  `json:"agent_id" binding:"required"`
	AgentVersion   string                  `json:"agent_version" binding:"required"`
}

func (p *ManageReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Action.required":       `"action"不能为空`,
		"AgentID.required":      `"agent_id"不能为空`,
		"AgentVersion.required": `"agent_version"不能为空`,
	}
}

func (p *ManageReq) ReqCheck() (err error) {
	return nil
}
