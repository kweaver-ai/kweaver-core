package daconfvalobj

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
)

// Config 表示agent配置
type ConfigForShow struct {
	Input         *Input                                `json:"input" binding:"required"` // 输入参数
	SystemPrompt  string                                `json:"system_prompt"`            // 系统提示词
	Dolphin       string                                `json:"dolphin"`                  // Dolphin语句
	IsDolphinMode cdaenum.DolphinMode                   `json:"is_dolphin_mode"`          // 是否是dolphin模式
	DataSource    *datasourcevalobj.RetrieverDataSource `json:"data_source"`              // 数据源
	Skill         *skillvalobj.Skill                    `json:"skills"`                   // 技能
	Llms          []*LlmItem                            `json:"llms"`                     // LLM配置

	IsDataFlowSetEnabled int                  `json:"is_data_flow_set_enabled"`   // 是否启用数据流设置
	OpeningRemarkConfig  *OpeningRemarkConfig `json:"opening_remark_config"`      // 开场白配置
	PresetQuestions      []*PresetQuestion    `json:"preset_questions"`           // 预设问题列表
	Output               *Output              `json:"output"  binding:"required"` // 输出结果
}
