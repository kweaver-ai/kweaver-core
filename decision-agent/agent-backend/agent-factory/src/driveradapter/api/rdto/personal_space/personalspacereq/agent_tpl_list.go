package personalspacereq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/pkg/errors"
)

// AgentTplListReq 个人空间Agent模板列表请求
type AgentTplListReq struct {
	// common.PageSize
	Name string `form:"name" json:"name"` // 模板名称（模糊搜索）

	ProductKey string `form:"product_key" json:"product_key"` // 产品标识

	CategoryID string `form:"category_id" json:"category_id"` // 分类ID

	PublishStatus       cdaenum.Status             `form:"publish_status" json:"publish_status"`                 // 发布状态 ("unpublished", "published")
	AgentTplCreatedType daenum.AgentTplCreatedType `form:"agent_tpl_created_type" json:"agent_tpl_created_type"` // 模板创建类型 ("copy_from_agent", "copy_from_tpl")

	// CreatedBy string `form:"-" json:"-"` // 创建人（用于个人空间查询）

	Size int `form:"size,default=10" json:"size" binding:"numeric,max=1000"` // 每页显示数量

	PaginationMarkerStr string `form:"pagination_marker_str" json:"pagination_marker_str"` // 分页标记

	Marker *personalspaceresp.PTplListPaginationMarker `json:"-"` // 分页标记
}

func (req *AgentTplListReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}

// CustomCheck 自定义参数校验
func (req *AgentTplListReq) CustomCheck() error {
	// 校验发布状态
	if req.PublishStatus != "" {
		if err := req.PublishStatus.EnumCheck(); err != nil {
			return errors.Wrap(err, "[AgentTplListReq]: publish_status is invalid")
		}
	}

	// 校验模板创建类型
	if req.AgentTplCreatedType != "" {
		if err := req.AgentTplCreatedType.EnumCheck(); err != nil {
			return errors.Wrap(err, "[AgentTplListReq]: agent_tpl_created_type is invalid")
		}
	}

	return nil
}

func (req *AgentTplListReq) LoadMarkerStr() (err error) {
	if req.PaginationMarkerStr == "" {
		return
	}

	req.Marker = personalspaceresp.NewPTplListPaginationMarker()

	err = req.Marker.LoadFromStr(req.PaginationMarkerStr)
	if err != nil {
		return
	}

	return
}
