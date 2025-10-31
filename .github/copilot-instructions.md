# Copilot Instructions for VIP Hosting Panel v2

## Project Overview
VIP Hosting Panel is a modern, multi-tenant web hosting control panel that centralizes provisioning, management, and billing across cost-efficient infrastructure providers (Hetzner, DigitalOcean, Vultr, AWS). It provides server management, site deployment, DNS/SSL automation, email services, monitoring, and reseller billing capabilities.

**Target Users**: Hosting providers, MSPs, agencies, resellers seeking lower costs and faster operations without cPanel/Plesk licensing overhead.

## Architecture & Technology Stack

### Backend (Go)
- **Framework**: Fiber (fast HTTP framework)
- **Language**: Go 1.21+
- **Authentication**: JWT with RBAC, 2FA (TOTP), optional SSO/OIDC
- **Job Queue**: Asynq (Redis-based) for async provisioning, backups, SSL renewal
- **Template Engine**: Templ (type-safe Go templates)

### Frontend (No React)
- **HTMX**: Hypermedia-driven interactions, server-side rendering
- **Alpine.js**: Minimal client-side reactivity for modals, dropdowns, forms
- **CSS**: Tailwind CSS + DaisyUI components
- **Real-time**: Server-Sent Events (SSE) for live status updates

### Data Layer
- **PostgreSQL 15+**: Core relational data (users, servers, sites, domains, billing, audit logs)
- **TimescaleDB**: Time-series metrics (CPU, RAM, disk, network, uptime)
- **Redis 7+**: Session cache, job queue, rate limiting

### Automation & Infrastructure
- **Ansible**: Server provisioning, configuration management, security hardening
- **Python Scripts**: Orchestration helpers for complex multi-step operations
- **SSH**: Agentless server management (optional lightweight agent for telemetry)

### Deployment (No Docker)
- **Direct Installation**: Systemd services on Ubuntu 22.04/24.04 or Debian 11/12
- **Services**: `vip-panel-api` (web server), `vip-panel-worker` (background jobs)
- **Reverse Proxy**: Nginx for SSL termination and static asset serving
- **Process Management**: systemd for service lifecycle

### Environment
- **Development**: Windows with Git Bash terminal
- **Production**: Linux (Ubuntu/Debian) with systemd
- **Repository**: GitHub with `master` as working branch

## Core Principles

### Real Data Fetching
- **Always fetch real data** from live sources (Hetzner API, server metrics, database queries, Stripe API)
- **Never use mock/placeholder data** in production code
- **Fallback pattern**: Display `N/A` or `--` when data is unavailable, with retry logic
- **Error handling**: Log failures, show user-friendly messages, continue attempting to fetch in background

### Data Fetching Patterns
```go
// Example: Server metrics with N/A fallback
func GetServerMetrics(serverID string) (*Metrics, error) {
    metrics, err := metricsRepo.FetchLatest(serverID)
    if err != nil {
        log.Warn("Failed to fetch metrics for server %s: %v", serverID, err)
        return &Metrics{
            CPU: "N/A",
            RAM: "N/A",
            Disk: "N/A",
            Status: "Unknown",
        }, nil // Return N/A values, don't fail the request
    }
    return metrics, nil
}
```

### No Docker Policy
- **Direct binary deployment**: Build Go binaries, deploy to `/usr/local/bin`
- **System packages**: Install PostgreSQL, Redis, Nginx via `apt`
- **Systemd services**: Manage lifecycle with `.service` files
- **Configuration**: YAML files in `/etc/vip-panel/` or local `configs/`
- **Logs**: System journal (`journalctl`) and application logs in `/var/log/vip-panel/`

### Frontend Architecture (HTMX + Alpine.js)
- **Server-rendered templates**: Use Templ for type-safe HTML generation
- **HTMX attributes**: `hx-get`, `hx-post`, `hx-swap` for dynamic content updates
- **Alpine.js**: Small components (modals, dropdowns, form validation)
- **No React, Vue, Angular**: Stick to hypermedia architecture
- **Progressive enhancement**: Forms work without JS, enhanced with HTMX

