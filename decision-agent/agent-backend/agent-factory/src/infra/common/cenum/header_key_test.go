package cenum

import (
	"testing"
)

func TestHeaderKey_String(t *testing.T) {
	t.Parallel()

	t.Run("string conversion", func(t *testing.T) {
		t.Parallel()

		key := HeaderKey("test-key")
		result := key.String()

		if result != "test-key" {
			t.Errorf("Expected 'test-key', got '%s'", result)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		t.Parallel()

		key := HeaderKey("")
		result := key.String()

		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

func TestHeaderKey_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid HeaderXAccountType", func(t *testing.T) {
		t.Parallel()

		err := HeaderXAccountType.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for HeaderXAccountType, got %v", err)
		}
	})

	t.Run("valid HeaderXAccountTypeOld", func(t *testing.T) {
		t.Parallel()

		err := HeaderXAccountTypeOld.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for HeaderXAccountTypeOld, got %v", err)
		}
	})

	t.Run("valid HeaderXAccountID", func(t *testing.T) {
		t.Parallel()

		err := HeaderXAccountID.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for HeaderXAccountID, got %v", err)
		}
	})

	t.Run("valid HeaderXAccountIDOld", func(t *testing.T) {
		t.Parallel()

		err := HeaderXAccountIDOld.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for HeaderXAccountIDOld, got %v", err)
		}
	})

	t.Run("valid HeaderXBizDomainID", func(t *testing.T) {
		t.Parallel()

		err := HeaderXBizDomainID.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for HeaderXBizDomainID, got %v", err)
		}
	})

	t.Run("invalid header key", func(t *testing.T) {
		t.Parallel()

		key := HeaderKey("invalid-key")
		err := key.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid header key")
		}
	})

	t.Run("empty header key", func(t *testing.T) {
		t.Parallel()

		key := HeaderKey("")
		err := key.EnumCheck()

		if err == nil {
			t.Error("Expected error for empty header key")
		}
	})
}

func TestHeaderKey_Constants(t *testing.T) {
	t.Parallel()

	t.Run("HeaderXAccountType constant", func(t *testing.T) {
		t.Parallel()

		if HeaderXAccountType != "x-account-type" {
			t.Errorf("Expected HeaderXAccountType to be 'x-account-type', got '%s'", HeaderXAccountType)
		}
	})

	t.Run("HeaderXAccountTypeOld constant", func(t *testing.T) {
		t.Parallel()

		if HeaderXAccountTypeOld != "x-visitor-type" {
			t.Errorf("Expected HeaderXAccountTypeOld to be 'x-visitor-type', got '%s'", HeaderXAccountTypeOld)
		}
	})

	t.Run("HeaderXAccountID constant", func(t *testing.T) {
		t.Parallel()

		if HeaderXAccountID != "x-account-id" {
			t.Errorf("Expected HeaderXAccountID to be 'x-account-id', got '%s'", HeaderXAccountID)
		}
	})

	t.Run("HeaderXAccountIDOld constant", func(t *testing.T) {
		t.Parallel()

		if HeaderXAccountIDOld != "x-user" {
			t.Errorf("Expected HeaderXAccountIDOld to be 'x-user', got '%s'", HeaderXAccountIDOld)
		}
	})

	t.Run("HeaderXBizDomainID constant", func(t *testing.T) {
		t.Parallel()

		if HeaderXBizDomainID != "x-business-domain" {
			t.Errorf("Expected HeaderXBizDomainID to be 'x-business-domain', got '%s'", HeaderXBizDomainID)
		}
	})
}
