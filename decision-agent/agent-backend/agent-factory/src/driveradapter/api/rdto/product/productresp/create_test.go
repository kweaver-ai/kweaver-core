package productresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRes_StructFields(t *testing.T) {
	t.Parallel()

	create := &CreateRes{
		Key: "test-product",
		ID:  12345,
	}

	assert.Equal(t, "test-product", create.Key)
	assert.Equal(t, int64(12345), create.ID)
}

func TestCreateRes_Empty(t *testing.T) {
	t.Parallel()

	create := &CreateRes{}

	assert.Empty(t, create.Key)
	assert.Equal(t, int64(0), create.ID)
}

func TestCreateRes_WithAllFields(t *testing.T) {
	t.Parallel()

	create := &CreateRes{
		Key: "smart-customer-service",
		ID:  67890,
	}

	assert.Equal(t, "smart-customer-service", create.Key)
	assert.Equal(t, int64(67890), create.ID)
}
