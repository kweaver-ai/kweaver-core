package daenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// AgentCreatedType agent创建类型
type AgentCreatedType string

const (
	// AgentCreatedTypeCreate 手动创建
	AgentCreatedTypeCreate AgentCreatedType = "create"

	// AgentCreatedTypeCopy 模板创建
	AgentCreatedTypeCopy AgentCreatedType = "copy"

	// AgentCreatedTypeImport 导入创建
	AgentCreatedTypeImport AgentCreatedType = "import"
)

func (b AgentCreatedType) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]AgentCreatedType{AgentCreatedTypeCreate, AgentCreatedTypeCopy, AgentCreatedTypeImport}, b) {
		err = errors.New("[AgentCreatedType]: invalid agent created type")
		return
	}

	return
}
