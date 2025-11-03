package vault

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrSecretNotFound     = errors.New("secret not found")
	ErrSecretAlreadyExists = errors.New("secret already exists")
	ErrInvalidSecretPath   = errors.New("invalid secret path")
)

// Secret represents a stored secret with metadata
type Secret struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Path        string     `db:"path" json:"path"`                 // e.g., "database/postgres/password"
	Value       string     `db:"value" json:"-"`                   // Encrypted value (never in JSON)
	Description string     `db:"description" json:"description"`   // Human-readable description
	CreatedBy   uuid.UUID  `db:"created_by" json:"created_by"`     // User ID who created it
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	Version     int        `db:"version" json:"version"`           // Version number for rotation
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"` // Optional expiration
	Metadata    string     `db:"metadata" json:"metadata,omitempty"` // JSON metadata (tags, etc.)
}

// SecretVersion represents a historical version of a secret
type SecretVersion struct {
	ID        uuid.UUID `db:"id" json:"id"`
	SecretID  uuid.UUID `db:"secret_id" json:"secret_id"`
	Version   int       `db:"version" json:"version"`
	Value     string    `db:"value" json:"-"`
	UpdatedBy uuid.UUID `db:"updated_by" json:"updated_by"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// AuditLog represents an access log entry
type AuditLog struct {
	ID        uuid.UUID `db:"id" json:"id"`
	SecretID  uuid.UUID `db:"secret_id" json:"secret_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Action    string    `db:"action" json:"action"` // "read", "write", "delete", "rotate"
	IPAddress string    `db:"ip_address" json:"ip_address"`
	Success   bool      `db:"success" json:"success"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}

// StorageBackend handles database operations for secrets
type StorageBackend struct {
	db *sqlx.DB
}

// NewStorageBackend creates a new storage backend
func NewStorageBackend(db *sqlx.DB) (*StorageBackend, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	backend := &StorageBackend{db: db}

	// Initialize database schema
	if err := backend.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return backend, nil
}

// initSchema creates the necessary database tables
func (s *StorageBackend) initSchema() error {
	schema := `
	-- Secrets table
	CREATE TABLE IF NOT EXISTS vault_secrets (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		path VARCHAR(500) UNIQUE NOT NULL,
		value TEXT NOT NULL,
		description TEXT,
		created_by UUID NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		version INTEGER NOT NULL DEFAULT 1,
		expires_at TIMESTAMP,
		metadata TEXT,
		CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Secret versions table (for rotation history)
	CREATE TABLE IF NOT EXISTS vault_secret_versions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		secret_id UUID NOT NULL,
		version INTEGER NOT NULL,
		value TEXT NOT NULL,
		updated_by UUID NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_secret_id FOREIGN KEY (secret_id) REFERENCES vault_secrets(id) ON DELETE CASCADE,
		CONSTRAINT fk_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE (secret_id, version)
	);

	-- Audit log table
	CREATE TABLE IF NOT EXISTS vault_audit_logs (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		secret_id UUID,
		user_id UUID NOT NULL,
		action VARCHAR(50) NOT NULL,
		ip_address VARCHAR(45),
		success BOOLEAN NOT NULL DEFAULT TRUE,
		timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_audit_secret_id FOREIGN KEY (secret_id) REFERENCES vault_secrets(id) ON DELETE SET NULL,
		CONSTRAINT fk_audit_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_vault_secrets_path ON vault_secrets(path);
	CREATE INDEX IF NOT EXISTS idx_vault_secrets_created_by ON vault_secrets(created_by);
	CREATE INDEX IF NOT EXISTS idx_vault_secrets_expires_at ON vault_secrets(expires_at);
	CREATE INDEX IF NOT EXISTS idx_vault_secret_versions_secret_id ON vault_secret_versions(secret_id);
	CREATE INDEX IF NOT EXISTS idx_vault_audit_logs_secret_id ON vault_audit_logs(secret_id);
	CREATE INDEX IF NOT EXISTS idx_vault_audit_logs_user_id ON vault_audit_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_vault_audit_logs_timestamp ON vault_audit_logs(timestamp);
	`

	_, err := s.db.Exec(schema)
	return err
}

// CreateSecret creates a new secret
func (s *StorageBackend) CreateSecret(ctx context.Context, secret *Secret) error {
	if secret.Path == "" {
		return ErrInvalidSecretPath
	}

	query := `
		INSERT INTO vault_secrets (path, value, description, created_by, version, expires_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(
		ctx, query,
		secret.Path, secret.Value, secret.Description,
		secret.CreatedBy, secret.Version, secret.ExpiresAt, secret.Metadata,
	).Scan(&secret.ID, &secret.CreatedAt, &secret.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrSecretAlreadyExists
		}
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetSecret retrieves a secret by path
func (s *StorageBackend) GetSecret(ctx context.Context, path string) (*Secret, error) {
	if path == "" {
		return nil, ErrInvalidSecretPath
	}

	var secret Secret
	query := `
		SELECT id, path, value, description, created_by, created_at, updated_at,
		       version, expires_at, metadata
		FROM vault_secrets
		WHERE path = $1
	`

	err := s.db.GetContext(ctx, &secret, query, path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Check if secret has expired
	if secret.ExpiresAt != nil && secret.ExpiresAt.Before(time.Now()) {
		return nil, ErrSecretNotFound
	}

	return &secret, nil
}

// UpdateSecret updates an existing secret and creates a version history
func (s *StorageBackend) UpdateSecret(ctx context.Context, path string, newValue string, updatedBy uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current secret
	var currentSecret Secret
	err = tx.GetContext(ctx, &currentSecret, "SELECT * FROM vault_secrets WHERE path = $1 FOR UPDATE", path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSecretNotFound
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Save current version to history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO vault_secret_versions (secret_id, version, value, updated_by)
		VALUES ($1, $2, $3, $4)
	`, currentSecret.ID, currentSecret.Version, currentSecret.Value, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to save version history: %w", err)
	}

	// Update secret with new value and increment version
	_, err = tx.ExecContext(ctx, `
		UPDATE vault_secrets
		SET value = $1, version = version + 1, updated_at = NOW()
		WHERE path = $2
	`, newValue, path)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return tx.Commit()
}

// DeleteSecret deletes a secret
func (s *StorageBackend) DeleteSecret(ctx context.Context, path string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM vault_secrets WHERE path = $1", path)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrSecretNotFound
	}

	return nil
}

// ListSecrets lists all secrets (without values) matching a path prefix
func (s *StorageBackend) ListSecrets(ctx context.Context, pathPrefix string) ([]Secret, error) {
	var secrets []Secret
	query := `
		SELECT id, path, description, created_by, created_at, updated_at,
		       version, expires_at, metadata
		FROM vault_secrets
		WHERE path LIKE $1
		ORDER BY path
	`

	err := s.db.SelectContext(ctx, &secrets, query, pathPrefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}

// GetSecretVersion retrieves a specific version of a secret
func (s *StorageBackend) GetSecretVersion(ctx context.Context, path string, version int) (*SecretVersion, error) {
	var secretVersion SecretVersion
	query := `
		SELECT sv.id, sv.secret_id, sv.version, sv.value, sv.updated_by, sv.updated_at
		FROM vault_secret_versions sv
		JOIN vault_secrets s ON sv.secret_id = s.id
		WHERE s.path = $1 AND sv.version = $2
	`

	err := s.db.GetContext(ctx, &secretVersion, query, path, version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret version: %w", err)
	}

	return &secretVersion, nil
}

// ListSecretVersions lists all versions of a secret
func (s *StorageBackend) ListSecretVersions(ctx context.Context, path string) ([]SecretVersion, error) {
	var versions []SecretVersion
	query := `
		SELECT sv.id, sv.secret_id, sv.version, sv.updated_by, sv.updated_at
		FROM vault_secret_versions sv
		JOIN vault_secrets s ON sv.secret_id = s.id
		WHERE s.path = $1
		ORDER BY sv.version DESC
	`

	err := s.db.SelectContext(ctx, &versions, query, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secret versions: %w", err)
	}

	return versions, nil
}

// LogAccess logs an access event to the audit log
func (s *StorageBackend) LogAccess(ctx context.Context, log *AuditLog) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO vault_audit_logs (secret_id, user_id, action, ip_address, success)
		VALUES ($1, $2, $3, $4, $5)
	`, log.SecretID, log.UserID, log.Action, log.IPAddress, log.Success)

	if err != nil {
		return fmt.Errorf("failed to log access: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a secret
func (s *StorageBackend) GetAuditLogs(ctx context.Context, secretID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := `
		SELECT id, secret_id, user_id, action, ip_address, success, timestamp
		FROM vault_audit_logs
		WHERE secret_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	err := s.db.SelectContext(ctx, &logs, query, secretID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}

// CleanupExpiredSecrets removes expired secrets
func (s *StorageBackend) CleanupExpiredSecrets(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM vault_secrets
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired secrets: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// isUniqueViolation checks if error is a unique constraint violation
func isUniqueViolation(err error) bool {
	// PostgreSQL unique violation error code: 23505
	return err != nil && (
		err.Error() == "pq: duplicate key value violates unique constraint \"vault_secrets_path_key\"" ||
		err.Error() == "UNIQUE constraint failed: vault_secrets.path")
}
