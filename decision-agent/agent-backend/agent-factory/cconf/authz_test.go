package cconf

import (
	"testing"
)

func TestAuthzCfg_Fields(t *testing.T) {
	t.Run("authorization config with private service", func(t *testing.T) {
		authz := AuthzCfg{
			PrivateSvc: &SvcConf{
				Protocol: "http",
				Host:     "authz-private.local",
				Port:     9000,
			},
			PublicSvc: nil,
		}

		if authz.PrivateSvc == nil {
			t.Fatal("Expected PrivateSvc to be non-nil")
		}

		if authz.PrivateSvc.Host != "authz-private.local" {
			t.Errorf("Expected PrivateSvc.Host to be 'authz-private.local', got '%s'", authz.PrivateSvc.Host)
		}

		if authz.PublicSvc != nil {
			t.Error("Expected PublicSvc to be nil")
		}
	})
}

func TestAuthzCfg_WithPublicSvc(t *testing.T) {
	t.Run("authorization config with public service", func(t *testing.T) {
		authz := AuthzCfg{
			PrivateSvc: nil,
			PublicSvc: &SvcConf{
				Protocol: "https",
				Host:     "authz-public.local",
				Port:     443,
			},
		}

		if authz.PrivateSvc != nil {
			t.Error("Expected PrivateSvc to be nil")
		}

		if authz.PublicSvc == nil {
			t.Fatal("Expected PublicSvc to be non-nil")
		}

		if authz.PublicSvc.Protocol != "https" {
			t.Errorf("Expected PublicSvc.Protocol to be 'https', got '%s'", authz.PublicSvc.Protocol)
		}
	})
}
