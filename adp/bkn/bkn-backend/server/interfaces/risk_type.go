// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	// 内置风险评估函数信息
	BuiltinToolBoxID         = "bkn-internal_risk-assessment"
	BuiltinToolBoxName       = "BKN风险评估工具"
	BuiltinToolConfigVersion = "0.5.0"
	BuiltinToolToolID        = "bkn_common_risk_assessment_tool"
)

// RiskType 风险类
type RiskType struct {
	RTID       string `json:"id" mapstructure:"id"`
	RTName     string `json:"name" mapstructure:"name"`
	CommonInfo `mapstructure:",squash"`
	KNID       string `json:"kn_id" mapstructure:"kn_id"`
	Branch     string `json:"branch" mapstructure:"branch"`

	Creator    AccountInfo `json:"creator" mapstructure:"creator"`
	CreateTime int64       `json:"create_time" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time" mapstructure:"update_time"`
	ModuleType string      `json:"module_type" mapstructure:"module_type"`

	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

// RiskTypesQueryParams 风险类查询参数
type RiskTypesQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	Tag         string
	Branch      string
	KNID        string
}

var (
	RiskTypeSort = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}
)

// RiskTypes 风险类列表
type RiskTypes struct {
	Entries     []*RiskType `json:"entries"`
	TotalCount  int64       `json:"total_count,omitempty"`
	SearchAfter []any       `json:"search_after,omitempty"`
	OverallMs   int64       `json:"overall_ms"`
}
