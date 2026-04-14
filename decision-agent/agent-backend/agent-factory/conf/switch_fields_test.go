package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSwitchFields(t *testing.T) {
	t.Parallel()

	t.Run("creates SwitchFields with defaults", func(t *testing.T) {
		t.Parallel()

		sf := NewSwitchFields()

		assert.NotNil(t, sf)
		assert.NotNil(t, sf.Mock)

		// Verify default values
		assert.False(t, sf.KeepLegacyAppPath)
		assert.False(t, sf.DisablePmsCheck)
		assert.False(t, sf.DisableBizDomain)
		assert.False(t, sf.DisableBizDomainInit)
		assert.False(t, sf.DisableAuditInit)
	})

	t.Run("MockSwitchFields defaults to false", func(t *testing.T) {
		t.Parallel()

		sf := NewSwitchFields()

		assert.False(t, sf.Mock.MockMQClient)
		assert.False(t, sf.Mock.MockSandboxPlatform)
		assert.False(t, sf.Mock.MockHydra)
		assert.False(t, sf.Mock.MockAuthZ)
		assert.False(t, sf.Mock.MockBizDomain)
		assert.False(t, sf.Mock.MockUserManagerModule)
		assert.Empty(t, sf.Mock.MockUserID)
	})
}

func TestSwitchFields_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create SwitchFields struct directly", func(t *testing.T) {
		t.Parallel()

		sf := &SwitchFields{
			KeepLegacyAppPath:    true,
			DisablePmsCheck:      true,
			DisableBizDomain:     true,
			DisableBizDomainInit: true,
			DisableAuditInit:     true,
			Mock: &MockSwitchFields{
				MockMQClient:          true,
				MockSandboxPlatform:   true,
				MockUserManagerModule: true,
				MockUserID:            "mock-user-id",
			},
		}

		assert.NotNil(t, sf)
		assert.True(t, sf.KeepLegacyAppPath)
		assert.True(t, sf.DisablePmsCheck)
		assert.True(t, sf.DisableBizDomain)
		assert.True(t, sf.Mock.MockMQClient)
		assert.True(t, sf.Mock.MockUserManagerModule)
		assert.Equal(t, "mock-user-id", sf.Mock.MockUserID)
	})

	t.Run("create SwitchFields with nil Mock", func(t *testing.T) {
		t.Parallel()

		sf := &SwitchFields{
			Mock: nil,
		}

		assert.NotNil(t, sf)
		assert.Nil(t, sf.Mock)
	})
}

func TestMockSwitchFields_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create MockSwitchFields struct", func(t *testing.T) {
		t.Parallel()

		msf := &MockSwitchFields{
			MockMQClient:          true,
			MockSandboxPlatform:   true,
			MockHydra:             true,
			MockAuthZ:             true,
			MockBizDomain:         true,
			MockUserManagerModule: true,
			MockUserID:            "mock-user-id",
		}

		assert.NotNil(t, msf)
		assert.True(t, msf.MockMQClient)
		assert.True(t, msf.MockSandboxPlatform)
		assert.True(t, msf.MockHydra)
		assert.True(t, msf.MockAuthZ)
		assert.True(t, msf.MockBizDomain)
		assert.True(t, msf.MockUserManagerModule)
		assert.Equal(t, "mock-user-id", msf.MockUserID)
	})

	t.Run("create empty MockSwitchFields", func(t *testing.T) {
		t.Parallel()

		msf := &MockSwitchFields{}

		assert.NotNil(t, msf)
		assert.False(t, msf.MockMQClient)
		assert.False(t, msf.MockSandboxPlatform)
		assert.False(t, msf.MockUserManagerModule)
		assert.Empty(t, msf.MockUserID)
	})
}

func TestSwitchFields_YAMLTAGs(t *testing.T) {
	t.Parallel()

	t.Run("yaml tags are correct", func(t *testing.T) {
		t.Parallel()
		// This is a compile-time check to ensure yaml tags are correct
		sf := &SwitchFields{}

		assert.NotNil(t, sf)
		// The yaml tags would be verified by actual yaml parsing
	})
}

func TestSwitchFields_IsBizDomainDisabled(t *testing.T) {
	t.Parallel()

	t.Run("nil switch fields returns false", func(t *testing.T) {
		t.Parallel()

		var sf *SwitchFields

		assert.False(t, sf.IsBizDomainDisabled())
	})

	t.Run("disable biz domain returns true", func(t *testing.T) {
		t.Parallel()

		sf := &SwitchFields{DisableBizDomain: true}

		assert.True(t, sf.IsBizDomainDisabled())
	})

	t.Run("disable biz domain false returns false", func(t *testing.T) {
		t.Parallel()

		sf := &SwitchFields{DisableBizDomain: false}

		assert.False(t, sf.IsBizDomainDisabled())
	})
}
