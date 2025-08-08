#!/bin/bash

# End-to-End Test Runner Script
# This script sets up the environment and runs E2E tests locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-30m}
TEST_SUITE=${TEST_SUITE:-all}
PARALLEL=${PARALLEL:-1}
VERBOSE=${VERBOSE:-false}
CLEANUP=${CLEANUP:-true}
DB_CONTAINER_NAME="autodevs-test-db"
REDIS_CONTAINER_NAME="autodevs-test-redis"

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -s, --suite SUITE       Test suite to run (all, happy-path, error-scenarios, edge-cases, performance)
    -t, --timeout TIMEOUT   Test timeout (default: 30m)
    -p, --parallel N        Number of parallel test processes (default: 1)
    -v, --verbose           Verbose output
    --no-cleanup            Skip cleanup of test resources
    --setup-only            Only setup test environment, don't run tests
    --cleanup-only          Only cleanup test resources
    
Test Suites:
    all                     Run all test suites
    happy-path              Run happy path scenarios
    error-scenarios         Run error handling tests
    edge-cases              Run edge case tests
    performance             Run performance tests

Examples:
    $0                      # Run all tests
    $0 -s happy-path -v     # Run happy path tests with verbose output
    $0 -s performance -t 45m # Run performance tests with 45m timeout
    $0 --cleanup-only       # Clean up test resources only

Environment Variables:
    TEST_TIMEOUT            Test timeout duration
    TEST_SUITE              Test suite to run
    PARALLEL                Number of parallel processes
    VERBOSE                 Enable verbose output (true/false)
    CLEANUP                 Enable cleanup (true/false)
