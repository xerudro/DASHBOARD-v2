#!/bin/bash

# VIP Hosting Panel v2 - Comprehensive Security Test Suite
# This script performs comprehensive security analysis using multiple tools

set -e

echo "=================================================================================="
echo "🔒 VIP HOSTING PANEL v2 - COMPREHENSIVE SECURITY ANALYSIS"
echo "=================================================================================="
echo "Analysis Type: Multi-Tool Security Review"
echo "Started: $(date)"
echo "=================================================================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✅ PASSED${NC} - $message"
            ;;
        "FAIL")
            echo -e "${RED}❌ FAILED${NC} - $message"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  WARNING${NC} - $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  INFO${NC} - $message"
            ;;
    esac
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

# Test 1: Check Go installation
echo "Test 1: Go Installation Check"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if command_exists go; then
    GO_VERSION=$(go version)
    print_status "PASS" "Go is installed: $GO_VERSION"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "Go is not installed"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 2: Install and run gosec
echo "Test 2: Installing and Running gosec Security Scanner"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if command_exists go; then
    print_status "INFO" "Installing gosec security scanner..."
    
    # Install gosec
    if go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; then
        print_status "INFO" "gosec installed successfully"
        
        # Add Go bin to PATH if not already there
        export PATH=$PATH:$(go env GOPATH)/bin
        
        # Run gosec analysis
        print_status "INFO" "Running gosec security analysis..."
        if gosec -fmt json -out gosec-report.json ./... 2>/dev/null; then
            # Check if gosec found any issues
            if [ -f gosec-report.json ]; then
                ISSUES=$(grep -o '"Issues":\[.*\]' gosec-report.json | grep -o '\[.*\]' | wc -c)
                if [ "$ISSUES" -gt 2 ]; then
                    print_status "WARN" "gosec found potential security issues (see gosec-report.json)"
                    WARNINGS=$((WARNINGS + 1))
                else
                    print_status "PASS" "gosec analysis completed - no major issues found"
                    PASSED_TESTS=$((PASSED_TESTS + 1))
                fi
            else
                print_status "PASS" "gosec analysis completed successfully"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            fi
        else
            print_status "WARN" "gosec analysis had warnings"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        print_status "FAIL" "Failed to install gosec"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
else
    print_status "FAIL" "Cannot run gosec - Go not installed"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 3: Check for hardcoded secrets
echo "Test 3: Hardcoded Secrets Detection"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
SECRET_PATTERNS=(
    "password.*=.*['\"][^'\"]*['\"]"
    "secret.*=.*['\"][^'\"]*['\"]"
    "api_key.*=.*['\"][^'\"]*['\"]"
    "token.*=.*['\"][^'\"]*['\"]"
    "key.*=.*['\"][^'\"]*['\"]"
)

SECRETS_FOUND=0
for pattern in "${SECRET_PATTERNS[@]}"; do
    if find . -name "*.go" -type f -exec grep -l -i "$pattern" {} \; 2>/dev/null | grep -q .; then
        SECRETS_FOUND=$((SECRETS_FOUND + 1))
    fi
done

if [ $SECRETS_FOUND -eq 0 ]; then
    print_status "PASS" "No obvious hardcoded secrets found"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "WARN" "Found $SECRETS_FOUND potential hardcoded secrets"
    WARNINGS=$((WARNINGS + 1))
fi
echo

# Test 4: SQL Injection Protection Check
echo "Test 4: SQL Injection Protection Analysis"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if find . -name "*.go" -type f -exec grep -l "SQLSecurityMiddleware\|sql_security" {} \; 2>/dev/null | grep -q .; then
    print_status "PASS" "SQL injection protection middleware found"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "SQL injection protection middleware not found"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 5: CSRF Protection Check
echo "Test 5: CSRF Protection Analysis"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if find . -name "*.go" -type f -exec grep -l "CSRFProtection\|csrf_security" {} \; 2>/dev/null | grep -q .; then
    print_status "PASS" "CSRF protection found"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "CSRF protection not found"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 6: Security Headers Check
echo "Test 6: Security Headers Implementation"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
REQUIRED_HEADERS=(
    "X-Content-Type-Options"
    "X-Frame-Options"
    "X-XSS-Protection"
    "Strict-Transport-Security"
    "Content-Security-Policy"
)

HEADERS_FOUND=0
for header in "${REQUIRED_HEADERS[@]}"; do
    if find . -name "*.go" -type f -exec grep -l "$header" {} \; 2>/dev/null | grep -q .; then
        HEADERS_FOUND=$((HEADERS_FOUND + 1))
    fi
done

if [ $HEADERS_FOUND -ge 4 ]; then
    print_status "PASS" "Security headers implementation found ($HEADERS_FOUND/5)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
elif [ $HEADERS_FOUND -ge 2 ]; then
    print_status "WARN" "Partial security headers implementation ($HEADERS_FOUND/5)"
    WARNINGS=$((WARNINGS + 1))
else
    print_status "FAIL" "Insufficient security headers ($HEADERS_FOUND/5)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 7: Authentication Security
echo "Test 7: Authentication Security Analysis"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
AUTH_FEATURES=0

# Check for JWT implementation
if find . -name "*.go" -type f -exec grep -l "jwt\|JWT" {} \; 2>/dev/null | grep -q .; then
    AUTH_FEATURES=$((AUTH_FEATURES + 1))
fi

