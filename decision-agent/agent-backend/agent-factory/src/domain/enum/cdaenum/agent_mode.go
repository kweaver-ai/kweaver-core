package cdaenum

import "github.com/pkg/errors"

// AgentMode 表示 agent 配置模式。
type AgentMode string

const (
	AgentModeDefault AgentMode = "default"
	AgentModeDolphin AgentMode = "dolphin"
	AgentModeReact   AgentMode = "react"
)

func (m AgentMode) EnumCheck() (err error) {
	switch m {
	case AgentModeDefault, AgentModeDolphin, AgentModeReact:
		return nil
	default:
		return errors.New("agent mode is invalid")
	}
}
