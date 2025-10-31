.PHONY: help build dev test clean migrate rollback seed install-services uninstall-services setup-nginx logs status restart

# Variables
BINARY_API=vip-panel-api
BINARY_WORKER=vip-panel-worker
BINARY_AGENT=vip-panel-agent
BINARY_CLI=vip-panel-cli

BUILD_DIR=./build
INSTALL_DIR=/opt/vip-panel
CONFIG_DIR=/etc/vip-panel
LOG_DIR=/var/log/vip-panel
DATA_DIR=/var/lib/vip-panel
SYSTEMD_DIR=/etc/systemd/system

GO=go
GOFLAGS=-ldflags="-s -w"
TEMPL=templ
NPM=npm

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

## help: Display this help message
help:
	@echo "VIP Hosting Panel - Make Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev              - Run in development mode with hot reload"
	@echo "  make build            - Build all binaries"
	@echo "  make test             - Run tests"
	@echo "  make clean            - Clean build artifacts"
	@echo ""
	@echo "Database:"
	@echo "  make migrate          - Run database migrations"
	@echo "  make rollback         - Rollback last migration"
	@echo "  make seed             - Seed database with test data"
	@echo ""
	@echo "Production (requires sudo):"
	@echo "  make install          - Full installation"
	@echo "  make install-services - Install systemd services"
	@echo "  make uninstall-services - Remove systemd services"
	@echo "  make setup-nginx      - Configure Nginx reverse proxy"
	@echo "  make status           - Check service status"
	@echo "  make logs             - Tail service logs"
	@echo "  make restart          - Restart all services"
	@echo ""

## build: Build all binaries
build: build-templ
	@echo "$(GREEN)Building binaries...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_API) ./cmd/api
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_WORKER) ./cmd/worker
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_AGENT) ./cmd/agent
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_CLI) ./cmd/cli
	@echo "$(GREEN)Build complete!$(NC)"

## build-templ: Generate Go code from Templ templates
build-templ:
	@echo "$(GREEN)Generating Templ templates...$(NC)"
	@$(TEMPL) generate

## build-css: Build Tailwind CSS
build-css:
	@echo "$(GREEN)Building CSS...$(NC)"
	@$(NPM) run build:css

## dev: Run in development mode with hot reload
dev: build-templ
	@echo "$(GREEN)Starting development server...$(NC)"
	@$(GO) run ./cmd/api/main.go &
	@$(GO) run ./cmd/worker/main.go &
	@$(NPM) run watch:css &
	@$(TEMPL) generate --watch &
	@wait

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Tests complete! Coverage report: coverage.html$(NC)"

## test-unit: Run unit tests only
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	@$(GO) test -v ./tests/unit/...

