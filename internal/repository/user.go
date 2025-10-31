package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, two_factor_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
		user.TwoFactorEnabled,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, role, status, 
		       two_factor_enabled, two_factor_secret, last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, role, status, 
		       two_factor_enabled, two_factor_secret, last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByTenant retrieves users by tenant ID
func (r *UserRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, role, status, 
		       two_factor_enabled, two_factor_secret, last_login_at, created_at, updated_at
		FROM users 
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var users []*models.User
	err := r.db.SelectContext(ctx, &users, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by tenant: %w", err)
	}

	return users, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, first_name = $3, last_name = $4, role = $5, status = $6, 
		    two_factor_enabled = $7, updated_at = $8
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
		user.TwoFactorEnabled,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdatePassword updates a user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users 
		SET password_hash = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET last_login_at = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, userID, now, now)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateTwoFactorSecret updates the user's 2FA secret
func (r *UserRepository) UpdateTwoFactorSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	query := `
		UPDATE users 
		SET two_factor_secret = $2, two_factor_enabled = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, secret, true, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update 2FA secret: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// DisableTwoFactor disables 2FA for a user
func (r *UserRepository) DisableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET two_factor_enabled = $2, two_factor_secret = NULL, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, false, time.Now())
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, models.UserStatusInactive, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// CountByTenant counts users in a tenant
func (r *UserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND status != $2`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID, models.UserStatusInactive)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by tenant: %w", err)
	}

	return count, nil
}

// Authenticate validates user credentials and returns user if valid
func (r *UserRepository) Authenticate(ctx context.Context, email, passwordHash string) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, role, status, 
		       two_factor_enabled, two_factor_secret, last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1 AND password_hash = $2 AND status = $3
	`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, email, passwordHash, models.UserStatusActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	return user, nil
}

// GetByTenantAndRole retrieves users by tenant and role
func (r *UserRepository) GetByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role string) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, role, status, 
		       two_factor_enabled, two_factor_secret, last_login_at, created_at, updated_at
		FROM users 
		WHERE tenant_id = $1 AND role = $2 AND status = $3
		ORDER BY created_at DESC
	`

	var users []*models.User
	err := r.db.SelectContext(ctx, &users, query, tenantID, role, models.UserStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by tenant and role: %w", err)
	}

	return users, nil
}