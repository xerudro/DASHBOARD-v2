# CloudPanel - Comprehensive Product Requirements Document (PRD)
## Complete Feature, Function & Functionality Analysis

**Document Version:** 1.0  
**Date:** November 2, 2025  
**Source:** Official CloudPanel Documentation (https://www.cloudpanel.io)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Overview](#system-overview)
3. [Technology Stack](#technology-stack)
4. [Installation & Setup](#installation--setup)
5. [Site Management](#site-management)
6. [Supported Applications](#supported-applications)
7. [Varnish Cache System](#varnish-cache-system)
8. [Database Management](#database-management)
9. [SSL/TLS Management](#ssltls-management)
10. [File & User Management](#file--user-management)
11. [Security Features](#security-features)
12. [Vhost Configuration](#vhost-configuration)
13. [Cron Jobs](#cron-jobs)
14. [Logs & Monitoring](#logs--monitoring)
15. [Admin Area](#admin-area)
16. [Cloud Features](#cloud-features)
17. [Remote Backups](#remote-backups)
18. [Dploy Deployment](#dploy-deployment)
19. [CLI Tools](#cli-tools)
20. [Performance Optimization](#performance-optimization)
21. [Technical Requirements](#technical-requirements)

---

## 1. Executive Summary

CloudPanel is a **free, open-source hosting control panel** designed with an obsessive focus on simplicity and performance. Built for modern cloud infrastructure, it provides a lightweight, fast, and secure solution for managing web servers and applications.

### Key Differentiators
- **Completely Free** - No licensing costs, 100% open source (MIT License for core components)
- **Lightweight Architecture** - Minimal resource consumption (1-Core, 2GB RAM minimum)
- **High Performance Stack** - NGINX, PHP 7.1-8.4, MySQL, Redis, Varnish Cache
- **ARM Ready** - Full support for ARM64 architecture (40% higher performance, 20% lower cost)
- **Multi-Application Support** - PHP, Node.js, Python, Static HTML, Reverse Proxies
- **Varnish Cache Integration** - 100-250x faster page loads with turn-key caching
- **Modern Cloud Integration** - Native support for AWS, Digital Ocean, Google Cloud, Azure, Hetzner, Vultr

### Product Philosophy
- **Simplicity First** - Complex operations made simple through intuitive UI
- **Performance Obsessed** - Optimized stack for maximum speed
- **Developer Friendly** - CLI tools, SSH access, Git integration
- **Security Focused** - Site isolation, UFW firewall, 2FA, Let's Encrypt

---

## 2. System Overview

### 2.1 Architecture Model

**Single-Server Design:**
CloudPanel operates on a single-server model, where all components run on one instance:

```
┌─────────────────────────────────────────────────────┐
│                CloudPanel Instance                  │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │          CloudPanel UI (Port 8443)         │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     NGINX Web Server (Ports 80/443)        │    │
│  │     ├─ SSL/TLS Termination                │    │
│  │     ├─ Static File Serving                │    │
│  │     └─ Reverse Proxy Layer                │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │      Varnish Cache (Port 6081)             │    │
│  │      └─ In-Memory Page Caching             │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     PHP-FPM (Multiple Versions)            │    │
│  │     ├─ PHP 7.1 - 8.4                       │    │
│  │     ├─ Per-Site PHP Version                │    │
│  │     └─ OPcache Enabled                     │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     MySQL 8.0 / MariaDB                    │    │
│  │     └─ Database Server                     │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     Node.js (via NVM)                      │    │
│  │     └─ Multiple Node.js Versions           │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     Python Applications                    │    │
│  │     └─ Python Virtual Environments         │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     Redis (In-Memory Cache)                │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
│  ┌───────────────────────────────────────────┐    │
│  │     ProFTPD (FTP/SFTP Server)              │    │
│  └───────────────────────────────────────────┘    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 2.2 Site Isolation

**Multi-Tenancy Architecture:**
- Each site runs under a dedicated Unix user account
- Complete file system isolation
- Separate PHP-FPM pools per site
- Individual log files per site
- Resource limits per site user

**Security Benefits:**
- Compromised site cannot affect other sites
- Prevents cross-site file access
- Contains security breaches
- Facilitates site-specific resource management

### 2.3 Target Users
- **Developers** - Fast project deployment and testing
- **Digital Agencies** - Multiple client site management
- **Web Hosting Users** - Personal or small business hosting
- **IT Administrators** - Server and application management
- **Infrastructure Providers** - Cloud hosting services

---

## 3. Technology Stack

### 3.1 Core Components

**Operating Systems:**
- **Ubuntu** 24.04 LTS, 22.04 LTS
- **Debian** 12 (Bookworm), 11 (Bullseye)
- **Architecture Support:** x86_64 and ARM64 (aarch64)

**Web Server:**
- **NGINX** (Latest Stable)
  - High-performance edge web server
  - Minimal memory footprint
  - Event-based architecture
  - 10x faster than Apache for static content
  - Concurrent connection handling
  - Built-in reverse proxy capability
  - HTTP/3 support (experimental on Ubuntu 24.04/Debian 12)

**PHP:**
- **Multiple Concurrent Versions:** PHP 7.1 - 8.4
- **Frameworks Supported:**
  - Laravel
  - Symfony
  - CodeIgniter
  - CakePHP
  - Slim
  - And all other PHP frameworks
- **OPcache:** Bytecode caching enabled by default
- **Composer:** Dependency management tool included

**Database:**
- **MySQL 8.0**
  - ACID compliance
  - InnoDB storage engine
  - Full-text search
  - Replication support
- **MariaDB 10.6+**
  - MySQL fork with enhancements
  - Better performance
  - Additional storage engines
  - Compatible with MySQL

**Caching:**
- **Varnish Cache 7.x**
  - HTTP reverse proxy cache
  - 100-250x performance improvement
  - In-memory page caching
  - Intelligent cache invalidation
  - Support for major PHP applications
  
- **Redis 7.x**
  - In-memory data structure store
  - Object caching
  - Session storage
  - Pub/Sub messaging
  - LRU eviction
  - Persistence options

**Node.js:**
- **Node Version Manager (NVM)**
  - Multiple Node.js versions
  - Per-site version selection
  - Easy version switching
  - npm and Yarn package managers

**Python:**
- **Python 3.x**
  - Virtual environment support
  - pip package manager
  - Flask, Django support
  - Async/await support

**FTP Server:**
- **ProFTPD**
  - GPL-licensed FTP server
  - Highly configurable
  - Support for FTP and FTPS
  - Virtual users support
  - Bandwidth throttling

**Package Managers:**
- **Yarn** - Fast, reliable JavaScript package management
- **Composer** - PHP dependency manager
- **npm** - Node.js package manager
- **pip** - Python package installer

### 3.2 System Architecture Benefits

**Lightweight Design:**
- Minimal memory overhead
- Fast boot times
- Low CPU usage
- Efficient disk I/O
- Optimal price-performance ratio

**Performance Characteristics:**
- Event-driven NGINX
- OPcache for PHP acceleration
- Varnish for page caching
- Redis for object caching
- Optimized MySQL/MariaDB configurations

**Security Features:**
- Site-level user isolation
- UFW firewall integration
- Let's Encrypt SSL automation
- Two-factor authentication
- Regular security updates

---

## 4. Installation & Setup

### 4.1 System Requirements

**Minimum Specifications:**
- **CPU:** 1 Core (physical or virtual)
- **RAM:** 2 GB
- **Storage:** 10 GB disk space
- **Architecture:** x86_64 or ARM64
- **Network:** Public IP address with internet access

**Recommended Specifications:**
- **CPU:** 2+ Cores
- **RAM:** 4 GB+
- **Storage:** 20 GB+ SSD
- **Network:** 100 Mbps+ bandwidth

**Operating System:**
- Ubuntu 24.04 LTS (recommended)
- Ubuntu 22.04 LTS
- Debian 12 (Bookworm)
- Debian 11 (Bullseye)
- **Clean installation required** (no pre-existing web services)

### 4.2 Supported Cloud Providers

CloudPanel provides one-click installation or marketplace images for:

**Amazon Web Services (AWS):**
- EC2 instances
- AMI (Amazon Machine Image) available
- Automated backup via AMIs
- Elastic IP support

**Digital Ocean:**
- Droplets
- Marketplace app available
- Automated backups via Spaces
- Floating IP support

**Hetzner Cloud:**
- Cloud servers
- Installer script available
- Snapshot support
- Flexible networking

**Google Compute Engine (GCE):**
- VM instances
- Installer script available
- Automated snapshots
- Cloud Storage integration

**Microsoft Azure:**
- Virtual Machines
- Installer script available
- Azure Backup integration
- Managed disks support

**Vultr:**
- Cloud Compute instances
- Marketplace app available
- Snapshot support
- Block storage

**Other Providers:**
- Any VPS or dedicated server
- Self-hosted infrastructure
- Private cloud environments

### 4.3 Installation Process

**One-Command Installation:**

**Ubuntu 24.04 / 22.04:**
```bash
curl -sS https://installer.cloudpanel.io/ce/v2/install.sh -o install.sh; \
echo "3b639332b2a7e9f2c05e3e1b1e90d29ae3bd9f98bf8d4c60cbb78a966b32e3e4 install.sh" | \
sha256sum -c && sudo bash install.sh
```

**Debian 12:**
```bash
curl -sS https://installer.cloudpanel.io/ce/v2/install.sh -o install.sh; \
echo "3b639332b2a7e9f2c05e3e1b1e90d29ae3bd9f98bf8d4c60cbb78a966b32e3e4 install.sh" | \
sha256sum -c && sudo bash install.sh
```

**Debian 11:**
```bash
curl -sS https://installer.cloudpanel.io/ce/v2/install.sh -o install.sh; \
echo "3b639332b2a7e9f2c05e3e1b1e90d29ae3bd9f98bf8d4c60cbb78a966b32e3e4 install.sh" | \
sha256sum -c && sudo bash install.sh
```

**Installation Steps:**
1. **Verify System:** Clean OS installation, root access
2. **Download Installer:** Use curl to fetch installation script
3. **Verify Checksum:** SHA256 verification for security
4. **Execute Installation:** Run as root/sudo
5. **Automated Setup:**
   - Install system dependencies
   - Configure NGINX
   - Install PHP versions (7.1-8.4)
   - Setup MySQL/MariaDB
   - Install Redis, Varnish
   - Configure firewall (UFW)
   - Create CloudPanel admin user
6. **Access Interface:** https://server-ip:8443

**Installation Time:** 5-10 minutes (depending on server speed and internet connection)

### 4.4 Initial Configuration

**First Access:**
1. Navigate to `https://your-server-ip:8443`
2. Create admin account
3. Set admin password (strong password required)
4. Configure timezone
5. Complete initial setup wizard

**Post-Installation Tasks:**
- Configure custom CloudPanel domain (optional)
- Setup Let's Encrypt SSL for CloudPanel interface
- Configure firewall rules
- Create first site/user
- Configure remote backups (optional)
- Setup external database server (optional)

### 4.5 Cloud-Specific Features

**AWS Installation:**
- Use AMI from AWS Marketplace
- Automatic IAM role configuration
- Elastic IP association
- Security group setup

**Digital Ocean:**
- One-click app from marketplace
- Automatic floating IP support
- Droplet snapshot integration

**Hetzner Cloud:**
- Installer script
- Snapshot scheduling
- Private networking support

**Google Cloud:**
- Installer script
- Snapshot automation
- Cloud Storage integration

---

## 5. Site Management

### 5.1 Site Types

CloudPanel supports six different site types:

**1. WordPress Sites:**
- One-click WordPress installation
- Latest WordPress version
- Pre-configured for optimal performance
- Database auto-creation
- Admin credentials generated
- WP-CLI available

**2. PHP Sites:**
- Generic PHP application hosting
- Support for all PHP frameworks
- Configurable document root
- Multiple PHP version support
- Composer integration

**3. Node.js Sites:**
- Node.js application hosting
- NVM for version management
- Custom app port configuration
- PM2 process management
- npm and Yarn support

**4. Python Sites:**
- Python application hosting
- Virtual environment support
- Custom app port configuration
- pip package management
- WSGI/ASGI support

**5. Static HTML Sites:**
- Pure HTML/CSS/JS hosting
- No server-side processing
- Optimal NGINX configuration
- CDN-ready

**6. Reverse Proxy:**
- Proxy to internal services
- SSL termination
- Load balancing capability
- WebSocket support

### 5.2 Creating Sites

**Via CloudPanel UI:**

**WordPress Site Creation:**
1. Click "+ Add Site"
2. Select "Create a WordPress Site"
3. Configure:
   - Domain Name (with/without www)
   - Site User (SSH/SFTP username)
   - Site User Password
   - PHP Version (7.1-8.4)
4. Click "Create"
5. Receive WordPress admin credentials
6. Access site immediately

**PHP Site Creation:**
1. Click "+ Add Site"
2. Select "Create a PHP Site"
3. Configure:
   - Application template (Laravel, Symfony, CakePHP, etc.)
   - Domain Name
   - PHP Version
   - Site User credentials
4. Click "Create"
5. Upload files via SSH/SFTP or use Composer

**Node.js Site Creation:**
1. Click "+ Add Site"
2. Select "Create a Node.js Site"
3. Configure:
   - Domain Name
   - Node.js Version
   - App Port (3000, 8000, etc.)
   - Site User credentials
4. Click "Create"
5. Upload application files
6. Configure start script

**Python Site Creation:**
1. Click "+ Add Site"
2. Select "Create a Python Site"
3. Configure:
   - Domain Name
   - Python Version
   - App Port
   - Site User credentials
4. Click "Create"
5. Upload application
6. Configure WSGI/ASGI

**Static HTML Site Creation:**
1. Click "+ Add Site"
2. Select "Create a Static HTML Site"
3. Configure:
   - Domain Name
   - Site User credentials
4. Click "Create"
5. Upload HTML/CSS/JS files

**Reverse Proxy Creation:**
1. Click "+ Add Site"
2. Select "Create a Reverse Proxy"
3. Configure:
   - Domain Name
   - Reverse Proxy URL (internal service)
   - Site User credentials
4. Click "Create"

**Via CloudPanel CLI:**

**WordPress:**
```bash
clpctl site:add:wordpress --domainName=www.example.com \
  --siteUser='john-doe' --siteUserPassword='secure123!' \
  --phpVersion=8.3
```

**PHP:**
```bash
clpctl site:add:php --domainName=www.example.com \
  --phpVersion=8.4 --vhostTemplate='Laravel 12' \
  --siteUser='john-doe' --siteUserPassword='secure123!'
```

**Node.js:**
```bash
clpctl site:add:nodejs --domainName=www.example.com \
  --nodejsVersion=20 --appPort=3000 \
  --siteUser='john-doe' --siteUserPassword='secure123!'
```

**Python:**
```bash
clpctl site:add:python --domainName=www.example.com \
  --pythonVersion=3.11 --appPort=8000 \
  --siteUser='john-doe' --siteUserPassword='secure123!'
```

**Static HTML:**
```bash
clpctl site:add:static --domainName=www.example.com \
  --siteUser='john-doe' --siteUserPassword='secure123!'
```

**Reverse Proxy:**
```bash
clpctl site:add:reverse-proxy --domainName=www.example.com \
  --reverseProxyUrl='http://127.0.0.1:8080' \
  --siteUser='john-doe' --siteUserPassword='secure123!'
```

### 5.3 Automatic Features

**Automatic Redirections:**
- **www Handling:**
  - Enter domain with www: redirects non-www to www
  - Enter domain without www: redirects www to non-www
- **HTTP to HTTPS:** All HTTP requests automatically redirected to HTTPS (301 permanent)
- **SEO Friendly:** Proper redirect codes used

**Automatic SSL:**
- Self-signed certificate generated on site creation
- Can be replaced with Let's Encrypt certificate
- Free SSL for all sites
- Automatic renewal (Let's Encrypt)

**Site User Benefits:**
- Isolated Unix user per site
- SSH/SFTP access
- Home directory: `/home/$siteUser`
- Separate PHP-FPM pool
- Individual resource limits
- Isolated logs

### 5.4 Site Management Operations

**PHP Version Changes:**
- Switch PHP version without downtime
- Available versions: 7.1 - 8.4
- Change via UI or CLI
- PHP-FPM pool automatically reconfigured

**Document Root Changes:**
- Modify document root path
- Common for frameworks (e.g., Laravel uses /public)
- Update via Vhost editor
- No downtime required

**Site Deletion:**
- Complete site removal
- Deletes user, files, databases
- Backup recommended before deletion
- Irreversible operation

**Site Disabling:**
- Temporary site suspension
- Maintenance mode
- Preserves all data
- Can be re-enabled

---

## 6. Supported Applications

CloudPanel includes pre-configured templates for 20+ applications. Each template provides optimized NGINX vhost configuration, PHP settings, and rewrite rules.

### 6.1 Content Management Systems

**1. WordPress (Latest Version):**
- **Features:**
  - One-click installation
  - WP-CLI pre-installed
  - Optimized NGINX configuration
  - Varnish Cache support
  - Multi-site support
  - Auto-updates support
- **PHP Versions:** 7.4 - 8.4
- **Installation:**
  ```bash
  clpctl site:add:wordpress --domainName=www.domain.com \
    --siteUser='john-doe' --siteUserPassword='pass123!' \
    --phpVersion=8.3
  ```

**2. Joomla 5:**
- **Features:**
  - Pre-configured rewrite rules
  - SEO-friendly URLs
  - Admin panel optimizations
  - Cache configuration
- **PHP Versions:** 8.1 - 8.3
- **Installation Process:**
  1. Create PHP site with Joomla 5 template
  2. Download and extract Joomla
  3. Create database
  4. Run web installer
- **Commands:**
  ```bash
  mkdir ~/tmp/joomla
  curl -Lso joomla.tar.gz https://downloads.joomla.org/cms/joomla5/latest
  tar xf joomla.tar -C ~/tmp/joomla/
  cp -R ~/tmp/joomla/* ~/htdocs/www.domain.com/
  ```

**3. Drupal 11:**
- **Features:**
  - Clean URLs support
  - Private files directory
  - Cron configuration
  - Admin toolbar optimization
- **PHP Versions:** 8.2 - 8.4
- **Installation via Composer:**
  ```bash
  php8.4 /usr/local/bin/composer create-project \
    drupal/recommended-project:^11 www.domain.com
  ```

**4. TYPO3 (Latest):**
- **Features:**
  - Backend routing
  - Install tool access
  - Scheduler configuration
  - Image processing
- **PHP Versions:** 8.1 - 8.4

### 6.2 E-Commerce Platforms

**5. WooCommerce:**
- **Features:**
  - WordPress + WooCommerce
  - Varnish Cache support
  - Cart/checkout exclusions
  - Payment gateway optimizations
  - Product image handling
- **PHP Versions:** 7.4 - 8.4
- **Installation:** Same as WordPress + WooCommerce plugin

**6. Magento 2:**
- **Features:**
  - Production mode configuration
  - Static content serving
  - Media cache handling
  - Cron job setup
  - Composer authentication
- **PHP Versions:** 8.1 - 8.3
- **Installation via Composer:**
  ```bash
  php8.3 /usr/local/bin/composer create-project \
    --repository-url=https://repo.magento.com/ \
    magento/project-community-edition www.domain.com
  ```
- **Setup Script:**
  ```bash
  php8.3 bin/magento setup:install \
    --base-url=https://www.domain.com \
    --db-host=127.0.0.1 \
    --db-name=database_name \
    --db-user=database_user \
    --db-password=database_password \
    --admin-firstname=Admin \
    --admin-lastname=User \
    --admin-email=admin@domain.com \
    --admin-user=admin \
    --admin-password=Admin123! \
    --language=en_US \
    --currency=USD \
    --timezone=America/New_York
  ```

**7. PrestaShop 1.7:**
- **Features:**
  - Pretty URLs
  - Image rewrite rules
  - Admin panel security
  - REST API support
- **PHP Versions:** 7.3 - 8.1
- **Installation:**
  ```bash
  curl -sL https://github.com/PrestaShop/PrestaShop/archive/1.7.8.7.tar.gz | tar xfz -
  ```

**8. Shopware 6:**
- **Features:**
  - Storefront optimization
  - Admin panel routing
  - API endpoints
  - Media handling
- **PHP Versions:** 8.1 - 8.4

### 6.3 PHP Frameworks

**9. Laravel 11/12:**
- **Features:**
  - Public directory as document root
  - Artisan command support
  - Queue worker configuration
  - Scheduler setup
  - Varnish Cache support
- **PHP Versions:** 8.2 - 8.4
- **Installation:**
  ```bash
  php8.4 /usr/local/bin/composer create-project \
    laravel/laravel:^12.0 www.domain.com
  ```
- **Post-Installation:**
  ```bash
  cd ~/htdocs/www.domain.com
  php8.4 artisan key:generate
  php8.4 artisan migrate
  ```

**10. Symfony 7:**
- **Features:**
  - Public directory routing
  - Development/production environments
  - Asset compilation
  - Console command access
  - Varnish Cache support
- **PHP Versions:** 8.2 - 8.4
- **Installation:**
  ```bash
  php8.4 /usr/local/bin/composer create-project \
    symfony/skeleton:^7.0 www.domain.com
  ```

**11. CakePHP 5:**
- **Features:**
  - Webroot as document root
  - Debug mode configuration
  - Asset handling
  - Cache configuration
  - Varnish Cache support
- **PHP Versions:** 8.1 - 8.4
- **Installation:**
  ```bash
  php8.4 /usr/local/bin/composer create-project \
    --prefer-dist cakephp/app:~5.0 www.domain.com
  ```

**12. CodeIgniter 4:**
- **Features:**
  - Public directory routing
  - Environment configuration
  - Spark CLI support
  - Cache drivers
  - Varnish Cache support
- **PHP Versions:** 7.4 - 8.4

**13. Slim 4:**
- **Features:**
  - Minimalist framework setup
  - Public directory routing
  - Middleware support
  - PSR-7 compliance
  - Varnish Cache support
- **PHP Versions:** 7.4 - 8.4
- **Installation:**
  ```bash
  php8.4 /usr/local/bin/composer create-project \
    slim/slim-skeleton:^4 www.domain.com
  ```

### 6.4 Collaboration & Cloud Storage

**14. Nextcloud 31:**
- **Features:**
  - WebDAV support
  - Desktop/mobile sync
  - File sharing
  - Calendar/Contacts
  - Office document editing
- **PHP Versions:** 8.2 - 8.4
- **Installation:**
  ```bash
  curl -sLo nextcloud.zip https://download.nextcloud.com/server/releases/latest.zip
  unzip nextcloud.zip -d nextcloud
  cp -R nextcloud/* ~/htdocs/www.domain.com/
  ```

**15. OwnCloud 10:**
- **Features:**
  - File synchronization
  - File sharing
  - WebDAV access
  - Desktop clients
- **PHP Versions:** 7.4 - 8.1
- **Installation:**
  ```bash
  curl -sL https://download.owncloud.com/server/stable/owncloud-complete-latest.tar.bz2 | tar xfj -
  cp -R owncloud/* ~/htdocs/www.domain.com/
  ```

### 6.5 Other CMS Platforms

**16. Neos 9:**
- **Features:**
  - Modern content management
  - Inline editing
  - Media management
  - Multi-language support
- **PHP Versions:** 8.2 - 8.4
- **Installation:**
  ```bash
  php8.4 /usr/local/bin/composer create-project \
    neos/neos-base-distribution:^9.0 www.domain.com
  ```

### 6.6 Generic Templates

**17. Generic PHP:**
- **Features:**
  - Basic PHP configuration
  - No framework assumptions
  - Customizable document root
  - Universal PHP app support
- **PHP Versions:** 7.1 - 8.4
- **Use Cases:**
  - Custom PHP applications
  - Legacy applications
  - Proprietary systems
  - Micro-frameworks

### 6.7 Application Template Benefits

**Pre-Configured Settings:**
- Optimized NGINX vhost configuration
- Framework-specific rewrite rules
- Security directives (block sensitive files)
- Performance optimizations
- Cache headers
- Gzip compression

**Security Features:**
- Block access to sensitive files (.env, .git, etc.)
- Admin panel protection
- Configuration file restrictions
- Directory listing disabled

**Performance Optimizations:**
- Static file caching
- Gzip/Brotli compression
- Browser cache headers
- CDN-ready configurations

---

## 7. Varnish Cache System

### 7.1 Overview

**Varnish Cache** is a powerful HTTP reverse proxy integrated into CloudPanel as a turn-key solution, providing dramatic performance improvements.

**Performance Benefits:**
- **100-250x faster page loads**
- **80% infrastructure cost savings**
- **Improved user experience**
- **Better search engine rankings** (faster sites rank higher)
- **Reduced server load**

**How It Works:**
```
User Request → NGINX (SSL/TLS) → Varnish Cache (Port 6081) → NGINX (Port 8080) → PHP-FPM
                                       ↓
                                  Cache Hit?
                                       ↓
                                   Yes → Return from memory
                                   No → Forward to PHP
```

**Cache Storage:**
- Pages stored in RAM after first visit
- Compressed page source cached
- Static files (CSS, JS, images) served directly by NGINX (not cached in Varnish)
- Intelligent cache invalidation

### 7.2 Supported Applications

Currently, Varnish Cache has first-class support for:

1. **WordPress** - Full support with automatic purging
2. **WooCommerce** - Cart/checkout exclusions
3. **Laravel** - Framework integration
4. **Symfony** - Controller-based caching
5. **CodeIgniter** - Page caching
6. **Slim** - Framework support
7. **Generic PHP** - Custom application support

**Continuous Development:**
- Additional applications being added
- Community contributions welcome
- Join Discord for feature requests

### 7.3 Enabling Varnish Cache

**Via CloudPanel UI:**
1. Navigate to site
2. Click "Varnish Cache" tab
3. Click "Enable Varnish Cache"
4. Configure settings:
   - Cache Tag Prefix
   - Excluded Parameters
   - Excludes (paths/files)
   - Cache Lifetime
5. Click "Save"

**Automatic Configuration:**
- PHP controller file added automatically
- Vhost updated with Varnish proxy
- Cache rules applied
- No code changes required

### 7.4 Cache Settings

**Cache Tag Prefix:**
- Main cache tag identifier
- Used as prefix for all cache tags
- Auto-generated unique identifier
- Format: 4-character alphanumeric (e.g., "aac6")
- Used for selective cache purging

**Cache Lifetime:**
- Default: 604800 seconds (7 days)
- Configurable per application
- Can be overridden in PHP controller
- Controlled via Cache-Control headers

**Excluded Parameters:**
- GET parameters that disable caching
- Default exclusions:
  - noCache=1
  - __SID (session IDs)
  - utm_* (marketing parameters)
  - fbclid (Facebook click IDs)
- Custom parameters can be added

**Example Excluded URLs:**
```
https://www.domain.com/?noCache=1  ← NOT cached
https://www.domain.com/?page=1&__SID=xyz123  ← NOT cached
https://www.domain.com/?page=1  ← Cached
```

**Excludes (Paths/Files):**
Specify paths or files that should never be cached:
- `^/my-account/` - URLs starting with /my-account/
- `/cart/` - URLs containing /cart/
- `wp-login.php` - Specific file
- `^/admin/` - Admin areas
- `/checkout/` - E-commerce checkout

**WordPress/WooCommerce Default Exclusions:**
- `/wp-admin/` - Admin panel
- `wp-login.php` - Login page
- `/cart/` - Shopping cart
- `/checkout/` - Checkout process
- `/my-account/` - User accounts

### 7.5 Cache Purging

**Purge Entire Cache:**
- Removes all cached pages for a site
- Button: "Purge Entire Cache"
- Under the hood: PURGE request with X-Cache-Tags header
- Command:
  ```bash
  curl -v -X PURGE -H 'X-Cache-Tags: aac6' 127.0.0.1:6081
  ```

**Selective Purging:**
- Purge specific URLs
- Purge by cache tags
- Comma-separated tag list
- Example: Purge products with specific IDs

**Automatic Purging:**
- WordPress: Post/page updates trigger purge
- WooCommerce: Product updates, order changes
- Laravel/Symfony: Model updates (configurable)
- Custom triggers in PHP controller

### 7.6 Varnish Architecture Details

**Request Flow:**

**Static Files (CSS, JS, Images):**
```
User → NGINX → Serve directly from disk
```

**Dynamic Content (Cached):**
```
User → NGINX (SSL/TLS) → Varnish Cache → Return from memory
```

**Dynamic Content (Not Cached):**
```
User → NGINX (SSL/TLS) → Varnish → NGINX (Port 8080) → PHP-FPM → Database
```

**Port Configuration:**
- **443/80:** NGINX SSL/TLS termination
- **6081:** Varnish Cache
- **8080:** NGINX backend (PHP processing)

### 7.7 PHP Controller

**Controller File Location:**
```
/home/$siteUser/.varnish-cache/controller.php
```

**Auto-Prepend:**
Controller added via `auto_prepend_file` directive in NGINX:
```nginx
fastcgi_param PHP_VALUE "auto_prepend_file=/home/siteUser/.varnish-cache/controller.php";
```

**Controller Responsibilities:**
1. **Determine if page can be cached**
2. **Set cache lifetime** for specific pages
3. **Add cache tags** for selective purging
4. **Exclude specific paths/conditions**
5. **Purge cache** on content changes

**Headers Sent by Controller:**
- `X-Cache-Lifetime` - How long to cache (seconds)
- `X-Cache-Tags` - Tags for selective purging
- `Cache-Control` - Browser and Varnish caching directive

**Example Headers:**
```
Cache-Control: public, max-age=604800, s-maxage=604800
X-Cache-Lifetime: 604800
X-Cache-Tags: aac6,aac6-post-123,aac6-home
```

**Response Headers (Visible to User):**
- `x-cache-age` - Age of cached page in seconds
- `x-cache-lifetime` - Remaining seconds before expiration
- `x-cache-tags` - Cache tags for this page

**Custom Applications:**

Set custom cache lifetime:
```php
ClpVarnish::setCacheLifetime(600); // 10 minutes
```

Add custom cache tags:
```php
$cacheTagPrefix = ClpVarnish::getCacheTagPrefix();
$cacheTag = sprintf('%s-%s', $cacheTagPrefix, 'product-456');
ClpVarnish::addCacheTag($cacheTag);
```

**Developer Mode:**

Enable purge logging:
```php
define('VARNISH_DEVELOPER_MODE', true);
```

View purge log:
```bash
tail -f ~/logs/varnish-cache/purge.log -n1000
```

### 7.8 WordPress/WooCommerce Specifics

**Automatic Purging Triggers:**
- Post/page publish or update
- Comment approval
- Theme change
- Plugin activation/deactivation
- Media upload
- Menu update

**WooCommerce Triggers:**
- Product update
- Order status change
- Stock level change
- Price update

**Excluded by Default:**
- Admin panel (/wp-admin/)
- Login page (wp-login.php)
- Cart (/cart/)
- Checkout (/checkout/)
- My Account (/my-account/)
- AJAX requests

**Logged-in Users:**
- No caching for logged-in users
- Cookie detection (wordpress_logged_in_*)
- Ensures personalized experience

---

## 8. Database Management

### 8.1 Database Engines

**Supported Databases:**
- MySQL 8.0
- MariaDB 10.6+

**Installation:**
- Database server installed during CloudPanel installation
- Runs on localhost (127.0.0.1)
- Port 3306
- Root credentials generated

### 8.2 Database Operations

**Adding a Database:**
1. Navigate to site → Databases tab
2. Click "Add Database"
3. Enter:
   - Database Name
   - Database User Name
   - Database User Password
4. Click "Add Database"

**Database Naming:**
- Alphanumeric characters
- Underscores allowed
- No spaces
- Case-sensitive (Linux)

**Database User Management:**

**Adding Database User:**
1. Databases tab
2. Click "Add Database User"
3. Configure:
   - Database User Name
   - Database User Password
   - Select Database
   - Permissions (SELECT, INSERT, UPDATE, DELETE, etc.)
4. Click "Add"

**Permissions Options:**
- SELECT - Read data
- INSERT - Add new rows
- UPDATE - Modify existing data
- DELETE - Remove data
- CREATE - Create tables
- DROP - Delete tables
- INDEX - Create indexes
- ALTER - Modify table structure
- ALL PRIVILEGES - Full access

**Deleting Database/User:**
- Click "Delete" button
- Confirm action
- Irreversible operation
- **Warning:** All data permanently lost

### 8.3 phpMyAdmin

**Access:**
- Click "Manage" button next to database
- Automatic single sign-on
- No manual credentials needed
- Opens in new browser tab

**Features:**
- Database structure browser
- SQL query interface
- Table operations
- Import/Export (⚠️ not recommended)
- User management
- Server status

**⚠️ Important Warning:**
- **DO NOT use phpMyAdmin for export/import**
- Can destroy database with large datasets
- Timeout issues
- Memory problems
- Use CLI tools instead

### 8.4 Database Export

**Using clpctl (Recommended):**

Login via SSH:
```bash
ssh site-user@server-ip
```

Navigate to export directory:
```bash
cd ~/tmp/
```

Export database:
```bash
clpctl db:export --databaseName=my-database --file=dump.sql.gz
```

**Compression:**
- `.sql.gz` - Gzipped (compressed)
- `.sql` - Uncompressed

**Export Process:**
- Uses `mysqldump` utility
- Optimal for large databases
- No memory issues
- Can be scheduled via cron

**Export Options:**
```bash
# Compressed export
clpctl db:export --databaseName=wordpress_db --file=wordpress_backup.sql.gz

# Uncompressed export
clpctl db:export --databaseName=wordpress_db --file=wordpress_backup.sql
```

### 8.5 Database Import

**Using clpctl (Recommended):**

Navigate to dump directory:
```bash
cd ~/tmp/
```

Import database:
```bash
clpctl db:import --databaseName=my-database --file=dump.sql.gz
```

**Import Process:**
- Uses `mysql` command-line client
- Handles large files
- Supports compressed files (.sql.gz)
- No timeout issues

**Import Steps:**
1. Create target database (if needed)
2. Upload dump file via SFTP
3. Run import command
4. Verify import success

### 8.6 Database Backups

**Automatic Backups:**
- **Schedule:** Every night at 3:15 AM
- **Retention:** 7 days by default
- **Location:** `/home/$site-user/backups/`
- **Format:** Gzipped SQL dumps
- **Naming:** `database-name_YYYY-MM-DD_HH-MM-SS.sql.gz`

**Customizing Backup Schedule:**

Edit cron job:
```bash
sudo nano /etc/cron.d/clp
```

Modify backup line:
```cron
# Default (once daily at 3:15 AM)
15 3 * * * clp /usr/bin/bash -c "/usr/bin/clpctl db:backup --ignoreDatabases='db1,db2' --retentionPeriod=7" &> /dev/null

# Twice daily (3:15 AM and 3:15 PM)
15 3,15 * * * clp /usr/bin/bash -c "/usr/bin/clpctl db:backup --retentionPeriod=7" &> /dev/null

# Every 6 hours
15 */6 * * * clp /usr/bin/bash -c "/usr/bin/clpctl db:backup --retentionPeriod=7" &> /dev/null
```

**Backup Options:**
- `--ignoreDatabases='db1,db2'` - Exclude specific databases
- `--retentionPeriod=7` - Days to keep backups (default: 7)

**Changing Retention Period:**
```bash
# Keep backups for 30 days
--retentionPeriod=30

# Keep backups for 14 days
--retentionPeriod=14
```

**Backup File Access:**
```bash
# List backups
ls -lah /home/site-user/backups/

# Download specific backup via SFTP
sftp site-user@server-ip
get /home/site-user/backups/database_2025-11-02_03-15-00.sql.gz
```

### 8.7 Database Master Credentials

**Retrieving Root Credentials:**

SSH as root:
```bash
sudo su root
```

Show credentials:
```bash
clpctl db:show:master-credentials
```

**Output Example:**
```
Host:       127.0.0.1
Port:       3306
User:       root
Password:   XyZ123AbC456DeF
Connect Command: mysql -u root -p'XyZ123AbC456DeF' -h 127.0.0.1
```

**Use Cases:**
- Direct database access
- Advanced administration
- Replication setup
- Third-party tool configuration

### 8.8 Remote Database Access

**Setup:**
1. Admin Area → Security → Firewall
2. Add rule:
   - Name: MySQL Remote
   - Port: 3306
   - Source: Your IP address
   - Action: Allow
3. Save rule

**Security Considerations:**
- Whitelist specific IPs only
- Use strong passwords
- Consider SSH tunneling instead
- Enable firewall rule temporarily

**SSH Tunnel (Recommended):**
```bash
# Create SSH tunnel
ssh -L 3307:127.0.0.1:3306 site-user@server-ip

# Connect via tunnel
mysql -h 127.0.0.1 -P 3307 -u database_user -p
```

### 8.9 External Database Server

**Use Cases:**
- Managed database services (AWS RDS, Digital Ocean Managed MySQL)
- Better performance
- Automatic backups
- Point-in-time recovery
- High availability

**Supported Versions:**
- MySQL 5.7
- MySQL 8.0
- MariaDB 10.6+

**Configuration:**
1. Admin Area → Settings → Database
2. Enter:
   - Database Host
   - Database Port (default: 3306)
   - Database Root User
   - Database Root Password
3. Test connection
4. Save settings

**Benefits:**
- Offload database processing
- Professional database management
- Automated backups
- Scalability
- High availability (if using cloud provider features)

---

## 9. SSL/TLS Management

### 9.1 Default Certificate

**Self-Signed Certificate:**
- Generated automatically on site creation
- Provides HTTPS immediately
- Browser warning (not trusted)
- Development/testing use
- Should be replaced for production

### 9.2 Let's Encrypt Certificates

**Free SSL Certificates:**
- 100% free forever
- Trusted by all browsers
- Automatic renewal
- Multi-domain support (SAN)
- Wildcard support (requires DNS validation)

**Issuing Let's Encrypt:**
1. Site → SSL/TLS tab
2. Click Actions → New Let's Encrypt Certificate
3. Add domain names:
   - Main domain: www.domain.com
   - Add alternative names: domain.com
4. Click "Create and Install"
5. Certificate issued in seconds

**DNS Requirements:**
- Valid A record pointing to server IP
- DNS propagation completed
- Port 80 accessible for validation

**Certificate Details:**
- Validity: 90 days
- Auto-renewal: Every 60 days
- Algorithm: ECDSA or RSA
- Chain: Full trust chain included

**Multiple Domains:**
```
www.domain.com
domain.com
subdomain.domain.com
another.domain.com
```

**Renewal:**
- Automatic via cron job
- Runs daily
- Renews certificates expiring in <30 days
- No manual intervention needed

### 9.3 Cloudflare Integration

**Cloudflare SSL:**
- Use Cloudflare's free certificates
- No Let's Encrypt needed
- Cloudflare handles SSL

**Configuration:**

1. **Enable Proxy in Cloudflare:**
   - DNS Records → Enable proxy (orange cloud)
   
2. **Set SSL Mode to Full:**
   - Cloudflare Dashboard → SSL/TLS
   - Select "Full" mode (not "Full (Strict)")
   
3. **No Certificate Needed:**
   - Keep self-signed certificate
   - Cloudflare handles browser-facing SSL
   - Server-to-Cloudflare uses self-signed

**Security Recommendation:**
- Enable "Allow traffic from Cloudflare only"
- Prevents direct server access
- Forces traffic through Cloudflare
- Protects against DDoS

**Full (Strict) Mode:**
- Requires valid certificate on server
- Use Let's Encrypt for this
- More secure
- Recommended for production

### 9.4 Import Custom Certificate

**When to Use:**
- Extended Validation (EV) certificates
- Organization Validated (OV) certificates
- Specific CA requirements
- Pre-purchased certificates
- Wildcard certificates

**Import Process:**
1. Site → SSL/TLS tab
2. Actions → Import Certificate
3. Provide:
   - Private Key (RSA/ECDSA)
   - Certificate (PEM format)
   - Certificate Chain (Intermediate + Root)
4. Click "Import and Install"

**Certificate Requirements:**
- PEM format
- Private key (unencrypted)
- Full certificate chain
- Must match domain name

**Example Private Key:**
```
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----
```

**Example Certificate:**
```
-----BEGIN CERTIFICATE-----
MIIFXzCCBEegAwIBAgIQBHmxLqZvKNhWAo+4bQb5OjANBgkqh...
-----END CERTIFICATE-----
```

**Example Certificate Chain:**
```
-----BEGIN CERTIFICATE-----
[Intermediate Certificate]
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
[Root Certificate]
-----END CERTIFICATE-----
```

### 9.5 HTTP to HTTPS Redirection

**Automatic Redirection:**
- Enabled by default for all sites
- 301 Permanent Redirect
- Configured in NGINX vhost
- SEO-friendly
- Cannot be disabled (security best practice)

**Redirect Logic:**
```nginx
if ($scheme != "https") {
    return 301 https://$host$request_uri;
}
```

### 9.6 HSTS (HTTP Strict Transport Security)

**Configuration:**
Added in vhost:
```nginx
add_header Strict-Transport-Security "max-age=31536000" always;
```

**Benefits:**
- Forces HTTPS for all future visits
- Prevents SSL stripping attacks
- Improves security
- Better SEO

### 9.7 Certificate Management

**View Certificate:**
- Details visible in SSL/TLS tab
- Expiration date
- Issuer information
- Domain names covered
- Algorithm type

**Replace Certificate:**
- Issue new Let's Encrypt
- Import different certificate
- No downtime
- Old certificate immediately replaced

**Certificate Files:**
```
/etc/ssl/certs/clp/www.domain.com.crt
/etc/ssl/private/clp/www.domain.com.key
```

---

## 10. File & User Management

### 10.1 Site User Concept

**Site User = Unix System User:**
- Each site has dedicated user
- Home directory: `/home/$siteUser/`
- SSH/SFTP access
- Isolated from other sites
- Individual PHP-FPM pool
- Separate logs

**Directory Structure:**
```
/home/$siteUser/
├── htdocs/
│   └── www.domain.com/    # Website files
├── logs/
│   ├── access.log         # NGINX access log
│   ├── error.log          # NGINX error log
│   └── php-fpm.log        # PHP-FPM log
├── backups/               # Database backups
├── tmp/                   # Temporary files
└── .varnish-cache/        # Varnish controller (if enabled)
    └── controller.php
```

### 10.2 SSH Access

**Login:**
```bash
ssh site-user@server-ip-address
```

**Available Tools:**
- Composer
- WP-CLI (for WordPress sites)
- clpctl (CloudPanel CLI)
- Git
- npm/Yarn
- PHP CLI
- MySQL client

**Common Tasks via SSH:**
```bash
# Install Composer dependencies
cd ~/htdocs/www.domain.com
composer install

# Laravel artisan commands
php artisan migrate

# WordPress CLI
wp plugin list
wp theme activate twentytwentyfour

# Git operations
git clone https://github.com/user/repo.git
git pull origin main
```

**Security:**
- Password authentication
- SSH key authentication supported
- Per-site isolation
- No root access
- Limited to site directory

### 10.3 SFTP Access

**Connection Details:**
- Host: server-ip-address
- Port: 22
- Protocol: SFTP
- Username: site-user
- Password: site-user-password

**SFTP Clients:**
- FileZilla
- Cyberduck
- WinSCP (Windows)
- Transmit (Mac)
- CLI: sftp command

**Operations:**
- Upload files
- Download files
- Edit files remotely
- Change permissions
- Create directories

### 10.4 File Manager (via CloudPanel)

**Not Available:**
CloudPanel does not include a web-based file manager.

**File Management Options:**
1. **SFTP** - Recommended for file uploads
2. **SSH** - Command-line access
3. **Git** - Version control deployment

**Rationale:**
- Security (web file managers are attack vectors)
- Performance (CLI is faster)
- Best practices (version control)

### 10.5 File Permissions

**Default Permissions:**
- Directories: 755 (rwxr-xr-x)
- Files: 644 (rw-r--r--)
- Owned by: site-user:site-user

**Secure Permissions:**
```bash
# Set directory permissions
find ~/htdocs/www.domain.com -type d -exec chmod 755 {} \;

# Set file permissions
find ~/htdocs/www.domain.com -type f -exec chmod 644 {} \;

# Writable directories (uploads, cache)
chmod 775 ~/htdocs/www.domain.com/wp-content/uploads
```

**Laravel/Symfony Permissions:**
```bash
# Storage and cache directories
chmod -R 775 storage bootstrap/cache
```

### 10.6 Disk Usage

**Monitoring:**
- View in CloudPanel dashboard
- Per-site disk usage
- Alert on quota limits

**Check via SSH:**
```bash
# Disk usage summary
du -sh ~/htdocs/www.domain.com

# Detailed breakdown
du -h --max-depth=1 ~/htdocs/www.domain.com

# Find large files
find ~/htdocs -type f -size +10M -exec ls -lh {} \;
```

---

## 11. Security Features

### 11.1 Site Isolation

**System-Level Isolation:**
- Each site = separate Unix user
- File system permissions
- Process isolation (PHP-FPM pools)
- Memory limits per site
- CPU limits per site

**Security Benefits:**
- Compromised site cannot access other sites
- File permissions prevent cross-site access
- Resource limits prevent resource exhaustion attacks
- Isolated log files

**Technical Implementation:**
```
Site 1: User john-doe → /home/john-doe/ → PHP-FPM pool john-doe
Site 2: User jane-smith → /home/jane-smith/ → PHP-FPM pool jane-smith
```

### 11.2 Firewall Management

**UFW (Uncomplicated Firewall):**
- Built-in firewall
- User-friendly interface
- Rule management
- Port control

**Default Rules:**
- Port 22 (SSH): Open
- Port 80 (HTTP): Open
- Port 443 (HTTPS): Open
- Port 8443 (CloudPanel): Open
- Port 3306 (MySQL): Closed
- All other ports: Closed

**Managing Rules:**

1. Admin Area → Security → Firewall
2. Add Rule:
   - Name: Description
   - Port: Port number or range
   - Protocol: TCP/UDP
   - Source: IP address, CIDR, or anywhere
   - Action: Allow/Deny
3. Save Rule

**Best Practices:**
- **Restrict CloudPanel Access (Port 8443):**
  - Whitelist office/home IP
  - VPN IP addresses only
  - Block public access
  
- **Restrict SSH Access (Port 22):**
  - Whitelist trusted IPs
  - Consider VPN access
  - Disable password authentication (use SSH keys)

**Cloud Provider Firewalls:**
- **Recommended:** Use provider's security groups
- AWS Security Groups
- GCE Firewall Rules
- Azure Network Security Groups
- Benefits:
  - Traffic blocked before reaching server
  - Better performance
  - DDoS protection
  - Centralized management

### 11.3 Basic Authentication

**CloudPanel Protection:**
- Extra security layer for CloudPanel interface
- HTTP Basic Auth
- Username + Password prompt
- Protects port 8443

**Enable via UI:**
1. Admin Area → Security → Basic Auth
2. Enter:
   - User Name
   - Password
3. Click "Save"

**Enable via CLI:**
```bash
sudo su root
clpctl cloudpanel:enable:basic-auth --userName='john.doe' --password='SecurePass123!'
```

**Disable via UI:**
1. Admin Area → Security → Basic Auth
2. Click "Disable Basic Auth"

**Disable via CLI:**
```bash
clpctl cloudpanel:disable:basic-auth
```

**Use Cases:**
- Additional security when IP whitelisting not possible
- Temporary access for contractors
- Defense in depth strategy

### 11.4 Two-Factor Authentication

**2FA for Users:**
- TOTP (Time-based One-Time Password)
- Authenticator app required (Google Authenticator, Authy, etc.)
- Per-user configuration
- Enhances login security

**Enable 2FA:**
1. User account → Two-Factor Authentication
2. Scan QR code with authenticator app
3. Enter verification code
4. Save backup codes
5. 2FA enabled

**Login Process:**
1. Enter username + password
2. Enter 6-digit code from authenticator app
3. Access granted

**Backup Codes:**
- Generated during setup
- Save securely
- Use if phone lost/unavailable
- One-time use

### 11.5 Cloudflare Integration

**Traffic Routing:**
- Route all traffic through Cloudflare
- DDoS protection
- Web Application Firewall (WAF)
- Rate limiting
- Bot management

**Configuration:**
1. Add domain to Cloudflare
2. Update nameservers
3. Enable proxy (orange cloud)
4. Configure firewall rules

**Allow Traffic from Cloudflare Only:**

**Benefits:**
- Attackers cannot bypass Cloudflare
- Direct IP access blocked
- Enhanced DDoS protection
- True origin IP hidden

**Implementation:**
1. Site → Security
2. Enable "Allow traffic from Cloudflare only"
3. NGINX restricts connections to Cloudflare IP ranges

**Cloudflare IP Ranges:**
Automatically updated from:
- https://www.cloudflare.com/ips-v4
- https://www.cloudflare.com/ips-v6

### 11.6 IP & Bot Blocker

**IP Blocker:**
- Block traffic from specific IPs
- Block entire IP ranges (CIDR)
- Block countries (via Cloudflare)

**Bot Blocker:**
- Block known bad bots
- Allow good bots (Google, Bing)
- User-agent filtering
- Rate limiting

**Configuration:**
- Site level blocking
- Server level blocking
- Regex pattern matching

### 11.7 Security Hardening Best Practices

**Operating System:**
- Keep Ubuntu/Debian updated
- Apply security patches promptly
- Remove unnecessary software
- Disable unused services

**Update Commands:**
```bash
sudo apt update
sudo apt upgrade
sudo apt dist-upgrade
```

**CloudPanel Updates:**
- Update promptly when available
- Check release notes
- Test in staging first (if possible)

**Password Policy:**
- Strong passwords (16+ characters)
- Unique passwords per site
- Password manager recommended
- Rotate passwords periodically

**SSH Hardening:**
```bash
# Disable root login
sudo nano /etc/ssh/sshd_config
PermitRootLogin no

# Disable password authentication (use keys)
PasswordAuthentication no

# Restart SSH
sudo systemctl restart sshd
```

**Fail2ban (Optional):**
- Install separately
- Ban IPs after failed login attempts
- Protect SSH, CloudPanel, FTP

**Regular Backups:**
- Automated backups essential
- Test restore procedures
- Off-server storage
- Encryption at rest

### 11.8 Application Security

**WordPress Security:**
- Keep WordPress core updated
- Update plugins/themes regularly
- Use security plugins (Wordfence, Sucuri)
- Strong admin passwords
- Limit login attempts
- Disable file editing in dashboard

**PHP Security:**
- Disable dangerous functions
- Enable open_basedir restrictions
- Disable allow_url_fopen/include
- Configure error reporting (production)

**Database Security:**
- Strong database passwords
- Remove test databases
- Disable remote access (if not needed)
- Regular backups

---

## 12. Vhost Configuration

### 12.1 Vhost Editor

**Purpose:**
- Customize NGINX configuration
- Add rewrites/redirects
- Configure reverse proxies
- Advanced customization
- Security directives

**Access:**
1. Site → Vhost tab
2. Edit vhost configuration
3. Save changes

**Safety Features:**
- Syntax validation before saving
- Automatic rollback on error
- Configuration backup
- No downtime on syntax errors

**Vhost Templates:**
- Available on GitHub: https://github.com/cloudpanel-io/vhost-templates/tree/master/v2
- Automatically updated nightly
- Application-specific configurations
- Reference implementations

### 12.2 HTTP/3 Support

**Availability:**
- Ubuntu 24.04 (NGINX 1.26+)
- Debian 12 (NGINX 1.26+)
- Experimental support
- Must have valid SSL certificate (not self-signed)

**Requirements:**
- Valid SSL/TLS certificate (Let's Encrypt or imported)
- UDP Port 443 open in firewall
- Modern browser support

**Enable HTTP/3:**
1. Site → Vhost tab
2. Find line: `http3 off;`
3. Change to: `http3 on;`
4. Click "Save"

**Verification:**
- Use HTTP/3 Check: https://http3check.net/
- Check browser developer tools
- Look for "h3" protocol

**Performance Benefits:**
- Faster connection establishment
- Better performance on lossy networks
- Reduced latency
- Improved mobile experience

### 12.3 Common Vhost Customizations

**Custom Rewrites:**
```nginx
# Example: Redirect old URLs
location /old-path {
    return 301 /new-path;
}

# Rewrite with regex
rewrite ^/blog/(.*)$ /news/$1 permanent;
```

**IP Whitelisting:**
```nginx
# Restrict admin area by IP
location /admin {
    allow 203.0.113.0/24;
    deny all;
}
```

**Custom Headers:**
```nginx
# Security headers
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;
```

**Rate Limiting:**
```nginx
# Limit requests per IP
limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;

location /wp-login.php {
    limit_req zone=login burst=2 nodelay;
}
```

**Reverse Proxy:**
```nginx
location /api {
    proxy_pass http://backend-server:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

**Client Max Body Size:**
```nginx
# Increase upload size limit
client_max_body_size 100M;
```

### 12.4 Vhost Structure

**Main Sections:**
1. **HTTP Block** - Redirect to HTTPS
2. **HTTPS Block** - Main site configuration
3. **Location Blocks** - Specific path handling
4. **PHP Processing** - FastCGI configuration
5. **Static Files** - Cache headers, Gzip

**Example Structure:**
```nginx
# HTTP redirect to HTTPS
server {
    listen 80;
    server_name www.domain.com domain.com;
    return 301 https://www.domain.com$request_uri;
}

# HTTPS main configuration
server {
    listen 443 ssl http2;
    http3 off;
    server_name www.domain.com;
    root /home/siteuser/htdocs/www.domain.com;
    
    # SSL configuration
    ssl_certificate /path/to/cert;
    ssl_certificate_key /path/to/key;
    
    # Varnish proxy (if enabled)
    location / {
        proxy_pass http://127.0.0.1:6081;
        # proxy headers...
    }
    
    # Static files
    location ~* \.(jpg|jpeg|png|gif|css|js|ico)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
    
    # PHP processing
    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php8.3-fpm-siteuser.sock;
        # FastCGI parameters...
    }
}
```

---

## 13. Cron Jobs

### 13.1 Overview

**Cron:**
- Unix task scheduler
- Time-based job execution
- Background task automation
- Per-site user cron jobs

**Use Cases:**
- WordPress scheduled tasks (wp-cron)
- Laravel task scheduler
- Database cleanup
- Log rotation
- Backup scripts
- Email queue processing
- Cache clearing

### 13.2 Adding Cron Jobs

**Via CloudPanel UI:**
1. Site → Cron Jobs tab
2. Click "Add Cron Job"
3. Select template or custom schedule
4. Enter command to execute
5. Click "Add"

**Schedule Templates:**
- Every Minute
- Every 5 Minutes
- Every 15 Minutes
- Every 30 Minutes
- Hourly
- Daily
- Weekly
- Monthly
- Custom

**Example Commands:**
```bash
# Laravel scheduler
cd /home/user/htdocs/www.domain.com && php artisan schedule:run >> /dev/null 2>&1

# WordPress cron
cd /home/user/htdocs/www.domain.com && php wp-cron.php > /dev/null 2>&1

# Custom script
/usr/bin/php8.3 /home/user/scripts/cleanup.php

# Database backup
clpctl db:export --databaseName=mydb --file=~/backups/manual-backup.sql.gz
```

**Via Command Line:**
1. SSH as site user
2. Edit crontab:
   ```bash
   crontab -e
   ```
3. Add cron entry:
   ```cron
   # Laravel scheduler (every minute)
   * * * * * cd /home/user/htdocs/www.domain.com && php artisan schedule:run >> /dev/null 2>&1
   
   # Daily backup at 2 AM
   0 2 * * * clpctl db:export --databaseName=mydb --file=~/backups/daily.sql.gz
   
   # WordPress cron every 15 minutes
   */15 * * * * cd /home/user/htdocs/www.domain.com && php wp-cron.php
   ```

### 13.3 Cron Syntax

**Format:**
```
* * * * * command
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, 0=Sunday)
│ │ │ └────── Month (1-12)
│ │ └───────── Day of month (1-31)
│ └──────────── Hour (0-23)
└───────────────── Minute (0-59)
```

**Examples:**
```cron
# Every minute
* * * * * command

# Every 5 minutes
*/5 * * * * command

# Every hour
0 * * * * command

# Daily at 3:30 AM
30 3 * * * command

# Every Monday at 9 AM
0 9 * * 1 command

# First day of month at midnight
0 0 1 * * command

# Every weekday at 6 PM
0 18 * * 1-5 command
```

### 13.4 Deleting Cron Jobs

**Via CloudPanel:**
1. Cron Jobs tab
2. Click "Delete" next to cron job
3. Confirm deletion

**Via Command Line:**
```bash
# Edit crontab
crontab -e

# Remove the line
# Save and exit
```

**View Current Cron Jobs:**
```bash
crontab -l
```

### 13.5 Cron Job Logging

**Output Redirection:**
```bash
# Discard all output
command > /dev/null 2>&1

# Log to file
command >> /home/user/logs/cron.log 2>&1

# Log errors only
command 2>> /home/user/logs/cron-errors.log

# Both stdout and stderr to file
command >> /home/user/logs/cron.log 2>&1
```

**System Cron Logs:**
```bash
# View cron execution log
sudo tail -f /var/log/syslog | grep CRON
```

### 13.6 Common Cron Issues

**Command Not Found:**
- Use absolute paths
- Set PATH in crontab
- Source environment variables

**Permission Denied:**
- Check file permissions
- Ensure script is executable
- Run as correct user

**Mail Notifications:**
- Cron sends output via email
- Configure mail server or redirect output
- Use `> /dev/null 2>&1` to suppress

---

## 14. Logs & Monitoring

### 14.1 Log Viewer

**Access:**
- Site → Logs tab
- View recent log entries
- Real-time log viewing
- Filter by log type

**Available Logs:**
- NGINX Access Log
- NGINX Error Log
- PHP-FPM Error Log

**Log Locations:**
```
/home/$siteUser/logs/
├── access.log           # NGINX access log
├── error.log            # NGINX error log
└── php-fpm.log          # PHP-FPM error log
```

### 14.2 Access Logs

**Format:**
```
IP_ADDRESS - - [DATE] "REQUEST" STATUS SIZE "REFERRER" "USER_AGENT"
```

**Example Entry:**
```
203.0.113.45 - - [02/Nov/2025:10:15:30 +0000] "GET /index.php HTTP/1.1" 200 4523 "https://google.com" "Mozilla/5.0..."
```

**Information:**
- Client IP address
- Request timestamp
- HTTP method and path
- Response status code
- Response size
- Referrer URL
- User agent string

**View via SSH:**
```bash
# Latest entries
tail -f ~/logs/access.log

# Last 100 lines
tail -n 100 ~/logs/access.log

# Search for specific IP
grep "203.0.113.45" ~/logs/access.log

# Count requests per IP
awk '{print $1}' ~/logs/access.log | sort | uniq -c | sort -rn

# Find 404 errors
grep " 404 " ~/logs/access.log
```

### 14.3 Error Logs

**NGINX Error Log:**
- Server-level errors
- Configuration issues
- Upstream errors (PHP-FPM)
- SSL/TLS errors

**PHP-FPM Error Log:**
- PHP runtime errors
- Warnings and notices
- Fatal errors
- Application exceptions

**View via SSH:**
```bash
# Latest errors
tail -f ~/logs/error.log

# PHP errors
tail -f ~/logs/php-fpm.log

# Search for specific error
grep "Fatal error" ~/logs/php-fpm.log

# Last 50 errors
tail -n 50 ~/logs/error.log
```

### 14.4 Log Rotation

**Automatic Rotation:**
- Logs rotated automatically
- Prevents disk space issues
- Compressed old logs
- Retention period: 7-14 days

**Configuration:**
```
/etc/logrotate.d/cloudpanel
```

**Manual Rotation:**
```bash
sudo logrotate -f /etc/logrotate.d/cloudpanel
```

### 14.5 Monitoring Dashboard

**CloudPanel Dashboard:**
- CPU usage
- Memory usage
- Disk usage
- Network traffic
- Load average

**Per-Site Metrics:**
- Disk usage
- Database size
- Traffic statistics
- Request counts

**Real-Time Monitoring:**
- Live CPU/Memory graphs
- Historical data (24h, 7d, 30d)
- Alert thresholds

### 14.6 System Logs

**Important System Logs:**
```bash
# System log
sudo tail -f /var/log/syslog

# Authentication log
sudo tail -f /var/log/auth.log

# CloudPanel log
sudo tail -f /var/log/cloudpanel.log

# NGINX access/error
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# MySQL error log
sudo tail -f /var/log/mysql/error.log
```

---

## 15. Admin Area

### 15.1 Users Management

**User Roles:**

**1. Admin:**
- Full permissions
- Access to all areas
- Can manage all sites
- Can create/delete users
- Access to admin area
- Billing access

**2. Site Manager:**
- Manage all sites
- Create new sites
- Delete sites
- No access to admin area
- No billing access
- No user management

**3. User (Restricted):**
- Access to assigned sites only
- Cannot see other sites
- Cannot create sites
- No admin area access
- Limited operations

**Adding Users:**
1. Admin Area → Users
2. Click "Add User"
3. Configure:
   - User Name
   - Email
   - Password
   - Role
   - Timezone
   - (If User role) Assign sites
4. Click "Add User"

**User Properties:**
- Timezone: Important for monitoring graphs
- Email: Notifications and password resets
- Role: Determines permissions
- Site Access: Specific to User role

**Deleting Users:**
1. Admin Area → Users
2. Click username
3. Click "Delete" (bottom left)
4. Confirm action

**Important:**
- Cannot delete currently logged-in user
- Admin role: At least one admin must exist

### 15.2 Settings

**General Settings:**

**Custom CloudPanel Domain:**
1. Settings → General
2. Enter domain name (e.g., cp.domain.com)
3. Create DNS A record pointing to server
4. Click "Save"
5. Wait for DNS propagation
6. Certificate issued automatically

**Alternative: Reverse Proxy:**
- Create reverse proxy site
- Domain: cp.domain.com
- Proxy URL: https://127.0.0.1:8443
- Import custom SSL certificate
- Benefits: Use own SSL certificate

**Database Settings:**

**External Database Server:**
- Use managed MySQL service (AWS RDS, DO Managed MySQL)
- Benefits:
  - Better performance
  - Point-in-time recovery
  - Automated backups
  - High availability

**Supported Versions:**
- MySQL 5.7, 8.0
- MariaDB 10.6+

**Configuration:**
1. Settings → Database
2. Enter:
   - Host
   - Port (default: 3306)
   - Root username
   - Root password
3. Test connection
4. Save

### 15.3 Instance Management

**Reboot Instance:**
1. Admin Area → Instance
2. Click "Reboot" (top right)
3. Confirm reboot
4. Server restarts (2-5 minutes)

**Service Management:**
Restart individual services:
- NGINX
- MySQL/MariaDB
- PHP-FPM (all versions)
- Redis
- ProFTPD
- Varnish Cache

**Restart Process:**
1. Admin Area → Instance
2. Select service
3. Click "Restart"
4. Brief service interruption (~5 seconds)

**Use Cases:**
- Apply configuration changes
- Clear service cache
- Resolve service issues
- After software updates

### 15.4 Security Settings

**Firewall Management:**
- Add/Remove firewall rules
- Port management
- IP whitelisting
- Service access control

**Basic Authentication:**
- Enable/Disable for CloudPanel
- Username/password configuration
- Extra security layer

**IP Blocker:**
- Block specific IPs
- Block IP ranges
- Country blocking (via Cloudflare)

**Bot Blocker:**
- Block known bad bots
- Allow good bots
- User-agent filtering

### 15.5 Cloud Features

**Available by Provider:**
- AWS: Automated AMIs
- Digital Ocean: Droplet Snapshots
- Google Cloud: Automated Snapshots
- Hetzner: Server Snapshots
- Azure: VM Snapshots

**Configuration:**
- Provider-specific settings
- Frequency configuration
- Retention period
- Automatic vs. manual

---

## 16. Cloud Features

### 16.1 Amazon Web Services (AWS)

**Automatic AMIs (Amazon Machine Images):**

**Setup:**
1. Create IAM user for CloudPanel
2. Attach policy with EC2 permissions
3. Generate access keys
4. In CloudPanel: Admin Area → Amazon Web Services
5. Enter:
   - Access Key ID
   - Secret Access Key
   - Region
6. Click "Save"

**Enable Automatic Images:**
1. Settings tab
2. Select:
   - Frequency: Hourly, Every 3h, 6h, 12h, or Daily
   - Retention Period: Days to keep (7-30)
3. Click "Save"

**Example Configuration:**
- Frequency: Twice per day
- Retention: 7 days
- Result: 14 AMIs maintained

**Benefits:**
- Full instance backup
- Incremental backups
- Point-in-time recovery
- Disaster recovery
- Zero downtime backups

**Manual AMI Creation:**
1. Images tab
2. Click "Create Image"
3. Enter image name
4. Click "Create"
5. Image created in 5-15 minutes

**AMI Management:**
- View all AMIs
- Delete old AMIs
- Restore from AMI
- Cross-region copying

### 16.2 Digital Ocean

**Droplet Snapshots:**

**Automated Backups:**
- Enable on droplet creation
- Weekly automated backups
- Managed by Digital Ocean
- Additional cost

**Spaces Integration:**
- Object storage for backups
- Integrated with Remote Backups
- S3-compatible API

### 16.3 Google Compute Engine (GCE)

**Automated Snapshots:**

**Setup:**
1. Create service account
2. Generate JSON key file
3. In CloudPanel: Admin Area → Google Compute Engine
4. Upload key file
5. Select project

**Enable Automatic Snapshots:**
1. Settings tab
2. Configure:
   - Frequency: Daily, twice daily, etc.
   - Retention Period: Days
3. Save

**Example:**
- 4 snapshots per day
- 7-day retention
- 28 snapshots maintained

**Manual Snapshots:**
1. Snapshots tab
2. Click "Create Snapshot"
3. Enter name
4. Click "Create"

**Benefits:**
- Incremental backups
- Fast recovery
- Cost-effective
- Automatic retention management

### 16.4 Hetzner Cloud

**Server Snapshots:**

**Configuration:**
- Manual snapshot creation
- Snapshot-based backups
- Quick instance recovery

**Backup Strategy:**
- Regular snapshot schedule
- Before major updates
- Disaster recovery

### 16.5 Microsoft Azure

**VM Snapshots:**

**Azure Backup:**
- Native Azure Backup integration
- Automated backup policies
- Point-in-time recovery

### 16.6 Vultr

**Snapshot Support:**
- Manual snapshots
- Automated snapshot scheduling
- Quick recovery
- Cross-region snapshots

---

## 17. Remote Backups

### 17.1 Overview

**Supported Storage Providers:**
- Amazon S3
- Wasabi
- Digital Ocean Spaces
- Backblaze B2
- Google Drive
- Dropbox
- SFTP/SSH
- Custom Rclone configurations (40+ providers)

**Backup Contents:**
- All website files
- Database dumps
- Configuration files
- Logs (optional)

**Backup Frequency:**
- Manual backups
- Scheduled backups (via cron)
- On-demand backups

### 17.2 Amazon S3

**Setup:**
1. Create S3 bucket
2. Create IAM user with S3 permissions
3. Generate access keys
4. CloudPanel: Admin Area → Backups → Add Backup
5. Select "Amazon S3"
6. Configure:
   - Access Key ID
   - Secret Access Key
   - Bucket Name
   - Region
   - Path (optional)
   - Excludes (optional)
7. Save
8. Click "Create Backup" to test

**Exclusions:**
```
/logs/*
/cache/*
/tmp/*
node_modules/*
```

### 17.3 Backblaze B2

**Setup:**
1. Create B2 bucket
2. Generate application key
3. CloudPanel: Backups → Add Backup
4. Select "Custom Rclone Config"
5. SSH as root:
   ```bash
   sudo su root
   ```
6. Create rclone config:
   ```bash
   rclone config
   ```
7. Select "New remote"
8. Name: `remote`
9. Type: `b2`
10. Enter:
    - Account ID
    - Application Key
    - Endpoint (optional)
11. Exit config
12. Test upload:
    ```bash
    echo "test" > /tmp/test-file
    rclone copy /tmp/test-file remote:my-bucket/backups/
    ```
13. CloudPanel: Enter bucket details
14. Click "Create Backup" to test

**Example Config:**
```ini
[remote]
type = b2
account = 004b6959c792d2a0000000008
key = K004QyHV7vAe5UaRbaXsZKHxE9Se87
endpoint = 
```

### 17.4 Wasabi

**Setup:**
1. Create Wasabi bucket
2. Generate access keys
3. CloudPanel: Backups → Add Backup
4. Select "Wasabi"
5. Configure:
   - Access Key ID
   - Secret Access Key
   - Bucket Name
   - Endpoint
6. Save and test

**Endpoints:**
- us-east-1: s3.wasabisys.com
- us-west-1: s3.us-west-1.wasabisys.com
- eu-central-1: s3.eu-central-1.wasabisys.com

### 17.5 Digital Ocean Spaces

**Setup:**
1. Create Space
2. Generate Spaces access keys
3. CloudPanel: Backups → Add Backup
4. Select "Digital Ocean Spaces"
5. Configure:
   - Access Key
   - Secret Key
   - Space Name
   - Endpoint
6. Save and test

**Endpoint Format:**
```
nyc3.digitaloceanspaces.com
fra1.digitaloceanspaces.com
sfo3.digitaloceanspaces.com
```

### 17.6 Google Drive

**Setup:**
1. CloudPanel: Backups → Add Backup
2. Select "Google Drive"
3. Click "Request Access Code"
4. Authorize CloudPanel access
5. Enter access code
6. Configure:
   - Folder name (optional)
   - Excludes (optional)
7. Save and test

**Important:**
- CloudPanel only accesses Apps/CloudPanel/ folder
- No access to other Google Drive files
- Secure OAuth2 authentication

### 17.7 Dropbox

**Setup:**
1. CloudPanel: Backups → Add Backup
2. Select "Dropbox"
3. Click "Request Access Code"
4. Authorize CloudPanel
5. Enter access code
6. Configure excludes (optional)
7. Save and test

**Storage Location:**
- Apps/CloudPanel/ in Dropbox
- Isolated from other files
- Secure OAuth2

### 17.8 SFTP/SSH

**Setup:**
1. Remote server with SSH/SFTP access
2. CloudPanel: Backups → Add Backup
3. Select "SFTP"
4. Configure:
   - Host
   - Port (22)
   - Username
   - Authentication:
     - Password, or
     - SSH Key
   - Remote path
   - Excludes
5. Test connection
6. Save

**SSH Key Authentication (Recommended):**
1. Generate key on CloudPanel server:
   ```bash
   ssh-keygen -t ed25519 -f my-private-key
   ```
2. Copy public key to remote server:
   ```bash
   ssh-copy-id -i my-private-key.pub user@remote-server
   ```
3. Test connection:
   ```bash
   ssh -i my-private-key user@remote-server
   ```
4. Enter private key path in CloudPanel

### 17.9 Custom Rclone Configurations

**Supported Providers (40+):**
- Alibaba Cloud OSS
- Azure Blob Storage
- Box
- Dropbox Business
- Google Cloud Storage
- IBM COS S3
- Mega
- Microsoft OneDrive
- Oracle Cloud Storage
- Scaleway
- UpCloud
- And many more...

**Setup:**
1. SSH as root
2. Run rclone config:
   ```bash
   rclone config
   ```
3. Create new remote named "remote"
4. Follow provider-specific wizard
5. Test configuration:
   ```bash
   rclone copy /tmp/test-file remote:bucket/path/
   ```
6. CloudPanel: Select "Custom Rclone Config"
7. Enter bucket/container details
8. Save and test

**Configuration File:**
```
/root/.config/rclone/rclone.conf
```

### 17.10 Backup Restoration

**Restoring Files:**

**Via File Manager (if available):**
1. Download backup from storage provider
2. Extract archive locally
3. Upload files via SFTP

**Via SSH:**
1. Download backup:
   ```bash
   cd ~/tmp/
   # Example for SFTP
   sftp user@backup-server
   get /backups/domain.com_2025-11-02.tar
   exit
   ```
2. Extract backup:
   ```bash
   tar xf domain.com_2025-11-02.tar
   ```
3. Copy files:
   ```bash
   cp -R extracted-files/* ~/htdocs/www.domain.com/
   ```

**Database Restore:**
1. Extract database dump
2. Import using clpctl:
   ```bash
   clpctl db:import --databaseName=mydb --file=backup.sql.gz
   ```

---

## 18. Dploy Deployment

### 18.1 Overview

**dploy** - Fast, Simple Code Deployment

**Features:**
- Open source (MIT License)
- Setup in < 60 seconds
- Zero downtime deployments
- Continuous deployment support
- Git-based deployment
- Multiple application support

**Supported Applications:**
- PHP applications
- Node.js applications
- Python applications
- Static HTML sites

**GitHub Repository:**
https://github.com/cloudpanel-io/dploy

### 18.2 How It Works

**Deployment Process:**
1. Code push to Git repository
2. Webhook triggers dploy
3. dploy pulls latest code
4. Runs build/deployment scripts
5. Switches to new version (symlink)
6. Zero downtime achieved

**Directory Structure:**
```
/home/siteuser/deployments/
├── current -> releases/20251102101530/
├── releases/
│   ├── 20251102101530/
│   ├── 20251102093020/
│   └── 20251101161545/
├── shared/
│   ├── .env
│   ├── storage/
│   └── uploads/
└── .dploy/
    └── config.yml
```

### 18.3 Zero Downtime Deployment

**Symlink Strategy:**
- `current` symlink points to active release
- New release deployed to separate directory
- After successful deployment, symlink updated
- Instant switch with no downtime
- Easy rollback (just update symlink)

**Example:**
```bash
# Before deployment
current -> releases/20251102093020/

# After deployment
current -> releases/20251102101530/

# Rollback
current -> releases/20251102093020/
```

### 18.4 Configuration

**dploy Config File:**
```yaml
# .dploy/config.yml
repository: https://github.com/user/repo.git
branch: main
shared_directories:
  - storage
  - uploads
shared_files:
  - .env
build_commands:
  - composer install --no-dev
  - npm install
  - npm run build
deployment_commands:
  - php artisan migrate --force
  - php artisan cache:clear
keep_releases: 5
```

**Configuration Options:**
- Repository URL
- Branch to deploy
- Shared directories (persist across deployments)
- Shared files (config, .env)
- Build commands
- Deployment commands
- Number of releases to keep

### 18.5 Continuous Deployment

**Webhook Setup:**
1. Configure dploy in CloudPanel
2. Generate webhook URL
3. Add webhook to Git provider (GitHub, GitLab, Bitbucket)
4. Push to branch triggers automatic deployment

**Workflow:**
```
Developer commits → Git push → Webhook → dploy → Deploy → Live
```

**Benefits:**
- Automated deployments
- Faster release cycles
- Reduced human error
- Consistent deployments

### 18.6 Supported Git Providers

- GitHub
- GitLab
- Bitbucket
- Self-hosted Git
- Any Git repository with HTTP/HTTPS access

### 18.7 Rollback

**Quick Rollback:**
1. Access dploy interface
2. Select previous release
3. Click "Rollback"
4. Symlink updated
5. Previous version live

**Manual Rollback:**
```bash
# List releases
ls -la ~/deployments/releases/

# Update symlink
cd ~/deployments/
ln -sfn releases/20251102093020 current

# Restart services if needed
```

---

## 19. CLI Tools

### 19.1 clpctl (CloudPanel CLI)

**Overview:**
- Command-line interface for CloudPanel
- Root-level commands
- User-level commands
- Automation support

**Root User Commands:**

**List Commands:**
```bash
clpctl
```

**Site Management:**
```bash
# Add WordPress site
clpctl site:add:wordpress --domainName=www.domain.com \
  --siteUser='john-doe' --siteUserPassword='pass123!' \
  --phpVersion=8.3

# Add PHP site
clpctl site:add:php --domainName=www.domain.com \
  --phpVersion=8.4 --vhostTemplate='Laravel 12' \
  --siteUser='john-doe' --siteUserPassword='pass123!'

# Add Node.js site
clpctl site:add:nodejs --domainName=www.domain.com \
  --nodejsVersion=20 --appPort=3000 \
  --siteUser='john-doe' --siteUserPassword='pass123!'

# Add Python site
clpctl site:add:python --domainName=www.domain.com \
  --pythonVersion=3.11 --appPort=8000 \
  --siteUser='john-doe' --siteUserPassword='pass123!'

# Add Static HTML site
clpctl site:add:static --domainName=www.domain.com \
  --siteUser='john-doe' --siteUserPassword='pass123!'

# Add Reverse Proxy
clpctl site:add:reverse-proxy --domainName=www.domain.com \
  --reverseProxyUrl='http://127.0.0.1:8080' \
  --siteUser='john-doe' --siteUserPassword='pass123!'
```

**Database Management:**
```bash
# Show master credentials
clpctl db:show:master-credentials

# Backup databases
clpctl db:backup --ignoreDatabases='test,old' --retentionPeriod=7

# Import database
clpctl db:import --databaseName=mydb --file=dump.sql.gz

# Export database
clpctl db:export --databaseName=mydb --file=backup.sql.gz
```

**Security:**
```bash
# Enable Basic Auth
clpctl cloudpanel:enable:basic-auth --userName='admin' --password='pass123'

# Disable Basic Auth
clpctl cloudpanel:disable:basic-auth
```

### 19.2 Site User Commands

**Database Operations:**
```bash
# Export database
clpctl db:export --databaseName=mydb --file=backup.sql.gz

# Import database
clpctl db:import --databaseName=mydb --file=backup.sql.gz
```

### 19.3 Other CLI Tools

**Composer:**
```bash
# Install dependencies
composer install

# Update dependencies
composer update

# Create Laravel project
php8.4 /usr/local/bin/composer create-project laravel/laravel myproject
```

**WP-CLI (WordPress):**
```bash
# List plugins
wp plugin list

# Update WordPress
wp core update

# Install plugin
wp plugin install wordfence --activate

# Create admin user
wp user create newadmin admin@domain.com --role=administrator
```

**npm/Yarn:**
```bash
# Install dependencies
npm install
yarn install

# Build assets
npm run build
yarn build

# Development mode
npm run dev
yarn dev
```

**Git:**
```bash
# Clone repository
git clone https://github.com/user/repo.git

# Pull latest
git pull origin main

# Deploy workflow
git pull && composer install && npm run build
```

---

## 20. Performance Optimization

### 20.1 Varnish Cache

**Already Covered in Section 7**

Performance improvement: 100-250x faster

### 20.2 OPcache

**PHP Bytecode Caching:**
- Enabled by default
- Stores compiled PHP scripts
- Reduces CPU usage
- Faster script execution

**Verification:**
```bash
php -i | grep opcache
```

**Configuration:**
```ini
opcache.enable=1
opcache.memory_consumption=128
opcache.interned_strings_buffer=8
opcache.max_accelerated_files=10000
opcache.revalidate_freq=2
```

**Performance Impact:**
- 2-3x faster PHP execution
- Reduced CPU load
- Lower response times

### 20.3 Redis

**Object Caching:**
- In-memory data store
- Cache database queries
- Session storage
- WordPress object cache

**Installation per Site:**
- Redis already installed
- Configure application to use Redis
- Connection: 127.0.0.1:6379

**WordPress Redis:**
```bash
# Install plugin
wp plugin install redis-cache --activate

# Enable Redis
wp redis enable
```

### 20.4 HTTP/2 & HTTP/3

**HTTP/2:**
- Enabled by default
- Multiplexing
- Header compression
- Server push

**HTTP/3:**
- Enable in vhost
- QUIC protocol
- Better mobile performance
- Reduced latency

### 20.5 Gzip Compression

**Automatic Compression:**
- Enabled by default in NGINX
- Compresses text files
- Reduces bandwidth
- Faster page loads

**Compressed Types:**
- HTML
- CSS
- JavaScript
- JSON
- XML
- SVG

### 20.6 Browser Caching

**Cache Headers:**
- Automatic configuration
- Static files cached
- Expires headers
- Cache-Control headers

**Configuration:**
```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff|woff2)$ {
    expires 30d;
    add_header Cache-Control "public, immutable";
}
```

### 20.7 Database Optimization

**MySQL/MariaDB:**
- Optimized default configuration
- InnoDB buffer pool sizing
- Query cache (MariaDB)
- Connection pooling

**Regular Maintenance:**
```bash
# Optimize tables
mysqlcheck -o database_name -u username -p

# Repair tables
mysqlcheck -r database_name -u username -p

# Analyze tables
mysqlcheck -a database_name -u username -p
```

### 20.8 CDN Integration

**Content Delivery Network:**
- Cloudflare (recommended)
- CloudFront
- StackPath
- BunnyCDN

**Benefits:**
- Global content distribution
- Reduced latency
- Bandwidth savings
- DDoS protection

---

## 21. Technical Requirements

### 21.1 System Requirements

**Minimum:**
- **CPU:** 1 Core
- **RAM:** 2 GB
- **Storage:** 10 GB
- **Architecture:** x86_64 or ARM64

**Recommended:**
- **CPU:** 2+ Cores
- **RAM:** 4 GB+
- **Storage:** 20 GB+ SSD
- **Architecture:** x86_64 or ARM64

**Operating System:**
- Ubuntu 24.04 LTS
- Ubuntu 22.04 LTS
- Debian 12 (Bookworm)
- Debian 11 (Bullseye)

**Clean Installation Required:**
- No pre-installed web servers
- No existing NGINX, Apache
- No existing PHP, MySQL
- Fresh OS installation

### 21.2 Network Requirements

**Required Ports (Inbound):**
- 22 (SSH)
- 80 (HTTP)
- 443 (HTTPS)
- 8443 (CloudPanel)

**Optional Ports:**
- 21 (FTP)
- 3306 (MySQL remote access)
- Custom application ports

**Firewall:**
- UFW automatically configured
- Cloud provider security groups recommended

### 21.3 Software Stack Versions

**Web Server:**
- NGINX 1.26+ (Ubuntu 24.04/Debian 12)
- NGINX 1.24+ (Ubuntu 22.04/Debian 11)

**PHP:**
- 7.1, 7.2, 7.3, 7.4
- 8.0, 8.1, 8.2, 8.3, 8.4

**Database:**
- MySQL 8.0
- MariaDB 10.6+

**Caching:**
- Varnish Cache 7.x
- Redis 7.x

**Node.js:**
- Multiple versions via NVM
- 14, 16, 18, 20, 22

**Python:**
- Python 3.9, 3.10, 3.11, 3.12

**FTP:**
- ProFTPD (latest stable)

---

## Appendix A: Quick Reference Commands

### Site Management
```bash
# Create WordPress site
clpctl site:add:wordpress --domainName=www.domain.com --siteUser='user' --siteUserPassword='pass' --phpVersion=8.3

# Create PHP site
clpctl site:add:php --domainName=www.domain.com --phpVersion=8.4 --vhostTemplate='Laravel 12' --siteUser='user' --siteUserPassword='pass'

# Create Node.js site
clpctl site:add:nodejs --domainName=www.domain.com --nodejsVersion=20 --appPort=3000 --siteUser='user' --siteUserPassword='pass'
```

### Database Management
```bash
# Export database
clpctl db:export --databaseName=mydb --file=backup.sql.gz

# Import database
clpctl db:import --databaseName=mydb --file=backup.sql.gz

# Show master credentials
clpctl db:show:master-credentials
```

### Logs
```bash
# View access log
tail -f ~/logs/access.log

# View error log
tail -f ~/logs/error.log

# View PHP errors
tail -f ~/logs/php-fpm.log
```

### Security
```bash
# Enable Basic Auth
clpctl cloudpanel:enable:basic-auth --userName='admin' --password='pass123'

# Disable Basic Auth
clpctl cloudpanel:disable:basic-auth
```

---

## Appendix B: Port Reference

| Port | Service | Access |
|------|---------|--------|
| 22 | SSH/SFTP | Remote access |
| 80 | HTTP | Web traffic (redirects to HTTPS) |
| 443 | HTTPS | Secure web traffic |
| 3306 | MySQL | Database (local/optional remote) |
| 6081 | Varnish | Internal caching (not public) |
| 6379 | Redis | Internal cache (not public) |
| 8080 | NGINX Backend | Internal (Varnish backend) |
| 8443 | CloudPanel | Admin interface |

---

## Appendix C: Supported PHP Versions

| Version | Status | Ubuntu 24.04 | Ubuntu 22.04 | Debian 12 | Debian 11 |
|---------|--------|--------------|--------------|-----------|-----------|
| PHP 7.1 | Legacy | ✓ | ✓ | ✓ | ✓ |
| PHP 7.2 | Legacy | ✓ | ✓ | ✓ | ✓ |
| PHP 7.3 | Legacy | ✓ | ✓ | ✓ | ✓ |
| PHP 7.4 | Security | ✓ | ✓ | ✓ | ✓ |
| PHP 8.0 | Security | ✓ | ✓ | ✓ | ✓ |
| PHP 8.1 | Active | ✓ | ✓ | ✓ | ✓ |
| PHP 8.2 | Active | ✓ | ✓ | ✓ | ✓ |
| PHP 8.3 | Active | ✓ | ✓ | ✓ | ✓ |
| PHP 8.4 | Latest | ✓ | ✓ | ✓ | ✓ |

---

## Document Change History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | November 2, 2025 | Initial comprehensive documentation |

---

## Conclusion

CloudPanel is a free, lightweight, and powerful hosting control panel designed for simplicity and performance. With its modern technology stack, Varnish Cache integration, support for multiple application types, and extensive cloud provider integration, it provides an excellent solution for developers, agencies, and hosting providers.

**Key Strengths:**
1. **Free & Open Source** - Zero licensing costs
2. **Performance** - Varnish Cache, OPcache, Redis, HTTP/3
3. **Flexibility** - PHP, Node.js, Python, Static, Reverse Proxy
4. **Security** - Site isolation, UFW firewall, 2FA, Let's Encrypt
5. **Simplicity** - Clean UI, easy setup, minimal learning curve
6. **Modern Stack** - NGINX, PHP 8.4, MySQL 8.0, latest tools
7. **Cloud Ready** - Native support for all major cloud providers
8. **Developer Friendly** - CLI tools, SSH access, Git integration
9. **ARM Support** - 40% better performance, 20% lower cost
10. **Active Development** - Regular updates, growing community

This document provides a complete reference for implementing CloudPanel features in new hosting control panel projects or understanding CloudPanel's comprehensive capabilities.

---

**Document Prepared By:** AI Research Assistant  
**Source Data:** https://www.cloudpanel.io  
**Research Date:** November 2, 2025  
**Format:** Markdown for easy conversion to other formats
