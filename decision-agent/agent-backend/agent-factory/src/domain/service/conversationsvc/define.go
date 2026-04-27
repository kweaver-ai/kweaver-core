package conversationsvc

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iagentexecutorhttp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/isandboxhtpp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iv2agentexecutorhttp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver"
)

type conversationSvc struct {
	*service.SvcBase
	logger              icmp.Logger
	conversationRepo    idbaccess.IConversationRepo
	conversationMsgRepo idbaccess.IConversationMsgRepo
	agentExecutorV1     iagentexecutorhttp.IAgentExecutor
	agentExecutorV2     iv2agentexecutorhttp.IV2AgentExecutor
	sandboxPlatform     isandboxhtpp.ISandboxPlatform
	sandboxPlatformConf *conf.SandboxPlatformConf
}

var _ iportdriver.IConversationSvc = &conversationSvc{}

type NewConversationSvcDto struct {
	SvcBase             *service.SvcBase
	ConversationRepo    idbaccess.IConversationRepo
	ConversationMsgRepo idbaccess.IConversationMsgRepo
	Logger              icmp.Logger
	OpenAICmp           icmp.IOpenAI
	UmHttp              iusermanagementacc.UserMgnt
	AgentExecutorV1     iagentexecutorhttp.IAgentExecutor
	AgentExecutorV2     iv2agentexecutorhttp.IV2AgentExecutor
	SandboxPlatform     isandboxhtpp.ISandboxPlatform
	SandboxPlatformConf *conf.SandboxPlatformConf
}

func NewConversationService(dto *NewConversationSvcDto) iportdriver.IConversationSvc {
	impl := &conversationSvc{
		SvcBase:             dto.SvcBase,
		conversationRepo:    dto.ConversationRepo,
		conversationMsgRepo: dto.ConversationMsgRepo,
		logger:              dto.Logger,
		agentExecutorV1:     dto.AgentExecutorV1,
		agentExecutorV2:     dto.AgentExecutorV2,
		sandboxPlatform:     dto.SandboxPlatform,
		sandboxPlatformConf: dto.SandboxPlatformConf,
	}

	return impl
}