EOF
}

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if ! command -v make &> /dev/null; then
        missing_tools+=("make")
    fi
    
    if ! command -v migrate &> /dev/null; then
        log_warning "migrate tool not found, attempting to install..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        if ! command -v migrate &> /dev/null; then
            missing_tools+=("migrate")
        fi
    fi
    
    if [[ ${#missing_tools[@]} -ne 0 ]]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_error "Please install the missing tools and try again"
        exit 1
    fi
    
    log_success "All prerequisites are satisfied"
}

# Start required services
start_services() {
    log_info "Starting test services..."
    
    # Check if containers are already running
    if docker ps | grep -q $DB_CONTAINER_NAME; then
        log_info "Database container already running"
    else
        log_info "Starting PostgreSQL database..."
        docker run -d \
            --name $DB_CONTAINER_NAME \
            -e POSTGRES_PASSWORD=postgres \
            -e POSTGRES_DB=autodevs_test \
            -p 5432:5432 \
            --health-cmd="pg_isready -U postgres" \
            --health-interval=10s \
            --health-timeout=5s \
            --health-retries=5 \
            postgres:15
    fi
    
    if docker ps | grep -q $REDIS_CONTAINER_NAME; then
        log_info "Redis container already running"
    else
        log_info "Starting Redis..."
        docker run -d \
            --name $REDIS_CONTAINER_NAME \
            -p 6379:6379 \
            --health-cmd="redis-cli ping" \
            --health-interval=10s \
            --health-timeout=5s \
            --health-retries=5 \
            redis:7
    fi
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        if docker inspect --format='{{.State.Health.Status}}' $DB_CONTAINER_NAME | grep -q "healthy" && \
           docker inspect --format='{{.State.Health.Status}}' $REDIS_CONTAINER_NAME | grep -q "healthy"; then
            log_success "All services are healthy"
            return 0
        fi
        
        log_info "Waiting for services to become healthy... (attempt $((attempt + 1))/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "Services failed to become healthy within timeout"
    return 1
}

# Setup test environment
setup_environment() {
    log_info "Setting up test environment..."
    
    # Create test environment file
    if [[ ! -f .env.test ]]; then
        log_info "Creating .env.test file..."
        cat << EOF > .env.test
# Test Environment Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=autodevs_test

REDIS_HOST=localhost
REDIS_PORT=6379

# Test-specific settings
TEST_TIMEOUT=$TEST_TIMEOUT
LOG_LEVEL=debug
GIN_MODE=test

# Mock settings
MOCK_AI_SERVICES=true
MOCK_GITHUB_API=true
MOCK_GIT_OPERATIONS=true
EOF
    fi
    
    # Install Go dependencies
    log_info "Installing Go dependencies..."
    go mod download
    go mod tidy
    
    # Generate mocks
    log_info "Generating mocks..."
    make mocks
    
    # Run migrations
    log_info "Running database migrations..."
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USERNAME=postgres
    export DB_PASSWORD=postgres
    export DB_NAME=autodevs_test
    make migrate-up
    
    log_success "Test environment setup completed"
}

# Run specific test suite
run_test_suite() {
    local suite=$1
    log_info "Running test suite: $suite"
    
    local test_pattern=""
    local timeout_flag="-timeout $TEST_TIMEOUT"
    local verbose_flag=""
    local parallel_flag=""
    
    if [[ $VERBOSE == "true" ]]; then
        verbose_flag="-v"
    fi
    
    if [[ $PARALLEL -gt 1 ]]; then
        parallel_flag="-parallel $PARALLEL"
    fi
    
    # Set test patterns based on suite
    case $suite in
        happy-path)
            test_pattern="-run TestCompleteTaskAutomationFlow|TestMultiTaskProjectWorkflow|TestPlanGenerationAndApproval|TestImplementationExecution|TestPRCreationAndMonitoring|TestWebSocketRealTimeUpdates"
            ;;
        error-scenarios)
            test_pattern="-run TestPlanningServiceFailure|TestImplementationServiceFailure|TestGitOperationFailures|TestGitHubAPIFailures|TestDatabaseConnectionFailures|TestConcurrencyIssues|TestResourceExhaustion"
            ;;
        edge-cases)
            test_pattern="-run TestEdgeCasesTaskStates|TestLargeDataHandling|TestSystemLimits|TestDataConsistency"
            ;;
        performance)
            test_pattern="-run TestHighVolumeTaskProcessing|TestWebSocketPerformance|BenchmarkTaskOperations"
            timeout_flag="-timeout 45m"  # Performance tests need more time
            ;;
        all)
            test_pattern=""  # Run all tests
            timeout_flag="-timeout 60m"  # All tests need more time
            ;;
        *)
            log_error "Unknown test suite: $suite"
            return 1
            ;;
    esac
    
    # Create test results directory
    mkdir -p test-results
    
    # Set environment variables for tests
    export DB_HOST=localhost
    export DB_USERNAME=postgres
    export DB_PASSWORD=postgres
    export DB_NAME=autodevs_test
    export REDIS_HOST=localhost
    export REDIS_PORT=6379
    
    # Run the tests
    local test_cmd="go test $timeout_flag $verbose_flag $parallel_flag $test_pattern ./internal/e2e_test"
    
    log_info "Executing: $test_cmd"
    
    if eval $test_cmd; then
        log_success "Test suite '$suite' completed successfully"
        return 0
    else
        log_error "Test suite '$suite' failed"
        return 1
    fi
}

# Generate test reports
generate_reports() {
    log_info "Generating test reports..."
    
    # Install go-junit-report if not available
    if ! command -v go-junit-report &> /dev/null; then
        go install github.com/jstemmer/go-junit-report/v2@latest
    fi
    
    # Re-run tests with JSON output for reporting
    local suites=()
    case $TEST_SUITE in
        all)
            suites=("happy-path" "error-scenarios" "edge-cases" "performance")
            ;;
        *)
            suites=("$TEST_SUITE")
            ;;
    esac
    
    for suite in "${suites[@]}"; do
        log_info "Generating report for $suite..."
        
        local test_pattern=""
        local timeout_flag="-timeout $TEST_TIMEOUT"
        
        case $suite in
            happy-path)
                test_pattern="-run TestCompleteTaskAutomationFlow|TestMultiTaskProjectWorkflow|TestPlanGenerationAndApproval|TestImplementationExecution|TestPRCreationAndMonitoring|TestWebSocketRealTimeUpdates"
                ;;
            error-scenarios)
                test_pattern="-run TestPlanningServiceFailure|TestImplementationServiceFailure|TestGitOperationFailures|TestGitHubAPIFailures|TestDatabaseConnectionFailures|TestConcurrencyIssues|TestResourceExhaustion"
                ;;
            edge-cases)
                test_pattern="-run TestEdgeCasesTaskStates|TestLargeDataHandling|TestSystemLimits|TestDataConsistency"
                ;;
            performance)
                test_pattern="-run TestHighVolumeTaskProcessing|TestWebSocketPerformance|BenchmarkTaskOperations"
                timeout_flag="-timeout 45m"
                ;;
        esac
        
        # Generate JSON report
        go test -json $timeout_flag $test_pattern ./internal/e2e_test > "test-results/${suite}.json" 2>&1 || true
        
        # Convert to JUnit XML
        if [[ -f "test-results/${suite}.json" ]]; then
            cat "test-results/${suite}.json" | go-junit-report -set-exit-code > "test-results/${suite}.xml" || true
        fi
    done
    
    log_success "Test reports generated in test-results/"
}

