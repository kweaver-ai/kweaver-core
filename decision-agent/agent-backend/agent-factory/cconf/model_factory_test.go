package cconf

import (
	"testing"
)

func TestModelFactoryConf_Fields(t *testing.T) {
	t.Run("model factory config fields are accessible", func(t *testing.T) {
		mf := ModelFactoryConf{
			ModelApiSvc: SvcConf{
				Protocol: "http",
				Host:     "model-api.local",
				Port:     8000,
			},
			ModelManagerSvc: SvcConf{
				Protocol: "http",
				Host:     "model-manager.local",
				Port:     8001,
			},
			LLM: LLMConf{
				DefaultModelName: "gpt-4",
				APIKey:           "test-api-key",
			},
		}

		if mf.ModelApiSvc.Host != "model-api.local" {
			t.Errorf("Expected ModelApiSvc.Host to be 'model-api.local', got '%s'", mf.ModelApiSvc.Host)
		}

		if mf.ModelManagerSvc.Port != 8001 {
			t.Errorf("Expected ModelManagerSvc.Port to be 8001, got %d", mf.ModelManagerSvc.Port)
		}

		if mf.LLM.DefaultModelName != "gpt-4" {
			t.Errorf("Expected LLM.DefaultModelName to be 'gpt-4', got '%s'", mf.LLM.DefaultModelName)
		}
	})
}

func TestLLMConf_Fields(t *testing.T) {
	t.Run("LLM config fields are accessible", func(t *testing.T) {
		llm := LLMConf{
			DefaultModelName: "gpt-3.5-turbo",
			APIKey:           "sk-123456",
		}

		if llm.DefaultModelName != "gpt-3.5-turbo" {
			t.Errorf("Expected DefaultModelName to be 'gpt-3.5-turbo', got '%s'", llm.DefaultModelName)
		}

		if llm.APIKey != "sk-123456" {
			t.Errorf("Expected APIKey to be 'sk-123456', got '%s'", llm.APIKey)
		}
	})
}
