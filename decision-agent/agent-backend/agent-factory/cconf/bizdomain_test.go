package cconf

import (
	"testing"
)

func TestBizDomainConf_Fields(t *testing.T) {
	t.Run("business domain config fields are accessible", func(t *testing.T) {
		bd := BizDomainConf{
			PrivateSvc: &BizDomainSvcConf{
				Host:     "bizdomain.local",
				Port:     7000,
				Protocol: "grpc",
			},
		}

		if bd.PrivateSvc == nil {
			t.Fatal("Expected PrivateSvc to be non-nil")
		}

		if bd.PrivateSvc.Host != "bizdomain.local" {
			t.Errorf("Expected PrivateSvc.Host to be 'bizdomain.local', got '%s'", bd.PrivateSvc.Host)
		}

		if bd.PrivateSvc.Protocol != "grpc" {
			t.Errorf("Expected PrivateSvc.Protocol to be 'grpc', got '%s'", bd.PrivateSvc.Protocol)
		}
	})
}

func TestBizDomainSvcConf_Fields(t *testing.T) {
	t.Run("business domain service config fields are accessible", func(t *testing.T) {
		bds := BizDomainSvcConf{
			Host:     "localhost",
			Port:     7000,
			Protocol: "http",
		}

		if bds.Host != "localhost" {
			t.Errorf("Expected Host to be 'localhost', got '%s'", bds.Host)
		}

		if bds.Port != 7000 {
			t.Errorf("Expected Port to be 7000, got %d", bds.Port)
		}
	})
}

func TestBizDomainConf_NilPrivateSvc(t *testing.T) {
	t.Run("business domain config with nil private service", func(t *testing.T) {
		bd := BizDomainConf{
			PrivateSvc: nil,
		}

		if bd.PrivateSvc != nil {
			t.Error("Expected PrivateSvc to be nil")
		}
	})
}
