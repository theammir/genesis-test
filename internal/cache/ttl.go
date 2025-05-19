package cache

import (
	"sync"
	"time"
)

type TTLitem[T any] struct {
	v         T
	timestamp time.Time
}

type TTLCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]TTLitem[V]
	TTL   time.Duration
}

func NewTTLCache[K comparable, V any](TTL time.Duration) *TTLCache[K, V] {
	return &TTLCache[K, V]{mu: sync.RWMutex{}, items: make(map[K]TTLitem[V]), TTL: TTL}
}

func (c *TTLCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = TTLitem[V]{v: value, timestamp: time.Now()}
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		var v V
		return v, false
	}

	if time.Since(item.timestamp) > c.TTL {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		var v V
		return v, false
	}
	return item.v, true
}
