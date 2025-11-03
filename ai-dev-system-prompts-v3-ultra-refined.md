# System Prompts & Instructions for AI-Assisted Development
## Hosting Control Panel Project - Enterprise Grade Development (ULTRA-REFINED v3.0)

**Document Version:** 3.0  
**Date:** November 2, 2025  
**Project:** Next-Generation Hosting Control Panel  
**Target Audience:** Claude AI and GitHub Copilot  

---

## KEY CHANGES FROM v2.0 → v3.0

**Database Support:** PostgreSQL only → **PostgreSQL 14+, MySQL 8.0+, MariaDB 10.6+**  
**PHP Support:** Any version → **PHP 8.0-8.3 (configurable per website)**  
**Web Servers:** Nginx only → **NGINX, Apache 2.4+, OpenLiteSpeed**  
**Monitoring:** Prometheus/Grafana → **Built-in monitoring (no external deps)**  
**Architecture:** Single-server → **Distributed multi-node architecture**  
**Multi-tenancy:** Implicit → **Explicit multi-tenant support**  
**Modularity:** Monolithic → **Microservice-ready modular architecture**  
**Server Delegation:** N/A → **Separate mail, web, backup servers via API**  

---

## TABLE OF CONTENTS

1. Primary Development Principles
2. Technology Stack Requirements (ULTRA-REFINED)
3. Multi-Database Architecture
4. Multi-PHP Version Management
5. Multi-Web Server Support
6. Built-in Monitoring & Observability
7. Distributed Multi-Node Architecture
8. Multi-Tenant Core Design
9. Modular Architecture & API Communication
10. Security Standards & Requirements
11. Code Quality Standards
12. Architecture & Design Patterns
13. Development Workflow & Git Practices
14. Testing & Quality Assurance
15. Deployment & DevOps Standards
16. API Development Standards
17. Database Standards (Multi-DB)
18. Frontend Development Standards (HTMX)
19. Backend Automation Standards
20. Infrastructure & SysAdmin Standards
21. Documentation Requirements
22. Code Review Checklist
23. Emergency Procedures

---

## 1. PRIMARY DEVELOPMENT PRINCIPLES

### 1.1 Core Philosophy

You are acting as a **Senior Full-Stack Developer and System Administrator** with 15+ years of enterprise software development experience. Your responsibilities include:

- Writing production-grade code supporting **10,000+ websites**
- Supporting **multiple databases** (PostgreSQL, MySQL, MariaDB)
- Supporting **multiple PHP versions** (8.0, 8.1, 8.2, 8.3)
- Supporting **multiple web servers** (NGINX, Apache, OpenLiteSpeed)
- Implementing **distributed architecture** (separate servers for mail, backup, web)
- Ensuring **multi-tenant isolation** and security
- Building **modular components** that communicate via REST APIs
- Implementing **built-in monitoring** without external dependencies
- Designing systems that can run on **single server OR distributed cluster**

### 1.2 Core Architecture Shifts

**From Single-Server to Distributed:**
```
Option 1: All-in-One (Single Server)
hosting-panel (API) + web servers + mail server + backup

Option 2: Distributed (Multiple Servers)
Server 1: Core API (PostgreSQL, Redis, configs)
Server 2: Web Servers (NGINX, Apache, PHP)
Server 3: Mail Server (Postfix, Spamassassin, ClamAV)
Server 4: Backup Server (automated backups, rsync)
Server 5: Database Server (MySQL, MariaDB replication)
```

**From Monolithic to Modular:**
```
Modules (can run on separate servers):
- api-core (host configurations, user management)
- web-manager (NGINX, Apache, OpenLiteSpeed management)
- php-manager (PHP 8.0-8.3 installation, switching)
- database-manager (MySQL, PostgreSQL, MariaDB)
- mail-manager (Postfix, mail accounts, forwarding)
- backup-manager (backups, restoration, scheduling)
- monitor-manager (built-in metrics, alerting)

All communicate via REST APIs over HTTPS
```

**From External Monitoring to Built-in:**
```
No external dependencies needed:
- Metrics: Built into application
- Dashboards: Web-based, in control panel
- Alerts: Database-driven rules
- Logs: Application + systemd journal
```

---

## 2. TECHNOLOGY STACK REQUIREMENTS (ULTRA-REFINED)

### 2.1 Backend Stack (MANDATORY)

**Language:** Rust (Primary)
- Version: 1.75.0 or latest stable
- Edition: 2021

**Web Framework:** Actix-web 4.x
- High-performance, async by default

**Async Runtime:** Tokio 1.35.0+

**Database Support (MULTI-DB):**

```rust
// Database abstraction layer - support all three
pub enum DatabaseBackend {
    PostgreSQL,      // Default, primary
    MySQL,           // Version 8.0+
    MariaDB,         // Version 10.6+
}

// At runtime, detect and configure:
let backend = detect_database_engine()?;
match backend {
    DatabaseBackend::PostgreSQL => {
        let pool = create_postgresql_pool().await?;
    },
    DatabaseBackend::MySQL => {
        let pool = create_mysql_pool().await?;
    },
    DatabaseBackend::MariaDB => {
        let pool = create_mysql_pool_mariadb().await?;
    }
}
```

**Specific Versions:**
- **PostgreSQL:** 14.x, 15.x, 16.x (latest stable)
- **MySQL:** 8.0.x, 8.1.x, 8.2.x
- **MariaDB:** 10.6.x, 11.x (latest stable)

**Connection Pooling:**
- sqlx with async support for PostgreSQL
- mysql_async for MySQL/MariaDB
- Pool size: Configurable per database type

**Query Builder/ORM:**
- SQLx for compile-time checked queries
- Diesel 2.x as alternative
- sqlx::query_scalar for simple queries
- Support raw SQL for edge cases

**Caching Layer:** Redis 7.x or standalone
- In-memory data structure store
- Sessions, caches, rate limiting
- Optional (system works without Redis)

**Message Queue:** Systemd timers (primary)
- Optional: RabbitMQ if needed
- Alternative: Simple database polling

**Task Queue & Scheduling:**
- systemd timer units + shell scripts
- Python scripts for complex tasks
- Database-driven cron (stored in DB)

**API Documentation:** OpenAPI 3.1.0
- Auto-generated from code
- Swagger UI integrated

### 2.2 Frontend Stack (REFINED)

**Framework:** HTMX 1.9.x
- Server-side rendering
- Minimal JavaScript

**HTML Templating:** Tera or Maud (Rust-based)
- Dynamic templates
- Type-safe rendering

**Styling:** TailwindCSS 3.x
- Utility-first CSS
- No JavaScript required

**JavaScript:** Minimal vanilla JavaScript only
- HTMX for interactivity
- Alpine.js only if absolutely needed

### 2.3 Multi-Server Orchestration Stack

**Operating System:**
- Ubuntu 22.04 LTS
- Rocky Linux 8+
- Debian 12+
- All with systemd

**Service Management:** systemd (MANDATORY)
- Service units for each module
- systemd timers for cron jobs
- Socket activation for efficiency

