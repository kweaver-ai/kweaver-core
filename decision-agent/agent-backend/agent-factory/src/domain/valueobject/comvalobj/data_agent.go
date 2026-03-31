package comvalobj

import (
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/cdaconstant"
)

type DataAgentUniqFlag struct {
	AgentID      string `json:"agent_id"`
	AgentVersion string `json:"agent_version"`
}

func NewDataAgentUniqFlag(agentID, agentVersion string) *DataAgentUniqFlag {
	return &DataAgentUniqFlag{
		AgentID:      agentID,
		AgentVersion: agentVersion,
	}
}

func (p *DataAgentUniqFlag) ValObjCheck() (err error) {
	if p.AgentID == "" {
		err = errors.New("agent_id is required")
		return
	}

	if p.AgentVersion == "" {
		err = errors.New("agent_version is required")
		return
	}

	return
}

func (p *DataAgentUniqFlag) IsUnpublish() bool {
	return p.AgentVersion == cdaconstant.AgentVersionUnpublished
}
