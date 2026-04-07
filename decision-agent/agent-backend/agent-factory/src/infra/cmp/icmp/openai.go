package icmp

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

//go:generate mockgen -package cmpmock -source openai.go -destination ./cmpmock/openai_mock.go
type IOpenAI interface {
	StreamChat(ctx context.Context, userMsg, sysMsg string) (stream *openai.ChatCompletionStream, err error)
	Chat(ctx context.Context, userMsg, sysMsg string) (chatRes openai.ChatCompletionResponse, err error)
}
