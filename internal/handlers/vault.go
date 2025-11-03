package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xerudro/DASHBOARD-v2/internal/middleware"
	"github.com/xerudro/DASHBOARD-v2/internal/vault"
)

// VaultHandler handles vault-related HTTP requests
type VaultHandler struct {
	vault *vault.Vault
}

// NewVaultHandler creates a new vault handler
func NewVaultHandler(v *vault.Vault) *VaultHandler {
	return &VaultHandler{vault: v}
}

// RegisterRoutes registers vault routes
func (h *VaultHandler) RegisterRoutes(app *fiber.App, jwtMiddleware *middleware.JWTMiddleware) {
	v := app.Group("/api/vault")

	// Require authentication for all vault operations
	v.Use(jwtMiddleware.Protect())

	// Require SuperAdmin or Admin role for vault operations
	v.Use(jwtMiddleware.RequireRole("SuperAdmin", "Admin"))

	// Vault management
	v.Post("/unlock", h.Unlock)
	v.Post("/lock", h.Lock)
	v.Get("/status", h.Status)
	v.Get("/health", h.Health)

	// Secret operations
	v.Post("/secrets", h.CreateSecret)
	v.Get("/secrets", h.ListSecrets)
	v.Get("/secrets/*", h.GetSecret)
	v.Put("/secrets/*", h.UpdateSecret)
	v.Delete("/secrets/*", h.DeleteSecret)

	// Secret versioning
	v.Get("/secrets/*/versions", h.ListSecretVersions)
	v.Get("/secrets/*/versions/:version", h.GetSecretVersion)

	// Secret rotation
	v.Post("/secrets/*/rotate", h.RotateSecret)
	v.Post("/rotate-all", h.RotateAllSecrets)

	// Audit logs
	v.Get("/secrets/*/audit", h.GetAuditLogs)

	// Maintenance
	v.Post("/cleanup", h.CleanupExpiredSecrets)

	// Token generation
	v.Post("/generate-token", h.GenerateToken)
}

// Unlock unlocks the vault
func (h *VaultHandler) Unlock(c *fiber.Ctx) error {
	type UnlockRequest struct {
		MasterKey string `json:"master_key" validate:"required"`
	}

	var req UnlockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.vault.Unlock(req.MasterKey); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Vault unlocked successfully",
	})
}

// Lock locks the vault
func (h *VaultHandler) Lock(c *fiber.Ctx) error {
	h.vault.Lock()

	return c.JSON(fiber.Map{
		"message": "Vault locked successfully",
	})
}

// Status returns vault status
func (h *VaultHandler) Status(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"locked": h.vault.IsLocked(),
	})
}

// Health returns vault health information
func (h *VaultHandler) Health(c *fiber.Ctx) error {
	return c.JSON(h.vault.Health())
}

// CreateSecret creates a new secret
func (h *VaultHandler) CreateSecret(c *fiber.Ctx) error {
	type CreateSecretRequest struct {
		Path        string  `json:"path" validate:"required"`
		Value       string  `json:"value" validate:"required"`
		Description string  `json:"description"`
		ExpiresIn   *string `json:"expires_in"` // e.g., "24h", "7d", "30d"
	}

	var req CreateSecretRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user ID from JWT token
	userID := c.Locals("user_id").(uuid.UUID)

	// Parse expiration duration
	var expiresIn *time.Duration
	if req.ExpiresIn != nil {
		duration, err := time.ParseDuration(*req.ExpiresIn)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid expiration duration",
			})
		}
		expiresIn = &duration
	}

	// Create secret
	err := h.vault.CreateSecret(c.Context(), req.Path, req.Value, req.Description, userID, expiresIn)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		if err == vault.ErrSecretAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Secret already exists at this path",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create secret",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Secret created successfully",
		"path":    req.Path,
	})
}

// GetSecret retrieves a secret
func (h *VaultHandler) GetSecret(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is required",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	ipAddress := c.IP()

	value, err := h.vault.GetSecret(c.Context(), path, userID, ipAddress)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		if err == vault.ErrSecretNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Secret not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve secret",
		})
	}

	return c.JSON(fiber.Map{
		"path":  path,
		"value": value,
	})
}

