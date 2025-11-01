package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

// JWTClaims represents JWT token claims with enhanced security
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	DeviceID string    `json:"device_id,omitempty"` // Device binding for security
	JTI      string    `json:"jti,omitempty"`       // JWT ID for revocation
	jwt.RegisteredClaims
}

// JWTManager handles JWT token creation and validation with Redis-based revocation
type JWTManager struct {
	secretKey       string
	tokenDuration   time.Duration
	refreshDuration time.Duration
	redisClient     *redis.Client
}

// NewJWTManager creates a new JWT manager with Redis support
func NewJWTManager(secretKey string, tokenDuration, refreshDuration time.Duration, redisClient *redis.Client) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		tokenDuration:   tokenDuration,
		refreshDuration: refreshDuration,
		redisClient:     redisClient,
	}
}

// NewJWTManagerWithoutRedis creates a JWT manager without Redis (for backward compatibility)
func NewJWTManagerWithoutRedis(secretKey string, tokenDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		tokenDuration:   tokenDuration,
		refreshDuration: refreshDuration,
		redisClient:     nil,
	}
}

// GenerateToken generates a new JWT token for a user with enhanced security
func (m *JWTManager) GenerateToken(user *models.User) (string, error) {
	return m.GenerateTokenWithDevice(user, "")
}

// GenerateTokenWithDevice generates a new JWT token with device binding
func (m *JWTManager) GenerateTokenWithDevice(user *models.User, deviceID string) (string, error) {
	// Generate unique JWT ID for revocation capability
	jti := uuid.New().String()
	
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		DeviceID: deviceID,
		JTI:      jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vip-hosting-panel",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	
	// Store JTI in Redis for token revocation capability
	if m.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		key := fmt.Sprintf("jwt:%s", jti)
		err = m.redisClient.Set(ctx, key, user.ID.String(), m.tokenDuration).Err()
		if err != nil {
			log.Error().
				Err(err).
				Str("jti", jti).
				Str("user_id", user.ID.String()).
				Msg("Failed to store JWT in Redis - token revocation will not work")
			// Don't fail token generation if Redis is down, just log the error
		}
	}
	
	return signedToken, nil
}

// GenerateRefreshToken generates a refresh token with enhanced security
func (m *JWTManager) GenerateRefreshToken(user *models.User) (string, error) {
	return m.GenerateRefreshTokenWithDevice(user, "")
}

// GenerateRefreshTokenWithDevice generates a refresh token with device binding
func (m *JWTManager) GenerateRefreshTokenWithDevice(user *models.User, deviceID string) (string, error) {
	// Generate unique JWT ID for revocation capability
	jti := uuid.New().String()
	
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		DeviceID: deviceID,
		JTI:      jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vip-hosting-panel",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	
	// Store JTI in Redis for token revocation capability
	if m.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		key := fmt.Sprintf("jwt:refresh:%s", jti)
		err = m.redisClient.Set(ctx, key, user.ID.String(), m.refreshDuration).Err()
		if err != nil {
			log.Error().
				Err(err).
				Str("jti", jti).
				Str("user_id", user.ID.String()).
				Msg("Failed to store refresh JWT in Redis - token revocation will not work")
		}
	}
	
	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims with enhanced security
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing algorithm to prevent algorithm confusion attacks
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		// Check if the error is due to token expiration
		if err.Error() == "token has invalid claims: token is expired" || 
		   err.Error() == "token is expired" {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Additional expiration check (defensive)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}
	
	// Check if token is revoked (if Redis is available)
	if m.redisClient != nil && claims.JTI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		key := fmt.Sprintf("jwt:%s", claims.JTI)
		exists, err := m.redisClient.Exists(ctx, key).Result()
		if err != nil {
			log.Error().
				Err(err).
				Str("jti", claims.JTI).
				Msg("Failed to check token revocation status in Redis")
			// Don't fail validation if Redis is down, just log the error
		} else if exists == 0 {
			return nil, ErrTokenRevoked
		}
	}

	return claims, nil
}

// RevokeToken revokes a JWT token by removing it from Redis
func (m *JWTManager) RevokeToken(jti string) error {
	if m.redisClient == nil {
		return fmt.Errorf("Redis client not configured - token revocation not available")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	key := fmt.Sprintf("jwt:%s", jti)
	err := m.redisClient.Del(ctx, key).Err()
	if err != nil {
		log.Error().
			Err(err).
			Str("jti", jti).
			Msg("Failed to revoke token")
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	
	log.Info().
		Str("jti", jti).
		Msg("Token revoked successfully")
	
	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (m *JWTManager) RevokeAllUserTokens(userID uuid.UUID) error {
	if m.redisClient == nil {
		return fmt.Errorf("Redis client not configured - token revocation not available")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Scan for all JWT keys and check their user ID
	var cursor uint64
	var revokedCount int
	
	for {
		keys, nextCursor, err := m.redisClient.Scan(ctx, cursor, "jwt:*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan tokens: %w", err)
		}
		
		for _, key := range keys {
			storedUserID, err := m.redisClient.Get(ctx, key).Result()
			if err == nil && storedUserID == userID.String() {
				if err := m.redisClient.Del(ctx, key).Err(); err == nil {
					revokedCount++
				}
			}
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	log.Info().
		Str("user_id", userID.String()).
		Int("revoked_count", revokedCount).
		Msg("All user tokens revoked")
	
	return nil
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
