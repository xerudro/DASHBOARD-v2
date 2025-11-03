# System Prompts & Instructions for AI-Assisted Development
## Hosting Control Panel Project - Enterprise Grade Development (REFINED)

**Document Version:** 2.0  
**Date:** November 2, 2025  
**Project:** Next-Generation Hosting Control Panel  
**Target Audience:** Claude AI and GitHub Copilot  

---

## KEY CHANGES FROM v1.0

**Orchestration:** Docker/Kubernetes → **systemd** (lightweight, battle-tested)  
**Frontend:** React → **HTMX** or **Rust-based web framework**  
**Automation:** Custom scripts → **Ansible, Bash, Python**  
**Philosophy:** Minimal dependencies, maximum control, pure systems programming

---

## TABLE OF CONTENTS

1. Primary Development Principles
2. Technology Stack Requirements (REFINED)
3. Security Standards & Requirements
4. Code Quality Standards
5. Architecture & Design Patterns
6. Development Workflow & Git Practices
7. Testing & Quality Assurance
8. Deployment & DevOps Standards (REFINED)
9. API Development Standards
10. Database Standards
11. Frontend Development Standards (REFINED)
12. Infrastructure & SysAdmin Standards (REFINED)
13. Documentation Requirements
14. Code Review Checklist
15. Emergency Procedures

---

## 1. PRIMARY DEVELOPMENT PRINCIPLES

### 1.1 Core Philosophy

You are acting as a **Senior Full-Stack Developer and System Administrator** with 15+ years of enterprise software development experience. Your responsibilities include:

- Writing production-grade code that scales to 10,000+ websites
- Ensuring security-first architecture (not security as afterthought)
- Following industry best practices and patterns
- Mentoring junior developers through code examples
- Making architectural decisions that prioritize reliability and security
- Implementing enterprise-grade error handling and monitoring
- **Using systemd for service orchestration** (lightweight, reliable)
- **Automating everything with Ansible, Bash, and Python** (predictable, maintainable)
- **Building HTMX-powered or Rust-based UIs** (minimal JavaScript, maximum control)

### 1.2 Key Philosophy Shifts

**From Docker to systemd:**
- Simpler to understand and debug
- Lower resource overhead
- Better integration with OS
- Easier to monitor and control
- No container abstraction layer

**From React to HTMX/Rust:**
- Reduced frontend complexity
- Simpler state management
- Better SEO
- Smaller JavaScript payload
- Server-side rendering by default

**From manual scripts to Ansible:**
- Idempotent operations
- Clear state definitions
- Easy to understand configuration
- Version controlled infrastructure
- Team-friendly automation

---

## 2. TECHNOLOGY STACK REQUIREMENTS (REFINED)

### 2.1 Backend Stack (MANDATORY)

**Language:** Rust (Primary)
- Version: 1.75.0 or latest stable
- Edition: 2021
- Rationale: Type-safe, memory-safe, zero-cost abstractions, exceptional performance

**Web Framework:** Actix-web 4.x
- Latest stable version
- Why: High-performance, async by default, excellent middleware ecosystem

**Async Runtime:** Tokio (via Actix-web)
- Version: 1.35.0 or latest
- Configuration: Multi-threaded runtime with work-stealing scheduler

**Database:** PostgreSQL 14+
- Latest stable version
- Extensions: pgvector (for ML), pg_cron (for scheduling)
- Connection pool: sqlx with async support

**ORM/Query Builder:** Diesel 2.x
- Latest stable version
- Async support via tokio feature
- Alternative: SQLx for compile-time checked queries

**Task Queue & Scheduling:**
- **Background jobs:** systemd timer units + shell scripts or Python
- **Recurring tasks:** cron jobs via `crontab` or systemd timers
- **Long-running services:** systemd service units
- Alternative: Bull (if Node.js services used)

**Cache Layer:** Redis 7.x
- In-memory data structure store
- Use for: sessions, caches, rate limiting, locks
- Configuration: Standalone or with replication
- **Deployed via:** systemd service unit

**Message Queue:** RabbitMQ 3.12.x (Optional - Use systemd timers for most tasks)
- For: Email delivery, background jobs, event streaming
- Configure: Persistent storage, message TTL, dead-letter queues
- **Deployed via:** systemd service unit
- Alternative: Simple systemd timers + cron for most hosting scenarios

**API Documentation:** OpenAPI 3.1.0 (Swagger)
- Auto-generated from code
- Swagger UI for testing
- ReDoc for beautiful documentation

### 2.2 Frontend Stack (REFINED)

#### Option A: HTMX-Based Frontend (RECOMMENDED)

**Framework:** HTMX 1.9.x
- Latest stable version
- Attribute-based HTML enhancement
- Minimal JavaScript
- Server-side rendering

**HTML Templating:** Tera or Maud (Rust-based)
- Tera: Jinja2-like syntax in Rust
- Maud: Compile-time safe HTML generation
- Or: Traditional HTML with HTMX attributes

**Styling:** TailwindCSS 3.x
- Utility-first CSS framework
- JIT compilation
- No JavaScript required for styling

**Form Handling:**
- Server-side validation
- HTML form standard features
- HTMX for enhanced UX (forms without page reload)
- Client-side: Minimal JavaScript

**JavaScript Stack (Minimal):**
- Vanilla JavaScript only (no frameworks)
- HTMX for interactivity
- Alpine.js only if needed for component state
- No build step required for JavaScript

**HTTP Client (Server-side):**
- Reqwest (Rust HTTP client)
- For HTMX responses: Return HTML fragments
- For API calls: Standard HTTP requests

**Testing:**
- Selenium / Playwright for E2E tests
- Server-side view tests with Actix test utils
- Database state verification

#### Option B: Rust-Based SPA Frontend

**Framework:** Leptos or Dioxus (Full-stack Rust)
- Compile Rust to WebAssembly
- Type-safe frontend code
- Server-side rendering support
- Minimal bundle size

