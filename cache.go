package pasis

import (
	"sync"
	"time"
)

// TokenCache defines the interface for token storage.
type TokenCache interface {
	// Get retrieves the stored tokens and expiration time.
	// Returns empty strings and zero time if no tokens are cached.
	Get() (token string, refreshToken string, expiresAt time.Time, err error)

	// Set stores the tokens and expiration time.
	Set(token, refreshToken string, expiresAt time.Time) error

	// Clear removes all stored tokens.
	Clear() error
}

// InMemoryTokenCache is a thread-safe in-memory implementation of TokenCache.
type InMemoryTokenCache struct {
	mu           sync.RWMutex
	token        string
	refreshToken string
	expiresAt    time.Time
}

// NewInMemoryTokenCache creates a new in-memory token cache.
func NewInMemoryTokenCache() *InMemoryTokenCache {
	return &InMemoryTokenCache{}
}

// Get retrieves the stored tokens and expiration time.
func (c *InMemoryTokenCache) Get() (string, string, time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token, c.refreshToken, c.expiresAt, nil
}

// Set stores the tokens and expiration time.
func (c *InMemoryTokenCache) Set(token, refreshToken string, expiresAt time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	c.refreshToken = refreshToken
	c.expiresAt = expiresAt
	return nil
}

// Clear removes all stored tokens.
func (c *InMemoryTokenCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
	c.refreshToken = ""
	c.expiresAt = time.Time{}
	return nil
}
