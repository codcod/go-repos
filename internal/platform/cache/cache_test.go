package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestMemoryCache_NewMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	if cache == nil {
		t.Fatal("NewMemoryCache() returned nil")
	}

	if cache.items == nil {
		t.Fatal("cache items map is nil")
	}

	if cache.Size() != 0 {
		t.Errorf("Expected empty cache, got size %d", cache.Size())
	}
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	// Test setting and getting a value
	key := "test-key"
	value := "test-value"
	ttl := 10 * time.Second

	cache.Set(key, value, ttl)

	retrievedValue, exists := cache.Get(key)
	if !exists {
		t.Error("Expected value to exist in cache")
	}

	if retrievedValue != value {
		t.Errorf("Expected %v, got %v", value, retrievedValue)
	}
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache()

	_, exists := cache.Get("non-existent-key")
	if exists {
		t.Error("Expected false for non-existent key")
	}
}

func TestMemoryCache_ExpiredItems(t *testing.T) {
	cache := NewMemoryCache()

	key := "expired-key"
	value := "expired-value"
	ttl := 100 * time.Millisecond

	cache.Set(key, value, ttl)

	// Verify item exists immediately
	_, exists := cache.Get(key)
	if !exists {
		t.Error("Expected item to exist immediately after setting")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify item is expired
	_, exists = cache.Get(key)
	if exists {
		t.Error("Expected item to be expired")
	}
}

func TestMemoryCache_ZeroTTL(t *testing.T) {
	cache := NewMemoryCache()

	key := "no-ttl-key"
	value := "no-ttl-value"

	cache.Set(key, value, 0)

	retrievedValue, exists := cache.Get(key)
	if !exists {
		t.Error("Expected value to exist with zero TTL")
	}

	if retrievedValue != value {
		t.Errorf("Expected %v, got %v", value, retrievedValue)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	key := "delete-key"
	value := "delete-value"
	ttl := 10 * time.Second

	cache.Set(key, value, ttl)

	// Verify item exists
	_, exists := cache.Get(key)
	if !exists {
		t.Error("Expected item to exist before deletion")
	}

	// Delete the item
	cache.Delete(key)

	// Verify item is gone
	_, exists = cache.Get(key)
	if exists {
		t.Error("Expected item to be deleted")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()

	// Add multiple items
	cache.Set("key1", "value1", 10*time.Second)
	cache.Set("key2", "value2", 10*time.Second)
	cache.Set("key3", "value3", 10*time.Second)

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	// Clear all items
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// Verify items are gone
	_, exists := cache.Get("key1")
	if exists {
		t.Error("Expected key1 to be cleared")
	}
}

func TestMemoryCache_Size(t *testing.T) {
	cache := NewMemoryCache()

	// Test empty cache
	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}

	// Add items
	cache.Set("key1", "value1", 10*time.Second)
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	cache.Set("key2", "value2", 10*time.Second)
	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Overwrite existing item
	cache.Set("key1", "new-value1", 10*time.Second)
	if cache.Size() != 2 {
		t.Errorf("Expected size to remain 2 after overwrite, got %d", cache.Size())
	}
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := fmt.Sprintf("key-%d", n)
			value := fmt.Sprintf("value-%d", n)
			cache.Set(key, value, 10*time.Second)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if cache.Size() != 10 {
		t.Errorf("Expected size 10 after concurrent writes, got %d", cache.Size())
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := fmt.Sprintf("key-%d", n)
			expectedValue := fmt.Sprintf("value-%d", n)
			value, exists := cache.Get(key)
			if !exists || value != expectedValue {
				t.Errorf("Concurrent read failed for key %s", key)
			}
			done <- true
		}(i)
	}

	// Wait for all read goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
