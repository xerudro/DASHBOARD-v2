package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Site represents a hosted website
type Site struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	TenantID      uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	ServerID      uuid.UUID   `json:"server_id" db:"server_id"`
	Name          string      `json:"name" db:"name"`
	Domain        string      `json:"domain" db:"domain"`
	Type          string      `json:"type" db:"type"`
	PHPVersion    *string     `json:"php_version,omitempty" db:"php_version"`
	NodeJSVersion *string     `json:"nodejs_version,omitempty" db:"nodejs_version"`
	Webserver     string      `json:"webserver" db:"webserver"`
	RootPath      *string     `json:"root_path,omitempty" db:"root_path"`
	Status        string      `json:"status" db:"status"`
	GitRepo       *string     `json:"git_repo,omitempty" db:"git_repo"`
	GitBranch     *string     `json:"git_branch,omitempty" db:"git_branch"`
	SSLEnabled    bool        `json:"ssl_enabled" db:"ssl_enabled"`
	SSLAutoRenew  bool        `json:"ssl_auto_renew" db:"ssl_auto_renew"`
	Config        SiteConfig  `json:"config,omitempty" db:"config"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
	DeployedAt    *time.Time  `json:"deployed_at,omitempty" db:"deployed_at"`
	DeletedAt     *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SiteType constants
const (
	SiteTypePHP      = "php"
	SiteTypeNodeJS   = "nodejs"
	SiteTypeStatic   = "static"
	SiteTypePython   = "python"
	SiteTypeWordPress = "wordpress"
	SiteTypeLaravel  = "laravel"
)

// SiteStatus constants
const (
	SiteStatusPending   = "pending"
	SiteStatusDeploying = "deploying"
	SiteStatusActive    = "active"
	SiteStatusSuspended = "suspended"
	SiteStatusFailed    = "failed"
)

// SiteConfig represents site-specific configuration
type SiteConfig struct {
	CacheEnabled     bool              `json:"cache_enabled,omitempty"`
	WAFEnabled       bool              `json:"waf_enabled,omitempty"`
	CompressionEnabled bool            `json:"compression_enabled,omitempty"`
	BasicAuthEnabled bool              `json:"basic_auth_enabled,omitempty"`
	EnvVars          map[string]string `json:"env_vars,omitempty"`
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`
	Redirects        []Redirect        `json:"redirects,omitempty"`
}

// Redirect represents a URL redirect rule
type Redirect struct {
	From       string `json:"from"`
	To         string `json:"to"`
	StatusCode int    `json:"status_code"`
	Type       string `json:"type"` // permanent, temporary
}

// Value implements driver.Valuer for SiteConfig
func (sc SiteConfig) Value() (driver.Value, error) {
	return json.Marshal(sc)
}

// Scan implements sql.Scanner for SiteConfig
func (sc *SiteConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sc)
}

// IsDeployed returns true if site is deployed
func (s *Site) IsDeployed() bool {
	return s.Status == SiteStatusActive && s.DeployedAt != nil
}

// IsActive returns true if site is active
func (s *Site) IsActive() bool {
	return s.Status == SiteStatusActive
}

// GetStatusBadge returns a color for status display
func (s *Site) GetStatusBadge() string {
	switch s.Status {
	case SiteStatusActive:
		return "success"
	case SiteStatusDeploying:
		return "warning"
	case SiteStatusFailed:
		return "error"
	case SiteStatusSuspended:
		return "neutral"
	default:
		return "neutral"
	}
}

// GetFullURL returns the full URL of the site
func (s *Site) GetFullURL() string {
	protocol := "http"
	if s.SSLEnabled {
		protocol = "https"
	}
	return protocol + "://" + s.Domain
}
