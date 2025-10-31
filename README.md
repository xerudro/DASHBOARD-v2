VIP Hosting Panel v2 - Enterprise Edition
Modern, secure, and high-performance hosting control panel built with Go, HTMX, and Python/Ansible.
🚀 Features
Infrastructure Management

Multi-Database Support: MySQL 5.7/8.0, PostgreSQL 12-16, MongoDB 4.4-7.0, Redis 6-7
Multi-PHP Support: PHP 5.6, 7.0, 7.1, 7.2, 7.3, 7.4, 8.0, 8.1, 8.2, 8.3
Multi-Node.js Support: Node.js 14.x, 16.x, 18.x, 20.x, 21.x
Multi-Webserver: Nginx, Apache2, LiteSpeed, Caddy (with auto-switching)
Server Providers: Hetzner, DigitalOcean, Vultr, AWS, Custom SSH

Security Features

WAF Integration: ModSecurity with OWASP Core Rule Set
Custom WAF Rules: Per-site and global rule management
Anti Code Injection: Input validation and sanitization
XSS Protection: Content Security Policy headers
SQL Injection Prevention: Prepared statements enforcement
ClamAV Antivirus: Real-time file scanning
Fail2ban: Brute force protection for SSH, FTP, Email, Panel
Firewall Manager: UFW/iptables with preset rules
SSL/TLS Management: Let's Encrypt auto-renewal + custom certs
2FA Authentication: TOTP support for all user roles
Audit Logging: Comprehensive action tracking

Email Services

Mail Server: Postfix (SMTP) + Dovecot (IMAP/POP3)
Webmail: Roundcube and SnappyMail integrated
Email Accounts: Create/delete/manage mailboxes
Email Aliases: Unlimited aliases per domain
Email Forwarding: Local and external forwarding
Spam Protection: SpamAssassin integration
DKIM/SPF/DMARC: Auto-configuration
Mail Quotas: Per-account storage limits
Email Filters: Sieve-based filtering

DNS Management

DNS Servers: Bind9 (primary) or PowerDNS
DNS Providers: Cloudflare, Route53, Custom
Zone Management: A, AAAA, CNAME, MX, TXT, SRV, CAA records
DNSSEC Support: Automated key management
DNS Templates: Quick setup for common configurations
Geo DNS: Location-based routing
Health Checks: Automatic failover

Website Management

One-Click Apps: WordPress, Laravel, Node.js, Static sites
Git Deployment: GitHub/GitLab webhooks
Zero-Downtime Deploy: Blue-green deployments
Staging Environments: Isolated testing environments
File Manager: Web-based with upload/download
FTP/SFTP: Per-user account management
Cron Jobs: Visual cron editor
Environment Variables: Secure secret management
Custom nginx/Apache configs: Per-site overrides

Database Management

phpMyAdmin: MySQL/MariaDB management
pgAdmin: PostgreSQL management
Adminer: Universal database tool
MongoDB Compass: MongoDB GUI
Redis Commander: Redis management
Database Backups: Automated dumps
Query Monitor: Slow query detection

Monitoring & Analytics

Real-time Metrics: CPU, RAM, Disk, Network
Application Performance: Response times, error rates
Uptime Monitoring: HTTP/HTTPS checks
SSL Certificate Expiry: Alerts before expiration
Resource Alerts: Customizable thresholds
Log Aggregation: Centralized log viewer
Traffic Analytics: Bandwidth usage per site

Backup & Recovery

Automated Backups: Scheduled full/incremental
Backup Storage: Local, S3, FTP, SFTP
One-Click Restore: Point-in-time recovery
Backup Encryption: AES-256 encryption
Retention Policies: Configurable cleanup
Backup Verification: Automated integrity checks

Billing & Reseller

Multi-tenant: Unlimited reseller levels
Stripe Integration: Subscriptions and one-time payments
Invoice Generation: Automated PDF invoices
Usage-Based Billing: Metered resources
Reseller Margins: Configurable markup
Client Portal: Self-service management
Payment Methods: Card, PayPal, Bank transfer

