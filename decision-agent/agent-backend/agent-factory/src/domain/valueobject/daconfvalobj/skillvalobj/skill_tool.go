package skillvalobj

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Tool 表示工具配置
type SkillTool struct {
	ToolID                          string                  `json:"tool_id" binding:"required"`        // 工具ID
	ToolBoxID                       string                  `json:"tool_box_id" binding:"required"`    // 工具箱ID
	ToolTimeout                     int                     `json:"tool_timeout" `                     // 工具调用超时时间
	ToolInput                       json.RawMessage         `json:"tool_input"`                        // 工具输入
	Intervention                    bool                    `json:"intervention"`                      // 是否启用干预
	InterventionConfirmationMessage string                  `json:"intervention_confirmation_message"` // 人工干预确认消息
	ResultProcessStrategies         []ResultProcessStrategy `json:"result_process_strategies"`         // 结果处理策略
}

type ResultProcessStrategy struct {
	Category Category `json:"category"` // 结果处理策略类型
	Strategy Strategy `json:"strategy"` // 结果处理策略
}
type Category struct {
	ID          string `json:"id"`          // 类型ID
	Name        string `json:"name"`        // 类型名称
	Description string `json:"description"` // 类型描述
}
type Strategy struct {
	ID          string `json:"id"`          // 策略ID
	Name        string `json:"name"`        // 策略名称
	Description string `json:"description"` // 策略描述
}

// ValObjCheck 验证工具配置
func (p *SkillTool) ValObjCheck() (err error) {
	// 检查ToolID是否为空
	if p.ToolID == "" {
		err = errors.New("[Tool]: tool_id is required")
		return
	}

	// 检查ToolBoxID是否为空
	if p.ToolBoxID == "" {
		err = errors.New("[Tool]: tool_box_id is required")
		return
	}

	return
}
