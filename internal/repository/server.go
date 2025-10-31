package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/vip-hosting-panel/internal/models"
)

// ServerRepository handles server database operations
type ServerRepository struct {
	db *sqlx.DB
}

// NewServerRepository creates a new server repository
func NewServerRepository(db *sqlx.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// Create creates a new server
func (r *ServerRepository) Create(ctx context.Context, server *models.Server) error {
	query := `
		INSERT INTO servers (id, tenant_id, name, provider, region, plan, external_id, ip_address, 
		                    status, specs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	server.ID = uuid.New()
	server.CreatedAt = now
	server.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		server.ID,
		server.TenantID,
		server.Name,
		server.Provider,
		server.Region,
		server.Plan,
		server.ExternalID,
		server.IPAddress,
		server.Status,
		server.Specs,
		server.CreatedAt,
		server.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return nil
}

// GetByID retrieves a server by ID
func (r *ServerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Server, error) {
	query := `
		SELECT id, tenant_id, name, provider, region, plan, external_id, ip_address, 
		       status, specs, created_at, updated_at
		FROM servers 
		WHERE id = $1
	`

	server := &models.Server{}
	err := r.db.GetContext(ctx, server, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("server not found")
		}
		return nil, fmt.Errorf("failed to get server by ID: %w", err)
	}

	return server, nil
}

// GetByTenant retrieves servers by tenant ID with N/A fallback
func (r *ServerRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Server, error) {
	query := `
		SELECT s.id, s.tenant_id, s.name, s.provider, s.region, s.plan, s.external_id, 
		       s.ip_address, s.status, s.specs, s.created_at, s.updated_at
		FROM servers s
		WHERE s.tenant_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var servers []*models.Server
	err := r.db.SelectContext(ctx, &servers, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers by tenant: %w", err)
	}

	// Apply N/A fallback pattern for each server
	for _, server := range servers {
		if server.IPAddress == "" {
			server.IPAddress = "N/A"
		}
		if server.Status == "" {
			server.Status = models.ServerStatusUnknown
		}
	}

	return servers, nil
}

// GetByProvider retrieves servers by provider
func (r *ServerRepository) GetByProvider(ctx context.Context, tenantID uuid.UUID, provider string) ([]*models.Server, error) {
	query := `
		SELECT id, tenant_id, name, provider, region, plan, external_id, ip_address, 
		       status, specs, created_at, updated_at
		FROM servers 
		WHERE tenant_id = $1 AND provider = $2
		ORDER BY created_at DESC
	`

	var servers []*models.Server
	err := r.db.SelectContext(ctx, &servers, query, tenantID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers by provider: %w", err)
	}

	return servers, nil
}

// GetByStatus retrieves servers by status
func (r *ServerRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*models.Server, error) {
	query := `
		SELECT id, tenant_id, name, provider, region, plan, external_id, ip_address, 
		       status, specs, created_at, updated_at
		FROM servers 
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	var servers []*models.Server
	err := r.db.SelectContext(ctx, &servers, query, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers by status: %w", err)
	}

	return servers, nil
}

// Update updates a server
func (r *ServerRepository) Update(ctx context.Context, server *models.Server) error {
	query := `
		UPDATE servers 
		SET name = $2, provider = $3, region = $4, plan = $5, external_id = $6, 
		    ip_address = $7, status = $8, specs = $9, updated_at = $10
		WHERE id = $1
	`

	server.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		server.ID,
		server.Name,
		server.Provider,
		server.Region,
		server.Plan,
		server.ExternalID,
		server.IPAddress,
		server.Status,
		server.Specs,
		server.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("server not found")
	}

	return nil
}

// UpdateStatus updates server status
func (r *ServerRepository) UpdateStatus(ctx context.Context, serverID uuid.UUID, status string) error {
	query := `
		UPDATE servers 
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, serverID, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("server not found")
	}

	return nil
}

// UpdateIPAddress updates server IP address
func (r *ServerRepository) UpdateIPAddress(ctx context.Context, serverID uuid.UUID, ipAddress string) error {
	query := `
		UPDATE servers 
		SET ip_address = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, serverID, ipAddress, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update server IP address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("server not found")
	}

	return nil
}

// Delete soft deletes a server
func (r *ServerRepository) Delete(ctx context.Context, serverID uuid.UUID) error {
	query := `
		UPDATE servers 
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, serverID, models.ServerStatusDeleted, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("server not found")
	}

	return nil
}

