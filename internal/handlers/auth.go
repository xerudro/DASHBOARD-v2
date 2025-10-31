package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/vip-hosting/panel-v2/internal/auth"
	"github.com/vip-hosting/panel-v2/internal/middleware"
	"github.com/vip-hosting/panel-v2/internal/models"
	"github.com/vip-hosting/panel-v2/internal/repository"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo *repository.UserRepository
	jwt      *middleware.JWTMiddleware
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	// JWT config will be injected from main
	return &AuthHandler{
		userRepo: userRepo,
	}
}

// SetJWT sets the JWT middleware (called from main after config is loaded)
func (h *AuthHandler) SetJWT(jwt *middleware.JWTMiddleware) {
	h.jwt = jwt
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,safe_string"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email,safe_string"`
	Password  string `json:"password" validate:"required,strong_password"`
	FirstName string `json:"first_name" validate:"required,min=2,max=50,safe_string"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50,safe_string"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success      bool   `json:"success"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         *UserResponse `json:"user,omitempty"`
	Message      string `json:"message,omitempty"`
}

// UserResponse represents user in API responses
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
}

// Login handles user login (API)
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Str("ip", c.IP()).Msg("Invalid login request format")
		return c.Status(fiber.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	// Validate input using middleware validator
	if validationResp := middleware.ValidateAndRespond(c, &req); validationResp != nil {
		log.Warn().
			Str("email", req.Email).
			Str("ip", c.IP()).
			Interface("errors", validationResp.Errors).
			Msg("Login validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: validationResp.Message,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Hash password to compare
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password for comparison")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Authentication failed",
		})
	}

	// Find user
	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Warn().
			Str("email", req.Email).
			Str("ip", c.IP()).
			Msg("Login attempt with invalid email")
		return c.Status(fiber.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	// Verify password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		log.Warn().
			Str("email", req.Email).
			Str("user_id", user.ID.String()).
			Str("ip", c.IP()).
			Msg("Login attempt with invalid password")
		return c.Status(fiber.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return c.Status(fiber.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Account is not active",
		})
	}

	// Generate tokens
	token, err := h.jwt.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Authentication failed",
		})
	}

	refreshToken, err := h.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Authentication failed",
		})
	}

	// Update last login
	if err := h.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Warn().Err(err).Msg("Failed to update last login")
	}

	// Set cookie for web interface
	h.jwt.SetTokenCookie(c, token)

	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("role", user.Role).
		Str("tenant_id", user.TenantID.String()).
		Str("ip", c.IP()).
		Str("user_agent", c.Get("User-Agent")).
		Msg("User login successful")

	return c.JSON(LoginResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Status:    user.Status,
		},
	})
}

// Register handles user registration (API)
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "All fields are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(LoginResponse{
			Success: false,
			Message: "User already exists",
		})
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Registration failed",
		})
	}

	// For now, assign to default tenant (this should be improved)
	// In a real implementation, you'd handle tenant creation properly
	defaultTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // From seed data

	// Create user
	user := &models.User{
		TenantID:     defaultTenantID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RoleClient,
		Status:       models.UserStatusActive,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Registration failed",
		})
	}

	log.Info().
		Str("email", user.Email).
		Str("role", user.Role).
		Msg("User registered")

	return c.Status(fiber.StatusCreated).JSON(LoginResponse{
		Success: true,
		Message: "Registration successful",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "Invalid request format",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "Refresh token is required",
		})
	}

	// Validate refresh token
	userID, err := h.jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Invalid refresh token",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "User not found",
		})
	}

	// Generate new tokens
	token, err := h.jwt.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Token refresh failed",
		})
	}

	refreshToken, err := h.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Token refresh failed",
		})
	}

	// Set cookie
	h.jwt.SetTokenCookie(c, token)

	return c.JSON(LoginResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// LoginPage renders login page (HTML)
func (h *AuthHandler) LoginPage(c *fiber.Ctx) error {
	// This would render a Templ template
	// For now, return a simple HTML response
	return c.Type("html").SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>VIP Hosting Panel - Login</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-md p-6">
            <h2 class="text-2xl font-bold text-center mb-6">Login</h2>
            <form method="POST" action="/login">
                <div class="mb-4">
                    <label class="block text-sm font-medium mb-2">Email</label>
                    <input type="email" name="email" required 
                           class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <div class="mb-6">
                    <label class="block text-sm font-medium mb-2">Password</label>
                    <input type="password" name="password" required 
                           class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <button type="submit" 
                        class="w-full bg-blue-500 text-white rounded-lg py-2 hover:bg-blue-600 transition-colors">
                    Login
                </button>
            </form>
            <p class="text-center mt-4">
                <a href="/register" class="text-blue-500 hover:underline">Register</a>
            </p>
        </div>
    </div>
</body>
</html>
	`)
}

// RegisterPage renders registration page (HTML)
func (h *AuthHandler) RegisterPage(c *fiber.Ctx) error {
	return c.Type("html").SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>VIP Hosting Panel - Register</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-md p-6">
            <h2 class="text-2xl font-bold text-center mb-6">Register</h2>
            <form method="POST" action="/register">
                <div class="mb-4">
                    <label class="block text-sm font-medium mb-2">First Name</label>
                    <input type="text" name="first_name" required 
                           class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <div class="mb-4">
                    <label class="block text-sm font-medium mb-2">Last Name</label>
                    <input type="text" name="last_name" required 
                           class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <div class="mb-4">
                    <label class="block text-sm font-medium mb-2">Email</label>
                    <input type="email" name="email" required 
                           class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <div class="mb-6">
                    <label class="block text-sm font-medium mb-2">Password</label>
                    <input type="password" name="password" required 
                           class="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <button type="submit" 
                        class="w-full bg-blue-500 text-white rounded-lg py-2 hover:bg-blue-600 transition-colors">
                    Register
                </button>
            </form>
            <p class="text-center mt-4">
                <a href="/login" class="text-blue-500 hover:underline">Login</a>
            </p>
        </div>
    </div>
</body>
</html>
	`)
}

// LoginForm handles form-based login (HTML)
func (h *AuthHandler) LoginForm(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return c.Redirect("/login?error=missing_fields")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user
	user, err := h.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return c.Redirect("/login?error=invalid_credentials")
	}

	// Verify password
	if !auth.CheckPassword(password, user.PasswordHash) {
		return c.Redirect("/login?error=invalid_credentials")
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return c.Redirect("/login?error=account_inactive")
	}

	// Generate token
	token, err := h.jwt.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return c.Redirect("/login?error=login_failed")
	}

	// Update last login
	if err := h.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Warn().Err(err).Msg("Failed to update last login")
	}

	// Set cookie
	h.jwt.SetTokenCookie(c, token)

	log.Info().
		Str("email", user.Email).
		Str("role", user.Role).
		Msg("User logged in via form")

	return c.Redirect("/dashboard")
}

// RegisterForm handles form-based registration (HTML)
func (h *AuthHandler) RegisterForm(c *fiber.Ctx) error {
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		return c.Redirect("/register?error=missing_fields")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return c.Redirect("/register?error=user_exists")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return c.Redirect("/register?error=registration_failed")
	}

	// Create user
	defaultTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	user := &models.User{
		TenantID:     defaultTenantID,
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.RoleClient,
		Status:       models.UserStatusActive,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return c.Redirect("/register?error=registration_failed")
	}

	log.Info().
		Str("email", user.Email).
		Str("role", user.Role).
		Msg("User registered via form")

	return c.Redirect("/login?success=registration_complete")
}