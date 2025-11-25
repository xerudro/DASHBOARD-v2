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

// SiteRepositoryInterface defines the interface for site repository operations
type SiteRepositoryInterface interface {
	Create(ctx context.Context, site *models.Site) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Site, error)
	GetByDomain(ctx context.Context, domain string) (*models.Site, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Site, error)
	Update(ctx context.Context, site *models.Site) error
	Delete(ctx context.Context, siteID uuid.UUID) error
	GetByType(ctx context.Context, tenantID uuid.UUID, siteType string) ([]*models.Site, error)
}

// SiteRepository handles site database operations
type SiteRepository struct {
	db *sqlx.DB
}

// NewSiteRepository creates a new site repository
func NewSiteRepository(db *sqlx.DB) *SiteRepository {
	return &SiteRepository{db: db}
}

// Create creates a new site
func (r *SiteRepository) Create(ctx context.Context, site *models.Site) error {
	query := `
		INSERT INTO sites (id, tenant_id, server_id, name, domain, type, php_version,
		                  nodejs_version, webserver, root_path, status, git_repo, git_branch,
		                  ssl_enabled, ssl_auto_renew, config, created_at, updated_at,
		                  deployed_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`

	now := time.Now()
	site.ID = uuid.New()
	site.CreatedAt = now
	site.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		site.ID,
		site.TenantID,
		site.ServerID,
		site.Name,
		site.Domain,
		site.Type,
		site.PHPVersion,
		site.NodeJSVersion,
		site.Webserver,
		site.RootPath,
		site.Status,
		site.GitRepo,
		site.GitBranch,
		site.SSLEnabled,
		site.SSLAutoRenew,
		site.Config,
		site.CreatedAt,
		site.UpdatedAt,
		site.DeployedAt,
		site.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create site: %w", err)
	}

	return nil
}

// GetByID retrieves a site by ID
func (r *SiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Site, error) {
	query := `
		SELECT id, tenant_id, server_id, name, domain, type, php_version, nodejs_version,
		       webserver, root_path, status, git_repo, git_branch, ssl_enabled, ssl_auto_renew,
		       config, created_at, updated_at, deployed_at, deleted_at
		FROM sites
		WHERE id = $1 AND deleted_at IS NULL
	`

	site := &models.Site{}
	err := r.db.GetContext(ctx, site, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("site not found")
		}
		return nil, fmt.Errorf("failed to get site by ID: %w", err)
	}

	return site, nil
}

// GetByDomain retrieves a site by domain
func (r *SiteRepository) GetByDomain(ctx context.Context, domain string) (*models.Site, error) {
	query := `
		SELECT id, tenant_id, server_id, name, domain, type, php_version, nodejs_version,
		       webserver, root_path, status, git_repo, git_branch, ssl_enabled, ssl_auto_renew,
		       config, created_at, updated_at, deployed_at, deleted_at
		FROM sites
		WHERE domain = $1 AND deleted_at IS NULL
	`

	site := &models.Site{}
	err := r.db.GetContext(ctx, site, query, domain)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil instead of error for domain checks
		}
		return nil, fmt.Errorf("failed to get site by domain: %w", err)
	}

	return site, nil
}

// ListByTenant retrieves sites by tenant ID
func (r *SiteRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Site, error) {
	query := `
		SELECT id, tenant_id, server_id, name, domain, type, php_version, nodejs_version,
		       webserver, root_path, status, git_repo, git_branch, ssl_enabled, ssl_auto_renew,
		       config, created_at, updated_at, deployed_at, deleted_at
		FROM sites
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var sites []*models.Site
	err := r.db.SelectContext(ctx, &sites, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites by tenant: %w", err)
	}

	return sites, nil
}

// ListByServer retrieves sites by server ID
func (r *SiteRepository) ListByServer(ctx context.Context, serverID string) ([]*models.Site, error) {
	// Parse string server ID to UUID
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID format: %w", err)
	}

	query := `
		SELECT id, tenant_id, server_id, name, domain, type, php_version, nodejs_version,
		       webserver, root_path, status, git_repo, git_branch, ssl_enabled, ssl_auto_renew,
		       config, created_at, updated_at, deployed_at, deleted_at
		FROM sites
		WHERE server_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var sites []*models.Site
	err = r.db.SelectContext(ctx, &sites, query, serverUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites by server: %w", err)
	}

	return sites, nil
}

// Update updates a site
func (r *SiteRepository) Update(ctx context.Context, site *models.Site) error {
	query := `
		UPDATE sites
		SET tenant_id = $2, server_id = $3, name = $4, domain = $5, type = $6,
		    php_version = $7, nodejs_version = $8, webserver = $9, root_path = $10,
		    status = $11, git_repo = $12, git_branch = $13, ssl_enabled = $14,
		    ssl_auto_renew = $15, config = $16, updated_at = $17, deployed_at = $18,
		    deleted_at = $19
		WHERE id = $1
	`

	site.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		site.ID,
		site.TenantID,
		site.ServerID,
		site.Name,
		site.Domain,
		site.Type,
		site.PHPVersion,
		site.NodeJSVersion,
		site.Webserver,
		site.RootPath,
		site.Status,
		site.GitRepo,
		site.GitBranch,
		site.SSLEnabled,
		site.SSLAutoRenew,
		site.Config,
		site.UpdatedAt,
		site.DeployedAt,
		site.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update site: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("site not found")
	}

	return nil
}

// UpdateStatus updates site status
func (r *SiteRepository) UpdateStatus(ctx context.Context, siteID uuid.UUID, status string) error {
	query := `
		UPDATE sites
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, siteID, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update site status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("site not found")
	}

	return nil
}

// Delete soft deletes a site
func (r *SiteRepository) Delete(ctx context.Context, siteID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE sites
		SET status = $2, deleted_at = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, siteID, models.SiteStatusSuspended, now, now)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("site not found")
	}

	return nil
}

// CountByTenant counts sites in a tenant
func (r *SiteRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM sites WHERE tenant_id = $1 AND deleted_at IS NULL`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count sites by tenant: %w", err)
	}

	return count, nil
}

// CountByStatus counts sites by status
func (r *SiteRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID, status string) (int, error) {
	query := `SELECT COUNT(*) FROM sites WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count sites by status: %w", err)
	}

	return count, nil
}

// GetByType retrieves sites by type
func (r *SiteRepository) GetByType(ctx context.Context, tenantID uuid.UUID, siteType string) ([]*models.Site, error) {
	query := `
		SELECT id, tenant_id, server_id, name, domain, type, php_version, nodejs_version,
		       webserver, root_path, status, git_repo, git_branch, ssl_enabled, ssl_auto_renew,
		       config, created_at, updated_at, deployed_at, deleted_at
		FROM sites
		WHERE tenant_id = $1 AND type = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var sites []*models.Site
	err := r.db.SelectContext(ctx, &sites, query, tenantID, siteType)
	if err != nil {
		return nil, fmt.Errorf("failed to get sites by type: %w", err)
	}

	return sites, nil
}
