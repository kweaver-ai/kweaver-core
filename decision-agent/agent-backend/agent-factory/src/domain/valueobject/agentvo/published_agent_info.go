package agentvo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// PublishUserInfo 发布者信息
type PublishUserInfo struct {
	UserID   string `json:"user_id"`  // 用户id
	Username string `json:"username"` // 用户显示名
}

// PublishedAgentInfo 已发布智能体信息
type PublishedAgentInfo struct {
	// PublishUserInfo *PublishUserInfo   `json:"publish_user_info"` // 发布者信息
	PublishedBy     string `json:"published_by"`      // 发布人
	PublishedByName string `json:"published_by_name"` // 发布人名称
	PublishedAt     int64  `json:"published_at"`      // 发布时间

	Profile    string             `json:"profile"`     // agent 简介
	Version    string             `json:"version"`     // agent 版本
	AvatarType cdaenum.AvatarType `json:"avatar_type"` // 头像类型: 1-内置头像, 2-用户上传头像, 3-AI生成头像
	Avatar     string             `json:"avatar"`      // 头像信息

	dapo.PublishedToBeStruct
}

func NewPublishedAgentInfo() *PublishedAgentInfo {
	return &PublishedAgentInfo{}
}

func (d *PublishedAgentInfo) LoadFromReleaseAgentPO(po *dapo.PublishedJoinPo) (err error) {
	err = cutil.CopyStructUseJSON(d, po)
	if err != nil {
		return
	}

	d.Version = po.Version

	return
}
