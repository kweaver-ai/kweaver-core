package categorysvc

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorySvc_StructFields(t *testing.T) {
	t.Parallel()

	svc := NewCategorySvc()

	// Type assertion to access internal fields
	impl, ok := svc.(*categorySvc)
	assert.True(t, ok, "Service should be *categorySvc type")
	assert.NotNil(t, impl.SvcBase)
	assert.NotNil(t, impl.categoryRepo)
}

func TestCategorySvc_ConcurrentCreation(t *testing.T) {
	t.Parallel()

	done := make(chan bool)

	// Create multiple goroutines that all call NewCategorySvc
	for i := 0; i < 10; i++ {
		go func() {
			svc := NewCategorySvc()
			assert.NotNil(t, svc)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// All should return the same instance
	svc1 := NewCategorySvc()
	svc2 := NewCategorySvc()
	assert.Same(t, svc1, svc2)
}

func TestCategorySvc_IsSingleton(t *testing.T) {
	t.Parallel()

	// Verify singleton pattern works correctly
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			svc := NewCategorySvc()
			assert.NotNil(t, svc)
		}()
	}

	wg.Wait()

	// All instances should be the same
	svc1 := NewCategorySvc()
	svc2 := NewCategorySvc()
	assert.Same(t, svc1, svc2)
}
