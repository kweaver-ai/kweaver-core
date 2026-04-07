// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import cond "ontology-query/common/condition"

// 行动查询请求体
type ActionQuery struct {
	InstanceIdentities []map[string]any `json:"_instance_identities,omitempty"`

	KNID         string `json:"-"`
	Branch       string `json:"-"`
	ActionTypeID string `json:"-"`
	CommonQueryParameters
}

// 行动查询返回体
type Actions struct {
	ActionType   *ActionType   `json:"action_type,omitempty"`
	ActionSource ActionSource  `json:"action_source"`
	Actions      []ActionParam `json:"actions"`
	TotalCount   int           `json:"total_count,omitempty"`
	OverallMs    int64         `json:"overall_ms"`
}

// 实例化后的行动参数
type ActionParam struct {
	InstanceID       any            `json:"_instance_id,omitempty"`       // 实例ID
	InstanceIdentity any            `json:"_instance_identity,omitempty"` // 实例唯一标识
	Display          any            `json:"display,omitempty"`            // 显示值
	Parameters       map[string]any `json:"parameters"`                   // 填入了实参的参数
	DynamicParams    map[string]any `json:"dynamic_params"`               // 动态参数map
}

type ActionType struct {
	ATID         string        `json:"id"`
	ATName       string        `json:"name"`
	ActionType   string        `json:"action_type"`
	ObjectTypeID string        `json:"object_type_id"`
	Condition    *cond.CondCfg `json:"condition,omitempty"`
	Affect       *ActionAffect `json:"affect"`
	ActionSource ActionSource  `json:"action_source"`
	Parameters   []Parameter   `json:"parameters"`
	Schedule     Schedule      `json:"schedule"`
}

type ActionAffect struct {
	ObjectTypeID string `json:"object_type_id,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

type ActionSource struct {
	Type string `json:"type" mapstructure:"type"`
	// 互斥字段，根据Type选择
	// type 为 tool
	BoxID  string `json:"box_id,omitempty"`
	ToolID string `json:"tool_id,omitempty"`
	// type 为 mcp
	McpID    string `json:"mcp_id,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
}

type Schedule struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}
