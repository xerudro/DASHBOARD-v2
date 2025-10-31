#!/bin/bash

# VIP Hosting Panel v2 - Static Security Analysis
# Comprehensive security testing of code and configuration

TEST_COUNT=0
PASSED_COUNT=0
FAILED_COUNT=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_result() {
    local status=$1
    local test_name=$2
    local details=$3
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    if [ "$status" = "PASSED" ]; then
        echo -e "${GREEN}✅ PASSED${NC} - $test_name"
        PASSED_COUNT=$((PASSED_COUNT + 1))
    else
        echo -e "${RED}❌ FAILED${NC} - $test_name"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
    
    if [ -n "$details" ]; then
        echo -e "   ${BLUE}Details:${NC} $details"
    fi
    echo
}

print_header() {
    echo
    echo "================================================================================"
    echo "🔒 VIP HOSTING PANEL v2 - STATIC SECURITY ANALYSIS"
    echo "================================================================================"
    echo "Analysis Type: Code Security Review"
    echo "Started: $(date)"
    echo "================================================================================"
    echo
}

print_summary() {
    echo
    echo "================================================================================"
    echo "📊 SECURITY ANALYSIS RESULTS"
    echo "================================================================================"
    echo "Total Tests: $TEST_COUNT"
    echo "Passed: $PASSED_COUNT"
    echo "Failed: $FAILED_COUNT"
    
    if [ $TEST_COUNT -gt 0 ]; then
        SUCCESS_RATE=$(( (PASSED_COUNT * 100) / TEST_COUNT ))
        echo "Success Rate: ${SUCCESS_RATE}%"
        
        if [ $FAILED_COUNT -eq 0 ]; then
            echo -e "${GREEN}🎉 All security checks passed! Your code is well-secured.${NC}"
        elif [ $SUCCESS_RATE -ge 75 ]; then
            echo -e "${YELLOW}⚠️  Most security checks passed, but some issues need attention.${NC}"
        else
            echo -e "${RED}🚨 Multiple security issues detected. Please address them immediately.${NC}"
        fi
    fi
    echo "================================================================================"
}

# Test 1: Check for security middleware implementation
test_security_middleware() {
    echo "Testing security middleware implementation..."
    
    issues_found=0
    
    # Check for rate limiting middleware
    if ! find . -name "*.go" -type f -exec grep -l "RateLimit\|rate.*limit" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for security headers middleware
    if ! find . -name "*.go" -type f -exec grep -l "SecurityHeaders\|X-Frame-Options\|X-XSS-Protection" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for input validation middleware
    if ! find . -name "*.go" -type f -exec grep -l "ValidateInput\|validation" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    if [ $issues_found -eq 0 ]; then
        print_result "PASSED" "Security Middleware" "All security middleware components found"
    else
        print_result "FAILED" "Security Middleware" "$issues_found security middleware components missing"
    fi
}

# Test 2: SQL Injection Protection Analysis
test_sql_injection_protection() {
    echo "Testing SQL injection protection..."
    
    # Look for parameterized queries and prepared statements
    parameterized_queries=$(find . -name "*.go" -type f -exec grep -l "\$[0-9]\+\|Query.*\$\|Exec.*\$\|QueryRow.*\$" {} \; | wc -l)
    
    # Look for dangerous direct string concatenation in SQL
    dangerous_sql=$(find . -name "*.go" -type f -exec grep -l "\".*SELECT.*\+\|\".*INSERT.*\+\|\".*UPDATE.*\+\|\".*DELETE.*\+" {} \; | wc -l)
    
    if [ $parameterized_queries -gt 0 ] && [ $dangerous_sql -eq 0 ]; then
        print_result "PASSED" "SQL Injection Protection" "Parameterized queries found, no dangerous SQL concatenation"
    else
        print_result "FAILED" "SQL Injection Protection" "Found $dangerous_sql potentially dangerous SQL patterns"
    fi
}

# Test 3: Password Security Analysis
test_password_security() {
    echo "Testing password security implementation..."
    
    issues_found=0
    
    # Check for password hashing
    if ! find . -name "*.go" -type f -exec grep -l "bcrypt\|scrypt\|argon2\|HashPassword" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for password validation
    if ! find . -name "*.go" -type f -exec grep -l "password.*validation\|validatePassword\|password.*length" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check that passwords aren't logged
    if find . -name "*.go" -type f -exec grep -l "log.*password\|Log.*password" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    if [ $issues_found -eq 0 ]; then
        print_result "PASSED" "Password Security" "Proper password hashing and validation found"
    else
        print_result "FAILED" "Password Security" "$issues_found password security issues found"
    fi
}

