package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	client     *redis.Client
	maxReqs    int
	window     time.Duration
	keyPrefix  string
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, maxReqs int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:    client,
		maxReqs:   maxReqs,
		window:    window,
		keyPrefix: "ratelimit:",
	}
}

// Middleware returns a Fiber middleware for distributed rate limiting
func (rrl *RedisRateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP + User-Agent for better uniqueness)
		clientID := c.IP() + ":" + c.Get("User-Agent")
		key := rrl.keyPrefix + clientID

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Use Redis pipeline for atomic operations
		pipe := rrl.client.Pipeline()

		// Increment counter
		incrCmd := pipe.Incr(ctx, key)

		// Set expiration on first request
		pipe.Expire(ctx, key, rrl.window)

		// Get TTL to check if key is new
		ttlCmd := pipe.TTL(ctx, key)

		// Execute pipeline
		_, err := pipe.Exec(ctx)
		if err != nil {
			// Log error but allow request (fail open)
			log.Error().
				Err(err).
				Str("client_id", clientID).
				Msg("Rate limiter Redis error - allowing request")
			return c.Next()
		}

		// Get counter value
		count := incrCmd.Val()
		ttl := ttlCmd.Val()

		// If TTL is -1, the key exists but has no expiration (edge case)
		// Set expiration again
		if ttl == -1 {
			rrl.client.Expire(ctx, key, rrl.window)
		}

		// Check if rate limit exceeded
		if count > int64(rrl.maxReqs) {
			// Calculate time until rate limit resets
			resetTime := int(ttl.Seconds())
			if resetTime < 0 {
				resetTime = int(rrl.window.Seconds())
			}

			// Set rate limit headers
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rrl.maxReqs))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Duration(resetTime)*time.Second).Unix()))
			c.Set("Retry-After", fmt.Sprintf("%d", resetTime))

			log.Warn().
				Str("client_id", clientID).
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Int64("count", count).
				Int("max", rrl.maxReqs).
				Msg("Rate limit exceeded (Redis)")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       true,
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": resetTime,
			})
		}

		// Set rate limit headers for successful requests
		remaining := rrl.maxReqs - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rrl.maxReqs))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rrl.window).Unix()))

		return c.Next()
	}
}

// AuthMiddleware returns a stricter rate limiter for auth endpoints
func (rrl *RedisRateLimiter) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Stricter limits for auth endpoints (IP only)
		clientID := c.IP()
		key := rrl.keyPrefix + "auth:" + clientID

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Use sliding window algorithm for auth endpoints
		now := time.Now().UnixNano()
		windowStart := now - int64(rrl.window)

		// Remove old entries and count current requests
		pipe := rrl.client.Pipeline()

		// Remove requests outside the time window
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// Add current request
		pipe.ZAdd(ctx, key, &redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d", now),
		})

		// Count requests in window
		countCmd := pipe.ZCard(ctx, key)

		// Set expiration
		pipe.Expire(ctx, key, rrl.window*2)

		// Execute pipeline
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Str("client_id", clientID).
				Msg("Auth rate limiter Redis error - allowing request")
			return c.Next()
		}

		count := countCmd.Val()

		// Check if rate limit exceeded (stricter limit: 10 requests per window)
		authLimit := 10
		if count > int64(authLimit) {
			log.Warn().
				Str("client_id", clientID).
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Int64("count", count).
				Int("limit", authLimit).
				Msg("Auth rate limit exceeded (Redis)")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Too many authentication attempts. Please wait before trying again.",
			})
		}

		return c.Next()
	}
}

// GetClientStats returns rate limit statistics for a client
func (rrl *RedisRateLimiter) GetClientStats(clientID string) (int64, time.Duration, error) {
	key := rrl.keyPrefix + clientID

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Get current count
	count, err := rrl.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, nil // No requests yet
		}
		return 0, 0, err
	}

	// Get TTL
	ttl, err := rrl.client.TTL(ctx, key).Result()
	if err != nil {
		return count, 0, err
	}

	return count, ttl, nil
}

// ResetClient resets rate limit for a specific client
func (rrl *RedisRateLimiter) ResetClient(clientID string) error {
	key := rrl.keyPrefix + clientID

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := rrl.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit for client %s: %w", clientID, err)
	}

	log.Info().
		Str("client_id", clientID).
		Msg("Rate limit reset for client")

	return nil
}

// CleanupExpiredKeys removes expired keys (optional maintenance)
func (rrl *RedisRateLimiter) CleanupExpiredKeys() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Scan for rate limit keys
	var cursor uint64
	var deletedCount int

	for {
		keys, nextCursor, err := rrl.client.Scan(ctx, cursor, rrl.keyPrefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		// Check TTL and delete keys without expiration
		for _, key := range keys {
			ttl, err := rrl.client.TTL(ctx, key).Result()
			if err != nil {
				continue
			}

			if ttl == -1 {
				// Key has no expiration, set it
				rrl.client.Expire(ctx, key, rrl.window)
			}
		}

		deletedCount += len(keys)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	log.Info().
		Int("cleaned_keys", deletedCount).
		Msg("Rate limiter cleanup completed")

	return nil
}

// GetGlobalStats returns global rate limiting statistics
func (rrl *RedisRateLimiter) GetGlobalStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cursor uint64
	var totalClients int
	var totalRequests int64

	for {
		keys, nextCursor, err := rrl.client.Scan(ctx, cursor, rrl.keyPrefix+"*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		totalClients += len(keys)

		// Sum up all request counts
		for _, key := range keys {
			count, err := rrl.client.Get(ctx, key).Int64()
			if err == nil {
				totalRequests += count
			}
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return map[string]interface{}{
		"total_clients":  totalClients,
		"total_requests": totalRequests,
		"window":         rrl.window.String(),
		"max_per_window": rrl.maxReqs,
	}, nil
}
