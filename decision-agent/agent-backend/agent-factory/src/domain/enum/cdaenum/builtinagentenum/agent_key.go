package builtinagentenum

import (
	"errors"

	"golang.org/x/exp/slices"
)

// AgentKey 内置agent key
type AgentKey string

const (
	AgentKeyDocQA        AgentKey = "DocQA_Agent"
	AgentKeyGraphQA      AgentKey = "GraphQA_Agent"
	AgentKeyOnlineSearch AgentKey = "OnlineSearch_Agent"
	AgentKeyPlan         AgentKey = "Plan_Agent"
	AgentKeySimpleChat   AgentKey = "SimpleChat_Agent"
	AgentKeySummary      AgentKey = "Summary_Agent"
	AgentKeyDeepSearch   AgentKey = "deepsearch"
)

func (e AgentKey) String() string {
	return string(e)
}

func (e AgentKey) EnumCheck() (err error) {
	if !slices.Contains([]AgentKey{AgentKeyDocQA, AgentKeyGraphQA, AgentKeyOnlineSearch, AgentKeyPlan, AgentKeySimpleChat, AgentKeySummary, AgentKeyDeepSearch}, e) {
		err = errors.New("[builtinagentenum][AgentKey]: invalid agent key")
		return
	}

	return
}

func (key AgentKey) IsDocQA() bool {
	return key == AgentKeyDocQA
}

func (key AgentKey) IsGraphQA() bool {
	return key == AgentKeyGraphQA
}