# Test 4: JWT Token Security Analysis
test_jwt_security() {
    echo "Testing JWT token security..."
    
    issues_found=0
    
    # Check for JWT implementation
    if ! find . -name "*.go" -type f -exec grep -l "jwt\|JWT" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for token validation
    if ! find . -name "*.go" -type f -exec grep -l "ValidateToken\|ParseToken\|VerifyToken" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check that JWT secrets aren't hardcoded
    if find . -name "*.go" -type f -exec grep -l "jwt.*secret.*=.*\".*\"" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    if [ $issues_found -eq 0 ]; then
        print_result "PASSED" "JWT Security" "Proper JWT implementation found"
    else
        print_result "FAILED" "JWT Security" "$issues_found JWT security issues found"
    fi
}

# Test 5: Input Validation Analysis
test_input_validation() {
    echo "Testing input validation implementation..."
    
    # Check for validation libraries
    if find . -name "*.go" -type f -exec grep -l "validator\|validation\|ValidateInput" {} \; | head -1 > /dev/null; then
        # Check for email validation
        if find . -name "*.go" -type f -exec grep -l "email.*validation\|ValidateEmail" {} \; | head -1 > /dev/null; then
            print_result "PASSED" "Input Validation" "Comprehensive input validation implementation found"
        else
            print_result "FAILED" "Input Validation" "Basic validation found but missing email validation"
        fi
    else
        print_result "FAILED" "Input Validation" "No input validation implementation found"
    fi
}

# Test 6: Error Handling Security
test_error_handling() {
    echo "Testing secure error handling..."
    
    issues_found=0
    
    # Check for generic error responses
    if ! find . -name "*.go" -type f -exec grep -l "Internal.*Server.*Error\|internal.*error" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check that database errors aren't exposed
    if find . -name "*.go" -type f -exec grep -l "return.*err\|json.*err" {} \; | head -1 > /dev/null; then
        if ! find . -name "*.go" -type f -exec grep -l "log.*err\|Error.*err" {} \; | head -1 > /dev/null; then
            issues_found=$((issues_found + 1))
        fi
    fi
    
    if [ $issues_found -eq 0 ]; then
        print_result "PASSED" "Error Handling" "Secure error handling patterns found"
    else
        print_result "FAILED" "Error Handling" "$issues_found error handling security issues"
    fi
}

# Test 7: Configuration Security
test_configuration_security() {
    echo "Testing configuration security..."
    
    issues_found=0
    
    # Check for environment variable usage
    if ! find . -name "*.go" -type f -exec grep -l "os.Getenv\|viper\|config" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for hardcoded secrets
    if find . -name "*.go" -type f -exec grep -l "password.*=.*\"\|secret.*=.*\"\|key.*=.*\"" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for example config files
    if [ -f "configs/config.yaml.example" ] || [ -f "config.yaml.example" ]; then
        # Good - example config exists
        true
    else
        issues_found=$((issues_found + 1))
    fi
    
    if [ $issues_found -eq 0 ]; then
        print_result "PASSED" "Configuration Security" "Secure configuration practices found"
    else
        print_result "FAILED" "Configuration Security" "$issues_found configuration security issues"
    fi
}

# Test 8: CORS Security Analysis
test_cors_security() {
    echo "Testing CORS security configuration..."
    
    # Check for CORS implementation
    if find . -name "*.go" -type f -exec grep -l "cors\|CORS" {} \; | head -1 > /dev/null; then
        # Check for wildcard origins (security risk)
        if find . -name "*.go" -type f -exec grep -l "AllowOrigins.*\*\|origin.*\*" {} \; | head -1 > /dev/null; then
            print_result "FAILED" "CORS Security" "Wildcard (*) origins detected - security risk"
        else
            print_result "PASSED" "CORS Security" "Secure CORS configuration found"
        fi
    else
        print_result "FAILED" "CORS Security" "No CORS implementation found"
    fi
}

