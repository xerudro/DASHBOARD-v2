package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/xerudro/DASHBOARD-v2/internal/audit"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
	"github.com/xerudro/DASHBOARD-v2/internal/services/providers"
)

const (
	TypeServerProvisioning = "server:provision"
	TypeServerDeletion     = "server:delete"
	TypeServerResize       = "server:resize"
)

// ServerProvisioningPayload contains server provisioning job data
type ServerProvisioningPayload struct {
	ServerID   string
	TenantID   string
	UserID     int64
	Provider   string
	ServerType string
	Location   string
	Image      string
	SSHKeys    []int64
	UserData   string
	Labels     map[string]string
}

// ServerDeletionPayload contains server deletion job data
type ServerDeletionPayload struct {
	ServerID   string
	TenantID   string
	UserID     int64
	Provider   string
	ProviderID string
}

// ServerResizePayload contains server resize job data
type ServerResizePayload struct {
	ServerID      string
	TenantID      string
	UserID        int64
	Provider      string
	ProviderID    string
	NewServerType string
}

// ServerProvisioningJob handles server provisioning
type ServerProvisioningJob struct {
	hetznerProvider *providers.HetznerProvider
	serverRepo      *repository.ServerRepository
	auditLogger     *audit.AuditLogger
}

// NewServerProvisioningJob creates a new server provisioning job handler
func NewServerProvisioningJob(
	hetznerProvider *providers.HetznerProvider,
	serverRepo *repository.ServerRepository,
	auditLogger *audit.AuditLogger,
) *ServerProvisioningJob {
	return &ServerProvisioningJob{
		hetznerProvider: hetznerProvider,
		serverRepo:      serverRepo,
		auditLogger:     auditLogger,
	}
}

// EnqueueServerProvisioning enqueues a server provisioning job
func EnqueueServerProvisioning(client *asynq.Client, payload ServerProvisioningPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeServerProvisioning, data)

	info, err := client.Enqueue(task,
		asynq.Queue("critical"),
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("task_id", info.ID).
		Str("server_id", payload.ServerID).
		Msg("Server provisioning job enqueued")

	return nil
}

// ProcessServerProvisioning processes a server provisioning job
func (j *ServerProvisioningJob) ProcessServerProvisioning(ctx context.Context, task *asynq.Task) error {
	var payload ServerProvisioningPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("server_id", payload.ServerID).
		Str("provider", payload.Provider).
		Str("type", payload.ServerType).
		Msg("Starting server provisioning")

	// Update server status to provisioning
	err := j.updateServerStatus(ctx, payload.ServerID, "provisioning", "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to update server status")
	}

	// Provision server based on provider
	var providerServerID string
	var publicIP string

	switch payload.Provider {
	case "hetzner":
		serverInfo, err := j.provisionHetznerServer(ctx, payload)
		if err != nil {
			j.updateServerStatus(ctx, payload.ServerID, "failed", err.Error())
			j.logProvisioningFailure(payload, err)
			return fmt.Errorf("Hetzner provisioning failed: %w", err)
		}
		providerServerID = fmt.Sprintf("%d", serverInfo.ID)
		publicIP = serverInfo.PublicIPv4

	default:
		err := fmt.Errorf("unsupported provider: %s", payload.Provider)
		j.updateServerStatus(ctx, payload.ServerID, "failed", err.Error())
		return err
	}

	// Update server with provider information
	err = j.updateServerProvisioned(ctx, payload.ServerID, providerServerID, publicIP)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update server after provisioning")
		return err
	}

	// Wait for server to be ready
	err = j.waitForServerReady(ctx, payload.Provider, providerServerID)
	if err != nil {
		log.Error().Err(err).Msg("Server failed to become ready")
		j.updateServerStatus(ctx, payload.ServerID, "failed", err.Error())
		return err
	}

	// Final status update
	err = j.updateServerStatus(ctx, payload.ServerID, "ready", "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to update final server status")
	}

	// Log success
	j.logProvisioningSuccess(payload, providerServerID, publicIP)

	log.Info().
		Str("server_id", payload.ServerID).
		Str("provider_id", providerServerID).
		Str("ip", publicIP).
		Msg("Server provisioning completed successfully")

	return nil
}

// provisionHetznerServer provisions a server on Hetzner Cloud
func (j *ServerProvisioningJob) provisionHetznerServer(ctx context.Context, payload ServerProvisioningPayload) (*providers.ServerInfo, error) {
	config := providers.ServerConfig{
		Name:             payload.ServerID, // Use our internal ID as name
		ServerType:       payload.ServerType,
		Location:         payload.Location,
		Image:            payload.Image,
		SSHKeys:          payload.SSHKeys,
		UserData:         payload.UserData,
		Labels:           payload.Labels,
		StartAfterCreate: true,
	}

	serverInfo, err := j.hetznerProvider.CreateServer(ctx, config)
	if err != nil {
		return nil, err
	}

	return serverInfo, nil
}

// waitForServerReady waits for server to become ready
func (j *ServerProvisioningJob) waitForServerReady(ctx context.Context, provider string, providerID string) error {
	maxAttempts := 60 // 5 minutes with 5-second intervals
	interval := 5 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			// Check server status
			switch provider {
			case "hetzner":
				serverID := parseHetznerID(providerID)
				if serverID == 0 {
					return fmt.Errorf("invalid Hetzner server ID: %s", providerID)
				}

				info, err := j.hetznerProvider.GetServer(ctx, serverID)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to get server status")
					continue
				}

				if info.Status == "running" {
					log.Info().
						Str("provider_id", providerID).
						Int("attempts", attempt+1).
						Msg("Server is ready")
					return nil
				}

				log.Debug().
					Str("provider_id", providerID).
					Str("status", info.Status).
					Int("attempt", attempt+1).
					Msg("Waiting for server to be ready")
			}
		}
	}

	return fmt.Errorf("server did not become ready within timeout")
}

