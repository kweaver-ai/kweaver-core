package auditconstant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditOperations_Basic(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "create", CREATE)
	assert.Equal(t, "delete", DELETE)
	assert.Equal(t, "update", UPDATE)
	assert.Equal(t, "copy", COPY)
	assert.Equal(t, "publish", PUBLISH)
	assert.Equal(t, "unpublish", UNPUBLISH)
}

func TestAuditOperations_Advanced(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "import", IMPORT)
	assert.Equal(t, "export", EXPORT)
	assert.Equal(t, "modify_publish", MODIFY_PUBLISH)
	assert.Equal(t, "copy_publish", COPY_PUBLISH)
}

func TestAuditOperations_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, CREATE)
	assert.NotEmpty(t, DELETE)
	assert.NotEmpty(t, UPDATE)
	assert.NotEmpty(t, PUBLISH)
	assert.NotEmpty(t, UNPUBLISH)
}

func TestAuditOperations_AreUnique(t *testing.T) {
	t.Parallel()

	operations := []string{
		CREATE, DELETE, UPDATE, COPY, PUBLISH, UNPUBLISH,
		IMPORT, EXPORT, MODIFY_PUBLISH, COPY_PUBLISH,
	}

	uniqueOps := make(map[string]bool)
	for _, op := range operations {
		assert.False(t, uniqueOps[op], "Duplicate operation found: %s", op)
		uniqueOps[op] = true
	}
}

func TestAuditOperation_ValidValues(t *testing.T) {
	t.Parallel()

	validOperations := []string{
		CREATE, DELETE, UPDATE, COPY, PUBLISH, UNPUBLISH,
		IMPORT, EXPORT, MODIFY_PUBLISH, COPY_PUBLISH,
	}

	for _, op := range validOperations {
		assert.NotEmpty(t, op)
	}
}