🏗️ Architecture
┌─────────────────────────────────────────────┐
│     FRONTEND: HTMX + Alpine.js              │
│     • Server-Side Rendering (Templ)         │
│     • Real-time updates (SSE)               │
│     • Tailwind CSS + DaisyUI                │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│     API GATEWAY: Go + Fiber                 │
│     • JWT Authentication                    │
│     • Rate Limiting                         │
│     • RBAC Middleware                       │
│     • Request Validation                    │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│     SERVICES LAYER (Go)                     │
├─────────────────────────────────────────────┤
│ • Server Management                         │
│ • Site Deployment                           │
│ • DNS Management                            │
│ • Email Management                          │
│ • Database Management                       │
│ • SSL/Certificate Management                │
│ • Backup Management                         │
│ • Monitoring & Alerts                       │
│ • Billing & Invoicing                       │
└─────────────────────────────────────────────┘
                    ↕
┌──────────────┬──────────────┬───────────────┐
│  PostgreSQL  │    Redis     │   TimescaleDB │
│  (Core Data) │(Cache/Queue) │   (Metrics)   │
└──────────────┴──────────────┴───────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│     WORKER POOL (Go + Asynq)                │
│     • Async job processing                  │
│     • Server provisioning                   │
│     • Backup execution                      │
│     • SSL renewal                           │
│     • Health checks                         │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│     AUTOMATION LAYER                        │
│     Python Scripts + Ansible Playbooks      │
├─────────────────────────────────────────────┤
│ • Server provisioning & hardening           │
│ • Web server configuration                  │
│ • Database installation                     │
│ • Email server setup                        │
│ • Security configuration                    │
│ • Monitoring agent installation             │
└─────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────┐
│     MANAGED SERVERS                         │
│     • Agent (optional): Metrics collection  │
│     • SSH Access: Ansible-based management  │
│     • Services: Nginx, PHP-FPM, MySQL, etc. │
└─────────────────────────────────────────────┘
📁 Project Structure
vip-hosting-panel-v2/
├── cmd/
│   ├── api/                    # Main web server
│   │   └── main.go
│   ├── worker/                 # Background job processor
│   │   └── main.go
│   ├── agent/                  # Server monitoring agent (optional)
│   │   └── main.go
│   └── cli/                    # CLI tools
│       └── main.go
│
├── internal/
│   ├── auth/                   # Authentication & authorization
│   │   ├── jwt.go
│   │   ├── rbac.go
│   │   ├── session.go
│   │   └── twofa.go
│   │
│   ├── handlers/               # HTTP request handlers
│   │   ├── auth.go
│   │   ├── dashboard.go
│   │   ├── servers.go
│   │   ├── sites.go
│   │   ├── databases.go
│   │   ├── email.go
│   │   ├── dns.go
│   │   ├── ssl.go
│   │   ├── backups.go
│   │   ├── monitoring.go
│   │   ├── firewall.go
│   │   ├── users.go
│   │   └── billing.go
│   │
│   ├── services/               # Business logic
│   │   ├── servers/
│   │   │   ├── provisioner.go
│   │   │   ├── hetzner.go
│   │   │   ├── digitalocean.go
│   │   │   └── manager.go
│   │   ├── sites/
│   │   │   ├── deployer.go
│   │   │   ├── templates.go
│   │   │   └── manager.go
│   │   ├── databases/
│   │   │   ├── mysql.go
│   │   │   ├── postgresql.go
│   │   │   ├── mongodb.go
│   │   │   └── redis.go
│   │   ├── email/
│   │   │   ├── postfix.go
│   │   │   ├── dovecot.go
│   │   │   ├── accounts.go
│   │   │   ├── aliases.go
│   │   │   └── spam.go
│   │   ├── dns/
│   │   │   ├── bind9.go
│   │   │   ├── cloudflare.go
│   │   │   ├── route53.go
│   │   │   └── manager.go
│   │   ├── ssl/
│   │   │   ├── acme.go
│   │   │   ├── letsencrypt.go
│   │   │   └── manager.go
│   │   ├── security/
│   │   │   ├── waf.go
│   │   │   ├── firewall.go
│   │   │   ├── fail2ban.go
│   │   │   ├── clamav.go
│   │   │   └── scanner.go
│   │   ├── backups/
│   │   │   ├── scheduler.go
│   │   │   ├── s3.go
│   │   │   └── restore.go
│   │   ├── monitoring/
│   │   │   ├── metrics.go
│   │   │   ├── alerts.go
│   │   │   └── uptime.go
│   │   └── billing/
│   │       ├── stripe.go
│   │       ├── invoices.go
│   │       └── usage.go
│   │
│   ├── models/                 # Database models
│   │   ├── user.go
│   │   ├── server.go
│   │   ├── site.go
│   │   ├── database.go
│   │   ├── email.go
│   │   ├── dns.go
│   │   ├── backup.go
│   │   └── invoice.go
│   │
│   ├── jobs/                   # Background jobs
│   │   ├── server_provisioning.go
│   │   ├── site_deployment.go
│   │   ├── backup_execution.go
│   │   ├── ssl_renewal.go
│   │   ├── health_check.go
│   │   └── cleanup.go
│   │
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go
│   │   ├── rbac.go
│   │   ├── rate_limit.go
│   │   ├── logging.go
│   │   └── cors.go
│   │
│   ├── repository/             # Database access layer
│   │   ├── user.go
│   │   ├── server.go
│   │   ├── site.go
│   │   └── ...
│   │
│   └── utils/                  # Helper functions
│       ├── crypto.go
│       ├── validator.go
│       ├── ssh.go
│       └── helpers.go
│
├── web/
│   ├── templates/              # Templ HTML templates
│   │   ├── layouts/
│   │   │   ├── base.templ
│   │   │   ├── dashboard.templ
│   │   │   └── auth.templ
│   │   ├── pages/
│   │   │   ├── dashboard.templ
│   │   │   ├── servers/
│   │   │   ├── sites/
│   │   │   ├── databases/
│   │   │   ├── email/
│   │   │   ├── dns/
│   │   │   ├── ssl/
│   │   │   ├── backups/
│   │   │   ├── monitoring/
│   │   │   └── settings/
│   │   └── components/
│   │       ├── navbar.templ
│   │       ├── sidebar.templ
│   │       ├── server_card.templ
│   │       ├── site_card.templ
│   │       ├── modals.templ
│   │       └── forms.templ
│   │
│   └── static/
│       ├── css/
│       │   ├── tailwind.css
│       │   └── custom.css
│       ├── js/
│       │   ├── alpine.js
│       │   └── htmx.js
│       ├── images/
│       └── icons/
│
├── automation/
│   ├── playbooks/              # Ansible playbooks
│   │   ├── provision-server.yml
│   │   ├── setup-webserver.yml
│   │   ├── install-php.yml
│   │   ├── install-nodejs.yml
│   │   ├── install-mysql.yml
│   │   ├── install-postgresql.yml
│   │   ├── install-mongodb.yml
│   │   ├── install-redis.yml
│   │   ├── setup-email.yml
│   │   ├── setup-dns.yml
│   │   ├── configure-firewall.yml
│   │   ├── install-waf.yml
│   │   ├── install-fail2ban.yml
│   │   ├── install-clamav.yml
│   │   ├── configure-ssl.yml
│   │   ├── setup-backups.yml
│   │   ├── install-monitoring.yml
│   │   └── security-hardening.yml
│   │
│   ├── roles/                  # Ansible roles
│   │   ├── common/
│   │   ├── webserver/
│   │   ├── database/
│   │   ├── email/
│   │   ├── security/
│   │   └── monitoring/
│   │
│   ├── scripts/                # Python management scripts
│   │   ├── server_manager.py
│   │   ├── site_manager.py
│   │   ├── database_manager.py
│   │   ├── email_manager.py
│   │   ├── dns_manager.py
│   │   ├── ssl_manager.py
│   │   ├── backup_manager.py
│   │   ├── firewall_manager.py
│   │   └── monitoring_agent.py
│   │
│   ├── templates/              # Configuration templates
│   │   ├── nginx/
│   │   │   ├── vhost.conf.j2
│   │   │   ├── php.conf.j2
│   │   │   └── nodejs.conf.j2
│   │   ├── apache/
│   │   │   └── vhost.conf.j2
│   │   ├── php/
│   │   │   └── php.ini.j2
│   │   ├── mysql/
│   │   │   └── my.cnf.j2
│   │   ├── postfix/
│   │   │   ├── main.cf.j2
│   │   │   └── master.cf.j2
│   │   ├── dovecot/
│   │   │   └── dovecot.conf.j2
│   │   ├── fail2ban/
│   │   │   └── jail.local.j2
│   │   └── modsecurity/
│   │       └── custom-rules.conf.j2
│   │
│   └── inventory/              # Ansible inventory
│       ├── production.ini
│       └── staging.ini
│
├── migrations/                 # Database migrations
│   ├── 001_initial_schema.sql
│   ├── 002_add_email_tables.sql
│   ├── 003_add_dns_tables.sql
│   └── ...
│
├── configs/                    # Configuration files
│   ├── config.yaml.example
│   ├── providers.yaml.example
│   └── security.yaml.example
│
├── scripts/                    # Build and deployment scripts
│   ├── build.sh
│   ├── deploy.sh
│   ├── migrate.sh
│   └── setup-dev.sh
│
├── tests/                      # Tests
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── docs/                       # Documentation
│   ├── getting-started.md
│   ├── architecture.md
│   ├── api-reference.md
│   ├── ansible-playbooks.md
│   └── security.md
│
├── docker/                     # Docker configuration
│   ├── api/
│   │   └── Dockerfile
│   ├── worker/
│   │   └── Dockerfile
│   └── agent/
│       └── Dockerfile
│
├── docker-compose.yml          # Development environment
├── docker-compose.prod.yml     # Production setup
├── Makefile                    # Build automation
├── go.mod                      # Go dependencies
├── go.sum
├── package.json                # Frontend dependencies
├── tailwind.config.js          # Tailwind configuration
└── README.md                   # This file
🚦 Quick Start
Prerequisites

