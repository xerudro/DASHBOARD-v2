package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TenantID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Name             string     `json:"name" db:"name"`
	Role             string     `json:"role" db:"role"`
	Status           string     `json:"status" db:"status"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	TwoFactorEnabled bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret  *string    `json:"-" db:"two_factor_secret"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserRole constants
const (
	RoleSuperAdmin = "superadmin"
	RoleAdmin      = "admin"
	RoleReseller   = "reseller"
	RoleClient     = "client"
)

// UserStatus constants
const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusInactive  = "inactive"
)

// IsActive returns true if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// IsSuperAdmin returns true if user has superadmin role
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsAdmin returns true if user has admin role or higher
func (u *User) IsAdmin() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleAdmin
}

// IsReseller returns true if user has reseller role or higher
func (u *User) IsReseller() bool {
	return u.Role == RoleSuperAdmin || u.Role == RoleAdmin || u.Role == RoleReseller
}

// CanManageTenant checks if user can manage the given tenant
func (u *User) CanManageTenant(tenantID uuid.UUID) bool {
	if u.IsSuperAdmin() {
		return true // Superadmin can manage all tenants
	}
	return u.TenantID == tenantID
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Token        string    `json:"token" db:"token"`
	RefreshToken *string   `json:"refresh_token,omitempty" db:"refresh_token"`
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// IsExpired returns true if session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
