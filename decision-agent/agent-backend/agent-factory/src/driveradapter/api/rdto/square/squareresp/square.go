package squareresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
)

type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}
type ListAgentResp []AgentListItemResp

type AgentListItemResp struct {
	CategoryId   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	daconfeo.DataAgent

	PublishTime     int64    `json:"publish_time"`
	PublishUserId   string   `json:"-"`
	PublishUserInfo UserInfo `json:"publish_user_info"`
	UpdateUserInfo  UserInfo `json:"update_user_info"`
}
