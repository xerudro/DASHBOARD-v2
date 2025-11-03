# Quick Reference - v3.0 ULTRA-REFINED
## Multi-DB, Multi-PHP, Multi-Web Server, Distributed, Built-in Monitoring

**Print & Laminate!**

---

## KEY FEATURES v3.0

✅ **Multi-Database:** PostgreSQL 14+, MySQL 8.0+, MariaDB 10.6+  
✅ **Multi-PHP:** 8.0, 8.1, 8.2, 8.3 per website  
✅ **Multi-Web Server:** NGINX, Apache 2.4+, OpenLiteSpeed  
✅ **Distributed:** Separate servers for API, web, mail, backup, database  
✅ **Built-in Monitoring:** No Prometheus/Grafana needed  
✅ **Multi-tenant:** Explicit tenant isolation with role-based access  
✅ **Modular:** Independent components via REST APIs  
✅ **Tabular Config:** Per-website configuration table  

---

## TECH STACK v3.0

| Component | Support | Notes |
|-----------|---------|-------|
| **Database** | PostgreSQL 14+, MySQL 8.0+, MariaDB 10.6+ | Auto-detection at runtime |
| **PHP** | 8.0, 8.1, 8.2, 8.3 | Per-website configuration |
| **Web Server** | NGINX, Apache 2.4+, OpenLiteSpeed | Auto-detection, switching |
| **Backend** | Rust 1.75+ (Actix-web 4.x) | Multi-DB abstraction layer |
| **Frontend** | HTMX 1.9.x (Tera/Maud templates) | Server-driven |
| **Orchestration** | systemd (no Docker) | Distributed via REST APIs |
| **Monitoring** | Built-in (database stored) | No external tools |
| **Automation** | Ansible + Bash + Python | systemd timers for scheduling |

---

## DATABASE ABSTRACTION LAYER

```rust
// Select at runtime
pub enum DatabaseBackend {
    PostgreSQL,    // Default
    MySQL,         // 8.0+
    MariaDB,       // 10.6+
}

// All queries are database-agnostic
sqlx::query_as::<_, User>(
    "SELECT * FROM users WHERE tenant_id = ? AND id = ?"
)
.bind(tenant_id)
.bind(user_id)
.fetch_one(&db.pool)
.await
```

### Connection Pools
```
PostgreSQL: sqlx (async)
MySQL: mysql_async (async)
MariaDB: mysql_async (compatible)
Max connections: Configurable per DB type
```

---

## MULTI-PHP VERSION MANAGEMENT

### Installation
```bash
# Install all supported versions
ansible-playbook playbooks/install-php-all.yml

# Result: /usr/bin/php8.0, /usr/bin/php8.1, /usr/bin/php8.2, /usr/bin/php8.3
# Running: /etc/systemd/system/php8.0-fpm.service, php8.1-fpm, etc.
```

### Per-Website Selection
```rust
// Set website to PHP 8.3
php_manager.set_website_php(website_id, "8.3").await?;

// Updates:
// 1. websites.php_version = "8.3"
// 2. Web server config (NGINX/Apache)
// 3. Restarts PHP-FPM for version 8.3
// 4. Logs action in audit trail
```

### Configuration
```toml
# Per PHP version
[php."8.0"]
memory_limit = "256M"
max_execution_time = 300
opcache_enabled = true

[php."8.1"]
memory_limit = "512M"
max_execution_time = 300

[php."8.2"]
memory_limit = "512M"
max_execution_time = 300

[php."8.3"]
memory_limit = "512M"
max_execution_time = 300
jit_enabled = true
```

---

## MULTI-WEB SERVER SUPPORT

### Supported Servers
```
NGINX     - Production, high performance
Apache    - Traditional, widely compatible
OpenLiteSpeed - Proprietary, high performance
```

### Auto-Detection
```rust
// Automatically detects installed web server
let server = WebManager::detect_web_server().await?;

// Or configure explicitly
web_server: WebServerType::NGINX
```

