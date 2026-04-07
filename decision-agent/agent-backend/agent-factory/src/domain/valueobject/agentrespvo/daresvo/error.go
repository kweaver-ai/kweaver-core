package daresvo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentresperr"
)

func (r *DataAgentRes) GetExecutorError() (respErr *agentresperr.RespError) {
	if r.Error == nil {
		return
	}

	respErr = agentresperr.NewRespError(agentresperr.RespErrorTypeAgentExecutor, r.Error)

	return
}