# Cleanup test resources
cleanup_resources() {
    if [[ $CLEANUP != "true" ]]; then
        log_info "Cleanup disabled, skipping resource cleanup"
        return 0
    fi
    
    log_info "Cleaning up test resources..."
    
    # Stop and remove containers
    if docker ps -q -f name=$DB_CONTAINER_NAME | grep -q .; then
        log_info "Stopping database container..."
        docker stop $DB_CONTAINER_NAME || true
        docker rm $DB_CONTAINER_NAME || true
    fi
    
    if docker ps -q -f name=$REDIS_CONTAINER_NAME | grep -q .; then
        log_info "Stopping Redis container..."
        docker stop $REDIS_CONTAINER_NAME || true
        docker rm $REDIS_CONTAINER_NAME || true
    fi
    
    # Clean up test database (if using local instance)
    # This is handled by container removal above
    
    log_success "Cleanup completed"
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -s|--suite)
                TEST_SUITE="$2"
                shift 2
                ;;
            -t|--timeout)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            -p|--parallel)
                PARALLEL="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --no-cleanup)
                CLEANUP=false
                shift
                ;;
            --setup-only)
                SETUP_ONLY=true
                shift
                ;;
            --cleanup-only)
                CLEANUP_ONLY=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate test suite option
validate_options() {
    case $TEST_SUITE in
        all|happy-path|error-scenarios|edge-cases|performance)
            ;;
        *)
            log_error "Invalid test suite: $TEST_SUITE"
            log_error "Valid options: all, happy-path, error-scenarios, edge-cases, performance"
            exit 1
            ;;
    esac
}

# Main execution function
main() {
    log_info "Starting E2E test execution..."
    log_info "Configuration:"
    log_info "  Test Suite: $TEST_SUITE"
    log_info "  Timeout: $TEST_TIMEOUT"
    log_info "  Parallel: $PARALLEL"
    log_info "  Verbose: $VERBOSE"
    log_info "  Cleanup: $CLEANUP"
    
    local exit_code=0
    
    # Handle special modes
    if [[ ${CLEANUP_ONLY:-false} == "true" ]]; then
        cleanup_resources
        exit $?
    fi
    
    # Setup phase
    check_prerequisites || exit_code=$?
    start_services || exit_code=$?
    setup_environment || exit_code=$?
    
    if [[ ${SETUP_ONLY:-false} == "true" ]]; then
        log_success "Setup completed successfully"
        exit $exit_code
    fi
    
    if [[ $exit_code -eq 0 ]]; then
        # Test execution phase
        run_test_suite "$TEST_SUITE" || exit_code=$?
        
        # Report generation
        generate_reports || log_warning "Report generation failed"
    fi
    
    # Cleanup phase
    cleanup_resources || log_warning "Cleanup failed"
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "E2E test execution completed successfully"
    else
        log_error "E2E test execution failed"
    fi
    
    exit $exit_code
}

# Handle script interruption
trap cleanup_resources EXIT INT TERM

# Parse arguments and run main function
parse_arguments "$@"
validate_options
main