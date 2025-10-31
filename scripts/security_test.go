package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SecurityTestSuite runs comprehensive security tests against the API
type SecurityTestSuite struct {
	baseURL string
	client  *http.Client
	token   string
}

// TestResults holds the results of security tests
type TestResults struct {
	TestName    string
	Passed      bool
	Description string
	Details     string
}

// NewSecurityTestSuite creates a new security test suite
func NewSecurityTestSuite(baseURL string) *SecurityTestSuite {
	return &SecurityTestSuite{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RunAllTests executes all security tests
func (s *SecurityTestSuite) RunAllTests() []TestResults {
	fmt.Println("🔒 Running comprehensive security test suite...")
	
	var results []TestResults
	
	// Test 1: Security Headers
	results = append(results, s.testSecurityHeaders())
	
	// Test 2: CORS Configuration
	results = append(results, s.testCORSConfiguration())
	
	// Test 3: Rate Limiting
	results = append(results, s.testRateLimiting())
	
	// Test 4: Input Validation
	results = append(results, s.testInputValidation())
	
	// Test 5: SQL Injection Protection
	results = append(results, s.testSQLInjectionProtection())
	
	// Test 6: XSS Protection
	results = append(results, s.testXSSProtection())
	
	// Test 7: Authentication Security
	results = append(results, s.testAuthenticationSecurity())
	
	// Test 8: JWT Token Security
	results = append(results, s.testJWTTokenSecurity())
	
	// Test 9: Unauthorized Access
	results = append(results, s.testUnauthorizedAccess())
	
	// Test 10: Information Disclosure
	results = append(results, s.testInformationDisclosure())
	
	return results
}

// testSecurityHeaders verifies security headers are properly set
func (s *SecurityTestSuite) testSecurityHeaders() TestResults {
	resp, err := s.client.Get(s.baseURL + "/health")
	if err != nil {
		return TestResults{
			TestName:    "Security Headers",
			Passed:      false,
			Description: "Check if security headers are properly configured",
			Details:     fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()
	
	// Check required security headers
	requiredHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
	}
	
	var missingHeaders []string
	for header, expectedValue := range requiredHeaders {
		actualValue := resp.Header.Get(header)
		if !strings.Contains(actualValue, expectedValue) {
			missingHeaders = append(missingHeaders, fmt.Sprintf("%s (expected: %s, got: %s)", header, expectedValue, actualValue))
		}
	}
	
	if len(missingHeaders) == 0 {
		return TestResults{
			TestName:    "Security Headers",
			Passed:      true,
			Description: "Check if security headers are properly configured",
			Details:     "All required security headers are present",
		}
	}
	
	return TestResults{
		TestName:    "Security Headers",
		Passed:      false,
		Description: "Check if security headers are properly configured",
		Details:     fmt.Sprintf("Missing or incorrect headers: %s", strings.Join(missingHeaders, ", ")),
	}
}

// testCORSConfiguration verifies CORS is properly configured
func (s *SecurityTestSuite) testCORSConfiguration() TestResults {
	req, _ := http.NewRequest("OPTIONS", s.baseURL+"/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://malicious-site.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return TestResults{
			TestName:    "CORS Configuration",
			Passed:      false,
			Description: "Check CORS configuration prevents unauthorized origins",
			Details:     fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()
	
	// Check that malicious origin is not allowed
	allowedOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowedOrigin == "*" || allowedOrigin == "https://malicious-site.com" {
		return TestResults{
			TestName:    "CORS Configuration",
			Passed:      false,
			Description: "Check CORS configuration prevents unauthorized origins",
			Details:     fmt.Sprintf("CORS allows unauthorized origin: %s", allowedOrigin),
		}
	}
	
	return TestResults{
		TestName:    "CORS Configuration",
		Passed:      true,
		Description: "Check CORS configuration prevents unauthorized origins",
		Details:     "CORS properly restricts origins",
	}
}

// testRateLimiting verifies rate limiting is working
func (s *SecurityTestSuite) testRateLimiting() TestResults {
	endpoint := s.baseURL + "/api/v1/auth/login"
	
	// Send many requests rapidly
	successCount := 0
	rateLimitedCount := 0
	
	for i := 0; i < 15; i++ {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonPayload, _ := json.Marshal(payload)
		
		resp, err := s.client.Post(endpoint, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
		} else {
			successCount++
		}
		resp.Body.Close()
		
		time.Sleep(100 * time.Millisecond)
	}
	
	if rateLimitedCount > 0 {
		return TestResults{
			TestName:    "Rate Limiting",
			Passed:      true,
			Description: "Check if rate limiting prevents abuse",
			Details:     fmt.Sprintf("Rate limiting working: %d requests blocked", rateLimitedCount),
		}
	}
	
	return TestResults{
		TestName:    "Rate Limiting",
		Passed:      false,
		Description: "Check if rate limiting prevents abuse",
		Details:     "No rate limiting detected - all requests succeeded",
	}
}

// testInputValidation verifies input validation is working
func (s *SecurityTestSuite) testInputValidation() TestResults {
	maliciousPayloads := []map[string]interface{}{
		{"email": "", "password": "test"},                    // Empty email
		{"email": "invalid-email", "password": "test"},       // Invalid email format
		{"email": "test@test.com", "password": ""},           // Empty password
		{"email": "test@test.com", "password": "ab"},         // Too short password
		{"email": strings.Repeat("a", 300) + "@test.com"},    // Too long email
		{"email": "test@test.com", "password": strings.Repeat("a", 1000)}, // Too long password
	}
	
	validationErrorCount := 0
	
	for _, payload := range maliciousPayloads {
		jsonPayload, _ := json.Marshal(payload)
		resp, err := s.client.Post(s.baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusBadRequest {
			validationErrorCount++
		}
		resp.Body.Close()
	}
	
	if validationErrorCount >= len(maliciousPayloads)-1 { // Allow some tolerance
		return TestResults{
			TestName:    "Input Validation",
			Passed:      true,
			Description: "Check if input validation prevents malicious input",
			Details:     fmt.Sprintf("Input validation working: %d/%d payloads rejected", validationErrorCount, len(maliciousPayloads)),
		}
	}
	
	return TestResults{
		TestName:    "Input Validation",
		Passed:      false,
		Description: "Check if input validation prevents malicious input",
		Details:     fmt.Sprintf("Insufficient validation: only %d/%d payloads rejected", validationErrorCount, len(maliciousPayloads)),
	}
}

// testSQLInjectionProtection verifies SQL injection protection
func (s *SecurityTestSuite) testSQLInjectionProtection() TestResults {
	sqlInjectionPayloads := []string{
		"admin@test.com' OR '1'='1",
		"admin@test.com'; DROP TABLE users; --",
		"admin@test.com' UNION SELECT * FROM users --",
		"admin@test.com' AND 1=1 --",
	}
	
	protectedCount := 0
	
	for _, payload := range sqlInjectionPayloads {
		loginPayload := map[string]string{
			"email":    payload,
			"password": "password123",
		}
		jsonPayload, _ := json.Marshal(loginPayload)
		
		resp, err := s.client.Post(s.baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			continue
		}
		
		// Check that we get proper validation error, not internal server error
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
			protectedCount++
		}
		resp.Body.Close()
	}
	
	if protectedCount == len(sqlInjectionPayloads) {
		return TestResults{
			TestName:    "SQL Injection Protection",
			Passed:      true,
			Description: "Check protection against SQL injection attacks",
			Details:     "All SQL injection payloads properly handled",
		}
	}
	
	return TestResults{
		TestName:    "SQL Injection Protection",
		Passed:      false,
		Description: "Check protection against SQL injection attacks",
		Details:     fmt.Sprintf("Potential vulnerability: %d/%d payloads not properly handled", len(sqlInjectionPayloads)-protectedCount, len(sqlInjectionPayloads)),
	}
}

// testXSSProtection verifies XSS protection
func (s *SecurityTestSuite) testXSSProtection() TestResults {
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"<svg onload=alert('xss')>",
	}
	
	protectedCount := 0
	
	for _, payload := range xssPayloads {
		loginPayload := map[string]string{
			"email":    "test@test.com",
			"password": payload,
		}
		jsonPayload, _ := json.Marshal(loginPayload)
		
		resp, err := s.client.Post(s.baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			continue
		}
		
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		
		// Check that XSS payload is not reflected in response
		if !strings.Contains(bodyStr, payload) {
			protectedCount++
		}
		resp.Body.Close()
	}
	
	if protectedCount == len(xssPayloads) {
		return TestResults{
			TestName:    "XSS Protection",
			Passed:      true,
			Description: "Check protection against XSS attacks",
			Details:     "No XSS payloads reflected in responses",
		}
	}
	
	return TestResults{
		TestName:    "XSS Protection",
		Passed:      false,
		Description: "Check protection against XSS attacks",
		Details:     fmt.Sprintf("Potential XSS vulnerability: %d/%d payloads reflected", len(xssPayloads)-protectedCount, len(xssPayloads)),
	}
}

// testAuthenticationSecurity verifies authentication security
func (s *SecurityTestSuite) testAuthenticationSecurity() TestResults {
	// Test 1: Weak password should be rejected
	weakPasswords := []string{
		"123456",
		"password",
		"admin",
		"qwerty",
		"abc123",
	}
	
	rejectedCount := 0
	for _, password := range weakPasswords {
		payload := map[string]string{
			"email":     "test@example.com",
			"firstName": "Test",
			"lastName":  "User",
			"password":  password,
		}
		jsonPayload, _ := json.Marshal(payload)
		
		resp, err := s.client.Post(s.baseURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusBadRequest {
			rejectedCount++
		}
		resp.Body.Close()
	}
	
	if rejectedCount >= len(weakPasswords)/2 { // At least half should be rejected
		return TestResults{
			TestName:    "Authentication Security",
			Passed:      true,
			Description: "Check authentication security measures",
			Details:     fmt.Sprintf("Password policy working: %d/%d weak passwords rejected", rejectedCount, len(weakPasswords)),
		}
	}
	
	return TestResults{
		TestName:    "Authentication Security",
		Passed:      false,
		Description: "Check authentication security measures",
		Details:     fmt.Sprintf("Weak password policy: only %d/%d weak passwords rejected", rejectedCount, len(weakPasswords)),
	}
}

// testJWTTokenSecurity verifies JWT token security
func (s *SecurityTestSuite) testJWTTokenSecurity() TestResults {
	// Test invalid JWT tokens
	invalidTokens := []string{
		"invalid.jwt.token",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
		"",
		"Bearer",
		"Bearer ",
	}
	
	rejectedCount := 0
	for _, token := range invalidTokens {
		req, _ := http.NewRequest("GET", s.baseURL+"/api/v1/dashboard", nil)
		if token != "" {
			req.Header.Set("Authorization", token)
		}
		
		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusUnauthorized {
			rejectedCount++
		}
		resp.Body.Close()
	}
	
	if rejectedCount == len(invalidTokens) {
		return TestResults{
			TestName:    "JWT Token Security",
			Passed:      true,
			Description: "Check JWT token validation",
			Details:     "All invalid tokens properly rejected",
		}
	}
	
	return TestResults{
		TestName:    "JWT Token Security",
		Passed:      false,
		Description: "Check JWT token validation",
		Details:     fmt.Sprintf("Token validation issue: %d/%d invalid tokens not rejected", len(invalidTokens)-rejectedCount, len(invalidTokens)),
	}
}

// testUnauthorizedAccess verifies unauthorized access is prevented
func (s *SecurityTestSuite) testUnauthorizedAccess() TestResults {
	protectedEndpoints := []string{
		"/api/v1/dashboard",
		"/api/v1/servers",
		"/api/v1/users",
		"/api/v1/dashboard/stats",
	}
	
	blockedCount := 0
	for _, endpoint := range protectedEndpoints {
		resp, err := s.client.Get(s.baseURL + endpoint)
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusUnauthorized {
			blockedCount++
		}
		resp.Body.Close()
	}
	
	if blockedCount == len(protectedEndpoints) {
		return TestResults{
			TestName:    "Unauthorized Access",
			Passed:      true,
			Description: "Check protection of authenticated endpoints",
			Details:     "All protected endpoints require authentication",
		}
	}
	
	return TestResults{
		TestName:    "Unauthorized Access",
		Passed:      false,
		Description: "Check protection of authenticated endpoints",
		Details:     fmt.Sprintf("Access control issue: %d/%d endpoints accessible without auth", len(protectedEndpoints)-blockedCount, len(protectedEndpoints)),
	}
}

// testInformationDisclosure verifies no sensitive information is disclosed
func (s *SecurityTestSuite) testInformationDisclosure() TestResults {
	// Test error responses don't leak sensitive info
	resp, err := s.client.Get(s.baseURL + "/api/v1/nonexistent")
	if err != nil {
		return TestResults{
			TestName:    "Information Disclosure",
			Passed:      false,
			Description: "Check for information disclosure in error responses",
			Details:     fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.ToLower(string(body))
	
	// Check for sensitive information in error responses
	sensitiveInfo := []string{
		"database",
		"sql",
		"postgres",
		"redis",
		"stack trace",
		"panic",
		"internal server error",
		"jwt secret",
		"password",
	}
	
	var foundSensitive []string
	for _, info := range sensitiveInfo {
		if strings.Contains(bodyStr, info) {
			foundSensitive = append(foundSensitive, info)
		}
	}
	
	if len(foundSensitive) == 0 {
		return TestResults{
			TestName:    "Information Disclosure",
			Passed:      true,
			Description: "Check for information disclosure in error responses",
			Details:     "No sensitive information disclosed in error responses",
		}
	}
	
	return TestResults{
		TestName:    "Information Disclosure",
		Passed:      false,
		Description: "Check for information disclosure in error responses",
		Details:     fmt.Sprintf("Potential information disclosure: %s", strings.Join(foundSensitive, ", ")),
	}
}

// PrintResults prints the test results in a formatted way
func PrintResults(results []TestResults) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔒 SECURITY TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	
	passedCount := 0
	totalCount := len(results)
	
	for _, result := range results {
		status := "❌ FAILED"
		if result.Passed {
			status = "✅ PASSED"
			passedCount++
		}
		
		fmt.Printf("\n%s - %s\n", status, result.TestName)
		fmt.Printf("   Description: %s\n", result.Description)
		fmt.Printf("   Details: %s\n", result.Details)
	}
	
	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Printf("📊 SUMMARY: %d/%d tests passed (%.1f%%)\n", 
		passedCount, totalCount, float64(passedCount)/float64(totalCount)*100)
	
	if passedCount == totalCount {
		fmt.Println("🎉 All security tests passed! Your application is well-secured.")
	} else if passedCount >= totalCount*3/4 {
		fmt.Println("⚠️  Most security tests passed, but some issues need attention.")
	} else {
		fmt.Println("🚨 Multiple security issues detected. Please address them immediately.")
	}
	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run security_test.go <base_url>")
		fmt.Println("Example: go run security_test.go http://localhost:8080")
		return
	}
	
	baseURL := os.Args[1]
	suite := NewSecurityTestSuite(baseURL)
	results := suite.RunAllTests()
	PrintResults(results)
}