# Quick Start Guide - Phase 3 Improvements

This guide helps you quickly integrate the new security and performance improvements into your VIP Hosting Panel v2 application.

---

## 🚀 5-Minute Integration

### 1. Add Imports

```go
import (
    "github.com/xerudro/DASHBOARD-v2/internal/middleware"
    "github.com/xerudro/DASHBOARD-v2/internal/database"
    "github.com/xerudro/DASHBOARD-v2/internal/cache"
    "github.com/xerudro/DASHBOARD-v2/internal/audit"
    "github.com/xerudro/DASHBOARD-v2/internal/shutdown"
)
```

### 2. Initialize Components (Add to main.go)

```go
// After database and Redis initialization

// 1. Redis Rate Limiter
rateLimiter := middleware.NewRedisRateLimiter(redis, 1000, time.Minute)

// 2. Request Validator
requestValidator := middleware.NewRequestValidator()

// 3. Audit Logger
auditLogger := audit.NewAuditLogger(db, redis, true)

// 4. Database Optimizer
dbOptimizer := database.NewPoolOptimizer(db)

// 5. Query Cache
queryCache := cache.NewRedisCache(redis, "app:", 5*time.Minute)

// 6. Graceful Shutdown
gracefulShutdown := shutdown.NewGracefulShutdown(30 * time.Second)
```

### 3. Apply Middleware (Add to Fiber app)

```go
// Order matters - apply in this sequence
app.Use(requestValidator.Middleware())   // First: Validate
app.Use(rateLimiter.Middleware())        // Second: Rate limit
app.Use(auditLogger.Middleware())        // Third: Audit log
// ... your other middleware ...
```

### 4. Register Shutdown Handlers

```go
gracefulShutdown.RegisterShutdownFunc("fiber", 40, 15*time.Second,
    shutdown.FiberShutdown(app))
gracefulShutdown.RegisterShutdownFunc("database", 70, 10*time.Second,
    shutdown.DatabaseShutdown(db.Close))
gracefulShutdown.RegisterShutdownFunc("cache", 60, 5*time.Second,
    shutdown.GenericShutdown("cache", func() error { return nil }))
```

### 5. Start Application with Graceful Shutdown

```go
// Start server in goroutine
go func() {
    if err := app.Listen(":8080"); err != nil {
        log.Fatal(err)
    }
}()

// Wait for shutdown signal
<-gracefulShutdown.Start()
```

---

## 📝 Common Use Cases

### Use Case 1: Rate Limit an Endpoint

```go
// Stricter rate limiting for sensitive endpoints
authGroup := app.Group("/api/auth")
authGroup.Use(rateLimiter.AuthMiddleware()) // 10 requests/minute
authGroup.Post("/login", handleLogin)
```

### Use Case 2: Log Security Event

```go
func handleLogin(c *fiber.Ctx) error {
    // ... login logic ...

    if !authenticated {
        auditLogger.LogAuthAttempt(c, email, false, "invalid password")
        return fiber.ErrUnauthorized
    }

    auditLogger.LogAuthAttempt(c, email, true, "")
    return c.JSON(response)
}
```

### Use Case 3: Cache Database Query

```go
func (r *ServerRepository) GetAll(ctx context.Context) ([]*models.Server, error) {
    var servers []*models.Server
    cacheKey := "servers:all"

    err := r.cache.GetOrSet(ctx, cacheKey, &servers, func() (interface{}, error) {
        // Fetch from database
        return r.db.Select(&servers, "SELECT * FROM servers")
    }, cache.CacheOptions{
        TTL: 5 * time.Minute,
        Tags: []string{"servers"},
    })

    return servers, err
}
```

### Use Case 4: Invalidate Cache on Update

```go
func (r *ServerRepository) Update(ctx context.Context, server *models.Server) error {
    err := r.db.Update(server)
    if err != nil {
        return err
    }

    // Invalidate cache
    r.cache.DeleteByTag(ctx, "servers")
    return nil
}
```

### Use Case 5: Optimized Database Query

```go
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User

    // Use optimized query with retry logic
    err := dbOptimizer.GetWithContext(ctx, &user,
        "SELECT * FROM users WHERE email = $1", email)

    return &user, err
}
```

---

## 🔧 Configuration Snippets

### Minimum Configuration

```yaml
# configs/config.yaml (add these sections)

rate_limit:
  enabled: true
  max_requests: 1000
  window: 60s

cache:
  enabled: true
  default_ttl: 300s

audit:
  enabled: true
  retention_days: 90
```

### Production Configuration

```yaml
rate_limit:
  enabled: true
  backend: redis
  max_requests: 5000
  window: 60s
  auth_max_requests: 10
  auth_window: 60s

cache:
  enabled: true
  backend: redis
  default_ttl: 300s
  max_memory: 500MB
  eviction_policy: allkeys-lru

database:
  max_connections: 100
  max_idle_connections: 25
  max_lifetime: 2h
  query_timeout: 30s
  slow_query_threshold: 1s
  enable_prepared_stmt_cache: true

audit:
  enabled: true
  log_to_redis: true
  retention_days: 90
  buffer_size: 5000
  sync_interval: 1s

shutdown:
  timeout: 60s
  drain_timeout: 30s
```

