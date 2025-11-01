package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/rs/zerolog/log"
)

// PerformanceMiddleware configures performance optimizations
func PerformanceMiddleware() []fiber.Handler {
	return []fiber.Handler{
		// ETag for conditional requests
		etag.New(etag.Config{
			Weak: true,
		}),
		
		// Compression for response bodies
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed, // Balance between speed and compression ratio
			Next: func(c *fiber.Ctx) bool {
				// Skip compression for already compressed content
				contentType := c.Get("Content-Type")
				return strings.Contains(contentType, "gzip") ||
					strings.Contains(contentType, "deflate") ||
					strings.Contains(contentType, "br")
			},
		}),
		
		// Cache middleware for static responses
		cache.New(cache.Config{
			Next: func(c *fiber.Ctx) bool {
				// Only cache GET requests
				return c.Method() != "GET"
			},
			Expiration:   30 * time.Second,
			CacheControl: true,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.Path() + "?" + c.Request().URI().QueryArgs().String()
			},
			Storage: nil, // Use in-memory cache, can be replaced with Redis
		}),
		
		// Custom performance logging
		performanceLogger(),
	}
}

// performanceLogger logs slow requests and performance metrics
func performanceLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Continue with request
		err := c.Next()
		
		// Calculate request duration
		duration := time.Since(start)
		
		// Log slow requests (>500ms)
		if duration > 500*time.Millisecond {
			log.Warn().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Str("ip", c.IP()).
				Dur("duration", duration).
				Int("status", c.Response().StatusCode()).
				Msg("Slow request detected")
		}
		
		// Log very slow requests (>2s) as errors
		if duration > 2*time.Second {
			log.Error().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Str("ip", c.IP()).
				Dur("duration", duration).
				Int("status", c.Response().StatusCode()).
				Msg("Very slow request - investigate performance issue")
		}
		
		return err
	}
}

// DatabaseOptimization provides database performance optimizations
type DatabaseOptimization struct {
	connectionPool *ConnectionPool
	queryCache     *QueryCache
}

// ConnectionPool manages database connection pooling
type ConnectionPool struct {
	MaxConnections     int
	MaxIdleConnections int
	MaxLifetime        time.Duration
	IdleTimeout        time.Duration
}

// OptimizedConnectionPool returns optimized connection pool settings
func OptimizedConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		MaxConnections:     50,  // Increase from default 25
		MaxIdleConnections: 15,  // Increase from default 10
		MaxLifetime:        2 * time.Hour, // Increase from default 1 hour
		IdleTimeout:        30 * time.Minute,
	}
}

// QueryCache provides in-memory query result caching
type QueryCache struct {
	cache map[string]CacheEntry
	ttl   time.Duration
}

// CacheEntry represents a cached query result
type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
}

// NewQueryCache creates a new query cache
func NewQueryCache(ttl time.Duration) *QueryCache {
	cache := &QueryCache{
		cache: make(map[string]CacheEntry),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a cached entry
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	entry, exists := qc.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check if entry has expired
	if time.Since(entry.Timestamp) > qc.ttl {
		delete(qc.cache, key)
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores a cache entry
func (qc *QueryCache) Set(key string, data interface{}) {
	qc.cache[key] = CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
	}
}

// cleanup removes expired cache entries
func (qc *QueryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		for key, entry := range qc.cache {
			if now.Sub(entry.Timestamp) > qc.ttl {
				delete(qc.cache, key)
			}
		}
	}
}

// ResponseOptimizer optimizes HTTP responses
type ResponseOptimizer struct {
	compressionLevel int
	minCompressionSize int
}

// NewResponseOptimizer creates a new response optimizer
func NewResponseOptimizer() *ResponseOptimizer {
	return &ResponseOptimizer{
		compressionLevel:   gzip.BestSpeed, // Balance speed vs compression
		minCompressionSize: 1024,          // Only compress responses > 1KB
	}
}

// OptimizeResponse applies response optimizations
func (ro *ResponseOptimizer) OptimizeResponse(c *fiber.Ctx) error {
	// Set caching headers for static content
	if strings.HasPrefix(c.Path(), "/static/") {
		c.Set("Cache-Control", "public, max-age=31536000") // 1 year
		c.Set("Expires", time.Now().Add(365*24*time.Hour).Format(time.RFC1123))
	}
	
	// Set caching headers for API responses
	if strings.HasPrefix(c.Path(), "/api/") {
		switch c.Method() {
		case "GET":
			// Cache GET requests for 5 minutes
			c.Set("Cache-Control", "private, max-age=300")
		case "POST", "PUT", "DELETE":
			// Don't cache mutations
			c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
	}
	
	return c.Next()
}

// MemoryOptimizer provides memory usage optimizations
type MemoryOptimizer struct {
	maxResponseSize int64
	poolSize        int
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		maxResponseSize: 10 * 1024 * 1024, // 10MB max response size
		poolSize:        100,               // Buffer pool size
	}
}

// OptimizeMemory provides memory optimization middleware
func (mo *MemoryOptimizer) OptimizeMemory() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Limit request body size
		c.Request().SetBodyStream(io.LimitReader(
			c.Request().BodyStream(),
			mo.maxResponseSize,
		), int(mo.maxResponseSize))

		return c.Next()
	}
}

