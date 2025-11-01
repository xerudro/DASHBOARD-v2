package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// PoolOptimizer provides advanced connection pool management
type PoolOptimizer struct {
	db                  *sqlx.DB
	preparedStmtCache   map[string]*sqlx.Stmt
	preparedStmtMutex   sync.RWMutex
	healthCheckInterval time.Duration
	connectionTimeout   time.Duration
	queryTimeout        time.Duration
	slowQueryThreshold  time.Duration
	maxRetries          int
	retryDelay          time.Duration
	metrics             *PoolMetrics
	stopHealthCheck     chan struct{}
}

// PoolMetrics tracks connection pool metrics
type PoolMetrics struct {
	mu                    sync.RWMutex
	TotalQueries          int64
	SlowQueries           int64
	FailedQueries         int64
	AvgQueryDuration      time.Duration
	totalQueryDuration    time.Duration
	PreparedStmtCacheHits int64
	PreparedStmtCacheMiss int64
	ConnectionErrors      int64
	Retries               int64
}

// NewPoolOptimizer creates an optimized connection pool
func NewPoolOptimizer(db *sqlx.DB) *PoolOptimizer {
	po := &PoolOptimizer{
		db:                  db,
		preparedStmtCache:   make(map[string]*sqlx.Stmt),
		healthCheckInterval: 30 * time.Second,
		connectionTimeout:   10 * time.Second,
		queryTimeout:        30 * time.Second,
		slowQueryThreshold:  1 * time.Second,
		maxRetries:          3,
		retryDelay:          100 * time.Millisecond,
		metrics:             &PoolMetrics{},
		stopHealthCheck:     make(chan struct{}),
	}

	// Start health check goroutine
	go po.healthCheckLoop()

	// Start metrics reporting
	go po.metricsReportLoop()

	return po
}

// GetPreparedStmt returns a prepared statement from cache or creates a new one
func (po *PoolOptimizer) GetPreparedStmt(ctx context.Context, query string) (*sqlx.Stmt, error) {
	// Check cache first
	po.preparedStmtMutex.RLock()
	stmt, exists := po.preparedStmtCache[query]
	po.preparedStmtMutex.RUnlock()

	if exists {
		po.metrics.mu.Lock()
		po.metrics.PreparedStmtCacheHits++
		po.metrics.mu.Unlock()
		return stmt, nil
	}

	// Cache miss - prepare new statement
	po.metrics.mu.Lock()
	po.metrics.PreparedStmtCacheMiss++
	po.metrics.mu.Unlock()

	// Create context with timeout
	prepareCtx, cancel := context.WithTimeout(ctx, po.connectionTimeout)
	defer cancel()

	stmt, err := po.db.PreparexContext(prepareCtx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Add to cache
	po.preparedStmtMutex.Lock()
	po.preparedStmtCache[query] = stmt
	po.preparedStmtMutex.Unlock()

	return stmt, nil
}

// QueryWithContext executes a query with context, timeout, and retry logic
func (po *PoolOptimizer) QueryWithContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		po.recordQuery(duration)

		if duration > po.slowQueryThreshold {
			log.Warn().
				Dur("duration", duration).
				Str("query", query).
				Msg("Slow query detected")
		}
	}()

	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, po.queryTimeout)
	defer cancel()

	var rows *sqlx.Rows
	var err error

	// Retry logic for transient errors
	for attempt := 0; attempt <= po.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := po.retryDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)

			po.metrics.mu.Lock()
			po.metrics.Retries++
			po.metrics.mu.Unlock()

			log.Warn().
				Int("attempt", attempt).
				Str("query", query).
				Msg("Retrying query")
		}

		rows, err = po.db.QueryxContext(queryCtx, query, args...)
		if err == nil {
			return rows, nil
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			break
		}

		// Check if context is still valid
		if queryCtx.Err() != nil {
			break
		}
	}

	po.metrics.mu.Lock()
	po.metrics.FailedQueries++
	po.metrics.mu.Unlock()

	return nil, fmt.Errorf("query failed after %d attempts: %w", po.maxRetries+1, err)
}

// GetWithContext executes a GET query with optimization
func (po *PoolOptimizer) GetWithContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		po.recordQuery(duration)
	}()

	queryCtx, cancel := context.WithTimeout(ctx, po.queryTimeout)
	defer cancel()

	err := po.db.GetContext(queryCtx, dest, query, args...)
	if err != nil {
		po.metrics.mu.Lock()
		po.metrics.FailedQueries++
		po.metrics.mu.Unlock()
	}

	return err
}

// SelectWithContext executes a SELECT query with optimization
func (po *PoolOptimizer) SelectWithContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		po.recordQuery(duration)
	}()

	queryCtx, cancel := context.WithTimeout(ctx, po.queryTimeout)
	defer cancel()

	err := po.db.SelectContext(queryCtx, dest, query, args...)
	if err != nil {
		po.metrics.mu.Lock()
		po.metrics.FailedQueries++
		po.metrics.mu.Unlock()
	}

	return err
}

// ExecWithContext executes a mutation query with optimization
func (po *PoolOptimizer) ExecWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		po.recordQuery(duration)
	}()

	queryCtx, cancel := context.WithTimeout(ctx, po.queryTimeout)
	defer cancel()

	result, err := po.db.ExecContext(queryCtx, query, args...)
	if err != nil {
		po.metrics.mu.Lock()
		po.metrics.FailedQueries++
		po.metrics.mu.Unlock()
		return nil, err
	}

	return result, nil
}

