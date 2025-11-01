package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

// Constants for better maintainability
const (
	// Redis key prefixes
	RedisKeyJWT        = "jwt:"
	RedisKeyJWTRefresh = "jwt:refresh:"
	RedisKeyJWTMeta    = "jwt:meta:"
	RedisKeyRateLimit  = "rate:token:gen:"
	RedisKeyAudit      = "audit:token:"

	// Rate limits
	TokenGenerationMaxPerMinute = 10

	// Timeouts
	RedisOperationTimeout = 2 * time.Second
	RedisCleanupTimeout   = 30 * time.Second
	RedisScanTimeout      = 5 * time.Second
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

// TokenAuditLog represents audit trail for security-sensitive operations
type TokenAuditLog struct {
	JTI       string    `json:"jti"`
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"` // "generated", "validated", "revoked", "refreshed"
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Success   bool      `json:"success"`
	Reason    string    `json:"reason,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// TokenSession represents an active user session
type TokenSession struct {
	JTI       string    `json:"jti"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	DeviceID  string    `json:"device_id"`
	CreatedAt time.Time `json:"created_at"`
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
	return m.GenerateTokenWithMetadata(user, deviceID, "", "")
}

// GenerateTokenWithMetadata generates a new JWT token with full metadata tracking
func (m *JWTManager) GenerateTokenWithMetadata(user *models.User, deviceID string, ip string, userAgent string) (string, error) {
	// Check rate limit
	if err := m.CheckTokenGenerationRateLimit(user.ID); err != nil {
		m.logTokenEvent(TokenAuditLog{
			UserID:    user.ID,
			Action:    "generated",
			IP:        ip,
			UserAgent: userAgent,
			Success:   false,
			Reason:    "rate limit exceeded",
			Timestamp: time.Now(),
		})
		return "", err
	}

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
		m.logTokenEvent(TokenAuditLog{
			JTI:       jti,
			UserID:    user.ID,
			Action:    "generated",
			IP:        ip,
			UserAgent: userAgent,
			Success:   false,
			Reason:    fmt.Sprintf("signing failed: %v", err),
			Timestamp: time.Now(),
		})
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Store JTI in Redis for token revocation capability
	if m.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
		defer cancel()

		key := fmt.Sprintf("%s%s", RedisKeyJWT, jti)
		err = m.redisClient.Set(ctx, key, user.ID.String(), m.tokenDuration).Err()
		if err != nil {
			log.Error().
				Err(err).
				Str("jti", jti).
				Str("user_id", user.ID.String()).
				Msg("Failed to store JWT in Redis - token revocation will not work")
			// Don't fail token generation if Redis is down, just log the error
		}

		// Store metadata for security checks
		metaKey := fmt.Sprintf("%s%s", RedisKeyJWTMeta, jti)
		m.redisClient.HSet(ctx, metaKey, map[string]interface{}{
			"ip":         ip,
			"user_agent": userAgent,
			"device_id":  deviceID,
			"created_at": time.Now().Unix(),
		})
		m.redisClient.Expire(ctx, metaKey, m.tokenDuration)
	}

	// Log successful generation
	m.logTokenEvent(TokenAuditLog{
		JTI:       jti,
		UserID:    user.ID,
		Action:    "generated",
		IP:        ip,
		UserAgent: userAgent,
		Success:   true,
		Timestamp: time.Now(),
	})

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
		ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
		defer cancel()

		key := fmt.Sprintf("%s%s", RedisKeyJWTRefresh, jti)
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
		m.logTokenEvent(TokenAuditLog{
			JTI:       claims.JTI,
			UserID:    claims.UserID,
			Action:    "validated",
			Success:   false,
			Reason:    "token expired",
			Timestamp: time.Now(),
		})
		return nil, ErrTokenExpired
	}

	// Check if token is revoked (if Redis is available)
	if m.redisClient != nil && claims.JTI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
		defer cancel()

		key := fmt.Sprintf("%s%s", RedisKeyJWT, claims.JTI)
		exists, err := m.redisClient.Exists(ctx, key).Result()
		if err != nil {
			log.Error().
				Err(err).
				Str("jti", claims.JTI).
				Msg("Failed to check token revocation status in Redis")
			// Don't fail validation if Redis is down, just log the error
		} else if exists == 0 {
			m.logTokenEvent(TokenAuditLog{
				JTI:       claims.JTI,
				UserID:    claims.UserID,
				Action:    "validated",
				Success:   false,
				Reason:    "token revoked",
				Timestamp: time.Now(),
			})
			return nil, ErrTokenRevoked
		}
	}

	// Log successful validation
	m.logTokenEvent(TokenAuditLog{
		JTI:       claims.JTI,
		UserID:    claims.UserID,
		Action:    "validated",
		Success:   true,
		Timestamp: time.Now(),
	})

	return claims, nil
}