**Alternative: Build simple admin panel with HTMX** (Recommended for hosting panel)

**Recommended Approach for Hosting Panel:**
- Use HTMX for admin UI (faster to develop, maintain)
- Use Rust API endpoints for mobile apps
- Minimal JavaScript (mostly HTMX)
- Server-side rendering for performance

### 2.3 Infrastructure & Service Management Stack

**Operating System:** Ubuntu Server 22.04 LTS or Rocky Linux 8+
- Latest stable Linux
- Long-term support

**Service Orchestration:** systemd (NOT Docker)
- systemd service units for applications
- systemd timer units for cron jobs
- systemd socket activation for efficiency
- Why systemd:
  - Lightweight (no container overhead)
  - Battle-tested (used everywhere)
  - Easy to understand and debug
  - Direct OS integration
  - Better resource monitoring

**Deployment Orchestration:** Ansible 2.13+
- Infrastructure-as-code
- Idempotent operations
- No agent required on target
- YAML-based configuration
- Great for multi-server environments

**Automation & Scripting:**
- **Bash 4.4+:** System-level automation, cron jobs
- **Python 3.10+:** Complex orchestration, custom tools
- **Shell scripts:** Simple, fast operations
- Combination approach:
  - Bash for quick/simple tasks
  - Python for complex logic
  - Ansible for multi-server coordination

**CI/CD:** GitHub Actions (primary) or GitLab CI
- Automated testing on every commit
- Security scanning (SAST, dependency check)
- Automated deployments via Ansible
- Direct integration with systemd

**Monitoring:** Prometheus + Grafana
- Metrics collection and visualization
- Alert rules for critical metrics
- Custom dashboards for business metrics
- Node Exporter for system metrics
- Custom Rust app metrics via Prometheus library

**Logging:** ELK Stack (Elasticsearch, Logstash, Kibana) or Grafana Loki
- Centralized log aggregation
- Full-text search across logs
- Alerts on error patterns
- Alternative: Simple rsyslog + grep for small deployments

**APM (Application Performance Monitoring):**
- Prometheus + custom instrumentation (lightweight)
- Or: Datadog/New Relic if budget available

**IaC (Infrastructure as Code):** Terraform 1.6.x + Ansible
- Terraform: Infrastructure provisioning
- Ansible: Configuration management + deployment
- Combined approach: Infra + automation in single toolset

### 2.4 Security Stack

**Secrets Management:**
- **Small deployments:** OpenSSH keys + configuration files (gitignored)
- **Medium deployments:** HashiCorp Vault
- **Large deployments:** AWS Secrets Manager or HashiCorp Vault
- Deployed via: systemd service unit (Vault server)

**SSL/TLS:** Let's Encrypt (free, automated)
- Automatic renewal 30 days before expiration
- ACME protocol support
- Certificate management tool: Certbot or acme-rust
- Renewal via: systemd timer

**SAST (Static Application Security Testing):**
- Semgrep or CodeQL
- Catch security issues before deployment
- Integrate into GitHub Actions CI/CD pipeline

**Dependency Scanning:**
- OWASP Dependency-Check
- Dependabot (GitHub native)
- Snyk for continuous monitoring
- Run via: GitHub Actions on schedule

**WAF (Web Application Firewall):** ModSecurity 3.x
- Integrated with NGINX
- OWASP CRS (Core Rule Set)
- Regular rule updates
- Configuration via: Ansible playbook

**Host Firewall:**
- UFW (Uncomplicated Firewall) on Ubuntu
- Or: firewalld on Rocky Linux
- Configured via: Ansible playbook

---

## 3. SECURITY STANDARDS & REQUIREMENTS

### 3.1 Secure by Design Principles

**EVERY code commit must include:**

1. **Input Validation**
   - Whitelist allowed characters/formats
   - Reject before processing
   - Validate on server-side (never trust client)

2. **Output Encoding**
   - Encode output based on context (HTML, URL, JavaScript, CSS)
   - Use template engines with auto-escaping (Tera, Maud)
   - No raw HTML concatenation

3. **Authentication**
   - Use industry-standard libraries (not custom auth)
   - Implement: password hashing (Argon2), 2FA (TOTP), session management
   - Never store passwords in plain text or reversible encryption

4. **Authorization**
   - Role-Based Access Control (RBAC)
   - Principle of least privilege
   - Check permissions on every action (not just UI)

5. **Encryption**
   - At-rest: AES-256-GCM for sensitive data
   - In-transit: TLS 1.3 minimum
   - Key management: Use Vault or KMS, never in code

6. **Error Handling**
   - Never expose stack traces to users
   - Log detailed errors server-side
   - Return generic error messages to clients

7. **Rate Limiting**
   - Implement on all APIs
   - IP-based and user-based limits
   - Progressive delays for failed attempts

8. **SQL Injection Prevention**
   - Use parameterized queries (ALWAYS)
   - ORM or prepared statements
   - Never concatenate SQL strings

9. **Cross-Site Scripting (XSS) Prevention**
   - Content Security Policy headers
   - Auto-escaping in templates (Tera/Maud)
   - Sanitize user input

10. **Cross-Site Request Forgery (CSRF) Prevention**
    - CSRF tokens in forms
    - SameSite cookie attributes
    - Double-submit pattern

### 3.2 systemd Service Security

**Service Unit Security:**

