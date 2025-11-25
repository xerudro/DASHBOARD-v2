package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/xerudro/DASHBOARD-v2/internal/middleware"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
	"github.com/xerudro/DASHBOARD-v2/internal/services/sites"
)

// SiteHandler handles site-related HTTP endpoints
type SiteHandler struct {
	siteManager *sites.SiteManager
}

// NewSiteHandler creates a new site handler
func NewSiteHandler(siteManager *sites.SiteManager) *SiteHandler {
	return &SiteHandler{
		siteManager: siteManager,
	}
}

// CreateSite handles POST /api/v1/sites
func (h *SiteHandler) CreateSite(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	var req sites.CreateSiteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Set tenant ID from auth context
	req.TenantID = tenantID.String()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.siteManager.CreateSite(ctx, &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create site")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": response.Message,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": response.Message,
		"data":    response,
	})
}

// ListSites handles GET /api/v1/sites
func (h *SiteHandler) ListSites(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	// Parse pagination parameters
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sitesList, err := h.siteManager.ListSites(ctx, tenantID.String(), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list sites")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load sites",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sitesList,
		"total":   len(sitesList),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetSite handles GET /api/v1/sites/:id
func (h *SiteHandler) GetSite(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	siteID := c.Params("id")
	if siteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Site ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	site, err := h.siteManager.GetSiteStatus(ctx, siteID)
	if err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to get site")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Site not found",
		})
	}

	// Verify site belongs to tenant
	if site.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    site,
	})
}

// UpdateSite handles PUT /api/v1/sites/:id
func (h *SiteHandler) UpdateSite(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	siteID := c.Params("id")
	if siteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Site ID is required",
		})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify site exists and belongs to tenant
	site, err := h.siteManager.GetSiteStatus(ctx, siteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Site not found",
		})
	}

	if site.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// Update site
	if err := h.siteManager.UpdateSite(ctx, siteID, updates); err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to update site")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update site",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Site updated successfully",
	})
}

// DeleteSite handles DELETE /api/v1/sites/:id
func (h *SiteHandler) DeleteSite(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	siteID := c.Params("id")
	if siteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Site ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify site exists and belongs to tenant
	site, err := h.siteManager.GetSiteStatus(ctx, siteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Site not found",
		})
	}

	if site.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// Delete site
	if err := h.siteManager.DeleteSite(ctx, siteID); err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to delete site")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete site",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Site deleted successfully",
	})
}

// RedeploySite handles POST /api/v1/sites/:id/redeploy
func (h *SiteHandler) RedeploySite(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	siteID := c.Params("id")
	if siteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Site ID is required",
		})
	}

	var req struct {
		TemplateID  string            `json:"template_id"`
		Environment map[string]string `json:"environment"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Verify site exists and belongs to tenant
	site, err := h.siteManager.GetSiteStatus(ctx, siteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Site not found",
		})
	}

	if site.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// Redeploy site
	if err := h.siteManager.RedeploySite(ctx, siteID, req.TemplateID, req.Environment); err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to redeploy site")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to redeploy site",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Site redeployment started",
	})
}

// GetSiteMetrics handles GET /api/v1/sites/:id/metrics
func (h *SiteHandler) GetSiteMetrics(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	siteID := c.Params("id")
	if siteID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Site ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify site exists and belongs to tenant
	site, err := h.siteManager.GetSiteStatus(ctx, siteID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Site not found",
		})
	}

	if site.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// Get metrics
	metrics, err := h.siteManager.GetSiteMetrics(ctx, siteID)
	if err != nil {
		log.Error().Err(err).Str("site_id", siteID).Msg("Failed to get site metrics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load metrics",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    metrics,
	})
}

// ValidateDomain handles POST /api/v1/sites/validate-domain
func (h *SiteHandler) ValidateDomain(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.siteManager.ValidateDomain(ctx, req.Domain, tenantID.String())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"valid":   false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"valid":   true,
		"message": "Domain is available",
	})
}

// ListTemplates handles GET /api/v1/sites/templates
func (h *SiteHandler) ListTemplates(c *fiber.Ctx) error {
	templateMgr := sites.NewTemplateManager()

	// Optional filter by type
	siteType := c.Query("type")

	var templates []*sites.TemplateConfig
	if siteType != "" {
		templates = templateMgr.ListTemplatesByType(siteType)
	} else {
		templates = templateMgr.ListTemplates()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    templates,
		"total":   len(templates),
	})
}

// GetTemplate handles GET /api/v1/sites/templates/:id
func (h *SiteHandler) GetTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	if templateID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Template ID is required",
		})
	}

	templateMgr := sites.NewTemplateManager()
	template, err := templateMgr.GetTemplate(templateID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Template not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    template,
	})
}

// GetSiteTypes handles GET /api/v1/sites/types
func (h *SiteHandler) GetSiteTypes(c *fiber.Ctx) error {
	siteTypes := []fiber.Map{
		{"value": models.SiteTypeStatic, "label": "Static Site", "icon": "file-code"},
		{"value": models.SiteTypeWordPress, "label": "WordPress", "icon": "wordpress"},
		{"value": models.SiteTypePHP, "label": "PHP Application", "icon": "php"},
		{"value": models.SiteTypeLaravel, "label": "Laravel", "icon": "laravel"},
		{"value": models.SiteTypeNodeJS, "label": "Node.js", "icon": "node-js"},
		{"value": models.SiteTypePython, "label": "Python", "icon": "python"},
		{"value": "docker", "label": "Docker", "icon": "docker"},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    siteTypes,
	})
}
