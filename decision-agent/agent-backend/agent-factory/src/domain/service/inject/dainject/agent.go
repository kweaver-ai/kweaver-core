package dainject

import (
	"sync"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	agentsvc "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/agentrunsvc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/conversationdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/conversationmsgdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/httpinject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	agentSvcOnce sync.Once
	agentSvcImpl iportdriver.IAgent
)

func NewAgentSvc() iportdriver.IAgent {
	agentSvcOnce.Do(func() {
		dto := &agentsvc.NewAgentSvcDto{
			SvcBase:             service.NewSvcBase(),
			Logger:              logger.GetLogger(),
			SquareSvc:           NewSquareSvc(),
			AgentExecutorV1:     httpinject.NewAgentExecutorV1HttpAcc(),
			AgentExecutorV2:     httpinject.NewAgentExecutorV2HttpAcc(),
			ConversationSvc:     NewConversationSvc(),
			SessionSvc:          NewSessionSvc(),
			SandboxPlatform:     httpinject.NewSandboxPlatformHttpAcc(),
			SandboxPlatformConf: global.GConfig.SandboxPlatformConf,
			ConversationRepo:    conversationdbacc.NewConversationRepo(),
			ConversationMsgRepo: conversationmsgdbacc.NewConversationMsgRepo(),
			// NOTE: streamDiffFrequency must be greater than 0
			StreamDiffFrequency: max(global.GConfig.StreamDiffFrequency, 1),
		}

		agentSvcImpl = agentsvc.NewAgentSvc(dto)
	})

	return agentSvcImpl
}
