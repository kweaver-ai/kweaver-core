package productresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/producteo"
	"github.com/stretchr/testify/assert"
)

func TestNewDetailRes(t *testing.T) {
	t.Parallel()

	detail := NewDetailRes()

	assert.NotNil(t, detail)
	assert.IsType(t, &DetailRes{}, detail)
}

func TestDetailRes_LoadFromEo(t *testing.T) {
	t.Parallel()

	t.Run("load from product eo", func(t *testing.T) {
		t.Parallel()

		detail := NewDetailRes()
		eo := &producteo.Product{}

		// Set values in the entity
		eo.ID = 123
		eo.Name = "Test Product"
		eo.Key = "test-product"
		eo.Profile = "Test Profile"
		eo.CreatedAt = 1234567890
		eo.UpdatedAt = 1234567899

		err := detail.LoadFromEo(eo)
		assert.NoError(t, err)
		assert.Equal(t, int64(123), detail.ID)
		assert.Equal(t, "Test Product", detail.Name)
		assert.Equal(t, "test-product", detail.Key)
		assert.Equal(t, "Test Profile", detail.Profile)
		assert.Equal(t, int64(1234567890), detail.CreatedAt)
		assert.Equal(t, int64(1234567899), detail.UpdatedAt)
	})

	t.Run("with nil eo", func(t *testing.T) {
		t.Parallel()

		detail := NewDetailRes()

		assert.Panics(t, func() {
			_ = detail.LoadFromEo(nil)
		})
	})
}

func TestDetailRes_StructFields(t *testing.T) {
	t.Parallel()

	detail := &DetailRes{
		ID:        456,
		Name:      "Another Product",
		Key:       "another-key",
		Profile:   "Another Profile",
		CreatedAt: 1234567890,
		UpdatedAt: 1234567899,
	}

	assert.Equal(t, int64(456), detail.ID)
	assert.Equal(t, "Another Product", detail.Name)
	assert.Equal(t, "another-key", detail.Key)
	assert.Equal(t, "Another Profile", detail.Profile)
	assert.Equal(t, int64(1234567890), detail.CreatedAt)
	assert.Equal(t, int64(1234567899), detail.UpdatedAt)
}

func TestDetailRes_Empty(t *testing.T) {
	t.Parallel()

	detail := &DetailRes{}

	assert.Equal(t, int64(0), detail.ID)
	assert.Empty(t, detail.Name)
	assert.Empty(t, detail.Key)
	assert.Empty(t, detail.Profile)
	assert.Equal(t, int64(0), detail.CreatedAt)
	assert.Equal(t, int64(0), detail.UpdatedAt)
}

func TestDetailRes_WithAllFields(t *testing.T) {
	t.Parallel()

	detail := &DetailRes{
		ID:        789,
		Name:      "Complete Product",
		Key:       "complete-product",
		Profile:   "Complete product description for testing",
		CreatedAt: 1000000000,
		UpdatedAt: 2000000000,
	}

	assert.Equal(t, int64(789), detail.ID)
	assert.Equal(t, "Complete Product", detail.Name)
	assert.Equal(t, "complete-product", detail.Key)
	assert.Equal(t, "Complete product description for testing", detail.Profile)
	assert.Equal(t, int64(1000000000), detail.CreatedAt)
	assert.Equal(t, int64(2000000000), detail.UpdatedAt)
}
