#!/bin/bash

# VIP Hosting Panel v2 - Security Test Suite
# Comprehensive security testing without Go dependency

BASE_URL="http://localhost:8080"
TEST_COUNT=0
PASSED_COUNT=0
FAILED_COUNT=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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
    echo "🔒 VIP HOSTING PANEL v2 - SECURITY TEST SUITE"
    echo "================================================================================"
    echo "Target: $BASE_URL"
    echo "Started: $(date)"
    echo "================================================================================"
    echo
}

print_summary() {
    echo
    echo "================================================================================"
    echo "📊 SECURITY TEST RESULTS SUMMARY"
    echo "================================================================================"
    echo "Total Tests: $TEST_COUNT"
    echo "Passed: $PASSED_COUNT"
    echo "Failed: $FAILED_COUNT"
    
    if [ $TEST_COUNT -gt 0 ]; then
        SUCCESS_RATE=$(( (PASSED_COUNT * 100) / TEST_COUNT ))
        echo "Success Rate: ${SUCCESS_RATE}%"
        
        if [ $FAILED_COUNT -eq 0 ]; then
            echo -e "${GREEN}🎉 All security tests passed! Your application is well-secured.${NC}"
        elif [ $SUCCESS_RATE -ge 75 ]; then
            echo -e "${YELLOW}⚠️  Most security tests passed, but some issues need attention.${NC}"
        else
            echo -e "${RED}🚨 Multiple security issues detected. Please address them immediately.${NC}"
        fi
    fi
    echo "================================================================================"
}

# Test 1: Check if server is running
test_server_availability() {
    echo "Testing server availability..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" 2>/dev/null)
    
    if [ "$response" = "200" ] || [ "$response" = "404" ]; then
        print_result "PASSED" "Server Availability" "Server is responding (HTTP $response)"
        return 0
    else
        print_result "FAILED" "Server Availability" "Server not responding or unreachable"
        return 1
    fi
}

# Test 2: Security Headers
test_security_headers() {
    echo "Testing security headers..."
    response=$(curl -s -I "$BASE_URL/" 2>/dev/null)
    
    # Check for security headers
    missing_headers=""
    
    if ! echo "$response" | grep -q "X-Content-Type-Options"; then
        missing_headers="$missing_headers X-Content-Type-Options"
    fi
    
    if ! echo "$response" | grep -q "X-Frame-Options"; then
        missing_headers="$missing_headers X-Frame-Options"
    fi
    
    if ! echo "$response" | grep -q "X-XSS-Protection"; then
        missing_headers="$missing_headers X-XSS-Protection"
    fi
    
    if [ -z "$missing_headers" ]; then
        print_result "PASSED" "Security Headers" "All critical security headers present"
    else
        print_result "FAILED" "Security Headers" "Missing headers:$missing_headers"
    fi
}

# Test 3: CORS Configuration
test_cors_configuration() {
    echo "Testing CORS configuration..."
    response=$(curl -s -H "Origin: https://malicious-site.com" \
                   -H "Access-Control-Request-Method: POST" \
                   -X OPTIONS "$BASE_URL/api/v1/auth/login" \
                   -i 2>/dev/null)
    
    if echo "$response" | grep -q "Access-Control-Allow-Origin: \*"; then
        print_result "FAILED" "CORS Configuration" "Wildcard (*) origin allowed - security risk"
    elif echo "$response" | grep -q "Access-Control-Allow-Origin: https://malicious-site.com"; then
        print_result "FAILED" "CORS Configuration" "Malicious origin allowed"
    else
        print_result "PASSED" "CORS Configuration" "CORS properly restricts origins"
    fi
}

# Test 4: Rate Limiting
test_rate_limiting() {
    echo "Testing rate limiting..."
    rate_limited=false
    
    # Send rapid requests
    for i in {1..15}; do
        response=$(curl -s -o /dev/null -w "%{http_code}" \
                      -X POST "$BASE_URL/api/v1/auth/login" \
                      -H "Content-Type: application/json" \
                      -d '{"email":"test@example.com","password":"test123"}' 2>/dev/null)
        
        if [ "$response" = "429" ]; then
            rate_limited=true
            break
        fi
        sleep 0.1
    done
    
    if [ "$rate_limited" = true ]; then
        print_result "PASSED" "Rate Limiting" "Rate limiting is active (HTTP 429 received)"
    else
        print_result "FAILED" "Rate Limiting" "No rate limiting detected"
    fi
}

# Test 5: Input Validation
test_input_validation() {
    echo "Testing input validation..."
    
    # Test with invalid email
    response=$(curl -s -o /dev/null -w "%{http_code}" \
                  -X POST "$BASE_URL/api/v1/auth/login" \
                  -H "Content-Type: application/json" \
                  -d '{"email":"invalid-email","password":"test123"}' 2>/dev/null)
    
    if [ "$response" = "400" ]; then
        print_result "PASSED" "Input Validation" "Invalid input properly rejected (HTTP 400)"
    else
        print_result "FAILED" "Input Validation" "Invalid input not rejected (HTTP $response)"
    fi
}

