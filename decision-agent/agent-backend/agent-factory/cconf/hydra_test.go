package cconf

import (
	"testing"
)

func TestHydraCfg_Fields(t *testing.T) {
	t.Run("hydra config fields are accessible", func(t *testing.T) {
		hydra := HydraCfg{
			UserMgnt: UserMgntCfg{
				Host:     "user-mgmt.local",
				Port:     8080,
				Protocol: "http",
			},
			HydraPublic: HydraPublicCfg{
				Host: "hydra-public.local",
				Port: 4444,
			},
			HydraAdmin: HydraAdminCfg{
				Host: "hydra-admin.local",
				Port: 4445,
			},
		}

		if hydra.UserMgnt.Host != "user-mgmt.local" {
			t.Errorf("Expected UserMgnt.Host to be 'user-mgmt.local', got '%s'", hydra.UserMgnt.Host)
		}

		if hydra.HydraPublic.Port != 4444 {
			t.Errorf("Expected HydraPublic.Port to be 4444, got %d", hydra.HydraPublic.Port)
		}

		if hydra.HydraAdmin.Port != 4445 {
			t.Errorf("Expected HydraAdmin.Port to be 4445, got %d", hydra.HydraAdmin.Port)
		}
	})
}

func TestUserMgntCfg_Fields(t *testing.T) {
	t.Run("user management config fields are accessible", func(t *testing.T) {
		um := UserMgntCfg{
			Host:     "localhost",
			Port:     8080,
			Protocol: "https",
		}

		if um.Protocol != "https" {
			t.Errorf("Expected Protocol to be 'https', got '%s'", um.Protocol)
		}
	})
}
