package cmpopenai

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

// StreamChat 处理流式聊天请求并返回流式响应
func (s *OpenAICmp) StreamChat(ctx context.Context, userMsg, sysMsg string) (stream *openai.ChatCompletionStream, err error) {
	chatReq, err := s.genReq(userMsg, sysMsg, true)
	if err != nil {
		return
	}
	// 创建流式响应
	stream, err = s.client.CreateChatCompletionStream(ctx, chatReq)

	return
}

func (s *OpenAICmp) Chat(ctx context.Context, userMsg, sysMsg string) (chatRes openai.ChatCompletionResponse, err error) {
	chatReq, err := s.genReq(userMsg, sysMsg, false)
	if err != nil {
		return
	}

	// 创建非流式响应
	chatRes, err = s.client.CreateChatCompletion(ctx, chatReq)

	return
}

func (s *OpenAICmp) genReq(userMsg string, sysMsg string, isStream bool) (chatReq openai.ChatCompletionRequest, err error) {
	if userMsg == "" {
		err = errors.New("[OpenAICmp][StreamChat]: userMsg is empty")
		return
	}

	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMsg,
		},
	}

	if sysMsg != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: sysMsg,
		})
	}

	// 创建聊天请求
	chatReq = openai.ChatCompletionRequest{
		Model:    s.model,
		Messages: msgs,
		Stream:   isStream,
	}

	return
}