# Test 6: SQL Injection Protection
test_sql_injection() {
    echo "Testing SQL injection protection..."
    
    # Test with SQL injection payload
    response=$(curl -s -w "%{http_code}" \
                  -X POST "$BASE_URL/api/v1/auth/login" \
                  -H "Content-Type: application/json" \
                  -d '{"email":"admin@test.com'\'' OR '\''1'\''='\''1","password":"test123"}' 2>/dev/null)
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    # Should return 400 (bad request) or 401 (unauthorized), not 500 (internal server error)
    if [ "$http_code" = "400" ] || [ "$http_code" = "401" ]; then
        print_result "PASSED" "SQL Injection Protection" "SQL injection payload properly handled"
    elif [ "$http_code" = "500" ]; then
        print_result "FAILED" "SQL Injection Protection" "Potential SQL injection vulnerability (HTTP 500)"
    else
        print_result "FAILED" "SQL Injection Protection" "Unexpected response (HTTP $http_code)"
    fi
}

# Test 7: XSS Protection
test_xss_protection() {
    echo "Testing XSS protection..."
    
    # Test with XSS payload
    xss_payload='<script>alert("xss")</script>'
    response=$(curl -s \
                  -X POST "$BASE_URL/api/v1/auth/login" \
                  -H "Content-Type: application/json" \
                  -d "{\"email\":\"test@test.com\",\"password\":\"$xss_payload\"}" 2>/dev/null)
    
    if echo "$response" | grep -q "$xss_payload"; then
        print_result "FAILED" "XSS Protection" "XSS payload reflected in response"
    else
        print_result "PASSED" "XSS Protection" "XSS payload not reflected in response"
    fi
}

# Test 8: Authentication Endpoint Security
test_auth_security() {
    echo "Testing authentication endpoint security..."
    
    # Test login with empty credentials
    response=$(curl -s -o /dev/null -w "%{http_code}" \
                  -X POST "$BASE_URL/api/v1/auth/login" \
                  -H "Content-Type: application/json" \
                  -d '{"email":"","password":""}' 2>/dev/null)
    
    if [ "$response" = "400" ] || [ "$response" = "401" ]; then
        print_result "PASSED" "Authentication Security" "Empty credentials properly rejected"
    else
        print_result "FAILED" "Authentication Security" "Empty credentials not properly handled"
    fi
}

# Test 9: Unauthorized Access
test_unauthorized_access() {
    echo "Testing unauthorized access protection..."
    
    # Test protected endpoint without authentication
    response=$(curl -s -o /dev/null -w "%{http_code}" \
                  "$BASE_URL/api/v1/dashboard" 2>/dev/null)
    
    if [ "$response" = "401" ] || [ "$response" = "403" ]; then
        print_result "PASSED" "Unauthorized Access" "Protected endpoints require authentication"
    else
        print_result "FAILED" "Unauthorized Access" "Protected endpoint accessible without auth (HTTP $response)"
    fi
}

# Test 10: Information Disclosure
test_information_disclosure() {
    echo "Testing information disclosure..."
    
    # Test non-existent endpoint for error information
    response=$(curl -s "$BASE_URL/api/v1/nonexistent" 2>/dev/null)
    
    # Check for sensitive information in error responses
    if echo "$response" | grep -qi "database\|sql\|postgres\|redis\|stack trace\|panic"; then
        print_result "FAILED" "Information Disclosure" "Sensitive information in error responses"
    else
        print_result "PASSED" "Information Disclosure" "No sensitive information disclosed in errors"
    fi
}

# Test 11: HTTPS Enforcement (if available)
test_https_enforcement() {
    echo "Testing HTTPS enforcement..."
    
    # Check if HTTPS redirect is in place
    response=$(curl -s -I "http://localhost:8080/" 2>/dev/null)
    
    if echo "$response" | grep -q "301\|302"; then
        if echo "$response" | grep -q "https://"; then
            print_result "PASSED" "HTTPS Enforcement" "HTTP redirects to HTTPS"
        else
            print_result "FAILED" "HTTPS Enforcement" "Redirect present but not to HTTPS"
        fi
    else
        print_result "FAILED" "HTTPS Enforcement" "No HTTPS redirect detected"
    fi
}

# Test 12: Content Type Validation
test_content_type() {
    echo "Testing content type validation..."
    
    # Send request with wrong content type
    response=$(curl -s -o /dev/null -w "%{http_code}" \
                  -X POST "$BASE_URL/api/v1/auth/login" \
                  -H "Content-Type: text/plain" \
                  -d '{"email":"test@test.com","password":"test123"}' 2>/dev/null)
    
    if [ "$response" = "400" ] || [ "$response" = "415" ]; then
        print_result "PASSED" "Content Type Validation" "Wrong content type properly rejected"
    else
        print_result "FAILED" "Content Type Validation" "Wrong content type accepted (HTTP $response)"
    fi
}

# Main execution
main() {
    print_header
    
    # Test server availability first
    if ! test_server_availability; then
        echo "❌ Server is not available. Please start the server and try again."
        echo "Expected server at: $BASE_URL"
        echo
        echo "To start the server:"
        echo "  make dev"
        echo "  # or"
        echo "  go run cmd/api/main.go"
        exit 1
    fi
    
    # Run all security tests
    test_security_headers
    test_cors_configuration
    test_rate_limiting
    test_input_validation
    test_sql_injection
    test_xss_protection
    test_auth_security
    test_unauthorized_access
    test_information_disclosure
    test_https_enforcement
    test_content_type
    
    print_summary
}

# Handle command line arguments
if [ $# -gt 0 ]; then
    BASE_URL=$1
fi

# Run the tests
main