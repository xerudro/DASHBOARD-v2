package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xerudro/DASHBOARD-v2/internal/cache"
	"github.com/xerudro/DASHBOARD-v2/internal/middleware"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
)

// CacheService interface for cache operations needed by dashboard
type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, opts cache.CacheOptions) error
	InvalidateDashboardStats(ctx context.Context, tenantID uuid.UUID) error
}

// DashboardHandler handles dashboard endpoints
type DashboardHandler struct {
	userRepo     *repository.UserRepository
	serverRepo   *repository.ServerRepository
	cacheService CacheService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(userRepo *repository.UserRepository, serverRepo *repository.ServerRepository, cacheService CacheService) *DashboardHandler {
	return &DashboardHandler{
		userRepo:     userRepo,
		serverRepo:   serverRepo,
		cacheService: cacheService,
	}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	Servers struct {
		Total        int            `json:"total"`
		Ready        int            `json:"ready"`
		Provisioning int            `json:"provisioning"`
		Failed       int            `json:"failed"`
		ByProvider   map[string]int `json:"by_provider"`
	} `json:"servers"`
	Users struct {
		Total  int `json:"total"`
		Active int `json:"active"`
	} `json:"users"`
	RecentActivity []ActivityItem `json:"recent_activity"`
}

// ActivityItem represents a recent activity
type ActivityItem struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"` // server_created, user_login, etc.
	Description string    `json:"description"`
	UserEmail   string    `json:"user_email,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ServerSummary represents server summary for dashboard
type ServerSummary struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Provider string    `json:"provider"`
	Region   string    `json:"region"`
	Status   string    `json:"status"`
	CPU      string    `json:"cpu"`
	RAM      string    `json:"ram"`
	Disk     string    `json:"disk"`
	Uptime   string    `json:"uptime"`
}

// GetDashboard returns dashboard data (API)
func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	userID, tenantID, _, role := middleware.GetUserFromContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get dashboard stats
	stats, err := h.getDashboardStats(ctx, tenantID, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard stats")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load dashboard data",
		})
	}

	// Get recent servers with metrics
	servers, err := h.getRecentServers(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent servers")
		// Don't fail the request, just log and continue with empty servers
		servers = []*ServerSummary{}
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("tenant_id", tenantID.String()).
		Int("server_count", len(servers)).
		Msg("Dashboard data requested")

	return c.JSON(fiber.Map{
		"stats":   stats,
		"servers": servers,
	})
}

// GetStats returns dashboard statistics (API)
func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	_, tenantID, _, role := middleware.GetUserFromContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := h.getDashboardStats(ctx, tenantID, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard stats")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load statistics",
		})
	}

	return c.JSON(stats)
}

// DashboardPage renders dashboard page (HTML)
func (h *DashboardHandler) DashboardPage(c *fiber.Ctx) error {
	userID, tenantID, email, role := middleware.GetUserFromContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get dashboard stats with N/A fallbacks
	stats, err := h.getDashboardStats(ctx, tenantID, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard stats")
		// Use empty stats with N/A fallbacks
		stats = &DashboardStats{}
	}

	// Get recent servers
	servers, err := h.getRecentServers(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent servers")
		servers = []*ServerSummary{}
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("email", email).
		Str("role", role).
		Msg("Dashboard page requested")

	// For now, return a simple HTML dashboard
	// In production, this would use Templ templates
	return c.Type("html").SendString(h.renderDashboardHTML(email, role, stats, servers))
}

// getDashboardStats fetches dashboard statistics with Redis caching and N/A fallbacks
func (h *DashboardHandler) getDashboardStats(ctx context.Context, tenantID uuid.UUID, role string) (*DashboardStats, error) {
	// Create cache key based on tenant and role (admin sees user counts)
	cacheKey := fmt.Sprintf("dashboard:stats:%s:%s", tenantID.String(), role)

	// Try to get from cache first
	if h.cacheService != nil {
		var cachedStats DashboardStats
		found, err := h.cacheService.Get(ctx, cacheKey, &cachedStats)
		if err != nil {
			log.Warn().Err(err).Str("cache_key", cacheKey).Msg("Cache get error")
		} else if found {
			log.Debug().
				Str("tenant_id", tenantID.String()).
				Str("role", role).
				Msg("Dashboard stats served from cache")
			return &cachedStats, nil
		}
	}

	// Cache miss - fetch from database
	log.Debug().
		Str("tenant_id", tenantID.String()).
		Str("role", role).
		Msg("Dashboard stats cache miss - fetching from database")

	stats := &DashboardStats{}

	// Get server counts with fallbacks
	totalServers, err := h.serverRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count servers")
		totalServers = 0
	}
	stats.Servers.Total = totalServers

	// Get servers by status
	readyCount, err := h.serverRepo.CountByStatus(ctx, tenantID, models.ServerStatusReady)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count ready servers")
		readyCount = 0
	}
	stats.Servers.Ready = readyCount

	provisioningCount, err := h.serverRepo.CountByStatus(ctx, tenantID, models.ServerStatusProvisioning)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count provisioning servers")
		provisioningCount = 0
	}
	stats.Servers.Provisioning = provisioningCount

	failedCount, err := h.serverRepo.CountByStatus(ctx, tenantID, models.ServerStatusFailed)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count failed servers")
		failedCount = 0
	}
	stats.Servers.Failed = failedCount

	// Initialize provider counts (would be implemented with proper queries)
	stats.Servers.ByProvider = make(map[string]int)

	// Get user counts (only for admins)
	if middleware.IsAdmin(role) {
		totalUsers, err := h.userRepo.CountByTenant(ctx, tenantID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to count users")
			totalUsers = 0
		}
		stats.Users.Total = totalUsers
		stats.Users.Active = totalUsers // Simplified for now
	}

	// Recent activity (placeholder - would come from audit logs)
	stats.RecentActivity = []ActivityItem{
		{
			ID:          uuid.New(),
			Type:        "server_created",
			Description: "Server web-01 created",
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Type:        "user_login",
			Description: "User logged in",
			Timestamp:   time.Now().Add(-4 * time.Hour),
		},
	}

	// Cache the result for 30 seconds with proper tags for invalidation
	if h.cacheService != nil {
		cacheOpts := cache.CacheOptions{
			TTL:  30 * time.Second, // 30-second cache as per performance requirements
			Tags: []string{"dashboard", "servers", tenantID.String()},
		}

		if err := h.cacheService.Set(ctx, cacheKey, stats, cacheOpts); err != nil {
			log.Error().
				Err(err).
				Str("cache_key", cacheKey).
				Msg("Failed to cache dashboard stats")
			// Don't return error - we have the data
		} else {
			log.Debug().
				Str("tenant_id", tenantID.String()).
				Str("role", role).
				Dur("ttl", cacheOpts.TTL).
				Msg("Dashboard stats cached successfully")
		}
	}

	return stats, nil
}

// InvalidateDashboardCache invalidates dashboard cache for a tenant
func (h *DashboardHandler) InvalidateDashboardCache(ctx context.Context, tenantID uuid.UUID) error {
	if h.cacheService == nil {
		return nil
	}

	return h.cacheService.InvalidateDashboardStats(ctx, tenantID)
}

// getRecentServers fetches recent servers with metrics
func (h *DashboardHandler) getRecentServers(ctx context.Context, tenantID uuid.UUID) ([]*ServerSummary, error) {
	// Get servers with metrics
	serversWithMetrics, err := h.serverRepo.GetWithMetrics(ctx, tenantID, 10, 0)
	if err != nil {
		return nil, err
	}

	summaries := make([]*ServerSummary, len(serversWithMetrics))
	for i, swm := range serversWithMetrics {
		// Helper function to safely get string from pointer
		region := "N/A"
		if swm.Server.Region != nil {
			region = *swm.Server.Region
		}

		summary := &ServerSummary{
			ID:       swm.Server.ID,
			Name:     swm.Server.Name,
			Provider: swm.GetProviderDisplay(),
			Region:   region,
			Status:   swm.GetStatusDisplay(),
			CPU:      swm.GetCPUDisplay(),
			RAM:      swm.GetMemoryDisplay(),
			Disk:     swm.GetDiskDisplay(),
			Uptime:   "N/A", // Uptime field not available in ServerMetrics
		}

		summaries[i] = summary
	}

	return summaries, nil
}

// renderDashboardHTML renders a simple dashboard HTML page
func (h *DashboardHandler) renderDashboardHTML(email, role string, stats *DashboardStats, servers []*ServerSummary) string {
	serversSection := ""
	for _, server := range servers {
		serversSection += `
			<div class="bg-white rounded-lg shadow p-4">
				<h3 class="font-semibold">` + server.Name + `</h3>
				<p class="text-sm text-gray-600">` + server.Provider + ` - ` + server.Region + `</p>
				<div class="mt-2 flex justify-between text-sm">
					<span class="` + getStatusColor(server.Status) + `">` + server.Status + `</span>
					<span>CPU: ` + server.CPU + `</span>
				</div>
				<div class="mt-1 flex justify-between text-sm text-gray-600">
					<span>RAM: ` + server.RAM + `</span>
					<span>Disk: ` + server.Disk + `</span>
				</div>
			</div>
		`
	}

	return `
<!DOCTYPE html>
<html>
<head>
    <title>VIP Hosting Panel - Dashboard</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.13.3/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100">
    <!-- Navigation -->
    <nav class="bg-white shadow-sm border-b">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between h-16">
                <div class="flex items-center">
                    <h1 class="text-xl font-semibold text-gray-900">VIP Hosting Panel</h1>
                </div>
                <div class="flex items-center space-x-4">
                    <span class="text-sm text-gray-700">` + email + ` (` + role + `)</span>
                    <a href="/servers" class="text-blue-600 hover:text-blue-800">Servers</a>
                    <a href="/logout" class="text-red-600 hover:text-red-800">Logout</a>
                </div>
            </div>
        </div>
    </nav>

    <!-- Main Content -->
    <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div class="px-4 py-6 sm:px-0">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Dashboard</h2>
            
            <!-- Stats Cards -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
                                    <span class="text-white font-semibold">S</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Total Servers</dt>
                                    <dd class="text-lg font-medium text-gray-900">` + intToString(stats.Servers.Total) + `</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                                    <span class="text-white font-semibold">R</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Ready</dt>
                                    <dd class="text-lg font-medium text-gray-900">` + intToString(stats.Servers.Ready) + `</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-yellow-500 rounded-full flex items-center justify-center">
                                    <span class="text-white font-semibold">P</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Provisioning</dt>
                                    <dd class="text-lg font-medium text-gray-900">` + intToString(stats.Servers.Provisioning) + `</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="bg-white overflow-hidden shadow rounded-lg">
                    <div class="p-5">
                        <div class="flex items-center">
                            <div class="flex-shrink-0">
                                <div class="w-8 h-8 bg-red-500 rounded-full flex items-center justify-center">
                                    <span class="text-white font-semibold">F</span>
                                </div>
                            </div>
                            <div class="ml-5 w-0 flex-1">
                                <dl>
                                    <dt class="text-sm font-medium text-gray-500 truncate">Failed</dt>
                                    <dd class="text-lg font-medium text-gray-900">` + intToString(stats.Servers.Failed) + `</dd>
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Recent Servers -->
            <div class="bg-white shadow rounded-lg">
                <div class="px-4 py-5 sm:p-6">
                    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Recent Servers</h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" 
                         hx-get="/api/v1/dashboard/stats" 
                         hx-trigger="every 30s"
                         hx-swap="innerHTML">
                        ` + serversSection + `
                    </div>
                    <div class="mt-4">
                        <a href="/servers" class="text-blue-600 hover:text-blue-800 font-medium">
                            View all servers →
                        </a>
                    </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
	`
}

// Helper functions
func getStatusColor(status string) string {
	switch status {
	case models.ServerStatusReady:
		return "text-green-600"
	case models.ServerStatusProvisioning:
		return "text-yellow-600"
	case models.ServerStatusFailed:
		return "text-red-600"
	default:
		return "text-gray-600"
	}
}

func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}
