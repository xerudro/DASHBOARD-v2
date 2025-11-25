package sites

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/xerudro/DASHBOARD-v2/internal/cache"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
)

// DeploymentRequest represents a site deployment request
type DeploymentRequest struct {
	SiteID      string                 `json:"site_id"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config,omitempty"`
	EnvVars     map[string]string      `json:"env_vars,omitempty"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Overrides   map[string]interface{} `json:"overrides,omitempty"`
}

// CreateSiteRequest represents a request to create a new site
type CreateSiteRequest struct {
	TenantID    string            `json:"tenant_id"`
	ServerID    string            `json:"server_id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Domain      string            `json:"domain"`
	Path        string            `json:"path,omitempty"`
	Config      models.SiteConfig `json:"config,omitempty"`
	TemplateID  string            `json:"template_id,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// UpdateSiteRequest represents a request to update a site
type UpdateSiteRequest struct {
	Name   string            `json:"name,omitempty"`
	Domain string            `json:"domain,omitempty"`
	Path   string            `json:"path,omitempty"`
	Config models.SiteConfig `json:"config,omitempty"`
}

// SiteResponse represents a response from site operations
type SiteResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SiteListResponse represents a response from site listing operations
type SiteListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    []*models.Site `json:"data,omitempty"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

// CreateSiteResponse represents the response from creating a site
type CreateSiteResponse struct {
	Site       *models.Site `json:"site"`
	Deployed   bool         `json:"deployed"`
	Message    string       `json:"message"`
	DeployLogs []string     `json:"deploy_logs,omitempty"`
}

// SiteManager handles site lifecycle operations
type SiteManager struct {
	siteRepo    repository.SiteRepositoryInterface
	serverRepo  *repository.ServerRepository
	tenantRepo  *repository.TenantRepository
	deployer    *Deployer
	templateMgr *TemplateManager
	cache       *cache.RedisCache
}

// NewSiteManager creates a new site manager
func NewSiteManager(
	siteRepo repository.SiteRepositoryInterface,
	serverRepo *repository.ServerRepository,
	tenantRepo *repository.TenantRepository,
	deployer *Deployer,
	templateMgr *TemplateManager,
	cache *cache.RedisCache,
) *SiteManager {
	return &SiteManager{
		siteRepo:    siteRepo,
		serverRepo:  serverRepo,
		tenantRepo:  tenantRepo,
		deployer:    deployer,
		templateMgr: templateMgr,
		cache:       cache,
	}
}

// CreateSite creates a new site
func (sm *SiteManager) CreateSite(ctx context.Context, req *CreateSiteRequest) (*CreateSiteResponse, error) {
	response := &CreateSiteResponse{
		Deployed: false,
	}

	// Validate tenant
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		response.Message = "Invalid tenant ID format"
		return response, fmt.Errorf("invalid tenant ID: %w", err)
	}

	tenant, err := sm.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		response.Message = fmt.Sprintf("Invalid tenant: %v", err)
		return response, fmt.Errorf("invalid tenant: %w", err)
	}

	// Validate server
	serverID, err := uuid.Parse(req.ServerID)
	if err != nil {
		response.Message = "Invalid server ID format"
		return response, fmt.Errorf("invalid server ID: %w", err)
	}

	server, err := sm.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		response.Message = fmt.Sprintf("Invalid server: %v", err)
		return response, fmt.Errorf("invalid server: %w", err)
	}

	// Check if server belongs to tenant
	if server.TenantID != tenantID {
		response.Message = "Server does not belong to tenant"
		return response, fmt.Errorf("server does not belong to tenant")
	}

	// Check if domain is already in use
	existingSite, err := sm.siteRepo.GetByDomain(ctx, req.Domain)
	if err == nil && existingSite != nil {
		response.Message = fmt.Sprintf("Domain %s is already in use", req.Domain)
		return response, fmt.Errorf("domain already in use: %s", req.Domain)
	}

	// Prepare deployment configuration (this validates the template)
	// Convert SiteConfig to map[string]interface{}
	configMap := map[string]interface{}{
		"cache_enabled":       req.Config.CacheEnabled,
		"waf_enabled":         req.Config.WAFEnabled,
		"compression_enabled": req.Config.CompressionEnabled,
		"basic_auth_enabled":  req.Config.BasicAuthEnabled,
		"env_vars":            req.Config.EnvVars,
		"custom_headers":      req.Config.CustomHeaders,
		"redirects":           req.Config.Redirects,
	}
	deployConfig, err := sm.templateMgr.PrepareDeploymentConfig(req.TemplateID, configMap)
	if err != nil {
		response.Message = fmt.Sprintf("Invalid configuration: %v", err)
		return response, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create site record
	site := &models.Site{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ServerID:     serverID,
		Name:         req.Name,
		Domain:       req.Domain,
		Type:         req.Type,
		Status:       models.SiteStatusPending,
		SSLEnabled:   false,
		SSLAutoRenew: true,
		Config:       req.Config,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Set site configuration from template
	if phpVersion, ok := deployConfig["php_version"].(string); ok {
		site.PHPVersion = &phpVersion
	}
	if nodeVersion, ok := deployConfig["nodejs_version"].(string); ok {
		site.NodeJSVersion = &nodeVersion
	}
	if webserver, ok := deployConfig["webserver"].(string); ok {
		site.Webserver = webserver
	}

	// Save site to database
	if err := sm.siteRepo.Create(ctx, site); err != nil {
		response.Message = fmt.Sprintf("Failed to create site record: %v", err)
		return response, fmt.Errorf("failed to create site: %w", err)
	}

	response.Site = site
	response.Message = "Site created successfully"

	// Attempt deployment
	deployReq := &DeploymentRequest{
		SiteID:      site.ID.String(),
		Type:        req.Type,
		Config:      deployConfig,
		TemplateID:  req.TemplateID,
		Environment: req.Environment,
		Overrides:   deployConfig,
	}

	deployResult, err := sm.deployer.Deploy(ctx, deployReq)
	if err != nil {
		log.Error().Err(err).Str("site_id", site.ID.String()).Msg("Site deployment failed")

		// Update site status to failed
		site.Status = models.SiteStatusFailed
		if updateErr := sm.siteRepo.Update(ctx, site); updateErr != nil {
			log.Error().Err(updateErr).Str("site_id", site.ID.String()).Msg("Failed to update site status")
		}

		response.Message = fmt.Sprintf("Site created but deployment failed: %v", err)
		response.DeployLogs = deployResult.Logs
		return response, nil // Don't return error, site was created
	}

	// Update site status to deployed
	site.Status = models.SiteStatusActive
	site.DeployedAt = &deployResult.DeployedAt
	if updateErr := sm.siteRepo.Update(ctx, site); updateErr != nil {
		log.Error().Err(updateErr).Str("site_id", site.ID.String()).Msg("Failed to update site status")
	}

	response.Deployed = true
	response.Message = "Site created and deployed successfully"
	response.DeployLogs = deployResult.Logs

	// Invalidate cache
	sm.invalidateSiteCache(ctx, site.ID.String(), tenant.ID.String())

	log.Info().
		Str("site_id", site.ID.String()).
		Str("domain", site.Domain).
		Str("tenant_id", tenant.ID.String()).
		Msg("Site created and deployed successfully")

	return response, nil
}

// UpdateSite updates an existing site
func (sm *SiteManager) UpdateSite(ctx context.Context, siteID string, updates map[string]interface{}) error {
	// Parse site ID
	siteUUID, err := uuid.Parse(siteID)
	if err != nil {
		return fmt.Errorf("invalid site ID: %w", err)
	}

	// Get existing site
	site, err := sm.siteRepo.GetByID(ctx, siteUUID)
	if err != nil {
		return fmt.Errorf("site not found: %w", err)
	}

	// Apply updates
	if domain, ok := updates["domain"].(string); ok {
		// Check if new domain is available
		existingSite, err := sm.siteRepo.GetByDomain(ctx, domain)
		if err == nil && existingSite != nil && existingSite.ID != siteUUID {
			return fmt.Errorf("domain already in use: %s", domain)
		}
		site.Domain = domain
	}

	if config, ok := updates["config"].(map[string]interface{}); ok {
		// Update site config - this would need to be properly implemented
		// For now, we'll skip this as SiteConfig structure is complex
		_ = config
	}

	site.UpdatedAt = time.Now()

	// Save updates
	if err := sm.siteRepo.Update(ctx, site); err != nil {
		return fmt.Errorf("failed to update site: %w", err)
	}

	// Invalidate cache
	sm.invalidateSiteCache(ctx, siteID, site.TenantID.String())

	log.Info().
		Str("site_id", siteID).
		Interface("updates", updates).
		Msg("Site updated successfully")

	return nil
}

// DeleteSite deletes a site
func (sm *SiteManager) DeleteSite(ctx context.Context, siteID string) error {
	// Parse site ID
	siteUUID, err := uuid.Parse(siteID)
	if err != nil {
		return fmt.Errorf("invalid site ID: %w", err)
	}

	// Get site
	site, err := sm.siteRepo.GetByID(ctx, siteUUID)
	if err != nil {
		return fmt.Errorf("site not found: %w", err)
	}

	// Update status to deleting
	site.Status = models.SiteStatusDeleting
	site.UpdatedAt = time.Now()
	if err := sm.siteRepo.Update(ctx, site); err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to update site status to deleting")
	}

	// TODO: Run Ansible playbook to remove site from server
	// For now, just mark as deleted

	// Soft delete the site
	if err := sm.siteRepo.Delete(ctx, siteUUID); err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	// Invalidate cache
	sm.invalidateSiteCache(ctx, siteID, site.TenantID.String())

	log.Info().
		Str("site_id", siteID).
		Str("domain", site.Domain).
		Msg("Site deleted successfully")

	return nil
}

