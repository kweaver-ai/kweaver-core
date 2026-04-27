package cmpopenai

import (
	"testing"
)

func TestNewOpenAICmp(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		apiKey := "test-api-key"
		baseURL := "https://api.openai.com/v1"
		model := "gpt-4"

		cmp := NewOpenAICmp(apiKey, baseURL, model, false)

		if cmp == nil {
			t.Fatal("Expected component to be created, got nil")
		}

		openaiCmp, ok := cmp.(*OpenAICmp)
		if !ok {
			t.Fatal("Expected OpenAICmp type")
		}

		if openaiCmp.apiKey != apiKey {
			t.Errorf("Expected apiKey to be '%s', got '%s'", apiKey, openaiCmp.apiKey)
		}

		if openaiCmp.baseURL != baseURL {
			t.Errorf("Expected baseURL to be '%s', got '%s'", baseURL, openaiCmp.baseURL)
		}

		if openaiCmp.model != model {
			t.Errorf("Expected model to be '%s', got '%s'", model, openaiCmp.model)
		}

		if openaiCmp.client == nil {
			t.Error("Expected client to be initialized")
		}
	})

	t.Run("with TLS insecure skip verify", func(t *testing.T) {
		t.Parallel()

		apiKey := "test-api-key"
		baseURL := "https://api.openai.com/v1"
		model := "gpt-4"

		cmp := NewOpenAICmp(apiKey, baseURL, model, true)

		if cmp == nil {
			t.Fatal("Expected component to be created, got nil")
		}

		openaiCmp, ok := cmp.(*OpenAICmp)
		if !ok {
			t.Fatal("Expected OpenAICmp type")
		}

		if openaiCmp.client == nil {
			t.Error("Expected client to be initialized")
		}
	})

	t.Run("empty strings", func(t *testing.T) {
		t.Parallel()

		cmp := NewOpenAICmp("", "", "", false)

		if cmp == nil {
			t.Fatal("Expected component to be created even with empty strings")
		}

		openaiCmp, ok := cmp.(*OpenAICmp)
		if !ok {
			t.Fatal("Expected OpenAICmp type")
		}

		if openaiCmp.apiKey != "" {
			t.Errorf("Expected apiKey to be empty, got '%s'", openaiCmp.apiKey)
		}
	})
}
