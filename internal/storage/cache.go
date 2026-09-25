package storage

import (
	"sync"
)

// Cache represents a generic, thread-safe in-memory cache.
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// NewCache initializes and returns a new Cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
	}
}

// Set adds or updates a key-value pair in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Get a key-value pair in the cache.
func (c *Cache[K, V]) Get(key K) V {
	c.mu.Lock()
	defer c.mu.Unlock()
	value := c.items[key]
	return value
}

// Items returns a shallow copy of all elements stored in the cache.
func (c *Cache[K, V]) Items() map[K]V {
	// 1. Acquire a Read Lock to allow multiple parallel reads but block writes
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 2. Allocate a new map to prevent external mutations from affecting internal state
	results := make(map[K]V, len(c.items))

	// 3. Copy elements from the internal cache map to the result map
	for k, v := range c.items {
		results[k] = v
	}

	return results
}
