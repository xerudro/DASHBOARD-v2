package middleware

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// RequestValidator provides comprehensive request validation
type RequestValidator struct {
	maxBodySize       int64
	maxHeaderSize     int
	maxURLLength      int
	allowedMethods    []string
	allowedContentTypes []string
	blockSuspiciousPatterns bool
}

// NewRequestValidator creates a new request validator with secure defaults
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		maxBodySize:    10 * 1024 * 1024, // 10MB default
		maxHeaderSize:  8 * 1024,         // 8KB default
		maxURLLength:   2048,              // 2KB default
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		allowedContentTypes: []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
			"text/plain",
			"application/xml",
		},
		blockSuspiciousPatterns: true,
	}
}

// Middleware returns the request validation middleware
func (rv *RequestValidator) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Validate HTTP method
		if err := rv.validateMethod(c); err != nil {
			return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		// Validate URL length
		if err := rv.validateURLLength(c); err != nil {
			return c.Status(fiber.StatusRequestURITooLong).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		// Validate headers size
		if err := rv.validateHeadersSize(c); err != nil {
			return c.Status(fiber.StatusRequestHeaderFieldsTooLarge).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		// Validate content type
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			if err := rv.validateContentType(c); err != nil {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"error":   true,
					"message": err.Error(),
				})
			}
		}

		// Validate body size
		if err := rv.validateBodySize(c); err != nil {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		// Check for suspicious patterns
		if rv.blockSuspiciousPatterns {
			if err := rv.checkSuspiciousPatterns(c); err != nil {
				log.Warn().
					Err(err).
					Str("ip", c.IP()).
					Str("path", c.Path()).
					Str("user_agent", c.Get("User-Agent")).
					Msg("Suspicious pattern detected")

				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid request",
				})
			}
		}

		// Validate specific headers
		if err := rv.validateHeaders(c); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.Next()
	}
}

// validateMethod checks if the HTTP method is allowed
func (rv *RequestValidator) validateMethod(c *fiber.Ctx) error {
	method := c.Method()
	for _, allowed := range rv.allowedMethods {
		if method == allowed {
			return nil
		}
	}
	return fmt.Errorf("HTTP method %s not allowed", method)
}

// validateURLLength checks if the URL is within acceptable length
func (rv *RequestValidator) validateURLLength(c *fiber.Ctx) error {
	url := c.OriginalURL()
	if len(url) > rv.maxURLLength {
		return fmt.Errorf("URL length exceeds maximum allowed (%d bytes)", rv.maxURLLength)
	}
	return nil
}

// validateHeadersSize checks if headers size is within limits
func (rv *RequestValidator) validateHeadersSize(c *fiber.Ctx) error {
	totalSize := 0
	c.Request().Header.VisitAll(func(key, value []byte) {
		totalSize += len(key) + len(value) + 2 // +2 for ": "
	})

	if totalSize > rv.maxHeaderSize {
		return fmt.Errorf("headers size exceeds maximum allowed (%d bytes)", rv.maxHeaderSize)
	}
	return nil
}

// validateContentType checks if the content type is allowed
func (rv *RequestValidator) validateContentType(c *fiber.Ctx) error {
	contentType := c.Get("Content-Type")
	if contentType == "" {
		return nil // Allow empty content type for some endpoints
	}

	// Parse content type (remove charset and other parameters)
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("invalid content type: %w", err)
	}

	for _, allowed := range rv.allowedContentTypes {
		if mediaType == allowed {
			return nil
		}
	}

	return fmt.Errorf("content type %s not allowed", mediaType)
}

// validateBodySize checks if the request body size is within limits
func (rv *RequestValidator) validateBodySize(c *fiber.Ctx) error {
	contentLength := c.Request().Header.ContentLength()

	if contentLength > int(rv.maxBodySize) {
		return fmt.Errorf("request body size exceeds maximum allowed (%d bytes)", rv.maxBodySize)
	}

	// Also limit the actual body reading
	if contentLength > 0 {
		c.Request().SetBodyStream(io.LimitReader(c.Request().BodyStream(), rv.maxBodySize))
	}

	return nil
}

// checkSuspiciousPatterns checks for suspicious patterns in the request
func (rv *RequestValidator) checkSuspiciousPatterns(c *fiber.Ctx) error {
	// Check URL for suspicious patterns
	url := strings.ToLower(c.OriginalURL())

	suspiciousPatterns := []string{
		"../",           // Path traversal
		"..\\",          // Path traversal (Windows)
		"%2e%2e",        // Encoded path traversal
		"<script",       // XSS attempt
		"javascript:",   // XSS attempt
		"onerror=",      // XSS attempt
		"onload=",       // XSS attempt
		"eval(",         // Code injection
		"exec(",         // Code injection
		";base64,",      // Base64 encoding (suspicious)
		"<?php",         // PHP code injection
		"<?xml",         // XML injection (in URL - suspicious)
		"union select",  // SQL injection
		"' or '1'='1",   // SQL injection
		"admin'--",      // SQL injection
		"<iframe",       // XSS/Clickjacking
		"cmd.exe",       // Command injection
		"/bin/bash",     // Command injection
		"/bin/sh",       // Command injection
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(url, pattern) {
			return fmt.Errorf("suspicious pattern detected: %s", pattern)
		}
	}

	// Check headers for suspicious patterns
	userAgent := strings.ToLower(c.Get("User-Agent"))
	suspiciousAgents := []string{
		"sqlmap",        // SQL injection tool
		"nikto",         // Web scanner
		"nmap",          // Port scanner
		"masscan",       // Port scanner
		"havij",         // SQL injection tool
		"acunetix",      // Web vulnerability scanner
		"nessus",        // Vulnerability scanner
		"openvas",       // Vulnerability scanner
		"metasploit",    // Exploitation framework
	}

	for _, agent := range suspiciousAgents {
		if strings.Contains(userAgent, agent) {
			return fmt.Errorf("suspicious user agent detected")
		}
	}

	return nil
}