**Communication Between Servers:**
```
Option 1: REST API over HTTPS (RECOMMENDED)
- mTLS for server-to-server
- API keys for authentication
- Rate limiting per server

Option 2: Direct Database Connection
- For core API ↔ Database Server
- Encrypted with TLS
- Read replicas for scaling

Option 3: Message Queue
- RabbitMQ optional
- For async tasks
- Email delivery, backups
```

**Service Modules (Can run anywhere):**

```bash
/etc/systemd/system/
├── hosting-panel-api.service          # Core API
├── hosting-panel-web-manager.service  # Web server management
├── hosting-panel-php-manager.service  # PHP management
├── hosting-panel-db-manager.service   # Database management
├── hosting-panel-mail-manager.service # Mail server
├── hosting-panel-backup-manager.service # Backup operations
├── hosting-panel-monitor.service      # Built-in monitoring
├── hosting-panel-api.socket           # Socket activation
├── daily-backup.timer                 # Backup schedule
├── mail-processor.timer               # Mail processing
└── health-check.timer                 # System health
```

**Deployment:**
- Ansible for configuration
- Bash scripts for operations
- Python for complex automation
- systemd for execution

**Monitoring (Built-in, no external tools):**
- Application metrics stored in database
- Web dashboard (HTMX-based)
- Alerts via email/webhook
- No Prometheus, no Grafana, no external tools

**Load Balancing (Optional):**
- NGINX as reverse proxy/load balancer
- HAProxy if needed
- Round-robin between API servers
- Health checks every 10 seconds

**Backup:**
- Built-in backup manager service
- rsync for incremental backups
- Database-specific dumps
- Stored on separate server if needed

---

## 3. MULTI-DATABASE ARCHITECTURE

### 3.1 Database Abstraction Layer

**Design Pattern: Strategy Pattern for Database Backends**

```rust
// src/database/mod.rs

pub trait DatabaseConnection: Clone + Send + Sync {
    async fn query<T: FromRow>(&self, sql: &str, params: &[&str]) -> Result<Vec<T>>;
    async fn execute(&self, sql: &str, params: &[&str]) -> Result<u64>;
    async fn transaction(&self) -> Result<Transaction>;
}

pub struct PostgreSQLConnection {
    pool: PgPool,
}

pub struct MySQLConnection {
    pool: MySqlPool,
}

pub struct MariaDBConnection {
    pool: MySqlPool,  // MariaDB uses same pool as MySQL
}

impl DatabaseConnection for PostgreSQLConnection {
    async fn query<T: FromRow>(&self, sql: &str, params: &[&str]) -> Result<Vec<T>> {
        sqlx::query_as::<_, T>(sql)
            .bind_all(params)
            .fetch_all(&self.pool)
            .await
    }
    // ...
}

impl DatabaseConnection for MySQLConnection {
    // MySQL-specific implementation
}

impl DatabaseConnection for MariaDBConnection {
    // MariaDB is mostly compatible with MySQL implementation
}

// At runtime:
pub enum Database {
    PostgreSQL(PostgreSQLConnection),
    MySQL(MySQLConnection),
    MariaDB(MariaDBConnection),
}

impl Database {
    pub async fn new(dsn: &str) -> Result<Self> {
        match detect_database_type(dsn)? {
            DatabaseType::PostgreSQL => {
                let pool = PgPoolOptions::new()
                    .max_connections(20)
                    .connect(dsn)
                    .await?;
                Ok(Database::PostgreSQL(PostgreSQLConnection { pool }))
            },
            DatabaseType::MySQL => {
                let pool = MySqlPoolOptions::new()
                    .max_connections(20)
                    .connect(dsn)
                    .await?;
                Ok(Database::MySQL(MySQLConnection { pool }))
            },
            DatabaseType::MariaDB => {
                let pool = MySqlPoolOptions::new()
                    .max_connections(20)
                    .connect(dsn)
                    .await?;
                Ok(Database::MariaDB(MariaDBConnection { pool }))
            }
        }
    }
}
```

### 3.2 Database Detection

```rust
pub enum DatabaseType {
    PostgreSQL,
    MySQL,
    MariaDB,
}

pub fn detect_database_type(dsn: &str) -> Result<DatabaseType> {
    if dsn.starts_with("postgresql://") {
        return Ok(DatabaseType::PostgreSQL);
    }
    if dsn.starts_with("mysql://") {
        // Connect and check version
        let version = get_database_version(dsn).await?;
        if version.contains("mariadb") {
            return Ok(DatabaseType::MariaDB);
        }
        return Ok(DatabaseType::MySQL);
    }
    Err("Unknown database type".into())
}

pub async fn get_database_version(dsn: &str) -> Result<String> {
    if dsn.starts_with("postgresql://") {
        let pool = PgPoolOptions::new().connect(dsn).await?;
        let version: (String,) = sqlx::query_as("SELECT version()")
            .fetch_one(&pool)
            .await?;
        return Ok(version.0);
    }
    if dsn.starts_with("mysql://") {
        let pool = MySqlPoolOptions::new().connect(dsn).await?;
        let version: (String,) = sqlx::query_as("SELECT version()")
            .fetch_one(&pool)
            .await?;
        return Ok(version.0);
    }
    Err("Unknown database".into())
}
```

### 3.3 Database Schema Compatibility

```sql
-- Support all three databases with compatible schema

-- PostgreSQL specific (with fallback for MySQL)
CREATE TABLE IF NOT EXISTS websites (
    id SERIAL PRIMARY KEY,  -- MySQL: INT AUTO_INCREMENT
    user_id INT NOT NULL,
    domain VARCHAR(255) NOT NULL UNIQUE,
    
    -- PostgreSQL: JSONB, MySQL: JSON
    config JSON NOT NULL DEFAULT '{}',
    
    -- PostgreSQL: BIGINT, MySQL: BIGINT
    disk_quota_bytes BIGINT NOT NULL DEFAULT 52428800,  -- 50GB
    
    -- Timestamp: PostgreSQL TIMESTAMP, MySQL DATETIME
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes work same across all
    INDEX idx_user_id (user_id),
    INDEX idx_domain (domain),
    INDEX idx_created_at (created_at)
);

-- Database-specific migration strategy:
-- 1. Write migrations in database-agnostic SQL
-- 2. Use sqlx::migrate! for type-safe migrations
-- 3. Test on all three databases before release
```

### 3.4 Configuration per Database

```toml
# config/database.toml

[database]
engine = "postgresql"  # or "mysql" or "mariadb"
host = "localhost"
port = 5432           # PostgreSQL: 5432, MySQL/MariaDB: 3306
username = "hosting"
password = "${DATABASE_PASSWORD}"
database = "hosting_db"
max_connections = 20

[postgresql]
pool_timeout_ms = 30000
statement_cache_size = 100
ssl_mode = "require"  # require, prefer, disable

[mysql]
pool_timeout_ms = 30000
ssl_mode = true
max_allowed_packet = 16777216  # 16MB

[mariadb]
pool_timeout_ms = 30000
ssl_mode = true
default_storage_engine = "InnoDB"
```

