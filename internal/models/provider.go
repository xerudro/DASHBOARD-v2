package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Provider represents a cloud infrastructure provider
type Provider struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	TenantID       uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	Name           string         `json:"name" db:"name"`
	Type           string         `json:"type" db:"type"`
	APIToken       string         `json:"-" db:"api_token"`
	Config         ProviderConfig `json:"config,omitempty" db:"config"`
	Status         string         `json:"status" db:"status"`
	LastVerifiedAt *time.Time     `json:"last_verified_at,omitempty" db:"last_verified_at"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// Provider status constants
const (
	ProviderStatusActive   = "active"
	ProviderStatusInactive = "inactive"
	ProviderStatusError    = "error"
)

// Provider type constants
const (
	ProviderTypeHetzner      = "hetzner"
	ProviderTypeDigitalOcean = "digitalocean"
	ProviderTypeVultr        = "vultr"
	ProviderTypeAWS          = "aws"
)

// ProviderConfig represents provider-specific configuration
type ProviderConfig struct {
	Region       string            `json:"region,omitempty"`
	DefaultImage string            `json:"default_image,omitempty"`
	SSHKeyID     string            `json:"ssh_key_id,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// Value implements driver.Valuer for ProviderConfig
func (p ProviderConfig) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for ProviderConfig
func (p *ProviderConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}

// IsActive returns true if provider is active and verified
func (p *Provider) IsActive() bool {
	return p.Status == ProviderStatusActive && p.LastVerifiedAt != nil
}

// GetDisplayName returns a user-friendly display name
func (p *Provider) GetDisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Type
}
