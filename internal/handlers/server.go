package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xerudro/DASHBOARD-v2/internal/middleware"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
)

// ServerHandler handles server endpoints
type ServerHandler struct {
	serverRepo *repository.ServerRepository
}

// NewServerHandler creates a new server handler
func NewServerHandler(serverRepo *repository.ServerRepository) *ServerHandler {
	return &ServerHandler{
		serverRepo: serverRepo,
	}
}

// CreateServerRequest represents server creation request
type CreateServerRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Provider string `json:"provider" validate:"required,oneof=hetzner digitalocean vultr aws"`
	Region   string `json:"region" validate:"required,min=2,max=20"`
	Plan     string `json:"plan" validate:"required,min=2,max=50"`
}

// ServerResponse represents server in API responses
type ServerResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Provider   string    `json:"provider"`
	Region     string    `json:"region"`
	Plan       string    `json:"plan"`
	ExternalID string    `json:"external_id,omitempty"`
	IPAddress  string    `json:"ip_address"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Metrics    *models.ServerMetrics `json:"metrics,omitempty"`
}

// List returns servers for the current tenant (API)
func (h *ServerHandler) List(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	// Parse pagination parameters
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get servers with metrics
	serversWithMetrics, err := h.serverRepo.GetWithMetrics(ctx, tenantID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get servers")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load servers",
		})
	}

	// Convert to response format
	responses := make([]*ServerResponse, len(serversWithMetrics))
	for i, swm := range serversWithMetrics {
		responses[i] = &ServerResponse{
			ID:         swm.Server.ID,
			Name:       swm.Server.Name,
			Provider:   swm.Server.Provider,
			Region:     swm.Server.Region,
			Plan:       swm.Server.Plan,
			ExternalID: swm.Server.ExternalID,
			IPAddress:  swm.Server.IPAddress,
			Status:     swm.Server.Status,
			CreatedAt:  swm.Server.CreatedAt,
			UpdatedAt:  swm.Server.UpdatedAt,
			Metrics:    swm.Metrics,
		}
	}

	// Get total count for pagination
	totalCount, err := h.serverRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count servers")
		totalCount = len(responses) // Fallback
	}

	return c.JSON(fiber.Map{
		"servers": responses,
		"pagination": fiber.Map{
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// Get returns a single server (API)
func (h *ServerHandler) Get(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid server ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server, err := h.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Server not found",
		})
	}

	// Verify tenant ownership
	if server.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	response := &ServerResponse{
		ID:         server.ID,
		Name:       server.Name,
		Provider:   server.Provider,
		Region:     server.Region,
		Plan:       server.Plan,
		ExternalID: server.ExternalID,
		IPAddress:  server.IPAddress,
		Status:     server.Status,
		CreatedAt:  server.CreatedAt,
		UpdatedAt:  server.UpdatedAt,
	}

	return c.JSON(response)
}

// Create creates a new server (API)
func (h *ServerHandler) Create(c *fiber.Ctx) error {
	userID, tenantID, _, _ := middleware.GetUserFromContext(c)

	var req CreateServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request format",
		})
	}

	// Basic validation
	if req.Name == "" || req.Provider == "" || req.Region == "" || req.Plan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Name, provider, region, and plan are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create server record
	server := &models.Server{
		TenantID: tenantID,
		Name:     req.Name,
		Provider: req.Provider,
		Region:   req.Region,
		Plan:     req.Plan,
		Status:   models.ServerStatusQueued,
		Specs:    fmt.Sprintf(`{"plan": "%s", "region": "%s"}`, req.Plan, req.Region),
	}

	if err := h.serverRepo.Create(ctx, server); err != nil {
		log.Error().Err(err).Msg("Failed to create server")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create server",
		})
	}

	// TODO: Enqueue provisioning job
	// jobQueue.Enqueue(&ProvisionServerJob{ServerID: server.ID})

	log.Info().
		Str("server_id", server.ID.String()).
		Str("user_id", userID.String()).
		Str("name", server.Name).
		Str("provider", server.Provider).
		Msg("Server creation requested")

	response := &ServerResponse{
		ID:        server.ID,
		Name:      server.Name,
		Provider:  server.Provider,
		Region:    server.Region,
		Plan:      server.Plan,
		IPAddress: server.IPAddress,
		Status:    server.Status,
		CreatedAt: server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// Update updates a server (API)
func (h *ServerHandler) Update(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid server ID",
		})
	}

	var req struct {
		Name string `json:"name" validate:"min=2,max=50"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request format",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get existing server
	server, err := h.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Server not found",
		})
	}

	// Verify tenant ownership
	if server.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// Update fields
	if req.Name != "" {
		server.Name = req.Name
	}

	if err := h.serverRepo.Update(ctx, server); err != nil {
		log.Error().Err(err).Msg("Failed to update server")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update server",
		})
	}

	response := &ServerResponse{
		ID:        server.ID,
		Name:      server.Name,
		Provider:  server.Provider,
		Region:    server.Region,
		Plan:      server.Plan,
		IPAddress: server.IPAddress,
		Status:    server.Status,
		CreatedAt: server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
	}

	return c.JSON(response)
}