// UpdateSecret updates a secret
func (h *VaultHandler) UpdateSecret(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is required",
		})
	}

	type UpdateSecretRequest struct {
		Value string `json:"value" validate:"required"`
	}

	var req UpdateSecretRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	err := h.vault.UpdateSecret(c.Context(), path, req.Value, userID)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		if err == vault.ErrSecretNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Secret not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update secret",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Secret updated successfully",
		"path":    path,
	})
}

// DeleteSecret deletes a secret
func (h *VaultHandler) DeleteSecret(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is required",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	err := h.vault.DeleteSecret(c.Context(), path, userID)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		if err == vault.ErrSecretNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Secret not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete secret",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Secret deleted successfully",
		"path":    path,
	})
}

// ListSecrets lists secrets under a path prefix
func (h *VaultHandler) ListSecrets(c *fiber.Ctx) error {
	prefix := c.Query("prefix", "")

	secrets, err := h.vault.ListSecrets(c.Context(), prefix)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list secrets",
		})
	}

	return c.JSON(fiber.Map{
		"secrets": secrets,
		"count":   len(secrets),
	})
}

// ListSecretVersions lists all versions of a secret
func (h *VaultHandler) ListSecretVersions(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is required",
		})
	}

	versions, err := h.vault.ListSecretVersions(c.Context(), path)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list secret versions",
		})
	}

	return c.JSON(fiber.Map{
		"versions": versions,
		"count":    len(versions),
	})
}

// GetSecretVersion retrieves a specific version of a secret
func (h *VaultHandler) GetSecretVersion(c *fiber.Ctx) error {
	path := c.Params("*")
	versionStr := c.Params("version")

	if path == "" || versionStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path and version are required",
		})
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid version number",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	value, err := h.vault.GetSecretVersion(c.Context(), path, version, userID)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		if err == vault.ErrSecretNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Secret version not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve secret version",
		})
	}

	return c.JSON(fiber.Map{
		"path":    path,
		"version": version,
		"value":   value,
	})
}

// RotateSecret rotates a secret with a new master key
func (h *VaultHandler) RotateSecret(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is required",
		})
	}

	type RotateRequest struct {
		NewMasterKey string `json:"new_master_key" validate:"required"`
	}

	var req RotateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	err := h.vault.RotateSecret(c.Context(), path, req.NewMasterKey, userID)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to rotate secret",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Secret rotated successfully",
		"path":    path,
	})
}

// RotateAllSecrets rotates all secrets with a new master key
func (h *VaultHandler) RotateAllSecrets(c *fiber.Ctx) error {
	type RotateAllRequest struct {
		NewMasterKey string `json:"new_master_key" validate:"required"`
	}

	var req RotateAllRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	err := h.vault.RotateAllSecrets(c.Context(), req.NewMasterKey, userID)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to rotate all secrets",
		})
	}

	return c.JSON(fiber.Map{
		"message": "All secrets rotated successfully",
	})
}

// GetAuditLogs retrieves audit logs for a secret
func (h *VaultHandler) GetAuditLogs(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Path is required",
		})
	}

	limit := c.QueryInt("limit", 100)
	if limit > 1000 {
		limit = 1000
	}

	logs, err := h.vault.GetAuditLogs(c.Context(), path, limit)
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve audit logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"count": len(logs),
	})
}

// CleanupExpiredSecrets removes expired secrets
func (h *VaultHandler) CleanupExpiredSecrets(c *fiber.Ctx) error {
	count, err := h.vault.CleanupExpiredSecrets(c.Context())
	if err != nil {
		if err == vault.ErrVaultLocked {
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{
				"error": "Vault is locked",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cleanup expired secrets",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Expired secrets cleaned up successfully",
		"count":   count,
	})
}

// GenerateToken generates a secure random token
func (h *VaultHandler) GenerateToken(c *fiber.Ctx) error {
	type GenerateTokenRequest struct {
		Length int `json:"length" validate:"required,min=16,max=128"`
	}

	var req GenerateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	token, err := h.vault.GenerateSecureToken(req.Length)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token":  token,
		"length": req.Length,
	})
}