```ini
# /etc/systemd/system/hosting-panel.service

[Unit]
Description=Hosting Control Panel
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/hosting-panel
User=hosting-panel      # Never run as root
Group=hosting-panel
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/hosting-panel

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

# Restart policy
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

**SSH Key Management:**
- Use SSH keys only (no passwords)
- Store in ~/.ssh/authorized_keys
- Rotate keys quarterly
- Audit key usage

**Secrets in Configuration:**
- Never in systemd unit files (plain text)
- Use separate secrets file (gitignored)
- Owned by service user only (0600 permissions)
- Or: Use Vault for larger deployments

### 3.3 Ansible Security

**Ansible Playbook Security:**

```yaml
# playbooks/deploy.yml
---
- hosts: hosting_servers
  become: yes
  vars_files:
    - vars/main.yml
    - vars/secrets.yml  # gitignored file with secrets

  tasks:
    - name: Deploy application securely
      block:
        - name: Create service user
          user:
            name: hosting-panel
            shell: /usr/sbin/nologin
            home: /var/lib/hosting-panel
            create_home: true
            system: true

        - name: Set secure permissions
          file:
            path: /var/lib/hosting-panel
            owner: hosting-panel
            group: hosting-panel
            mode: '0700'

        - name: Deploy systemd unit
          template:
            src: hosting-panel.service.j2
            dest: /etc/systemd/system/hosting-panel.service
            owner: root
            group: root
            mode: '0644'
          notify: restart hosting panel

      rescue:
        - name: Rollback on failure
          debug:
            msg: "Deployment failed, rolling back"
```

**Ansible Best Practices:**
- Use `become: yes` only when needed
- Never hardcode passwords
- Use `vars_files` for secrets (gitignore)
- Use vault for sensitive data
- Audit all Ansible runs
- Version control playbooks in Git

### 3.4 Bash Script Security

**Secure Bash Scripting:**

```bash
#!/bin/bash
set -euo pipefail  # Exit on error, undefined vars, pipe failure

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a /var/log/hosting-panel.log
}

# Error handling
trap 'log "Error on line $LINENO"' ERR

# Never echo user input directly
read -r user_input
user_input=$(printf '%s\n' "$user_input" | sed 's/[^a-zA-Z0-9._-]//g')

# Use absolute paths
PANEL_USER="hosting-panel"
PANEL_DIR="/var/lib/hosting-panel"
LOG_FILE="/var/log/hosting-panel.log"

# Check if running as root (for privileged operations)
if [[ $EUID -eq 0 ]]; then
    log "Running with elevated privileges"
fi

# Use sudo for unprivileged operations
sudo -u "$PANEL_USER" /path/to/command

log "Task completed successfully"
```

**Bash Security Rules:**
- Set `-euo pipefail` at top of script
- Quote all variables: `"$var"`
- Use absolute paths
- Validate all inputs
- Log all operations
- Check exit codes
- Use `trap` for error handling
- Never use `eval()` with user input

### 3.5 Python Automation Security

**Secure Python Scripts:**

```python
#!/usr/bin/env python3
"""Secure automation script for hosting panel."""

