package agentconfigresp

import (
	_ "embed"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type FieldInfo struct {
	Name        string       `json:"name"`               // 字段名称
	Type        string       `json:"type"`               // 字段类型
	Description string       `json:"description"`        // 字段描述
	Children    []*FieldInfo `json:"children,omitempty"` // 子字段列表
}

type SelfConfigField FieldInfo

func NewSelfConfigField() *SelfConfigField {
	return &SelfConfigField{}
}

//go:embed self_config_fields.json
var SelfConfigFieldJSONStr string

func (f *SelfConfigField) LoadFromJSONStr() (err error) {
	err = cutil.JSON().Unmarshal([]byte(SelfConfigFieldJSONStr), f)

	return
}
