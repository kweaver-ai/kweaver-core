package uniquerydto

// ViewResults 基于数据视图的查询结果
type ViewResults struct {
	View        *DataView     `json:"view,omitempty"`         // 数据视图对象信息
	Entries     []interface{} `json:"entries"`                // 视图数据
	TotalCount  int           `json:"total_count,omitempty"`  // 不分页情况下的数据总数
	PitID       string        `json:"pit_id,omitempty"`       // point_in_time 的 id
	SearchAfter []interface{} `json:"search_after,omitempty"` // 开启分页时返回，表示返回的最后一个文档的排序值
}

// DataView 数据视图对象信息
type DataView struct {
	ID          string `json:"id,omitempty"`          // 视图ID
	Name        string `json:"name,omitempty"`        // 视图名称
	Description string `json:"description,omitempty"` // 视图描述
	// 可以根据需要添加更多字段
}