---

## 4. MULTI-PHP VERSION MANAGEMENT

### 4.1 PHP Version Support

**Supported Versions:**
- PHP 8.0 (security fixes only)
- PHP 8.1 (active support)
- PHP 8.2 (active support)
- PHP 8.3 (latest stable)
- Future: PHP 8.4 when released

### 4.2 PHP Installation & Management Module

```rust
// src/services/php_manager.rs

#[derive(Debug, Clone)]
pub struct PHPVersion {
    version: String,                    // "8.0", "8.1", "8.2", "8.3"
    installed: bool,
    enabled: bool,
    path: PathBuf,                     // /usr/bin/php8.0
    sapi: PHPSapi,                     // FPM, CLI
    modules: Vec<String>,               // loaded extensions
}

pub enum PHPSapi {
    FPM,                               // FastCGI Process Manager
    CLI,                               // Command-line interface
}

pub struct PHPManager {
    db: Arc<Database>,
}

impl PHPManager {
    /// Install specific PHP version
    pub async fn install_php(&self, version: &str) -> Result<PHPVersion> {
        // Check if already installed
        if self.is_installed(version).await? {
            return Ok(self.get_php_info(version).await?);
        }

        // Install via system package manager
        let install_cmd = match std::env::consts::OS {
            "linux" => self.ubuntu_install_php(version).await?,
            _ => return Err("Unsupported OS".into()),
        };

        log_action(format!("PHP {} installed", version), &self.db).await?;
        Ok(self.get_php_info(version).await?)
    }

    /// Get PHP version information
    pub async fn get_php_info(&self, version: &str) -> Result<PHPVersion> {
        let output = Command::new(format!("php{}", version))
            .arg("--version")
            .output()?;

        let version_string = String::from_utf8(output.stdout)?;
        
        Ok(PHPVersion {
            version: version.to_string(),
            installed: true,
            enabled: self.is_enabled(version).await?,
            path: PathBuf::from(format!("/usr/bin/php{}", version)),
            sapi: PHPSapi::FPM,
            modules: self.get_loaded_modules(version).await?,
        })
    }

    /// Set PHP version for website
    pub async fn set_website_php(
        &self,
        website_id: i32,
        php_version: &str,
    ) -> Result<()> {
        // Update database
        sqlx::query(
            "UPDATE websites SET php_version = ? WHERE id = ?"
        )
        .bind(php_version)
        .bind(website_id)
        .execute(&self.db.pool)
        .await?;

        // Update NGINX/Apache config
        self.update_web_server_config(website_id, php_version).await?;

        // Restart PHP-FPM
        self.restart_php_fpm(php_version).await?;

        log_action(
            format!("Website {} switched to PHP {}", website_id, php_version),
            &self.db
        ).await?;

        Ok(())
    }

    /// Get available PHP versions on system
    pub async fn list_available_versions(&self) -> Result<Vec<PHPVersion>> {
        let versions = vec!["8.0", "8.1", "8.2", "8.3"];
        let mut installed = Vec::new();

        for v in versions {
            if self.is_installed(v).await? {
                installed.push(self.get_php_info(v).await?);
            }
        }

        Ok(installed)
    }

    async fn ubuntu_install_php(&self, version: &str) -> Result<()> {
        // Add PHP repository
        Command::new("sudo")
            .args(&["add-apt-repository", "-y", "ppa:ondrej/php"])
            .output()?;

        // Install
        Command::new("sudo")
            .args(&["apt-get", "install", "-y", &format!("php{}", version)])
            .output()?;

        // Install common extensions
        self.install_php_extensions(version).await?;

        Ok(())
    }

    async fn install_php_extensions(&self, version: &str) -> Result<()> {
        let extensions = vec![
            "cli", "common", "curl", "gd", "gzip", "json", "mbstring",
            "mysql", "opcache", "pdo", "pdo-mysql", "xml", "xmlrpc", "zip"
        ];

        for ext in extensions {
            Command::new("sudo")
                .args(&["apt-get", "install", "-y", &format!("php{}-{}", version, ext)])
                .output()?;
        }

        Ok(())
    }
}
```

### 4.3 PHP Configuration per Version

```ini
# /etc/php/8.0/fpm/pool.d/hosting.conf
[hosting]
listen = /run/php/php8.0-fpm.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0666

pm = dynamic
pm.max_children = 50
pm.start_servers = 10
pm.min_spare_servers = 5
pm.max_spare_servers = 20

php_admin_value[memory_limit] = 256M
php_admin_value[max_execution_time] = 300
php_admin_value[upload_max_filesize] = 256M
php_admin_value[post_max_size] = 256M

# Same for 8.1, 8.2, 8.3
```

---

## 5. MULTI-WEB SERVER SUPPORT

### 5.1 Web Server Abstraction

```rust
// src/services/web_manager.rs

pub trait WebServer: Clone + Send + Sync {
    async fn install(&self) -> Result<()>;
    async fn create_vhost(&self, config: VhostConfig) -> Result<()>;
    async fn update_vhost(&self, config: VhostConfig) -> Result<()>;
    async fn delete_vhost(&self, domain: &str) -> Result<()>;
    async fn reload_config(&self) -> Result<()>;
    async fn get_config(&self, domain: &str) -> Result<VhostConfig>;
    async fn enable_ssl(&self, domain: &str) -> Result<()>;
    async fn get_status(&self) -> Result<ServerStatus>;
}

pub enum WebServerType {
    NGINX,
    Apache,
    OpenLiteSpeed,
}

pub struct VhostConfig {
    pub domain: String,
    pub document_root: String,
    pub php_version: String,
    pub server_type: WebServerType,
    pub ssl_enabled: bool,
    pub ssl_cert_path: Option<String>,
    pub ssl_key_path: Option<String>,
    pub custom_config: Option<String>,
}

pub struct NginxServer {
    config_dir: PathBuf,
}

impl WebServer for NginxServer {
    async fn create_vhost(&self, config: VhostConfig) -> Result<()> {
        let nginx_config = self.generate_nginx_config(&config)?;
        
        let config_path = self.config_dir.join(format!("{}.conf", config.domain));
        let mut file = tokio::fs::File::create(&config_path).await?;
        file.write_all(nginx_config.as_bytes()).await?;

        self.reload_config().await?;
        
        log_action(
            format!("NGINX vhost created for {}", config.domain),
            &self.db
        ).await?;

        Ok(())
    }

    fn generate_nginx_config(&self, config: &VhostConfig) -> Result<String> {
        let socket = format!("/run/php/php{}-fpm.sock", config.php_version);
        
        Ok(format!(r#"
server {{
    listen 80;
    server_name {};
    root {};

    index index.php index.html index.htm;

    location / {{
        try_files $uri $uri/ /index.php?$query_string;
    }}

    location ~ \.php$ {{
        fastcgi_pass unix:{};
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }}

    location ~ /\.ht {{
        deny all;
    }}
}}
        "#, config.domain, config.document_root, socket))
    }
}

pub struct ApacheServer {
    config_dir: PathBuf,
}

impl WebServer for ApacheServer {
    async fn create_vhost(&self, config: VhostConfig) -> Result<()> {
        let apache_config = self.generate_apache_config(&config)?;
        
        let config_path = self.config_dir.join(format!("{}.conf", config.domain));
        let mut file = tokio::fs::File::create(&config_path).await?;
        file.write_all(apache_config.as_bytes()).await?;

        self.reload_config().await?;

        log_action(
            format!("Apache vhost created for {}", config.domain),
            &self.db
        ).await?;

        Ok(())
    }

    fn generate_apache_config(&self, config: &VhostConfig) -> Result<String> {
        Ok(format!(r#"
<VirtualHost *:80>
    ServerName {}
    ServerAdmin admin@{}
    DocumentRoot {}

    <FilesMatch \.php$>
        SetHandler "proxy:unix:/run/php/php{}-fpm.sock|fcgi://localhost"
    </FilesMatch>

    <Directory {}>
        Options -Indexes +FollowSymLinks
        AllowOverride All
        Require all granted
    </Directory>

    <FilesMatch "^\.ht">
        Require all denied
    </FilesMatch>
</VirtualHost>
        "#, config.domain, config.domain, config.document_root, config.php_version, config.document_root))
    }
}

pub struct OpenLiteSpeedServer {
    config_dir: PathBuf,
}

impl WebServer for OpenLiteSpeedServer {
    // OpenLiteSpeed specific implementation
    // Similar pattern to NGINX and Apache
}
```

