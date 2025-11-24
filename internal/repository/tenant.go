package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

// TenantRepository provides access to tenant records.
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new TenantRepository.
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetByID fetches a tenant by ID.
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	const query = `
        SELECT id, name, slug, plan, status, parent_tenant_id, created_at, updated_at, deleted_at
        FROM tenants
        WHERE id = $1
    `

	tenant := &models.Tenant{}
	if err := r.db.GetContext(ctx, tenant, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}
