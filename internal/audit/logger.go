package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// AuditLogger provides comprehensive security audit logging
type AuditLogger struct {
	db            *sqlx.DB
	redis         *redis.Client
	enableRedis   bool
	retention     time.Duration
	asyncBuffer   chan AuditEvent
	stopChan      chan struct{}
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	ID           string                 `json:"id" db:"id"`
	TenantID     *string                `json:"tenant_id,omitempty" db:"tenant_id"`
	UserID       *int64                 `json:"user_id,omitempty" db:"user_id"`
	Action       string                 `json:"action" db:"action"`
	Resource     string                 `json:"resource" db:"resource"`
	ResourceID   *string                `json:"resource_id,omitempty" db:"resource_id"`
	Status       string                 `json:"status" db:"status"`
	IP           string                 `json:"ip" db:"ip"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	Method       string                 `json:"method" db:"method"`
	Path         string                 `json:"path" db:"path"`
	StatusCode   int                    `json:"status_code" db:"status_code"`
	Duration     int64                  `json:"duration" db:"duration"`
	Error        *string                `json:"error,omitempty" db:"error"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
}

// AuditEventType defines types of audit events
type AuditEventType string

const (
	EventTypeAuthentication AuditEventType = "authentication"
	EventTypeAuthorization  AuditEventType = "authorization"
	EventTypeDataAccess     AuditEventType = "data_access"
	EventTypeDataModify     AuditEventType = "data_modify"
	EventTypeSystemConfig   AuditEventType = "system_config"
	EventTypeSecurityChange AuditEventType = "security_change"
	EventTypeAPICall        AuditEventType = "api_call"
	EventTypeError          AuditEventType = "error"
)

// AuditStatus defines the status of an audit event
type AuditStatus string

const (
	StatusSuccess AuditStatus = "success"
	StatusFailure AuditStatus = "failure"
	StatusDenied  AuditStatus = "denied"
	StatusError   AuditStatus = "error"
)

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sqlx.DB, redis *redis.Client, enableRedis bool) *AuditLogger {
	al := &AuditLogger{
		db:          db,
		redis:       redis,
		enableRedis: enableRedis,
		retention:   90 * 24 * time.Hour, // 90 days default retention
		asyncBuffer: make(chan AuditEvent, 1000),
		stopChan:    make(chan struct{}),
	}

	// Start async processing goroutine
	go al.processEvents()

	// Start cleanup goroutine
	go al.cleanupOldEvents()

	return al
}

// Log logs an audit event
func (al *AuditLogger) Log(event AuditEvent) {
	// Generate ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("audit_%d_%s", time.Now().UnixNano(), generateRandomString(8))
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Send to async buffer
	select {
	case al.asyncBuffer <- event:
		// Event sent successfully
	default:
		// Buffer full, log synchronously
		log.Warn().Msg("Audit buffer full, logging synchronously")
		al.logSync(event)
	}
}

// LogFromContext extracts information from Fiber context and logs
func (al *AuditLogger) LogFromContext(c *fiber.Ctx, action string, resource string, status AuditStatus, metadata map[string]interface{}) {
	event := AuditEvent{
		Action:     action,
		Resource:   resource,
		Status:     string(status),
		IP:         c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Method:     c.Method(),
		Path:       c.Path(),
		StatusCode: c.Response().StatusCode(),
		Metadata:   metadata,
	}

	// Extract tenant and user ID from context if available
	if tenantID, ok := c.Locals("tenant_id").(string); ok {
		event.TenantID = &tenantID
	}

	if userID, ok := c.Locals("user_id").(int64); ok {
		event.UserID = &userID
	}

	al.Log(event)
}

// LogAuthAttempt logs an authentication attempt
func (al *AuditLogger) LogAuthAttempt(c *fiber.Ctx, email string, success bool, reason string) {
	status := StatusSuccess
	var errorMsg *string

	if !success {
		status = StatusFailure
		errorMsg = &reason
	}

	metadata := map[string]interface{}{
		"email": email,
	}

	if !success {
		metadata["reason"] = reason
	}

	event := AuditEvent{
		Action:     "login",
		Resource:   "authentication",
		Status:     string(status),
		IP:         c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Method:     c.Method(),
		Path:       c.Path(),
		StatusCode: c.Response().StatusCode(),
		Error:      errorMsg,
		Metadata:   metadata,
	}

	al.Log(event)
}

