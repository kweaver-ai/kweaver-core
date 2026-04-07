package pubedresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// 已发布智能体列表项响应
type PAListItemResp struct {
	ID            string `json:"id"`              // agentID
	Version       string `json:"version"`         // agent版本
	Key           string `json:"key"`             // agent标识
	IsBuiltIn     int    `json:"is_built_in"`     // 是否内置 (0: 否, 1: 是)
	IsSystemAgent int    `json:"is_system_agent"` // 是否系统智能体 (0: 否, 1: 是)

	Name    string `json:"name"`    // 智能体名称
	Profile string `json:"profile"` // 智能体简介

	AvatarType cdaenum.AvatarType `json:"avatar_type"`
	Avatar     string             `json:"avatar"`

	// CreatedBy     string `json:"created_by"`      // 创建人
	//CreatedByName string `json:"created_by_name"` // 创建人名称
	//CreatedAt     int64  `json:"created_at"`      // 创建时间
	//
	//UpdatedBy     string `json:"updated_by"`      // 更新人
	//UpdatedByName string `json:"updated_by_name"` // 更新人名称
	//UpdatedAt     int64  `json:"updated_at"`      // 更新时间

	ReleaseID string `json:"release_id"`

	PublishedAt     int64  `json:"published_at"`      // 发布时间
	PublishedBy     string `json:"published_by"`      // 发布人
	PublishedByName string `json:"published_by_name"` // 发布人名称

	PublishInfo *publishvo.ListPublishInfo `json:"publish_info"` // 发布信息

	BusinessDomainID string `json:"business_domain_id"` // 业务域ID
}

func NewPAListItemResp() *PAListItemResp {
	return &PAListItemResp{
		PublishInfo: publishvo.NewListPublishInfo(),
	}
}

// 已发布智能体列表响应
type PubedAgentListResp struct {
	Entries []*PAListItemResp `json:"entries"` // 智能体列表
	// Total   int64             `json:"total"`   // 智能体总数
	PaginationMarkerStr string `json:"pagination_marker_str"`
	IsLastPage          bool   `json:"is_last_page"`
}

func NewPAListResp() *PubedAgentListResp {
	return &PubedAgentListResp{
		// Total:   total,
		Entries: []*PAListItemResp{},
	}
}

func (l *PubedAgentListResp) LoadFromEos(eos []*pubedeo.PublishedAgentEo, agentID2BdIDMap map[string]string) (err error) {
	for _, eo := range eos {
		item := NewPAListItemResp()

		err = cutil.CopyStructUseJSON(item, eo)
		if err != nil {
			return
		}

		item.Version = eo.Version

		err = cutil.CopyStructUseJSON(item.PublishInfo, eo.PublishedToBeStruct)
		if err != nil {
			return
		}

		item.BusinessDomainID = agentID2BdIDMap[eo.ID]

		l.Entries = append(l.Entries, item)
	}

	l.PaginationMarkerStr, err = l.genMarkerStr()
	if err != nil {
		return
	}

	return
}

func (l *PubedAgentListResp) genMarkerStr() (markerStr string, err error) {
	marker := NewPAListPaginationMarker()

	if len(l.Entries) == 0 || l.IsLastPage {
		return
	}

	// 1. 取最后一个
	lastItem := l.Entries[len(l.Entries)-1]

	// 2. 设置 marker
	marker.PublishedAt = lastItem.PublishedAt
	marker.LastReleaseID = lastItem.ReleaseID

	// 3. 转换为字符串
	markerStr, err = marker.ToString()
	if err != nil {
		return
	}

	return
}
