package sessionsvc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iagentexecutorhttp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/iredisaccess/isessionredis"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver"
)

type sessionSvc struct {
	logger          icmp.Logger
	sessionRedisAcc isessionredis.ISessionRedisAcc
	agentExecutorV1 iagentexecutorhttp.IAgentExecutor
}

var _ iportdriver.ISessionSvc = &sessionSvc{}

type NewSessionSvcDto struct {
	Logger          icmp.Logger
	SessionRedis    isessionredis.ISessionRedisAcc
	AgentExecutorV1 iagentexecutorhttp.IAgentExecutor
}

func NewSessionService(dto *NewSessionSvcDto) iportdriver.ISessionSvc {
	return &sessionSvc{
		logger:          dto.Logger,
		sessionRedisAcc: dto.SessionRedis,
		agentExecutorV1: dto.AgentExecutorV1,
	}
}