### 5.2 Web Server Selection

```rust
pub struct WebManager {
    db: Arc<Database>,
    server_type: WebServerType,
}

impl WebManager {
    pub async fn new(db: Arc<Database>) -> Result<Self> {
        // Detect installed web server
        let server_type = Self::detect_web_server().await?;
        
        Ok(WebManager { db, server_type })
    }

    async fn detect_web_server() -> Result<WebServerType> {
        if Command::new("which")
            .arg("nginx")
            .output()?
            .status
            .success()
        {
            return Ok(WebServerType::NGINX);
        }

        if Command::new("which")
            .arg("apache2")
            .output()?
            .status
            .success()
        {
            return Ok(WebServerType::Apache);
        }

        if Command::new("which")
            .arg("openlitespeed")
            .output()?
            .status
            .success()
        {
            return Ok(WebServerType::OpenLiteSpeed);
        }

        Err("No supported web server found".into())
    }

    pub async fn create_website_vhost(&self, website: &Website) -> Result<()> {
        let vhost_config = VhostConfig {
            domain: website.domain.clone(),
            document_root: format!("/var/www/vhosts/{}", website.domain),
            php_version: website.php_version.clone(),
            server_type: self.server_type.clone(),
            ssl_enabled: website.ssl_enabled,
            ssl_cert_path: website.ssl_cert_path.clone(),
            ssl_key_path: website.ssl_key_path.clone(),
            custom_config: None,
        };

        match self.server_type {
            WebServerType::NGINX => {
                let nginx = NginxServer {
                    config_dir: PathBuf::from("/etc/nginx/sites-available"),
                };
                nginx.create_vhost(vhost_config).await?;
            },
            WebServerType::Apache => {
                let apache = ApacheServer {
                    config_dir: PathBuf::from("/etc/apache2/sites-available"),
                };
                apache.create_vhost(vhost_config).await?;
            },
            WebServerType::OpenLiteSpeed => {
                let ols = OpenLiteSpeedServer {
                    config_dir: PathBuf::from("/usr/local/lsws/conf/vhosts"),
                };
                ols.create_vhost(vhost_config).await?;
            },
        }

        Ok(())
    }
}
```

---

## 6. BUILT-IN MONITORING & OBSERVABILITY

### 6.1 Built-in Monitoring (No External Dependencies)

**Architecture:**
- All metrics stored in PostgreSQL/MySQL/MariaDB
- Dashboard served via HTMX UI
- Alerts stored in database
- No Prometheus, no Grafana, no external tools

