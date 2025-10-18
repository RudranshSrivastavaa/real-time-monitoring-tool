#!/bin/bash

# Simple deployment script for Real-time Monitoring Tool
# Perfect for college projects - no Docker needed!

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
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

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install Node.js 18 or later."
        exit 1
    fi
    
    # Check MongoDB
    if ! command -v mongod &> /dev/null; then
        print_error "MongoDB is not installed. Please install MongoDB."
        exit 1
    fi
    
    print_success "All prerequisites are installed"
}

# Setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    # Create data directory
    mkdir -p data/db
    
    # Create backend .env if it doesn't exist
    if [ ! -f backend/.env ]; then
        cat > backend/.env << EOF
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=realtime_monitor
PORT=8080
HOST=localhost
ENVIRONMENT=production
ALLOWED_ORIGINS=http://localhost:3000
EOF
        print_success "Created backend/.env file"
    fi
}

# Build backend
build_backend() {
    print_status "Building backend..."
    
    cd backend
    
    # Download dependencies
    go mod download
    go mod tidy
    
    # Build the application
    go build -o monitoring-tool main.go
    
    print_success "Backend built successfully"
    cd ..
}

# Build frontend
build_frontend() {
    print_status "Building frontend..."
    
    cd frontend/monitoring-dashboard
    
    # Install dependencies
    npm install
    
    # Build for production
    npm run build
    
    print_success "Frontend built successfully"
    cd ../..
}

# Start services
start_services() {
    print_status "Starting services..."
    
    # Start MongoDB
    if ! pgrep -x "mongod" > /dev/null; then
        print_status "Starting MongoDB..."
        mongod --dbpath ./data/db --fork --logpath ./data/mongodb.log
        sleep 3
    else
        print_status "MongoDB is already running"
    fi
    
    # Start backend
    print_status "Starting backend..."
    cd backend
    ./monitoring-tool &
    BACKEND_PID=$!
    echo $BACKEND_PID > ../backend.pid
    cd ..
    sleep 3
    
    # Start frontend (using serve)
    print_status "Starting frontend..."
    cd frontend/monitoring-dashboard
    
    # Install serve if not present
    if ! command -v serve &> /dev/null; then
        npm install -g serve
    fi
    
    serve -s build -l 3000 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../../frontend.pid
    cd ../..
    
    print_success "All services started successfully"
}

# Check health
check_health() {
    print_status "Checking service health..."
    
    sleep 5
    
    # Check backend
    if curl -f http://localhost:8080/api/v1/health &> /dev/null; then
        print_success "Backend is healthy"
    else
        print_warning "Backend health check failed"
    fi
    
    # Check frontend
    if curl -f http://localhost:3000 &> /dev/null; then
        print_success "Frontend is healthy"
    else
        print_warning "Frontend health check failed"
    fi
}

# Show deployment info
show_info() {
    print_success "Deployment completed successfully!"
    echo
    print_status "Application URLs:"
    echo "  Frontend: http://localhost:3000"
    echo "  Backend API: http://localhost:8080/api/v1"
    echo "  WebSocket: ws://localhost:8080/ws"
    echo "  Health Check: http://localhost:8080/api/v1/health"
    echo
    print_status "Useful commands:"
    echo "  Stop services: ./deploy.sh stop"
    echo "  View logs: Check the terminal output"
    echo "  Restart: ./deploy.sh restart"
}

# Stop services
stop_services() {
    print_status "Stopping services..."
    
    # Stop backend
    if [ -f "backend.pid" ]; then
        BACKEND_PID=$(cat backend.pid)
        if kill -0 $BACKEND_PID 2>/dev/null; then
            kill $BACKEND_PID
            rm backend.pid
            print_success "Backend stopped"
        fi
    fi
    
    # Stop frontend
    if [ -f "frontend.pid" ]; then
        FRONTEND_PID=$(cat frontend.pid)
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            kill $FRONTEND_PID
            rm frontend.pid
            print_success "Frontend stopped"
        fi
    fi
    
    # Stop MongoDB
    if pgrep -x "mongod" > /dev/null; then
        pkill mongod
        print_success "MongoDB stopped"
    fi
}

# Main function
main() {
    print_status "Starting Real-time Monitoring Tool deployment..."
    echo
    
    check_prerequisites
    setup_environment
    build_backend
    build_frontend
    start_services
    check_health
    show_info
}

# Handle arguments
case "${1:-}" in
    "stop")
        stop_services
        ;;
    "restart")
        stop_services
        sleep 2
        main
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo
        echo "Commands:"
        echo "  (no command)  Deploy the application"
        echo "  stop          Stop all services"
        echo "  restart       Restart all services"
        echo "  help          Show this help message"
        ;;
    *)
        main
        ;;
esac
