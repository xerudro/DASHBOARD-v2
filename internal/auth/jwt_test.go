package auth_test

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/xerudro/DASHBOARD-v2/internal/auth"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

func TestJWTEnhancements(t *testing.T) {
	// Setup
	secretKey := "test-secret-key-for-jwt-testing-12345"
	tokenDuration := 24 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	t.Run("JWT with JTI generation", func(t *testing.T) {
		// Test without Redis (backward compatibility)
		jwtManager := auth.NewJWTManagerWithoutRedis(secretKey, tokenDuration, refreshDuration)

		testUser := &models.User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "test@example.com",
			Role:     models.RoleAdmin,
		}

		token, err := jwtManager.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if token == "" {
			t.Fatal("Generated token is empty")
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.UserID != testUser.ID {
			t.Errorf("UserID mismatch: got %v, want %v", claims.UserID, testUser.ID)
		}

		if claims.Email != testUser.Email {
			t.Errorf("Email mismatch: got %v, want %v", claims.Email, testUser.Email)
		}

		if claims.JTI == "" {
			t.Error("JTI should be generated but is empty")
		}
	})

	t.Run("JWT with device binding", func(t *testing.T) {
		jwtManager := auth.NewJWTManagerWithoutRedis(secretKey, tokenDuration, refreshDuration)

		testUser := &models.User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "device-test@example.com",
			Role:     models.RoleClient,
		}

		deviceID := "device-12345-abcde"
		token, err := jwtManager.GenerateTokenWithDevice(testUser, deviceID)
		if err != nil {
			t.Fatalf("Failed to generate token with device: %v", err)
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.DeviceID != deviceID {
			t.Errorf("DeviceID mismatch: got %v, want %v", claims.DeviceID, deviceID)
		}
	})

	t.Run("Algorithm confusion prevention", func(t *testing.T) {
		jwtManager := auth.NewJWTManagerWithoutRedis(secretKey, tokenDuration, refreshDuration)

		// Create a token with "none" algorithm (security vulnerability)
		maliciousToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoiMTIzIiwiZW1haWwiOiJoYWNrZXJAZXhhbXBsZS5jb20ifQ."

		_, err := jwtManager.ValidateToken(maliciousToken)
		if err == nil {
			t.Error("Should reject token with 'none' algorithm")
		}
	})

	t.Run("Invalid token rejection", func(t *testing.T) {
		jwtManager := auth.NewJWTManagerWithoutRedis(secretKey, tokenDuration, refreshDuration)

		invalidTokens := []string{
			"invalid.jwt.token",
			"",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
		}

		for _, token := range invalidTokens {
			_, err := jwtManager.ValidateToken(token)
			if err == nil {
				t.Errorf("Should reject invalid token: %s", token)
			}
		}
	})

	t.Run("Expired token detection", func(t *testing.T) {
		shortDuration := 1 * time.Millisecond
		jwtManager := auth.NewJWTManagerWithoutRedis(secretKey, shortDuration, refreshDuration)

		testUser := &models.User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "expiry-test@example.com",
			Role:     models.RoleClient,
		}

		token, err := jwtManager.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = jwtManager.ValidateToken(token)
		if err != auth.ErrTokenExpired {
			t.Errorf("Expected ErrTokenExpired, got: %v", err)
		}
	})
}

