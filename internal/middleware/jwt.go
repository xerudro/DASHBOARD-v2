package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string        `mapstructure:"secret_key"`
	ExpireTime    time.Duration `mapstructure:"expire_time"`
	RefreshTime   time.Duration `mapstructure:"refresh_time"`
	Issuer        string        `mapstructure:"issuer"`
	CookieName    string        `mapstructure:"cookie_name"`
	CookieSecure  bool          `mapstructure:"cookie_secure"`
	CookieHTTPOnly bool         `mapstructure:"cookie_http_only"`
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	config JWTConfig
}

// NewJWT creates a new JWT middleware
func NewJWT(config JWTConfig) *JWTMiddleware {
	// Set defaults
	if config.ExpireTime == 0 {
		config.ExpireTime = 24 * time.Hour
	}
	if config.RefreshTime == 0 {
		config.RefreshTime = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "vip-hosting-panel"
	}
	if config.CookieName == "" {
		config.CookieName = "auth_token"
	}

	return &JWTMiddleware{config: config}
}

// Protect returns a middleware function that validates JWT tokens
func (m *JWTMiddleware) Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := m.extractToken(c)
		if token == "" {
			return m.unauthorized(c)
		}

		claims, err := m.validateToken(token)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid JWT token")
			return m.unauthorized(c)
		}

		// Store user info in context
		ctx := context.WithValue(c.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "role", claims.Role)
		c.SetUserContext(ctx)

		// Store in Fiber locals for easier access
		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// RequireRole returns a middleware that checks user role
func (m *JWTMiddleware) RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role").(string)
		
		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Insufficient permissions",
		})
	}
}

// RequireTenant ensures the user belongs to the requested tenant
func (m *JWTMiddleware) RequireTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantIDParam := c.Params("tenant_id")
		if tenantIDParam == "" {
			return c.Next() // No tenant parameter, skip check
		}

		requestedTenantID, err := uuid.Parse(tenantIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid tenant ID",
			})
		}

		userTenantID := c.Locals("tenant_id").(uuid.UUID)
		userRole := c.Locals("role").(string)

		// Superadmin can access any tenant
		if userRole == "superadmin" {
			return c.Next()
		}

		// Regular users can only access their own tenant
		if userTenantID != requestedTenantID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied to this tenant",
			})
		}

		return c.Next()
	}
}

// GenerateToken generates a new JWT token
func (m *JWTMiddleware) GenerateToken(userID, tenantID uuid.UUID, email, role string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.ExpireTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

// GenerateRefreshToken generates a refresh token
func (m *JWTMiddleware) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    m.config.Issuer,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTime)),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

// ValidateRefreshToken validates a refresh token
func (m *JWTMiddleware) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// SetTokenCookie sets JWT token as HTTP-only cookie
func (m *JWTMiddleware) SetTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     m.config.CookieName,
		Value:    token,
		Expires:  time.Now().Add(m.config.ExpireTime),
		HTTPOnly: m.config.CookieHTTPOnly,
		Secure:   m.config.CookieSecure,
		SameSite: "Lax",
	})
}

// ClearTokenCookie clears the JWT token cookie
func (m *JWTMiddleware) ClearTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     m.config.CookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: m.config.CookieHTTPOnly,
		Secure:   m.config.CookieSecure,
		SameSite: "Lax",
	})
}

// extractToken extracts JWT token from Authorization header or cookie
func (m *JWTMiddleware) extractToken(c *fiber.Ctx) string {
	// Try Authorization header first
	auth := c.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try cookie
	return c.Cookies(m.config.CookieName)
}

// validateToken validates and parses JWT token
func (m *JWTMiddleware) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// unauthorized returns unauthorized response
func (m *JWTMiddleware) unauthorized(c *fiber.Ctx) error {
	// Return JSON for API routes
	if strings.HasPrefix(c.Path(), "/api/") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized",
		})
	}

	// Redirect to login for web routes
	return c.Redirect("/login")
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(c *fiber.Ctx) (userID, tenantID uuid.UUID, email, role string) {
	userID = c.Locals("user_id").(uuid.UUID)
	tenantID = c.Locals("tenant_id").(uuid.UUID)
	email = c.Locals("email").(string)
	role = c.Locals("role").(string)
	return
}

// IsAdmin checks if user has admin role
func IsAdmin(role string) bool {
	return role == "superadmin" || role == "admin"
}

// IsReseller checks if user has reseller role or higher
func IsReseller(role string) bool {
	return role == "superadmin" || role == "admin" || role == "reseller"
}