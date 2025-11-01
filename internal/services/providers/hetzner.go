package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/rs/zerolog/log"
	"github.com/xerudro/DASHBOARD-v2/internal/cache"
)

// HetznerProvider implements Hetzner Cloud API integration
type HetznerProvider struct {
	client *hcloud.Client
	cache  *cache.RedisCache
}

// ServerConfig represents server creation configuration
type ServerConfig struct {
	Name             string
	ServerType       string // cx11, cx21, cx31, etc.
	Location         string // fsn1, nbg1, hel1, ash
	Image            string // ubuntu-22.04, debian-11, etc.
	SSHKeys          []int64
	UserData         string
	Labels           map[string]string
	Firewalls        []int64
	Networks         []int64
	Volumes          []int64
	StartAfterCreate bool
}

// ServerInfo represents detailed server information
type ServerInfo struct {
	ID              int64
	Name            string
	Status          string
	ServerType      string
	PublicIPv4      string
	PublicIPv6      string
	PrivateIPv4     string
	Location        string
	Datacenter      string
	Created         time.Time
	IncludedTraffic int64
	OutgoingTraffic int64
	IncomingTraffic int64
	Labels          map[string]string
	Protection      Protection
	Rescue          bool
	Locked          bool
	BackupWindow    string
	Image           ImageInfo
}

// ImageInfo represents OS image information
type ImageInfo struct {
	ID          int64
	Name        string
	Description string
	Type        string
	OSFlavor    string
	OSVersion   string
	RapidDeploy bool
}

// Protection represents server protection settings
type Protection struct {
	Delete  bool
	Rebuild bool
}

// Pricing represents Hetzner pricing information
type Pricing struct {
	ServerTypes map[string]ServerTypePricing
	Traffic     TrafficPricing
	FloatingIPs FloatingIPPricing
	Volumes     VolumePricing
	Backups     BackupPricing
}

// ServerTypePricing represents pricing for a server type
type ServerTypePricing struct {
	Hourly  float64
	Monthly float64
	Name    string
	Cores   int
	Memory  float64 // GB
	Disk    int     // GB
}

// TrafficPricing represents traffic pricing
type TrafficPricing struct {
	PerTB float64
}

// FloatingIPPricing represents floating IP pricing
type FloatingIPPricing struct {
	Monthly float64
}

// VolumePricing represents volume pricing
type VolumePricing struct {
	PerGBMonthly float64
}

// BackupPricing represents backup pricing
type BackupPricing struct {
	Percentage float64 // Percentage of server price
}

// NewHetznerProvider creates a new Hetzner provider instance
func NewHetznerProvider(token string, cache *cache.RedisCache) (*HetznerProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("Hetzner API token is required")
	}

	client := hcloud.NewClient(hcloud.WithToken(token))

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _, err := client.ServerType.List(ctx, hcloud.ServerTypeListOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Hetzner API: %w", err)
	}

	log.Info().Msg("Successfully connected to Hetzner Cloud API")

	return &HetznerProvider{
		client: client,
		cache:  cache,
	}, nil
}

