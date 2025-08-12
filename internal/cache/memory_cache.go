package cache

import (
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a cached exchange rate with expiration
type CacheItem struct {
	Rate      float64
	ExpiresAt time.Time
}

// MemoryCache implements an in-memory cache for exchange rates
type MemoryCache struct {
	data map[string]CacheItem
	mu   sync.RWMutex
	ttl  time.Duration
}

// NewMemoryCache creates a new in-memory cache with specified TTL
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]CacheItem),
		ttl:  ttl,
	}

	// Start background cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// generateKey creates a cache key for a currency pair and date
func (c *MemoryCache) generateKey(from, to, date string) string {
	if date == "" {
		return fmt.Sprintf("%s_%s_latest", from, to)
	}
	return fmt.Sprintf("%s_%s_%s", from, to, date)
}

// Get retrieves an exchange rate from cache
func (c *MemoryCache) Get(from, to, date string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(from, to, date)
	item, exists := c.data[key]

	if !exists {
		return 0, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		// Item expired, but we'll let the cleanup goroutine handle removal
		return 0, false
	}

	return item.Rate, true
}

// Set stores an exchange rate in cache
func (c *MemoryCache) Set(from, to, date string, rate float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(from, to, date)
	c.data[key] = CacheItem{
		Rate:      rate,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes an exchange rate from cache
func (c *MemoryCache) Delete(from, to, date string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(from, to, date)
	delete(c.data, key)
}

// Clear removes all items from cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]CacheItem)
}

// Size returns the number of items in cache
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// GetStats returns cache statistics
func (c *MemoryCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	validItems := 0
	expiredItems := 0
	now := time.Now()

	for _, item := range c.data {
		if now.After(item.ExpiresAt) {
			expiredItems++
		} else {
			validItems++
		}
	}

	return map[string]interface{}{
		"total_items":   len(c.data),
		"valid_items":   validItems,
		"expired_items": expiredItems,
		"ttl_seconds":   c.ttl.Seconds(),
	}
}

// cleanupExpired removes expired items from cache periodically
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()

			for key, item := range c.data {
				if now.After(item.ExpiresAt) {
					delete(c.data, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

// CacheInterface defines the interface for caching operations
type CacheInterface interface {
	Get(from, to, date string) (float64, bool)
	Set(from, to, date string, rate float64)
	Delete(from, to, date string)
	Clear()
	Size() int
	GetStats() map[string]interface{}
}