// Delete deletes a server (API)
func (h *ServerHandler) Delete(c *fiber.Ctx) error {
	userID, tenantID, _, _ := middleware.GetUserFromContext(c)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid server ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get server to verify ownership
	server, err := h.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Server not found",
		})
	}

	// Verify tenant ownership
	if server.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// Soft delete
	if err := h.serverRepo.Delete(ctx, serverID); err != nil {
		log.Error().Err(err).Msg("Failed to delete server")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to delete server",
		})
	}

	// TODO: Enqueue server destruction job
	// jobQueue.Enqueue(&DestroyServerJob{ServerID: serverID})

	log.Info().
		Str("server_id", serverID.String()).
		Str("user_id", userID.String()).
		Str("name", server.Name).
		Msg("Server deletion requested")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Server deleted successfully",
	})
}

// GetMetrics returns server metrics (API)
func (h *ServerHandler) GetMetrics(c *fiber.Ctx) error {
	_, tenantID, _, _ := middleware.GetUserFromContext(c)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid server ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify server exists and tenant ownership
	server, err := h.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Server not found",
		})
	}

	if server.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	// TODO: Get actual metrics from TimescaleDB
	// For now, return mock data with N/A fallback
	return c.JSON(fiber.Map{
		"cpu":    "N/A",
		"memory": "N/A",
		"disk":   "N/A",
		"uptime": "N/A",
		"status": "Unknown",
		"message": "Metrics collection not yet implemented",
	})
}

// ServersPage renders servers page (HTML)
func (h *ServerHandler) ServersPage(c *fiber.Ctx) error {
	_, tenantID, email, role := middleware.GetUserFromContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get servers with metrics
	serversWithMetrics, err := h.serverRepo.GetWithMetrics(ctx, tenantID, 50, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get servers for page")
		serversWithMetrics = []*models.ServerWithMetrics{} // Empty fallback
	}

	return c.Type("html").SendString(h.renderServersHTML(email, role, serversWithMetrics))
}

// CreateServerPage renders server creation page (HTML)
func (h *ServerHandler) CreateServerPage(c *fiber.Ctx) error {
	_, _, email, role := middleware.GetUserFromContext(c)

	return c.Type("html").SendString(h.renderCreateServerHTML(email, role))
}

