package idbaccess

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation_message/conversationmsgreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

//go:generate mockgen -package idbaccessmock -destination ./idbaccessmock/conversation_message.go github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess IConversationMsgRepo
type IConversationMsgRepo interface {
	IDBAccBaseRepo

	Create(ctx context.Context, po *dapo.ConversationMsgPO) (id string, err error)
	Update(ctx context.Context, po *dapo.ConversationMsgPO) (err error)
	Delete(ctx context.Context, id string) (err error)
	DeleteByConversationID(ctx context.Context, tx *sql.Tx, conversationID string) (err error)
	DeleteByAPPKey(ctx context.Context, tx *sql.Tx, appKey string) (err error)

	GetByID(ctx context.Context, id string) (po *dapo.ConversationMsgPO, err error)
	GetMaxIndexByID(ctx context.Context, id string) (maxIndex int, err error)

	List(ctx context.Context, req conversationmsgreq.ListReq) (rt []*dapo.ConversationMsgPO, err error)
	GetRecentMessages(ctx context.Context, conversationID string, limit int) (rt []*dapo.ConversationMsgPO, err error)
	GetLatestMsgByConversationID(ctx context.Context, conversationID string) (po *dapo.ConversationMsgPO, err error)
}