// RedeploySite redeploys an existing site
func (sm *SiteManager) RedeploySite(ctx context.Context, siteID string, templateID string, environment map[string]string) error {
	// Parse site ID
	siteUUID, err := uuid.Parse(siteID)
	if err != nil {
		return fmt.Errorf("invalid site ID: %w", err)
	}

	// Get site
	site, err := sm.siteRepo.GetByID(ctx, siteUUID)
	if err != nil {
		return fmt.Errorf("site not found: %w", err)
	}

	// Update status to deploying
	site.Status = models.SiteStatusDeploying
	site.UpdatedAt = time.Now()
	if err := sm.siteRepo.Update(ctx, site); err != nil {
		return fmt.Errorf("failed to update site status: %w", err)
	}

	// Prepare deployment
	deployReq := &DeploymentRequest{
		SiteID:      siteID,
		TemplateID:  templateID,
		Environment: environment,
	}

	// Deploy
	result, err := sm.deployer.Deploy(ctx, deployReq)
	if err != nil {
		// Update status to failed
		site.Status = models.SiteStatusFailed
		if updateErr := sm.siteRepo.Update(ctx, site); updateErr != nil {
			log.Error().Err(updateErr).Str("site_id", siteID).Msg("Failed to update site status")
		}
		return fmt.Errorf("redeployment failed: %w", err)
	}

	// Update status to active
	site.Status = models.SiteStatusActive
	site.DeployedAt = &result.DeployedAt
	if err := sm.siteRepo.Update(ctx, site); err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to update site status")
	}

	// Invalidate cache
	sm.invalidateSiteCache(ctx, siteID, site.TenantID.String())

	log.Info().
		Str("site_id", siteID).
		Str("domain", site.Domain).
		Msg("Site redeployed successfully")

	return nil
}

