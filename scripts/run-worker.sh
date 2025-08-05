#!/bin/bash

# Job Worker Runner Script
# This script runs the background job processor

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
WORKER_NAME="worker-$(date +%s)"
VERBOSE=false
REDIS_HOST=${REDIS_HOST:-"localhost"}
REDIS_PORT=${REDIS_PORT:-"6379"}

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n, --name NAME     Worker name (default: auto-generated)"
    echo "  -v, --verbose       Enable verbose logging"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  REDIS_HOST          Redis host (default: localhost)"
    echo "  REDIS_PORT          Redis port (default: 6379)"
    echo "  REDIS_PASSWORD      Redis password (optional)"
    echo "  REDIS_DB            Redis database (default: 0)"
    echo ""
    echo "Examples:"
    echo "  $0                           # Run with default settings"
    echo "  $0 -n planning-worker-1      # Run with custom name"
    echo "  $0 -v                        # Run with verbose logging"
    echo "  REDIS_HOST=redis.example.com $0  # Run with custom Redis host"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            WORKER_NAME="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if Redis is running
check_redis() {
    print_info "Checking Redis connection..."

    if command -v redis-cli &> /dev/null; then
        if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping &> /dev/null; then
            print_success "Redis is running at $REDIS_HOST:$REDIS_PORT"
        else
            print_error "Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
            print_info "Make sure Redis is running and accessible"
            exit 1
        fi
    else
        print_warning "redis-cli not found, skipping Redis connection check"
    fi
}

# Check if the worker binary exists
check_binary() {
    if [ ! -f "./worker" ]; then
        print_info "Building worker binary..."
        go build -o worker ./cmd/worker
        if [ $? -ne 0 ]; then
            print_error "Failed to build worker binary"
            exit 1
        fi
        print_success "Worker binary built successfully"
    fi
}

# Main execution
main() {
    print_info "Starting Job Worker"
    print_info "Worker Name: $WORKER_NAME"
    print_info "Redis Host: $REDIS_HOST:$REDIS_PORT"

    # Check dependencies
    check_redis
    check_binary

    # Set up environment variables
    export REDIS_HOST
    export REDIS_PORT
    export REDIS_PASSWORD=${REDIS_PASSWORD:-""}
    export REDIS_DB=${REDIS_DB:-"0"}

    # Build command
    CMD="./worker -worker=$WORKER_NAME"
    if [ "$VERBOSE" = true ]; then
        CMD="$CMD -verbose"
    fi

    print_info "Starting worker with command: $CMD"
    print_info "Press Ctrl+C to stop the worker"
    echo ""

    # Run the worker
    exec $CMD
}

# Run main function
main "$@"