package imodelfactoryacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/sashabaranov/go-openai"
)

type IModelApiAcc interface {
	StreamChatCompletion(ctx context.Context, req *ChatCompletionReq) (chan string, chan error, error)
	ChatCompletion(ctx context.Context, req *ChatCompletionReq) (openai.ChatCompletionResponse, error)
}
type ChatCompletionReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`

	UserID      string            `json:"-"`
	AccountType cenum.AccountType `json:"-"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	// Usage             Usage                  `json:"usage"`
	// SystemFingerprint string                 `json:"system_fingerprint"`
}

type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Delta        ChatCompletionMessage `json:"delta"`
	FinishReason string                `json:"finish_reason"`
}
type ChatCompletionMessage struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}