// CountByTenant counts servers in a tenant
func (r *ServerRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM servers WHERE tenant_id = $1 AND status != $2`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID, models.ServerStatusDeleted)
	if err != nil {
		return 0, fmt.Errorf("failed to count servers by tenant: %w", err)
	}

	return count, nil
}

// CountByStatus counts servers by status
func (r *ServerRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID, status string) (int, error) {
	query := `SELECT COUNT(*) FROM servers WHERE tenant_id = $1 AND status = $2`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count servers by status: %w", err)
	}

	return count, nil
}

// GetWithMetrics retrieves servers with their latest metrics
func (r *ServerRepository) GetWithMetrics(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.ServerWithMetrics, error) {
	query := `
		SELECT s.id, s.tenant_id, s.name, s.provider, s.region, s.plan, s.external_id, 
		       s.ip_address, s.status, s.specs, s.created_at, s.updated_at,
		       m.cpu_percent, m.memory_percent, m.disk_percent, m.load_average, 
		       m.uptime, m.status as metrics_status, m.collected_at
		FROM servers s
		LEFT JOIN (
			SELECT DISTINCT ON (server_id) server_id, cpu_percent, memory_percent, 
			       disk_percent, load_average, uptime, status, collected_at
			FROM server_metrics 
			ORDER BY server_id, collected_at DESC
		) m ON s.id = m.server_id
		WHERE s.tenant_id = $1 AND s.status != $2
		ORDER BY s.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, models.ServerStatusDeleted, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers with metrics: %w", err)
	}
	defer rows.Close()

	var servers []*models.ServerWithMetrics
	for rows.Next() {
		server := &models.Server{}
		metrics := &models.ServerMetrics{}
		
		var cpuPercent, memoryPercent, diskPercent, loadAverage, uptime sql.NullFloat64
		var metricsStatus sql.NullString
		var collectedAt sql.NullTime

		err := rows.Scan(
			&server.ID, &server.TenantID, &server.Name, &server.Provider, &server.Region,
			&server.Plan, &server.ExternalID, &server.IPAddress, &server.Status,
			&server.Specs, &server.CreatedAt, &server.UpdatedAt,
			&cpuPercent, &memoryPercent, &diskPercent, &loadAverage, &uptime,
			&metricsStatus, &collectedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server with metrics: %w", err)
		}

		// Apply N/A fallback pattern
		if server.IPAddress == "" {
			server.IPAddress = "N/A"
		}

		// Handle metrics with N/A fallback
		if cpuPercent.Valid {
			metrics.CPUPercent = cpuPercent.Float64
			metrics.MemoryPercent = memoryPercent.Float64
			metrics.DiskPercent = diskPercent.Float64
			metrics.LoadAverage = loadAverage.Float64
			metrics.Uptime = uptime.Float64
			metrics.Status = metricsStatus.String
			metrics.CollectedAt = collectedAt.Time
		} else {
			// No metrics available - use N/A pattern
			metrics = nil
		}

		serverWithMetrics := &models.ServerWithMetrics{
			Server:  server,
			Metrics: metrics,
		}

		servers = append(servers, serverWithMetrics)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate server rows: %w", err)
	}

	return servers, nil
}

// GetByExternalID retrieves a server by external provider ID
func (r *ServerRepository) GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*models.Server, error) {
	query := `
		SELECT id, tenant_id, name, provider, region, plan, external_id, ip_address, 
		       status, specs, created_at, updated_at
		FROM servers 
		WHERE tenant_id = $1 AND external_id = $2
	`

	server := &models.Server{}
	err := r.db.GetContext(ctx, server, query, tenantID, externalID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("server not found")
		}
		return nil, fmt.Errorf("failed to get server by external ID: %w", err)
	}

	return server, nil
}

// models/server_with_metrics.go - Add this struct to models
type ServerWithMetrics struct {
	Server  *models.Server        `json:"server"`
	Metrics *models.ServerMetrics `json:"metrics,omitempty"`
}