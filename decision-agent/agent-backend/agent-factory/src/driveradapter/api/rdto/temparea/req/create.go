package tempareareq

type CreateReq struct {
	AgentID      string     `json:"agent_id" binding:"required"`
	AgentVersion string     `json:"agent_version" binding:"required"`
	Source       []TempArea `json:"source" binding:"required"`
	TempAreaID   string     `json:"-"`
	UserID       string     `json:"-"`
}

type TempArea struct {
	ID      string      `json:"id" binding:"required"`
	Type    string      `json:"type"` // doc
	Details interface{} `json:"details"`
}