Ubuntu 22.04/24.04 or Debian 11/12
Go 1.21+
Node.js 18+
PostgreSQL 15+
Redis 7+
systemd (included in Ubuntu/Debian)
Nginx (as reverse proxy)

Production Installation

Clone the repository

bashgit clone https://github.com/yourusername/vip-hosting-panel-v2.git
cd vip-hosting-panel-v2

Run the automated installer

bashsudo bash scripts/install.sh
This will:

Install all system dependencies (PostgreSQL, Redis, Nginx)
Build the Go binaries
Setup systemd services
Configure firewall
Create database and run migrations
Generate SSL certificates
Start all services

Manual Installation

Install system dependencies

bash# Update system
sudo apt update && sudo apt upgrade -y

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Install Redis
sudo apt install redis-server -y

# Install Nginx
sudo apt install nginx -y

# Install Go (if not installed)
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install nodejs -y

# Install Python dependencies
sudo apt install python3 python3-pip ansible -y
pip3 install -r automation/requirements.txt

Setup database

bashsudo -u postgres psql << EOF
CREATE DATABASE vip_hosting;
CREATE USER vip_panel WITH PASSWORD 'change_this_password';
GRANT ALL PRIVILEGES ON DATABASE vip_hosting TO vip_panel;
\c vip_hosting
GRANT ALL ON SCHEMA public TO vip_panel;
EOF

