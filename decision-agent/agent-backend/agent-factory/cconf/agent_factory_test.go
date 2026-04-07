package cconf

import (
	"testing"
)

func TestAgentFactoryConf_Fields(t *testing.T) {
	t.Run("agent factory config fields are accessible", func(t *testing.T) {
		af := AgentFactoryConf{
			PrivateSvc: &SvcConf{
				Protocol: "http",
				Host:     "localhost",
				Port:     8080,
			},
		}

		if af.PrivateSvc == nil {
			t.Fatal("Expected PrivateSvc to be non-nil")
		}

		if af.PrivateSvc.Protocol != "http" {
			t.Errorf("Expected Protocol to be 'http', got '%s'", af.PrivateSvc.Protocol)
		}
	})
}

func TestAgentFactoryConf_NilPrivateSvc(t *testing.T) {
	t.Run("agent factory config with nil private service", func(t *testing.T) {
		af := AgentFactoryConf{
			PrivateSvc: nil,
		}

		if af.PrivateSvc != nil {
			t.Error("Expected PrivateSvc to be nil")
		}
	})
}