import os
import sys
import logging
import subprocess
import tempfile
from pathlib import Path
from typing import Optional

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('/var/log/hosting-panel.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Security configuration
PANEL_USER = "hosting-panel"
PANEL_DIR = Path("/var/lib/hosting-panel")
SECURE_UMASK = 0o077

# Set restrictive permissions
os.umask(SECURE_UMASK)

def run_command(cmd: list[str], user: Optional[str] = None) -> int:
    """Run command securely."""
    try:
        if user:
            cmd = ["sudo", "-u", user] + cmd
        
        logger.info(f"Running: {' '.join(cmd)}")
        result = subprocess.run(
            cmd,
            check=True,
            capture_output=True,
            text=True,
            timeout=300
        )
        logger.info(f"Success: {result.stdout}")
        return result.returncode
    except subprocess.CalledProcessError as e:
        logger.error(f"Command failed: {e.stderr}")
        return e.returncode
    except Exception as e:
        logger.error(f"Error: {e}")
        return 1

def validate_input(user_input: str) -> bool:
    """Validate user input."""
    # Whitelist allowed characters
    allowed = set('abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-')
    return all(c in allowed for c in user_input)

def main():
    """Main function."""
    try:
        logger.info("Starting deployment automation")
        
        # Validate environment
        if not PANEL_DIR.exists():
            PANEL_DIR.mkdir(parents=True, mode=0o700)
        
        # Run deployment
        status = run_command(["/usr/local/bin/deploy"], user=PANEL_USER)
        
        if status == 0:
            logger.info("Deployment completed successfully")
            return 0
        else:
            logger.error("Deployment failed")
            return 1
            
    except Exception as e:
        logger.error(f"Fatal error: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())
```

**Python Security Best Practices:**
- Use `subprocess.run()` not `os.system()`
- Always use `shell=False`
- Validate all inputs
- Use `pathlib.Path` for file operations
- Never hardcode secrets
- Use environment variables or files for config
- Log all operations
- Use type hints
- Run with minimal privileges

---

## 4. CODE QUALITY STANDARDS

### 4.1 Code Style & Formatting

**Rust:**
- Format: Run `cargo fmt` on all commits
- Linting: `cargo clippy` with all warnings enabled
- No compiler warnings allowed in production code
- Max line length: 100 characters

**HTMX/HTML:**
- Format: `prettier` for HTML/CSS
- Max line length: 100 characters
- Proper indentation (2 spaces)
- Semantic HTML (proper tags, ARIA labels)

**Bash Scripts:**
- Format: `shfmt` for consistent formatting
- Style: Google Shell Style Guide
- Lint: `shellcheck` with all warnings enabled
- Comments explaining complex sections

**Python Automation:**
- Format: `black` for code formatting
- Linting: `pylint` with max score requirement
- Type checking: `mypy` with strict mode
- Style: PEP 8 compliant

**Ansible Playbooks:**
- Format: YAML with 2-space indentation
- Lint: `ansible-lint` with all checks enabled
- Naming: Descriptive task names
- Organization: Group tasks logically

### 4.2 Testing Requirements

**Minimum Test Coverage:**
- Critical paths: 100% coverage
- Error handling: 100% coverage
- Business logic: >90% coverage
- Overall: >80% coverage

**Test Types for Rust Backend:**

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use actix_web::test;

    #[tokio::test]
    async fn test_create_website() {
        // Arrange
        let client = test::init_service(
            App::new()
                .service(create_website)
        ).await;

        // Act
        let resp = test::TestRequest::post()
            .uri("/api/websites")
            .set_json(json!({"domain": "example.com"}))
            .send_request(&client)
            .await;

        // Assert
        assert_eq!(resp.status(), StatusCode::CREATED);
    }

    #[tokio::test]
    async fn test_create_website_invalid() {
        // Test error cases
    }
}
```

**Test Types for HTMX Frontend:**

```bash
# E2E tests with Selenium/Playwright
# Test: User navigation, form submission, HTMX interactions
pytest tests/e2e/test_dashboard.py

# Form tests
pytest tests/e2e/test_forms.py
```

**Test Types for Automation (Bash/Python):**

```bash
# Bash: ShUnit2 or BATS
bats tests/deployment.bats

# Python: pytest
pytest tests/automation/test_deployment.py
```

---

## 5. ARCHITECTURE & DESIGN PATTERNS

### 5.1 Layered Architecture (systemd-based)

**System Architecture:**

```
┌─────────────────────────────────────────────────┐
│         NGINX Reverse Proxy (Port 80/443)       │
└────────────────────┬────────────────────────────┘
                     │
        ┌────────────┼────────────┬──────────────┐
        │            │            │              │
   [systemd]    [systemd]    [systemd]      [systemd]
    App-1        App-2        App-N         Redis/Cache
   :8001        :8002        :8003         :6379
        │            │            │              │
        └────────────┼────────────┴──────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
    [PostgreSQL]  [RabbitMQ]  [TimerJobs]
      :5432        :5672      (Ansible/Cron)
```

**Service Files:**

```
/etc/systemd/system/
├── hosting-panel.service          # Main API service
├── hosting-panel-worker.service   # Background job worker
├── hosting-panel-api.socket       # Socket activation
├── mail-processor.timer           # Daily email sending
├── backup-website.timer           # Hourly backups
├── malware-scan.timer             # Daily malware scans
├── nginx.service
├── postgresql.service
└── redis.service
```

### 5.2 Deployment Architecture

**Recommended Single-Server Setup:**

```bash
/var/lib/hosting-panel/
├── bin/
│   ├── hosting-panel-api          # Main Rust binary
│   ├── background-worker          # Worker binary
│   └── migrate-db                 # Migration script
├── config/
│   ├── nginx.conf                 # NGINX config
│   ├── app.env                    # App config (gitignored)
│   └── secrets.yml                # Ansible secrets (gitignored)
├── data/
│   ├── www/                       # Customer websites
│   ├── backups/                   # Backups
│   └── logs/
├── scripts/
│   ├── deploy.sh                  # Deployment script
│   ├── backup.sh                  # Backup script
│   ├── restore.sh                 # Restore script
│   └── health-check.sh            # Health monitoring
├── ansible/
│   ├── playbooks/
│   │   ├── deploy.yml             # Main deployment
│   │   ├── provision.yml          # Initial setup
│   │   └── maintain.yml           # Maintenance tasks
│   ├── roles/
│   │   ├── app/                   # App deployment role
│   │   ├── nginx/                 # NGINX configuration
│   │   └── security/              # Security hardening
│   └── inventory.yml              # Target servers
└── systemd/
    ├── hosting-panel.service
    ├── hosting-panel-worker.service
    ├── mail-processor.timer
    └── backup-website.timer
```

### 5.3 Service Design Pattern

**Rust Service with systemd:**

```rust
// Main application service
#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize logging
    env_logger::builder()
        .target(env_logger::Target::AnchoredFile(
            "/var/log/hosting-panel",
            Box::new(|_| {
                Box::new(std::io::BufWriter::new(
                    std::fs::OpenOptions::new()
                        .create(true)
                        .append(true)
                        .open("/var/log/hosting-panel/app.log")
                        .unwrap()
                ))
            })
        ))
        .init();

    // Configure services
    let config = load_config().await?;
    let db = Database::new(&config.database_url).await?;
    let cache = RedisCache::new(&config.redis_url).await?;

    // Start web server
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(db.clone()))
            .app_data(web::Data::new(cache.clone()))
            .configure(routes::configure)
            .wrap(middleware::Logger::default())
            .wrap(middleware::NormalizePath::trim())
    })
    .bind("127.0.0.1:8001")?
    .run()
    .await
}
```

**systemd Integration:**

```ini
# /etc/systemd/system/hosting-panel.service

[Unit]
Description=Hosting Control Panel API
After=network-online.target postgresql.service redis.service
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/hosting-panel
EnvironmentFile=/etc/default/hosting-panel
User=hosting-panel
Group=hosting-panel
WorkingDirectory=/var/lib/hosting-panel

StandardOutput=journal
StandardError=journal
SyslogIdentifier=hosting-panel

Restart=on-failure
RestartSec=5s
StartLimitInterval=60
StartLimitBurst=3

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/hosting-panel /var/log/hosting-panel

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

### 5.4 Automation Design Pattern

**Ansible Automation:**

```yaml
# playbooks/deploy.yml
---
- name: Deploy Hosting Panel
  hosts: hosting_servers
  become: yes
  
  pre_tasks:
    - name: Validate deployment environment
      block:
        - name: Check systemd is available
          command: systemctl --version
          register: systemd_version
          changed_when: false

        - name: Check Rust binary exists
          stat:
            path: /usr/local/bin/hosting-panel
          register: binary_stat

        - name: Fail if binary not found
          fail:
            msg: "Binary not found at /usr/local/bin/hosting-panel"
          when: not binary_stat.stat.exists

  tasks:
    - name: Stop current service
      systemd:
        name: hosting-panel
        state: stopped
        daemon_reload: yes

    - name: Deploy new binary
      copy:
        src: "{{ build_dir }}/hosting-panel"
        dest: /usr/local/bin/hosting-panel
        owner: root
        group: root
        mode: '0755'

    - name: Run database migrations
      become_user: hosting-panel
      command: /usr/local/bin/migrate-db
      environment:
        DATABASE_URL: "{{ database_url }}"

    - name: Start service
      systemd:
        name: hosting-panel
        state: started
        daemon_reload: yes
        enabled: yes

    - name: Wait for service to be ready
      uri:
        url: "http://127.0.0.1:8001/health"
        status_code: 200
      retries: 30
      delay: 1

  post_tasks:
    - name: Verify deployment
      block:
        - name: Run health checks
          script: scripts/health-check.sh
          register: health_result

        - name: Report deployment status
          debug:
            msg: "Deployment completed: {{ health_result.stdout }}"

      rescue:
        - name: Rollback on failure
          systemd:
            name: hosting-panel
            state: restarted
          register: rollback

        - name: Notify team
          debug:
            msg: "Deployment failed, rolled back"
```

---

## 6. FRONTEND DEVELOPMENT STANDARDS (HTMX)

### 6.1 HTMX Architecture

**Server-Side Rendering with HTMX:**

```rust
// src/handlers/dashboard.rs

#[get("/dashboard")]
async fn dashboard(user: AuthUser, db: web::Data<Database>) -> Result<HttpResponse> {
    let websites = db.get_websites_for_user(user.id).await?;
    
    let html = render_template("dashboard.html", TemplateContext {
        websites,
        user,
    })?;

    Ok(HttpResponse::Ok()
        .content_type("text/html; charset=utf-8")
        .body(html))
}

// Handle HTMX requests (return HTML fragments)
#[post("/api/websites/{id}/enable")]
async fn enable_website(
    path: web::Path<i32>,
    user: AuthUser,
    db: web::Data<Database>,
) -> Result<HttpResponse> {
    let id = path.into_inner();
    
    // Update database
    db.enable_website(id, user.id).await?;
    
    // Return updated HTML fragment
    let website = db.get_website(id).await?;
    let html = render_template("fragments/website_row.html", TemplateContext {
        website,
    })?;

    Ok(HttpResponse::Ok()
        .content_type("text/html; charset=utf-8")
        .body(html))
}
```

**HTML Template with HTMX:**

```html
<!-- templates/dashboard.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - Hosting Panel</title>
    
    <!-- TailwindCSS -->
    <script src="https://cdn.tailwindcss.com"></script>
    
    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org"></script>
    
    <!-- Custom styles -->
    <link rel="stylesheet" href="/static/style.css">
</head>
<body class="bg-gray-50">
    <div class="container mx-auto">
        <h1 class="text-3xl font-bold py-6">Dashboard</h1>
        
        <!-- Websites list with HTMX -->
        <div id="websites-container"
             hx-get="/api/websites/list"
             hx-trigger="load, websiteUpdated from:body"
             hx-swap="innerHTML">
            <p class="text-gray-500">Loading websites...</p>
        </div>
    </div>
    
    <!-- Minimal custom JavaScript (HTMX handles most) -->
    <script src="/static/app.js" defer></script>
</body>
</html>

<!-- templates/fragments/websites_list.html -->
<table class="w-full border-collapse">
    <thead>
        <tr class="bg-gray-200">
            <th class="p-3 text-left">Domain</th>
            <th class="p-3 text-left">Status</th>
            <th class="p-3 text-left">Actions</th>
        </tr>
    </thead>
    <tbody>
        {% for website in websites %}
        <tr class="border-b hover:bg-gray-100" id="website-{{ website.id }}">
            <td class="p-3">{{ website.domain }}</td>
            <td class="p-3">
                <span class="px-2 py-1 rounded bg-green-100 text-green-800">
                    {{ website.status }}
                </span>
            </td>
            <td class="p-3">
                <button hx-post="/api/websites/{{ website.id }}/enable"
                        hx-target="#website-{{ website.id }}"
                        hx-swap="outerHTML"
                        class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                    Enable
                </button>
            </td>
        </tr>
        {% endfor %}
    </tbody>
</table>
```

### 6.2 Form Handling with HTMX

```html
<!-- templates/forms/create_website.html -->
<form hx-post="/api/websites"
      hx-target="#websites-container"
      hx-swap="innerHTML"
      hx-on="htmx:afterSwap: showNotification('Website created!')"
      class="space-y-4">
      
    <div>
        <label for="domain" class="block font-semibold">Domain</label>
        <input type="text" 
               id="domain" 
               name="domain" 
               required
               hx-post="/api/validate/domain"
               hx-target="#domain-error"
               hx-swap="outerHTML"
               class="w-full border rounded px-3 py-2">
        <div id="domain-error"></div>
    </div>

    <div>
        <label for="plan" class="block font-semibold">Plan</label>
        <select id="plan" name="plan" required class="w-full border rounded px-3 py-2">
            <option value="">Select a plan</option>
            <option value="starter">Starter</option>
            <option value="pro">Professional</option>
            <option value="business">Business</option>
        </select>
    </div>

    <button type="submit" class="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600">
        Create Website
    </button>
</form>
```

### 6.3 Real-Time Updates

```rust
// Server-Sent Events for real-time updates (without WebSockets)
#[get("/api/events/stream")]
async fn events_stream(user: AuthUser) -> impl Responder {
    let (tx, rx) = mpsc::unbounded_channel();

    // Subscribe user to updates
    // When events happen, send to all subscribed users
    
    let stream = ReceiverStream::new(rx)
        .map(|msg| {
            Ok(web::Bytes::from(format!("data: {}\n\n", msg)))
        });

    HttpResponse::Ok()
        .content_type("text/event-stream")
        .streaming_body(body::BodyStream::new(stream))
}
```

```html
<!-- Client-side HTMX + Server-Sent Events -->
<div hx-sse="connect:/api/events/stream"
     hx-trigger="sse:update-websites"
     hx-get="/api/websites/list"
     hx-target="#websites-container"
     hx-swap="innerHTML">
</div>
```

---

## 7. BACKEND AUTOMATION STANDARDS

### 7.1 Background Jobs with systemd Timers

**Instead of RabbitMQ/Celery, use systemd timers:**

```ini
# /etc/systemd/system/mail-processor.timer

[Unit]
Description=Mail Processor Timer
Requires=mail-processor.service

[Timer]
OnBootSec=5min
OnUnitActiveSec=5min
Accuracy=1s

[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/system/mail-processor.service

[Unit]
Description=Send Pending Emails
After=postgresql.service

[Service]
Type=oneshot
ExecStart=/usr/local/bin/hosting-panel-cli mail-send
User=hosting-panel
StandardOutput=journal
StandardError=journal
```

**Rust CLI for background tasks:**

```rust
// src/bin/cli.rs

#[derive(Parser)]
#[command(name = "hosting-panel-cli")]
#[command(about = "CLI for hosting panel background tasks")]
struct Args {
    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand)]
enum Command {
    /// Send pending emails
    MailSend,
    
    /// Process daily backups
    BackupProcess,
    
    /// Run malware scans
    MalwareScan,
    
    /// Clean old logs
    LogCleanup,
}

#[actix_web::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    match args.command {
        Command::MailSend => send_pending_emails().await?,
        Command::BackupProcess => process_backups().await?,
        Command::MalwareScan => run_malware_scans().await?,
        Command::LogCleanup => cleanup_old_logs().await?,
    }

    Ok(())
}

