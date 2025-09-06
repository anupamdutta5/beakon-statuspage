# Makefile for the status page application

# Variables
DOCKER_COMPOSE = docker-compose
DOCKER_COMPOSE_V2 = docker compose
GO_VERSION = 1.21
APP_NAME = statuspage
VERSION ?= latest

# Check if docker-compose or docker compose is available
ifeq ($(shell command -v docker-compose),)
    COMPOSE_CMD = $(DOCKER_COMPOSE_V2)
else
    COMPOSE_CMD = $(DOCKER_COMPOSE)
endif

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development targets
.PHONY: dev
dev: ## Start development environment
	$(COMPOSE_CMD) up -d

.PHONY: dev-build
dev-build: ## Build and start development environment
	$(COMPOSE_CMD) up -d --build

.PHONY: dev-logs
dev-logs: ## Show development logs
	$(COMPOSE_CMD) logs -f

.PHONY: dev-stop
dev-stop: ## Stop development environment
	$(COMPOSE_CMD) down

.PHONY: dev-clean
dev-clean: ## Clean development environment (remove volumes)
	$(COMPOSE_CMD) down -v

# Build targets
.PHONY: build
build: ## Build the application
	$(COMPOSE_CMD) build

.PHONY: build-app
build-app: ## Build only the application container
	$(COMPOSE_CMD) build app

.PHONY: build-frontend
build-frontend: ## Build only the frontend container
	$(COMPOSE_CMD) build frontend

# Test targets
.PHONY: test
test: ## Run all tests
	./scripts/test.sh all

.PHONY: test-unit
test-unit: ## Run unit tests only
	./scripts/test.sh unit

.PHONY: test-integration
test-integration: ## Run integration tests only
	./scripts/test.sh integration

.PHONY: test-load
test-load: ## Run load tests only
	./scripts/test.sh load

.PHONY: test-security
test-security: ## Run security tests only
	./scripts/test.sh security

# Database targets
.PHONY: db-migrate
db-migrate: ## Run database migrations
	$(COMPOSE_CMD) run --rm app go run cmd/migrate/main.go

.PHONY: db-seed
db-seed: ## Seed database with sample data
	$(COMPOSE_CMD) run --rm app go run cmd/seed/main.go

.PHONY: db-reset
db-reset: ## Reset database (drop, create, migrate, seed)
	$(COMPOSE_CMD) down -v
	$(COMPOSE_CMD) up -d postgres redis
	sleep 10
	$(COMPOSE_CMD) run --rm app go run cmd/migrate/main.go
	$(COMPOSE_CMD) run --rm app go run cmd/seed/main.go

.PHONY: db-backup
db-backup: ## Create database backup
	./scripts/deploy.sh backup

.PHONY: db-restore
db-restore: ## Restore database from backup
	@echo "Please specify backup file: make db-restore BACKUP_FILE=backup.sql"
	@if [ -z "$(BACKUP_FILE)" ]; then exit 1; fi
	$(COMPOSE_CMD) exec -T postgres psql -U postgres -d statuspage < $(BACKUP_FILE)

# Deployment targets
.PHONY: deploy
deploy: ## Deploy the application
	./scripts/deploy.sh deploy

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging environment
	./scripts/deploy.sh deploy -e .env.staging

.PHONY: deploy-production
deploy-production: ## Deploy to production environment
	./scripts/deploy.sh deploy -e .env.production

.PHONY: rollback
rollback: ## Rollback to previous version
	./scripts/deploy.sh rollback

# Monitoring targets
.PHONY: monitor
monitor: ## Run comprehensive health check
	./scripts/monitor.sh health

.PHONY: monitor-app
monitor-app: ## Check application health
	./scripts/monitor.sh app

.PHONY: monitor-db
monitor-db: ## Check database health
	./scripts/monitor.sh db

.PHONY: monitor-redis
monitor-redis: ## Check Redis health
	./scripts/monitor.sh redis

.PHONY: monitor-resources
monitor-resources: ## Check system resource usage
	./scripts/monitor.sh resources

.PHONY: logs
logs: ## Show application logs
	./scripts/monitor.sh logs

.PHONY: logs-follow
logs-follow: ## Follow application logs
	./scripts/monitor.sh logs -f

