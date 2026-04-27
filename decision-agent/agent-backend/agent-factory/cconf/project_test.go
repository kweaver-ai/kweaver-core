package cconf

import (
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.uber.org/zap/zapcore"
)

func TestProject_Fields(t *testing.T) {
	t.Run("project fields are accessible", func(t *testing.T) {
		project := Project{
			Host:        "0.0.0.0",
			Port:        30777,
			Language:    rest.SimplifiedChinese,
			LoggerLevel: zapcore.InfoLevel,
			LogFile:     "/var/log/agent-factory.log",
		}

		if project.Host != "0.0.0.0" {
			t.Errorf("Expected Host to be '0.0.0.0', got '%s'", project.Host)
		}

		if project.Port != 30777 {
			t.Errorf("Expected Port to be 30777, got %d", project.Port)
		}

		if project.Language != rest.SimplifiedChinese {
			t.Errorf("Expected Language to be SimplifiedChinese, got %v", project.Language)
		}

		if project.LogFile != "/var/log/agent-factory.log" {
			t.Errorf("Expected LogFile to be '/var/log/agent-factory.log', got '%s'", project.LogFile)
		}
	})
}

func TestProject_EmptyHost(t *testing.T) {
	t.Run("empty host is valid", func(t *testing.T) {
		project := Project{
			Host:     "",
			Port:     8080,
			Language: rest.SimplifiedChinese,
		}

		err := project.Check()
		if err != nil {
			t.Errorf("Expected Check to succeed with empty host, got %v", err)
		}
	})
}

func TestProject_ZeroPort(t *testing.T) {
	t.Run("zero port is valid", func(t *testing.T) {
		project := Project{
			Host:     "localhost",
			Port:     0,
			Language: rest.SimplifiedChinese,
		}

		err := project.Check()
		if err != nil {
			t.Errorf("Expected Check to succeed with zero port, got %v", err)
		}
	})
}
