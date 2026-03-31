// Package cache defines the cache interface.
package cache

import (
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// InMemoryCache is a simple in-memory cache implementation.
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

// NewInMemoryCache creates a new instance of InMemoryCache.
func NewInMemoryCache() interfaces.Cache {
	return &InMemoryCache{
		items: make(map[string]interface{}),
	}
}

// Set stores a value in the cache with the given key.
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Get retrieves a value from the cache by key.
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

// Delete removes a value from the cache by key.
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Size returns the number of items in the cache.
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
