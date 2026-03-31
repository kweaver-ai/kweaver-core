package valueobject

type AgentInfo struct {
	AgentID      string `json:"agent_id"`
	AgentName    string `json:"agent_name"`
	AgentStatus  string `json:"agent_status"`
	AgentVersion string `json:"agent_version"`
}