## Security & Compliance

### Security Features (Priority P0)
- **RBAC**: Superadmin → Admin → Reseller → Client hierarchy
- **Multi-tenant isolation**: Database-level tenant separation with middleware enforcement
- **Audit logging**: Immutable event trails for all privileged actions
- **2FA**: TOTP for all user roles, enforced at org level
- **WAF**: ModSecurity with OWASP CRS per site
- **Firewall**: UFW/iptables managed via API
- **Fail2ban**: SSH, FTP, Email, Panel login protection
- **SSL/TLS**: Let's Encrypt auto-renewal, custom cert support
- **Secrets**: Encrypted at rest in PostgreSQL, vaulted for sensitive config

### Security Patterns
- **Input validation**: Validate all user input (domains, SSH keys, emails)
- **SQL injection prevention**: Always use parameterized queries
- **XSS protection**: CSP headers, HTML escaping in templates
- **CSRF**: Token-based protection on all state-changing requests
- **Rate limiting**: Per-endpoint limits via middleware
- **SSH key management**: Rotate credentials, least-privilege access

### CodeQL Integration
- **Languages**: Go (primary), Python (automation scripts), JavaScript (Alpine.js)
- **Scans**: On push/PR to master/main, weekly scheduled (Mondays 16:44 UTC)
- **Manual trigger**: `gh workflow run "CodeQL Advanced"`

## Development Workflows

### Local Development
```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
npm install

# Setup local PostgreSQL and Redis (must be installed on system)
make setup-dev-db

# Run database migrations
make migrate

# Start API server (development mode with hot reload)
make dev-api

# Start worker (separate terminal)
make dev-worker

# Build Tailwind CSS
npm run build:css
```

### Building for Production
```bash
# Build Go binaries
make build

# Install systemd services
sudo make install-services

# Setup Nginx reverse proxy
sudo make setup-nginx

# Start services
sudo systemctl enable vip-panel-api vip-panel-worker
sudo systemctl start vip-panel-api vip-panel-worker
```

### Database Migrations
- **Tool**: golang-migrate or custom migration runner
- **Location**: `migrations/` directory with sequential numbered SQL files
- **Naming**: `001_initial_schema.sql`, `002_add_email_tables.sql`
- **Rollback**: Always provide `down` migrations for schema changes

### Ansible Automation
```bash
# Provision new server
ansible-playbook automation/playbooks/provision-server.yml \
  -i inventory/production.ini \
  --extra-vars "server_id=srv_xyz123"

# Deploy site
ansible-playbook automation/playbooks/deploy-site.yml \
  --extra-vars "site_id=site_abc456 domain=example.com"

# Install PHP version
ansible-playbook automation/playbooks/install-php.yml \
  --extra-vars "php_version=8.3"
```

## Key Architectural Patterns

### Multi-Tenancy
- **Tenant isolation**: Every table has `tenant_id` foreign key
- **Query scoping**: All queries filtered by current user's tenant context
- **Middleware**: `TenantMiddleware` extracts tenant from JWT, injects into request context
- **Reseller hierarchy**: Resellers can create sub-tenants (clients) with isolated billing/quotas

### Job Processing (Asynq)
- **Queue**: Redis-backed job queue for long-running tasks
- **Job types**: Server provisioning, site deployment, SSL renewal, backups, health checks
- **Retries**: Exponential backoff for transient failures
- **Monitoring**: Job status tracked in PostgreSQL, real-time updates via SSE

### Provider Integration
- **Abstraction**: `ProviderInterface` with implementations for Hetzner, DigitalOcean, Vultr, AWS
- **API clients**: Vendor SDKs or custom HTTP clients
- **Rate limiting**: Respect provider quotas, implement backoff
- **Error handling**: Graceful degradation when provider APIs fail

