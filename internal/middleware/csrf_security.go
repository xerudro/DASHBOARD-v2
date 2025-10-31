package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// CSRFConfig represents CSRF protection configuration
type CSRFConfig struct {
	TokenLength    int
	TokenLookup    string
	CookieName     string
	CookieExpiry   time.Duration
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
}

// DefaultCSRFConfig returns default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    32,
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieExpiry:   24 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
	}
}

// CSRFProtection provides Cross-Site Request Forgery protection
func CSRFProtection(config ...CSRFConfig) fiber.Handler {
	cfg := DefaultCSRFConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		// Skip CSRF for safe methods
		if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
			// Generate token for safe methods
			token := generateCSRFToken(cfg.TokenLength)
			
			// Set cookie
			c.Cookie(&fiber.Cookie{
				Name:     cfg.CookieName,
				Value:    token,
				Expires:  time.Now().Add(cfg.CookieExpiry),
				Secure:   cfg.CookieSecure,
				HTTPOnly: cfg.CookieHTTPOnly,
				SameSite: cfg.CookieSameSite,
			})
			
			// Set header for JavaScript access
			c.Set("X-CSRF-Token", token)
			
			return c.Next()
		}

		// For unsafe methods, verify token
		cookieToken := c.Cookies(cfg.CookieName)
		headerToken := c.Get("X-CSRF-Token")
		
		if cookieToken == "" || headerToken == "" {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("CSRF token missing")
			
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token missing",
			})
		}

		if cookieToken != headerToken {
			log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Str("method", c.Method()).
				Msg("CSRF token mismatch")
			
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token invalid",
			})
		}

		return c.Next()
	}
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Error().Err(err).Msg("Failed to generate CSRF token")
		// Fallback to timestamp-based token (less secure but functional)
		return fmt.Sprintf("fallback_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// ConfigSecurityMiddleware provides configuration-based security settings
func ConfigSecurityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent information disclosure
		c.Set("Server", "VIP-Panel")
		c.Set("X-Powered-By", "")
		
		// Content type security
		if c.Get("Content-Type") == "" {
			c.Set("Content-Type", "application/json")
		}
		
		// Additional security headers for configuration endpoints
		if contains(c.Path(), []string{"/api/config", "/api/settings"}) {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		}
		
		return c.Next()
	}
}

// contains checks if a string contains any of the given substrings
func contains(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// SecureLoggingMiddleware prevents sensitive data from being logged
func SecureLoggingMiddleware() fiber.Handler {
	sensitiveFields := []string{
		"password", "token", "secret", "key", "authorization",
		"x-api-key", "x-auth-token", "cookie", "session",
	}
	
	return func(c *fiber.Ctx) error {
		// Store original headers for logging
		headers := make(map[string]string)
		c.GetReqHeaders(func(key, value string) {
			keyLower := strings.ToLower(key)
			
			// Check if header contains sensitive information
			isSensitive := false
			for _, field := range sensitiveFields {
				if keyLower == field || strings.Contains(keyLower, field) {
					isSensitive = true
					break
				}
			}
			
			if isSensitive {
				headers[key] = "[REDACTED]"
			} else {
				headers[key] = value
			}
		})
		
		// Continue processing
		err := c.Next()
		
		// Log request (excluding sensitive data)
		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Int("status", c.Response().StatusCode()).
			Interface("headers", headers).
			Msg("Request processed")
		
		return err
	}
}