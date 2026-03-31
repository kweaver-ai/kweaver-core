package cconf

import (
	"testing"
)

func TestPublicService_Fields(t *testing.T) {
	t.Run("public service fields are accessible", func(t *testing.T) {
		ps := PublicService{
			PublicHost:     "example.com",
			PublicPort:     443,
			PublicProtocol: "https",
		}

		if ps.PublicHost != "example.com" {
			t.Errorf("Expected PublicHost to be 'example.com', got '%s'", ps.PublicHost)
		}

		if ps.PublicPort != 443 {
			t.Errorf("Expected PublicPort to be 443, got %d", ps.PublicPort)
		}

		if ps.PublicProtocol != "https" {
			t.Errorf("Expected PublicProtocol to be 'https', got '%s'", ps.PublicProtocol)
		}
	})
}

func TestSvcConf_Fields(t *testing.T) {
	t.Run("service config fields are accessible", func(t *testing.T) {
		sc := SvcConf{
			Protocol: "http",
			Host:     "localhost",
			Port:     8080,
		}

		if sc.Protocol != "http" {
			t.Errorf("Expected Protocol to be 'http', got '%s'", sc.Protocol)
		}

		if sc.Host != "localhost" {
			t.Errorf("Expected Host to be 'localhost', got '%s'", sc.Host)
		}

		if sc.Port != 8080 {
			t.Errorf("Expected Port to be 8080, got %d", sc.Port)
		}
	})
}