### Server Provisioning Flow
1. User creates server via UI (HTMX form submission)
2. API validates input, creates `Server` record with status `queued`
3. Job enqueued: `ProvisionServerJob{ServerID, ProviderConfig}`
4. Worker picks job, calls Ansible playbook via SSH
5. Playbook: create VM, harden security, install packages, configure firewall
6. Worker updates server status: `queued` → `provisioning` → `ready`
7. SSE pushes status updates to UI in real-time

### Site Deployment Flow
1. User creates site, selects template (WordPress, Laravel, Node.js, Static)
2. API creates `Site` record, enqueues `DeploySiteJob`
3. Worker runs Ansible playbook: configure Nginx vhost, PHP-FPM pool, create DB if needed
4. Git deployment: clone repo, install dependencies, build assets
5. Zero-downtime: symlink swap for blue-green deployments
6. SSL issuance: ACME challenge via Let's Encrypt, auto-renew cron

### DNS Management
- **Providers**: Cloudflare API, Route53, or panel-managed DNS (Bind9/PowerDNS)
- **Auto-provisioning**: When site is created, auto-create A/AAAA records
- **Validation**: Check DNS propagation before SSL issuance
- **DNSSEC**: Optional for advanced users

### Monitoring & Alerts
- **Metrics collection**: Lightweight agent or SSH-based exporters (node_exporter)
- **Storage**: TimescaleDB for time-series data (retention policies for cost control)
- **Uptime checks**: HTTP/HTTPS pings from control plane to managed servers
- **Alerting**: Email/webhook when CPU > 80%, disk > 85%, service down, SSL expiring < 7 days

### Billing Integration
- **Stripe**: Subscriptions for hosting plans, usage-based metering for overage
- **Invoicing**: Auto-generated monthly invoices with PDF export
- **Reseller margins**: Configurable markup percentages per reseller
- **Cost tracking**: Provider costs (Hetzner VM charges) vs. revenue per tenant

## File Structure & Conventions

### Go Code Organization
- `cmd/`: Entrypoints (`api/main.go`, `worker/main.go`, `cli/main.go`)
- `internal/`: Private application code (handlers, services, models, jobs)
- `internal/handlers/`: HTTP request handlers (one file per resource: `servers.go`, `sites.go`)
- `internal/services/`: Business logic (provider clients, deployment logic)
- `internal/models/`: Database models (GORM or sqlc)
- `internal/jobs/`: Asynq background job definitions
- `internal/middleware/`: Authentication, RBAC, rate limiting
- `internal/repository/`: Data access layer (SQL queries)

### Frontend Templates (Templ)
- `web/templates/layouts/`: Base layouts (`base.templ`, `dashboard.templ`)
- `web/templates/pages/`: Full page templates (`servers/list.templ`, `sites/create.templ`)
- `web/templates/components/`: Reusable components (`navbar.templ`, `server_card.templ`)
- **Naming**: Use `.templ` extension, compile to `.go` files with `templ generate`

### Ansible Playbooks
- `automation/playbooks/`: Task-specific playbooks (one per major operation)
- `automation/roles/`: Reusable roles (`common`, `webserver`, `database`, `security`)
- `automation/templates/`: Jinja2 templates for config files (Nginx vhosts, PHP.ini)
- `automation/scripts/`: Python orchestration helpers

### Configuration Files
- `configs/config.yaml`: Main app config (server, database, Redis, JWT, features)
- `configs/providers.yaml`: API tokens for Hetzner, DigitalOcean, Cloudflare, Stripe
- `configs/security.yaml`: Firewall rules, WAF settings, SSL policies
- **Never commit**: Use `.example` files, load secrets from env vars or vault

## Data Fetching Best Practices

### Always Use Real Data
- **Provider APIs**: Fetch server list, costs, regions from Hetzner/DO APIs
- **Server metrics**: Query TimescaleDB for CPU/RAM/disk from agent or SSH
- **Site status**: Check Nginx access logs, PHP-FPM status, database connections
- **Billing**: Fetch invoices, payments, usage from Stripe API and local DB

