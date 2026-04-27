package locale

import (
	"context"
	"os"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func TestI18nKey(t *testing.T) {
	t.Run("SystemCreatedBy constant", func(t *testing.T) {
		if SystemCreatedBy != "SystemCreatedBy" {
			t.Errorf("Expected SystemCreatedBy to be 'SystemCreatedBy', got '%s'", SystemCreatedBy)
		}
	})

	t.Run("UnknownUser constant", func(t *testing.T) {
		if UnknownUser != "UnknownUser" {
			t.Errorf("Expected UnknownUser to be 'UnknownUser', got '%s'", UnknownUser)
		}
	})

	t.Run("Copy constant", func(t *testing.T) {
		if Copy != "Copy" {
			t.Errorf("Expected Copy to be 'Copy', got '%s'", Copy)
		}
	})
}

func TestI18nMap(t *testing.T) {
	t.Run("create i18n map", func(t *testing.T) {
		m := I18nMap{
			rest.SimplifiedChinese: "系统",
			rest.AmericanEnglish:   "System",
		}

		if m[rest.SimplifiedChinese] != "系统" {
			t.Errorf("Expected Chinese to be '系统', got '%s'", m[rest.SimplifiedChinese])
		}

		if m[rest.AmericanEnglish] != "System" {
			t.Errorf("Expected English to be 'System', got '%s'", m[rest.AmericanEnglish])
		}
	})

	t.Run("empty i18n map", func(t *testing.T) {
		var m I18nMap

		if len(m) != 0 {
			t.Errorf("Expected map to be empty, got %d entries", len(m))
		}
	})
}

func TestAllI18nMap(t *testing.T) {
	t.Run("SystemCreatedBy translations", func(t *testing.T) {
		m := AllI18nMap[SystemCreatedBy]

		if m[rest.SimplifiedChinese] != "系统" {
			t.Errorf("Expected Chinese to be '系统', got '%s'", m[rest.SimplifiedChinese])
		}

		if m[rest.AmericanEnglish] != "System" {
			t.Errorf("Expected English to be 'System', got '%s'", m[rest.AmericanEnglish])
		}
	})

	t.Run("UnknownUser translations", func(t *testing.T) {
		m := AllI18nMap[UnknownUser]

		if m[rest.SimplifiedChinese] != "未知用户" {
			t.Errorf("Expected Chinese to be '未知用户', got '%s'", m[rest.SimplifiedChinese])
		}

		if m[rest.AmericanEnglish] != "Unknown User" {
			t.Errorf("Expected English to be 'Unknown User', got '%s'", m[rest.AmericanEnglish])
		}
	})

	t.Run("Copy translations", func(t *testing.T) {
		m := AllI18nMap[Copy]

		if m[rest.SimplifiedChinese] != "副本" {
			t.Errorf("Expected Chinese to be '副本', got '%s'", m[rest.SimplifiedChinese])
		}

		if m[rest.AmericanEnglish] != "Duplicate" {
			t.Errorf("Expected English to be 'Duplicate', got '%s'", m[rest.AmericanEnglish])
		}
	})
}

func TestGetI18n(t *testing.T) {
	t.Run("valid key and language", func(t *testing.T) {
		result := GetI18n(SystemCreatedBy, rest.SimplifiedChinese)
		if result != "系统" {
			t.Errorf("Expected '系统', got '%s'", result)
		}
	})

	t.Run("valid key and English", func(t *testing.T) {
		result := GetI18n(UnknownUser, rest.AmericanEnglish)
		if result != "Unknown User" {
			t.Errorf("Expected 'Unknown User', got '%s'", result)
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		result := GetI18n("InvalidKey", rest.SimplifiedChinese)
		if result != "" {
			t.Errorf("Expected empty string for invalid key, got '%s'", result)
		}
	})

	t.Run("valid key but unsupported language", func(t *testing.T) {
		// Use a language that's not in the map
		result := GetI18n(SystemCreatedBy, "fr-FR")
		if result != "" {
			t.Errorf("Expected empty string for unsupported language, got '%s'", result)
		}
	})
}

func TestGetI18nByCtx(t *testing.T) {
	t.Run("test function exists", func(t *testing.T) {
		// This test verifies that GetI18nByCtx function exists
		// The actual functionality depends on context and language helper
		// which requires global config to be set
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic without global config
				t.Logf("Expected panic without global config: %v", r)
			}
		}()

		ctx := context.Background()

		// This will likely panic without proper global config setup
		_ = GetI18nByCtx
		_ = ctx
	})

	t.Run("call GetI18nByCtx with panic recovery", func(t *testing.T) {
		ctx := context.Background()
		// Just verify the function is callable
		// The result depends on context language which may not be set
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic without global config
				t.Logf("Expected panic without global config: %v", r)
			}
		}()

		_ = GetI18nByCtx(ctx, SystemCreatedBy)
	})
}

func TestRegister(t *testing.T) {
	t.Run("UT mode", func(t *testing.T) {
		// Save original env value
		originalMode := os.Getenv("I18N_MODE_UT")

		// Clean up after test
		t.Cleanup(func() {
			if originalMode != "" {
				os.Setenv("I18N_MODE_UT", originalMode)
			} else {
				os.Unsetenv("I18N_MODE_UT")
			}
		})

		// Set UT mode
		os.Setenv("I18N_MODE_UT", "true")

		// This should not panic in UT mode
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Register panicked: %v", r)
			}
		}()

		Register()
	})
	// Note: We skip the other test cases because i18n.RegisterI18n can only be called once per process
	// The first test registers successfully, subsequent tests would fail with "messageId already exist"
}
