#!/bin/bash

# Monitoring script for the status page application

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if Docker is running
check_docker() {
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! command_exists docker-compose && ! docker compose version >/dev/null 2>&1; then
        print_error "Docker Compose is not available. Please install Docker Compose first."
        exit 1
    fi
}

# Function to get container status
get_container_status() {
    local container_name="$1"
    
    if command_exists docker-compose; then
        docker-compose ps "$container_name" --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
    else
        docker compose ps "$container_name" --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
    fi
}

# Function to check application health
check_application_health() {
    print_status "Checking application health..."
    
    # Check if application is responding
    if curl -f -s http://localhost:8080/api/v1/status >/dev/null 2>&1; then
        print_success "Application is healthy"
        return 0
    else
        print_error "Application is not responding"
        return 1
    fi
}

# Function to check database health
check_database_health() {
    print_status "Checking database health..."
    
    # Check if database container is running
    local db_status=$(get_container_status "postgres")
    if echo "$db_status" | grep -q "Up"; then
        print_success "Database container is running"
    else
        print_error "Database container is not running"
        return 1
    fi
    
    # Check database connection
    if command_exists docker-compose; then
        if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
            print_success "Database is accepting connections"
        else
            print_error "Database is not accepting connections"
            return 1
        fi
    else
        if docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
            print_success "Database is accepting connections"
        else
            print_error "Database is not accepting connections"
            return 1
        fi
    fi
}

# Function to check Redis health
check_redis_health() {
    print_status "Checking Redis health..."
    
    # Check if Redis container is running
    local redis_status=$(get_container_status "redis")
    if echo "$redis_status" | grep -q "Up"; then
        print_success "Redis container is running"
    else
        print_error "Redis container is not running"
        return 1
    fi
    
    # Check Redis connection
    if command_exists docker-compose; then
        if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
            print_success "Redis is accepting connections"
        else
            print_error "Redis is not accepting connections"
            return 1
        fi
    else
        if docker compose exec -T redis redis-cli ping >/dev/null 2>&1; then
            print_success "Redis is accepting connections"
        else
            print_error "Redis is not accepting connections"
            return 1
        fi
    fi
}

# Function to check disk usage
check_disk_usage() {
    print_status "Checking disk usage..."
    
    # Check disk usage
    local disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$disk_usage" -lt 80 ]; then
        print_success "Disk usage is normal: ${disk_usage}%"
    elif [ "$disk_usage" -lt 90 ]; then
        print_warning "Disk usage is high: ${disk_usage}%"
    else
        print_error "Disk usage is critical: ${disk_usage}%"
        return 1
    fi
}

# Function to check memory usage
check_memory_usage() {
    print_status "Checking memory usage..."
    
    # Check memory usage
    local memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    
    if [ "$memory_usage" -lt 80 ]; then
        print_success "Memory usage is normal: ${memory_usage}%"
    elif [ "$memory_usage" -lt 90 ]; then
        print_warning "Memory usage is high: ${memory_usage}%"
    else
        print_error "Memory usage is critical: ${memory_usage}%"
        return 1
    fi
}

# Function to check CPU usage
check_cpu_usage() {
    print_status "Checking CPU usage..."
    
    # Check CPU usage
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    
    if [ "$cpu_usage" -lt 80 ]; then
        print_success "CPU usage is normal: ${cpu_usage}%"
    elif [ "$cpu_usage" -lt 90 ]; then
        print_warning "CPU usage is high: ${cpu_usage}%"
    else
        print_error "CPU usage is critical: ${cpu_usage}%"
        return 1
    fi
}

# Function to check container resource usage
check_container_resources() {
    print_status "Checking container resource usage..."
    
    # Check container resource usage
    if command_exists docker-compose; then
        docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
    else
        docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
    fi
}

# Function to check application logs
check_application_logs() {
    print_status "Checking application logs..."
    
    # Get recent application logs
    if command_exists docker-compose; then
        docker-compose logs --tail=50 app
    else
        docker compose logs --tail=50 app
    fi
}

