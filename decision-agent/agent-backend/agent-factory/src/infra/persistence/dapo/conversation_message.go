package dapo

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"

type ConversationMsgPO struct {
	ID             string `json:"id" db:"f_id"`
	AgentAPPKey    string `json:"agent_app_key" db:"f_agent_app_key"`
	ConversationID string `json:"conversation_id" db:"f_conversation_id"`
	AgentID        string `json:"agent_id" db:"f_agent_id"`
	AgentVersion   string `json:"agent_version" db:"f_agent_version"`
	ReplyID        string `json:"reply_id" db:"f_reply_id"`

	Index       int                                `json:"index" db:"f_index"`
	Role        cdaenum.ConversationMsgRole        `json:"origin" db:"f_role"`
	Content     *string                            `json:"content" db:"f_content"`
	ContentType cdaenum.ConversationMsgContentType `json:"content_type" db:"f_content_type"`
	Status      cdaenum.ConversationMsgStatus      `json:"status" db:"f_status"`

	Ext *string `json:"ext" db:"f_ext"`

	CreateTime int64  `json:"create_time" db:"f_create_time"`
	UpdateTime int64  `json:"update_time" db:"f_update_time"`
	CreateBy   string `json:"create_by" db:"f_create_by"`
	UpdateBy   string `json:"update_by" db:"f_update_by"`
	IsDeleted  int    `json:"is_deleted" db:"f_is_deleted"`
}

func (p *ConversationMsgPO) TableName() string {
	return "t_data_agent_conversation_message"
}