# Test 9: Logging Security
test_logging_security() {
    echo "Testing logging security..."
    
    issues_found=0
    
    # Check for logging implementation
    if ! find . -name "*.go" -type f -exec grep -l "log\|Log" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check that sensitive data isn't logged
    if find . -name "*.go" -type f -exec grep -l "log.*password\|log.*token\|log.*secret" {} \; | head -1 > /dev/null; then
        issues_found=$((issues_found + 1))
    fi
    
    # Check for security event logging
    if find . -name "*.go" -type f -exec grep -l "log.*login\|log.*auth\|log.*security" {} \; | head -1 > /dev/null; then
        # Good - security events are logged
        true
    else
        issues_found=$((issues_found + 1))
    fi
    
    if [ $issues_found -eq 0 ]; then
        print_result "PASSED" "Logging Security" "Secure logging practices found"
    else
        print_result "FAILED" "Logging Security" "$issues_found logging security issues"
    fi
}

# Test 10: Dependencies Security
test_dependencies_security() {
    echo "Testing dependencies security..."
    
    if [ -f "go.mod" ]; then
        # Check for known secure libraries
        secure_libs=0
        
        if grep -q "golang.org/x/crypto" go.mod; then
            secure_libs=$((secure_libs + 1))
        fi
        
        if grep -q "github.com/golang-jwt/jwt" go.mod; then
            secure_libs=$((secure_libs + 1))
        fi
        
        if grep -q "github.com/go-playground/validator" go.mod; then
            secure_libs=$((secure_libs + 1))
        fi
        
        if [ $secure_libs -ge 2 ]; then
            print_result "PASSED" "Dependencies Security" "Secure libraries detected in dependencies"
        else
            print_result "FAILED" "Dependencies Security" "Missing recommended security libraries"
        fi
    else
        print_result "FAILED" "Dependencies Security" "No go.mod file found"
    fi
}

# Test 11: File Structure Security
test_file_structure() {
    echo "Testing secure file structure..."
    
    security_score=0
    
    # Check for proper directory structure
    if [ -d "internal" ]; then
        security_score=$((security_score + 1))
    fi
    
    if [ -d "internal/middleware" ]; then
        security_score=$((security_score + 1))
    fi
    
    if [ -d "configs" ]; then
        security_score=$((security_score + 1))
    fi
    
    # Check for security documentation
    if [ -f "SECURITY.md" ] || [ -f "SECURITY_PERFORMANCE_REPORT.md" ]; then
        security_score=$((security_score + 1))
    fi
    
    # Check for .gitignore (prevents committing secrets)
    if [ -f ".gitignore" ]; then
        if grep -q "config.yaml\|\.env\|secrets" .gitignore; then
            security_score=$((security_score + 1))
        fi
    fi
    
    if [ $security_score -ge 4 ]; then
        print_result "PASSED" "File Structure Security" "Secure project structure found ($security_score/5 criteria met)"
    else
        print_result "FAILED" "File Structure Security" "Insecure project structure ($security_score/5 criteria met)"
    fi
}

# Test 12: Security Documentation
test_security_documentation() {
    echo "Testing security documentation..."
    
    docs_found=0
    
    if [ -f "SECURITY.md" ]; then
        docs_found=$((docs_found + 1))
    fi
    
    if [ -f "SECURITY_PERFORMANCE_REPORT.md" ]; then
        docs_found=$((docs_found + 1))
    fi
    
    if [ -f "README.md" ]; then
        if grep -q -i "security\|authentication\|authorization" README.md; then
            docs_found=$((docs_found + 1))
        fi
    fi
    
    if [ $docs_found -ge 2 ]; then
        print_result "PASSED" "Security Documentation" "Comprehensive security documentation found"
    else
        print_result "FAILED" "Security Documentation" "Insufficient security documentation"
    fi
}

# Main execution
main() {
    print_header
    
    # Run all static security tests
    test_security_middleware
    test_sql_injection_protection
    test_password_security
    test_jwt_security
    test_input_validation
    test_error_handling
    test_configuration_security
    test_cors_security
    test_logging_security
    test_dependencies_security
    test_file_structure
    test_security_documentation
    
    print_summary
    
    echo
    echo "🔍 SECURITY RECOMMENDATIONS:"
    echo "1. Run \`go mod audit\` to check for vulnerable dependencies"
    echo "2. Use \`gosec\` for additional static analysis: go install github.com/securecodewarrior/gosec/v2/cmd/gosec"
    echo "3. Consider using \`govulncheck\` for vulnerability scanning"
    echo "4. Implement automated security testing in CI/CD pipeline"
    echo "5. Regular security audits and penetration testing"
    echo
}

# Run the tests
main