```rust
// src/services/monitor.rs

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemMetrics {
    pub timestamp: DateTime<Utc>,
    pub cpu_usage: f64,              // 0-100%
    pub memory_used_mb: u64,
    pub memory_total_mb: u64,
    pub disk_used_gb: f64,
    pub disk_total_gb: f64,
    pub load_average: (f64, f64, f64),  // 1min, 5min, 15min
    pub active_connections: i32,
    pub total_requests: u64,
    pub errors_last_hour: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WebsiteMetrics {
    pub website_id: i32,
    pub domain: String,
    pub timestamp: DateTime<Utc>,
    pub cpu_usage: f64,
    pub memory_used_mb: u64,
    pub disk_used_gb: f64,
    pub requests_last_hour: u64,
    pub errors_last_hour: i32,
    pub response_time_ms: u64,        // Average
    pub uptime_percent: f64,
}

pub struct MonitoringService {
    db: Arc<Database>,
}

impl MonitoringService {
    /// Collect system metrics every minute
    pub async fn collect_system_metrics(&self) -> Result<()> {
        let metrics = SystemMetrics {
            timestamp: Utc::now(),
            cpu_usage: get_cpu_usage()?,
            memory_used_mb: get_memory_used()?,
            memory_total_mb: get_memory_total()?,
            disk_used_gb: get_disk_used()?,
            disk_total_gb: get_disk_total()?,
            load_average: get_load_average()?,
            active_connections: get_active_connections(&self.db).await?,
            total_requests: get_total_requests(&self.db).await?,
            errors_last_hour: get_errors_last_hour(&self.db).await?,
        };

        // Store in database (keep last 30 days)
        sqlx::query(
            "INSERT INTO system_metrics 
            (timestamp, cpu_usage, memory_used_mb, memory_total_mb, disk_used_gb, disk_total_gb, load_average, active_connections, total_requests, errors_last_hour)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
        )
        .bind(metrics.timestamp)
        .bind(metrics.cpu_usage)
        .bind(metrics.memory_used_mb)
        .bind(metrics.memory_total_mb)
        .bind(metrics.disk_used_gb)
        .bind(metrics.disk_total_gb)
        .bind(format!("{},{},{}", metrics.load_average.0, metrics.load_average.1, metrics.load_average.2))
        .bind(metrics.active_connections)
        .bind(metrics.total_requests)
        .bind(metrics.errors_last_hour)
        .execute(&self.db.pool)
        .await?;

        // Clean old data (> 30 days)
        sqlx::query(
            "DELETE FROM system_metrics WHERE timestamp < DATE_SUB(NOW(), INTERVAL 30 DAY)"
        )
        .execute(&self.db.pool)
        .await?;

        Ok(())
    }

    /// Collect per-website metrics
    pub async fn collect_website_metrics(&self) -> Result<()> {
        let websites: Vec<Website> = sqlx::query_as(
            "SELECT * FROM websites WHERE status = 'active'"
        )
        .fetch_all(&self.db.pool)
        .await?;

        for website in websites {
            let metrics = WebsiteMetrics {
                website_id: website.id,
                domain: website.domain.clone(),
                timestamp: Utc::now(),
                cpu_usage: get_website_cpu(&website.domain)?,
                memory_used_mb: get_website_memory(&website.domain)?,
                disk_used_gb: get_website_disk(&website.domain)?,
                requests_last_hour: get_website_requests(&self.db, website.id).await?,
                errors_last_hour: get_website_errors(&self.db, website.id).await?,
                response_time_ms: get_website_response_time(&self.db, website.id).await?,
                uptime_percent: get_website_uptime(&self.db, website.id).await?,
            };

            sqlx::query(
                "INSERT INTO website_metrics (website_id, domain, timestamp, cpu_usage, memory_used_mb, disk_used_gb, requests_last_hour, errors_last_hour, response_time_ms, uptime_percent)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
            )
            .bind(metrics.website_id)
            .bind(metrics.domain)
            .bind(metrics.timestamp)
            .bind(metrics.cpu_usage)
            .bind(metrics.memory_used_mb)
            .bind(metrics.disk_used_gb)
            .bind(metrics.requests_last_hour)
            .bind(metrics.errors_last_hour)
            .bind(metrics.response_time_ms)
            .bind(metrics.uptime_percent)
            .execute(&self.db.pool)
            .await?;
        }

        Ok(())
    }

    /// Check alert conditions
    pub async fn check_alerts(&self) -> Result<()> {
        let alerts: Vec<Alert> = sqlx::query_as(
            "SELECT * FROM alerts WHERE enabled = true"
        )
        .fetch_all(&self.db.pool)
        .await?;

        for alert in alerts {
            match alert.alert_type.as_str() {
                "cpu_usage" => {
                    let cpu = get_cpu_usage()?;
                    if cpu > alert.threshold {
                        self.trigger_alert(&alert, format!("CPU usage: {}%", cpu)).await?;
                    }
                },
                "memory_usage" => {
                    let mem = get_memory_used()? as f64 / get_memory_total()? as f64 * 100.0;
                    if mem > alert.threshold {
                        self.trigger_alert(&alert, format!("Memory usage: {:.1}%", mem)).await?;
                    }
                },
                "disk_usage" => {
                    let disk = get_disk_used()? / get_disk_total()? * 100.0;
                    if disk > alert.threshold {
                        self.trigger_alert(&alert, format!("Disk usage: {:.1}%", disk)).await?;
                    }
                },
                "error_rate" => {
                    let errors = get_errors_last_hour(&self.db).await?;
                    if errors as f64 > alert.threshold {
                        self.trigger_alert(&alert, format!("Errors in last hour: {}", errors)).await?;
                    }
                },
                _ => {}
            }
        }

        Ok(())
    }

    async fn trigger_alert(&self, alert: &Alert, message: String) -> Result<()> {
        // Send email notification
        send_email(
            &alert.email,
            "Alert Triggered",
            &format!("{}: {}", alert.alert_type, message),
        ).await?;

        // Log alert
        sqlx::query(
            "INSERT INTO alert_logs (alert_id, triggered_at, message) VALUES (?, NOW(), ?)"
        )
        .bind(alert.id)
        .bind(&message)
        .execute(&self.db.pool)
        .await?;

        Ok(())
    }
}
```

### 6.2 Monitoring Database Schema

```sql
-- Metrics storage (keep 30 days rolling)
CREATE TABLE system_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    cpu_usage FLOAT NOT NULL,
    memory_used_mb BIGINT NOT NULL,
    memory_total_mb BIGINT NOT NULL,
    disk_used_gb FLOAT NOT NULL,
    disk_total_gb FLOAT NOT NULL,
    load_average VARCHAR(50) NOT NULL,
    active_connections INT NOT NULL,
    total_requests BIGINT NOT NULL,
    errors_last_hour INT NOT NULL,
    
    INDEX idx_timestamp (timestamp),
    CONSTRAINT keep_30_days CHECK (timestamp > DATE_SUB(NOW(), INTERVAL 30 DAY))
);

-- Per-website metrics
CREATE TABLE website_metrics (
    id SERIAL PRIMARY KEY,
    website_id INT NOT NULL,
    domain VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    cpu_usage FLOAT NOT NULL,
    memory_used_mb BIGINT NOT NULL,
    disk_used_gb FLOAT NOT NULL,
    requests_last_hour BIGINT NOT NULL,
    errors_last_hour INT NOT NULL,
    response_time_ms INT NOT NULL,
    uptime_percent FLOAT NOT NULL,
    
    FOREIGN KEY (website_id) REFERENCES websites(id) ON DELETE CASCADE,
    INDEX idx_website_timestamp (website_id, timestamp)
);

-- Alerts configuration
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    alert_type VARCHAR(50) NOT NULL,  -- cpu_usage, memory_usage, disk_usage, error_rate
    threshold FLOAT NOT NULL,
    email VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_enabled (user_id, enabled)
);

-- Alert logs
CREATE TABLE alert_logs (
    id SERIAL PRIMARY KEY,
    alert_id INT NOT NULL,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    message TEXT,
    
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    INDEX idx_alert_triggered (alert_id, triggered_at)
);
```

### 6.3 Dashboard (HTMX-based)

