package pubedresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// 已发布模板列表项响应
type PubedTplListItemResp struct {
	ID int64 `json:"id"` // 发布ID（对应t_data_agent_config_tpl_published表的id）

	TplID     int64  `json:"tpl_id"`
	Key       string `json:"key"`         // 模板标识
	IsBuiltIn int    `json:"is_built_in"` // 是否内置 (0: 否, 1: 是)

	Name    string `json:"name"`    // 模板名称
	Profile string `json:"profile"` // 模板简介

	PublishedAt     int64  `json:"published_at"`      // 发布时间
	PublishedBy     string `json:"published_by"`      // 发布人
	PublishedByName string `json:"published_by_name"` // 发布人名称
}

// 已发布模板列表响应
type PublishedAgentTplListResp struct {
	Entries []*PubedTplListItemResp `json:"entries"` // 模板列表
	// Total   int64                            `json:"total"`   // 模板总数

	PaginationMarkerStr string `json:"pagination_marker_str"`
	IsLastPage          bool   `json:"is_last_page"`
}

func NewPublishedAgentTplListResp() *PublishedAgentTplListResp {
	return &PublishedAgentTplListResp{
		// Total:   total,
		Entries:    []*PubedTplListItemResp{},
		IsLastPage: true,
	}
}

func (l *PublishedAgentTplListResp) LoadFromEos(eos []*pubedeo.PublishedTplListEo) (err error) {
	for _, eo := range eos {
		item := &PubedTplListItemResp{}

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

func (l *PublishedAgentTplListResp) genMarkerStr() (markerStr string, err error) {
	marker := NewPTplListPaginationMarker()

	if len(l.Entries) == 0 || l.IsLastPage {
		return
	}

	// 1. 取最后一个
	lastItem := l.Entries[len(l.Entries)-1]

	// 2. 设置 marker
	marker.LastPubedTplID = lastItem.ID

	// 3. 转换为字符串
	markerStr, err = marker.ToString()
	if err != nil {
		return
	}

	return
}