func TestJWTWithRedis(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a test database
	})

	ctx := redisClient.Context()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisClient.Close()

	// Clean up test keys after test
	defer redisClient.FlushDB(ctx)

	secretKey := "test-secret-key-with-redis-12345"
	tokenDuration := 24 * time.Hour
	refreshDuration := 7 * 24 * time.Hour

	t.Run("Token revocation", func(t *testing.T) {
		jwtManager := auth.NewJWTManager(secretKey, tokenDuration, refreshDuration, redisClient)

		testUser := &models.User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "revoke-test@example.com",
			Role:     models.RoleAdmin,
		}

		token, err := jwtManager.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Token should be valid initially
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			t.Fatalf("Token should be valid: %v", err)
		}

		// Revoke the token
		err = jwtManager.RevokeToken(claims.JTI)
		if err != nil {
			t.Fatalf("Failed to revoke token: %v", err)
		}

		// Token should now be invalid
		_, err = jwtManager.ValidateToken(token)
		if err != auth.ErrTokenRevoked {
			t.Errorf("Expected ErrTokenRevoked, got: %v", err)
		}
	})

	t.Run("Revoke all user tokens", func(t *testing.T) {
		jwtManager := auth.NewJWTManager(secretKey, tokenDuration, refreshDuration, redisClient)

		testUser := &models.User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "revoke-all-test@example.com",
			Role:     models.RoleReseller,
		}

		// Generate multiple tokens for the same user
		token1, _ := jwtManager.GenerateToken(testUser)
		token2, _ := jwtManager.GenerateToken(testUser)
		token3, _ := jwtManager.GenerateTokenWithDevice(testUser, "device-1")

		// All tokens should be valid
		if _, err := jwtManager.ValidateToken(token1); err != nil {
			t.Errorf("Token 1 should be valid: %v", err)
		}
		if _, err := jwtManager.ValidateToken(token2); err != nil {
			t.Errorf("Token 2 should be valid: %v", err)
		}
		if _, err := jwtManager.ValidateToken(token3); err != nil {
			t.Errorf("Token 3 should be valid: %v", err)
		}

		// Revoke all tokens for the user
		err := jwtManager.RevokeAllUserTokens(testUser.ID)
		if err != nil {
			t.Fatalf("Failed to revoke all user tokens: %v", err)
		}

		// All tokens should now be invalid
		_, err = jwtManager.ValidateToken(token1)
		if err != auth.ErrTokenRevoked {
			t.Errorf("Token 1 should be revoked, got: %v", err)
		}
		_, err = jwtManager.ValidateToken(token2)
		if err != auth.ErrTokenRevoked {
			t.Errorf("Token 2 should be revoked, got: %v", err)
		}
		_, err = jwtManager.ValidateToken(token3)
		if err != auth.ErrTokenRevoked {
			t.Errorf("Token 3 should be revoked, got: %v", err)
		}
	})
}

func TestJWTClaimsMethods(t *testing.T) {
	superadminClaims := &auth.JWTClaims{Role: models.RoleSuperAdmin}
	adminClaims := &auth.JWTClaims{Role: models.RoleAdmin}
	resellerClaims := &auth.JWTClaims{Role: models.RoleReseller}
	clientClaims := &auth.JWTClaims{Role: models.RoleClient}

	t.Run("Role checks", func(t *testing.T) {
		if !superadminClaims.IsSuperAdmin() {
			t.Error("Superadmin should return true for IsSuperAdmin")
		}
		if !superadminClaims.IsAdmin() {
			t.Error("Superadmin should return true for IsAdmin")
		}
		if !adminClaims.IsAdmin() {
			t.Error("Admin should return true for IsAdmin")
		}
		if adminClaims.IsSuperAdmin() {
			t.Error("Admin should return false for IsSuperAdmin")
		}
		if !resellerClaims.IsReseller() {
			t.Error("Reseller should return true for IsReseller")
		}
		if clientClaims.IsAdmin() {
			t.Error("Client should return false for IsAdmin")
		}
	})

	t.Run("Tenant access", func(t *testing.T) {
		tenant1 := uuid.New()
		tenant2 := uuid.New()

		superadminClaims.TenantID = tenant1
		adminClaims.TenantID = tenant1

		// Superadmin can access any tenant
		if !superadminClaims.CanAccessTenant(tenant2) {
			t.Error("Superadmin should access any tenant")
		}

		// Admin can only access their own tenant
		if !adminClaims.CanAccessTenant(tenant1) {
			t.Error("Admin should access their own tenant")
		}
		if adminClaims.CanAccessTenant(tenant2) {
			t.Error("Admin should not access other tenants")
		}
	})
}
