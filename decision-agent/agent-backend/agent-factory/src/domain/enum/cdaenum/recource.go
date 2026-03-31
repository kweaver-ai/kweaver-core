package cdaenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type ResourceType string

const (
	ResourceTypeDataAgent    ResourceType = "agent"     // 数据智能体
	ResourceTypeDataAgentTpl ResourceType = "agent_tpl" // 数据智能体模板
)

func (e ResourceType) String() string {
	return string(e)
}

// EnumCheck
func (e ResourceType) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]ResourceType{ResourceTypeDataAgent, ResourceTypeDataAgentTpl}, e) {
		err = errors.New("[ResourceType]: invalid resource type")
		return
	}

	return
}
