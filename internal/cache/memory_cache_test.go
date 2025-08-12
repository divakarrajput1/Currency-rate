package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	// Test Set and Get
	cache.Set("USD", "INR", "", 83.5)
	rate, found := cache.Get("USD", "INR", "")
	assert.True(t, found)
	assert.Equal(t, 83.5, rate)

	// Test Get non-existent key
	_, found = cache.Get("EUR", "JPY", "")
	assert.False(t, found)
}

func TestMemoryCache_HistoricalRates(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	// Test historical rate
	cache.Set("USD", "INR", "2023-01-01", 82.0)
	rate, found := cache.Get("USD", "INR", "2023-01-01")
	assert.True(t, found)
	assert.Equal(t, 82.0, rate)

	// Different date should not be found
	_, found = cache.Get("USD", "INR", "2023-01-02")
	assert.False(t, found)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(100 * time.Millisecond)
	defer cache.Clear()

	// Set a rate
	cache.Set("USD", "INR", "", 83.5)

	// Should be found immediately
	_, found := cache.Get("USD", "INR", "")
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiration
	_, found = cache.Get("USD", "INR", "")
	assert.False(t, found)
}

func TestMemoryCache_KeyGeneration(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	// Test different key combinations
	cache.Set("USD", "INR", "", 83.5)
	cache.Set("USD", "INR", "2023-01-01", 82.0)
	cache.Set("INR", "USD", "", 0.012)

	// All should be distinct
	rate1, found1 := cache.Get("USD", "INR", "")
	rate2, found2 := cache.Get("USD", "INR", "2023-01-01")
	rate3, found3 := cache.Get("INR", "USD", "")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)
	assert.Equal(t, 83.5, rate1)
	assert.Equal(t, 82.0, rate2)
	assert.Equal(t, 0.012, rate3)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	// Set and verify
	cache.Set("USD", "INR", "", 83.5)
	_, found := cache.Get("USD", "INR", "")
	assert.True(t, found)

	// Delete and verify
	cache.Delete("USD", "INR", "")
	_, found = cache.Get("USD", "INR", "")
	assert.False(t, found)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)

	// Set multiple rates
	cache.Set("USD", "INR", "", 83.5)
	cache.Set("EUR", "USD", "", 1.1)
	cache.Set("GBP", "JPY", "", 150.0)

	assert.Equal(t, 3, cache.Size())

	// Clear all
	cache.Clear()
	assert.Equal(t, 0, cache.Size())

	// Verify none are found
	_, found1 := cache.Get("USD", "INR", "")
	_, found2 := cache.Get("EUR", "USD", "")
	_, found3 := cache.Get("GBP", "JPY", "")

	assert.False(t, found1)
	assert.False(t, found2)
	assert.False(t, found3)
}

func TestMemoryCache_Size(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	assert.Equal(t, 0, cache.Size())

	cache.Set("USD", "INR", "", 83.5)
	assert.Equal(t, 1, cache.Size())

	cache.Set("EUR", "USD", "", 1.1)
	assert.Equal(t, 2, cache.Size())

	cache.Delete("USD", "INR", "")
	assert.Equal(t, 1, cache.Size())
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache(100 * time.Millisecond)
	defer cache.Clear()

	// Add some rates
	cache.Set("USD", "INR", "", 83.5)
	cache.Set("EUR", "USD", "", 1.1)

	stats := cache.GetStats()
	assert.Equal(t, 2, stats["total_items"])
	assert.Equal(t, 2, stats["valid_items"])
	assert.Equal(t, 0, stats["expired_items"])
	assert.Equal(t, 0.1, stats["ttl_seconds"])

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	stats = cache.GetStats()
	assert.Equal(t, 2, stats["total_items"])
	assert.Equal(t, 0, stats["valid_items"])
	assert.Equal(t, 2, stats["expired_items"])
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Set("USD", "INR", "", float64(id*numOperations+j))
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Get("USD", "INR", "")
			}
		}()
	}

	wg.Wait()

	// Verify cache is still functional
	cache.Set("TEST", "PAIR", "", 123.45)
	rate, found := cache.Get("TEST", "PAIR", "")
	assert.True(t, found)
	assert.Equal(t, 123.45, rate)
}

func TestMemoryCache_MixedOperations(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)
	defer cache.Clear()

	var wg sync.WaitGroup
	const numGoroutines = 10

	// Mixed concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Set
			cache.Set("CURR1", "CURR2", "", float64(id))

			// Get
			cache.Get("CURR1", "CURR2", "")

			// Delete
			if id%2 == 0 {
				cache.Delete("CURR1", "CURR2", "")
			}

			// Size
			cache.Size()

			// Stats
			cache.GetStats()
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	assert.NotNil(t, cache)
}