// GetSiteStatus returns the current status of a site
func (sm *SiteManager) GetSiteStatus(ctx context.Context, siteID string) (*models.Site, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("site:status:%s", siteID)
	var cachedSite models.Site
	found, err := sm.cache.Get(ctx, cacheKey, &cachedSite)
	if err == nil && found {
		return &cachedSite, nil
	}

	// Parse site ID
	siteUUID, err := uuid.Parse(siteID)
	if err != nil {
		return nil, fmt.Errorf("invalid site ID: %w", err)
	}

	// Get from database
	site, err := sm.siteRepo.GetByID(ctx, siteUUID)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	// Cache for 5 minutes
	opts := cache.CacheOptions{TTL: 5 * time.Minute}
	sm.cache.Set(ctx, cacheKey, site, opts)

	return site, nil
}

// ListSites lists sites for a tenant
func (sm *SiteManager) ListSites(ctx context.Context, tenantID string, limit, offset int) ([]*models.Site, error) {
	// Parse tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Try cache first
	cacheKey := fmt.Sprintf("sites:list:%s:%d:%d", tenantID, limit, offset)
	var cached []*models.Site
	found, err := sm.cache.Get(ctx, cacheKey, &cached)
	if err == nil && found {
		return cached, nil
	}

	// Get from database
	sites, err := sm.siteRepo.ListByTenant(ctx, tenantUUID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites: %w", err)
	}

	// Cache for 2 minutes
	opts := cache.CacheOptions{TTL: 2 * time.Minute}
	sm.cache.Set(ctx, cacheKey, sites, opts)

	return sites, nil
}