// validateHeaders validates specific security-critical headers
func (rv *RequestValidator) validateHeaders(c *fiber.Ctx) error {
	// Check for null bytes in headers (can cause issues)
	var nullByteFound bool
	c.Request().Header.VisitAll(func(key, value []byte) {
		if bytes.Contains(key, []byte{0}) || bytes.Contains(value, []byte{0}) {
			nullByteFound = true
		}
	})

	if nullByteFound {
		return fmt.Errorf("null bytes detected in headers")
	}

	// Validate Host header
	host := c.Get("Host")
	if host == "" && c.Protocol() == "HTTP/1.1" {
		return fmt.Errorf("Host header is required for HTTP/1.1")
	}

	// Check for header injection attempts
	headers := []string{"Host", "User-Agent", "Referer", "Origin"}
	for _, header := range headers {
		value := c.Get(header)
		if strings.Contains(value, "\r") || strings.Contains(value, "\n") {
			return fmt.Errorf("header injection attempt detected")
		}
	}

	return nil
}

// SetMaxBodySize sets the maximum allowed body size
func (rv *RequestValidator) SetMaxBodySize(size int64) {
	rv.maxBodySize = size
	log.Info().
		Int64("max_body_size", size).
		Msg("Max body size updated")
}

// SetMaxHeaderSize sets the maximum allowed headers size
func (rv *RequestValidator) SetMaxHeaderSize(size int) {
	rv.maxHeaderSize = size
}

// SetMaxURLLength sets the maximum allowed URL length
func (rv *RequestValidator) SetMaxURLLength(length int) {
	rv.maxURLLength = length
}

// AddAllowedContentType adds an allowed content type
func (rv *RequestValidator) AddAllowedContentType(contentType string) {
	rv.allowedContentTypes = append(rv.allowedContentTypes, contentType)
}

// SetAllowedMethods sets the allowed HTTP methods
func (rv *RequestValidator) SetAllowedMethods(methods []string) {
	rv.allowedMethods = methods
}

// EnableSuspiciousPatternBlocking enables or disables suspicious pattern blocking
func (rv *RequestValidator) EnableSuspiciousPatternBlocking(enable bool) {
	rv.blockSuspiciousPatterns = enable
}

// FileUploadValidator provides validation for file uploads
type FileUploadValidator struct {
	maxFileSize      int64
	allowedMimeTypes []string
	allowedExtensions []string
	maxFiles         int
}

// NewFileUploadValidator creates a new file upload validator
func NewFileUploadValidator() *FileUploadValidator {
	return &FileUploadValidator{
		maxFileSize: 5 * 1024 * 1024, // 5MB default
		allowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"application/pdf",
			"text/plain",
		},
		allowedExtensions: []string{
			".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt",
		},
		maxFiles: 10,
	}
}

// Middleware returns the file upload validation middleware
func (fuv *FileUploadValidator) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only validate multipart form data
		if !strings.Contains(c.Get("Content-Type"), "multipart/form-data") {
			return c.Next()
		}

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid multipart form data",
			})
		}

		// Validate number of files
		totalFiles := 0
		for _, files := range form.File {
			totalFiles += len(files)
		}

		if totalFiles > fuv.maxFiles {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Too many files (max %d allowed)", fuv.maxFiles),
			})
		}

		// Validate each file
		for _, files := range form.File {
			for _, file := range files {
				// Check file size
				if file.Size > fuv.maxFileSize {
					return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
						"error":   true,
						"message": fmt.Sprintf("File %s exceeds maximum size (%d bytes)", file.Filename, fuv.maxFileSize),
					})
				}

				// Check file extension
				validExtension := false
				filename := strings.ToLower(file.Filename)
				for _, ext := range fuv.allowedExtensions {
					if strings.HasSuffix(filename, ext) {
						validExtension = true
						break
					}
				}

				if !validExtension {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   true,
						"message": fmt.Sprintf("File %s has invalid extension", file.Filename),
					})
				}

				// Check MIME type
				validMimeType := false
				for _, mimeType := range fuv.allowedMimeTypes {
					if file.Header.Get("Content-Type") == mimeType {
						validMimeType = true
						break
					}
				}

				if !validMimeType {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   true,
						"message": fmt.Sprintf("File %s has invalid MIME type", file.Filename),
					})
				}
			}
		}

		return c.Next()
	}
}

// SetMaxFileSize sets the maximum file size
func (fuv *FileUploadValidator) SetMaxFileSize(size int64) {
	fuv.maxFileSize = size
}

// SetMaxFiles sets the maximum number of files
func (fuv *FileUploadValidator) SetMaxFiles(count int) {
	fuv.maxFiles = count
}

// AddAllowedMimeType adds an allowed MIME type
func (fuv *FileUploadValidator) AddAllowedMimeType(mimeType string) {
	fuv.allowedMimeTypes = append(fuv.allowedMimeTypes, mimeType)
}

// AddAllowedExtension adds an allowed file extension
func (fuv *FileUploadValidator) AddAllowedExtension(extension string) {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	fuv.allowedExtensions = append(fuv.allowedExtensions, strings.ToLower(extension))
}
