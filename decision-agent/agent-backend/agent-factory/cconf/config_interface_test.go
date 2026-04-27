package cconf

import (
	"testing"
)

// MockConfig implements IConf for testing
type MockConfig struct {
	debugMode bool
}

func (m *MockConfig) IsDebug() bool {
	return m.debugMode
}

func TestIConf_Interface(t *testing.T) {
	t.Run("IConf interface implementation", func(t *testing.T) {
		var _ IConf = (*Config)(nil)

		var _ IConf = (*MockConfig)(nil)
		// Test passes if compilation succeeds
	})
}

func TestMockConfig_IsDebug(t *testing.T) {
	t.Run("mock config debug true", func(t *testing.T) {
		mock := &MockConfig{debugMode: true}
		if !mock.IsDebug() {
			t.Error("Expected IsDebug to return true")
		}
	})

	t.Run("mock config debug false", func(t *testing.T) {
		mock := &MockConfig{debugMode: false}
		if mock.IsDebug() {
			t.Error("Expected IsDebug to return false")
		}
	})
}

func TestConfig_ImplementsIConf(t *testing.T) {
	t.Run("Config implements IConf", func(t *testing.T) {
		config := &Config{}

		var iConf IConf = config
		_ = iConf.IsDebug()
	})
}