async fn send_pending_emails() -> Result<()> {
    let db = Database::new().await?;
    let mailer = MailService::new()?;

    let emails = db.get_pending_emails().await?;
    for email in emails {
        match mailer.send(&email).await {
            Ok(_) => db.mark_email_sent(email.id).await?,
            Err(e) => {
                eprintln!("Failed to send email {}: {}", email.id, e);
                db.increment_email_retry(email.id).await?;
            }
        }
    }

    Ok(())
}
```

### 7.2 Bash Scripts for Operations

**Backup script:**

```bash
#!/bin/bash
set -euo pipefail

# Configuration
BACKUP_DIR="/var/lib/hosting-panel/backups"
RETENTION_DAYS=30
LOG_FILE="/var/log/hosting-panel/backup.log"

# Logging
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup PostgreSQL database
log "Starting PostgreSQL backup..."
BACKUP_FILE="$BACKUP_DIR/postgresql_$(date +%Y%m%d_%H%M%S).sql.gz"

sudo -u postgres pg_dump hosting_db | gzip > "$BACKUP_FILE"
chmod 600 "$BACKUP_FILE"
log "PostgreSQL backup completed: $BACKUP_FILE"

# Backup application data
log "Starting application data backup..."
DATA_BACKUP="$BACKUP_DIR/data_$(date +%Y%m%d_%H%M%S).tar.gz"

