package personalspacereq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/pkg/errors"
)

// AgentListReq 个人空间Agent列表请求
type AgentListReq struct {
	// common.PageSize
	Name string `form:"name" json:"name"` // Agent名称（模糊搜索）

	PublishStatus cdaenum.StatusThreeState `form:"publish_status" json:"publish_status"` // 发布状态 ("unpublished", "published", "published_edited")
	PublishToBe   cdaenum.PublishToBe      `form:"publish_to_be" json:"publish_to_be"`   // 发布为标识("api_agent", "web_sdk_agent", "skill_agent")

	AgentCreatedType daenum.AgentCreatedType `form:"agent_created_type" json:"agent_created_type"` // Agent创建类型 ("create", "copy")

	Size int `form:"size,default=10" json:"size" binding:"numeric,max=1000"` // 每页显示数量

	PaginationMarkerStr string `form:"pagination_marker_str" json:"pagination_marker_str"` // 分页标记

	Marker *personalspaceresp.PAListPaginationMarker `json:"-"`
}

func (req *AgentListReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}

// CustomCheck 自定义参数校验
func (req *AgentListReq) CustomCheck() error {
	// 校验发布状态
	if req.PublishStatus != "" {
		if err := req.PublishStatus.EnumCheck(); err != nil {
			return errors.Wrap(err, "[AgentListReq]: publish_status is invalid")
		}
	}

	// 校验Agent创建类型
	if req.AgentCreatedType != "" {
		if err := req.AgentCreatedType.EnumCheck(); err != nil {
			return errors.Wrap(err, "[AgentListReq]: agent_created_type is invalid")
		}
	}

	// 校验发布为标识
	if req.PublishToBe != "" {
		if err := req.PublishToBe.EnumCheck(); err != nil {
			return errors.Wrap(err, "[AgentListReq]: publish_to_be is invalid")
		}
	}

	return nil
}

func (req *AgentListReq) LoadMarkerStr() (err error) {
	if req.PaginationMarkerStr == "" {
		return
	}

	req.Marker = personalspaceresp.NewPAListPaginationMarker()

	err = req.Marker.LoadFromStr(req.PaginationMarkerStr)
	if err != nil {
		return
	}

	return
}
