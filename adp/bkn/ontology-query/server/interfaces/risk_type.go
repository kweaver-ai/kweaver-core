// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import cond "ontology-query/common/condition"

const (
	RISK_RULE_WHEN_TYPE_CONDITION = "condition"

	// 内置风险评估工具
	BuiltinToolBoxID  = "bkn-internal_risk-assessment"
	BuiltinToolToolID = "bkn_common_risk_assessment_tool"
)

// RiskType 风险类（用于 RiskType 风险评估）
type RiskType struct {
	RTID               string        `json:"id"`
	RTName             string        `json:"name"`
	MaxAcceptableLevel string        `json:"max_acceptable_level"`
	Parameters         []Parameter   `json:"parameters"`
	RiskRules          []RiskRule    `json:"risk_rules"`
	RiskFunction       *RiskFunction `json:"risk_function"`
}

// RiskRule 风险规则
type RiskRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	When        *RiskRuleWhen `json:"when"`
	Decision    string        `json:"decision"`
	Message     string        `json:"message"`
}

// RiskRuleWhen 命中条件
type RiskRuleWhen struct {
	Type            string        `json:"type"`
	Condition       *cond.CondCfg `json:"condition,omitempty"`
	NaturalLanguage string        `json:"natural_language,omitempty"`
}

// RiskFunction 风险评估函数
type RiskFunction struct {
	Type       string      `json:"type"`
	BoxID      string      `json:"box_id,omitempty"`
	ToolID     string      `json:"tool_id,omitempty"`
	McpID      string      `json:"mcp_id,omitempty"`
	ToolName   string      `json:"tool_name,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"` // 扁平列表，每个参数的 source 标记位置 path/query/header/body
}
