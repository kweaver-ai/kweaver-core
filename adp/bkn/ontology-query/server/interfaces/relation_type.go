// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	RELATION_TYPE_DIRECT    = "direct"
	RELATION_TYPE_DATA_VIEW = "data_view"
)

type RelationType struct {
	RTID               string `json:"id"`
	RTName             string `json:"name"`
	SourceObjectTypeID string `json:"source_object_type_id"`
	TargetObjectTypeID string `json:"target_object_type_id"`
	Type               string `json:"type"`
	MappingRules       any    `json:"mapping_rules"` // 根据type来决定是不同的映射方式，direct对应的结构体是[]Mapping
}

// 间接关联
type InDirectMapping struct {
	BackingDataSource  *ResourceInfo `json:"backing_data_source"`
	SourceMappingRules []Mapping     `json:"source_mapping_rules"`
	TargetMappingRules []Mapping     `json:"target_mapping_rules"`
}

// 直接关联
type Mapping struct {
	SourceProp SimpleProperty `json:"source_property"`
	TargetProp SimpleProperty `json:"target_property"`
}

type SimpleProperty struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}
