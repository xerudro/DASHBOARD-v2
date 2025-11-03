package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// ContentSecurityPolicy settings
	EnableCSP           bool
	CSPReportOnly       bool
	AllowInlineScripts  bool // Only use in development
	AllowInlineStyles   bool // Only use in development
	AllowedScriptDomains []string
	AllowedStyleDomains  []string
	AllowedImageDomains  []string
	AllowedFontDomains   []string

	// HSTS settings
	EnableHSTS       bool
	HSTSMaxAge       int  // In seconds (default: 31536000 = 1 year)
	HSTSPreload      bool
	HSTSSubdomains   bool

	// Additional settings
	EnableXSSProtection   bool
	DenyFraming           bool
	StrictReferrerPolicy  bool
}

// DefaultSecurityHeadersConfig returns default security configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableCSP:           true,
		CSPReportOnly:       false,
		AllowInlineScripts:  false, // Strict by default
		AllowInlineStyles:   true,  // Allow for Tailwind and dynamic styling
		AllowedScriptDomains: []string{
			"'self'",
			"https://cdn.tailwindcss.com",
			"https://unpkg.com",
		},
		AllowedStyleDomains: []string{
			"'self'",
			"https://cdn.tailwindcss.com",
			"https://fonts.googleapis.com",
		},
		AllowedImageDomains: []string{
			"'self'",
			"data:",
			"https:",
		},
		AllowedFontDomains: []string{
			"'self'",
			"https://fonts.gstatic.com",
		},
		EnableHSTS:          true,
		HSTSMaxAge:          31536000, // 1 year
		HSTSPreload:         false,
		HSTSSubdomains:      true,
		EnableXSSProtection: true,
		DenyFraming:         true,
		StrictReferrerPolicy: true,
	}
}

// SecurityHeaders adds comprehensive security headers to responses
func SecurityHeaders() fiber.Handler {
	config := DefaultSecurityHeadersConfig()
	return SecurityHeadersWithConfig(config)
}

// SecurityHeadersWithConfig adds security headers with custom configuration
func SecurityHeadersWithConfig(config SecurityHeadersConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Prevent clickjacking
		if config.DenyFraming {
			c.Set("X-Frame-Options", "DENY")
		}

		// 2. Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// 3. Enable XSS protection (legacy header for older browsers)
		if config.EnableXSSProtection {
			c.Set("X-XSS-Protection", "1; mode=block")
		}

		// 4. Enforce HTTPS (HSTS)
		if config.EnableHSTS {
			if c.Get("X-Forwarded-Proto") == "https" || c.Protocol() == "https" {
				hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
				if config.HSTSSubdomains {
					hstsValue += "; includeSubDomains"
				}
				if config.HSTSPreload {
					hstsValue += "; preload"
				}
				c.Set("Strict-Transport-Security", hstsValue)
			}
		}

		// 5. Content Security Policy (CSP)
		if config.EnableCSP {
			csp := buildCSP(config)
			headerName := "Content-Security-Policy"
			if config.CSPReportOnly {
				headerName = "Content-Security-Policy-Report-Only"
			}
			c.Set(headerName, csp)
		}

		// 6. Referrer Policy
		if config.StrictReferrerPolicy {
			c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		} else {
			c.Set("Referrer-Policy", "no-referrer-when-downgrade")
		}

		// 7. Feature Policy / Permissions Policy
		// Restrict access to sensitive browser features
		permissionsPolicy := []string{
			"geolocation=()",
			"microphone=()",
			"camera=()",
			"payment=()",
			"usb=()",
			"magnetometer=()",
			"gyroscope=()",
			"accelerometer=()",
		}
		c.Set("Permissions-Policy", strings.Join(permissionsPolicy, ", "))

		// 8. Cross-Origin Policies
		c.Set("Cross-Origin-Opener-Policy", "same-origin")
		c.Set("Cross-Origin-Resource-Policy", "same-origin")
		c.Set("Cross-Origin-Embedder-Policy", "require-corp")

		return c.Next()
	}
}

// buildCSP constructs the Content-Security-Policy header value
func buildCSP(config SecurityHeadersConfig) string {
	var cspDirectives []string

	// default-src: Fallback for other directives
	cspDirectives = append(cspDirectives, "default-src 'self'")

	// script-src: Control script execution
	scriptSrc := "script-src"
	if config.AllowInlineScripts {
		scriptSrc += " 'unsafe-inline'"
	}
	scriptSrc += " " + strings.Join(config.AllowedScriptDomains, " ")
	cspDirectives = append(cspDirectives, scriptSrc)

	// style-src: Control stylesheet loading
	styleSrc := "style-src"
	if config.AllowInlineStyles {
		styleSrc += " 'unsafe-inline'"
	}
	styleSrc += " " + strings.Join(config.AllowedStyleDomains, " ")
	cspDirectives = append(cspDirectives, styleSrc)

	// img-src: Control image sources
	imgSrc := "img-src " + strings.Join(config.AllowedImageDomains, " ")
	cspDirectives = append(cspDirectives, imgSrc)

	// font-src: Control font sources
	fontSrc := "font-src " + strings.Join(config.AllowedFontDomains, " ")
	cspDirectives = append(cspDirectives, fontSrc)

	// connect-src: Control AJAX, WebSocket, EventSource
	cspDirectives = append(cspDirectives, "connect-src 'self'")

	// object-src: Control <object>, <embed>, <applet>
	cspDirectives = append(cspDirectives, "object-src 'none'")

	// base-uri: Restrict <base> tag URLs
	cspDirectives = append(cspDirectives, "base-uri 'self'")

	// form-action: Restrict form submission targets
	cspDirectives = append(cspDirectives, "form-action 'self'")

	// frame-ancestors: Control embedding (replaces X-Frame-Options)
	cspDirectives = append(cspDirectives, "frame-ancestors 'none'")

	// upgrade-insecure-requests: Automatically upgrade HTTP to HTTPS
	cspDirectives = append(cspDirectives, "upgrade-insecure-requests")

	return strings.Join(cspDirectives, "; ")
}

// GenerateNonce generates a cryptographically secure nonce for CSP
func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}

// CSPNonceMiddleware generates and injects CSP nonce into context
func CSPNonceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		nonce, err := GenerateNonce()
		if err != nil {
			// Log error but don't fail the request
			return c.Next()
		}

		// Store nonce in context for use in templates
		c.Locals("csp_nonce", nonce)

		// Add nonce to CSP header if it exists
		csp := c.Get("Content-Security-Policy")
		if csp != "" {
			// Add nonce to script-src and style-src
			csp = strings.ReplaceAll(csp, "script-src", fmt.Sprintf("script-src 'nonce-%s'", nonce))
			c.Set("Content-Security-Policy", csp)
		}

		return c.Next()
	}
}

// CORS returns a properly configured CORS middleware
func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		
		// Define allowed origins (should be configurable)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://yourdomain.com", // Replace with actual domain
		}
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		// Set CORS headers only for allowed origins
		if allowed {
			c.Set("Access-Control-Allow-Origin", origin)
		}
		
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400") // 24 hours
		
		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}
		
		return c.Next()
	}
}