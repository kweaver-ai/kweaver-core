package producteo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestProduct_Embedded(t *testing.T) {
	t.Parallel()

	po := &dapo.ProductPo{
		ID:   123,
		Name: "Test Product",
		Key:  "test-product",
	}

	product := &Product{
		ProductPo: *po,
	}

	assert.Equal(t, int64(123), product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "test-product", product.Key)
}

func TestProduct_Empty(t *testing.T) {
	t.Parallel()

	product := &Product{}
	assert.Empty(t, product.ID)
	assert.Empty(t, product.Name)
}
