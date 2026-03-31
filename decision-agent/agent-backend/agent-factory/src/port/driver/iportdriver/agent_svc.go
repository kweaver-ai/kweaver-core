package iportdriver

import (
	"context"

	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

//go:generate mockgen -source=./agent_svc.go -destination ./iportdrivermock/agent_svc.go -package iportdrivermock
type IAgent interface {
	Chat(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error)
	// ResumeChat 恢复聊天（Session恢复）
	ResumeChat(ctx context.Context, conversationID string) (chan []byte, error)
	// TerminateChat 终止聊天
	// 如果 agentRunID 不为空，先调用 Executor 终止，再执行原有逻辑
	// 如果 interruptedAssistantMessageID 不为空，更新消息状态为 cancelled
	TerminateChat(ctx context.Context, conversationID string, agentRunID string, interruptedAssistantMessageID string) error
	GetAPIDoc(ctx context.Context, req *agentreq.GetAPIDocReq) (interface{}, error)

	// ConversationSessionInit(ctx context.Context, req *agentreq.ConversationSessionInitReq) (resp *agentresp.ConversationSessionInitResp, err error)
}
