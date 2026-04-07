package productp2e

import (
	"context"
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
		po      *dapo.ProductPo
		wantErr bool
		checkEO func(t *testing.T, eo *producteo.Product)
	}{
		{
			name: "valid product PO",
			po: &dapo.ProductPo{
				ID:      1,
				Name:    "Test Product",
				Key:     "test-product",
				Profile: "Test Description",
			},
			wantErr: false,
			checkEO: func(t *testing.T, eo *producteo.Product) {
				assert.Equal(t, int64(1), eo.ProductPo.ID)
				assert.Equal(t, "Test Product", eo.ProductPo.Name)
				assert.Equal(t, "test-product", eo.ProductPo.Key)
				assert.Equal(t, "Test Description", eo.ProductPo.Profile)
			},
		},
		{
			name:    "nil PO",
			po:      nil,
			wantErr: false,
			checkEO: func(t *testing.T, eo *producteo.Product) {
				assert.Nil(t, eo)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo, err := Product(tt.po)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.checkEO != nil {
					tt.checkEO(t, eo)
				}
			}
		})
	}
}

func TestProducts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name     string
		pos      []*dapo.ProductPo
		wantErr  bool
		checkEOs func(t *testing.T, eos []*producteo.Product)
	}{
		{
			name: "multiple valid products",
			pos: []*dapo.ProductPo{
				{
					ID:   1,
					Name: "Product 1",
					Key:  "product-1",
				},
				{
					ID:   2,
					Name: "Product 2",
					Key:  "product-2",
				},
			},
			wantErr: false,
			checkEOs: func(t *testing.T, eos []*producteo.Product) {
				assert.Len(t, eos, 2)
				assert.Equal(t, int64(1), eos[0].ProductPo.ID)
				assert.Equal(t, "Product 1", eos[0].ProductPo.Name)
				assert.Equal(t, int64(2), eos[1].ProductPo.ID)
				assert.Equal(t, "Product 2", eos[1].ProductPo.Name)
			},
		},
		{
			name:    "empty slice",
			pos:     []*dapo.ProductPo{},
			wantErr: false,
			checkEOs: func(t *testing.T, eos []*producteo.Product) {
				assert.Len(t, eos, 0)
			},
		},
		{
			name:    "nil slice",
			pos:     nil,
			wantErr: false,
			checkEOs: func(t *testing.T, eos []*producteo.Product) {
				assert.Len(t, eos, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eos, err := Products(ctx, tt.pos)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.checkEOs != nil {
					tt.checkEOs(t, eos)
				}
			}
		})
	}
}
