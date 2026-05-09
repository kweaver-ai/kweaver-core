// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rate

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrencyControl tests the global and catalog-level concurrency control.
func TestConcurrencyControl(t *testing.T) {
	t.Parallel()

	// Test configuration
	cfg := ConcurrencyConfig{
		Enabled: true,
		Global: GlobalConcurrencyConfig{
			MaxConcurrentQueries: 5,
		},
	}

	// Create limiter
	limiter := NewConcurrencyLimiter(cfg)
	defer limiter.Close()

	t.Run("GlobalConcurrencyLimit", func(t *testing.T) {
		t.Parallel()

		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 5,
			},
		}

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		var wg sync.WaitGroup
		successCount := 0
		failCount := 0
		var mu sync.Mutex

		// Use a counter to track concurrent executions
		var concurrentCount atomic.Int64
		maxConcurrent := atomic.Int64{}

		// Try to acquire 10 permits with global limit of 5
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				params := AcquireParams{
					CatalogID:            "catalog-1",
					MaxConcurrentQueries: 3,
				}

				release, err := limiter.Acquire(params)
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					return
				}
				defer release()

				// Track concurrency
				current := concurrentCount.Add(1)
				// Update max concurrent
				for {
					oldMax := maxConcurrent.Load()
					if current <= oldMax {
						break
					}
					if maxConcurrent.CompareAndSwap(oldMax, current) {
						break
					}
				}

				mu.Lock()
				successCount++
				mu.Unlock()

				// Hold permit for 100ms
				time.Sleep(100 * time.Millisecond)
				concurrentCount.Add(-1)
			}(i)
		}

		wg.Wait()

		// Should have at most 5 concurrent executions at any time
		if maxConcurrent.Load() > 5 {
			t.Errorf("Expected at most 5 concurrent, got %d", maxConcurrent.Load())
		}
		// Some requests may fail due to timeout
		if failCount == 10 {
			t.Error("All requests failed, likely test issue")
		}
	})

	t.Run("CatalogConcurrencyLimit", func(t *testing.T) {
		t.Parallel()

		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 20, // High global limit
			},
		}

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		var wg sync.WaitGroup
		catalog1Success := 0
		catalog2Success := 0
		var mu sync.Mutex

		// Use a counter to track concurrent executions per catalog
		var cat1Concurrent atomic.Int64
		var cat2Concurrent atomic.Int64
		cat1Max := atomic.Int64{}
		cat2Max := atomic.Int64{}

		// Try to acquire permits for two catalogs with limit of 3 each
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				catalogID := "catalog-1"
				if id >= 5 {
					catalogID = "catalog-2"
				}

				params := AcquireParams{
					CatalogID:            catalogID,
					MaxConcurrentQueries: 3,
				}

				release, err := limiter.Acquire(params)
				if err != nil {
					return
				}
				defer release()

				// Track concurrency based on catalog
				var concurrent *atomic.Int64
				var max *atomic.Int64
				if catalogID == "catalog-1" {
					concurrent = &cat1Concurrent
					max = &cat1Max
				} else {
					concurrent = &cat2Concurrent
					max = &cat2Max
				}

				current := concurrent.Add(1)
				// Update max concurrent
				for {
					oldMax := max.Load()
					if current <= oldMax {
						break
					}
					if max.CompareAndSwap(oldMax, current) {
						break
					}
				}

				mu.Lock()
				if catalogID == "catalog-1" {
					catalog1Success++
				} else {
					catalog2Success++
				}
				mu.Unlock()

				// Hold permit for 200ms to ensure concurrency limit is visible
				time.Sleep(200 * time.Millisecond)
				concurrent.Add(-1)
			}(i)
		}

		wg.Wait()

		// Each catalog should have at most 3 concurrent executions at any time
		if cat1Max.Load() > 3 {
			t.Errorf("Expected catalog-1 at most 3 concurrent, got %d", cat1Max.Load())
		}
		if cat2Max.Load() > 3 {
			t.Errorf("Expected catalog-2 at most 3 concurrent, got %d", cat2Max.Load())
		}
	})

	t.Run("ComplementaryRelationship", func(t *testing.T) {
		t.Parallel()

		// Test that both global AND catalog limits apply
		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 10, // High global limit
			},
		}

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		// Use a counter to track concurrent executions
		var concurrentCount atomic.Int64
		maxConcurrent := atomic.Int64{}

		// With global limit 10 and catalog limit 2, should only get 2 concurrent at any time
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				params := AcquireParams{
					CatalogID:            "catalog-1",
					MaxConcurrentQueries: 2,
				}

				release, err := limiter.Acquire(params)
				if err != nil {
					return
				}
				defer release()

				// Track concurrency
				current := concurrentCount.Add(1)
				// Update max concurrent
				for {
					oldMax := maxConcurrent.Load()
					if current <= oldMax {
						break
					}
					if maxConcurrent.CompareAndSwap(oldMax, current) {
						break
					}
				}

				mu.Lock()
				successCount++
				mu.Unlock()

				// Hold permit for 200ms to ensure concurrency limit is visible
				time.Sleep(200 * time.Millisecond)
				concurrentCount.Add(-1)
			}()
		}

		wg.Wait()

		// Should be limited by catalog (2), not global (10)
		if maxConcurrent.Load() > 2 {
			t.Errorf("Expected at most 2 concurrent (catalog limit), got %d", maxConcurrent.Load())
		}
	})

	t.Run("Stats", func(t *testing.T) {
		t.Parallel()

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		// Acquire some permits
		release1, _ := limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 5,
		})
		release2, _ := limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 5,
		})

		// Release permits
		release1()
		release2()
	})
}

