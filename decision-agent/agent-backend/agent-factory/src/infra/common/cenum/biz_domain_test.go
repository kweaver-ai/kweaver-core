package cenum

import (
	"testing"
)

func TestBizDomainID_ToString(t *testing.T) {
	t.Parallel()

	t.Run("valid business domain", func(t *testing.T) {
		t.Parallel()

		domain := BizDomainID("test-domain")
		result := domain.ToString()

		if result != "test-domain" {
			t.Errorf("Expected 'test-domain', got '%s'", result)
		}
	})

	t.Run("empty business domain", func(t *testing.T) {
		t.Parallel()

		domain := BizDomainID("")
		result := domain.ToString()

		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

func TestBizDomainID_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid public domain", func(t *testing.T) {
		t.Parallel()

		err := BizDomainPublic.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for BizDomainPublic, got %v", err)
		}
	})

	t.Run("invalid domain", func(t *testing.T) {
		t.Parallel()

		domain := BizDomainID("invalid-domain")
		err := domain.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid domain")
		}
	})

	t.Run("empty domain", func(t *testing.T) {
		t.Parallel()

		domain := BizDomainID("")
		err := domain.EnumCheck()

		if err == nil {
			t.Error("Expected error for empty domain")
		}
	})
}

func TestBizDomainID_Constant(t *testing.T) {
	t.Parallel()

	t.Run("BizDomainPublic constant", func(t *testing.T) {
		t.Parallel()

		if BizDomainPublic != "bd_public" {
			t.Errorf("Expected BizDomainPublic to be 'bd_public', got '%s'", BizDomainPublic)
		}
	})
}
