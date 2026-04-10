package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/conversation.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IConversationRepo
type IConversationRepo interface {
	IDBAccBaseRepo

	Create(ctx context.Context, po *dapo.ConversationPO) (rt *dapo.ConversationPO, err error)
	Update(ctx context.Context, po *dapo.ConversationPO) (err error)
	Delete(ctx context.Context, tx *sql.Tx, id string) (err error)
	DeleteByAPPKey(ctx context.Context, tx *sql.Tx, appKey string) (err error)

	GetByID(ctx context.Context, id string) (po *dapo.ConversationPO, err error)

	List(ctx context.Context, req conversationreq.ListReq) (rt []*dapo.ConversationPO, count int64, err error)
	ListByAgentID(ctx context.Context, agentID, title string, page, size int) (rt []*dapo.ConversationPO, count int64, err error)
}
