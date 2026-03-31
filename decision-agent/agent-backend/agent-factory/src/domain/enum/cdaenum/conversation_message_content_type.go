package cdaenum

import "github.com/pkg/errors"

// 对话来源
type ConversationMsgContentType string

func (t ConversationMsgContentType) EnumCheck() (err error) {
	if t != MsgText && t != ExploreMsgType && t != PromptMsgType && t != OtherMsgType {
		err = errors.New("消息类型不合法")
		return
	}

	return
}

const (
	// 文本类型
	MsgText ConversationMsgContentType = "text"

	ExploreMsgType ConversationMsgContentType = "explore"
	PromptMsgType  ConversationMsgContentType = "prompt"
	OtherMsgType   ConversationMsgContentType = "other"
)