### Per-Website Server
```sql
-- Table: website_configs
website_id | web_server | php_version | database_backend
-----------+------------+-------------+------------------
1          | nginx      | 8.3         | mysql
2          | apache     | 8.1         | postgresql
3          | openlitespeed | 8.2      | mariadb
```

### Configuration Templates
```ini
# NGINX - /etc/nginx/sites-available/{domain}.conf
server {
    listen 80;
    server_name example.com;
    root /var/www/vhosts/example.com;
    
    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_param SCRIPT_FILENAME ...;
    }
}

# Apache - /etc/apache2/sites-available/{domain}.conf
<VirtualHost *:80>
    ServerName example.com
    DocumentRoot /var/www/vhosts/example.com
    
    <FilesMatch \.php$>
        SetHandler "proxy:unix:/run/php/php8.3-fpm.sock|fcgi://..."
    </FilesMatch>
</VirtualHost>

# OpenLiteSpeed - /usr/local/lsws/conf/vhosts/{domain}/
Similar configuration...
```

---

## BUILT-IN MONITORING (NO EXTERNAL TOOLS)

### Metrics Collection (Every Minute)
```rust
// Automatic collection via systemd timer
monitoring.collect_system_metrics().await?;
monitoring.collect_website_metrics().await?;
monitoring.check_alerts().await?;
```

### Stored in Database
```sql
-- System metrics (30-day rolling)
SELECT * FROM system_metrics
WHERE timestamp > DATE_SUB(NOW(), INTERVAL 30 DAY)
ORDER BY timestamp DESC;

-- Website metrics
SELECT * FROM website_metrics
WHERE website_id = ? AND tenant_id = ?
ORDER BY timestamp DESC;

-- Alerts configuration
SELECT * FROM alerts WHERE tenant_id = ? AND enabled = true;

-- Alert logs
SELECT * FROM alert_logs WHERE alert_id = ? ORDER BY triggered_at DESC;
```

### Dashboard (HTMX, Real-time)
```html
<!-- Auto-refresh every 60 seconds -->
<div hx-get="/api/dashboard/metrics"
     hx-trigger="load, every 60s"
     hx-swap="innerHTML">
</div>

<!-- Charts -->
<canvas id="cpu-chart"
        hx-get="/api/dashboard/chart/cpu?hours=24"
        hx-trigger="load"></canvas>

<!-- Website metrics table -->
<table hx-get="/api/dashboard/websites"
       hx-trigger="load, every 120s"
       hx-swap="tbody">
</table>
```

### Alert Types
```
- cpu_usage (threshold: 80%, 95%)
- memory_usage (threshold: 85%, 95%)
- disk_usage (threshold: 90%, 95%)
- error_rate (threshold: 1%, 5%)
- service_down (immediate)
- certificate_expiry (days remaining)
```

---

## DISTRIBUTED MULTI-NODE ARCHITECTURE

### Single Server (All-in-One)
```
hosting-panel.example.com
├── API Core
├── NGINX + Apache + OpenLiteSpeed
├── PHP 8.0-8.3
├── PostgreSQL/MySQL/MariaDB
├── Postfix (mail)
├── Backup manager
└── Monitoring
```

### 3-Server Cluster
```
Server 1: API + Database
├── Hosting Panel API :8001
├── PostgreSQL/MySQL/MariaDB
└── Redis

Server 2: Web Servers
├── NGINX/Apache/OpenLiteSpeed
├── PHP 8.0-8.3
├── Agent service
└── Websites in /var/www/vhosts

Server 3: Mail + Backup
├── Postfix
├── Mail manager agent
├── Backup manager
└── rsync server
```

### 5+ Server Cluster
```
API Servers (HA): 2+
Database Servers: Primary + Replica
Web Servers: 3+
Mail Server: 1+
Backup Server: 1+
Load Balancer: NGINX/HAProxy (optional)
```