### N/A Fallback Pattern
```go
// Template rendering with N/A fallback
type ServerCardData struct {
    Name   string
    CPU    string // "45%" or "N/A"
    RAM    string // "2.1 GB / 4 GB" or "N/A"
    Status string // "Ready", "Provisioning", "Unknown"
}

// In handler
func (h *ServerHandler) GetServerCard(c *fiber.Ctx) error {
    server := h.repo.FindByID(serverID)

    metrics, err := h.metrics.GetLatest(serverID)
    cardData := ServerCardData{
        Name:   server.Name,
        CPU:    "N/A",
        RAM:    "N/A",
        Status: "Unknown",
    }

    if err == nil && metrics != nil {
        cardData.CPU = fmt.Sprintf("%.1f%%", metrics.CPUPercent)
        cardData.RAM = fmt.Sprintf("%.1f GB / %.1f GB", metrics.UsedRAM, metrics.TotalRAM)
        cardData.Status = metrics.Status
    } else {
        // Log error but don't fail the request
        log.Warn("Metrics unavailable for server %s: %v", serverID, err)

        // Schedule background retry
        h.jobs.Enqueue(&RefreshMetricsJob{ServerID: serverID})
    }

    return h.render(c, components.ServerCard(cardData))
}
```

### Background Refresh Strategy
- **Initial load**: Show N/A if data not immediately available
- **Polling**: HTMX `hx-trigger="every 30s"` to refresh metrics
- **SSE**: Push real-time updates when metrics become available
- **Retry queue**: Failed API calls enqueued for retry with exponential backoff

## UI/UX Patterns (HTMX + Alpine.js)

### Server-Side Rendering
```html
<!-- templates/pages/servers/list.templ -->
<div class="servers-grid" hx-get="/api/servers/cards" hx-trigger="every 30s">
  @for _, server := range servers {
    @components.ServerCard(server)
  }
</div>
```

### Dynamic Modals (Alpine.js)
```html
<!-- Modal component -->
<div x-data="{ open: false }" x-show="open" x-cloak>
  <button @click="open = true">Create Server</button>
  <div class="modal" x-show="open" @click.outside="open = false">
    <form hx-post="/api/servers" hx-swap="outerHTML">
      <!-- form fields -->
    </form>
  </div>
</div>
```

### Form Validation
- **Client-side**: Alpine.js for instant feedback
- **Server-side**: Always validate in Go handlers (never trust client)
- **Error display**: HTMX swaps error messages into DOM on validation failure

### Status Indicators
- **Color-coded chips**: `Ready` (green), `Provisioning` (yellow), `Failed` (red), `Unknown` (gray)
- **Real-time updates**: SSE pushes status changes, HTMX swaps chip content
- **Tooltips**: Show last updated timestamp on hover

## Security Checklist for Every Feature

- [ ] Input validation (whitelist, length limits, regex patterns)
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS protection (HTML escaping in templates)
- [ ] CSRF tokens on state-changing forms
- [ ] RBAC enforcement (check user role in handler)
- [ ] Audit logging (who, what, when for privileged actions)
- [ ] Rate limiting (per-user, per-IP)
- [ ] Secrets encryption (never log API tokens, passwords)
- [ ] HTTPS enforcement (redirect HTTP to HTTPS)
- [ ] SSH key validation (format, length, type)

## Testing Strategy

### Unit Tests
- **Go**: `*_test.go` files for services, models, utilities
- **Coverage target**: >70% for business logic
- **Mocking**: Use interfaces for provider clients, repositories

### Integration Tests
- **Database**: Test queries against real PostgreSQL (test database)
- **Jobs**: Test Asynq job execution with Redis test instance
- **Ansible**: Test playbooks against staging VMs (Vagrant or Hetzner test servers)

### E2E Tests
- **Selenium/Playwright**: Test critical flows (create server → deploy site → SSL issuance)
- **API tests**: Postman/Hurl for REST endpoints

## Performance Targets (from PRD)

