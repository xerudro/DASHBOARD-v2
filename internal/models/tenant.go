package models

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Slug           string     `json:"slug" db:"slug"`
	Plan           string     `json:"plan" db:"plan"`
	Status         string     `json:"status" db:"status"`
	ParentTenantID *uuid.UUID `json:"parent_tenant_id,omitempty" db:"parent_tenant_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TenantStatus constants
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusCanceled  = "canceled"
)

// IsActive returns true if tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive && t.DeletedAt == nil
}

// IsReseller returns true if tenant has a parent (is a reseller's client)
func (t *Tenant) IsReseller() bool {
	return t.ParentTenantID != nil
}
