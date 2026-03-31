package agentreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req/chatopt"
)

type DebugReq struct {
	AgentID        string         `json:"agent_id"`                 // agentID
	AgentVersion   string         `json:"agent_version"`            // agent版本
	Input          DebugInput     `json:"input"`                    // 输入
	ConversationID string         `json:"conversation_id"`          // 会话ID
	SelectedFiles  []SelectedFile `json:"selected_files,omitempty"` // 用户选择的临时区文件

	AgentRunID                string                              `json:"agent_run_id"`                     // Agent运行ID（中断恢复时由前端传入）
	ResumeInterruptInfo       *v2agentexecutordto.AgentResumeInfo `json:"resume_interrupt_info"`            // 中断恢复信息（为nil时走正常流程）
	InterruptedAssistantMsgID string                              `json:"interrupted_assistant_message_id"` // 中断的助手消息ID

	ChatMode string `json:"chat_mode"` // 聊天模式
	// NOTE: 新增stream参数，控制流式返回
	Stream    bool `json:"stream,omitempty"`     // 是否流式返回
	IncStream bool `json:"inc_stream,omitempty"` // 是否增量返回

	UserID      string `json:"-"` // 用户ID
	Token       string `json:"-"` // 用户token
	AgentAPPKey string `json:"-"`

	ExecutorVersion string `json:"executor_version"` // executor version v1 或 v2 默认v2

	// ConversationSessionID string `json:"conversation_session_id"`
	ChatOption chatopt.ChatOption `json:"chat_option"`
}

type DebugInput struct {
	Query        string                  `json:"query"`         // 查询内容
	CustomQuerys map[string]interface{}  `json:"custom_querys"` // 自定义查询
	History      []*comvalobj.LLMMessage `json:"history"`       // 历史
}