### Inter-Server Communication
```
Method 1: REST API over HTTPS (RECOMMENDED)
- mTLS between servers
- API key authentication
- Rate limiting per server
- Service discovery via DNS

Method 2: Direct Database
- Core API → Database Server
- Encrypted with TLS
- Connection pooling
- Read replicas for scaling

Method 3: Message Queue (Optional)
- RabbitMQ for async tasks
- Email delivery, backups
- Event streaming
```

---

## SERVER ROLES & CONFIGURATION

### Role: API Server
```toml
[roles]
enable_api_core = true
enable_web_manager = false
enable_php_manager = false
enable_mail_manager = false
enable_backup_manager = false

# Listens on :8001
# Communicates with other servers via API
# Stores all configs in database
```

### Role: Web Server
```toml
[roles]
enable_api_core = false
enable_web_manager = true
enable_php_manager = true
enable_mail_manager = false
enable_backup_manager = false

# Agent listens on :8001
# Receives commands from API server
# Manages NGINX/Apache/OpenLiteSpeed
# Manages PHP 8.0-8.3
```

### Role: Mail Server
```toml
[roles]
enable_api_core = false
enable_web_manager = false
enable_php_manager = false
enable_mail_manager = true
enable_backup_manager = false

# Agent listens on :8001
# Manages Postfix, SpamAssassin, ClamAV
# Communicates with API for mail accounts
```

### Role: Backup Server
```toml
[roles]
enable_api_core = false
enable_web_manager = false
enable_php_manager = false
enable_mail_manager = false
enable_backup_manager = true

# Agent listens on :8001
# Manages backups (rsync, database dumps)
# Restoration operations
# Scheduled via API
```

---

## MULTI-TENANT CORE

### Tenant Isolation
```rust
// EVERY query includes tenant_id
// No cross-tenant data access possible

// Wrong - Missing tenant_id
sqlx::query_as::<_, Website>("SELECT * FROM websites WHERE id = ?")
    .bind(website_id)

// Correct - Tenant verified
sqlx::query_as::<_, Website>(
    "SELECT * FROM websites WHERE id = ? AND tenant_id = ?"
)
.bind(website_id)
.bind(tenant_id)
```

### Tenant Table Structure
```sql
CREATE TABLE tenants (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    subdomain VARCHAR(255) UNIQUE,
    database_backend VARCHAR(50),  -- PostgreSQL, MySQL, MariaDB
    database_url TEXT,
    storage_quota_gb BIGINT,
    website_limit INT,
    email_accounts_limit INT,
    status VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Separate tenant database possible:
-- tenant1.db_url = "postgresql://host/tenant1_db"
-- tenant2.db_url = "mysql://host/tenant2_db"
-- Each tenant can use different database type!
```

### User Roles
```
Admin     - Full access to tenant
Reseller  - Can manage specific aspects
Client    - Can only manage own websites
```

---

## WEBSITE CONFIGURATION TABLE