---

## 🐛 Troubleshooting

### Rate Limiter Not Working

```bash
# Check Redis connection
redis-cli PING

# Check rate limit keys
redis-cli KEYS "ratelimit:*"

# Check specific client
redis-cli GET "ratelimit:192.168.1.1:Mozilla"
```

### Cache Not Caching

```go
// Enable cache metrics logging
metrics := queryCache.GetMetrics()
log.Info().
    Int64("hits", metrics.Hits).
    Int64("misses", metrics.Misses).
    Float64("hit_rate", float64(metrics.Hits)/float64(metrics.Hits+metrics.Misses)*100).
    Msg("Cache metrics")
```

### Slow Queries Not Detected

```go
// Lower the threshold for testing
dbOptimizer.SetSlowQueryThreshold(100 * time.Millisecond)
```

### Audit Logs Not Appearing

```bash
# Check PostgreSQL
SELECT COUNT(*) FROM audit_logs;

# Check Redis
redis-cli LLEN "audit:recent"

# Check buffer
# Look for log messages: "Audit buffer full"
```

### Graceful Shutdown Not Working

```go
// Add debug logging
gracefulShutdown.RegisterShutdownFunc("test", 1, 5*time.Second,
    func(ctx context.Context) error {
        log.Info().Msg("Shutdown function called")
        return nil
    })
```

---

## 📊 Quick Health Check

```bash
# 1. Check rate limiting
curl -I http://localhost:8080/api/health
# Look for: X-RateLimit-Limit, X-RateLimit-Remaining headers

# 2. Check audit logs
curl http://localhost:8080/monitoring/audit/summary

# 3. Check cache metrics
curl http://localhost:8080/monitoring/cache/metrics

# 4. Check database metrics
curl http://localhost:8080/monitoring/database/metrics

# 5. Test graceful shutdown
kill -SIGTERM $(pgrep vip-panel)
# Watch logs for "Graceful shutdown completed"
```

---

## 🎯 Performance Testing

### Load Test with Rate Limiting

```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Test rate limiting (should see 429 errors after limit)
hey -n 2000 -c 50 http://localhost:8080/api/health

# Test with authentication
hey -n 1000 -c 20 -m POST \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"test123"}' \
    http://localhost:8080/api/auth/login
```

### Cache Performance Test

```bash
# First request (cache miss)
time curl http://localhost:8080/api/servers

# Second request (cache hit - should be much faster)
time curl http://localhost:8080/api/servers
```

---

## 🔍 Monitoring Commands

### Redis Monitoring

```bash
# Monitor Redis in real-time
redis-cli MONITOR

# Check memory usage
redis-cli INFO memory

# Check key statistics
redis-cli INFO keyspace
```

### Database Monitoring

```sql
-- Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
WHERE mean_exec_time > 1000
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check connection pool
SELECT count(*) as connections,
       state
FROM pg_stat_activity
GROUP BY state;
```

### Application Monitoring

```bash
# Watch application logs
tail -f /var/log/vip-panel/api.log

# Filter for security events
tail -f /var/log/vip-panel/api.log | grep "Security audit event"

# Filter for slow queries
tail -f /var/log/vip-panel/api.log | grep "Slow query detected"
```

---

## 🚨 Alerts to Set Up

### Critical Alerts

```yaml
alerts:
  - name: high_rate_limit_violations
    condition: rate_limit_violations > 100 per minute
    severity: critical

  - name: cache_hit_rate_low
    condition: cache_hit_rate < 50%
    severity: warning

  - name: slow_query_count_high
    condition: slow_queries > 10 per minute
    severity: warning

  - name: failed_auth_attempts
    condition: failed_auth > 50 per hour from same IP
    severity: critical

  - name: graceful_shutdown_timeout
    condition: shutdown_duration > 60s
    severity: warning
```

---

## 📚 Next Steps

1. ✅ Review [PHASE-3-IMPROVEMENTS-SUMMARY.md](PHASE-3-IMPROVEMENTS-SUMMARY.md) for detailed documentation
2. ✅ Check [.github/DEVELOPMENT-PROGRESS.md](.github/DEVELOPMENT-PROGRESS.md) for implementation status
3. ✅ Set up monitoring dashboards
4. ✅ Configure alerting rules
5. ✅ Run load tests in staging
6. ✅ Deploy to production with monitoring

---

## 💡 Pro Tips

1. **Start with conservative rate limits** - You can always increase them
2. **Monitor cache hit rates** - Aim for >80% for frequently accessed data
3. **Review audit logs weekly** - Look for suspicious patterns
4. **Test graceful shutdown** - Regularly verify zero-downtime deployments
5. **Use prepared statement cache** - Significant performance boost for repeated queries
6. **Tag your cache entries** - Makes invalidation much easier
7. **Set appropriate TTLs** - Balance freshness vs. performance
8. **Monitor Redis memory** - Set eviction policies appropriately

---

**Need Help?** Check the detailed documentation in PHASE-3-IMPROVEMENTS-SUMMARY.md or review the inline code comments in each module.
