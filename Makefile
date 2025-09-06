# Enterprise Status Page - Makefile for Docker Operations

.PHONY: help dev prod build clean logs shell backup restore

# Default target
help: ## Show this help message
	@echo "Enterprise Status Page - Docker Operations"
	@echo "=========================================="
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
dev: ## Start development environment
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development environment started!"
	@echo "Status Page: http://localhost:8080"
	@echo "Admin Login: http://localhost:8080/admin/login"

dev-build: ## Build and start development environment
	docker-compose -f docker-compose.dev.yml up -d --build

dev-logs: ## Show development logs
	docker-compose -f docker-compose.dev.yml logs -f

dev-stop: ## Stop development environment
	docker-compose -f docker-compose.dev.yml down

dev-clean: ## Stop and remove development containers and volumes
	docker-compose -f docker-compose.dev.yml down -v --remove-orphans

# Production commands
prod: ## Start production environment
	docker-compose up -d
	@echo "Production environment started!"
	@echo "Status Page: http://localhost:8080"
	@echo "Admin Login: http://localhost:8080/admin/login"

prod-build: ## Build and start production environment
	docker-compose up -d --build

prod-monitoring: ## Start production environment with monitoring
	docker-compose --profile monitoring up -d
	@echo "Production environment with monitoring started!"
	@echo "Status Page: http://localhost:8080"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000"

prod-logs: ## Show production logs
	docker-compose logs -f

prod-stop: ## Stop production environment
	docker-compose down

prod-clean: ## Stop and remove production containers and volumes
	docker-compose down -v --remove-orphans

# Build commands
build: ## Build the application image
	docker build -t statuspage:latest .

build-dev: ## Build the development image
	docker build -f Dockerfile.dev -t statuspage:dev .

# Utility commands
logs: ## Show all logs
	docker-compose logs -f

shell: ## Access application container shell
	docker-compose exec statuspage sh

db-shell: ## Access database shell
	docker-compose exec postgres psql -U postgres statuspage

nginx-shell: ## Access nginx container shell
	docker-compose exec nginx sh

# Backup and restore
backup: ## Backup database
	mkdir -p backups
	docker-compose exec postgres pg_dump -U postgres statuspage > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Database backup created in backups/ directory"

restore: ## Restore database from backup (usage: make restore BACKUP=backup_file.sql)
	@if [ -z "$(BACKUP)" ]; then echo "Usage: make restore BACKUP=backup_file.sql"; exit 1; fi
	docker-compose exec -T postgres psql -U postgres statuspage < $(BACKUP)
	@echo "Database restored from $(BACKUP)"

# Health checks
health: ## Check health of all services
	@echo "Checking service health..."
	@docker-compose ps
	@echo ""
	@echo "Application health:"
	@curl -s http://localhost:8080/health || echo "Application not responding"

# Cleanup commands
clean: ## Clean up Docker resources
	docker system prune -f
	docker volume prune -f

clean-all: ## Clean up all Docker resources (including images)
	docker system prune -a -f
	docker volume prune -f

# SSL setup
ssl-setup: ## Generate self-signed SSL certificates for development
	mkdir -p nginx/ssl
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout nginx/ssl/key.pem \
		-out nginx/ssl/cert.pem \
		-subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
	@echo "Self-signed SSL certificates generated in nginx/ssl/"

# Monitoring commands
prometheus-logs: ## Show Prometheus logs
	docker-compose logs -f prometheus

grafana-logs: ## Show Grafana logs
	docker-compose logs -f grafana

# Development utilities
test: ## Run tests
	docker-compose exec statuspage go test ./...

lint: ## Run linter
	docker-compose exec statuspage go vet ./...

fmt: ## Format code
	docker-compose exec statuspage go fmt ./...

# Status commands
status: ## Show status of all services
	docker-compose ps

restart: ## Restart all services
	docker-compose restart

restart-app: ## Restart only the application
	docker-compose restart statuspage
