package sdk

import (
	"context"
	"sync"
	"time"
)

type TokenCache interface {
	Get(key string) (*Token, bool)
	Set(key string, token *Token)
}

// MemoryTokenCache is a basic in-memory token cache implementation.
type MemoryTokenCache struct {
	tokens map[string]*Token

	lock sync.RWMutex
}

// NewMemoryTokenCache creates a new in memory token cache instance.
func NewMemoryTokenCache() *MemoryTokenCache {
	return &MemoryTokenCache{
		tokens: map[string]*Token{},
	}
}

// Set adds or updates a token with the given key in the cache.
func (c *MemoryTokenCache) Set(key string, token *Token) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tokens[key] = token
}

// Get retrieves the token associated with the given key from the cache. The bool
// return value will be false if no matching key is found, and true otherwise.
func (c *MemoryTokenCache) Get(key string) (*Token, bool) {
	c.lock.RLock()
	token, ok := c.tokens[key]
	c.lock.RUnlock()

	if ok && token.Expired() {
		c.lock.Lock()
		delete(c.tokens, key)
		c.lock.Unlock()

		return nil, false
	}

	return token, ok
}

// StartGC starts garbage collection of expired tokens.
func (c *MemoryTokenCache) StartGC(ctx context.Context, gcInterval time.Duration) {
	if gcInterval <= 0 {
		return
	}

	ticker := time.NewTicker(gcInterval)

	for {
		select {
		case <-ticker.C:
			c.clearExpired()
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

// clearExpired removes all expired tokens from the cache.
func (c *MemoryTokenCache) clearExpired() {
	for key, token := range c.tokens {
		if token.Expired() {
			c.lock.Lock()
			delete(c.tokens, key)
			c.lock.Unlock()
		}
	}
}
