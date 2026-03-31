package pubedresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// 已发布智能体列表项响应
type PublishedAgentInfoListItem struct {
	ID            string `json:"id"`              // 模板ID
	Version       string `json:"version"`         // 模板版本
	Key           string `json:"key"`             // 模板标识
	IsBuiltIn     int    `json:"is_built_in"`     // 是否内置 (0: 否, 1: 是)
	IsSystemAgent int    `json:"is_system_agent"` // 是否系统智能体 (0: 否, 1: 是)

	Name    string `json:"name"`    // 模板名称
	Profile string `json:"profile"` // 模板简介

	AvatarType cdaenum.AvatarType `json:"avatar_type"`
	Avatar     string             `json:"avatar"`

	Config *daconfvalobj.Config `json:"config"`

	// CreatedBy     string `json:"created_by"`      // 创建人
	//CreatedByName string `json:"created_by_name"` // 创建人名称
	//CreatedAt     int64  `json:"created_at"`      // 创建时间
	//
	//UpdatedBy     string `json:"updated_by"`      // 更新人
	//UpdatedByName string `json:"updated_by_name"` // 更新人名称
	//UpdatedAt     int64  `json:"updated_at"`      // 更新时间

	PublishedAt     int64  `json:"published_at"`      // 发布时间
	PublishedBy     string `json:"published_by"`      // 发布人
	PublishedByName string `json:"published_by_name"` // 发布人名称

	PublishInfo *publishvo.ListPublishInfo `json:"publish_info"` // 发布信息
}

func NewPublishedAgentInfoListItem() *PublishedAgentInfoListItem {
	return &PublishedAgentInfoListItem{
		PublishInfo: publishvo.NewListPublishInfo(),
	}
}

// HlConfig 根据请求参数配置返回的配置字段
// 只暴露需要的指定字段，默认只暴露input字段
func (i *PublishedAgentInfoListItem) HlConfig(needConfigFields []string) {
	newConfig := daconfvalobj.NewConfig()

	for _, field := range needConfigFields {
		if field == "input" {
			newConfig.Input = i.Config.Input
		}
	}

	i.Config = newConfig
}

// 已发布智能体列表响应
type PAInfoListResp struct {
	Entries []*PublishedAgentInfoListItem `json:"entries"` // 智能体列表
}

func NewPublishedAgentInfoListResp() *PAInfoListResp {
	return &PAInfoListResp{
		Entries: []*PublishedAgentInfoListItem{},
	}
}

func (l *PAInfoListResp) LoadFromEos(eos []*pubedeo.PublishedAgentEo, needConfigFields []string) (err error) {
	for _, eo := range eos {
		item := NewPublishedAgentInfoListItem()

		err = cutil.CopyStructUseJSON(item, eo)
		if err != nil {
			return
		}

		item.Version = eo.Version

		err = cutil.CopyStructUseJSON(item.PublishInfo, eo.PublishedToBeStruct)
		if err != nil {
			return
		}

		item.HlConfig(needConfigFields)
		l.Entries = append(l.Entries, item)
	}

	return
}
