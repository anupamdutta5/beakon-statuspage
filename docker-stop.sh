#!/bin/bash

# Beakon Microservices - Docker Stop Script
# This script stops all microservices and cleans up containers

set -e

echo "🛑 Stopping Beakon Microservices Platform"
echo "=========================================="

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

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    print_error "Docker is not running."
    exit 1
fi

# Parse command line arguments
CLEAN_VOLUMES=false
REMOVE_IMAGES=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN_VOLUMES=true
            shift
            ;;
        --remove-images)
            REMOVE_IMAGES=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --clean         Remove volumes (deletes all data)"
            echo "  --remove-images Remove built images"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

print_status "Stopping all running containers..."

# Stop all containers
if $CLEAN_VOLUMES; then
    print_warning "Stopping containers and removing volumes (all data will be lost)..."
    docker-compose down --volumes --remove-orphans
else
    docker-compose down --remove-orphans
fi

if [ $? -eq 0 ]; then
    print_success "All containers stopped successfully"
else
    print_error "Some containers failed to stop properly"
fi

# Remove built images if requested
if $REMOVE_IMAGES; then
    print_warning "Removing built images..."

    # Get list of images built by docker-compose
    images=$(docker images --filter "label=com.docker.compose.project=beakon" -q)

    if [ -n "$images" ]; then
        docker rmi $images 2>/dev/null || true
        print_success "Built images removed"
    else
        print_status "No built images found to remove"
    fi

    # Clean up dangling images
    docker image prune -f >/dev/null 2>&1 || true
fi

# Show remaining containers (if any)
running_containers=$(docker-compose ps -q 2>/dev/null | wc -l)
if [ "$running_containers" -gt 0 ]; then
    print_warning "Some containers are still running:"
    docker-compose ps
else
    print_success "All containers have been stopped"
fi

# Clean up system resources if requested
if $CLEAN_VOLUMES; then
    print_status "Cleaning up unused Docker resources..."
    docker system prune -f >/dev/null 2>&1 || true
fi

echo ""
print_success "🎉 Beakon Platform Stopped Successfully!"
echo ""
echo "📋 Summary:"
if $CLEAN_VOLUMES; then
    echo "   • All containers stopped and removed"
    echo "   • All volumes removed (data deleted)"
    echo "   • System resources cleaned up"
else
    echo "   • All containers stopped and removed"
    echo "   • Volumes preserved (data safe)"
fi

if $REMOVE_IMAGES; then
    echo "   • Built images removed"
fi

echo ""
echo "🚀 To start again: ./docker-start.sh"
echo "🧹 To clean everything: ./docker-stop.sh --clean --remove-images"