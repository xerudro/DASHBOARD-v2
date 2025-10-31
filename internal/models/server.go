package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Server represents a managed server
type Server struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	TenantID         uuid.UUID     `json:"tenant_id" db:"tenant_id"`
	ProviderID       uuid.UUID     `json:"provider_id" db:"provider_id"`
	Name             string        `json:"name" db:"name"`
	Hostname         *string       `json:"hostname,omitempty" db:"hostname"`
	IPAddress        *string       `json:"ip_address,omitempty" db:"ip_address"`
	ProviderServerID *string       `json:"provider_server_id,omitempty" db:"provider_server_id"`
	Region           *string       `json:"region,omitempty" db:"region"`
	Size             *string       `json:"size,omitempty" db:"size"`
	OS               *string       `json:"os,omitempty" db:"os"`
	Status           string        `json:"status" db:"status"`
	SSHPort          int           `json:"ssh_port" db:"ssh_port"`
	SSHKey           *string       `json:"-" db:"ssh_key"`
	Specs            ServerSpecs   `json:"specs,omitempty" db:"specs"`
	Tags             []string      `json:"tags,omitempty" db:"tags"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
	ProvisionedAt    *time.Time    `json:"provisioned_at,omitempty" db:"provisioned_at"`
	DeletedAt        *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ServerStatus constants
const (
	ServerStatusQueued       = "queued"
	ServerStatusProvisioning = "provisioning"
	ServerStatusReady        = "ready"
	ServerStatusDegraded     = "degraded"
	ServerStatusFailed       = "failed"
	ServerStatusStopped      = "stopped"
	ServerStatusTerminated   = "terminated"
)

// ServerSpecs represents server specifications
type ServerSpecs struct {
	CPUCores   int    `json:"cpu_cores,omitempty"`
	RAMTotal   int64  `json:"ram_total_mb,omitempty"`
	DiskTotal  int64  `json:"disk_total_gb,omitempty"`
	DiskType   string `json:"disk_type,omitempty"`
	Bandwidth  int64  `json:"bandwidth_gb,omitempty"`
	IPv4Count  int    `json:"ipv4_count,omitempty"`
	IPv6       bool   `json:"ipv6,omitempty"`
}

// Value implements driver.Valuer for ServerSpecs
func (s ServerSpecs) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for ServerSpecs
func (s *ServerSpecs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// IsProvisioned returns true if server is fully provisioned
func (s *Server) IsProvisioned() bool {
	return s.Status == ServerStatusReady && s.ProvisionedAt != nil
}

// IsAvailable returns true if server is available for hosting sites
func (s *Server) IsAvailable() bool {
	return s.Status == ServerStatusReady || s.Status == ServerStatusDegraded
}

// GetStatusBadge returns a color for status display
func (s *Server) GetStatusBadge() string {
	switch s.Status {
	case ServerStatusReady:
		return "success"
	case ServerStatusProvisioning:
		return "warning"
	case ServerStatusFailed:
		return "error"
	case ServerStatusStopped:
		return "neutral"
	default:
		return "neutral"
	}
}
