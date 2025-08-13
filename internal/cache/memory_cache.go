package cache

import (
	"fmt"
	"sync"
	"time"
)

type CacheItem struct {
	Rate      float64
	ExpiresAt time.Time
}

type MemoryCache struct {
	data map[string]CacheItem
	mu   sync.RWMutex
	ttl  time.Duration
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]CacheItem),
		ttl:  ttl,
	}

	go cache.cleanupExpired()

	return cache
}

func (c *MemoryCache) generateKey(from, to, date string) string {
	if date == "" {
		return fmt.Sprintf("%s_%s_latest", from, to)
	}
	return fmt.Sprintf("%s_%s_%s", from, to, date)
}

func (c *MemoryCache) Get(from, to, date string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(from, to, date)
	item, exists := c.data[key]

	if !exists {
		return 0, false
	}

	if time.Now().After(item.ExpiresAt) {
		return 0, false
	}

	return item.Rate, true
}

func (c *MemoryCache) Set(from, to, date string, rate float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(from, to, date)
	c.data[key] = CacheItem{
		Rate:      rate,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *MemoryCache) Delete(from, to, date string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(from, to, date)
	delete(c.data, key)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]CacheItem)
}

func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

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

func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) 
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

type CacheInterface interface {
	Get(from, to, date string) (float64, bool)
	Set(from, to, date string, rate float64)
	Delete(from, to, date string)
	Clear()
	Size() int
	GetStats() map[string]interface{}
}
