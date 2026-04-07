package producte2p

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		eo      *producteo.Product
		wantErr bool
		checkPO func(t *testing.T, po *dapo.ProductPo)
	}{
		{
			name: "valid product entity",
			eo: &producteo.Product{
				ProductPo: dapo.ProductPo{
					ID:      1,
					Name:    "Test Product",
					Key:     "test-product",
					Profile: "Test Description",
				},
			},
			wantErr: false,
			checkPO: func(t *testing.T, po *dapo.ProductPo) {
				assert.Equal(t, int64(1), po.ID)
				assert.Equal(t, "Test Product", po.Name)
				assert.Equal(t, "test-product", po.Key)
				assert.Equal(t, "Test Description", po.Profile)
			},
		},
		{
			name: "product with minimal fields",
			eo: &producteo.Product{
				ProductPo: dapo.ProductPo{
					ID:   2,
					Name: "Minimal Product",
				},
			},
			wantErr: false,
			checkPO: func(t *testing.T, po *dapo.ProductPo) {
				assert.Equal(t, int64(2), po.ID)
				assert.Equal(t, "Minimal Product", po.Name)
			},
		},
		{
			name:    "nil entity",
			eo:      nil,
			wantErr: false,
			checkPO: func(t *testing.T, po *dapo.ProductPo) {
				assert.Nil(t, po)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			po, err := Product(tt.eo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.checkPO != nil {
					tt.checkPO(t, po)
				}
			}
		})
	}
}

func TestProduct_EmptyProduct(t *testing.T) {
	t.Parallel()

	eo := &producteo.Product{}
	po, err := Product(eo)

	require.NoError(t, err)
	assert.NotNil(t, po)
}

func TestProduct_WithAllFields(t *testing.T) {
	t.Parallel()

	eo := &producteo.Product{
		ProductPo: dapo.ProductPo{
			ID:        100,
			Name:      "Complete Product",
			Key:       "complete-product",
			Profile:   "Complete Description",
			CreatedAt: 1234567890,
			UpdatedAt: 1234567891,
			CreatedBy: "user-1",
			UpdatedBy: "user-2",
		},
	}

	po, err := Product(eo)
	require.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, int64(100), po.ID)
	assert.Equal(t, "Complete Product", po.Name)
	assert.Equal(t, "complete-product", po.Key)
	assert.Equal(t, "Complete Description", po.Profile)
	assert.Equal(t, int64(1234567890), po.CreatedAt)
	assert.Equal(t, int64(1234567891), po.UpdatedAt)
	assert.Equal(t, "user-1", po.CreatedBy)
	assert.Equal(t, "user-2", po.UpdatedBy)
}

func TestProduct_WithChineseCharacters(t *testing.T) {
	t.Parallel()

	eo := &producteo.Product{
		ProductPo: dapo.ProductPo{
			ID:      1,
			Name:    "中文产品",
			Key:     "zhongwen-chanpin",
			Profile: "这是中文描述",
		},
	}

	po, err := Product(eo)
	require.NoError(t, err)
	assert.Equal(t, "中文产品", po.Name)
	assert.Equal(t, "zhongwen-chanpin", po.Key)
	assert.Equal(t, "这是中文描述", po.Profile)
}

func TestProduct_WithZeroValues(t *testing.T) {
	t.Parallel()

	eo := &producteo.Product{
		ProductPo: dapo.ProductPo{
			ID:   0,
			Name: "",
			Key:  "",
		},
	}

	po, err := Product(eo)
	require.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, int64(0), po.ID)
	assert.Equal(t, "", po.Name)
	assert.Equal(t, "", po.Key)
}
