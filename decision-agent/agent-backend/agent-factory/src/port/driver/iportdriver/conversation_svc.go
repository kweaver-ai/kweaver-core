package iportdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
)

//go:generate mockgen -source=./conversation_svc.go -destination ./iportdrivermock/conversation_svc.go -package iportdrivermock
type IConversationSvc interface {
	List(ctx context.Context, req conversationreq.ListReq) (agentList conversationresp.ListConversationResp, count int64, err error)
	Detail(ctx context.Context, id string) (res conversationresp.ConversationDetail, err error)
	Init(ctx context.Context, req conversationreq.InitReq) (rt conversationresp.InitConversationResp, err error)
	Update(ctx context.Context, req conversationreq.UpdateReq) (err error)
	Delete(ctx context.Context, id string) (err error)
	DeleteByAppKey(ctx context.Context, appKey string) (err error)
	MarkRead(ctx context.Context, id string, latest_read_index int) (err error)

	// NOTE: 获取会话中的历史上下文（新版本，支持多种策略）
	GetHistoryV2(ctx context.Context, id string, historyConfig *daconfvalobj.ConversationHistoryConfig, regenerateUserMsgID string, regenerateAssistantMsgID string) ([]*comvalobj.LLMMessage, error)

	// NOTE: 获取会话中的历史上下文（旧版本，保持兼容）
	GetHistory(ctx context.Context, id string, limit int, regenerateUserMsgID string, regenerateAssistantMsgID string) ([]*comvalobj.LLMMessage, error)

	// 根据agentID获取所有会话
	ListByAgentID(ctx context.Context, agentID, title string, page, size int, startTime, endTime int64) ([]conversationresp.ConversationDetail, int64, error)
}