tar --exclude='*.log' \
    --exclude='.git' \
    -czf "$DATA_BACKUP" \
    /var/lib/hosting-panel/data

chmod 600 "$DATA_BACKUP"
log "Data backup completed: $DATA_BACKUP"

# Remove old backups
log "Cleaning old backups..."
find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete
log "Old backups cleaned"

log "Backup completed successfully"
```

### 7.3 Python Automation Scripts

**Health monitoring:**

```python
#!/usr/bin/env python3
"""Health monitoring and alerting."""

import requests
import json
import subprocess
from typing import Dict, List
from pathlib import Path

CONFIG_FILE = Path("/etc/hosting-panel/monitoring.json")
LOG_FILE = Path("/var/log/hosting-panel/health.log")

def check_service_health(service: str) -> bool:
    """Check if systemd service is running."""
    try:
        result = subprocess.run(
            ["systemctl", "is-active", service],
            capture_output=True,
            text=True,
            timeout=5
        )
        return result.returncode == 0
    except Exception as e:
        log(f"Error checking service {service}: {e}")
        return False

def check_api_health(url: str) -> bool:
    """Check API health endpoint."""
    try:
        response = requests.get(
            f"{url}/health",
            timeout=5,
            verify=True
        )
        return response.status_code == 200
    except Exception as e:
        log(f"Error checking API health: {e}")
        return False

def check_disk_space() -> bool:
    """Check if disk usage is acceptable."""
    result = subprocess.run(
        ["df", "-h", "/var/lib/hosting-panel"],
        capture_output=True,
        text=True
    )
    
    lines = result.stdout.strip().split('\n')
    usage_percent = int(lines[1].split()[4].rstrip('%'))
    
    if usage_percent > 90:
        alert(f"Disk usage critical: {usage_percent}%")
        return False
    
    return True

def check_database() -> bool:
    """Check database connectivity."""
    try:
        result = subprocess.run(
            ["psql", "-U", "hosting", "-d", "hosting_db", "-c", "SELECT 1"],
            capture_output=True,
            timeout=5
        )
        return result.returncode == 0
    except Exception as e:
        log(f"Database check failed: {e}")
        return False

def run_health_checks() -> Dict[str, bool]:
    """Run all health checks."""
    checks = {
        "api": check_api_health("http://127.0.0.1:8001"),
        "database": check_database(),
        "disk": check_disk_space(),
        "service": check_service_health("hosting-panel"),
    }
    
    return checks

def log(message: str):
    """Log message."""
    with open(LOG_FILE, "a") as f:
        f.write(f"[{datetime.now()}] {message}\n")

def alert(message: str):
    """Send alert."""
    # Send to monitoring system (Prometheus, etc.)
    log(f"ALERT: {message}")

if __name__ == "__main__":
    results = run_health_checks()
    
    if all(results.values()):
        log("All health checks passed")
        exit(0)
    else:
        log(f"Health check failures: {results}")
        exit(1)
