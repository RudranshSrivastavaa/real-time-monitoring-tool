#!/bin/bash

# Health Check Script for Real-time Monitoring Tool
# This script checks the health of all services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:3000"
HEALTH_ENDPOINT="$BACKEND_URL/api/v1/health"

# Function to print colored output
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

# Check if URL is accessible
check_url() {
    local url=$1
    local service_name=$2
    
    print_status "Checking $service_name at $url..."
    
    if curl -f -s --max-time 10 "$url" > /dev/null 2>&1; then
        print_success "$service_name is healthy"
        return 0
    else
        print_error "$service_name is not accessible"
        return 1
    fi
}

# Check backend health endpoint
check_backend_health() {
    print_status "Checking backend health endpoint..."
    
    local response=$(curl -s --max-time 10 "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
    
    if [ -n "$response" ]; then
        local status=$(echo "$response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        
        if [ "$status" = "healthy" ]; then
            print_success "Backend health check passed"
            echo "Response: $response"
            return 0
        else
            print_error "Backend health check failed: $response"
            return 1
        fi
    else
        print_error "Backend health endpoint not accessible"
        return 1
    fi
}

# Check Docker containers
check_docker_containers() {
    print_status "Checking Docker containers..."
    
    local containers=("realtime_monitor_backend_prod" "realtime_monitor_frontend_prod" "realtime_monitor_db_prod")
    local all_healthy=true
    
    for container in "${containers[@]}"; do
        if docker ps --format "table {{.Names}}" | grep -q "$container"; then
            local status=$(docker inspect --format='{{.State.Status}}' "$container" 2>/dev/null)
            
            if [ "$status" = "running" ]; then
                print_success "Container $container is running"
            else
                print_error "Container $container is not running (status: $status)"
                all_healthy=false
            fi
        else
            print_error "Container $container is not found"
            all_healthy=false
        fi
    done
    
    if [ "$all_healthy" = true ]; then
        return 0
    else
        return 1
    fi
}

# Check database connection
check_database() {
    print_status "Checking database connection..."
    
    if docker exec realtime_monitor_db_prod mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
        print_success "Database is accessible"
        return 0
    else
        print_error "Database is not accessible"
        return 1
    fi
}

# Check WebSocket connection
check_websocket() {
    print_status "Checking WebSocket connection..."
    
    # This is a basic check - in production you might want a more sophisticated test
    if curl -f -s --max-time 10 "$BACKEND_URL/ws" > /dev/null 2>&1; then
        print_success "WebSocket endpoint is accessible"
        return 0
    else
        print_warning "WebSocket endpoint check failed (this might be expected)"
        return 1
    fi
}

# Get system resources
check_resources() {
    print_status "Checking system resources..."
    
    # Check disk space
    local disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$disk_usage" -lt 80 ]; then
        print_success "Disk usage is healthy: ${disk_usage}%"
    else
        print_warning "Disk usage is high: ${disk_usage}%"
    fi
    
    # Check memory usage
    local memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ "$memory_usage" -lt 80 ]; then
        print_success "Memory usage is healthy: ${memory_usage}%"
    else
        print_warning "Memory usage is high: ${memory_usage}%"
    fi
}

# Main health check function
main() {
    echo "🏥 Real-time Monitoring Tool Health Check"
    echo "=========================================="
    echo
    
    local exit_code=0
    
    # Check Docker containers
    if ! check_docker_containers; then
        exit_code=1
    fi
    
    echo
    
    # Check database
    if ! check_database; then
        exit_code=1
    fi
    
    echo
    
    # Check backend
    if ! check_backend_health; then
        exit_code=1
    fi
    
    echo
    
    # Check frontend
    if ! check_url "$FRONTEND_URL" "Frontend"; then
        exit_code=1
    fi
    
    echo
    
    # Check WebSocket
    check_websocket
    
    echo
    
    # Check system resources
    check_resources
    
    echo
    echo "=========================================="
    
    if [ $exit_code -eq 0 ]; then
        print_success "All health checks passed! ✅"
        echo
        echo "Service URLs:"
        echo "  Frontend: $FRONTEND_URL"
        echo "  Backend API: $BACKEND_URL/api/v1"
        echo "  Health Check: $HEALTH_ENDPOINT"
        echo "  WebSocket: ws://localhost:8080/ws"
    else
        print_error "Some health checks failed! ❌"
        echo
        echo "Troubleshooting tips:"
        echo "  1. Check service logs: ./deploy.sh logs"
        echo "  2. Restart services: ./deploy.sh restart"
        echo "  3. Check Docker containers: docker ps"
        echo "  4. Verify environment configuration"
    fi
    
    exit $exit_code
}

# Handle command line arguments
case "${1:-}" in
    "backend")
        check_backend_health
        ;;
    "frontend")
        check_url "$FRONTEND_URL" "Frontend"
        ;;
    "database")
        check_database
        ;;
    "containers")
        check_docker_containers
        ;;
    "resources")
        check_resources
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo
        echo "Commands:"
        echo "  (no command)  Run all health checks"
        echo "  backend       Check backend health only"
        echo "  frontend      Check frontend health only"
        echo "  database      Check database health only"
        echo "  containers    Check Docker containers only"
        echo "  resources     Check system resources only"
        echo "  help          Show this help message"
        ;;
    *)
        main
        ;;
esac
