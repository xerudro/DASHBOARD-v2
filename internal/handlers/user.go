package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/vip-hosting-panel/internal/middleware"
	"github.com/vip-hosting-panel/internal/repository"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userRepo *repository.UserRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// List returns users for the current tenant (API) - Admin only
func (h *UserHandler) List(c *fiber.Ctx) error {
	_, tenantID, _, role := middleware.GetUserFromContext(c)

	// Only admins can list users
	if !middleware.IsAdmin(role) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Insufficient permissions",
		})
	}

	// Parse pagination parameters
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := h.userRepo.GetByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load users",
		})
	}

	// Convert to safe response format (no password hashes)
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Status:    user.Status,
		}
	}

	// Get total count
	totalCount, err := h.userRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count users")
		totalCount = len(responses)
	}

	return c.JSON(fiber.Map{
		"users": responses,
		"pagination": fiber.Map{
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetProfile returns current user's profile (API)
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, _, _, _ := middleware.GetUserFromContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load profile",
		})
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Status:    user.Status,
	}

	return c.JSON(response)
}

// UpdateProfile updates current user's profile (API)
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, _, _, _ := middleware.GetUserFromContext(c)

	var req struct {
		FirstName string `json:"first_name" validate:"min=2,max=50"`
		LastName  string `json:"last_name" validate:"min=2,max=50"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request format",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current user
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load profile",
		})
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := h.userRepo.Update(ctx, user); err != nil {
		log.Error().Err(err).Msg("Failed to update user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to update profile",
		})
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Status:    user.Status,
	}

	log.Info().
		Str("user_id", userID.String()).
		Msg("User profile updated")

	return c.JSON(response)
}