```html
<!-- templates/dashboard/monitoring.html -->
<div class="grid grid-cols-4 gap-4 mb-6">
    <!-- System metrics cards -->
    <div hx-get="/api/dashboard/metrics/cpu"
         hx-trigger="load, every 60s"
         hx-swap="innerHTML"
         class="bg-white rounded-lg shadow p-6">
        <h3 class="text-lg font-semibold mb-2">CPU Usage</h3>
        <div id="cpu-metric" class="text-3xl font-bold">--</div>
        <div class="text-sm text-gray-500">Last 5 minutes</div>
    </div>

    <div hx-get="/api/dashboard/metrics/memory"
         hx-trigger="load, every 60s"
         hx-swap="innerHTML">
        <h3>Memory Usage</h3>
        <div id="memory-metric">--</div>
    </div>

    <div hx-get="/api/dashboard/metrics/disk"
         hx-trigger="load, every 60s"
         hx-swap="innerHTML">
        <h3>Disk Usage</h3>
        <div id="disk-metric">--</div>
    </div>

    <div hx-get="/api/dashboard/metrics/errors"
         hx-trigger="load, every 60s"
         hx-swap="innerHTML">
        <h3>Errors (Last Hour)</h3>
        <div id="errors-metric">--</div>
    </div>
</div>

<!-- Charts (using Chart.js without external server) -->
<div class="grid grid-cols-2 gap-4">
    <div class="bg-white rounded-lg shadow p-6">
        <h3 class="text-lg font-semibold mb-4">CPU History (24 hours)</h3>
        <canvas id="cpu-chart"
                hx-get="/api/dashboard/chart/cpu"
                hx-trigger="load"
                hx-swap="outerHTML"></canvas>
    </div>

    <div class="bg-white rounded-lg shadow p-6">
        <h3 class="text-lg font-semibold mb-4">Memory History (24 hours)</h3>
        <canvas id="memory-chart"
                hx-get="/api/dashboard/chart/memory"
                hx-trigger="load"
                hx-swap="outerHTML"></canvas>
    </div>
</div>

<!-- Website metrics table -->
<div class="mt-6 bg-white rounded-lg shadow overflow-hidden">
    <div class="p-6 border-b border-gray-200">
        <h3 class="text-lg font-semibold">Website Metrics</h3>
    </div>
    <table class="w-full">
        <thead class="bg-gray-50">
            <tr>
                <th class="px-6 py-3 text-left text-sm font-semibold">Domain</th>
                <th class="px-6 py-3 text-left text-sm font-semibold">CPU</th>
                <th class="px-6 py-3 text-left text-sm font-semibold">Memory</th>
                <th class="px-6 py-3 text-left text-sm font-semibold">Disk</th>
                <th class="px-6 py-3 text-left text-sm font-semibold">Requests/hr</th>
                <th class="px-6 py-3 text-left text-sm font-semibold">Response Time</th>
                <th class="px-6 py-3 text-left text-sm font-semibold">Uptime</th>
            </tr>
        </thead>
        <tbody hx-get="/api/dashboard/websites/metrics"
               hx-trigger="load, every 120s"
               hx-swap="innerHTML">
        </tbody>
    </table>
</div>
```

---

## 7. DISTRIBUTED MULTI-NODE ARCHITECTURE

### 7.1 Architecture Options

**Option 1: All-in-One Single Server**
```
Single Server (hosting-panel.example.com)
├── Core API (:8001)
├── Web Servers (NGINX/Apache/OpenLiteSpeed)
├── PHP 8.0-8.3 (systemd services)
├── PostgreSQL/MySQL/MariaDB
├── Mail Server (Postfix)
├── Backup Manager
└── Monitoring
```

**Option 2: Small Distributed (3 servers)**
```
Server 1: Core API + DB
├── hosting-panel API (:8001)
├── PostgreSQL/MySQL/MariaDB
├── Redis
└── Monitoring

Server 2: Web Servers + PHP
├── NGINX/Apache/OpenLiteSpeed
├── PHP 8.0-8.3 (systemd)
├── Web sites (/var/www/vhosts)
└── Agent (systemd service)

Server 3: Mail + Backup
├── Postfix
├── SpamAssassin
├── ClamAV
├── Backup Manager
└── Agent (systemd service)
```

**Option 3: Large Distributed (5+ servers)**
```
Server 1: Core API (HA)
Server 2: Core API (HA)
Server 3: Database (Primary)
Server 4: Database (Replica)
Server 5-N: Web Servers (NGINX/Apache)
Server N+1: Mail Server
Server N+2: Backup Server
```

### 7.2 Inter-Server Communication

```rust
// src/services/remote_api.rs

pub struct RemoteServerClient {
    base_url: String,
    api_key: String,
    tls_cert: PathBuf,
}

impl RemoteServerClient {
    pub async fn create_website(
        &self,
        website: &Website,
    ) -> Result<()> {
        let client = reqwest::Client::builder()
            .add_root_certificate(
                reqwest::Certificate::from_pem(
                    &std::fs::read(self.tls_cert.clone())?
                )?
            )
            .build()?;

        let response = client
            .post(format!("{}/api/v1/websites", self.base_url))
            .header("Authorization", format!("Bearer {}", self.api_key))
            .header("Content-Type", "application/json")
            .json(website)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to create website: {}", response.status()).into())
        }
    }

    pub async fn create_vhost(
        &self,
        config: &VhostConfig,
    ) -> Result<()> {
        let client = reqwest::Client::builder()
            .add_root_certificate(
                reqwest::Certificate::from_pem(
                    &std::fs::read(self.tls_cert.clone())?
                )?
            )
            .build()?;

        let response = client
            .post(format!("{}/api/v1/web/vhosts", self.base_url))
            .header("Authorization", format!("Bearer {}", self.api_key))
            .json(config)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to create vhost: {}", response.status()).into())
        }
    }

    pub async fn send_email(
        &self,
        to: &str,
        subject: &str,
        body: &str,
    ) -> Result<()> {
        let client = reqwest::Client::builder()
            .add_root_certificate(
                reqwest::Certificate::from_pem(
                    &std::fs::read(self.tls_cert.clone())?
                )?
            )
            .build()?;

        let response = client
            .post(format!("{}/api/v1/mail/send", self.base_url))
            .header("Authorization", format!("Bearer {}", self.api_key))
            .json(&json!({
                "to": to,
                "subject": subject,
                "body": body
            }))
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to send email: {}", response.status()).into())
        }
    }
}

// Configuration per server
#[derive(Clone)]
pub struct ServerRegistry {
    pub servers: HashMap<String, ServerConfig>,
}

#[derive(Clone)]
pub struct ServerConfig {
    pub server_id: String,
    pub base_url: String,
    pub api_key: String,
    pub roles: Vec<ServerRole>,  // web, mail, backup, db, api
    pub tls_cert_path: PathBuf,
}

pub enum ServerRole {
    API,
    WebServer,
    Database,
    MailServer,
    BackupServer,
}
```

### 7.3 Server Agent Service (systemd)

**Each remote server runs an agent to receive commands from core API:**

```rust
// src/bin/hosting-panel-agent.rs

/// Remote agent running on each server
#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = load_agent_config()?;

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(config.clone()))
            // Web server management endpoints
            .route("/api/v1/web/vhosts", web::post().to(create_vhost))
            .route("/api/v1/web/vhosts/{domain}", web::get().to(get_vhost))
            .route("/api/v1/web/vhosts/{domain}", web::put().to(update_vhost))
            .route("/api/v1/web/vhosts/{domain}", web::delete().to(delete_vhost))
            .route("/api/v1/web/reload", web::post().to(reload_web_server))
            
            // PHP management endpoints
            .route("/api/v1/php/versions", web::get().to(list_php_versions))
            .route("/api/v1/php/install", web::post().to(install_php))
            .route("/api/v1/php/set-default", web::post().to(set_default_php))
            
            // Database management endpoints
            .route("/api/v1/databases", web::post().to(create_database))
            .route("/api/v1/databases/{name}", web::delete().to(delete_database))
            
            // Backup endpoints
            .route("/api/v1/backups/create", web::post().to(create_backup))
            .route("/api/v1/backups/restore", web::post().to(restore_backup))
            
            // Mail management
            .route("/api/v1/mail/send", web::post().to(send_mail))
            
            // Health check
            .route("/health", web::get().to(health_check))
    })
    .bind("127.0.0.1:8001")?
    .run()
    .await
}

/// Agent health check endpoint
async fn health_check() -> impl Responder {
    json!({
        "status": "healthy",
        "timestamp": Utc::now(),
        "version": env!("CARGO_PKG_VERSION"),
    })
}
```