# Function to check database logs
check_database_logs() {
    print_status "Checking database logs..."
    
    # Get recent database logs
    if command_exists docker-compose; then
        docker-compose logs --tail=50 postgres
    else
        docker compose logs --tail=50 postgres
    fi
}

# Function to check Redis logs
check_redis_logs() {
    print_status "Checking Redis logs..."
    
    # Get recent Redis logs
    if command_exists docker-compose; then
        docker-compose logs --tail=50 redis
    else
        docker compose logs --tail=50 redis
    fi
}

# Function to run comprehensive health check
run_health_check() {
    print_status "Running comprehensive health check..."
    
    local health_status=0
    
    # Check application health
    if ! check_application_health; then
        health_status=1
    fi
    
    # Check database health
    if ! check_database_health; then
        health_status=1
    fi
    
    # Check Redis health
    if ! check_redis_health; then
        health_status=1
    fi
    
    # Check system resources
    if ! check_disk_usage; then
        health_status=1
    fi
    
    if ! check_memory_usage; then
        health_status=1
    fi
    
    if ! check_cpu_usage; then
        health_status=1
    fi
    
    if [ $health_status -eq 0 ]; then
        print_success "All health checks passed"
    else
        print_error "Some health checks failed"
    fi
    
    return $health_status
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  health      Run comprehensive health check"
    echo "  app         Check application health only"
    echo "  db          Check database health only"
    echo "  redis       Check Redis health only"
    echo "  resources   Check system resource usage"
    echo "  containers  Check container status and resources"
    echo "  logs        Show application logs"
    echo "  db-logs     Show database logs"
    echo "  redis-logs  Show Redis logs"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  -f, --follow    Follow logs in real-time"
    echo "  -n, --lines N   Number of log lines to show (default: 50)"
    echo ""
    echo "Examples:"
    echo "  $0 health"
    echo "  $0 app"
    echo "  $0 logs -n 100"
    echo "  $0 logs -f"
}

# Main script logic
main() {
    # Default values
    local follow=false
    local lines=50
    local command=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            health|app|db|redis|resources|containers|logs|db-logs|redis-logs|help)
                command="$1"
                shift
                ;;
            -f|--follow)
                follow=true
                shift
                ;;
            -n|--lines)
                lines="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Check prerequisites
    check_docker
    check_docker_compose
    
    # Execute command
    case "$command" in
        "health")
            run_health_check
            ;;
        "app")
            check_application_health
            ;;
        "db")
            check_database_health
            ;;
        "redis")
            check_redis_health
            ;;
        "resources")
            check_disk_usage
            check_memory_usage
            check_cpu_usage
            ;;
        "containers")
            check_container_resources
            ;;
        "logs")
            if [ "$follow" = true ]; then
                if command_exists docker-compose; then
                    docker-compose logs -f --tail="$lines" app
                else
                    docker compose logs -f --tail="$lines" app
                fi
            else
                if command_exists docker-compose; then
                    docker-compose logs --tail="$lines" app
                else
                    docker compose logs --tail="$lines" app
                fi
            fi
            ;;
        "db-logs")
            if [ "$follow" = true ]; then
                if command_exists docker-compose; then
                    docker-compose logs -f --tail="$lines" postgres
                else
                    docker compose logs -f --tail="$lines" postgres
                fi
            else
                if command_exists docker-compose; then
                    docker-compose logs --tail="$lines" postgres
                else
                    docker compose logs --tail="$lines" postgres
                fi
            fi
            ;;
        "redis-logs")
            if [ "$follow" = true ]; then
                if command_exists docker-compose; then
                    docker-compose logs -f --tail="$lines" redis
                else
                    docker compose logs -f --tail="$lines" redis
                fi
            else
                if command_exists docker-compose; then
                    docker-compose logs --tail="$lines" redis
                else
                    docker compose logs --tail="$lines" redis
                fi
            fi
            ;;
        "help"|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
