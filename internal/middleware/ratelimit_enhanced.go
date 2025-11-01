package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// EnhancedRateLimiter implements distributed rate limiting with multi-tier support
type EnhancedRateLimiter struct {
	limiter     *redis_rate.Limiter
	redisClient redis.UniversalClient
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter with redis_rate
func NewEnhancedRateLimiter(redisClient redis.UniversalClient) *EnhancedRateLimiter {
	return &EnhancedRateLimiter{
		limiter:     redis_rate.NewLimiter(redisClient),
		redisClient: redisClient,
	}
}

// Middleware returns a Fiber middleware for multi-tier distributed rate limiting
func (erl *EnhancedRateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		ip := c.IP()
		endpoint := c.Path()

		// Tier 1: IP-based rate limit (stricter - 60 requests per minute)
		ipKey := fmt.Sprintf("rate:ip:%s", ip)
		ipRes, err := erl.limiter.Allow(ctx, ipKey, redis_rate.PerMinute(60))
		if err != nil {
			log.Error().
				Err(err).
				Str("ip", ip).
				Msg("Rate limiter error - allowing request")
			return c.Next()
		}

		if ipRes.Allowed == 0 {
			// Set rate limit headers
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", ipRes.Limit.Burst))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", fmt.Sprintf("%.0f", ipRes.RetryAfter.Seconds()))

			log.Warn().
				Str("ip", ip).
				Str("path", endpoint).
				Msg("IP rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many requests from this IP",
				"retry_after": ipRes.RetryAfter.Seconds(),
			})
		}

		// Tier 2: User-based rate limit (more permissive for authenticated users - 120 requests per minute)
		userID := c.Locals("user_id")
		if userID != nil {
			userIDStr, ok := userID.(string)
			if ok && userIDStr != "" {
				userKey := fmt.Sprintf("rate:user:%s", userIDStr)
				userRes, err := erl.limiter.Allow(ctx, userKey, redis_rate.PerMinute(120))
				if err != nil {
					log.Error().
						Err(err).
						Str("user_id", userIDStr).
						Msg("User rate limiter error - allowing request")
					return c.Next()
				}

				if userRes.Allowed == 0 {
					// Set rate limit headers
					c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", userRes.Limit.Burst))
					c.Set("X-RateLimit-Remaining", "0")
					c.Set("Retry-After", fmt.Sprintf("%.0f", userRes.RetryAfter.Seconds()))

					log.Warn().
						Str("user_id", userIDStr).
						Str("path", endpoint).
						Msg("User rate limit exceeded")

					return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
						"error":       "Too many requests",
						"retry_after": userRes.RetryAfter.Seconds(),
					})
				}

				// Set success rate limit headers for user
				c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", userRes.Limit.Burst))
				c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", userRes.Remaining))
			}
		}

		// Tier 3: Endpoint-specific rate limits for expensive operations
		if isExpensiveEndpoint(endpoint) {
			expKey := fmt.Sprintf("rate:endpoint:%s:%s", ip, endpoint)
			if userID != nil {
				if userIDStr, ok := userID.(string); ok && userIDStr != "" {
					expKey = fmt.Sprintf("rate:endpoint:%s:%s", userIDStr, endpoint)
				}
			}

			expRes, err := erl.limiter.Allow(ctx, expKey, redis_rate.PerHour(10))
			if err != nil {
				log.Error().
					Err(err).
					Str("endpoint", endpoint).
					Msg("Endpoint rate limiter error - allowing request")
				return c.Next()
			}

			if expRes.Allowed == 0 {
				// Set rate limit headers
				c.Set("X-RateLimit-Limit", "10")
				c.Set("X-RateLimit-Remaining", "0")
				c.Set("Retry-After", fmt.Sprintf("%.0f", expRes.RetryAfter.Seconds()))

				log.Warn().
					Str("endpoint", endpoint).
					Str("ip", ip).
					Msg("Expensive endpoint rate limit exceeded")

				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":       "Rate limit exceeded for this operation",
					"retry_after": expRes.RetryAfter.Seconds(),
				})
			}
		}

		// Set default rate limit headers (IP-based)
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", ipRes.Limit.Burst))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", ipRes.Remaining))

		return c.Next()
	}
}

// isExpensiveEndpoint checks if an endpoint requires stricter rate limiting
func isExpensiveEndpoint(path string) bool {
	expensiveOps := []string{
		"/api/servers",      // Server provisioning
		"/api/sites/deploy", // Deployment
		"/api/backups",      // Backup operations
		"/api/v1/servers",   // Alternative server endpoint
		"/api/v1/sites/deploy",
		"/api/v1/backups",
	}

	for _, op := range expensiveOps {
		if strings.HasPrefix(path, op) {
			return true
		}
	}
	return false
}

// AuthMiddleware returns a stricter rate limiter for authentication endpoints
func (erl *EnhancedRateLimiter) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		// Use IP only for auth endpoints (stricter limit: 10 requests per minute)
		ip := c.IP()
		authKey := fmt.Sprintf("rate:auth:%s", ip)

		res, err := erl.limiter.Allow(ctx, authKey, redis_rate.PerMinute(10))
		if err != nil {
			log.Error().
				Err(err).
				Str("ip", ip).
				Msg("Auth rate limiter error - allowing request")
			return c.Next()
		}

		if res.Allowed == 0 {
			// Set rate limit headers
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", res.Limit.Burst))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", fmt.Sprintf("%.0f", res.RetryAfter.Seconds()))

			log.Warn().
				Str("ip", ip).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("Auth rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many authentication attempts. Please wait before trying again.",
				"retry_after": res.RetryAfter.Seconds(),
			})
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", res.Limit.Burst))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))

		return c.Next()
	}
}

// GetClientStats returns rate limit statistics for a client
func (erl *EnhancedRateLimiter) GetClientStats(clientID string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stats := make(map[string]interface{})

	// Get IP-based stats
	ipKey := fmt.Sprintf("rate:ip:%s", clientID)
	ipVal, err := erl.redisClient.Get(ctx, ipKey).Result()
	if err == nil {
		stats["ip_requests"] = ipVal
	}

	// Get user-based stats if applicable
	userKey := fmt.Sprintf("rate:user:%s", clientID)
	userVal, err := erl.redisClient.Get(ctx, userKey).Result()
	if err == nil {
		stats["user_requests"] = userVal
	}

	return stats, nil
}

// ResetClient resets rate limits for a specific client
func (erl *EnhancedRateLimiter) ResetClient(clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Reset all rate limit keys for this client
	patterns := []string{
		fmt.Sprintf("rate:ip:%s", clientID),
		fmt.Sprintf("rate:user:%s", clientID),
		fmt.Sprintf("rate:auth:%s", clientID),
		fmt.Sprintf("rate:endpoint:%s:*", clientID),
	}

	for _, pattern := range patterns {
		keys, err := erl.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			erl.redisClient.Del(ctx, key)
		}
	}

	log.Info().
		Str("client_id", clientID).
		Msg("Rate limits reset for client")

	return nil
}