.PHONY: logs-db
logs-db: ## Show database logs
	./scripts/monitor.sh db-logs

.PHONY: logs-redis
logs-redis: ## Show Redis logs
	./scripts/monitor.sh redis-logs

# Utility targets
.PHONY: clean
clean: ## Clean up build artifacts and containers
	$(COMPOSE_CMD) down -v --remove-orphans
	docker system prune -f

.PHONY: clean-all
clean-all: ## Clean up everything (including images)
	$(COMPOSE_CMD) down -v --remove-orphans
	docker system prune -af

.PHONY: format
format: ## Format Go code
	go fmt ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy Go modules
	go mod tidy

.PHONY: mod-download
mod-download: ## Download Go modules
	go mod download

.PHONY: mod-verify
mod-verify: ## Verify Go modules
	go mod verify

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	docker push $(APP_NAME):$(VERSION)

.PHONY: docker-pull
docker-pull: ## Pull Docker image from registry
	docker pull $(APP_NAME):$(VERSION)

# Development tools
.PHONY: shell
shell: ## Open shell in application container
	$(COMPOSE_CMD) exec app /bin/bash

.PHONY: shell-db
shell-db: ## Open shell in database container
	$(COMPOSE_CMD) exec postgres /bin/bash

.PHONY: shell-redis
shell-redis: ## Open shell in Redis container
	$(COMPOSE_CMD) exec redis /bin/bash

.PHONY: psql
psql: ## Connect to database with psql
	$(COMPOSE_CMD) exec postgres psql -U postgres -d statuspage

.PHONY: redis-cli
redis-cli: ## Connect to Redis with redis-cli
	$(COMPOSE_CMD) exec redis redis-cli

# Status targets
.PHONY: status
status: ## Show application status
	$(COMPOSE_CMD) ps

.PHONY: status-detailed
status-detailed: ## Show detailed application status
	$(COMPOSE_CMD) ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "Container resource usage:"
	docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"

# Health check targets
.PHONY: health
health: ## Run health check
	curl -f http://localhost:8080/api/v1/status || echo "Application is not healthy"

.PHONY: health-admin
health-admin: ## Run admin health check
	curl -f http://localhost:8080/api/v1/admin/status || echo "Admin endpoint is not healthy"

# Quick start targets
.PHONY: quick-start
quick-start: ## Quick start for development
	@echo "Starting development environment..."
	$(COMPOSE_CMD) up -d --build
	@echo "Waiting for services to be ready..."
	sleep 15
	@echo "Running database migrations..."
	$(COMPOSE_CMD) run --rm app go run cmd/migrate/main.go
	@echo "Seeding database..."
	$(COMPOSE_CMD) run --rm app go run cmd/seed/main.go
	@echo "Development environment is ready!"
	@echo "Application: http://localhost:8080"
	@echo "Admin: http://localhost:8080/admin"

.PHONY: quick-stop
quick-stop: ## Quick stop development environment
	$(COMPOSE_CMD) down

.PHONY: quick-restart
quick-restart: ## Quick restart development environment
	$(COMPOSE_CMD) restart

# Documentation targets
.PHONY: docs
docs: ## Generate documentation
	@echo "Documentation is available in the docs/ directory"
	@echo "API documentation: docs/API.md"
	@echo "User guide: docs/README.md"

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "Serving documentation at http://localhost:3000"
	cd docs && python3 -m http.server 3000

# Release targets
.PHONY: release
release: ## Create a new release
	@echo "Creating release for version $(VERSION)..."
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)

.PHONY: release-draft
release-draft: ## Create a draft release
	@echo "Creating draft release for version $(VERSION)..."
	git tag -a v$(VERSION) -m "Draft release version $(VERSION)"
	git push origin v$(VERSION)

# Security targets
.PHONY: security-scan
security-scan: ## Run security scan
	gosec ./...

.PHONY: security-audit
security-audit: ## Run security audit
	go list -json -deps ./... | nancy sleuth

# Performance targets
.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

.PHONY: profile
profile: ## Run profiling
	go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...

# Dependencies
.PHONY: deps
deps: ## Install dependencies
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

.PHONY: deps-check
deps-check: ## Check for outdated dependencies
	go list -u -m all

# Default target
.DEFAULT_GOAL := help