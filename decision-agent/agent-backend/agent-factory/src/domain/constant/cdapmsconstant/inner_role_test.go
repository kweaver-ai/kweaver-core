package cdapmsconstant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInnerRoleAppAdminName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "app_admin", InnerRoleAppAdminName)
}

func TestInnerRoleAppAdminID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "1572fb82-526f-11f0-bde6-e674ec8dde71", InnerRoleAppAdminID)
}

func TestInnerRoleConstants_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, InnerRoleAppAdminName)
	assert.NotEmpty(t, InnerRoleAppAdminID)
}

func TestInnerRoleConstants_AreUnique(t *testing.T) {
	t.Parallel()

	assert.NotEqual(t, InnerRoleAppAdminName, InnerRoleAppAdminID)
}