// TransactionWithContext executes a transaction with optimization
func (po *PoolOptimizer) TransactionWithContext(ctx context.Context, fn func(*sqlx.Tx) error) error {
	txCtx, cancel := context.WithTimeout(ctx, po.queryTimeout*2) // Give more time for transactions
	defer cancel()

	tx, err := po.db.BeginTxx(txCtx, nil)
	if err != nil {
		po.metrics.mu.Lock()
		po.metrics.ConnectionErrors++
		po.metrics.mu.Unlock()
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error().
				Err(rbErr).
				Msg("Failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// healthCheckLoop periodically checks connection pool health
func (po *PoolOptimizer) healthCheckLoop() {
	ticker := time.NewTicker(po.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := po.db.PingContext(ctx)
			cancel()

			if err != nil {
				po.metrics.mu.Lock()
				po.metrics.ConnectionErrors++
				po.metrics.mu.Unlock()

				log.Error().
					Err(err).
					Msg("Database health check failed")
			}

			// Log pool statistics
			stats := po.db.Stats()
			log.Debug().
				Int("open_connections", stats.OpenConnections).
				Int("in_use", stats.InUse).
				Int("idle", stats.Idle).
				Int64("wait_count", stats.WaitCount).
				Dur("wait_duration", stats.WaitDuration).
				Msg("Connection pool statistics")

		case <-po.stopHealthCheck:
			return
		}
	}
}

// metricsReportLoop periodically reports pool metrics
func (po *PoolOptimizer) metricsReportLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			po.metrics.mu.RLock()
			if po.metrics.TotalQueries > 0 {
				cacheHitRate := float64(po.metrics.PreparedStmtCacheHits) / float64(po.metrics.PreparedStmtCacheHits+po.metrics.PreparedStmtCacheMiss) * 100
				failureRate := float64(po.metrics.FailedQueries) / float64(po.metrics.TotalQueries) * 100

				log.Info().
					Int64("total_queries", po.metrics.TotalQueries).
					Int64("slow_queries", po.metrics.SlowQueries).
					Int64("failed_queries", po.metrics.FailedQueries).
					Float64("failure_rate", failureRate).
					Dur("avg_duration", po.metrics.AvgQueryDuration).
					Float64("cache_hit_rate", cacheHitRate).
					Int64("retries", po.metrics.Retries).
					Msg("Database pool metrics")
			}
			po.metrics.mu.RUnlock()

		case <-po.stopHealthCheck:
			return
		}
	}
}

// recordQuery records query metrics
func (po *PoolOptimizer) recordQuery(duration time.Duration) {
	po.metrics.mu.Lock()
	defer po.metrics.mu.Unlock()

	po.metrics.TotalQueries++
	po.metrics.totalQueryDuration += duration

	if po.metrics.TotalQueries > 0 {
		po.metrics.AvgQueryDuration = po.metrics.totalQueryDuration / time.Duration(po.metrics.TotalQueries)
	}

	if duration > po.slowQueryThreshold {
		po.metrics.SlowQueries++
	}
}

// GetMetrics returns current pool metrics
func (po *PoolOptimizer) GetMetrics() PoolMetrics {
	po.metrics.mu.RLock()
	defer po.metrics.mu.RUnlock()

	// Return a copy to avoid returning locked value
	return PoolMetrics{
		TotalQueries:          po.metrics.TotalQueries,
		SlowQueries:           po.metrics.SlowQueries,
		FailedQueries:         po.metrics.FailedQueries,
		AvgQueryDuration:      po.metrics.AvgQueryDuration,
		totalQueryDuration:    po.metrics.totalQueryDuration,
		PreparedStmtCacheHits: po.metrics.PreparedStmtCacheHits,
		PreparedStmtCacheMiss: po.metrics.PreparedStmtCacheMiss,
		ConnectionErrors:      po.metrics.ConnectionErrors,
		Retries:               po.metrics.Retries,
	}
}

// ClearPreparedStatements clears the prepared statement cache
func (po *PoolOptimizer) ClearPreparedStatements() error {
	po.preparedStmtMutex.Lock()
	defer po.preparedStmtMutex.Unlock()

	for query, stmt := range po.preparedStmtCache {
		if err := stmt.Close(); err != nil {
			log.Error().
				Err(err).
				Str("query", query).
				Msg("Failed to close prepared statement")
		}
	}

	po.preparedStmtCache = make(map[string]*sqlx.Stmt)

	log.Info().Msg("Prepared statement cache cleared")
	return nil
}

// Close closes the pool optimizer and cleans up resources
func (po *PoolOptimizer) Close() error {
	// Stop health check loop
	close(po.stopHealthCheck)

	// Clear prepared statements
	if err := po.ClearPreparedStatements(); err != nil {
		return err
	}

	log.Info().Msg("Pool optimizer closed")
	return nil
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable errors
	errStr := err.Error()

	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"temporary failure",
		"too many connections",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}

// SetQueryTimeout sets the query timeout
func (po *PoolOptimizer) SetQueryTimeout(timeout time.Duration) {
	po.queryTimeout = timeout
	log.Info().
		Dur("timeout", timeout).
		Msg("Query timeout updated")
}

// SetSlowQueryThreshold sets the slow query threshold
func (po *PoolOptimizer) SetSlowQueryThreshold(threshold time.Duration) {
	po.slowQueryThreshold = threshold
	log.Info().
		Dur("threshold", threshold).
		Msg("Slow query threshold updated")
}

// SetMaxRetries sets the maximum number of retries
func (po *PoolOptimizer) SetMaxRetries(maxRetries int) {
	po.maxRetries = maxRetries
	log.Info().
		Int("max_retries", maxRetries).
		Msg("Max retries updated")
}
