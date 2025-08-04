#!/bin/bash

# API Test Script for Auto-Devs
# This script runs the complete API test suite using Newman

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8098}"
COLLECTION_FILE="tests/postman/auto-devs-collection.json"
ENVIRONMENT_FILE="tests/postman/environment.json"
REPORTS_DIR="tests/reports"
TIMEOUT="${TIMEOUT:-30000}"

# Ensure reports directory exists
mkdir -p "${REPORTS_DIR}"

echo -e "${BLUE}Auto-Devs API Test Suite${NC}"
echo -e "${BLUE}========================${NC}"
echo ""

# Check if Newman is installed
if ! command -v newman &> /dev/null; then
    echo -e "${RED}Newman is not installed. Please install it with:${NC}"
    echo "npm install -g newman"
    echo "npm install -g newman-reporter-htmlextra"
    exit 1
fi

# Check if server is running
echo -e "${YELLOW}Checking if API server is running...${NC}"
if ! curl -s "${API_BASE_URL}/api/v1/health" > /dev/null 2>&1; then
    echo -e "${RED}API server is not running at ${API_BASE_URL}${NC}"
    echo "Please start the server first with: make run"
    exit 1
fi

echo -e "${GREEN}API server is running at ${API_BASE_URL}${NC}"
echo ""

# Update environment file with current base URL
echo -e "${YELLOW}Updating environment configuration...${NC}"
jq --arg url "${API_BASE_URL}" '.values[0].value = $url' "${ENVIRONMENT_FILE}" > "${ENVIRONMENT_FILE}.tmp"
mv "${ENVIRONMENT_FILE}.tmp" "${ENVIRONMENT_FILE}"

# Run the test suite
echo -e "${YELLOW}Running API test suite...${NC}"
echo ""

# Define output files
JUNIT_REPORT="${REPORTS_DIR}/api-tests-junit.xml"
HTML_REPORT="${REPORTS_DIR}/api-tests-report.html"
JSON_REPORT="${REPORTS_DIR}/api-tests-results.json"

# Run Newman with multiple reporters
newman run "${COLLECTION_FILE}" \
    --environment "${ENVIRONMENT_FILE}" \
    --timeout "${TIMEOUT}" \
    --reporters cli,junit,htmlextra,json \
    --reporter-junit-export "${JUNIT_REPORT}" \
    --reporter-htmlextra-export "${HTML_REPORT}" \
    --reporter-htmlextra-title "Auto-Devs API Test Report" \
    --reporter-htmlextra-darkTheme \
    --reporter-json-export "${JSON_REPORT}" \
    --color on \
    --delay-request 100

# Capture exit code
NEWMAN_EXIT_CODE=$?

echo ""
echo -e "${BLUE}Test Results:${NC}"
echo -e "${BLUE}=============${NC}"

if [ $NEWMAN_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
else
    echo -e "${RED}❌ Some tests failed (exit code: $NEWMAN_EXIT_CODE)${NC}"
fi

# Display report locations
echo ""
echo -e "${YELLOW}Reports generated:${NC}"
echo "📊 HTML Report: ${HTML_REPORT}"
echo "📋 JUnit Report: ${JUNIT_REPORT}"
echo "📄 JSON Report: ${JSON_REPORT}"

# Parse and display summary from JSON report
if [ -f "${JSON_REPORT}" ]; then
    echo ""
    echo -e "${YELLOW}Test Summary:${NC}"
    
    # Extract summary using jq
    if command -v jq &> /dev/null; then
        TOTAL_TESTS=$(jq -r '.run.stats.tests.total' "${JSON_REPORT}")
        PASSED_TESTS=$(jq -r '.run.stats.tests.passed' "${JSON_REPORT}")
        FAILED_TESTS=$(jq -r '.run.stats.tests.failed' "${JSON_REPORT}")
        TOTAL_ASSERTIONS=$(jq -r '.run.stats.assertions.total' "${JSON_REPORT}")
        PASSED_ASSERTIONS=$(jq -r '.run.stats.assertions.passed' "${JSON_REPORT}")
        FAILED_ASSERTIONS=$(jq -r '.run.stats.assertions.failed' "${JSON_REPORT}")
        
        echo "  Tests: ${PASSED_TESTS}/${TOTAL_TESTS} passed"
        echo "  Assertions: ${PASSED_ASSERTIONS}/${TOTAL_ASSERTIONS} passed"
        
        if [ "${FAILED_TESTS}" != "0" ]; then
            echo -e "${RED}  Failed Tests: ${FAILED_TESTS}${NC}"
        fi
        
        if [ "${FAILED_ASSERTIONS}" != "0" ]; then
            echo -e "${RED}  Failed Assertions: ${FAILED_ASSERTIONS}${NC}"
        fi
    fi
fi

echo ""

# Performance test option
if [ "$1" = "--performance" ]; then
    echo -e "${YELLOW}Running performance tests...${NC}"
    
    # Run with multiple iterations for performance testing
    newman run "${COLLECTION_FILE}" \
        --environment "${ENVIRONMENT_FILE}" \
        --iteration-count 10 \
        --delay-request 50 \
        --timeout 10000 \
        --reporters cli,json \
        --reporter-json-export "${REPORTS_DIR}/performance-results.json"
        
    echo -e "${GREEN}Performance test completed. Results saved to ${REPORTS_DIR}/performance-results.json${NC}"
fi

# Load test option
if [ "$1" = "--load" ]; then
    echo -e "${YELLOW}Running load tests...${NC}"
    
    # Run with high concurrency
    newman run "${COLLECTION_FILE}" \
        --environment "${ENVIRONMENT_FILE}" \
        --iteration-count 50 \
        --delay-request 10 \
        --timeout 5000 \
        --reporters cli,json \
        --reporter-json-export "${REPORTS_DIR}/load-test-results.json"
        
    echo -e "${GREEN}Load test completed. Results saved to ${REPORTS_DIR}/load-test-results.json${NC}"
fi

# Open HTML report if requested
if [ "$2" = "--open" ] || [ "$1" = "--open" ]; then
    if command -v open &> /dev/null; then
        open "${HTML_REPORT}"
    elif command -v xdg-open &> /dev/null; then
        xdg-open "${HTML_REPORT}"
    else
        echo -e "${YELLOW}Cannot open HTML report automatically. Please open: ${HTML_REPORT}${NC}"
    fi
fi

exit $NEWMAN_EXIT_CODE