package skillvalobj

import (
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum/skillenum"
	"github.com/pkg/errors"
)

type CurrentPmsCheckStatusT string

const (
	CurrentPmsCheckStatusSuccess CurrentPmsCheckStatusT = "success"
	CurrentPmsCheckStatusFailed  CurrentPmsCheckStatusT = "failed"
)

// Tool 表示工具配置
type SkillAgent struct {
	AgentKey                        string                 `json:"agent_key" binding:"required"`      // Agent Key
	AgentVersion                    string                 `json:"agent_version"`                     // Agent 版本
	AgentInput                      json.RawMessage        `json:"agent_input"`                       // 技能输入
	Intervention                    bool                   `json:"intervention"`                      // 是否启用干预
	InterventionConfirmationMessage string                 `json:"intervention_confirmation_message"` // 人工干预确认消息
	DataSourceConfig                *DataSourceConfig      `json:"data_source_config"`                // 数据源配置
	LlmConfig                       *LLMConfig             `json:"llm_config"`                        // LLM 配置
	AgentTimeout                    int                    `json:"agent_timeout"`                     // Agent工具 调用超时时间
	CurrentPmsCheckStatus           CurrentPmsCheckStatusT `json:"current_pms_check_status"`          // 当前此技能的使用权限状态 【注意】：这个字段是调用详情接口时设置的，不是配置时设置的。如果为空字符串，表示未检查（agent不存在或未发布）
	CurrentIsExistsAndPublished     bool                   `json:"current_is_exists_and_published"`   // 当前是否存在并已发布
}

// ValObjCheck 验证工具配置
func (p *SkillAgent) ValObjCheck() (err error) {
	// 检查AgentID是否为空
	if p.AgentKey == "" {
		err = errors.New("[Tool]: agent_key is required")
		return
	}

	return
}

type DataSourceConfig struct {
	Type            skillenum.Datasource                `json:"type"`
	SpecificInherit skillenum.DatasourceSpecificInherit `json:"specific_inherit"`
}

type LLMConfig struct {
	Type skillenum.LLM `json:"type"`
}
