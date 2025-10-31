package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey       string
	tokenDuration   time.Duration
	refreshDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		tokenDuration:   tokenDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vip-hosting-panel",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// GenerateRefreshToken generates a refresh token
func (m *JWTManager) GenerateRefreshToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vip-hosting-panel",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidSignature
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

// ExtractUserID extracts user ID from token claims
func (c *JWTClaims) ExtractUserID() uuid.UUID {
	return c.UserID
}

// ExtractTenantID extracts tenant ID from token claims
func (c *JWTClaims) ExtractTenantID() uuid.UUID {
	return c.TenantID
}

// IsSuperAdmin checks if user has superadmin role
func (c *JWTClaims) IsSuperAdmin() bool {
	return c.Role == models.RoleSuperAdmin
}

// IsAdmin checks if user has admin role or higher
func (c *JWTClaims) IsAdmin() bool {
	return c.Role == models.RoleSuperAdmin || c.Role == models.RoleAdmin
}

// IsReseller checks if user has reseller role or higher
func (c *JWTClaims) IsReseller() bool {
	return c.Role == models.RoleSuperAdmin || c.Role == models.RoleAdmin || c.Role == models.RoleReseller
}

// CanAccessTenant checks if user can access resources in the given tenant
func (c *JWTClaims) CanAccessTenant(tenantID uuid.UUID) bool {
	if c.IsSuperAdmin() {
		return true // Superadmin can access all tenants
	}
	return c.TenantID == tenantID
}
