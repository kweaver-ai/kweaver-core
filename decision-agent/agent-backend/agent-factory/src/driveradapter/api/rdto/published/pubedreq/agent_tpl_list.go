package pubedreq

import "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"

// 已发布模板列表请求对象
type PubedTplListReq struct {
	Name       string `form:"name" json:"name"`               // 根据名称模糊查询
	CategoryID string `form:"category_id" json:"category_id"` // 分类ID
	// common.PageSize

	Size int `form:"size,default=10" json:"size" binding:"numeric,max=1000"` // 每页显示数量

	PaginationMarkerStr string `form:"pagination_marker_str" json:"pagination_marker_str"` // 分页标记

	Marker *pubedresp.PTplListPaginationMarker `json:"-"` // 分页标记

	TplIDsByBd []string `json:"-"` // 根据业务域ID获取到的模板ID列表
}

func (req *PubedTplListReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}

// LoadMarkerStr
func (req *PubedTplListReq) LoadMarkerStr() (err error) {
	if req.PaginationMarkerStr == "" {
		return
	}

	req.Marker = pubedresp.NewPTplListPaginationMarker()

	err = req.Marker.LoadFromStr(req.PaginationMarkerStr)
	if err != nil {
		return
	}

	return
}
