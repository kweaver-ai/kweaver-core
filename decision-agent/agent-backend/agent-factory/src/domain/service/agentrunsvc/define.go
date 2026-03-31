package agentsvc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iagentexecutorhttp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/isandboxhtpp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iv2agentexecutorhttp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
)

type agentSvc struct {
	*service.SvcBase
	logger          icmp.Logger
	squareSvc       iv3portdriver.ISquareSvc
	agentExecutorV1 iagentexecutorhttp.IAgentExecutor
	agentExecutorV2 iv2agentexecutorhttp.IV2AgentExecutor
	conversationSvc iportdriver.IConversationSvc
	sessionSvc      iportdriver.ISessionSvc
	sandboxPlatform isandboxhtpp.ISandboxPlatform

	conversationRepo    idbaccess.IConversationRepo
	conversationMsgRepo idbaccess.IConversationMsgRepo
	streamDiffFrequency int
	sandboxPlatformConf *conf.SandboxPlatformConf

	SessionMap  sync.Map
	progressMap sync.Map
	progressSet sync.Map
}

var _ iportdriver.IAgent = &agentSvc{}

type NewAgentSvcDto struct {
	SvcBase             *service.SvcBase
	Logger              icmp.Logger
	SquareSvc           iv3portdriver.ISquareSvc
	AgentExecutorV1     iagentexecutorhttp.IAgentExecutor
	AgentExecutorV2     iv2agentexecutorhttp.IV2AgentExecutor
	ConversationSvc     iportdriver.IConversationSvc
	SessionSvc          iportdriver.ISessionSvc
	SandboxPlatform     isandboxhtpp.ISandboxPlatform
	SandboxPlatformConf *conf.SandboxPlatformConf
	ConversationRepo    idbaccess.IConversationRepo
	ConversationMsgRepo idbaccess.IConversationMsgRepo
	StreamDiffFrequency int
}

func NewAgentSvc(dto *NewAgentSvcDto) iportdriver.IAgent {
	impl := &agentSvc{
		SvcBase:             dto.SvcBase,
		logger:              dto.Logger,
		squareSvc:           dto.SquareSvc,
		agentExecutorV1:     dto.AgentExecutorV1,
		agentExecutorV2:     dto.AgentExecutorV2,
		conversationSvc:     dto.ConversationSvc,
		sessionSvc:          dto.SessionSvc,
		sandboxPlatform:     dto.SandboxPlatform,
		sandboxPlatformConf: dto.SandboxPlatformConf,
		conversationRepo:    dto.ConversationRepo,
		conversationMsgRepo: dto.ConversationMsgRepo,
		streamDiffFrequency: dto.StreamDiffFrequency,
		SessionMap:          sync.Map{},
		progressMap:         sync.Map{},
		progressSet:         sync.Map{},
	}

	return impl
}
