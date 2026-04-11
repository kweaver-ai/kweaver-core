package swagger

import (
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
)

// APIError API 错误响应
type APIError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"请求参数错误"`
	Details string `json:"details,omitempty" example:"详细错误信息"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Total int64 `json:"total" example:"100"`
	List  any   `json:"list"`
}

// ChatResp Agent 对话响应
type ChatResp = agentresp.ChatResp

// ChatReq Agent 对话请求
type ChatReq = agentreq.ChatReq

// ResumeReq 恢复对话请求
type ResumeReq = agentreq.ResumeReq

// ListReq 会话列表请求
type ListReq = conversationreq.ListReq
