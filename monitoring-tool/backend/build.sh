#!/bin/bash

# Build script for Real-time Monitoring Tool Backend
# This script builds the Go application for different platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Configuration
APP_NAME="monitoring-tool"
VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="build"
MAIN_FILE="main.go"

# Get current directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Clean previous builds
clean_builds() {
    print_status "Cleaning previous builds..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
}

# Check Go installation
check_go() {
    print_status "Checking Go installation..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_status "Download from: https://golang.org/dl/"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go version: $go_version"
}

# Download dependencies
download_deps() {
    print_status "Downloading Go dependencies..."
    
    if [ -f "go.mod" ]; then
        go mod download
        go mod tidy
        print_success "Dependencies downloaded successfully"
    else
        print_error "go.mod file not found. Make sure you're in the backend directory."
        exit 1
    fi
}

# Build for current platform
build_current() {
    local platform=$(go env GOOS)_$(go env GOARCH)
    local output_name="$APP_NAME-$platform"
    
    if [ "$platform" = "windows_amd64" ]; then
        output_name="$output_name.exe"
    fi
    
    print_status "Building for current platform: $platform"
    
    go build -ldflags "-X main.version=$VERSION -s -w" \
        -o "$BUILD_DIR/$output_name" \
        "$MAIN_FILE"
    
    if [ $? -eq 0 ]; then
        print_success "Build completed: $BUILD_DIR/$output_name"
        
        # Show file size
        local size=$(du -h "$BUILD_DIR/$output_name" | cut -f1)
        print_status "Binary size: $size"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Build for multiple platforms
build_all() {
    print_status "Building for multiple platforms..."
    
    local platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )
    
    for platform in "${platforms[@]}"; do
        local os=$(echo $platform | cut -d'/' -f1)
        local arch=$(echo $platform | cut -d'/' -f2)
        local output_name="$APP_NAME-$os-$arch"
        
        if [ "$os" = "windows" ]; then
            output_name="$output_name.exe"
        fi
        
        print_status "Building for $platform..."
        
        GOOS=$os GOARCH=$arch go build -ldflags "-X main.version=$VERSION -s -w" \
            -o "$BUILD_DIR/$output_name" \
            "$MAIN_FILE"
        
        if [ $? -eq 0 ]; then
            print_success "Built: $BUILD_DIR/$output_name"
        else
            print_error "Failed to build for $platform"
        fi
    done
}

# Build for production (optimized)
build_production() {
    local platform=$(go env GOOS)_$(go env GOARCH)
    local output_name="$APP_NAME-prod"
    
    if [ "$platform" = "windows_amd64" ]; then
        output_name="$output_name.exe"
    fi
    
    print_status "Building production binary (optimized)..."
    
    # Production build flags
    go build -ldflags "-X main.version=$VERSION -s -w" \
        -trimpath \
        -buildmode=pie \
        -o "$BUILD_DIR/$output_name" \
        "$MAIN_FILE"
    
    if [ $? -eq 0 ]; then
        print_success "Production build completed: $BUILD_DIR/$output_name"
        
        # Show file size and info
        local size=$(du -h "$BUILD_DIR/$output_name" | cut -f1)
        print_status "Binary size: $size"
        
        # Show binary info
        if command -v file &> /dev/null; then
            print_status "Binary info:"
            file "$BUILD_DIR/$output_name"
        fi
    else
        print_error "Production build failed"
        exit 1
    fi
}

# Test the build
test_build() {
    local binary="$BUILD_DIR/$APP_NAME-prod"
    if [ "$(go env GOOS)" = "windows" ]; then
        binary="$binary.exe"
    fi
    
    if [ ! -f "$binary" ]; then
        print_error "Binary not found: $binary"
        return 1
    fi
    
    print_status "Testing build..."
    
    # Test version flag
    local version_output=$($binary -version 2>&1 || echo "Version flag not implemented")
    print_status "Version output: $version_output"
    
    print_success "Build test completed"
}

# Show help
show_help() {
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  build        Build for current platform (default)"
    echo "  build-all    Build for all supported platforms"
    echo "  production   Build optimized production binary"
    echo "  test         Test the build"
    echo "  clean        Clean build directory"
    echo "  deps         Download dependencies only"
    echo "  help         Show this help message"
    echo
    echo "Environment Variables:"
    echo "  VERSION      Set application version (default: 1.0.0)"
    echo
    echo "Examples:"
    echo "  $0                    # Build for current platform"
    echo "  $0 production         # Build production binary"
    echo "  VERSION=2.0.0 $0      # Build with custom version"
}

# Main execution
main() {
    local command=${1:-build}
    
    case "$command" in
        "build")
            check_go
            clean_builds
            download_deps
            build_current
            ;;
        "build-all")
            check_go
            clean_builds
            download_deps
            build_all
            ;;
        "production")
            check_go
            clean_builds
            download_deps
            build_production
            test_build
            ;;
        "test")
            test_build
            ;;
        "clean")
            clean_builds
            ;;
        "deps")
            check_go
            download_deps
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