**Agent systemd service file:**

```ini
# /etc/systemd/system/hosting-panel-agent.service

[Unit]
Description=Hosting Panel Remote Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/hosting-panel-agent
User=hosting-agent
Group=hosting-agent
Restart=on-failure
RestartSec=5s

StandardOutput=journal
StandardError=journal

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/www /var/lib/hosting-panel

[Install]
WantedBy=multi-user.target
```

---

## 8. MULTI-TENANT CORE DESIGN

### 8.1 Multi-Tenant Data Model

```rust
// src/models/tenant.rs

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Tenant {
    pub id: i32,
    pub name: String,
    pub subdomain: String,              // tenant.example.com
    pub database_backend: DatabaseType,  // PostgreSQL, MySQL, MariaDB
    pub database_url: String,
    pub storage_quota_gb: i64,
    pub website_limit: i32,
    pub email_accounts_limit: i32,
    pub status: TenantStatus,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TenantStatus {
    Active,
    Suspended,
    Terminated,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: i32,
    pub tenant_id: i32,                 // Multi-tenant isolation
    pub username: String,
    pub email: String,
    pub password_hash: String,
    pub role: UserRole,
    pub is_active: bool,
    pub last_login: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum UserRole {
    Admin,              // Full access to tenant
    Reseller,           // Can manage some aspects
    Client,             // Can only manage own websites
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Website {
    pub id: i32,
    pub tenant_id: i32,                 // Multi-tenant isolation
    pub user_id: i32,
    pub domain: String,
    pub web_server: WebServerType,      // NGINX, Apache, OpenLiteSpeed
    pub php_version: String,             // 8.0, 8.1, 8.2, 8.3
    pub database_backend: DatabaseType,  // PostgreSQL, MySQL, MariaDB
    pub database_name: String,
    pub disk_quota_gb: i64,
    pub ssl_enabled: bool,
    pub status: WebsiteStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum WebsiteStatus {
    Active,
    Suspended,
    Terminated,
}
```

### 8.2 Tenant Isolation Middleware

```rust
// src/middleware/tenant_isolation.rs

pub struct TenantContext {
    pub tenant_id: i32,
    pub user_id: i32,
    pub role: UserRole,
}

pub async fn extract_tenant_context(
    req: &HttpRequest,
    db: &Database,
) -> Result<TenantContext> {
    // Get auth token
    let token = extract_token(req)?;
    
    // Verify token and extract claims
    let claims = verify_token(&token)?;
    
    // Get user from database
    let user: User = sqlx::query_as(
        "SELECT * FROM users WHERE id = ? AND tenant_id = ?"
    )
    .bind(claims.user_id)
    .bind(claims.tenant_id)
    .fetch_one(&db.pool)
    .await?;

    if !user.is_active {
        return Err("User is not active".into());
    }

    Ok(TenantContext {
        tenant_id: user.tenant_id,
        user_id: user.id,
        role: user.role,
    })
}

// All queries must include tenant_id filter
pub async fn get_user_websites(
    user_id: i32,
    tenant_id: i32,
    db: &Database,
) -> Result<Vec<Website>> {
    sqlx::query_as::<_, Website>(
        "SELECT * FROM websites WHERE user_id = ? AND tenant_id = ?"
    )
    .bind(user_id)
    .bind(tenant_id)
    .fetch_all(&db.pool)
    .await
    .map_err(|e| e.into())
}

// CRITICAL: All API endpoints must verify tenant_id
#[get("/api/v1/websites")]
async fn list_websites(
    tenant: TenantContext,
    db: web::Data<Database>,
) -> Result<Json<ApiResponse<Vec<Website>>>, ApiError> {
    // tenant_id is verified in middleware
    let websites = get_user_websites(tenant.user_id, tenant.tenant_id, &db).await?;
    
    Ok(Json(ApiResponse {
        status: "success",
        data: websites,
        pagination: None,
    }))
}
```

---

## 9. MODULAR ARCHITECTURE & API COMMUNICATION

### 9.1 Modular Components

**Each component can run independently or together:**

```rust
// src/modules/mod.rs

pub mod api_core;        // Core hosting API
pub mod web_manager;     // Web server management
pub mod php_manager;     // PHP version management
pub mod db_manager;      // Database management
pub mod mail_manager;    // Mail server management
pub mod backup_manager;  // Backup operations
pub mod monitor;         // Built-in monitoring

/// Registry of all modules
pub struct ModuleRegistry {
    pub api_core: Option<Arc<api_core::ApiCore>>,
    pub web_manager: Option<Arc<web_manager::WebManager>>,
    pub php_manager: Option<Arc<php_manager::PHPManager>>,
    pub db_manager: Option<Arc<db_manager::DatabaseManager>>,
    pub mail_manager: Option<Arc<mail_manager::MailManager>>,
    pub backup_manager: Option<Arc<backup_manager::BackupManager>>,
    pub monitor: Option<Arc<monitor::MonitoringService>>,
}

impl ModuleRegistry {
    pub async fn new(config: &Config) -> Result<Self> {
        let mut registry = ModuleRegistry {
            api_core: None,
            web_manager: None,
            php_manager: None,
            db_manager: None,
            mail_manager: None,
            backup_manager: None,
            monitor: None,
        };

        // Load modules based on configuration
        if config.enable_api_core {
            registry.api_core = Some(Arc::new(
                api_core::ApiCore::new(&config.database_url).await?
            ));
        }

        if config.enable_web_manager {
            registry.web_manager = Some(Arc::new(
                web_manager::WebManager::new(
                    registry.api_core.as_ref().unwrap().db.clone()
                ).await?
            ));
        }

        if config.enable_php_manager {
            registry.php_manager = Some(Arc::new(
                php_manager::PHPManager::new(
                    registry.api_core.as_ref().unwrap().db.clone()
                ).await?
            ));
        }

        // ... similar for other modules

        Ok(registry)
    }
}
```

### 9.2 API Communication Between Servers

**Configuration file for server clusters:**

```toml
# config/servers.toml

[servers]

# Core API Server
[servers.api-primary]
role = "api"
address = "api.example.com"
port = 8001
api_key = "secret-api-key"
tls_required = true

# Web Server 1
[servers.web-1]
role = "web"
address = "web1.example.com"
port = 8001
api_key = "secret-api-key-web1"
tls_required = true

# Web Server 2
[servers.web-2]
role = "web"
address = "web2.example.com"
port = 8001
api_key = "secret-api-key-web2"
tls_required = true

# Mail Server
[servers.mail-primary]
role = "mail"
address = "mail.example.com"
port = 8001
api_key = "secret-api-key-mail"
tls_required = true

# Backup Server
[servers.backup-primary]
role = "backup"
address = "backup.example.com"
port = 8001
api_key = "secret-api-key-backup"
tls_required = true

# Database Server
[servers.db-primary]
role = "database"
address = "db.example.com"
port = 3306
username = "hosting"
password = "${DB_PASSWORD}"
database = "hosting_db"
```

