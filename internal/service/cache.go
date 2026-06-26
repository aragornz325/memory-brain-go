package service

import (
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	value      any
	expiration time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	metrics *Metrics
}

func NewCache(metrics *Metrics) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		metrics: metrics,
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss()
		}
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss()
		}
		return nil, false
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit()
	}
	return entry.value, true
}

func (c *Cache) Set(key string, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(duration),
	}
}

func (c *Cache) Invalidate(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k := range c.entries {
		if strings.HasPrefix(k, prefix) {
			delete(c.entries, k)
		}
	}
}

func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}