## test-integration: Run integration tests
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	@$(GO) test -v ./tests/integration/...

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf web/templates/*_templ.go
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete!$(NC)"

## migrate: Run database migrations
migrate:
	@echo "$(GREEN)Running database migrations...$(NC)"
	@$(GO) run ./cmd/cli/main.go migrate up

## rollback: Rollback last migration
rollback:
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@$(GO) run ./cmd/cli/main.go migrate down

## seed: Seed database with test data
seed:
	@echo "$(GREEN)Seeding database...$(NC)"
	@$(GO) run ./cmd/cli/main.go seed

## install: Full installation (run as sudo)
install: build
	@echo "$(GREEN)Installing VIP Hosting Panel...$(NC)"
	
	# Create directories
	@mkdir -p $(INSTALL_DIR)
	@mkdir -p $(CONFIG_DIR)
	@mkdir -p $(LOG_DIR)
	@mkdir -p $(DATA_DIR)
	@mkdir -p $(DATA_DIR)/backups
	@mkdir -p $(DATA_DIR)/uploads
	
	# Copy binaries
	@cp $(BUILD_DIR)/$(BINARY_API) $(INSTALL_DIR)/
	@cp $(BUILD_DIR)/$(BINARY_WORKER) $(INSTALL_DIR)/
	@cp $(BUILD_DIR)/$(BINARY_AGENT) $(INSTALL_DIR)/
	@cp $(BUILD_DIR)/$(BINARY_CLI) $(INSTALL_DIR)/
	
	# Copy configuration files
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		cp configs/config.yaml.example $(CONFIG_DIR)/config.yaml; \
		echo "$(YELLOW)Created config.yaml - Please edit $(CONFIG_DIR)/config.yaml$(NC)"; \
	fi
	@if [ ! -f $(CONFIG_DIR)/providers.yaml ]; then \
		cp configs/providers.yaml.example $(CONFIG_DIR)/providers.yaml; \
		echo "$(YELLOW)Created providers.yaml - Please edit $(CONFIG_DIR)/providers.yaml$(NC)"; \
	fi
	
	# Copy web assets
	@cp -r web/static $(INSTALL_DIR)/
	
	# Set permissions
	@chmod +x $(INSTALL_DIR)/$(BINARY_API)
	@chmod +x $(INSTALL_DIR)/$(BINARY_WORKER)
	@chmod +x $(INSTALL_DIR)/$(BINARY_AGENT)
	@chmod +x $(INSTALL_DIR)/$(BINARY_CLI)
	
	# Create vip-panel user if doesn't exist
	@id -u vip-panel &>/dev/null || useradd -r -s /bin/false vip-panel
	
	# Set ownership
	@chown -R vip-panel:vip-panel $(LOG_DIR)
	@chown -R vip-panel:vip-panel $(DATA_DIR)
	
	@echo "$(GREEN)Installation complete!$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "1. Edit configuration: $(CONFIG_DIR)/config.yaml"
	@echo "2. Run database migration: make migrate"
	@echo "3. Install services: make install-services"
	@echo "4. Setup Nginx: make setup-nginx"

## install-services: Install and enable systemd services
install-services:
	@echo "$(GREEN)Installing systemd services...$(NC)"
	
	# Copy service files
	@cp scripts/systemd/vip-panel-api.service $(SYSTEMD_DIR)/
	@cp scripts/systemd/vip-panel-worker.service $(SYSTEMD_DIR)/
	
	# Reload systemd
	@systemctl daemon-reload
	
	# Enable services
	@systemctl enable vip-panel-api
	@systemctl enable vip-panel-worker
	
	# Start services
	@systemctl start vip-panel-api
	@systemctl start vip-panel-worker
	
	@echo "$(GREEN)Services installed and started!$(NC)"
	@echo "Check status with: make status"

## uninstall-services: Stop and remove systemd services
uninstall-services:
	@echo "$(YELLOW)Removing systemd services...$(NC)"
	
	# Stop services
	@systemctl stop vip-panel-api || true
	@systemctl stop vip-panel-worker || true
	
	# Disable services
	@systemctl disable vip-panel-api || true
	@systemctl disable vip-panel-worker || true
	
	# Remove service files
	@rm -f $(SYSTEMD_DIR)/vip-panel-api.service
	@rm -f $(SYSTEMD_DIR)/vip-panel-worker.service
	
	# Reload systemd
	@systemctl daemon-reload
	
	@echo "$(GREEN)Services removed!$(NC)"

## setup-nginx: Configure Nginx as reverse proxy
setup-nginx:
	@echo "$(GREEN)Setting up Nginx...$(NC)"
	@bash scripts/setup-nginx.sh

## status: Check service status
status:
	@echo "$(GREEN)Service Status:$(NC)"
	@echo ""
	@systemctl status vip-panel-api --no-pager
	@echo ""
	@systemctl status vip-panel-worker --no-pager

## logs: Tail service logs
logs:
	@echo "$(GREEN)Tailing logs (Ctrl+C to exit)...$(NC)"
	@journalctl -u vip-panel-api -u vip-panel-worker -f

## logs-api: Tail API service logs only
logs-api:
	@journalctl -u vip-panel-api -f

## logs-worker: Tail worker service logs only
logs-worker:
	@journalctl -u vip-panel-worker -f

## restart: Restart all services
restart:
	@echo "$(GREEN)Restarting services...$(NC)"
	@systemctl restart vip-panel-api
	@systemctl restart vip-panel-worker
	@echo "$(GREEN)Services restarted!$(NC)"

## restart-api: Restart API service only
restart-api:
	@systemctl restart vip-panel-api
	@echo "$(GREEN)API service restarted!$(NC)"

## restart-worker: Restart worker service only
restart-worker:
	@systemctl restart vip-panel-worker
	@echo "$(GREEN)Worker service restarted!$(NC)"

## stop: Stop all services
stop:
	@echo "$(YELLOW)Stopping services...$(NC)"
	@systemctl stop vip-panel-api
	@systemctl stop vip-panel-worker
	@echo "$(GREEN)Services stopped!$(NC)"

## start: Start all services
start:
	@echo "$(GREEN)Starting services...$(NC)"
	@systemctl start vip-panel-api
	@systemctl start vip-panel-worker
	@echo "$(GREEN)Services started!$(NC)"

## update: Update and restart services (for production deployments)
update: build
	@echo "$(GREEN)Updating VIP Hosting Panel...$(NC)"
	
	# Stop services
	@systemctl stop vip-panel-api
	@systemctl stop vip-panel-worker
	
	# Backup old binaries
	@cp $(INSTALL_DIR)/$(BINARY_API) $(INSTALL_DIR)/$(BINARY_API).backup || true
	@cp $(INSTALL_DIR)/$(BINARY_WORKER) $(INSTALL_DIR)/$(BINARY_WORKER).backup || true
	
	# Copy new binaries
	@cp $(BUILD_DIR)/$(BINARY_API) $(INSTALL_DIR)/
	@cp $(BUILD_DIR)/$(BINARY_WORKER) $(INSTALL_DIR)/
	@cp $(BUILD_DIR)/$(BINARY_CLI) $(INSTALL_DIR)/
	
	# Update web assets
	@cp -r web/static $(INSTALL_DIR)/
	
	# Run migrations
	@$(INSTALL_DIR)/$(BINARY_CLI) migrate up
	
	# Start services
	@systemctl start vip-panel-api
	@systemctl start vip-panel-worker
	
	@echo "$(GREEN)Update complete!$(NC)"

## backup: Create backup of configuration and data
backup:
	@echo "$(GREEN)Creating backup...$(NC)"
	@mkdir -p backups
	@tar czf backups/vip-panel-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz \
		$(CONFIG_DIR) \
		$(DATA_DIR) \
		--exclude=$(DATA_DIR)/backups
	@echo "$(GREEN)Backup created in backups/ directory$(NC)"

## health: Check system health
health:
	@echo "$(GREEN)System Health Check:$(NC)"
	@echo ""
	@echo "Services:"
	@systemctl is-active vip-panel-api && echo "  ✓ API: Running" || echo "  ✗ API: Stopped"
	@systemctl is-active vip-panel-worker && echo "  ✓ Worker: Running" || echo "  ✗ Worker: Stopped"
	@echo ""
	@echo "Database:"
	@systemctl is-active postgresql && echo "  ✓ PostgreSQL: Running" || echo "  ✗ PostgreSQL: Stopped"
	@echo ""
	@echo "Cache:"
	@systemctl is-active redis && echo "  ✓ Redis: Running" || echo "  ✗ Redis: Stopped"
	@echo ""
	@echo "Web Server:"
	@systemctl is-active nginx && echo "  ✓ Nginx: Running" || echo "  ✗ Nginx: Stopped"

## lint: Run linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@golangci-lint run ./...

## format: Format Go code
format:
	@echo "$(GREEN)Formatting code...$(NC)"
	@gofmt -s -w .
	@goimports -w .

## deps: Update dependencies
deps:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@$(NPM) update

## version: Show version
version:
	@$(GO) run ./cmd/cli/main.go version
