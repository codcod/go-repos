package cache

import (
	"sync"
	"time"
)

// MemoryCache implements core.Cache using in-memory storage
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	value     interface{}
	expiredAt time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.expiredAt) {
		// Remove expired item (will be cleaned up by background process)
		return nil, false
	}

	return item.value, true
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiredAt := time.Now().Add(ttl)
	if ttl <= 0 {
		// No expiration
		expiredAt = time.Now().Add(24 * time.Hour * 365) // 1 year
	}

	c.items[key] = &cacheItem{
		value:     value,
		expiredAt: expiredAt,
	}
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all values from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}

// cleanup removes expired items periodically
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiredAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size returns the number of items in the cache
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// NoOpCache implements core.Cache as a no-operation cache (for testing)
type NoOpCache struct{}

// NewNoOpCache creates a new no-operation cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns false (cache miss)
func (c *NoOpCache) Get(key string) (interface{}, bool) {
	return nil, false
}

// Set does nothing
func (c *NoOpCache) Set(key string, value interface{}, ttl time.Duration) {
	// No-op
}

// Delete does nothing
func (c *NoOpCache) Delete(key string) {
	// No-op
}

// Clear does nothing
func (c *NoOpCache) Clear() {
	// No-op
}
