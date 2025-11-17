package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// Cache provides caching for action outputs
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
}

// CacheEntry represents a cached action result
type CacheEntry struct {
	Key       string
	Value     map[string]interface{}
	CreatedAt time.Time
	TTL       time.Duration
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	cache := &Cache{
		entries: make(map[string]*CacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// generateKey creates a cache key from action name and inputs
func (c *Cache) generateKey(action string, inputs map[string]interface{}) string {
	// Serialize inputs deterministically
	data := map[string]interface{}{
		"action": action,
		"inputs": inputs,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached value
func (c *Cache) Get(action string, inputs map[string]interface{}) (map[string]interface{}, bool) {
	key := c.generateKey(action, inputs)

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > entry.TTL {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(action string, inputs map[string]interface{}, outputs map[string]interface{}, ttl time.Duration) {
	key := c.generateKey(action, inputs)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Key:       key,
		Value:     outputs,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// cleanup periodically removes expired entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.Sub(entry.CreatedAt) > entry.TTL {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size returns the number of cached entries
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}