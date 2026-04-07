package conversationhistoryacc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	conversationHistoryRepoOnce sync.Once
	conversationHistoryRepoImpl idbaccess.IConversationHistoryRepo
)

type conversationHistoryRepo struct {
	*drivenadapter.RepoBase

	db     *sqlx.DB
	logger icmp.Logger
}

// GetLatestVersionByAgentId implements idbaccess.ReleaseHistoryRepo.

var _ idbaccess.IConversationHistoryRepo = &conversationHistoryRepo{}

func NewConversationHistoryRepo() idbaccess.IConversationHistoryRepo {
	conversationHistoryRepoOnce.Do(func() {
		conversationHistoryRepoImpl = &conversationHistoryRepo{
			db:       global.GDB,
			logger:   logger.GetLogger(),
			RepoBase: drivenadapter.NewRepoBase(),
		}
	})

	return conversationHistoryRepoImpl
}
