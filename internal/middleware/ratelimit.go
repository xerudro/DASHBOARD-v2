package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*ClientData
	maxReqs  int
	window   time.Duration
	cleanup  time.Duration
}

// ClientData holds rate limiting data for a client
type ClientData struct {
	requests []time.Time
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*ClientData),
		maxReqs:  maxReqs,
		window:   window,
		cleanup:  window * 2, // Cleanup old entries every 2x window
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Middleware returns a Fiber middleware for rate limiting
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP + User-Agent for better uniqueness)
		clientID := c.IP() + ":" + c.Get("User-Agent")
		
		if !rl.isAllowed(clientID) {
			log.Warn().
				Str("client_id", clientID).
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Msg("Rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Rate limit exceeded. Please try again later.",
			})
		}

		return c.Next()
	}
}

// AuthMiddleware returns a stricter rate limiter for auth endpoints
func (rl *RateLimiter) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Stricter limits for auth endpoints
		clientID := c.IP()
		
		if !rl.isAllowed(clientID) {
			log.Warn().
				Str("client_id", clientID).
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Auth rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Too many authentication attempts. Please wait before trying again.",
			})
		}

		return c.Next()
	}
}

// isAllowed checks if the client is allowed to make a request
func (rl *RateLimiter) isAllowed(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// Get or create client data
	client, exists := rl.clients[clientID]
	if !exists {
		client = &ClientData{
			requests: make([]time.Time, 0),
			lastSeen: now,
		}
		rl.clients[clientID] = client
	}

	// Update last seen
	client.lastSeen = now

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if we're under the limit
	if len(client.requests) >= rl.maxReqs {
		return false
	}

	// Add this request
	client.requests = append(client.requests, now)
	return true
}

// cleanupLoop removes old client data to prevent memory leaks
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cleanup)
		
		for clientID, client := range rl.clients {
			if client.lastSeen.Before(cutoff) {
				delete(rl.clients, clientID)
			}
		}
		rl.mu.Unlock()
	}
}