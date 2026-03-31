package daresvo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
)

func (r *DataAgentRes) GetMiddleOutputVars() (middleOutputVarRes *agentrespvo.MiddleOutputVarRes, err error) {
	middleOutputVarRes = agentrespvo.NewMiddleOutputVarRes()

	outputVars := r.middleOutputVarsHelper.outputVariablesS.MiddleOutputVars

	if len(outputVars) == 0 {
		return
	}

	var m map[string]interface{}

	m, err = r.middleOutputVarsHelper.GetMiddleOutputVarsMap()
	if err != nil {
		return
	}

	outputVarInterventionMap := r.Answer.Interventions.ToOutputVarMap()

	err = middleOutputVarRes.LoadFrom(outputVars, m, outputVarInterventionMap)
	if err != nil {
		return
	}

	return
}