Copy and configure

bash# Copy config files
cp configs/config.yaml.example configs/config.yaml
cp configs/providers.yaml.example configs/providers.yaml

# Edit configurations
nano configs/config.yaml

Build the application

bashmake build

Install systemd services

bashsudo make install-services

Start services

bashsudo systemctl enable vip-panel-api
sudo systemctl enable vip-panel-worker
sudo systemctl start vip-panel-api
sudo systemctl start vip-panel-worker

Setup Nginx reverse proxy

bashsudo make setup-nginx
sudo systemctl reload nginx

Access the panel

https://your-domain.com or https://your-server-ip
Default credentials:
Email: admin@example.com
Password: admin123
Development Setup
bash# Install dependencies
go mod download
npm install

# Run migrations
make migrate

# Start in development mode
make dev
🔧 Configuration
Main Configuration (configs/config.yaml)
yamlserver:
  host: 0.0.0.0
  port: 3000
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  name: vip_hosting
  user: postgres
  password: postgres
  max_connections: 100

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-change-in-production"
  expiration: 24h

features:
  multi_php: true
  multi_nodejs: true
  multi_database: true
  email_server: true
  dns_management: true
  waf: true
  antivirus: true
Provider Configuration (configs/providers.yaml)
yamlproviders:
  hetzner:
    enabled: true
    api_token: "your-hetzner-api-token"
    default_region: nbg1
    
  digitalocean:
    enabled: false
    api_token: ""
    
  cloudflare:
    enabled: true
    api_token: "your-cloudflare-api-token"
    email: "your-email@example.com"

stripe:
  enabled: true
  secret_key: "sk_test_..."
  public_key: "pk_test_..."
  webhook_secret: "whsec_..."
