package pubedreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
	"github.com/pkg/errors"
)

// 已发布智能体列表请求对象
// 【注意】： 这里发生变化时，需要同步调整IsNoConditionAndFirstPage()
type PubedAgentListReq struct {
	Name             string              `form:"name" json:"name"`                       // 根据名称模糊查询
	IDs              []string            `json:"ids"`                                    // 根据ID查询
	AgentKeys        []string            `json:"agent_keys"`                             // 根据智能体标识查询
	ExcludeAgentKeys []string            `json:"exclude_agent_keys"`                     // 排除智能体标识
	CategoryID       string              `form:"category_id" json:"category_id"`         // 分类ID
	ToBeFlag         cdaenum.PublishToBe `form:"publish_to_be" json:"publish_to_be"`     // 发布为标识("api_agent", "web_sdk_agent", "skill_agent")
	CustomSpaceID    string              `form:"custom_space_id" json:"custom_space_id"` // 自定义空间ID

	IsToCustomSpace int `form:"is_to_custom_space" json:"is_to_custom_space"` // 获取发布到自定义空间的智能体

	IsToSquare int `form:"is_to_square" json:"is_to_square"` // 获取发布到广场的智能体

	BusinessDomainIDs []string `json:"business_domain_ids"` // 业务域ID数组；当未禁用业务域时，如果不传会使用 header 中的 "x-business-domain"，若该 header 也未传则默认使用“公共业务域”过滤

	Size int `form:"size,default=10" json:"size" binding:"numeric,max=1000"` // 每页显示数量

	PaginationMarkerStr string `form:"pagination_marker_str" json:"pagination_marker_str"` // 上一次查询的最后一条记录对应的pagination_marker_str

	Marker *pubedresp.PAListPaginationMarker `json:"-"` // 上一次查询的最后一条记录的marker
}

func (req *PubedAgentListReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}

func (req *PubedAgentListReq) CustomCheck() (err error) {
	if len(req.IDs) > 1000 {
		return errors.New("[PubedAgentListReq]: ids length is too long")
	}

	if len(req.AgentKeys) > 1000 {
		return errors.New("[PubedAgentListReq]: agent_keys length is too long")
	}

	if len(req.ExcludeAgentKeys) > 1000 {
		return errors.New("[PubedAgentListReq]: exclude_agent_keys length is too long")
	}

	if len(req.BusinessDomainIDs) > 2 {
		return errors.New("[PubedAgentListReq]: business_domain_ids length is too long (max 2)")
	}

	return
}

func (req *PubedAgentListReq) LoadMarkerStr() (err error) {
	if req.PaginationMarkerStr == "" {
		return
	}

	req.Marker = pubedresp.NewPAListPaginationMarker()

	err = req.Marker.LoadFromStr(req.PaginationMarkerStr)
	if err != nil {
		return
	}

	return
}
