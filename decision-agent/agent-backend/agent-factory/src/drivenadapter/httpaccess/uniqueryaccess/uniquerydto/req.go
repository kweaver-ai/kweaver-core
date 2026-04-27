package uniquerydto

// ReqDataView 基于数据视图的视图数据查询请求体
type ReqDataView struct {
	// 内部接口鉴权使用
	XAccountID   string `json:"-"`
	XAccountType string `json:"-"`

	Start          int64         `json:"start,omitempty"`            // 查询的开始时间，单位到毫秒
	End            int64         `json:"end,omitempty"`              // 查询的结束时间，单位到毫秒
	Filters        *GlobalFilter `json:"filters,omitempty"`          // 全局过滤条件
	Sort           []Sort        `json:"sort,omitempty"`             // 排序字段
	Offset         int           `json:"offset,omitempty"`           // 查询的起始位置，默认值 0
	Limit          int           `json:"limit,omitempty"`            // 返回的数量，默认值 10
	Format         string        `json:"format,omitempty"`           // 数据输出格式，original 或 flat
	SearchAfter    []interface{} `json:"search_after,omitempty"`     // 上次查询返回的最后一个文档的排序值
	PitID          string        `json:"pit_id,omitempty"`           // point_in_time 的 id
	PitKeepAlive   string        `json:"pit_keep_alive,omitempty"`   // point_in_time 的存活时间
	NeedTotal      bool          `json:"need_total,omitempty"`       // 是否需要总数，默认false
	UseSearchAfter bool          `json:"use_search_after,omitempty"` // 是否使用search after 翻页，默认false
	DateField      string        `json:"date_field,omitempty"`       // 时间字段，和start,end配合使用
	SQL            string        `json:"sql,omitempty"`              // 自定义sql查询
}

// GlobalFilter 全局过滤条件
type GlobalFilter struct {
	Operation     string      `json:"operation"`                // 操作符
	SubConditions []Condition `json:"sub_conditions,omitempty"` // 子过滤条件
}

// Condition 过滤条件
type Condition struct {
	Value         interface{} `json:"value,omitempty"`          // 字段值
	Operation     string      `json:"operation"`                // 操作符
	Field         string      `json:"field,omitempty"`          // 字段名称
	ValueFrom     string      `json:"value_from,omitempty"`     // 字段值来源
	SubConditions []Condition `json:"sub_conditions,omitempty"` // 子过滤条件
	DisplayName   string      `json:"display_name,omitempty"`   // 显示名
}

// Sort 排序字段和方向
type Sort struct {
	Field     string `json:"field"`     // 排序字段
	Direction string `json:"direction"` // 排序方向
}
