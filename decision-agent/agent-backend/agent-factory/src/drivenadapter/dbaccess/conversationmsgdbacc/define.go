package conversationmsgdbacc

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	conversationMsgRepoOnce sync.Once
	conversationMsgRepoImpl idbaccess.IConversationMsgRepo
)

type ConversationMsgRepo struct {
	idbaccess.IDBAccBaseRepo

	db *sqlx.DB

	logger icmp.Logger
}

var _ idbaccess.IConversationMsgRepo = &ConversationMsgRepo{}

func NewConversationMsgRepo() idbaccess.IConversationMsgRepo {
	conversationMsgRepoOnce.Do(func() {
		conversationMsgRepoImpl = &ConversationMsgRepo{
			db:             global.GDB,
			logger:         logger.GetLogger(),
			IDBAccBaseRepo: dbaccess.NewDBAccBase(),
		}
	})

	return conversationMsgRepoImpl
}