- **API latency**: p95 < 300ms
- **Job queue wait time**: p95 < 30s
- **Time-to-provision**: Server + site + SSL < 8 minutes
- **Uptime**: Control plane ≥ 99.9%, agent connectivity ≥ 99.5%
- **Provisioning success rate**: ≥ 95%
- **SSL issuance success**: ≥ 98%

## Critical Files & Paths

### Configuration
- [configs/config.yaml.example](configs/config.yaml.example) - Main app config template
- [configs/providers.yaml.example](configs/providers.yaml.example) - Provider API tokens
- `.env` - Local environment overrides (not committed)

### Core Backend
- [cmd/api/main.go](cmd/api/main.go) - API server entrypoint
- [cmd/worker/main.go](cmd/worker/main.go) - Background job processor
- [internal/handlers/](internal/handlers/) - HTTP request handlers
- [internal/services/](internal/services/) - Business logic layer
- [internal/jobs/](internal/jobs/) - Background job definitions

### Frontend
- [web/templates/](web/templates/) - Templ HTML templates
- [web/static/css/tailwind.css](web/static/css/tailwind.css) - Tailwind source
- [web/static/js/](web/static/js/) - Alpine.js and HTMX scripts

### Automation
- [automation/playbooks/](automation/playbooks/) - Ansible playbooks
- [automation/scripts/](automation/scripts/) - Python orchestration scripts
- [automation/templates/](automation/templates/) - Config file templates (Nginx, PHP, etc.)

### Database
- [migrations/](migrations/) - SQL migration files
- [internal/models/](internal/models/) - Database models

### Security
- [.gitignore](.gitignore) - Never commit secrets, credentials, API keys
- [SECURITY.md](SECURITY.md) - Vulnerability reporting policy
- [.github/workflows/codeql.yml](.github/workflows/codeql.yml) - CodeQL security scanning

## Common Tasks & Commands

### Development
```bash
# Start dev environment
make dev

# Run migrations
make migrate

# Generate Templ templates
templ generate

# Build Tailwind CSS
npm run build:css

# Run tests
make test

# Format code
gofmt -w .
```

### Deployment
```bash
# Build production binaries
make build

# Install on server (run on target server)
sudo bash scripts/install.sh

# Manual service management
sudo systemctl status vip-panel-api
sudo systemctl restart vip-panel-worker
sudo journalctl -u vip-panel-api -f
```

### Debugging
```bash
# View API logs
sudo journalctl -u vip-panel-api -f --since "10 minutes ago"

# Check Redis queue
redis-cli LLEN asynq:queues:default

# Test Ansible playbook
ansible-playbook automation/playbooks/provision-server.yml --check --diff
```

## Git Workflow

### Branch Strategy
- **master**: Working development branch (active development)
- **main**: Default branch (stable releases)
- **feature branches**: `feature/server-resize`, `feature/dns-cloudflare`
- **PRs**: Require CodeQL pass, manual review

### Commit Messages
- Format: `<type>: <description>` (e.g., `feat: add Hetzner server provisioning`)
- Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `security`

### Security Workflow
```bash
# Trigger manual CodeQL scan
gh workflow run "CodeQL Advanced"

# Check for secrets before commit (use pre-commit hook)
git secrets --scan
```

## When Adding New Features

1. **Check PRD alignment**: Ensure feature is in scope (P0/P1/P2 priority)
2. **Update CodeQL matrix**: If new language/framework introduced
3. **Add migrations**: Database schema changes go in `migrations/`
4. **Write tests**: Unit + integration for new services
5. **Update Ansible**: If feature requires server-side changes
6. **Real data**: Always fetch from real APIs, fallback to N/A
7. **Security review**: Complete checklist above
8. **Documentation**: Update API docs, architecture diagrams

## Resources & References

- **PRD**: [project-prd.md](project-prd.md) - Product requirements and success metrics
- **README**: [README.md](README.md) - Technical architecture and feature list
- **HTMX Docs**: https://htmx.org/docs/
- **Templ Guide**: https://templ.guide/
- **Fiber Docs**: https://docs.gofiber.io/
- **Ansible Best Practices**: https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html