// CPUOptimizer provides CPU usage optimizations
type CPUOptimizer struct {
	maxConcurrentRequests int
	requestQueue          chan struct{}
}

// NewCPUOptimizer creates a new CPU optimizer
func NewCPUOptimizer(maxConcurrent int) *CPUOptimizer {
	return &CPUOptimizer{
		maxConcurrentRequests: maxConcurrent,
		requestQueue:          make(chan struct{}, maxConcurrent),
	}
}

// OptimizeCPU provides CPU optimization middleware
func (co *CPUOptimizer) OptimizeCPU() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Acquire semaphore
		co.requestQueue <- struct{}{}
		defer func() {
			<-co.requestQueue
		}()
		
		return c.Next()
	}
}

// NetworkOptimizer provides network optimizations
type NetworkOptimizer struct {
	keepAliveTimeout time.Duration
	readTimeout      time.Duration
	writeTimeout     time.Duration
}

// NewNetworkOptimizer creates a new network optimizer
func NewNetworkOptimizer() *NetworkOptimizer {
	return &NetworkOptimizer{
		keepAliveTimeout: 30 * time.Second,
		readTimeout:      15 * time.Second,
		writeTimeout:     15 * time.Second,
	}
}

// OptimizedFiberConfig returns optimized Fiber configuration
func (no *NetworkOptimizer) OptimizedFiberConfig() fiber.Config {
	return fiber.Config{
		// Network optimizations
		ReadTimeout:  no.readTimeout,
		WriteTimeout: no.writeTimeout,
		IdleTimeout:  no.keepAliveTimeout,
		
		// Performance optimizations
		Prefork:                 false, // Set to true in production with load balancer
		DisableKeepalive:        false,
		DisableDefaultDate:      false,
		DisableDefaultContentType: false,
		DisableHeaderNormalizing: false,
		DisableStartupMessage:   false,
		
		// Memory optimizations
		BodyLimit:    10 * 1024 * 1024, // 10MB
		ReadBufferSize:  8192,          // 8KB
		WriteBufferSize: 8192,          // 8KB
		
		// Error handling
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			
			// Log performance-related errors
			if code >= 500 {
				log.Error().
					Err(err).
					Str("path", c.Path()).
					Str("method", c.Method()).
					Int("status", code).
					Msg("Server error occurred")
			}
			
			return c.Status(code).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	}
}

// PerformanceMonitor tracks application performance metrics
type PerformanceMonitor struct {
	requestCount    int64
	totalDuration   time.Duration
	slowRequests    int64
	errorCount      int64
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	monitor := &PerformanceMonitor{}
	
	// Start metrics reporting goroutine
	go monitor.reportMetrics()
	
	return monitor
}

// Monitor provides performance monitoring middleware
func (pm *PerformanceMonitor) Monitor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		err := c.Next()
		
		duration := time.Since(start)
		pm.requestCount++
		pm.totalDuration += duration
		
		if duration > 500*time.Millisecond {
			pm.slowRequests++
		}
		
		if c.Response().StatusCode() >= 500 {
			pm.errorCount++
		}
		
		return err
	}
}

// reportMetrics logs performance metrics periodically
func (pm *PerformanceMonitor) reportMetrics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		if pm.requestCount > 0 {
			avgDuration := pm.totalDuration / time.Duration(pm.requestCount)
			
			log.Info().
				Int64("total_requests", pm.requestCount).
				Dur("avg_duration", avgDuration).
				Int64("slow_requests", pm.slowRequests).
				Int64("error_count", pm.errorCount).
				Float64("error_rate", float64(pm.errorCount)/float64(pm.requestCount)*100).
				Msg("Performance metrics")
		}
	}
}

// OptimizationRecommendations provides performance optimization recommendations
func OptimizationRecommendations() map[string]string {
	return map[string]string{
		"database": "Use connection pooling, implement query caching, add database indexes, use prepared statements",
		"caching": "Implement Redis caching for frequently accessed data, use HTTP caching headers, implement response caching",
		"compression": "Enable gzip compression for responses, optimize static asset delivery, use CDN for static content",
		"monitoring": "Implement APM tools, monitor slow queries, track memory usage, monitor CPU utilization",
		"scaling": "Implement horizontal scaling, use load balancing, consider database sharding, implement auto-scaling",
		"security": "Use rate limiting, implement security headers, validate all inputs, use secure session management",
		"optimization": "Profile code regularly, optimize database queries, minimize memory allocations, use efficient algorithms",
	}
}