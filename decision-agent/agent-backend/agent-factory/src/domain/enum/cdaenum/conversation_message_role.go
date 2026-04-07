package cdaenum

import "github.com/pkg/errors"

// 消息所属角色
type ConversationMsgRole string

func (t ConversationMsgRole) EnumCheck() (err error) {
	if t != MsgRoleUser && t != MsgRoleAssistant {
		err = errors.New("角色不合法")
		return
	}

	return
}

const (
	// 用户角色
	MsgRoleUser ConversationMsgRole = "user"
	// 助手角色
	MsgRoleAssistant ConversationMsgRole = "assistant"
)
