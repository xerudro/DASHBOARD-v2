# JWT and Rate Limiting Security Enhancements

## Overview
This document describes the security enhancements made to JWT token management and distributed rate limiting in the VIP Hosting Panel.

## 1. JWT Token Security Enhancement

### Changes Made

#### Added Fields to JWTClaims
- **JTI (JWT ID)**: Unique identifier for each token, enabling token revocation
- **DeviceID**: Optional device binding for additional security

#### Enhanced Token Generation
- `GenerateTokenWithDevice(user, deviceID)`: Creates tokens bound to specific devices
- Automatic JTI generation using UUID for each token
- Redis-based token storage for revocation capability
- Improved error handling with graceful degradation when Redis is unavailable

#### Enhanced Token Validation
- **Algorithm Confusion Prevention**: Explicitly verifies HMAC signing method
- **Token Revocation Support**: Checks Redis for revoked tokens
- **Better Error Messages**: Distinguishes between expired, invalid, and revoked tokens

#### New Token Management Methods
- `RevokeToken(jti)`: Revoke a specific token by its JTI
- `RevokeAllUserTokens(userID)`: Revoke all tokens for a specific user (useful for logout, security events)

#### Backward Compatibility
- `NewJWTManagerWithoutRedis()`: Constructor for scenarios without Redis
- Graceful degradation when Redis is unavailable (logs errors but doesn't fail requests)

### Security Benefits

1. **Token Revocation**: Enables immediate invalidation of compromised tokens
2. **Device Binding**: Prevents token theft and replay attacks across different devices
3. **Algorithm Confusion Prevention**: Protects against JWT algorithm confusion attacks
4. **Unique Token Tracking**: Each token has a unique ID for audit and revocation purposes
5. **Not Before Claims**: Prevents premature token usage

### Usage Examples

```go
// Initialize with Redis support
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
jwtManager := auth.NewJWTManager(
    "your-secret-key",
    24*time.Hour,
    7*24*time.Hour,
    redisClient,
)

// Generate token with device binding
token, err := jwtManager.GenerateTokenWithDevice(user, "device-12345")

// Validate token (automatically checks revocation)
claims, err := jwtManager.ValidateToken(token)

// Revoke a specific token
err = jwtManager.RevokeToken(claims.JTI)

// Revoke all tokens for a user (e.g., on password change)
err = jwtManager.RevokeAllUserTokens(userID)
```

## 2. Enhanced Distributed Rate Limiting

### Changes Made

#### Multi-Tier Rate Limiting
Implements three tiers of rate limiting with different strictness levels:

1. **IP-Based Rate Limiting** (60 requests/minute)
   - Applied to all requests
   - Most restrictive tier
   - Protects against IP-based attacks

2. **User-Based Rate Limiting** (120 requests/minute)
   - Applied to authenticated users
   - More permissive than IP-based
   - Encourages authentication

3. **Endpoint-Specific Rate Limiting** (10 requests/hour)
   - Applied to expensive operations
   - Prevents resource exhaustion
   - Protects critical endpoints

#### Rate Limit Headers
Compliant with standard rate limit headers:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in current window
- `Retry-After`: Seconds to wait before retrying (on 429 response)

### Usage Examples

```go
// Initialize enhanced rate limiter
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
rateLimiter := middleware.NewEnhancedRateLimiter(redisClient)

// Apply general rate limiting
app.Use(rateLimiter.Middleware())

// Apply stricter rate limiting to auth endpoints
authGroup := app.Group("/auth")
authGroup.Use(rateLimiter.AuthMiddleware())
```

## Testing

### Run JWT Tests
```bash
go test -v ./internal/auth/...
```

### Run Rate Limiting Tests (requires Redis)
```bash
# Start Redis first
docker run -d -p 6379:6379 redis:7

# Run tests
go test -v ./internal/middleware/ratelimit_enhanced_test.go ./internal/middleware/ratelimit_enhanced.go
```

## Files Changed

### Modified Files
- `internal/auth/jwt.go` - Enhanced JWT implementation
- `internal/auth/errors.go` - Added ErrTokenRevoked
- `go.mod` / `go.sum` - Added redis_rate dependency

### New Files
- `internal/middleware/ratelimit_enhanced.go` - Enhanced rate limiting
- `internal/auth/jwt_test.go` - JWT tests
- `internal/middleware/ratelimit_enhanced_test.go` - Rate limiting tests
- `docs/JWT_RATE_LIMITING_ENHANCEMENTS.md` - This documentation
