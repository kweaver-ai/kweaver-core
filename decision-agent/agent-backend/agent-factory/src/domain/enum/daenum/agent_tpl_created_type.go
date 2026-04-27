package daenum

import "github.com/pkg/errors"

// AgentTplCreatedType agent模板创建类型
type AgentTplCreatedType string

const (
	// AgentTplCreatedTypeCopyFromAgent 模板创建
	AgentTplCreatedTypeCopyFromAgent AgentTplCreatedType = "copy_from_agent"

	// AgentTplCreatedTypeCopyFromTpl 模板创建
	AgentTplCreatedTypeCopyFromTpl AgentTplCreatedType = "copy_from_tpl"
)

func (a AgentTplCreatedType) EnumCheck() error {
	switch a {
	case AgentTplCreatedTypeCopyFromAgent, AgentTplCreatedTypeCopyFromTpl:
		return nil
	default:
		return errors.New("无效的AgentTplCreatedType")
	}
}
