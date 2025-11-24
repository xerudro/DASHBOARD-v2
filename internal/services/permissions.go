package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/xerudro/DASHBOARD-v2/internal/models"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
)

// PlanDefinition defines the allowed server creation features for a tenant plan.
type PlanDefinition struct {
	AllowServerCreation   bool `mapstructure:"allow_server_creation"`
	AllowVPSServers       bool `mapstructure:"allow_vps_servers"`
	AllowDedicatedServers bool `mapstructure:"allow_dedicated_servers"`
}

// PermissionService enforces plan-based permissions for provisioning operations.
type PermissionService struct {
	tenantRepo      *repository.TenantRepository
	planDefinitions map[string]PlanDefinition
	defaultPlan     PlanDefinition
}

// PermissionChecker exposes the contract used by handlers.
type PermissionChecker interface {
	CanCreateServer(ctx context.Context, tenantID uuid.UUID, role, serverClass string) (bool, error)
}

var defaultPlanDefinitions = map[string]PlanDefinition{
	"basic":     {AllowServerCreation: false},
	"starter":   {AllowServerCreation: true, AllowVPSServers: true},
	"vps":       {AllowServerCreation: true, AllowVPSServers: true},
	"vps-pro":   {AllowServerCreation: true, AllowVPSServers: true},
	"dedicated": {AllowServerCreation: true, AllowVPSServers: true, AllowDedicatedServers: true},
	"reseller":  {AllowServerCreation: true, AllowVPSServers: true, AllowDedicatedServers: true},
}

// NewPermissionService creates a PermissionService with optional custom plan definitions.
func NewPermissionService(tenantRepo *repository.TenantRepository, customDefinitions map[string]PlanDefinition) *PermissionService {
	merged := make(map[string]PlanDefinition)
	for key, def := range defaultPlanDefinitions {
		merged[key] = def
	}
	for key, def := range customDefinitions {
		merged[strings.ToLower(key)] = def
	}

	defaultPlan := merged["basic"]

	return &PermissionService{
		tenantRepo:      tenantRepo,
		planDefinitions: merged,
		defaultPlan:     defaultPlan,
	}
}

// CanCreateServer determines whether the tenant role is allowed to provision the requested server class.
func (s *PermissionService) CanCreateServer(ctx context.Context, tenantID uuid.UUID, role, serverClass string) (bool, error) {
	if role == models.RoleSuperAdmin || role == models.RoleAdmin {
		return true, nil
	}

	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return false, fmt.Errorf("failed to resolve tenant for permissions: %w", err)
	}

	plan := strings.ToLower(strings.TrimSpace(tenant.Plan))
	definition, ok := s.planDefinitions[plan]
	if !ok {
		definition = s.defaultPlan
	}

	switch strings.ToLower(serverClass) {
	case "dedicated":
		return definition.AllowDedicatedServers, nil
	default:
		if definition.AllowServerCreation {
			return true, nil
		}
		if definition.AllowVPSServers {
			return true, nil
		}
		return false, nil
	}
}