// CreateServer creates a new server on Hetzner Cloud
func (h *HetznerProvider) CreateServer(ctx context.Context, config ServerConfig) (*ServerInfo, error) {
	log.Info().
		Str("name", config.Name).
		Str("type", config.ServerType).
		Str("location", config.Location).
		Msg("Creating Hetzner server")

	// Get server type
	serverType, _, err := h.client.ServerType.GetByName(ctx, config.ServerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get server type: %w", err)
	}
	if serverType == nil {
		return nil, fmt.Errorf("server type %s not found", config.ServerType)
	}

	// Get location
	location, _, err := h.client.Location.GetByName(ctx, config.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	if location == nil {
		return nil, fmt.Errorf("location %s not found", config.Location)
	}

	// Get image
	image, _, err := h.client.Image.GetByName(ctx, config.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	if image == nil {
		return nil, fmt.Errorf("image %s not found", config.Image)
	}

	// Prepare SSH keys
	var sshKeys []*hcloud.SSHKey
	for _, keyID := range config.SSHKeys {
		key, _, err := h.client.SSHKey.GetByID(ctx, keyID)
		if err != nil {
			log.Warn().Int64("key_id", keyID).Msg("SSH key not found, skipping")
			continue
		}
		sshKeys = append(sshKeys, key)
	}

	// Prepare firewalls
	var firewalls []*hcloud.ServerCreateFirewall
	for _, fwID := range config.Firewalls {
		firewalls = append(firewalls, &hcloud.ServerCreateFirewall{
			Firewall: hcloud.Firewall{ID: fwID},
		})
	}

	// Create server
	createOpts := hcloud.ServerCreateOpts{
		Name:             config.Name,
		ServerType:       serverType,
		Image:            image,
		Location:         location,
		SSHKeys:          sshKeys,
		UserData:         config.UserData,
		Labels:           config.Labels,
		Firewalls:        firewalls,
		StartAfterCreate: &config.StartAfterCreate,
	}

	result, _, err := h.client.Server.Create(ctx, createOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	log.Info().
		Int64("server_id", result.Server.ID).
		Str("name", result.Server.Name).
		Msg("Server created successfully")

	// Invalidate cache
	if h.cache != nil {
		h.cache.DeleteByTag(ctx, "servers")
	}

	return h.convertServerInfo(result.Server), nil
}

// GetServer retrieves server information by ID
func (h *HetznerProvider) GetServer(ctx context.Context, serverID int64) (*ServerInfo, error) {
	// Try cache first
	if h.cache != nil {
		cacheKey := fmt.Sprintf("hetzner:server:%d", serverID)
		var cachedServer ServerInfo
		found, err := h.cache.Get(ctx, cacheKey, &cachedServer)
		if err == nil && found {
			return &cachedServer, nil
		}
	}

	server, _, err := h.client.Server.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("server %d not found", serverID)
	}

	info := h.convertServerInfo(server)

	// Cache result
	if h.cache != nil {
		cacheKey := fmt.Sprintf("hetzner:server:%d", serverID)
		h.cache.Set(ctx, cacheKey, info, cache.CacheOptions{
			TTL:  1 * time.Minute,
			Tags: []string{"servers", "hetzner"},
		})
	}

	return info, nil
}

// ListServers lists all servers
func (h *HetznerProvider) ListServers(ctx context.Context) ([]*ServerInfo, error) {
	// Try cache first
	if h.cache != nil {
		cacheKey := "hetzner:servers:all"
		var cachedServers []*ServerInfo
		found, err := h.cache.Get(ctx, cacheKey, &cachedServers)
		if err == nil && found {
			return cachedServers, nil
		}
	}

	servers, err := h.client.Server.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	result := make([]*ServerInfo, len(servers))
	for i, server := range servers {
		result[i] = h.convertServerInfo(server)
	}

	// Cache result
	if h.cache != nil {
		cacheKey := "hetzner:servers:all"
		h.cache.Set(ctx, cacheKey, result, cache.CacheOptions{
			TTL:  30 * time.Second,
			Tags: []string{"servers", "hetzner"},
		})
	}

	return result, nil
}

// DeleteServer deletes a server
func (h *HetznerProvider) DeleteServer(ctx context.Context, serverID int64) error {
	log.Info().Int64("server_id", serverID).Msg("Deleting Hetzner server")

	server := &hcloud.Server{ID: serverID}
	_, _, err := h.client.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	// Invalidate cache
	if h.cache != nil {
		h.cache.DeleteByTag(ctx, "servers")
	}

	log.Info().Int64("server_id", serverID).Msg("Server deleted successfully")
	return nil
}

// PowerOnServer powers on a server
func (h *HetznerProvider) PowerOnServer(ctx context.Context, serverID int64) error {
	server := &hcloud.Server{ID: serverID}
	_, _, err := h.client.Server.Poweron(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to power on server: %w", err)
	}

	// Invalidate cache
	if h.cache != nil {
		cacheKey := fmt.Sprintf("hetzner:server:%d", serverID)
		h.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// PowerOffServer powers off a server
func (h *HetznerProvider) PowerOffServer(ctx context.Context, serverID int64) error {
	server := &hcloud.Server{ID: serverID}
	_, _, err := h.client.Server.Poweroff(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to power off server: %w", err)
	}

	// Invalidate cache
	if h.cache != nil {
		cacheKey := fmt.Sprintf("hetzner:server:%d", serverID)
		h.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// RebootServer reboots a server
func (h *HetznerProvider) RebootServer(ctx context.Context, serverID int64) error {
	server := &hcloud.Server{ID: serverID}
	_, _, err := h.client.Server.Reboot(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to reboot server: %w", err)
	}

	return nil
}

// ResizeServer changes the server type
func (h *HetznerProvider) ResizeServer(ctx context.Context, serverID int64, newServerType string) error {
	log.Info().
		Int64("server_id", serverID).
		Str("new_type", newServerType).
		Msg("Resizing Hetzner server")

	serverType, _, err := h.client.ServerType.GetByName(ctx, newServerType)
	if err != nil {
		return fmt.Errorf("failed to get server type: %w", err)
	}

	server := &hcloud.Server{ID: serverID}
	_, _, err = h.client.Server.ChangeType(ctx, server, hcloud.ServerChangeTypeOpts{
		ServerType:  serverType,
		UpgradeDisk: true,
	})
	if err != nil {
		return fmt.Errorf("failed to resize server: %w", err)
	}

	// Invalidate cache
	if h.cache != nil {
		h.cache.DeleteByTag(ctx, "servers")
	}

	log.Info().Int64("server_id", serverID).Msg("Server resize initiated")
	return nil
}

// GetPricing retrieves current Hetzner pricing
func (h *HetznerProvider) GetPricing(ctx context.Context) (*Pricing, error) {
	// Try cache first
	if h.cache != nil {
		cacheKey := "hetzner:pricing"
		var cachedPricing Pricing
		found, err := h.cache.Get(ctx, cacheKey, &cachedPricing)
		if err == nil && found {
			return &cachedPricing, nil
		}
	}

	// TODO: Implement proper pricing from Hetzner API
	// pricing, _, err := h.client.Pricing.Get(ctx)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get pricing: %w", err)
	// }

	result := &Pricing{
		ServerTypes: make(map[string]ServerTypePricing),
		Traffic: TrafficPricing{
			PerTB: 1.0, // Default fallback - Hetzner API pricing structure varies
		},
		FloatingIPs: FloatingIPPricing{
			Monthly: 0.0, // Default fallback
		},
		Volumes: VolumePricing{
			PerGBMonthly: 0.0, // Default fallback
		},
		Backups: BackupPricing{
			Percentage: 20.0, // Hetzner backup is 20% of server price
		},
	}

	// Get server types with pricing
	serverTypes, err := h.client.ServerType.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server types: %w", err)
	}

	// Map server types with their specifications
	// Note: Pricing details from Hetzner API require complex nested structure parsing
	// Using fallback values for now - this should be enhanced with proper API parsing
	for _, st := range serverTypes {
		result.ServerTypes[st.Name] = ServerTypePricing{
			Hourly:  0.0, // Fallback - requires proper Hetzner API price parsing
			Monthly: 0.0, // Fallback - requires proper Hetzner API price parsing
			Name:    st.Name,
			Cores:   st.Cores,
			Memory:  float64(st.Memory),
			Disk:    st.Disk,
		}
	}

	// Cache result for 1 hour
	if h.cache != nil {
		cacheKey := "hetzner:pricing"
		h.cache.Set(ctx, cacheKey, result, cache.CacheOptions{
			TTL:  1 * time.Hour,
			Tags: []string{"pricing", "hetzner"},
		})
	}

	return result, nil
}

// CalculateMonthlyCost calculates monthly cost for a server configuration
func (h *HetznerProvider) CalculateMonthlyCost(ctx context.Context, serverType string, withBackups bool) (float64, error) {
	pricing, err := h.GetPricing(ctx)
	if err != nil {
		return 0, err
	}

	st, ok := pricing.ServerTypes[serverType]
	if !ok {
		return 0, fmt.Errorf("server type %s not found in pricing", serverType)
	}

	cost := st.Monthly

	if withBackups {
		cost += cost * (pricing.Backups.Percentage / 100.0)
	}

	return cost, nil
}

// ListLocations lists all available Hetzner locations
func (h *HetznerProvider) ListLocations(ctx context.Context) ([]*hcloud.Location, error) {
	// Try cache first
	if h.cache != nil {
		cacheKey := "hetzner:locations"
		var cachedLocations []*hcloud.Location
		found, err := h.cache.Get(ctx, cacheKey, &cachedLocations)
		if err == nil && found {
			return cachedLocations, nil
		}
	}

	locations, err := h.client.Location.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	// Cache for 24 hours (locations rarely change)
	if h.cache != nil {
		cacheKey := "hetzner:locations"
		h.cache.Set(ctx, cacheKey, locations, cache.CacheOptions{
			TTL:  24 * time.Hour,
			Tags: []string{"locations", "hetzner"},
		})
	}

	return locations, nil
}

// ListServerTypes lists all available server types
func (h *HetznerProvider) ListServerTypes(ctx context.Context) ([]*hcloud.ServerType, error) {
	// Try cache first
	if h.cache != nil {
		cacheKey := "hetzner:server_types"
		var cachedTypes []*hcloud.ServerType
		found, err := h.cache.Get(ctx, cacheKey, &cachedTypes)
		if err == nil && found {
			return cachedTypes, nil
		}
	}

	serverTypes, err := h.client.ServerType.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list server types: %w", err)
	}

	// Cache for 24 hours
	if h.cache != nil {
		cacheKey := "hetzner:server_types"
		h.cache.Set(ctx, cacheKey, serverTypes, cache.CacheOptions{
			TTL:  24 * time.Hour,
			Tags: []string{"server_types", "hetzner"},
		})
	}

	return serverTypes, nil
}

// ListImages lists available OS images
func (h *HetznerProvider) ListImages(ctx context.Context) ([]*hcloud.Image, error) {
	// Only list official images
	images, err := h.client.Image.AllWithOpts(ctx, hcloud.ImageListOpts{
		Type: []hcloud.ImageType{hcloud.ImageTypeSystem},
		Sort: []string{"name:asc"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return images, nil
}

// CreateSSHKey creates a new SSH key
func (h *HetznerProvider) CreateSSHKey(ctx context.Context, name string, publicKey string) (*hcloud.SSHKey, error) {
	key, _, err := h.client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %w", err)
	}

	return key, nil
}

// ListSSHKeys lists all SSH keys
func (h *HetznerProvider) ListSSHKeys(ctx context.Context) ([]*hcloud.SSHKey, error) {
	keys, err := h.client.SSHKey.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	return keys, nil
}

// GetServerMetrics retrieves server metrics
func (h *HetznerProvider) GetServerMetrics(ctx context.Context, serverID int64, metricType string, start, end time.Time) (*hcloud.ServerMetrics, error) {
	server := &hcloud.Server{ID: serverID}

	metrics, _, err := h.client.Server.GetMetrics(ctx, server, hcloud.ServerGetMetricsOpts{
		Types: []hcloud.ServerMetricType{hcloud.ServerMetricType(metricType)},
		Start: start,
		End:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get server metrics: %w", err)
	}

	return metrics, nil
}

// convertServerInfo converts hcloud.Server to ServerInfo
func (h *HetznerProvider) convertServerInfo(server *hcloud.Server) *ServerInfo {
	info := &ServerInfo{
		ID:              server.ID,
		Name:            server.Name,
		Status:          string(server.Status),
		ServerType:      server.ServerType.Name,
		Location:        server.Datacenter.Location.Name,
		Datacenter:      server.Datacenter.Name,
		Created:         server.Created,
		IncludedTraffic: int64(server.IncludedTraffic),
		OutgoingTraffic: int64(server.OutgoingTraffic),
		IncomingTraffic: 0, // Not available in Hetzner API
		Labels:          server.Labels,
		Protection: Protection{
			Delete:  server.Protection.Delete,
			Rebuild: server.Protection.Rebuild,
		},
		Rescue:       server.RescueEnabled,
		Locked:       server.Locked,
		BackupWindow: server.BackupWindow,
	}

	// Public IPv4
	if server.PublicNet.IPv4.IP != nil {
		info.PublicIPv4 = server.PublicNet.IPv4.IP.String()
	}

	// Public IPv6
	if server.PublicNet.IPv6.IP != nil {
		info.PublicIPv6 = server.PublicNet.IPv6.IP.String()
	}

	// Private networks
	if len(server.PrivateNet) > 0 {
		info.PrivateIPv4 = server.PrivateNet[0].IP.String()
	}

	// Image information
	if server.Image != nil {
		info.Image = ImageInfo{
			ID:          server.Image.ID,
			Name:        server.Image.Name,
			Description: server.Image.Description,
			Type:        string(server.Image.Type),
			OSFlavor:    server.Image.OSFlavor,
			OSVersion:   server.Image.OSVersion,
			RapidDeploy: server.Image.RapidDeploy,
		}
	}

	return info
}

// ConvertToModel converts ServerInfo to models.Server
// NOTE: This function needs to be refactored to match the actual Server model structure
// The current Server model uses UUIDs and different field types
// Commenting out to prevent compilation errors - needs implementation
// func (h *HetznerProvider) ConvertToModel(info *ServerInfo, tenantID string, userID int64) *models.Server {}