// CreateServerForm handles server creation form (HTML)
func (h *ServerHandler) CreateServerForm(c *fiber.Ctx) error {
	userID, tenantID, _, _ := middleware.GetUserFromContext(c)

	name := c.FormValue("name")
	provider := c.FormValue("provider")
	region := c.FormValue("region")
	plan := c.FormValue("plan")

	if name == "" || provider == "" || region == "" || plan == "" {
		return c.Redirect("/servers/create?error=missing_fields")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create server
	server := &models.Server{
		TenantID: tenantID,
		Name:     name,
		Provider: provider,
		Region:   region,
		Plan:     plan,
		Status:   models.ServerStatusQueued,
		Specs:    fmt.Sprintf(`{"plan": "%s", "region": "%s"}`, plan, region),
	}

	if err := h.serverRepo.Create(ctx, server); err != nil {
		log.Error().Err(err).Msg("Failed to create server")
		return c.Redirect("/servers/create?error=creation_failed")
	}

	log.Info().
		Str("server_id", server.ID.String()).
		Str("user_id", userID.String()).
		Str("name", server.Name).
		Msg("Server created via form")

	return c.Redirect("/servers?success=server_created")
}

// renderServersHTML renders the servers list page
func (h *ServerHandler) renderServersHTML(email, role string, servers []*models.ServerWithMetrics) string {
	serversSection := ""
	if len(servers) == 0 {
		serversSection = `
			<div class="text-center py-8">
				<p class="text-gray-500">No servers found</p>
				<a href="/servers/create" class="mt-4 inline-block bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
					Create Server
				</a>
			</div>
		`
	} else {
		for _, swm := range servers {
			serversSection += fmt.Sprintf(`
				<tr>
					<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">%s</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
					<td class="px-6 py-4 whitespace-nowrap">
						<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full %s">
							%s
						</span>
					</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
					<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
				</tr>
			`, swm.Server.Name, swm.Server.Provider, swm.Server.Region, swm.Server.IPAddress,
				getStatusBadgeColor(swm.Server.Status), swm.Server.Status,
				swm.GetCPUDisplay(), swm.GetMemoryDisplay(), swm.GetDiskDisplay())
		}
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>VIP Hosting Panel - Servers</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
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
                    <span class="text-sm text-gray-700">%s (%s)</span>
                    <a href="/dashboard" class="text-blue-600 hover:text-blue-800">Dashboard</a>
                    <a href="/logout" class="text-red-600 hover:text-red-800">Logout</a>
                </div>
            </div>
        </div>
    </nav>

    <!-- Main Content -->
    <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div class="px-4 py-6 sm:px-0">
            <div class="flex justify-between items-center mb-6">
                <h2 class="text-2xl font-bold text-gray-900">Servers</h2>
                <a href="/servers/create" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
                    Create Server
                </a>
            </div>
            
            <!-- Servers Table -->
            <div class="bg-white shadow overflow-hidden sm:rounded-lg">
                <table class="min-w-full divide-y divide-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">IP Address</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPU</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">RAM</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Disk</th>
                        </tr>
                    </thead>
                    <tbody class="bg-white divide-y divide-gray-200" 
                           hx-get="/api/v1/servers" 
                           hx-trigger="every 30s"
                           hx-swap="innerHTML">
                        %s
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</body>
</html>
	`, email, role, serversSection)
}

// renderCreateServerHTML renders the server creation form
func (h *ServerHandler) renderCreateServerHTML(email, role string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>VIP Hosting Panel - Create Server</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
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
                    <span class="text-sm text-gray-700">%s (%s)</span>
                    <a href="/dashboard" class="text-blue-600 hover:text-blue-800">Dashboard</a>
                    <a href="/servers" class="text-blue-600 hover:text-blue-800">Servers</a>
                    <a href="/logout" class="text-red-600 hover:text-red-800">Logout</a>
                </div>
            </div>
        </div>
    </nav>

    <!-- Main Content -->
    <div class="max-w-3xl mx-auto py-6 sm:px-6 lg:px-8">
        <div class="px-4 py-6 sm:px-0">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Create Server</h2>
            
            <div class="bg-white shadow rounded-lg p-6">
                <form method="POST" action="/servers">
                    <div class="grid grid-cols-1 gap-6">
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Server Name</label>
                            <input type="text" name="name" required 
                                   class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                        </div>
                        
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Provider</label>
                            <select name="provider" required 
                                    class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                                <option value="">Select Provider</option>
                                <option value="hetzner">Hetzner</option>
                                <option value="digitalocean">DigitalOcean</option>
                                <option value="vultr">Vultr</option>
                                <option value="aws">AWS</option>
                            </select>
                        </div>
                        
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Region</label>
                            <input type="text" name="region" required placeholder="e.g., us-east-1, fra1"
                                   class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                        </div>
                        
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Plan</label>
                            <input type="text" name="plan" required placeholder="e.g., cx11, s-1vcpu-1gb"
                                   class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                        </div>
                    </div>
                    
                    <div class="mt-6 flex justify-end space-x-3">
                        <a href="/servers" class="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50">
                            Cancel
                        </a>
                        <button type="submit" class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600">
                            Create Server
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</body>
</html>
	`, email, role)
}

// Helper function for status badge colors
func getStatusBadgeColor(status string) string {
	switch status {
	case models.ServerStatusReady:
		return "bg-green-100 text-green-800"
	case models.ServerStatusProvisioning:
		return "bg-yellow-100 text-yellow-800"
	case models.ServerStatusQueued:
		return "bg-blue-100 text-blue-800"
	case models.ServerStatusFailed:
		return "bg-red-100 text-red-800"
	default:
		return "bg-gray-100 text-gray-800"
	}
}