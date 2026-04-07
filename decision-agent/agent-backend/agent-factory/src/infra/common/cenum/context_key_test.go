package cenum

import (
	"testing"
)

func TestContextKey_String(t *testing.T) {
	t.Parallel()

	t.Run("string conversion", func(t *testing.T) {
		t.Parallel()

		key := ContextKey("test-key")
		result := key.String()

		if result != "test-key" {
			t.Errorf("Expected 'test-key', got '%s'", result)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		t.Parallel()

		key := ContextKey("")
		result := key.String()

		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

func TestContextKey_Constants(t *testing.T) {
	t.Parallel()

	t.Run("ContextKeyPrefix constant", func(t *testing.T) {
		t.Parallel()

		if ContextKeyPrefix != "ctx_key_" {
			t.Errorf("Expected ContextKeyPrefix to be 'ctx_key_', got '%s'", ContextKeyPrefix)
		}
	})

	t.Run("VisitLangCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := VisitLangCtxKey.String()
		expected := "ctx_key_visit_language"

		if result != expected {
			t.Errorf("Expected VisitLangCtxKey to be '%s', got '%s'", expected, result)
		}
	})

	t.Run("VisitUserIDCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := VisitUserIDCtxKey.String()
		expected := "ctx_key_visit_user_id"

		if result != expected {
			t.Errorf("Expected VisitUserIDCtxKey to be '%s', got '%s'", expected, result)
		}
	})

	t.Run("VisitUserTokenCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := VisitUserTokenCtxKey.String()
		expected := "ctx_key_visit_user_token"

		if result != expected {
			t.Errorf("Expected VisitUserTokenCtxKey to be '%s', got '%s'", expected, result)
		}
	})

	t.Run("TraceIDCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := TraceIDCtxKey.String()
		expected := "ctx_key_trace_id"

		if result != expected {
			t.Errorf("Expected TraceIDCtxKey to be '%s', got '%s'", expected, result)
		}
	})

	t.Run("VisitUserInfoCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := VisitUserInfoCtxKey.String()
		expected := "ctx_key_visit_user_info"

		if result != expected {
			t.Errorf("Expected VisitUserInfoCtxKey to be '%s', got '%s'", expected, result)
		}
	})

	t.Run("InternalAPIFlagCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := InternalAPIFlagCtxKey.String()
		expected := "ctx_key_internal_api_flag"

		if result != expected {
			t.Errorf("Expected InternalAPIFlagCtxKey to be '%s', got '%s'", expected, result)
		}
	})

	t.Run("BizDomainIDCtxKey constant", func(t *testing.T) {
		t.Parallel()

		result := BizDomainIDCtxKey.String()
		expected := "ctx_key_biz_domain_id"

		if result != expected {
			t.Errorf("Expected BizDomainIDCtxKey to be '%s', got '%s'", expected, result)
		}
	})
}
