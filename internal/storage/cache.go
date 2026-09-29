package storage

import (
	"slices"
	"sync"
)

// Cache represents a generic, thread-safe in-memory cache that preserves insertion order.
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
	order []K
}

// NewCache initializes and returns a new Cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
		order: make([]K, 0),
	}
}

// Set adds or updates a key-value pair in the cache while tracking insertion order.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.items[key]; !exists {
		c.order = append(c.order, key)
	}
	c.items[key] = value
}

// Get retrieves a key-value pair from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.items[key]
	return value, exists
}

// AllValues returns a shallow copy list of all values in their exact insertion order.
func (c *Cache[K, V]) AllValues() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.order))
	for _, key := range c.order {
		values = append(values, c.items[key])
	}
	return values
}

func (c *Cache[K, V]) Del(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	// Stops scanning immediately once the unique key is found
	if idx := slices.Index(c.order, key); idx != -1 {
		c.order = slices.Delete(c.order, idx, idx+1)
	}

	//c.order = slices.DeleteFunc(c.order, func(e K) bool {
	//return e == key
	//})
}

// Len returns the current size of the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.order)
}
