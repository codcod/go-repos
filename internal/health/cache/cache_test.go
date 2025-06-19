package cache

import (
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	if cache == nil {
		t.Fatal("Expected cache to be created")
	}

	if cache.items == nil {
		t.Error("Expected items map to be initialized")
	}
}

func TestMemoryCache_SetGet(t *testing.T) {
	cache := NewMemoryCache()

	key := "test-key"
	value := "test-value"
	ttl := 10 * time.Second

	// Set value
	cache.Set(key, value, ttl)

	// Get value
	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find the cached value")
	}

	if retrieved != value {
		t.Errorf("Expected '%s', got '%s'", value, retrieved)
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache()

	key := "test-key"
	value := "test-value"
	ttl := 100 * time.Millisecond

	// Set value with short TTL
	cache.Set(key, value, ttl)

	// Should be available immediately
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find the cached value immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be expired")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	key := "test-key"
	value := "test-value"
	ttl := 10 * time.Second

	// Set and verify
	cache.Set(key, value, ttl)
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find the cached value")
	}

	// Delete and verify
	cache.Delete(key)
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be deleted")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()

	// Set multiple values
	cache.Set("key1", "value1", 10*time.Second)
	cache.Set("key2", "value2", 10*time.Second)

	// Verify they exist
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	if !found1 || !found2 {
		t.Error("Expected to find both cached values")
	}

	// Clear cache
	cache.Clear()

	// Verify they're gone
	_, found1 = cache.Get("key1")
	_, found2 = cache.Get("key2")
	if found1 || found2 {
		t.Error("Expected all cached values to be cleared")
	}
}
