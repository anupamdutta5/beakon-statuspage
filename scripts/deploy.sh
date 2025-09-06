#!/bin/bash

# Deployment script for the status page application

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

# Function to check environment variables
check_env_vars() {
    local env_file="$1"
    
    if [ ! -f "$env_file" ]; then
        print_error "Environment file $env_file not found."
        exit 1
    fi
    
    # Check required environment variables
    local required_vars=(
        "DATABASE_URL"
        "REDIS_URL"
        "JWT_SECRET"
        "ADMIN_EMAIL"
        "ADMIN_PASSWORD"
    )
    
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" "$env_file"; then
            print_error "Required environment variable $var not found in $env_file"
            exit 1
        fi
    done
    
    print_success "Environment variables validated"
}

# Function to backup database
backup_database() {
    local backup_dir="$1"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$backup_dir/backup_$timestamp.sql"
    
    print_status "Creating database backup..."
    
    # Create backup directory if it doesn't exist
    mkdir -p "$backup_dir"
    
    # Create database backup
    if command_exists docker-compose; then
        docker-compose exec -T postgres pg_dump -U postgres statuspage > "$backup_file"
    else
        docker compose exec -T postgres pg_dump -U postgres statuspage > "$backup_file"
    fi
    
    if [ $? -eq 0 ]; then
        print_success "Database backup created: $backup_file"
    else
        print_error "Failed to create database backup"
        exit 1
    fi
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    # Build the application
    if command_exists docker-compose; then
        docker-compose build app
    else
        docker compose build app
    fi
    
    # Run migrations
    if command_exists docker-compose; then
        docker-compose run --rm app go run cmd/migrate/main.go
    else
        docker compose run --rm app go run cmd/migrate/main.go
    fi
    
    if [ $? -eq 0 ]; then
        print_success "Database migrations completed"
    else
        print_error "Database migrations failed"
        exit 1
    fi
}

# Function to deploy application
deploy_application() {
    local env_file="$1"
    
    print_status "Deploying application..."
    
    # Stop existing containers
    if command_exists docker-compose; then
        docker-compose down
    else
        docker compose down
    fi
    
    # Start new containers
    if command_exists docker-compose; then
        docker-compose --env-file "$env_file" up -d
    else
        docker compose --env-file "$env_file" up -d
    fi
    
    if [ $? -eq 0 ]; then
        print_success "Application deployed successfully"
    else
        print_error "Application deployment failed"
        exit 1
    fi
}

# Function to wait for application to be ready
wait_for_application() {
    local max_attempts=30
    local attempt=1
    
    print_status "Waiting for application to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s http://localhost:8080/api/v1/status >/dev/null 2>&1; then
            print_success "Application is ready"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - Application not ready yet, waiting..."
        sleep 10
        ((attempt++))
    done
    
    print_error "Application failed to become ready after $max_attempts attempts"
    exit 1
}

# Function to run health checks
run_health_checks() {
    print_status "Running health checks..."
    
    # Check if application is responding
    if ! curl -f -s http://localhost:8080/api/v1/status >/dev/null 2>&1; then
        print_error "Application health check failed"
        exit 1
    fi
    
    # Check if admin endpoint is accessible
    if ! curl -f -s http://localhost:8080/api/v1/admin/status >/dev/null 2>&1; then
        print_error "Admin endpoint health check failed"
        exit 1
    fi
    
    print_success "All health checks passed"
}

# Function to rollback deployment
rollback_deployment() {
    local backup_dir="$1"
    
    print_warning "Rolling back deployment..."
    
    # Stop current containers
    if command_exists docker-compose; then
        docker-compose down
    else
        docker compose down
    fi
    
    # Restore database from backup
    local latest_backup=$(ls -t "$backup_dir"/backup_*.sql 2>/dev/null | head -n1)
    if [ -n "$latest_backup" ]; then
        print_status "Restoring database from backup: $latest_backup"
        
        if command_exists docker-compose; then
            docker-compose up -d postgres
            sleep 10
            docker-compose exec -T postgres psql -U postgres -d statuspage < "$latest_backup"
        else
            docker compose up -d postgres
            sleep 10
            docker compose exec -T postgres psql -U postgres -d statuspage < "$latest_backup"
        fi
        
        print_success "Database restored from backup"
    else
        print_warning "No backup found for rollback"
    fi
    
    # Start previous version
    if command_exists docker-compose; then
        docker-compose up -d
    else
        docker compose up -d
    fi
    
    print_success "Rollback completed"
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  deploy      Deploy the application"
    echo "  rollback    Rollback to previous version"
    echo "  backup      Create database backup"
    echo "  migrate     Run database migrations"
    echo "  health      Run health checks"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  -e, --env-file FILE    Environment file (default: .env)"
    echo "  -b, --backup-dir DIR   Backup directory (default: ./backups)"
    echo "  -f, --force           Force deployment without confirmation"
    echo ""
    echo "Examples:"
    echo "  $0 deploy"
    echo "  $0 deploy -e .env.production"
    echo "  $0 rollback -b ./backups"
    echo "  $0 backup -b ./backups"
}

# Main script logic
main() {
    # Default values
    local env_file=".env"
    local backup_dir="./backups"
    local force=false
    local command=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            deploy|rollback|backup|migrate|health|help)
                command="$1"
                shift
                ;;
            -e|--env-file)
                env_file="$2"
                shift 2
                ;;
            -b|--backup-dir)
                backup_dir="$2"
                shift 2
                ;;
            -f|--force)
                force=true
                shift
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
        "deploy")
            if [ "$force" = false ]; then
                read -p "Are you sure you want to deploy? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    print_warning "Deployment cancelled"
                    exit 0
                fi
            fi
            
            check_env_vars "$env_file"
            backup_database "$backup_dir"
            run_migrations
            deploy_application "$env_file"
            wait_for_application
            run_health_checks
            print_success "Deployment completed successfully!"
            ;;
        "rollback")
            if [ "$force" = false ]; then
                read -p "Are you sure you want to rollback? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    print_warning "Rollback cancelled"
                    exit 0
                fi
            fi
            
            rollback_deployment "$backup_dir"
            wait_for_application
            run_health_checks
            print_success "Rollback completed successfully!"
            ;;
        "backup")
            backup_database "$backup_dir"
            ;;
        "migrate")
            check_env_vars "$env_file"
            run_migrations
            ;;
        "health")
            run_health_checks
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
