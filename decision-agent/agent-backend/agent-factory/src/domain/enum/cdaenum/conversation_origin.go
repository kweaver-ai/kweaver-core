package cdaenum

import "github.com/pkg/errors"

// 对话来源
type ConversationOrigin string

func (t ConversationOrigin) EnumCheck() (err error) {
	if t != ConversationWebChat && t != ConversationAPICall {
		err = errors.New("对话来源不合法")
		return
	}

	return
}

const (
	// 通过浏览器发起对话
	ConversationWebChat ConversationOrigin = "web_chat"

	// 通过API发起对话
	ConversationAPICall ConversationOrigin = "api_call"
)