// updateServerStatus updates server status in database
func (j *ServerProvisioningJob) updateServerStatus(ctx context.Context, serverID string, status string, errorMsg string) error {
	// This would update the database
	// Implementation depends on your repository pattern
	log.Info().
		Str("server_id", serverID).
		Str("status", status).
		Msg("Updating server status")

	return nil
}

// updateServerProvisioned updates server with provider details
func (j *ServerProvisioningJob) updateServerProvisioned(ctx context.Context, serverID string, providerID string, publicIP string) error {
	log.Info().
		Str("server_id", serverID).
		Str("provider_id", providerID).
		Str("ip", publicIP).
		Msg("Updating server with provider details")

	return nil
}

// logProvisioningSuccess logs successful provisioning
func (j *ServerProvisioningJob) logProvisioningSuccess(payload ServerProvisioningPayload, providerID string, publicIP string) {
	if j.auditLogger == nil {
		return
	}

	j.auditLogger.LogSecurityEvent(
		"server_provisioned",
		"info",
		fmt.Sprintf("Server %s provisioned successfully", payload.ServerID),
		map[string]interface{}{
			"server_id":   payload.ServerID,
			"provider":    payload.Provider,
			"provider_id": providerID,
			"server_type": payload.ServerType,
			"location":    payload.Location,
			"ip_address":  publicIP,
			"tenant_id":   payload.TenantID,
			"user_id":     payload.UserID,
		},
	)
}

// logProvisioningFailure logs provisioning failure
func (j *ServerProvisioningJob) logProvisioningFailure(payload ServerProvisioningPayload, err error) {
	if j.auditLogger == nil {
		return
	}

	j.auditLogger.LogSecurityEvent(
		"server_provisioning_failed",
		"error",
		fmt.Sprintf("Server %s provisioning failed: %v", payload.ServerID, err),
		map[string]interface{}{
			"server_id":   payload.ServerID,
			"provider":    payload.Provider,
			"server_type": payload.ServerType,
			"location":    payload.Location,
			"error":       err.Error(),
			"tenant_id":   payload.TenantID,
			"user_id":     payload.UserID,
		},
	)
}

// ProcessServerDeletion processes a server deletion job
func (j *ServerProvisioningJob) ProcessServerDeletion(ctx context.Context, task *asynq.Task) error {
	var payload ServerDeletionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("server_id", payload.ServerID).
		Str("provider_id", payload.ProviderID).
		Msg("Starting server deletion")

	// Delete server from provider
	switch payload.Provider {
	case "hetzner":
		serverID := parseHetznerID(payload.ProviderID)
		if serverID == 0 {
			return fmt.Errorf("invalid Hetzner server ID: %s", payload.ProviderID)
		}

		err := j.hetznerProvider.DeleteServer(ctx, serverID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to delete Hetzner server")
			return err
		}

	default:
		return fmt.Errorf("unsupported provider: %s", payload.Provider)
	}

	// Update database status
	err := j.updateServerStatus(ctx, payload.ServerID, "deleted", "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to update server status")
	}

	// Log deletion
	if j.auditLogger != nil {
		j.auditLogger.LogSecurityEvent(
			"server_deleted",
			"info",
			fmt.Sprintf("Server %s deleted successfully", payload.ServerID),
			map[string]interface{}{
				"server_id":   payload.ServerID,
				"provider":    payload.Provider,
				"provider_id": payload.ProviderID,
				"tenant_id":   payload.TenantID,
				"user_id":     payload.UserID,
			},
		)
	}

	log.Info().
		Str("server_id", payload.ServerID).
		Msg("Server deletion completed")

	return nil
}

// ProcessServerResize processes a server resize job
func (j *ServerProvisioningJob) ProcessServerResize(ctx context.Context, task *asynq.Task) error {
	var payload ServerResizePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("server_id", payload.ServerID).
		Str("new_type", payload.NewServerType).
		Msg("Starting server resize")

	// Resize server
	switch payload.Provider {
	case "hetzner":
		serverID := parseHetznerID(payload.ProviderID)
		if serverID == 0 {
			return fmt.Errorf("invalid Hetzner server ID: %s", payload.ProviderID)
		}

		err := j.hetznerProvider.ResizeServer(ctx, serverID, payload.NewServerType)
		if err != nil {
			log.Error().Err(err).Msg("Failed to resize Hetzner server")
			return err
		}

	default:
		return fmt.Errorf("unsupported provider: %s", payload.Provider)
	}

	// Wait for resize to complete
	err := j.waitForServerReady(ctx, payload.Provider, payload.ProviderID)
	if err != nil {
		log.Error().Err(err).Msg("Server resize failed")
		return err
	}

	// Log resize
	if j.auditLogger != nil {
		j.auditLogger.LogSecurityEvent(
			"server_resized",
			"info",
			fmt.Sprintf("Server %s resized to %s", payload.ServerID, payload.NewServerType),
			map[string]interface{}{
				"server_id":       payload.ServerID,
				"provider":        payload.Provider,
				"provider_id":     payload.ProviderID,
				"new_server_type": payload.NewServerType,
				"tenant_id":       payload.TenantID,
				"user_id":         payload.UserID,
			},
		)
	}

	log.Info().
		Str("server_id", payload.ServerID).
		Msg("Server resize completed")

	return nil
}

// parseHetznerID converts string ID to int64
func parseHetznerID(id string) int64 {
	var serverID int64
	fmt.Sscanf(id, "%d", &serverID)
	return serverID
}