### Schema (Per-Website Settings)
```sql
CREATE TABLE website_configs (
    id INT PRIMARY KEY,
    website_id INT,
    tenant_id INT,
    
    -- Web Server
    web_server VARCHAR(50),     -- nginx, apache, openlitespeed
    
    -- PHP
    php_version VARCHAR(10),    -- 8.0, 8.1, 8.2, 8.3
    php_memory_limit VARCHAR(10),
    php_max_execution_time INT,
    php_upload_max_filesize VARCHAR(10),
    php_post_max_size VARCHAR(10),
    php_extensions TEXT,        -- JSON array
    
    -- Database
    database_backend VARCHAR(50),  -- postgresql, mysql, mariadb
    database_host VARCHAR(255),
    database_port INT,
    database_name VARCHAR(255),
    database_charset VARCHAR(50),
    
    -- SSL/TLS
    ssl_enabled BOOLEAN,
    ssl_provider VARCHAR(50),   -- letsencrypt, custom
    ssl_auto_renew BOOLEAN,
    
    -- Performance
    gzip_enabled BOOLEAN,
    cache_enabled BOOLEAN,
    cache_ttl_seconds INT,
    
    -- Security
    http_security_headers TEXT, -- JSON
    allowed_ips TEXT,          -- JSON array
    blocked_ips TEXT,          -- JSON array
    
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Example Data
```sql
INSERT INTO website_configs VALUES (
    1,                          -- id
    1,                          -- website_id
    1,                          -- tenant_id
    'nginx',                    -- web_server
    '8.3',                      -- php_version
    '512M',                     -- php_memory_limit
    300,                        -- php_max_execution_time
    '256M',                     -- php_upload_max_filesize
    '256M',                     -- php_post_max_size
    '["gd","curl","zip"]',      -- php_extensions
    'mysql',                    -- database_backend
    'db.example.com',           -- database_host
    3306,                       -- database_port
    'website1_db',              -- database_name
    'utf8mb4',                  -- database_charset
    true,                       -- ssl_enabled
    'letsencrypt',              -- ssl_provider
    true,                       -- ssl_auto_renew
    true,                       -- gzip_enabled
    true,                       -- cache_enabled
    3600,                       -- cache_ttl_seconds
    '{}',                       -- http_security_headers
    '[]',                       -- allowed_ips
    '[]',                       -- blocked_ips
    NOW(),                      -- created_at
    NOW()                       -- updated_at
);
```

---

## DEPLOYMENT PATTERNS

### Pattern 1: All-in-One
```bash
# Single server with all components
ansible-playbook playbooks/all-in-one.yml

# Install scripts:
- setup-postgresql-mysql-mariadb.sh (choose one)
- setup-php-8.0-8.3.sh (all versions)
- setup-nginx-apache-openlitespeed.sh (choose one)
- setup-postfix-mail.sh
- setup-backup-manager.sh
- setup-monitoring.sh
```

### Pattern 2: API + Web + Mail + Backup
```bash
# API Server
ansible-playbook playbooks/deploy-api.yml -i api-servers.ini

# Web Servers (1+)
ansible-playbook playbooks/deploy-web.yml -i web-servers.ini

# Mail Server
ansible-playbook playbooks/deploy-mail.yml -i mail-servers.ini

# Backup Server
ansible-playbook playbooks/deploy-backup.yml -i backup-servers.ini
```

### Pattern 3: Full HA Cluster
```bash
# Load Balancer
ansible-playbook playbooks/deploy-lb.yml -i lb-servers.ini

# API Servers (HA)
ansible-playbook playbooks/deploy-api.yml -i api-servers.ini

# Database (Primary + Replica)
ansible-playbook playbooks/deploy-db-primary.yml -i db-primary.ini
ansible-playbook playbooks/deploy-db-replica.yml -i db-replica.ini

# Web Servers (3+)
ansible-playbook playbooks/deploy-web.yml -i web-servers.ini

# Mail Server
ansible-playbook playbooks/deploy-mail.yml -i mail-servers.ini

# Backup Server
ansible-playbook playbooks/deploy-backup.yml -i backup-servers.ini
```

---

## SYSTEMD SERVICES v3.0

```bash
# Core services
sudo systemctl start hosting-panel-api
sudo systemctl start hosting-panel-web-manager
sudo systemctl start hosting-panel-php-manager
sudo systemctl start hosting-panel-db-manager
sudo systemctl start hosting-panel-mail-manager
sudo systemctl start hosting-panel-backup-manager
sudo systemctl start hosting-panel-monitor

# PHP services (one per version)
sudo systemctl start php8.0-fpm
sudo systemctl start php8.1-fpm
sudo systemctl start php8.2-fpm
sudo systemctl start php8.3-fpm

# Web servers (choose one)
sudo systemctl start nginx
# OR
sudo systemctl start apache2
# OR
sudo systemctl start openlitespeed

