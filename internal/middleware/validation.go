package middleware

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// ValidationResponse represents validation error response
type ValidationResponse struct {
	Error   bool              `json:"error"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("safe_string", validateSafeString)
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []ValidationError {
	var errors []ValidationError
	
	if err := validate.Struct(s); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   strings.ToLower(err.Field()),
				Tag:     err.Tag(),
				Value:   err.Param(),
				Message: getErrorMessage(err),
			})
		}
	}
	
	return errors
}

// ValidationMiddleware returns a middleware that validates request bodies
func ValidationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Store validator in context for use in handlers
		c.Locals("validator", validate)
		return c.Next()
	}
}

// Helper function to generate user-friendly error messages
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters long"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters long"
	case "strong_password":
		return fe.Field() + " must contain at least 8 characters with uppercase, lowercase, number and special character"
	case "safe_string":
		return fe.Field() + " contains unsafe characters"
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	default:
		return fe.Field() + " is invalid"
	}
}

// Custom validator for strong passwords
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)
	
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// Custom validator for safe strings (prevent XSS)
func validateSafeString(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	
	// Check for common XSS patterns
	dangerous := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"eval(",
		"expression(",
	}
	
	lowerStr := strings.ToLower(str)
	for _, pattern := range dangerous {
		if strings.Contains(lowerStr, pattern) {
			log.Warn().
				Str("input", str).
				Str("pattern", pattern).
				Msg("Potentially dangerous input detected")
			return false
		}
	}
	
	return true
}

// ValidateAndRespond validates a struct and returns error response if invalid
func ValidateAndRespond(c *fiber.Ctx, s interface{}) *ValidationResponse {
	errors := ValidateStruct(s)
	if len(errors) > 0 {
		return &ValidationResponse{
			Error:   true,
			Message: "Validation failed",
			Errors:  errors,
		}
	}
	return nil
}