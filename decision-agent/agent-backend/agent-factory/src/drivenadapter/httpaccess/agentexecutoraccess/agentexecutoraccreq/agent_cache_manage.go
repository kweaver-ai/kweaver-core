package agentexecutoraccreq

type AgentCacheManageReq struct {
	AgentID      string               `json:"agent_id"`
	AgentVersion string               `json:"agent_version"`
	Action       AgentCacheActionType `json:"action"`
}