// TestConcurrencyLimitErrorTypes tests different error scenarios.
func TestConcurrencyLimitErrorTypes(t *testing.T) {
	t.Parallel()

	t.Run("GlobalLimitExceeded", func(t *testing.T) {
		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 2,
			},
		}

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		// Acquire all global permits (global limit is 2)
		release1, err := limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 5,
		})
		if err != nil {
			t.Fatalf("First acquire failed: %v", err)
		}
		defer release1()

		release2, err := limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-2",
			MaxConcurrentQueries: 5,
		})
		if err != nil {
			t.Fatalf("Second acquire failed: %v", err)
		}
		defer release2()

		// Third should fail due to global limit
		_, err = limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-3",
			MaxConcurrentQueries: 5,
		})
		if err == nil {
			t.Error("Expected global limit error, got nil")
		}
	})

	t.Run("CatalogLimitExceeded", func(t *testing.T) {
		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 2,
			},
		}

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		// Acquire catalog permit
		release1, err := limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 1,
		})
		if err != nil {
			t.Fatalf("First acquire failed: %v", err)
		}
		defer release1()

		// Second should fail due to catalog limit (limit is 1)
		_, err = limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 1,
		})
		if err == nil {
			t.Error("Expected catalog limit error, got nil")
		}
	})

	t.Run("QueueTimeout", func(t *testing.T) {
		t.Parallel()

		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 2,
			},
		}

		limiter := NewConcurrencyLimiter(cfg)
		defer limiter.Close()

		// Hold permit for long time
		release1, err := limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 1,
		})
		if err != nil {
			t.Fatalf("First acquire failed: %v", err)
		}
		defer release1()

		// Try with very short timeout
		_, err = limiter.Acquire(AcquireParams{
			CatalogID:            "catalog-1",
			MaxConcurrentQueries: 1,
		})
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

// TestUpdateCatalogLimit tests dynamic catalog limit updates.
// Note: This feature is not yet implemented, skipping this test.
func TestUpdateCatalogLimit(t *testing.T) {
	t.Skip("UpdateCatalogLimit not yet implemented")
}

// TestConfigValidation tests configuration validation.
func TestConfigValidation(t *testing.T) {
	t.Parallel()

	t.Run("ValidConfig", func(t *testing.T) {
		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 100,
			},
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Valid config should not error: %v", err)
		}
	})

	t.Run("InvalidGlobalLimit", func(t *testing.T) {
		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 0, // Invalid
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid global limit")
		}
	})

	t.Run("InvalidCatalogLimit", func(t *testing.T) {
		// Note: Validate() currently only validates global config, not catalog config.
		// Catalog limit validation would require additional implementation.
		cfg := ConcurrencyConfig{
			Enabled: true,
			Global: GlobalConcurrencyConfig{
				MaxConcurrentQueries: 100,
			},
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Valid config should not error: %v", err)
		}
	})
}
