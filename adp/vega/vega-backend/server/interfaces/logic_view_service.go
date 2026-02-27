// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
)

const (
	LogicType_Derived   = "derived"
	LogicType_Composite = "composite"

	//特征的配置项
	PropertyFeatureType_Keyword  = "keyword"
	PropertyFeatureType_Fulltext = "fulltext"
	PropertyFeatureType_Vector   = "vector"

	LogicDefinitionNodeType_Resource = "resource"
	LogicDefinitionNodeType_Join     = "join"
	LogicDefinitionNodeType_Union    = "union"
	LogicDefinitionNodeType_Sql      = "sql"
	LogicDefinitionNodeType_Output   = "output"

	// join的类型
	JoinType_Inner     = "inner"
	JoinType_Left      = "left"
	JoinType_Right     = "right"
	JoinType_FullOuter = "full outer"

	// union的类型
	UnionType_All      = "all"
	UnionType_Distinct = "distinct"

	// MaxRecursionDepth 逻辑视图最大嵌套深度，防止循环引用导致栈溢出
	MaxRecursionDepth = 10
)

const (
	// 视图字段名称、字段显示名、字段备注、字段特征备注的最大长度
	MaxLength_ViewPropertyName               = 255
	MaxLength_ViewPropertyDisplayName        = 255
	MaxLength_ViewPropertyFeatureName        = 255
	MaxLength_ViewPropertyDescription        = 1000
	MaxLength_ViewPropertyFeatureDescription = 1000

	RegexPattern_NonBuiltin_ViewID = "^[a-z0-9][a-z0-9_-]{0,39}$"
)

var (
	LogicDefinitionNodeTypeMap = map[string]struct{}{
		LogicDefinitionNodeType_Resource: {},
		LogicDefinitionNodeType_Join:     {},
		LogicDefinitionNodeType_Union:    {},
		LogicDefinitionNodeType_Sql:      {},
		LogicDefinitionNodeType_Output:   {},
	}

	JoinTypeMap = map[string]struct{}{
		JoinType_Inner:     {},
		JoinType_Left:      {},
		JoinType_Right:     {},
		JoinType_FullOuter: {},
	}

	UnionTypeMap = map[string]struct{}{
		UnionType_All:      {},
		UnionType_Distinct: {},
	}

	PropertyFeatureTypeMap = map[string]struct{}{
		PropertyFeatureType_Keyword:  {},
		PropertyFeatureType_Fulltext: {},
		PropertyFeatureType_Vector:   {},
	}
)

type LogicView struct {
	Resource
	FieldsMap      map[string]*ViewProperty `json:"fields_map,omitempty" mapstructure:"-"`
	IsSingleSource bool                     `json:"is_single_source,omitempty" mapstructure:"-"`
	RefResources   map[string]*Resource     `json:"ref_resources,omitempty" mapstructure:"-"`
}

// LogicDefinitionNode 表示图中的节点
type LogicDefinitionNode struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Inputs       []string        `json:"inputs"`
	Config       map[string]any  `json:"config"`
	OutputFields []*ViewProperty `json:"output_fields"`
}

// 节点类型为resource的节点配置
type ResourceNodeCfg struct {
	ResourceID string         `json:"resource_id" mapstructure:"resource_id"`
	Filters    *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
	Distinct   bool           `json:"distinct" mapstructure:"distinct"`
	Resource   *Resource      `json:"resource,omitempty" mapstructure:"resource"`
}

// 节点类型为join的节点配置
type JoinNodeCfg struct {
	JoinType string         `json:"join_type" mapstructure:"join_type"`
	JoinOn   []*JoinOn      `json:"join_on" mapstructure:"join_on"`
	Filters  *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
}

// join on 配置
type JoinOn struct {
	LeftField  string `json:"left_field" mapstructure:"left_field"`   //传递 name
	RightField string `json:"right_field" mapstructure:"right_field"` //传递 name
	Operator   string `json:"operator" mapstructure:"operator"`
}

// 节点类型为union的节点配置
type UnionNodeCfg struct {
	UnionType string         `json:"union_type" mapstructure:"union_type"`
	Filters   *FilterCondCfg `json:"filters,omitempty" mapstructure:"filters"`
}

type SQLNodeCfg struct {
	SQL string `json:"sql" mapstructure:"sql"`
}

// OutputFieldRef 表示 Union 对齐模式中 from 数组的元素
type OutputFieldRef struct {
	From     string `json:"from"`
	FromNode string `json:"from_node"`
}