// RevokeToken revokes a JWT token by removing it from Redis
func (m *JWTManager) RevokeToken(jti string) error {
	if m.redisClient == nil {
		return fmt.Errorf("Redis client not configured - token revocation not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
	defer cancel()

	// Remove both access and refresh tokens
	key := fmt.Sprintf("%s%s", RedisKeyJWT, jti)
	refreshKey := fmt.Sprintf("%s%s", RedisKeyJWTRefresh, jti)
	metaKey := fmt.Sprintf("%s%s", RedisKeyJWTMeta, jti)

	err := m.redisClient.Del(ctx, key, refreshKey, metaKey).Err()
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

	ctx, cancel := context.WithTimeout(context.Background(), RedisScanTimeout)
	defer cancel()

	// Scan for all JWT keys and check their user ID
	var cursor uint64
	var revokedCount int

	for {
		keys, nextCursor, err := m.redisClient.Scan(ctx, cursor, fmt.Sprintf("%s*", RedisKeyJWT), 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan tokens: %w", err)
		}

		for _, key := range keys {
			storedUserID, err := m.redisClient.Get(ctx, key).Result()
			if err == nil && storedUserID == userID.String() {
				// Extract JTI from key
				jti := strings.TrimPrefix(key, RedisKeyJWT)

				// Remove access token, refresh token, and metadata
				m.redisClient.Del(ctx,
					fmt.Sprintf("%s%s", RedisKeyJWT, jti),
					fmt.Sprintf("%s%s", RedisKeyJWTRefresh, jti),
					fmt.Sprintf("%s%s", RedisKeyJWTMeta, jti),
				)
				revokedCount++
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

// ==================== Enhanced Security Features ====================

// ValidateTokenWithSecurityChecks validates token with additional security checks for IP/User-Agent changes
func (m *JWTManager) ValidateTokenWithSecurityChecks(tokenString string, currentIP string, currentUserAgent string) (*JWTClaims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Optional: Check for IP/User-Agent changes (if stored during token creation)
	if m.redisClient != nil && claims.JTI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
		defer cancel()

		// Get stored metadata
		metaKey := fmt.Sprintf("%s%s", RedisKeyJWTMeta, claims.JTI)
		storedMeta, err := m.redisClient.HGetAll(ctx, metaKey).Result()
		if err == nil && len(storedMeta) > 0 {
			// Check for IP change (optional security measure - log warning only)
			if storedIP := storedMeta["ip"]; storedIP != "" && storedIP != currentIP {
				log.Warn().
					Str("jti", claims.JTI).
					Str("stored_ip", storedIP).
					Str("current_ip", currentIP).
					Msg("Token used from different IP address")

				// Optional: Revoke token or require re-authentication
				// Uncomment the line below for strict IP binding
				// return nil, fmt.Errorf("token used from different IP")
			}

			// Check for User-Agent change
			if storedUA := storedMeta["user_agent"]; storedUA != "" && storedUA != currentUserAgent {
				log.Warn().
					Str("jti", claims.JTI).
					Str("stored_ua", storedUA).
					Str("current_ua", currentUserAgent).
					Msg("Token used from different User-Agent")
			}
		}
	}

	return claims, nil
}

// CheckTokenGenerationRateLimit prevents token generation abuse
func (m *JWTManager) CheckTokenGenerationRateLimit(userID uuid.UUID) error {
	if m.redisClient == nil {
		return nil // Skip if Redis unavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
	defer cancel()

	// Allow max 10 token generations per minute per user
	key := fmt.Sprintf("%s%s", RedisKeyRateLimit, userID.String())
	count, err := m.redisClient.Incr(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to check rate limit")
		return nil // Don't fail if Redis is down
	}

	if count == 1 {
		m.redisClient.Expire(ctx, key, 1*time.Minute)
	}

	if count > TokenGenerationMaxPerMinute {
		log.Warn().
			Str("user_id", userID.String()).
			Int64("attempts", count).
			Msg("Token generation rate limit exceeded")
		return fmt.Errorf("too many token generation attempts, please try again later")
	}

	return nil
}

// RefreshAccessToken implements secure token refresh with rotation
func (m *JWTManager) RefreshAccessToken(refreshToken string, ip string, userAgent string) (newAccessToken string, newRefreshToken string, err error) {
	// Validate refresh token
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		m.logTokenEvent(TokenAuditLog{
			Action:    "refreshed",
			IP:        ip,
			UserAgent: userAgent,
			Success:   false,
			Reason:    fmt.Sprintf("invalid refresh token: %v", err),
			Timestamp: time.Now(),
		})
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify it's actually a refresh token
	if m.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
		defer cancel()

		refreshKey := fmt.Sprintf("%s%s", RedisKeyJWTRefresh, claims.JTI)
		exists, _ := m.redisClient.Exists(ctx, refreshKey).Result()
		if exists == 0 {
			m.logTokenEvent(TokenAuditLog{
				JTI:       claims.JTI,
				UserID:    claims.UserID,
				Action:    "refreshed",
				IP:        ip,
				UserAgent: userAgent,
				Success:   false,
				Reason:    "refresh token not found or revoked",
				Timestamp: time.Now(),
			})
			return "", "", fmt.Errorf("refresh token not found or revoked")
		}
	}

	// Create user object from claims
	user := &models.User{
		ID:       claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
		Role:     claims.Role,
	}

	// Generate new access token
	newAccessToken, err = m.GenerateTokenWithMetadata(user, claims.DeviceID, ip, userAgent)
	if err != nil {
		m.logTokenEvent(TokenAuditLog{
			JTI:       claims.JTI,
			UserID:    claims.UserID,
			Action:    "refreshed",
			IP:        ip,
			UserAgent: userAgent,
			Success:   false,
			Reason:    fmt.Sprintf("failed to generate new access token: %v", err),
			Timestamp: time.Now(),
		})
		return "", "", err
	}

	// Generate new refresh token (rotation)
	newRefreshToken, err = m.GenerateRefreshTokenWithDevice(user, claims.DeviceID)
	if err != nil {
		m.logTokenEvent(TokenAuditLog{
			JTI:       claims.JTI,
			UserID:    claims.UserID,
			Action:    "refreshed",
			IP:        ip,
			UserAgent: userAgent,
			Success:   false,
			Reason:    fmt.Sprintf("failed to generate new refresh token: %v", err),
			Timestamp: time.Now(),
		})
		return "", "", err
	}

	// Revoke old refresh token to prevent reuse
	m.RevokeToken(claims.JTI)

	m.logTokenEvent(TokenAuditLog{
		JTI:       claims.JTI,
		UserID:    claims.UserID,
		Action:    "refreshed",
		IP:        ip,
		UserAgent: userAgent,
		Success:   true,
		Timestamp: time.Now(),
	})

	log.Info().
		Str("user_id", user.ID.String()).
		Str("old_jti", claims.JTI).
		Msg("Token refreshed successfully")

	return newAccessToken, newRefreshToken, nil
}

// StartTokenCleanupJob starts a periodic cleanup job to remove expired tokens from Redis
func (m *JWTManager) StartTokenCleanupJob(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info().Dur("interval", interval).Msg("Starting token cleanup job")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Token cleanup job stopped")
			return
		case <-ticker.C:
			m.cleanupExpiredTokens()
		}
	}
}

func (m *JWTManager) cleanupExpiredTokens() {
	if m.redisClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), RedisCleanupTimeout)
	defer cancel()

	var cursor uint64
	var cleanedCount int

	for {
		keys, nextCursor, err := m.redisClient.Scan(ctx, cursor, "jwt:*", 1000).Result()
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan for expired tokens")
			break
		}

		// Redis TTL will auto-delete, but we can check for orphaned entries
		for _, key := range keys {
			ttl, _ := m.redisClient.TTL(ctx, key).Result()
			if ttl == -1 { // Key exists but has no TTL
				m.redisClient.Del(ctx, key)
				cleanedCount++
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if cleanedCount > 0 {
		log.Info().Int("cleaned", cleanedCount).Msg("Cleaned up orphaned tokens")
	}
}

// GetActiveUserSessions retrieves all active sessions for a user
func (m *JWTManager) GetActiveUserSessions(userID uuid.UUID) ([]TokenSession, error) {
	if m.redisClient == nil {
		return nil, fmt.Errorf("Redis not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), RedisScanTimeout)
	defer cancel()

	var sessions []TokenSession
	var cursor uint64

	for {
		keys, nextCursor, err := m.redisClient.Scan(ctx, cursor, fmt.Sprintf("%s*", RedisKeyJWTMeta), 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			meta, err := m.redisClient.HGetAll(ctx, key).Result()
			if err != nil || len(meta) == 0 {
				continue
			}

			// Extract JTI from key (jwt:meta:JTI)
			jti := strings.TrimPrefix(key, RedisKeyJWTMeta)

			// Check if this token belongs to the user
			tokenKey := fmt.Sprintf("%s%s", RedisKeyJWT, jti)
			storedUserID, err := m.redisClient.Get(ctx, tokenKey).Result()
			if err == nil && storedUserID == userID.String() {
				createdAt, _ := strconv.ParseInt(meta["created_at"], 10, 64)
				sessions = append(sessions, TokenSession{
					JTI:       jti,
					IP:        meta["ip"],
					UserAgent: meta["user_agent"],
					DeviceID:  meta["device_id"],
					CreatedAt: time.Unix(createdAt, 0),
				})
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return sessions, nil
}

// ValidateMultipleTokens validates tokens in bulk (useful for checking active sessions)
func (m *JWTManager) ValidateMultipleTokens(tokens []string) map[string]*JWTClaims {
	results := make(map[string]*JWTClaims)

	for _, token := range tokens {
		claims, err := m.ValidateToken(token)
		if err == nil {
			results[token] = claims
		}
	}

	return results
}

// logTokenEvent logs audit trail for security-sensitive operations
func (m *JWTManager) logTokenEvent(auditLog TokenAuditLog) {
	// Log to structured logging
	logEvent := log.Info()
	if !auditLog.Success {
		logEvent = log.Warn()
	}

	logEvent.
		Str("jti", auditLog.JTI).
		Str("user_id", auditLog.UserID.String()).
		Str("action", auditLog.Action).
		Str("ip", auditLog.IP).
		Bool("success", auditLog.Success).
		Str("reason", auditLog.Reason).
		Msg("Token audit event")

	// Optional: Store in Redis for audit trail
	if m.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), RedisOperationTimeout)
		defer cancel()

		auditKey := fmt.Sprintf("%s%s:%d", RedisKeyAudit, auditLog.UserID.String(), time.Now().Unix())
		data, _ := json.Marshal(auditLog)
		m.redisClient.Set(ctx, auditKey, data, 30*24*time.Hour) // Keep for 30 days
	}
}

// getErrorReason extracts a readable error reason from an error
func getErrorReason(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