# Timers
sudo systemctl start daily-backup.timer
sudo systemctl start mail-processor.timer
sudo systemctl start health-check.timer
```

---

## COMMANDS QUICK REFERENCE

### Database Operations
```bash
# Detect database type
curl -X GET https://api.example.com/api/v1/system/database/info

# Switch to MySQL
UPDATE tenants SET database_backend='mysql' WHERE id=1;

# Switch to PostgreSQL
UPDATE tenants SET database_backend='postgresql' WHERE id=1;

# List supported versions
SELECT DISTINCT php_version FROM websites;
```

### PHP Version Management
```bash
# Install PHP 8.3
ansible-playbook playbooks/install-php.yml -e php_version=8.3

# Switch website to PHP 8.3
curl -X PUT https://api.example.com/api/v1/websites/1 \
  -H "Content-Type: application/json" \
  -d '{"php_version":"8.3"}'

# List installed PHP versions
sudo ls /etc/php/
```

### Web Server Management
```bash
# Create vhost
curl -X POST https://api.example.com/api/v1/web/vhosts \
  -H "Content-Type: application/json" \
  -d '{"domain":"example.com","php_version":"8.3","web_server":"nginx"}'

# Test configuration
sudo nginx -t        # NGINX
sudo apache2ctl -t   # Apache
```

### Monitoring
```bash
# View system metrics
SELECT * FROM system_metrics ORDER BY timestamp DESC LIMIT 10;

# View website metrics
SELECT * FROM website_metrics WHERE website_id = 1 ORDER BY timestamp DESC;

# View alerts
SELECT * FROM alerts WHERE tenant_id = 1;

# Trigger health check
curl https://api.example.com/health
```

---

## TROUBLESHOOTING v3.0

### Issue: Website won't load
```
1. Check web server config:
   /etc/nginx/sites-available/{domain}.conf
   /etc/apache2/sites-available/{domain}.conf

2. Check PHP-FPM socket exists:
   ls -la /run/php/php8.3-fpm.sock

3. Check website_configs table:
   SELECT * FROM website_configs WHERE website_id = ?;

4. Verify database connection:
   Test connection string from database_backend + database_url

5. Check logs:
   sudo journalctl -u nginx -f
   sudo journalctl -u php8.3-fpm -f
```

### Issue: Distributed server not responding
```
1. Check network connectivity:
   ping server.example.com

2. Verify API key:
   grep api_key /etc/hosting-panel/servers.toml

3. Check TLS certificate:
   sudo openssl s_client -connect server.example.com:8001

4. View agent logs:
   sudo journalctl -u hosting-panel-agent -f

5. Check agent service:
   sudo systemctl status hosting-panel-agent
```

### Issue: Monitoring not updating
```
1. Check monitoring service:
   sudo systemctl status hosting-panel-monitor

2. Check database connectivity:
   Can API server reach database server?

3. View metric collection:
   sudo journalctl -u hosting-panel-monitor -f

4. Check alert rules:
   SELECT * FROM alerts WHERE enabled = true;
```

---

## CHECKLIST BEFORE DEPLOYMENT

### Single Server
- [ ] Database (PostgreSQL, MySQL, or MariaDB) installed
- [ ] PHP 8.0-8.3 installed with FPM
- [ ] Web server (NGINX, Apache, or OpenLiteSpeed) installed
- [ ] Postfix installed
- [ ] systemd services created
- [ ] Monitoring enabled
- [ ] Backups configured

### Distributed Servers
- [ ] API server running
- [ ] Web server(s) running with agent
- [ ] Database server running
- [ ] Mail server running with agent
- [ ] Backup server running with agent
- [ ] Monitoring enabled on API server
- [ ] mTLS configured between servers
- [ ] Firewall rules configured
- [ ] Network connectivity verified

---

**Version:** 3.0 (ULTRA-REFINED)  
**Date:** November 2, 2025  
**Status:** Production Ready  

**Ready to build enterprise hosting control panel with maximum flexibility!**