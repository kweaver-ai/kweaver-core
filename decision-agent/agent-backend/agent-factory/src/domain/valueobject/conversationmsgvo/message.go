package conversationmsgvo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

type Message struct {
	ID             string                        `json:"id"`
	ConversationID string                        `json:"conversation_id"`
	Role           cdaenum.ConversationMsgRole   `json:"role"`
	Content        interface{}                   `json:"content"`
	ContentType    chatresenum.AnswerType        `json:"content_type"`
	Status         cdaenum.ConversationMsgStatus `json:"status"`
	ReplyID        string                        `json:"reply_id"`
	AgentInfo      valueobject.AgentInfo         `json:"agent_info"`
	Index          int                           `json:"index"`
	Ext            *MessageExt                   `json:"ext"` // 扩展字段
}

func (m *Message) IsInterrupted() bool {
	return m.Ext != nil && m.Ext.IsInterrupted()
}

//role:user
type UserContent struct {
	Text          string                  `json:"text"`
	SelectedFiles []agentreq.SelectedFile `json:"selected_files"` // 用户选择的临时区文件
}

//role:assistant
type AssistantContent struct {
	FinalAnswer  FinalAnswer   `json:"final_answer"`
	MiddleAnswer *MiddleAnswer `json:"middle_answer"`
}

type FinalAnswer struct {
	Query                 string                  `json:"query"`
	Answer                Answer                  `json:"answer"`
	SelectedFiles         []agentreq.SelectedFile `json:"selected_files"` // 用户选择的临时区文件
	Thinking              string                  `json:"thinking"`
	SkillProcess          []*SkillsProcessItem    `json:"skill_process"`
	AnswerTypeOther       interface{}             `json:"answer_type_other"`       // 当content_type为other时使用
	OutputVariablesConfig *agentconfigvo.Variable `json:"output_variables_config"` // output 输出变量配置
}

type SkillsProcessItem struct {
	AgentName      string             `json:"agent_name"`
	Text           string             `json:"text"`
	Cites          interface{}        `json:"cites,omitempty"`
	Status         string             `json:"status"`
	Type           string             `json:"type"`
	Thinking       string             `json:"thinking"`
	InputMessage   interface{}        `json:"input_message"`
	Interrupted    bool               `json:"interrupted"`
	RelatedQueries []*RelatedQuestion `json:"related_queries"`
}

type RelatedQuestion struct {
	Query string `json:"query"`
}

type MiddleAnswer struct {
	Progress       []*agentrespvo.Progress        `json:"progress"` // Dolphin中间执行过程展示
	DocRetrieval   *agentrespvo.DocRetrievalField `json:"doc_retrieval"`
	GraphRetrieval any                            `json:"graph_retrieval"`
	OtherVariables map[string]interface{}         `json:"other_variables"` // 用于存储配置中output.variables.other_vars配置的其他变量
}

type Answer struct {
	Text  string      `json:"text"`
	Cites interface{} `json:"cites,omitempty"`
	Ask   interface{} `json:"ask,omitempty"`
}
