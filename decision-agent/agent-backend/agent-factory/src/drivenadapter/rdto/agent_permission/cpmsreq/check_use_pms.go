package cpmsreq

// CheckAgentRunReq 检查运行权限请求
type CheckAgentRunReq struct {
	AgentID       string `form:"agent_id" json:"agent_id" binding:"required"` // agent id
	CustomSpaceID string `form:"custom_space_id" json:"custom_space_id"`      // 自定义空间id（广场custom_space_id为空）

	UserID string `form:"user_id" json:"user_id"` // 用户id

	AppAccountID string `form:"app_account_id" json:"app_account_id"` // app_account_id
}

func (req *CheckAgentRunReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"AgentID.required": `"agent_id"不能为空`,
	}
}
