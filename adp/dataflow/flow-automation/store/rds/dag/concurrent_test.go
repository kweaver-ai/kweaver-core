package dagmodel

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestConcurrentCreateDag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}

	d := NewDagRepository().(*dag)
	ctx := context.Background()

	var wg sync.WaitGroup
	errors := make([]error, 10)
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			dag := &entity.Dag{
				Name: fmt.Sprintf("Concurrent Test DAG %d %d", time.Now().UnixNano(), idx),
				Vars: entity.DagVars{
					fmt.Sprintf("var_%d", idx): {DefaultValue: "val"},
				},
			}
			dag.Initial()

			_, err := d.CreateDag(ctx, dag)
			mu.Lock()
			errors[idx] = err
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Logf("Goroutine %d error: %v", i, err)
			assert.NotContains(t, err.Error(), "Deadlock", "Deadlock detected!")
		}
	}
}
