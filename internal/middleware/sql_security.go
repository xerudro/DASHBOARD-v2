package middleware

import (
	"strings"
	
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// SQLSecurityMiddleware provides protection against SQL injection attacks
func SQLSecurityMiddleware() fiber.Handler {
	// Common SQL injection patterns to detect
	dangerousPatterns := []string{
		"'",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"union",
		"select",
		"insert",
		"delete",
		"update",
		"drop",
		"create",
		"alter",
		"exec",
		"execute",
		"script",
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
	}

	return func(c *fiber.Ctx) error {
		// Check request body for SQL injection attempts
		body := string(c.Body())
		if body != "" {
			bodyLower := strings.ToLower(body)
			for _, pattern := range dangerousPatterns {
				if strings.Contains(bodyLower, pattern) {
					log.Warn().
						Str("ip", c.IP()).
						Str("pattern", pattern).
						Str("path", c.Path()).
						Msg("Potential SQL injection attempt detected in body")
					
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid request format",
					})
				}
			}
		}

		// Check query parameters
		c.Context().QueryArgs().VisitAll(func(key, value []byte) {
			valueLower := strings.ToLower(string(value))
			for _, pattern := range dangerousPatterns {
				if strings.Contains(valueLower, pattern) {
					log.Warn().
						Str("ip", c.IP()).
						Str("pattern", pattern).
						Str("param", string(key)).
						Str("path", c.Path()).
						Msg("Potential SQL injection attempt detected in query param")
				}
			}
		})

		return c.Next()
	}
}

// SafeQueryBuilder helps build safe SQL queries with parameter validation
type SafeQueryBuilder struct {
	allowedTables []string
	allowedFields []string
}

// NewSafeQueryBuilder creates a new safe query builder
func NewSafeQueryBuilder() *SafeQueryBuilder {
	return &SafeQueryBuilder{
		allowedTables: []string{
			"users", "tenants", "servers", "sites", "domains", 
			"metrics", "backups", "ssl_certificates", "audit_logs",
		},
		allowedFields: []string{
			"id", "name", "email", "created_at", "updated_at",
			"status", "type", "provider", "region",
		},
	}
}

// ValidateTableName ensures table name is in allowlist
func (sqb *SafeQueryBuilder) ValidateTableName(tableName string) bool {
	for _, allowed := range sqb.allowedTables {
		if allowed == tableName {
			return true
		}
	}
	return false
}

// ValidateFieldName ensures field name is in allowlist
func (sqb *SafeQueryBuilder) ValidateFieldName(fieldName string) bool {
	for _, allowed := range sqb.allowedFields {
		if allowed == fieldName {
			return true
		}
	}
	return false
}

// SanitizeInput removes potentially dangerous characters
func (sqb *SafeQueryBuilder) SanitizeInput(input string) string {
	// Remove common SQL injection characters
	replacements := map[string]string{
		"'":  "",
		"\"": "",
		";":  "",
		"--": "",
		"/*": "",
		"*/": "",
	}
	
	result := input
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	
	return result
}