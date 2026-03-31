package productsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProductService(t *testing.T) {
	t.Parallel()

	// First call creates the service
	svc1 := NewProductService()
	assert.NotNil(t, svc1)

	// Second call returns the same instance (singleton)
	svc2 := NewProductService()
	assert.Same(t, svc1, svc2)
}

func TestProductSvc_StructFields(t *testing.T) {
	t.Parallel()

	svc := NewProductService()

	// Type assertion to access internal fields
	impl, ok := svc.(*productSvc)
	assert.True(t, ok, "Service should be *productSvc type")
	assert.NotNil(t, impl.SvcBase)
	assert.NotNil(t, impl.redisCmp)
	assert.NotNil(t, impl.productRepo)
}

func TestProductSvc_ConcurrentCreation(t *testing.T) {
	t.Parallel()

	done := make(chan bool)

	// Create multiple goroutines that all call NewProductService
	for i := 0; i < 10; i++ {
		go func() {
			svc := NewProductService()
			assert.NotNil(t, svc)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// All should return the same instance
	svc1 := NewProductService()
	svc2 := NewProductService()
	assert.Same(t, svc1, svc2)
}
