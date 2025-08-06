#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[DEV-RUNNER]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to cleanup processes on exit
cleanup() {
    print_status "Shutting down development servers..."
    
    # Kill all child processes
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null
        print_status "Stopped Go server (PID: $SERVER_PID)"
    fi
    
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
        print_status "Stopped React frontend (PID: $FRONTEND_PID)"
    fi

    if [ ! -z "$WORKER_PID" ]; then
        kill $WORKER_PID 2>/dev/null
        print_status "Stopped worker (PID: $WORKER_PID)"
    fi
    
    # Kill any remaining processes in our process group
    pkill -P $$ 2>/dev/null
    
    print_success "Development servers stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM EXIT

print_status "Starting Auto-Devs development environment..."

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed or not in PATH"
        exit 1
    fi
    
    print_success "All dependencies found"
}

# Function to start the Go server
start_server() {
    print_status "Starting Go server..."
    
    # Check if we're in the right directory
    if [ ! -f "cmd/server/main.go" ]; then
        print_error "Cannot find cmd/server/main.go. Make sure you're running this from the project root."
        exit 1
    fi
    
    # Start the server in background
    go run cmd/server/main.go &
    SERVER_PID=$!
    
    print_success "Go server started (PID: $SERVER_PID)"
    print_status "Server will be available at: http://localhost:8098"
}

start_worker() {
    print_status "Starting worker..."
    go run cmd/worker/main.go &
    WORKER_PID=$!
    print_success "Worker started (PID: $WORKER_PID)"
}

# Function to start the React frontend
start_frontend() {
    print_status "Starting React frontend..."
    
    # Check if frontend directory exists
    if [ ! -d "frontend" ]; then
        print_error "Cannot find frontend directory"
        exit 1
    fi
    
    # Check if package.json exists
    if [ ! -f "frontend/package.json" ]; then
        print_error "Cannot find frontend/package.json"
        exit 1
    fi
    
    # Install dependencies if node_modules doesn't exist
    if [ ! -d "frontend/node_modules" ]; then
        print_warning "node_modules not found. Installing frontend dependencies..."
        cd frontend && npm install && cd ..
        if [ $? -ne 0 ]; then
            print_error "Failed to install frontend dependencies"
            exit 1
        fi
        print_success "Frontend dependencies installed"
    fi
    
    # Start the frontend in background
    cd frontend && npm run dev &
    FRONTEND_PID=$!
    cd ..
    
    print_success "React frontend started (PID: $FRONTEND_PID)"
    print_status "Frontend will be available at: http://localhost:5173"
}

# Function to wait for processes and monitor them
monitor_processes() {
    print_status "Monitoring development servers..."
    print_status "Press Ctrl+C to stop all servers"
    print_status ""
    print_status "Available services:"
    print_status "  • Go Server:      http://localhost:8098"
    print_status "  • Worker:         http://localhost:8098/worker"
    print_status "  • React Frontend: http://localhost:5173"
    print_status "  • Swagger UI:     http://localhost:8098/swagger/index.html"
    print_status ""
    
    # Wait for either process to exit
    while true; do
        # Check if server process is still running
        if ! kill -0 $SERVER_PID 2>/dev/null; then
            print_error "Go server process died unexpectedly"
            cleanup
        fi
        
        # Check if frontend process is still running
        if ! kill -0 $FRONTEND_PID 2>/dev/null; then
            print_error "React frontend process died unexpectedly"
            cleanup
        fi
        
        # Check if worker process is still running
        if ! kill -0 $WORKER_PID 2>/dev/null; then
            print_error "Worker process died unexpectedly"
            cleanup
        fi

        sleep 2
    done
}

# Main execution
main() {
    check_dependencies
    start_server
    sleep 2  # Give server a moment to start
    start_worker
    sleep 2  # Give frontend a moment to start
    start_frontend
    sleep 2  # Give worker a moment to start
    monitor_processes
}

# Run the main function
main