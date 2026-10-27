package binarytable

import (
	"container/list"
	"sync"
)

// Cache is a small thread-safe LRU cache. It is used by DeflateReader to keep
// recently decompressed blocks and is also useful to callers.
//
// The key type must be comparable; the value type may be any type supported by
// TinyGo.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	order    *list.List
	items    map[K]*list.Element
}

type cacheEntry[K comparable, V any] struct {
	key   K
	value V
}

// NewCache creates an LRU cache with the given capacity. A capacity <= 0
// disables caching.
func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	hint := capacity
	if hint < 0 {
		hint = 0
	}
	return &Cache[K, V]{
		capacity: capacity,
		order:    list.New(),
		items:    make(map[K]*list.Element, hint),
	}
}

// Capacity returns the configured capacity.
func (c *Cache[K, V]) Capacity() int { return c.capacity }

// Len returns the number of cached entries.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// Get returns the value for key and marks it as recently used.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero V
	element, ok := c.items[key]
	if !ok {
		return zero, false
	}
	c.order.MoveToFront(element)
	return element.Value.(*cacheEntry[K, V]).value, true
}

// Put inserts or replaces a value. Inserting into a zero-capacity cache is a
// no-op.
func (c *Cache[K, V]) Put(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.items[key]; ok {
		entry := element.Value.(*cacheEntry[K, V])
		entry.value = value
		c.order.MoveToFront(element)
		return
	}
	element := c.order.PushFront(&cacheEntry[K, V]{key: key, value: value})
	c.items[key] = element
	for c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest == nil {
			break
		}
		entry := oldest.Value.(*cacheEntry[K, V])
		c.order.Remove(oldest)
		delete(c.items, entry.key)
	}
}

// Contains reports whether key is present without changing recency.
func (c *Cache[K, V]) Contains(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.items[key]
	return ok
}

// Remove deletes key from the cache.
func (c *Cache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.items[key]; ok {
		c.order.Remove(element)
		delete(c.items, key)
	}
}

// Clear removes every entry.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.order.Init()
	hint := c.capacity
	if hint < 0 {
		hint = 0
	}
	c.items = make(map[K]*list.Element, hint)
}
