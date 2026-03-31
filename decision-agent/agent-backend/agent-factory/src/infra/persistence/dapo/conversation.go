package dapo

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"

type ConversationPO struct {
	ID          string `json:"id" db:"f_id"`
	AgentAPPKey string `json:"agent_app_key" db:"f_agent_app_key"`

	Title            string                     `json:"title" db:"f_title"`
	Origin           cdaenum.ConversationOrigin `json:"origin" db:"f_origin"`
	MessageIndex     int                        `json:"message_index" db:"f_message_index"`
	ReadMessageIndex int                        `json:"read_message_index" db:"f_read_message_index"`

	Ext *string `json:"ext" db:"f_ext"`

	CreateTime int64  `json:"create_time" db:"f_create_time"`
	UpdateTime int64  `json:"update_time" db:"f_update_time"`
	CreateBy   string `json:"create_by" db:"f_create_by"`
	UpdateBy   string `json:"update_by" db:"f_update_by"`
	IsDeleted  int    `json:"is_deleted" db:"f_is_deleted"`
}

func (p *ConversationPO) TableName() string {
	return "t_data_agent_conversation"
}
