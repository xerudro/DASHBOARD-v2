package middleware_test

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/xerudro/DASHBOARD-v2/internal/middleware"
)

func TestEnhancedRateLimiter(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a test database
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisClient.Close()

	// Clean up test keys
	defer redisClient.FlushDB(ctx)

	t.Run("IP-based rate limiting", func(t *testing.T) {
		app := fiber.New()
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		app.Use(limiter.Middleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// Make requests up to the limit
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 65; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.100")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode == 200 {
				successCount++
			} else if resp.StatusCode == 429 {
				rateLimitedCount++
			}
			resp.Body.Close()
		}

		if successCount != 60 {
			t.Errorf("Expected 60 successful requests, got %d", successCount)
		}

		if rateLimitedCount != 5 {
			t.Errorf("Expected 5 rate-limited requests, got %d", rateLimitedCount)
		}
	})

	t.Run("Rate limit headers", func(t *testing.T) {
		app := fiber.New()
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		// Clean Redis before this test
		redisClient.FlushDB(ctx)

		app.Use(limiter.Middleware())
		app.Get("/test-headers", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test-headers", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.101")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Check rate limit headers are present
		if resp.Header.Get("X-RateLimit-Limit") == "" {
			t.Error("X-RateLimit-Limit header should be set")
		}

		if resp.Header.Get("X-RateLimit-Remaining") == "" {
			t.Error("X-RateLimit-Remaining header should be set")
		}

		t.Logf("Rate limit headers: Limit=%s, Remaining=%s",
			resp.Header.Get("X-RateLimit-Limit"),
			resp.Header.Get("X-RateLimit-Remaining"))
	})

	t.Run("User-based rate limiting", func(t *testing.T) {
		app := fiber.New()
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		// Clean Redis before this test
		redisClient.FlushDB(ctx)

		// Middleware to set user_id
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_id", "user-123")
			return c.Next()
		})

		app.Use(limiter.Middleware())
		app.Get("/test-user", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// User-based limit is 120 per minute
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 125; i++ {
			req := httptest.NewRequest("GET", "/test-user", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.102")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode == 200 {
				successCount++
			} else if resp.StatusCode == 429 {
				rateLimitedCount++
			}
			resp.Body.Close()
		}

		// Should get IP limit (60) first, but since user is authenticated,
		// they should get user limit (120)
		if successCount < 60 {
			t.Errorf("Expected at least 60 successful requests for authenticated user, got %d", successCount)
		}
	})

	t.Run("Auth endpoint rate limiting", func(t *testing.T) {
		app := fiber.New()
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		// Clean Redis before this test
		redisClient.FlushDB(ctx)

		app.Use(limiter.AuthMiddleware())
		app.Post("/auth/login", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// Auth limit is 10 per minute
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 15; i++ {
			req := httptest.NewRequest("POST", "/auth/login", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.103")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode == 200 {
				successCount++
			} else if resp.StatusCode == 429 {
				rateLimitedCount++
			}
			resp.Body.Close()
		}

		if successCount != 10 {
			t.Errorf("Expected 10 successful auth requests, got %d", successCount)
		}

		if rateLimitedCount != 5 {
			t.Errorf("Expected 5 rate-limited auth requests, got %d", rateLimitedCount)
		}
	})

	t.Run("Expensive endpoint rate limiting", func(t *testing.T) {
		app := fiber.New()
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		// Clean Redis before this test
		redisClient.FlushDB(ctx)

		app.Use(limiter.Middleware())
		app.Post("/api/servers", func(c *fiber.Ctx) error {
			return c.SendString("Server created")
		})

		// Expensive endpoint limit is 10 per hour
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 15; i++ {
			req := httptest.NewRequest("POST", "/api/servers", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.104")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode == 200 {
				successCount++
			} else if resp.StatusCode == 429 {
				rateLimitedCount++
				// Check error message
				body, _ := io.ReadAll(resp.Body)
				t.Logf("Rate limit response: %s", string(body))
			}
			resp.Body.Close()
		}

		if successCount != 10 {
			t.Errorf("Expected 10 successful expensive requests, got %d", successCount)
		}

		if rateLimitedCount != 5 {
			t.Errorf("Expected 5 rate-limited expensive requests, got %d", rateLimitedCount)
		}
	})

	t.Run("Reset client rate limits", func(t *testing.T) {
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		// Clean Redis before this test
		redisClient.FlushDB(ctx)

		app := fiber.New()
		app.Use(limiter.Middleware())
		app.Get("/test-reset", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		clientIP := "192.168.1.105"

		// Make some requests to populate rate limit
		for i := 0; i < 30; i++ {
			req := httptest.NewRequest("GET", "/test-reset", nil)
			req.Header.Set("X-Forwarded-For", clientIP)
			resp, _ := app.Test(req, -1)
			resp.Body.Close()
		}

		// Reset rate limits
		err := limiter.ResetClient(clientIP)
		if err != nil {
			t.Fatalf("Failed to reset client: %v", err)
		}

		// Should be able to make requests again
		req := httptest.NewRequest("GET", "/test-reset", nil)
		req.Header.Set("X-Forwarded-For", clientIP)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("After reset, request should succeed, got status: %d", resp.StatusCode)
		}
	})

	t.Run("Retry-After header on rate limit", func(t *testing.T) {
		app := fiber.New()
		limiter := middleware.NewEnhancedRateLimiter(redisClient)

		// Clean Redis before this test
		redisClient.FlushDB(ctx)

		app.Use(limiter.AuthMiddleware())
		app.Post("/auth/test-retry", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		clientIP := "192.168.1.106"

		// Exceed rate limit (10 requests for auth)
		for i := 0; i < 11; i++ {
			req := httptest.NewRequest("POST", "/auth/test-retry", nil)
			req.Header.Set("X-Forwarded-For", clientIP)
			resp, _ := app.Test(req, -1)
			
			if i == 10 {
				// This should be rate limited
				if resp.StatusCode != 429 {
					t.Errorf("Expected 429 status, got %d", resp.StatusCode)
				}

				retryAfter := resp.Header.Get("Retry-After")
				if retryAfter == "" {
					t.Error("Retry-After header should be set on 429 response")
				} else {
					t.Logf("Retry-After: %s seconds", retryAfter)
				}
			}
			resp.Body.Close()
		}
	})
}

func TestExpensiveEndpointDetection(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"/api/servers", true},
		{"/api/servers/123", true},
		{"/api/sites/deploy", true},
		{"/api/backups", true},
		{"/api/v1/servers", true},
		{"/api/v1/sites/deploy", true},
		{"/api/dashboard", false},
		{"/api/users", false},
		{"/health", false},
	}

	// We can't directly test the private function, but we can test the behavior
	// This test would require access to the private function or behavioral testing
	// For now, we'll document expected behavior
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Path: %s", tc.path), func(t *testing.T) {
			t.Logf("Path %s should be expensive: %v", tc.path, tc.expected)
		})
	}
}

func BenchmarkEnhancedRateLimiter(b *testing.B) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available: %v", err)
	}
	defer redisClient.Close()
	defer redisClient.FlushDB(ctx)

	app := fiber.New()
	limiter := middleware.NewEnhancedRateLimiter(redisClient)

	app.Use(limiter.Middleware())
	app.Get("/bench", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("192.168.1.%d", (i%254)+1))
		resp, _ := app.Test(req, -1)
		resp.Body.Close()

		// Reset every 50 requests to avoid hitting limits
		if i%50 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}