// LogAccessDenied logs an access denied event
func (al *AuditLogger) LogAccessDenied(c *fiber.Ctx, resource string, reason string) {
	event := AuditEvent{
		Action:     "access",
		Resource:   resource,
		Status:     string(StatusDenied),
		IP:         c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Method:     c.Method(),
		Path:       c.Path(),
		StatusCode: c.Response().StatusCode(),
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}

	// Extract user context
	if tenantID, ok := c.Locals("tenant_id").(string); ok {
		event.TenantID = &tenantID
	}
	if userID, ok := c.Locals("user_id").(int64); ok {
		event.UserID = &userID
	}

	al.Log(event)
}

// LogSecurityEvent logs a security-related event
func (al *AuditLogger) LogSecurityEvent(eventType string, severity string, message string, metadata map[string]interface{}) {
	event := AuditEvent{
		Action:   eventType,
		Resource: "security",
		Status:   severity,
		Metadata: metadata,
	}

	// Add severity to metadata
	if event.Metadata == nil {
		event.Metadata = make(map[string]interface{})
	}
	event.Metadata["severity"] = severity
	event.Metadata["message"] = message

	al.Log(event)
}

// processEvents processes audit events asynchronously
func (al *AuditLogger) processEvents() {
	for {
		select {
		case event := <-al.asyncBuffer:
			al.logSync(event)
		case <-al.stopChan:
			// Drain remaining events
			for len(al.asyncBuffer) > 0 {
				event := <-al.asyncBuffer
				al.logSync(event)
			}
			return
		}
	}
}

// logSync logs an event synchronously
func (al *AuditLogger) logSync(event AuditEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Log to database
	if err := al.logToDatabase(ctx, event); err != nil {
		log.Error().
			Err(err).
			Str("event_id", event.ID).
			Msg("Failed to log audit event to database")
	}

	// Log to Redis if enabled
	if al.enableRedis && al.redis != nil {
		if err := al.logToRedis(ctx, event); err != nil {
			log.Error().
				Err(err).
				Str("event_id", event.ID).
				Msg("Failed to log audit event to Redis")
		}
	}

	// Log to application logger for critical events
	if event.Status == string(StatusFailure) || event.Status == string(StatusDenied) || event.Status == string(StatusError) {
		log.Warn().
			Str("event_id", event.ID).
			Str("action", event.Action).
			Str("resource", event.Resource).
			Str("status", event.Status).
			Str("ip", event.IP).
			Interface("metadata", event.Metadata).
			Msg("Security audit event")
	}
}