```

---

## 8. DEPLOYMENT & DEVOPS STANDARDS

### 8.1 Deployment with systemd + Ansible

**Deployment Process:**

1. **Build Phase** (CI/CD - GitHub Actions)
   - Compile Rust binary
   - Run tests
   - Security scan
   - Create artifact

2. **Deploy Phase** (Ansible)
   - Stop current service
   - Deploy new binary
   - Run migrations
   - Start service
   - Health check

**GitHub Actions Workflow:**

```yaml
name: Build and Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Rust
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      
      - name: Run tests
        run: cargo test --release
      
      - name: Security audit
        run: cargo audit
      
      - name: Build release
        run: cargo build --release
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: hosting-panel
          path: target/release/hosting-panel

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Download artifact
        uses: actions/download-artifact@v3
        with:
          name: hosting-panel
      
      - name: Deploy with Ansible
        run: |
          ansible-playbook -i hosts playbooks/deploy.yml \
            -e "binary_path=hosting-panel"
        env:
          ANSIBLE_HOST_KEY_CHECKING: "False"
```

### 8.2 Service Management Commands

**Common systemd commands:**

```bash
# Start/stop/restart service
sudo systemctl start hosting-panel
sudo systemctl stop hosting-panel
sudo systemctl restart hosting-panel

# Enable service to start on boot
sudo systemctl enable hosting-panel

# View service status
sudo systemctl status hosting-panel

# View service logs
sudo journalctl -u hosting-panel -f
sudo journalctl -u hosting-panel --since "1 hour ago"

# Check timer status
sudo systemctl list-timers
sudo systemctl status mail-processor.timer

# Reload systemd configuration
sudo systemctl daemon-reload

# View all services
sudo systemctl list-units --type service

# Monitor system resources
systemctl status -l
```

### 8.3 Monitoring Prometheus Metrics

**Rust app with Prometheus metrics:**

```rust
use prometheus::{Counter, Histogram, Registry};

// Create metrics
let http_requests: Counter = Counter::new(
    "http_requests_total",
    "Total HTTP requests"
).unwrap();

let request_duration: Histogram = Histogram::new(
    "http_request_duration_seconds",
    "HTTP request duration"
).unwrap();

// Register metrics
let registry = Registry::new();
registry.register(Box::new(http_requests.clone())).unwrap();
registry.register(Box::new(request_duration.clone())).unwrap();

// Use metrics
http_requests.inc();
let start = std::time::Instant::now();
// ... do work ...
request_duration.observe(start.elapsed().as_secs_f64());
```

**Prometheus scrape configuration:**

```yaml
# /etc/prometheus/prometheus.yml

global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'hosting-panel'
    static_configs:
      - targets: ['127.0.0.1:8001']
    metrics_path: '/metrics'
```

---

## 9. API DEVELOPMENT STANDARDS

### 9.1 REST API with Actix-web

**Standard API endpoints:**

```rust
#[get("/api/v1/websites")]
async fn list_websites(
    user: AuthUser,
    query: web::Query<ListQuery>,
    db: web::Data<Database>,
) -> Result<Json<ApiResponse<Vec<Website>>>, ApiError> {
    let limit = query.limit.unwrap_or(50).min(100);
    let offset = (query.page.unwrap_or(1) - 1) * limit;

    let websites = db.get_websites(user.id, limit, offset).await?;
    let total = db.count_websites(user.id).await?;

    Ok(Json(ApiResponse {
        status: "success",
        data: websites,
        pagination: Some(Pagination {
            page: query.page.unwrap_or(1),
            per_page: limit,
            total,
            total_pages: (total + limit - 1) / limit,
        }),
    }))
}

#[post("/api/v1/websites")]
async fn create_website(
    user: AuthUser,
    body: web::Json<CreateWebsiteRequest>,
    db: web::Data<Database>,
) -> Result<Json<ApiResponse<Website>>, ApiError> {
    // Validate input
    if body.domain.is_empty() {
        return Err(ApiError::ValidationError("domain is required".to_string()));
    }

    // Check authorization and permissions
    let account = db.get_user(user.id).await?;
    if !account.can_create_website() {
        return Err(ApiError::Forbidden);
    }

    // Create website
    let website = db.create_website(user.id, body.domain.clone()).await?;

    Ok(Json(ApiResponse {
        status: "success",
        data: website,
        pagination: None,
    }))
}
```

---

## 10. DATABASE STANDARDS

See previous document - PostgreSQL standards remain the same.

---

## 11. INFRASTRUCTURE & SYSADMIN STANDARDS (REFINED)

### 11.1 systemd Service Hardening

**Complete service template:**

```ini
[Unit]
Description=Hosting Control Panel
Documentation=man:hosting-panel(8)
After=network-online.target postgresql.service redis.service
Wants=network-online.target
BindsTo=postgresql.service redis.service

[Service]
Type=notify
ExecStart=/usr/local/bin/hosting-panel
ExecReload=/bin/kill -SIGHUP $MAINPID
EnvironmentFile=/etc/default/hosting-panel

# User and group
User=hosting-panel
Group=hosting-panel
DynamicUser=no

# Process management
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30s

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=hosting-panel
SyslogFacility=local7

# Restart policy
Restart=on-failure
RestartSec=5s
StartLimitInterval=600s
StartLimitBurst=3

# Security - Capabilities
AmbientCapabilities=
CapabilityBoundingSet=~CAP_SYS_ADMIN CAP_SYS_PTRACE
SecureBits=keep-caps
PrivateDevices=yes

# Security - Filesystem
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
NoNewPrivileges=yes
ReadWritePaths=/var/lib/hosting-panel /var/log/hosting-panel /run/hosting-panel

# Security - Process
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
RestrictNamespaces=yes
RestrictSUIDSGID=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
SystemCallFilter=@system-service
SystemCallFilter=~@privileged @resources

