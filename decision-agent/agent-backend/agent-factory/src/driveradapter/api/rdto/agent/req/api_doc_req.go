package agentreq

type GetAPIDocReq struct {
	AppKey       string `json:"app_key"`
	AgentID      string `json:"agent_id"`
	AgentVersion string `json:"agent_version"`
}