// logToDatabase logs an event to PostgreSQL
func (al *AuditLogger) logToDatabase(ctx context.Context, event AuditEvent) error {
	// Serialize metadata to JSON
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := `
		INSERT INTO audit_logs (
			id, tenant_id, user_id, action, resource, resource_id,
			status, ip, user_agent, method, path, status_code,
			duration, error, metadata, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err = al.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.UserID, event.Action,
		event.Resource, event.ResourceID, event.Status, event.IP,
		event.UserAgent, event.Method, event.Path, event.StatusCode,
		event.Duration, event.Error, metadataJSON, event.Timestamp,
	)

	return err
}

// logToRedis logs an event to Redis for real-time monitoring
func (al *AuditLogger) logToRedis(ctx context.Context, event AuditEvent) error {
	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Store in Redis list (recent events)
	key := "audit:recent"
	pipe := al.redis.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 999) // Keep last 1000 events
	pipe.Expire(ctx, key, 24*time.Hour)

	// Store by IP for monitoring
	if event.IP != "" {
		ipKey := fmt.Sprintf("audit:ip:%s", event.IP)
		pipe.LPush(ctx, ipKey, data)
		pipe.LTrim(ctx, ipKey, 0, 99) // Keep last 100 events per IP
		pipe.Expire(ctx, ipKey, 24*time.Hour)
	}

	// Store failed attempts for alerting
	if event.Status == string(StatusFailure) || event.Status == string(StatusDenied) {
		failKey := "audit:failures"
		pipe.LPush(ctx, failKey, data)
		pipe.LTrim(ctx, failKey, 0, 499) // Keep last 500 failures
		pipe.Expire(ctx, failKey, 24*time.Hour)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// cleanupOldEvents removes old audit events based on retention policy
func (al *AuditLogger) cleanupOldEvents() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			cutoff := time.Now().Add(-al.retention)

			query := `DELETE FROM audit_logs WHERE timestamp < $1`
			result, err := al.db.ExecContext(ctx, query, cutoff)
			cancel()

			if err != nil {
				log.Error().
					Err(err).
					Msg("Failed to cleanup old audit events")
			} else {
				rowsAffected, _ := result.RowsAffected()
				log.Info().
					Int64("deleted", rowsAffected).
					Time("cutoff", cutoff).
					Msg("Cleaned up old audit events")
			}

		case <-al.stopChan:
			return
		}
	}
}

// Query queries audit events with filters
func (al *AuditLogger) Query(ctx context.Context, filters map[string]interface{}, limit int, offset int) ([]AuditEvent, error) {
	query := `SELECT * FROM audit_logs WHERE 1=1`
	args := make([]interface{}, 0)
	argIndex := 1

	// Build dynamic query based on filters
	if tenantID, ok := filters["tenant_id"].(string); ok {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, tenantID)
		argIndex++
	}

	if userID, ok := filters["user_id"].(int64); ok {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if action, ok := filters["action"].(string); ok {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	if resource, ok := filters["resource"].(string); ok {
		query += fmt.Sprintf(" AND resource = $%d", argIndex)
		args = append(args, resource)
		argIndex++
	}

	if status, ok := filters["status"].(string); ok {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if ip, ok := filters["ip"].(string); ok {
		query += fmt.Sprintf(" AND ip = $%d", argIndex)
		args = append(args, ip)
		argIndex++
	}

	if from, ok := filters["from"].(time.Time); ok {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, from)
		argIndex++
	}

	if to, ok := filters["to"].(time.Time); ok {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, to)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY timestamp DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var events []AuditEvent
	err := al.db.SelectContext(ctx, &events, query, args...)

	return events, err
}

// GetRecentFailures returns recent failed authentication attempts
func (al *AuditLogger) GetRecentFailures(ctx context.Context, limit int) ([]AuditEvent, error) {
	query := `
		SELECT * FROM audit_logs
		WHERE status IN ('failure', 'denied')
		ORDER BY timestamp DESC
		LIMIT $1
	`

	var events []AuditEvent
	err := al.db.SelectContext(ctx, &events, query, limit)

	return events, err
}

// GetSuspiciousActivity detects suspicious activity patterns
func (al *AuditLogger) GetSuspiciousActivity(ctx context.Context, lookbackHours int) ([]map[string]interface{}, error) {
	query := `
		SELECT ip, COUNT(*) as attempt_count, MAX(timestamp) as last_attempt
		FROM audit_logs
		WHERE status IN ('failure', 'denied')
		AND timestamp > NOW() - INTERVAL '1 hour' * $1
		GROUP BY ip
		HAVING COUNT(*) > 10
		ORDER BY attempt_count DESC
	`

	rows, err := al.db.QueryContext(ctx, query, lookbackHours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var ip string
		var count int
		var lastAttempt time.Time

		if err := rows.Scan(&ip, &count, &lastAttempt); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"ip":           ip,
			"attempts":     count,
			"last_attempt": lastAttempt,
		})
	}

	return results, nil
}

// Close stops the audit logger
func (al *AuditLogger) Close() {
	close(al.stopChan)
	log.Info().Msg("Audit logger stopped")
}

// SetRetention sets the retention period for audit logs
func (al *AuditLogger) SetRetention(retention time.Duration) {
	al.retention = retention
	log.Info().
		Dur("retention", retention).
		Msg("Audit log retention updated")
}

// generateRandomString generates a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Middleware returns Fiber middleware for automatic audit logging
func (al *AuditLogger) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Determine status
		status := StatusSuccess
		var errorMsg *string

		if err != nil {
			status = StatusError
			msg := err.Error()
			errorMsg = &msg
		} else if c.Response().StatusCode() >= 400 {
			status = StatusFailure
		}

		// Log event
		event := AuditEvent{
			Action:     c.Method(),
			Resource:   c.Path(),
			Status:     string(status),
			IP:         c.IP(),
			UserAgent:  c.Get("User-Agent"),
			Method:     c.Method(),
			Path:       c.Path(),
			StatusCode: c.Response().StatusCode(),
			Duration:   duration.Milliseconds(),
			Error:      errorMsg,
		}

		// Extract context information
		if tenantID, ok := c.Locals("tenant_id").(string); ok {
			event.TenantID = &tenantID
		}
		if userID, ok := c.Locals("user_id").(int64); ok {
			event.UserID = &userID
		}

		al.Log(event)

		return err
	}
}