// 逻辑视图字段
type ViewProperty struct {
	Property
	From     string            `json:"from,omitempty"`      // Join 映射模式：源字段名 (当 from 为 string 时)
	FromNode string            `json:"from_node,omitempty"` // Join 映射模式：源节点ID
	FromList []*OutputFieldRef `json:"-"`                   // Union 对齐模式：多源对齐数组 (当 from 为 array 时)
}

// UnmarshalJSON 自定义反序列化，处理 output_fields 的 5 种形态
func (v *ViewProperty) UnmarshalJSON(data []byte) error {
	// 1. 探测是否为纯字符串（通配符模式 "*" 或 投影模式 "field_a"）
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v.Name = s
		return nil
	}

	// 2. 探测是否为对象
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 解码基类 Property 的字段 (Name, Type, DisplayName, OriginalName, Description, Features)
	type PropertyAlias Property
	var propAlias PropertyAlias
	if err := json.Unmarshal(data, &propAlias); err != nil {
		return err
	}
	v.Property = Property(propAlias)

	// 解码 from_node
	if rawFromNode, ok := raw["from_node"]; ok {
		_ = json.Unmarshal(rawFromNode, &v.FromNode)
	}

	// 解码 from: 可能是 string (映射模式) 或 array (对齐模式)
	if rawFrom, ok := raw["from"]; ok {
		// 尝试 string
		var fromStr string
		if err := json.Unmarshal(rawFrom, &fromStr); err == nil {
			v.From = fromStr
		} else {
			// 尝试 array
			var fromList []*OutputFieldRef
			if err := json.Unmarshal(rawFrom, &fromList); err == nil {
				v.FromList = fromList
			}
		}
	}

	return nil
}

// MarshalJSON 自定义序列化，为了精简输出并符合 5 种形态
func (v *ViewProperty) MarshalJSON() ([]byte, error) {
	// 如果只有 Name 且没有其他元数据或映射信息，序列化为纯字符串 (形态 1 & 2)
	// 判断条件：Name 非空，且 Type, From, FromNode, FromList, DisplayName 等其他关键字段均为空
	if v.Name != "" && v.Type == "" && v.From == "" && v.FromNode == "" &&
		len(v.FromList) == 0 && v.DisplayName == "" && v.OriginalName == "" &&
		v.Description == "" && len(v.Features) == 0 {
		return json.Marshal(v.Name)
	}

	// 否则序列化为对象 (形态 3, 4, 5)
	type Alias ViewProperty
	// 使用辅助结构体处理 from 字段的多态输出
	tmp := struct {
		*Alias
		From any `json:"from,omitempty"`
	}{
		Alias: (*Alias)(v),
	}

	if len(v.FromList) > 0 {
		tmp.From = v.FromList
	} else if v.From != "" {
		tmp.From = v.From
	}

	return json.Marshal(tmp)
}

func (v *ViewProperty) String() string {
	return fmt.Sprintf("ViewProperty{name: %s, type: %s, from: %s, from_node: %s, from_list_len: %d}",
		v.Name, v.Type, v.From, v.FromNode, len(v.FromList))
}

type DSLCfg struct {
	From           int              `json:"from"`
	Size           int              `json:"size"`
	Sort           []map[string]any `json:"sort,omitempty"`
	TrackScores    bool             `json:"track_scores,omitempty"`
	TrackTotalHits bool             `json:"track_total_hits,omitempty"`
	SearchAfter    []any            `json:"search_after,omitempty"`
	Query          struct {
		Bool struct {
			Should         []any `json:"should,omitempty"`
			Filter         []any `json:"filter,omitempty"`
			Must           []any `json:"must,omitempty"`
			MinShouldMatch int   `json:"minimum_should_match,omitempty"`
		} `json:"bool"`
	} `json:"query"`
	Pit *struct {
		ID        string `json:"id,omitempty"`
		KeepAlive string `json:"keep_alive,omitempty"`
	} `json:"pit,omitempty"`
}

func (dsl DSLCfg) String() string {
	bytes, _ := sonic.MarshalIndent(dsl, "", "  ")
	return string(bytes)
}

type SearchAfterParams struct {
	SearchAfter  []any  `json:"search_after"`
	PitID        string `json:"pit_id"`
	PitKeepAlive string `json:"pit_keep_alive"`
}

//go:generate mockgen -source ../interfaces/logic_view_service.go -destination ../interfaces/mock/mock_logic_view_service.go
type LogicViewService interface {
	// Query queries Resource data.
	Query(ctx context.Context, resource *Resource, params *ResourceDataQueryParams) ([]map[string]any, int64, error)
}