📚 API Documentation
Authentication
POST /api/auth/login
json{
  "email": "user@example.com",
  "password": "password123"
}
POST /api/auth/register
json{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
Server Management
GET /api/servers - List all servers
POST /api/servers - Create new server
GET /api/servers/:id - Get server details
PUT /api/servers/:id - Update server
DELETE /api/servers/:id - Delete server
POST /api/servers/:id/reboot - Reboot server
POST /api/servers/:id/resize - Resize server
Site Management
GET /api/sites - List all sites
POST /api/sites - Create new site
GET /api/sites/:id - Get site details
PUT /api/sites/:id - Update site
DELETE /api/sites/:id - Delete site
POST /api/sites/:id/deploy - Deploy site
Database Management
GET /api/databases - List all databases
POST /api/databases - Create database
DELETE /api/databases/:id - Delete database
Email Management
GET /api/email/accounts - List email accounts
POST /api/email/accounts - Create email account
GET /api/email/aliases - List aliases
POST /api/email/aliases - Create alias
DNS Management
GET /api/dns/zones - List DNS zones
POST /api/dns/zones - Create DNS zone
GET /api/dns/records - List DNS records
POST /api/dns/records - Create DNS record
🔐 Security Features
WAF Rules
The panel includes ModSecurity with OWASP Core Rule Set. Custom rules can be added per-site:
nginx# Custom WAF rule example
SecRule REQUEST_URI "@contains /admin" \
    "id:1000,phase:1,deny,status:403,msg:'Admin access blocked'"
Firewall Rules
Default firewall rules are automatically configured:

SSH (22) - Restricted to management IPs
HTTP (80) - Open
HTTPS (443) - Open
MySQL (3306) - Localhost only
PostgreSQL (5432) - Localhost only
SMTP (25, 587) - Open
IMAP (143, 993) - Open
POP3 (110, 995) - Open

Fail2ban Configuration
Automatic brute-force protection for:

SSH
FTP
Email (SMTP, IMAP, POP3)
Panel login
WordPress
Custom rules per application

📊 Monitoring
Metrics Collected

CPU usage (per core and total)
Memory usage (used, free, cached)
Disk usage (per partition)
Network traffic (in/out)
Process count
Load average
Database connections
Web server requests
Email queue size

Alert Conditions

CPU > 80% for 5 minutes
Memory > 90%
Disk > 85%
Service down
SSL certificate expiring in < 7 days
Backup failed
Security scan detected malware

🔄 Backup System
Backup Types

Full Backup: Complete server snapshot
Incremental Backup: Changes since last backup
Database Backup: SQL dumps
File Backup: Website files only

Backup Storage

Local storage (on-server)
S3-compatible (AWS, DigitalOcean Spaces, MinIO)
FTP/SFTP remote servers
Multiple destinations per backup

Restore Process

Select backup point
Choose restore type (full, files only, database only)
Select destination (original location or new location)
Verify and confirm
Automated rollback if restore fails

💳 Billing System
Subscription Plans

Tiered pricing based on resources
Custom pricing per reseller
Usage-based add-ons
One-time charges support

Invoice Generation

Automated monthly invoicing
Prorated charges for upgrades/downgrades
VAT/tax calculation
Multi-currency support
PDF generation with custom branding

🎯 Roadmap
v1.0 (Current - Months 1-2)

✅ Core infrastructure
✅ Multi-PHP/Node.js support
✅ Basic security features
✅ Email server
✅ DNS management
✅ Billing integration

v1.1 (Month 3)

 Advanced WAF rules
 Container support (Docker)
 Kubernetes integration
 Advanced analytics
 Mobile app

v1.2 (Month 4)

 White-labeling
 Plugin system
 Marketplace
 Advanced automation
 AI-powered optimization

📖 Documentation

Getting Started Guide
Architecture Overview
API Reference
Ansible Playbooks
Security Best Practices
Troubleshooting

🤝 Contributing
Contributions are welcome! Please read CONTRIBUTING.md for details.
📄 License
This project is licensed under the MIT License - see LICENSE file for details.
🙏 Acknowledgments

HTMX for the hypermedia approach
Templ for type-safe Go templates
Fiber for the blazing-fast web framework
Ansible for infrastructure automation
The open-source community

📞 Support

Documentation: https://docs.superhosting.vip
Community Forum: https://community.superhosting.vip
Discord: https://discord.gg/viphostingpanel
Email: support@vsuperhosting.vip


Made with ❤️ by the VIP Hosting Panel Team