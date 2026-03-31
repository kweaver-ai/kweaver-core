package personalspaceresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// AgentTplListItem 个人空间Agent模板列表项
type AgentTplListItem struct {
	ID        int64  `json:"id"`          // 模板ID
	Key       string `json:"key"`         // 模板标识
	IsBuiltIn int    `json:"is_built_in"` // 是否内置 (0: 否, 1: 是)

	Name    string `json:"name"`    // 模板名称
	Profile string `json:"profile"` // 模板简介

	Status              cdaenum.Status             `json:"status"`                 // 发布状态
	AgentTplCreatedType daenum.AgentTplCreatedType `json:"agent_tpl_created_type"` // 模板创建类型

	CreatedBy     string `json:"created_by"`      // 创建人
	CreatedByName string `json:"created_by_name"` // 创建人名称
	CreatedAt     int64  `json:"created_at"`      // 创建时间

	UpdatedBy     string `json:"updated_by"`      // 更新人
	UpdatedByName string `json:"updated_by_name"` // 更新人名称
	UpdatedAt     int64  `json:"updated_at"`      // 更新时间

	PublishedAt     int64  `json:"published_at"`      // 发布时间
	PublishedBy     string `json:"published_by"`      // 发布人
	PublishedByName string `json:"published_by_name"` // 发布人名称
}

// AgentTplListResp 个人空间Agent模板列表响应
type AgentTplListResp struct {
	Entries []*AgentTplListItem `json:"entries"` // 模板列表
	// Total   int64               `json:"total"`   // 总数
	PaginationMarkerStr string `form:"pagination_marker_str" json:"pagination_marker_str"` // 分页标记

	Marker *PTplListPaginationMarker `json:"-"`

	IsLastPage bool `json:"is_last_page"`
}

func NewAgentTplListResp() *AgentTplListResp {
	return &AgentTplListResp{
		Entries: []*AgentTplListItem{},
	}
}

func (l *AgentTplListResp) LoadFromEos(eos []*daconfeo.DataAgentTplListEo) (err error) {
	for _, eo := range eos {
		item := &AgentTplListItem{}

		err = cutil.CopyStructUseJSON(item, eo)
		if err != nil {
			return
		}

		l.Entries = append(l.Entries, item)
	}

	l.PaginationMarkerStr, err = l.genMarkerStr()
	if err != nil {
		return
	}

	return
}

func (l *AgentTplListResp) genMarkerStr() (markerStr string, err error) {
	marker := NewPTplListPaginationMarker()

	if len(l.Entries) == 0 || l.IsLastPage {
		return
	}

	// 1. 取最后一个
	lastItem := l.Entries[len(l.Entries)-1]

	// 2. 设置 marker
	marker.UpdatedAt = lastItem.UpdatedAt
	marker.LastTplID = lastItem.ID

	// 3. 转换为字符串
	markerStr, err = marker.ToString()
	if err != nil {
		return
	}

	return
}
