package umcmp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
)

func TestNewUmCmp(t *testing.T) {
	t.Parallel()

	t.Run("valid um component", func(t *testing.T) {
		t.Parallel()

		umConf := &cconf.UserMgntCfg{
			Protocol: "http",
			Host:     "localhost",
			Port:     8080,
		}

		um := NewUmCmp(umConf, nil)

		if um == nil {
			t.Fatal("Expected um component to be created, got nil")
		}

		if um.umConf != umConf {
			t.Error("Expected umConf to be set")
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		um := NewUmCmp(nil, nil)

		if um == nil {
			t.Fatal("Expected um component to be created even with nil config")
		}

		if um.umConf != nil {
			t.Error("Expected umConf to be nil")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		t.Parallel()

		umConf := &cconf.UserMgntCfg{
			Protocol: "http",
			Host:     "localhost",
			Port:     8080,
		}

		um := NewUmCmp(umConf, nil)

		if um == nil {
			t.Fatal("Expected um component to be created even with nil logger")
		}
	})
}
