package cmpopenai

import (
	"context"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestGenReq(t *testing.T) {
	t.Parallel()

	cmp := &OpenAICmp{
		model: "gpt-4",
	}

	t.Run("valid user message without system message", func(t *testing.T) {
		t.Parallel()

		userMsg := "Hello, how are you?"
		sysMsg := ""

		req, err := cmp.genReq(userMsg, sysMsg, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if req.Model != "gpt-4" {
			t.Errorf("Expected model to be 'gpt-4', got '%s'", req.Model)
		}

		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}

		if req.Messages[0].Role != openai.ChatMessageRoleUser {
			t.Errorf("Expected role to be '%s', got '%s'", openai.ChatMessageRoleUser, req.Messages[0].Role)
		}

		if req.Messages[0].Content != userMsg {
			t.Errorf("Expected content to be '%s', got '%s'", userMsg, req.Messages[0].Content)
		}

		if req.Stream {
			t.Error("Expected Stream to be false")
		}
	})

	t.Run("valid user message with system message", func(t *testing.T) {
		t.Parallel()

		userMsg := "Hello, how are you?"
		sysMsg := "You are a helpful assistant."

		req, err := cmp.genReq(userMsg, sysMsg, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}

		if req.Messages[0].Role != openai.ChatMessageRoleUser {
			t.Errorf("Expected first role to be '%s', got '%s'", openai.ChatMessageRoleUser, req.Messages[0].Role)
		}

		if req.Messages[1].Role != openai.ChatMessageRoleSystem {
			t.Errorf("Expected second role to be '%s', got '%s'", openai.ChatMessageRoleSystem, req.Messages[1].Role)
		}

		if req.Messages[1].Content != sysMsg {
			t.Errorf("Expected system content to be '%s', got '%s'", sysMsg, req.Messages[1].Content)
		}
	})

	t.Run("with stream enabled", func(t *testing.T) {
		t.Parallel()

		userMsg := "Stream this message"
		sysMsg := ""

		req, err := cmp.genReq(userMsg, sysMsg, true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !req.Stream {
			t.Error("Expected Stream to be true")
		}
	})

	t.Run("empty user message", func(t *testing.T) {
		t.Parallel()

		userMsg := ""
		sysMsg := "System message"

		req, err := cmp.genReq(userMsg, sysMsg, false)

		if err == nil {
			t.Error("Expected error for empty user message, got nil")
		}

		if err.Error() != "[OpenAICmp][StreamChat]: userMsg is empty" {
			t.Errorf("Expected specific error message, got '%s'", err.Error())
		}

		if req.Model != "" {
			t.Error("Expected request to be zero value on error")
		}
	})
}

func TestStreamChat_EmptyUserMessage(t *testing.T) {
	t.Parallel()

	cmp := &OpenAICmp{
		model: "gpt-4",
	}

	ctx := context.Background()

	stream, err := cmp.StreamChat(ctx, "", "System message")

	if err == nil {
		t.Error("Expected error for empty user message, got nil")
	}

	if err.Error() != "[OpenAICmp][StreamChat]: userMsg is empty" {
		t.Errorf("Expected specific error message, got '%s'", err.Error())
	}

	if stream != nil {
		t.Error("Expected stream to be nil on error")
	}
}

func TestChat_EmptyUserMessage(t *testing.T) {
	t.Parallel()

	cmp := &OpenAICmp{
		model: "gpt-4",
	}

	ctx := context.Background()

	res, err := cmp.Chat(ctx, "", "System message")

	if err == nil {
		t.Error("Expected error for empty user message, got nil")
	}

	if err.Error() != "[OpenAICmp][StreamChat]: userMsg is empty" {
		t.Errorf("Expected specific error message, got '%s'", err.Error())
	}

	if res.ID != "" {
		t.Error("Expected response to be zero value on error")
	}
}
