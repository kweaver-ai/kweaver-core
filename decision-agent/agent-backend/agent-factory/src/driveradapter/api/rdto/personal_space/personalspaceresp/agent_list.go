package personalspaceresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/cdaconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// AgentListItem 个人空间Agent列表项
type AgentListItem struct {
	ID            string `json:"id"`              // Agent ID
	Key           string `json:"key"`             // Agent标识
	IsBuiltIn     int    `json:"is_built_in"`     // 是否内置 (0: 否, 1: 是)
	IsSystemAgent int    `json:"is_system_agent"` // 是否系统Agent (0: 否, 1: 是)

	Name    string `json:"name"`    // Agent名称
	Profile string `json:"profile"` // Agent简介

	// CategoryID string `json:"category_id"` // 分类ID
	Version string `json:"version"` // Agent版本

	AvatarType int    `json:"avatar_type"` // 头像类型
	Avatar     string `json:"avatar"`      // 头像信息
	ProductKey string `json:"product_key"` // 所属产品标识

	Status      cdaenum.StatusThreeState `json:"status"`       // 状态 ("unpublished", "published", "published_edited")
	CreatedType daenum.AgentCreatedType  `json:"created_type"` // Agent创建类型

	UpdatedAt     int64  `json:"updated_at"`      // 最近更新时间
	UpdatedBy     string `json:"updated_by"`      // 最近更新人
	UpdatedByName string `json:"updated_by_name"` // 最近更新人名称

	CreatedBy     string `json:"created_by"`      // 创建人
	CreatedByName string `json:"created_by_name"` // 创建人名称
	CreatedAt     int64  `json:"created_at"`      // 创建时间

	PublishedAt int64 `json:"published_at"` // 发布时间
	// PublishedBy     string `json:"published_by"`      // 发布人
	// PublishedByName string `json:"published_by_name"` // 发布人名称

	PublishInfo *publishvo.ListPublishInfo `json:"publish_info"` // 发布信息
}

func NewAgentListItem() *AgentListItem {
	return &AgentListItem{
		PublishInfo: publishvo.NewListPublishInfo(),
	}
}

// AgentListResp 个人空间Agent列表响应
type AgentListResp struct {
	Entries []*AgentListItem `json:"entries"` // Agent列表
	// Total   int64            `json:"total"`   // 总数

	PaginationMarkerStr string `form:"pagination_marker_str" json:"pagination_marker_str"` // 分页标记

	Marker *PAListPaginationMarker `json:"-"`

	IsLastPage bool `json:"is_last_page"`
}

// NewAgentListResp 创建个人空间Agent列表响应
func NewAgentListResp() *AgentListResp {
	return &AgentListResp{
		Entries: make([]*AgentListItem, 0),
	}
}

// LoadFromEos 从实体对象列表加载数据
func (resp *AgentListResp) LoadFromEos(eos []*daconfeo.DataAgent, releaseAgentPoMap map[string]*dapo.PublishedJoinPo) (err error) {
	if len(eos) == 0 {
		return nil
	}

	resp.Entries = make([]*AgentListItem, 0, len(eos))

	for _, eo := range eos {
		item := NewAgentListItem()

		err = cutil.CopyStructUseJSON(item, eo)
		if err != nil {
			return
		}

		if item.Version == "" {
			item.Version = cdaconstant.AgentVersionUnpublished
		}

		// 设置发布相关信息
		if releaseAgentPoMap != nil {
			if releaseAgentPoMap[eo.ID] != nil {
				releaseAgentPo := releaseAgentPoMap[eo.ID]

				err = cutil.CopyStructUseJSON(item.PublishInfo, releaseAgentPo.PublishedToBeStruct)
				if err != nil {
					return
				}

				item.PublishedAt = releaseAgentPo.PublishedAt

				item.Version = releaseAgentPo.Version

				// 判断是否是 "发布后有修改" 状态
				// 当 status 是 unpublished 且有发布记录（updated_at > published_at）时，设置为 published_edited
				if item.Status == cdaenum.StatusThreeStateUnpublished && item.UpdatedAt > releaseAgentPo.PublishedAt {
					item.Status = cdaenum.StatusThreeStatePublishedEdited
				}
			}
		}

		resp.Entries = append(resp.Entries, item)
	}

	resp.PaginationMarkerStr, err = resp.genMarkerStr()
	if err != nil {
		return
	}

	return nil
}

func (resp *AgentListResp) genMarkerStr() (markerStr string, err error) {
	marker := NewPAListPaginationMarker()

	if len(resp.Entries) == 0 || resp.IsLastPage {
		return
	}

	// 1. 取最后一个
	lastItem := resp.Entries[len(resp.Entries)-1]

	// 2. 设置 marker
	marker.UpdatedAt = lastItem.UpdatedAt
	marker.LastAgentID = lastItem.ID

	// 3. 转换为字符串
	markerStr, err = marker.ToString()
	if err != nil {
		return
	}

	return
}
