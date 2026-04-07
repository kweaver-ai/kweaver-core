package squareresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
)

type RecentListAgentResp []RecentAgentListItem

type RecentAgentListItem struct {
	CategoryId   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	daconfeo.DataAgent

	PublishedAt     int64  `json:"published_at"`
	PublishedBy     string `json:"published_by"`
	PublishedByName string `json:"published_by_name"`

	// UpdateUserInfo UserInfo `json:"update_user_info"`

	PublishInfo *publishvo.ListPublishInfo `json:"publish_info"` // 发布信息
}