# Resources
LimitNOFILE=65535
LimitNPROC=4096
LimitCORE=0

[Install]
WantedBy=multi-user.target
```

### 11.2 Ansible Playbooks

**Initial provisioning:**

```yaml
# playbooks/provision.yml
---
- name: Provision hosting server
  hosts: all
  become: yes

  pre_tasks:
    - name: Update system packages
      apt:
        update_cache: yes
        upgrade: dist
      when: ansible_os_family == "Debian"

    - name: Install required packages
      apt:
        name:
          - build-essential
          - postgresql-client
          - redis-tools
          - curl
          - wget
          - git
          - htop
          - systemd
        state: present

  tasks:
    - name: Create hosting user
      user:
        name: hosting-panel
        home: /var/lib/hosting-panel
        shell: /usr/sbin/nologin
        system: yes
        create_home: yes

    - name: Create application directories
      file:
        path: "{{ item }}"
        state: directory
        owner: hosting-panel
        group: hosting-panel
        mode: '0755'
      loop:
        - /var/lib/hosting-panel
        - /var/lib/hosting-panel/config
        - /var/lib/hosting-panel/data
        - /var/lib/hosting-panel/scripts
        - /var/log/hosting-panel

    - name: Deploy systemd service files
      template:
        src: "{{ item }}.j2"
        dest: "/etc/systemd/system/{{ item }}"
        owner: root
        group: root
        mode: '0644'
      loop:
        - hosting-panel.service
        - hosting-panel-worker.service
        - mail-processor.timer
        - backup-website.timer
      notify: reload systemd

    - name: Configure UFW firewall
      ufw:
        rule: allow
        port: "{{ item }}"
        proto: tcp
      loop:
        - "80"
        - "443"
        - "22"

    - name: Setup log rotation
      template:
        src: logrotate.j2
        dest: /etc/logrotate.d/hosting-panel
        owner: root
        group: root
        mode: '0644'

  handlers:
    - name: reload systemd
      command: systemctl daemon-reload
```

---

## 12. DOCUMENTATION REQUIREMENTS

### Standard Documentation

- README with setup instructions
- API specification (OpenAPI 3.1)
- Architecture decision records (ADRs)
- Runbooks for operations
- Disaster recovery procedures
- Ansible playbook documentation

---

## 13. CODE REVIEW CHECKLIST

See previous document - remains largely the same, with additions for:
- HTMX usage best practices
- Bash script security checks
- Ansible playbook validation
- systemd service configuration

---

## 14. EMERGENCY PROCEDURES

### Procedure: Service Failure

```bash
#!/bin/bash
# Immediate response to service failure

# 1. Check service status
systemctl status hosting-panel

# 2. View recent logs
journalctl -u hosting-panel -n 50

# 3. Check resource usage
free -h
df -h

# 4. Restart service
systemctl restart hosting-panel

# 5. Verify recovery
curl http://127.0.0.1:8001/health

# 6. If still failing, investigate
systemctl start hosting-panel --no-block
sleep 5
journalctl -u hosting-panel -f
```

### Procedure: Database Connection Lost

```bash
#!/bin/bash
# Response to database connection errors

# 1. Check PostgreSQL service
systemctl status postgresql

# 2. Verify connectivity
psql -U hosting -d hosting_db -c "SELECT 1"

# 3. Check connection pool
psql -U postgres -d hosting_db -c "SELECT * FROM pg_stat_activity"

# 4. Restart if needed
systemctl restart postgresql

# 5. Restart application
systemctl restart hosting-panel
```

### Procedure: Disk Space Critical

```bash
#!/bin/bash
# Response to disk space issues

# 1. Check disk usage
df -h

# 2. Find large files
du -sh /var/lib/hosting-panel/* | sort -h

# 3. Check logs
du -sh /var/log/hosting-panel

# 4. Clean old backups
find /var/lib/hosting-panel/backups -mtime +30 -delete

# 5. Clean old logs
journalctl --vacuum=7d

# 6. Monitor
df -h
```

---

## SUMMARY OF KEY CHANGES v2.0

### Technology Choices

| Component | v1.0 | v2.0 | Reason |
|-----------|------|------|--------|
| Orchestration | Docker/Kubernetes | systemd | Simpler, lighter, direct OS integration |
| Frontend | React 18 | HTMX/Rust | Simpler, minimal JS, server-driven |
| Automation | Manual/Terraform | Ansible + Bash + Python | Idempotent, version-controlled, clear |
| Scheduling | RabbitMQ/Celery | systemd timers | Lightweight, built-in, reliable |
| Deployment | Manual | Ansible playbooks | Repeatable, version-controlled |

### Key Principles

✅ **Simplicity First** - systemd over Kubernetes, HTMX over React  
✅ **Automation Everything** - Ansible for all deployments  
✅ **Direct OS Integration** - Leverage Linux kernel features  
✅ **Minimal Dependencies** - Use built-in tools where possible  
✅ **Version Control** - All infrastructure and automation in Git  
✅ **Security by Default** - systemd hardening, Ansible validation  
✅ **Easy Debugging** - Direct systemd logs, clear service status  

---

## FINAL CHECKLIST

Before using these refined standards:

- [ ] Team understands systemd architecture
- [ ] Ansible playbooks tested locally
- [ ] HTMX frontend development workflow established
- [ ] Bash/Python automation scripts reviewed
- [ ] CI/CD pipeline configured for new stack
- [ ] Monitoring configured (Prometheus)
- [ ] Runbooks updated for systemd
- [ ] Emergency procedures practiced
- [ ] Team trained on new approach
- [ ] Documentation updated

---

**Version:** 2.0 (Refined)  
**Date:** November 2, 2025  
**Changes:** Removed Docker/Kubernetes, replaced with systemd; Removed React, replaced with HTMX; Added Ansible/Bash/Python automation  

**You are now ready for lightweight, maintainable, systemd-based infrastructure!**