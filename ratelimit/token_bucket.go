package chat_ratelimit

import (
	"net"
	"sync"
	"time"
)

// TokenBucket represents the configuration for a token bucket rate limiter.
// A bucket fills with tokens at a constant rate.
// Each message consumes one token.
// If the bucket is empty, messages are rejected.
// The bucket has a maximum capacity to limit burst size.
type TokenBucket struct {
	RefillInterval time.Duration // How often tokens are added
	RefillRate     uint8         // Number of tokens to add at each interval
	BucketLimit    uint8         // Maximum number of tokens the bucket can hold
}

// AvailableToken represents the number of available tokens for a connection
type AvailableToken uint8

// Ratelimit manages rate limiting for multiple connections
type Ratelimit struct {
	userRatelimitMap map[net.Conn]AvailableToken
	config           TokenBucket
	mu               sync.Mutex
}

// NewRatelimit initializes a new Ratelimit instance and starts the token refill routine
func NewRatelimit(config TokenBucket) *Ratelimit {
	ratelimit := &Ratelimit{
		userRatelimitMap: make(map[net.Conn]AvailableToken),
		config:           config,
	}

	go ratelimit.refillRoutine()

	return ratelimit
}

// refillRoutine periodically refills tokens for all connections
func (r *Ratelimit) refillRoutine() {
	ticker := time.NewTicker(r.config.RefillInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		for conn, tokens := range r.userRatelimitMap {
			if tokens < AvailableToken(r.config.BucketLimit) {
				newTokens := uint8(tokens) + r.config.RefillRate
				if newTokens > r.config.BucketLimit {
					newTokens = r.config.BucketLimit
				}
				r.userRatelimitMap[conn] = AvailableToken(newTokens)
			}
		}
		r.mu.Unlock()
	}
}

// Add adds a new connection to the rate limiter
func (r *Ratelimit) Add(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.userRatelimitMap[conn] = AvailableToken(r.config.BucketLimit)
}

// Remove removes a connection from the rate limiter
func (r *Ratelimit) Remove(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.userRatelimitMap, conn)
}

// Check checks if the connection is rate limited and consumes a token if available
// Returns true if a token was consumed, false if rate limited or connection not found
func (r *Ratelimit) Check(conn net.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	tokens, exists := r.userRatelimitMap[conn]
	if !exists || tokens == 0 {
		return false
	}
	r.userRatelimitMap[conn] = tokens - 1
	return true
}