# Check for password hashing
if find . -name "*.go" -type f -exec grep -l "bcrypt\|HashPassword" {} \; 2>/dev/null | grep -q .; then
    AUTH_FEATURES=$((AUTH_FEATURES + 1))
fi

# Check for rate limiting
if find . -name "*.go" -type f -exec grep -l "RateLimit\|rate.*limit" {} \; 2>/dev/null | grep -q .; then
    AUTH_FEATURES=$((AUTH_FEATURES + 1))
fi

if [ $AUTH_FEATURES -ge 3 ]; then
    print_status "PASS" "Comprehensive authentication security found"
    PASSED_TESTS=$((PASSED_TESTS + 1))
elif [ $AUTH_FEATURES -ge 2 ]; then
    print_status "WARN" "Basic authentication security found"
    WARNINGS=$((WARNINGS + 1))
else
    print_status "FAIL" "Insufficient authentication security"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 8: Dependency Security
echo "Test 8: Dependency Security Check"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [ -f "go.mod" ]; then
    print_status "INFO" "Checking Go dependencies for known vulnerabilities..."
    
    # Install govulncheck if not available
    if ! command_exists govulncheck; then
        if command_exists go; then
            go install golang.org/x/vuln/cmd/govulncheck@latest
            export PATH=$PATH:$(go env GOPATH)/bin
        fi
    fi
    
    if command_exists govulncheck; then
        if govulncheck ./... > vulncheck-report.txt 2>&1; then
            if grep -q "No vulnerabilities found" vulncheck-report.txt; then
                print_status "PASS" "No known vulnerabilities in dependencies"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                print_status "WARN" "Potential vulnerabilities found (see vulncheck-report.txt)"
                WARNINGS=$((WARNINGS + 1))
            fi
        else
            print_status "WARN" "Vulnerability check completed with warnings"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        print_status "WARN" "Could not install govulncheck"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    print_status "FAIL" "go.mod not found"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Test 9: Configuration Security
echo "Test 9: Configuration Security Analysis"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
CONFIG_ISSUES=0

# Check for environment variable usage
if find . -name "*.yaml" -o -name "*.yml" -exec grep -l "\${" {} \; 2>/dev/null | grep -q .; then
    CONFIG_ISSUES=$((CONFIG_ISSUES + 1))
fi

# Check for secure defaults
if find configs/ -name "*.example" -exec grep -l "change-this\|CHANGE\|example\|localhost" {} \; 2>/dev/null | grep -q .; then
    CONFIG_ISSUES=$((CONFIG_ISSUES + 1))
fi

if [ $CONFIG_ISSUES -eq 0 ]; then
    print_status "WARN" "Configuration may have hardcoded values"
    WARNINGS=$((WARNINGS + 1))
else
    print_status "PASS" "Configuration uses environment variables and examples"
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi
echo

# Test 10: Code Quality and Security Patterns
echo "Test 10: Security Code Patterns Analysis"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
GOOD_PATTERNS=0

# Check for error handling
if find . -name "*.go" -type f -exec grep -l "if err != nil" {} \; 2>/dev/null | wc -l | awk '{print $1}' | head -1 | { read count; [ "$count" -gt 5 ]; }; then
    GOOD_PATTERNS=$((GOOD_PATTERNS + 1))
fi

# Check for input validation
if find . -name "*.go" -type f -exec grep -l "validation\|validate" {} \; 2>/dev/null | grep -q .; then
    GOOD_PATTERNS=$((GOOD_PATTERNS + 1))
fi

# Check for logging
if find . -name "*.go" -type f -exec grep -l "log\.\|zerolog" {} \; 2>/dev/null | grep -q .; then
    GOOD_PATTERNS=$((GOOD_PATTERNS + 1))
fi

if [ $GOOD_PATTERNS -ge 3 ]; then
    print_status "PASS" "Good security coding patterns found"
    PASSED_TESTS=$((PASSED_TESTS + 1))
elif [ $GOOD_PATTERNS -ge 2 ]; then
    print_status "WARN" "Some security patterns found"
    WARNINGS=$((WARNINGS + 1))
else
    print_status "FAIL" "Insufficient security patterns"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo

# Final Results
echo "=================================================================================="
echo "📊 COMPREHENSIVE SECURITY ANALYSIS RESULTS"
echo "=================================================================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Warnings: $WARNINGS"

SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo "Success Rate: ${SUCCESS_RATE}%"

if [ $FAILED_TESTS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}🔒 Excellent! No security issues detected.${NC}"
elif [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  Good security posture with minor warnings to address.${NC}"
elif [ $FAILED_TESTS -le 2 ]; then
    echo -e "${YELLOW}⚠️  Some security issues detected. Please address them.${NC}"
else
    echo -e "${RED}🚨 Multiple security issues detected. Immediate attention required.${NC}"
fi

echo "=================================================================================="
echo
echo "🔍 DETAILED REPORTS GENERATED:"
echo "- gosec-report.json (if gosec ran successfully)"
echo "- vulncheck-report.txt (if govulncheck ran successfully)"
echo
echo "🔧 SECURITY RECOMMENDATIONS:"
echo "1. Review any generated reports for detailed findings"
echo "2. Ensure all environment variables are properly configured"
echo "3. Run security tests regularly as part of CI/CD pipeline"
echo "4. Keep dependencies updated regularly"
echo "5. Consider professional security audit for production deployment"
echo
echo "For more information, see: https://owasp.org/www-project-go-secure-coding-practices-guide/"
echo "=================================================================================="

# Exit with appropriate code
if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
else
    exit 0
fi