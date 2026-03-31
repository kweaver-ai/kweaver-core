package cdaenum

import "github.com/pkg/errors"

type ToolType string

const (
	// ToolTypeTool 工具
	ToolTypeTool ToolType = "tool"
	// ToolTypeAgent agent
	ToolTypeAgent ToolType = "agent"
)

func (t ToolType) EnumCheck() (err error) {
	if t != ToolTypeTool && t != ToolTypeAgent {
		err = errors.New("[ToolType]: invalid tool type")
		return
	}

	return
}
