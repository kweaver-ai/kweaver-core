// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package rate provides rate limiting and concurrency control for VEGA queries.
// This package implements a two-level concurrency control mechanism:
// 1. Global concurrency control: Protects VEGA service itself
// 2. Catalog-level concurrency control: Protects downstream data sources
//
// The two levels work in a complementary relationship (AND logic), not mutually exclusive.
// A query can only proceed if BOTH global and catalog permits are acquired.
package rate

import (
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

// ConcurrencyLimiter provides two-level concurrency control.
// Global level protects VEGA service, Catalog level protects downstream services.
type ConcurrencyLimiter interface {
	// Acquire attempts to acquire execution permits.
	// Returns a release function that MUST be called when query completes.
	// Returns error if:
	//   - ErrGlobalLimitExceeded: Global concurrency limit reached
	//   - ErrCatalogLimitExceeded: Catalog concurrency limit reached
	Acquire(params AcquireParams) (release ReleaseFunc, err error)

	// Close releases all resources.
	Close()
}

// AcquireParams contains parameters for acquiring concurrency permits.
type AcquireParams struct {
	CatalogID            string // Catalog ID (required for catalog-level control)
	MaxConcurrentQueries int64  // Max concurrent queries for this catalog
}

// ReleaseFunc is a function that releases acquired permits.
// It MUST be called exactly once when the query completes.
type ReleaseFunc func()

// concurrencyLimiter implements ConcurrencyLimiter with two-level control.
type concurrencyLimiter struct {
	cfg ConcurrencyConfig

	// Global semaphore (protects VEGA service itself)
	globalSem *semaphore.Weighted

	// catalog semaphores (protects downstream services)
	// Key: catalogID, Value: *catalogSemaphore
	catalogSems sync.Map
}

// catalogSemaphore wraps a semaphore with catalog-specific configuration.
type catalogSemaphore struct {
	sem *semaphore.Weighted
}

// NewConcurrencyLimiter creates a new concurrency limiter with the given configuration.
func NewConcurrencyLimiter(cfg ConcurrencyConfig) ConcurrencyLimiter {
	cl := &concurrencyLimiter{
		cfg:       cfg,
		globalSem: semaphore.NewWeighted(int64(cfg.Global.MaxConcurrentQueries)),
	}

	return cl
}

// Acquire attempts to acquire both global and catalog permits.
// Both must succeed for the query to proceed (AND logic).
func (cl *concurrencyLimiter) Acquire(params AcquireParams) (ReleaseFunc, error) {
	// Try to acquire global semaphore
	if !cl.globalSem.TryAcquire(1) {
		return nil, NewRateLimitError(ErrGlobalLimitExceeded,
			fmt.Sprintf("global concurrency limit exceeded, max=%d", cl.cfg.Global.MaxConcurrentQueries))
	}

	// Acquire catalog permit (protects downstream service)
	var catalogSem *catalogSemaphore

	if params.CatalogID != "" && params.MaxConcurrentQueries > 0 {
		catalogSem = cl.getOrCreateCatalogSemaphore(params.CatalogID, params.MaxConcurrentQueries)

		if !catalogSem.sem.TryAcquire(1) {
			// Catalog limit hit: release global permits
			cl.globalSem.Release(1)

			return nil, NewRateLimitError(ErrCatalogLimitExceeded,
				fmt.Sprintf("catalog concurrency limit exceeded for catalog=%s, max=%d",
					params.CatalogID, params.MaxConcurrentQueries))
		}
	}

	// Success, return release function
	return func() {
		// Release catalog permit (if acquired)
		if catalogSem != nil {
			catalogSem.sem.Release(1)
		}

		// Release global permit
		cl.globalSem.Release(1)
	}, nil
}

// Close releases all resources.
func (cl *concurrencyLimiter) Close() {
	// No explicit cleanup needed for semaphores
	// Just clear the maps
	cl.catalogSems.Range(func(key, value interface{}) bool {
		cl.catalogSems.Delete(key)
		return true
	})
}

// getOrCreateCatalogSemaphore gets or creates a catalog semaphore.
func (cl *concurrencyLimiter) getOrCreateCatalogSemaphore(catalogID string, maxConcurrentQueries int64) *catalogSemaphore {
	if sem, ok := cl.catalogSems.Load(catalogID); ok {
		return sem.(*catalogSemaphore)
	}

	// Create new semaphore
	newSem := &catalogSemaphore{
		sem: semaphore.NewWeighted(maxConcurrentQueries),
	}

	// Use LoadOrStore to avoid race condition
	actual, loaded := cl.catalogSems.LoadOrStore(catalogID, newSem)
	if loaded {
		return actual.(*catalogSemaphore)
	}

	return newSem
}