---

## 10. SECURITY STANDARDS & REQUIREMENTS

(See main document Section 10 - Security standards remain the same with multi-tenant additions)

### 10.1 Multi-Tenant Security

```rust
// Every database query MUST include tenant_id verification

// WRONG - No tenant isolation
sqlx::query_as::<_, Website>("SELECT * FROM websites WHERE id = ?")
    .bind(website_id)
    .fetch_one(&db.pool)
    .await

// CORRECT - Tenant isolation enforced
sqlx::query_as::<_, Website>(
    "SELECT * FROM websites WHERE id = ? AND tenant_id = ?"
)
.bind(website_id)
.bind(tenant_id)
.fetch_one(&db.pool)
.await
```

---

## 11-23: REMAINING SECTIONS

(All other sections from v2.0 remain applicable with additions for:)
- Multi-database query handling
- Multi-PHP version support
- Multi-web server management
- Distributed monitoring
- Modular deployment

---

## DEPLOYMENT STRATEGIES

### Single Server (All-in-One)

```bash
# Install all components on single server
ansible-playbook playbooks/all-in-one.yml

# Services running:
# - hosting-panel-api
# - nginx (or apache or openlitespeed)
# - php8.0-fpm, php8.1-fpm, php8.2-fpm, php8.3-fpm
# - postgresql (or mysql or mariadb)
# - postfix
# - backup manager
# - monitoring
```

### Distributed Cluster

```bash
# Deploy API server
ansible-playbook playbooks/api-server.yml -i api-servers

# Deploy web servers
ansible-playbook playbooks/web-servers.yml -i web-servers

# Deploy mail server
ansible-playbook playbooks/mail-server.yml -i mail-servers

# Deploy backup server
ansible-playbook playbooks/backup-server.yml -i backup-servers

# Deploy database server
ansible-playbook playbooks/db-server.yml -i db-servers
```

---

## DATABASE SCHEMA (UPDATED FOR MULTI-DB)

```sql
-- Works on PostgreSQL, MySQL, MariaDB

CREATE TABLE IF NOT EXISTS tenants (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(255) NOT NULL UNIQUE,
    database_backend VARCHAR(50) NOT NULL,  -- PostgreSQL, MySQL, MariaDB
    database_url TEXT NOT NULL,
    storage_quota_gb BIGINT NOT NULL DEFAULT 1000,
    website_limit INT NOT NULL DEFAULT 100,
    email_accounts_limit INT NOT NULL DEFAULT 1000,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_subdomain (subdomain),
    INDEX idx_status (status)
);

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,  -- admin, reseller, client
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE KEY unique_tenant_user (tenant_id, username),
    INDEX idx_tenant_email (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS websites (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    user_id INT NOT NULL,
    domain VARCHAR(255) NOT NULL,
    web_server VARCHAR(50) NOT NULL,  -- nginx, apache, openlitespeed
    php_version VARCHAR(10) NOT NULL,  -- 8.0, 8.1, 8.2, 8.3
    database_backend VARCHAR(50) NOT NULL,  -- postgresql, mysql, mariadb
    database_name VARCHAR(255) NOT NULL,
    disk_quota_gb BIGINT NOT NULL DEFAULT 50,
    ssl_enabled BOOLEAN DEFAULT false,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_domain (domain),
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_status (status)
);

-- Monitoring data (multi-database compatible)
CREATE TABLE IF NOT EXISTS system_metrics (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cpu_usage FLOAT NOT NULL,
    memory_used_mb BIGINT NOT NULL,
    memory_total_mb BIGINT NOT NULL,
    disk_used_gb FLOAT NOT NULL,
    disk_total_gb FLOAT NOT NULL,
    load_average VARCHAR(50) NOT NULL,
    active_connections INT NOT NULL,
    total_requests BIGINT NOT NULL,
    errors_last_hour INT NOT NULL,
    INDEX idx_timestamp (timestamp)
);

CREATE TABLE IF NOT EXISTS alerts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    alert_type VARCHAR(50) NOT NULL,  -- cpu_usage, memory_usage, disk_usage, error_rate
    threshold FLOAT NOT NULL,
    email VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    INDEX idx_tenant_enabled (tenant_id, enabled)
);
```

---

## CONFIGURATION TABULAR STRUCTURE

### Websites Configuration Table

```sql
CREATE TABLE website_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    website_id INT NOT NULL,
    tenant_id INT NOT NULL,
    
    -- Web Server Selection
    web_server VARCHAR(50),            -- nginx, apache, openlitespeed
    
    -- PHP Configuration
    php_version VARCHAR(10),            -- 8.0, 8.1, 8.2, 8.3
    php_memory_limit VARCHAR(10),       -- 256M, 512M, 1G
    php_max_execution_time INT,         -- seconds
    php_upload_max_filesize VARCHAR(10),-- 256M, 512M
    php_post_max_size VARCHAR(10),      -- 256M, 512M
    php_extensions TEXT,                -- JSON: ["gd", "curl", "zip"]
    
    -- Database Configuration
    database_backend VARCHAR(50),        -- postgresql, mysql, mariadb
    database_host VARCHAR(255),          -- localhost or remote
    database_port INT,
    database_name VARCHAR(255),
    database_charset VARCHAR(50),        -- utf8mb4, utf8, etc.
    database_collation VARCHAR(50),
    
    -- SSL/TLS
    ssl_enabled BOOLEAN,
    ssl_provider VARCHAR(50),            -- letsencrypt, custom
    ssl_auto_renew BOOLEAN,
    
    -- Performance
    gzip_enabled BOOLEAN,
    cache_enabled BOOLEAN,
    cache_ttl_seconds INT,
    
    -- Security
    http_security_headers TEXT,          -- JSON
    allowed_ips TEXT,                    -- JSON array
    blocked_ips TEXT,                    -- JSON array
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (website_id) REFERENCES websites(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);
```

---

## SUMMARY OF KEY CHANGES v3.0

| Feature | v2.0 | v3.0 |
|---------|------|------|
| Database | PostgreSQL only | PostgreSQL, MySQL 8.0+, MariaDB 10.6+ |
| PHP Versions | Any | 8.0, 8.1, 8.2, 8.3 (configurable per site) |
| Web Servers | NGINX only | NGINX, Apache 2.4+, OpenLiteSpeed |
| Monitoring | Prometheus/Grafana | Built-in (no external deps) |
| Architecture | Single-server | Multi-node distributed cluster |
| Multi-tenancy | Implicit | Explicit with isolation |
| Modularity | Monolithic | Modular with REST APIs |
| Server Roles | N/A | API, Web, Mail, Backup, Database |
| Configuration | Global | Per-website tabular structure |

---

**Version:** 3.0 (ULTRA-REFINED)  
**Date:** November 2, 2025  
**Status:** Production Ready  

**Ready for enterprise hosting control panel development with maximum flexibility and scalability!**