// GetSiteMetrics returns metrics for a site
func (sm *SiteManager) GetSiteMetrics(ctx context.Context, siteID string) (map[string]interface{}, error) {
	// Parse site ID
	siteUUID, err := uuid.Parse(siteID)
	if err != nil {
		return nil, fmt.Errorf("invalid site ID: %w", err)
	}

	site, err := sm.siteRepo.GetByID(ctx, siteUUID)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	server, err := sm.serverRepo.GetByID(ctx, site.ServerID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// TODO: Get actual metrics from monitoring service
	// For now, return mock data
	metrics := map[string]interface{}{
		"site_id":         site.ID,
		"domain":          site.Domain,
		"server_ip":       server.IPAddress,
		"status":          string(site.Status),
		"uptime":          "99.9%",
		"response_time":   "150ms",
		"requests_today":  1250,
		"bandwidth_today": "2.5GB",
		"last_updated":    time.Now(),
	}

	return metrics, nil
}

// ValidateDomain checks if a domain is valid and available
func (sm *SiteManager) ValidateDomain(ctx context.Context, domain string, tenantID string) error {
	// Basic domain validation
	if len(domain) == 0 {
		return fmt.Errorf("domain cannot be empty")
	}

	if len(domain) > 253 {
		return fmt.Errorf("domain too long")
	}

	// Parse tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Check if domain is already in use by this tenant
	existingSite, err := sm.siteRepo.GetByDomain(ctx, domain)
	if err == nil && existingSite != nil {
		if existingSite.TenantID == tenantUUID {
			return fmt.Errorf("domain already in use by your account")
		} else {
			return fmt.Errorf("domain already in use by another account")
		}
	}

	return nil
}

// invalidateSiteCache invalidates site-related cache entries
func (sm *SiteManager) invalidateSiteCache(ctx context.Context, siteID, tenantID string) {
	keys := []string{
		fmt.Sprintf("site:status:%s", siteID),
		fmt.Sprintf("sites:list:%s:*", tenantID), // Invalidate all list caches for tenant
	}

	for _, key := range keys {
		if err := sm.cache.Delete(ctx, key); err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Failed to invalidate cache")
		}
	}
}
