# Enhance Panel - Comprehensive Product Requirements Document (PRD)
## Complete Feature, Function & Functionality Analysis

**Document Version:** 1.0  
**Date:** November 2, 2025  
**Source:** Official Enhance Documentation (https://enhance.com/docs/)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Overview](#system-overview)
3. [Core Architecture](#core-architecture)
4. [Installation & Setup](#installation--setup)
5. [Server Management](#server-management)
6. [Application Role](#application-role)
7. [Database Management](#database-management)
8. [Email Services](#email-services)
9. [DNS Management](#dns-management)
10. [Backup System](#backup-system)
11. [Security Features](#security-features)
12. [WordPress Management](#wordpress-management)
13. [Website Management](#website-management)
14. [Customer & Reseller Management](#customer--reseller-management)
15. [Package Management](#package-management)
16. [Billing Integration](#billing-integration)
17. [Platform Settings](#platform-settings)
18. [Monitoring & Logs](#monitoring--logs)
19. [Migration Tools](#migration-tools)
20. [API & Automation](#api--automation)
21. [Performance Optimization](#performance-optimization)
22. [Technical Requirements](#technical-requirements)

---

## 1. Executive Summary

Enhance Panel is a modern, containerized web hosting control panel designed for hosting providers, offering multi-server cluster management with distributed roles. Built on Ubuntu 24.04 LTS with Docker containerization, it provides a comprehensive solution for managing websites, emails, databases, DNS, and backups across multiple servers.

### Key Differentiators
- **Zero per-server licensing costs** - Single license for unlimited servers
- **Containerized architecture** - All services run in isolated Docker containers
- **Flexible server roles** - Distribute services across multiple servers
- **Multi-webserver support** - Apache, Nginx, LiteSpeed, OpenLiteSpeed
- **Native WordPress toolkit** - Built-in WordPress management
- **Incremental backups** - Efficient space-saving backup system
- **Modern UI/UX** - Clean, streamlined interface

---

## 2. System Overview

### 2.1 Product Philosophy
Enhance Panel focuses on providing:
- **Simplicity** - Complex tasks performed with ease through streamlined workflows
- **Scalability** - Effortlessly scale platforms and manage load
- **Security** - Automatic SSL, containerized isolation, ModSecurity support
- **Performance** - Resource limiting, opcode caching, optimized for speed
- **Flexibility** - Run all roles on one server or spread across multiple servers

### 2.2 Target Users
- **Master Organization** - Primary installation owner with full administrative access
- **Resellers** - Customers who can create and manage sub-customers
- **End Users/Customers** - Website owners with limited administrative access
- **System Administrators** - Technical staff managing infrastructure

---

## 3. Core Architecture

### 3.1 Multi-Server Cluster Design

**Cluster Topology:**
- **Control Panel Server** - Central management server (required, single instance)
- **Application Servers** - Web hosting servers (multiple instances supported)
- **Database Servers** - MySQL/MariaDB/PostgreSQL servers (multiple instances supported)
- **Email Servers** - Mail service servers (multiple instances supported)
- **DNS Servers** - Nameserver infrastructure (minimum 2 recommended)
- **Backup Servers** - Dedicated backup storage servers (multiple instances supported)

### 3.2 Containerized Services

**Container Management:**
- All roles deployed as Docker containers
- Managed automatically by Enhance orchestration layer
- Automatic container lifecycle management (create, restart, destroy)
- Persistent data stored in volumes or host mounts
- Services isolated for security and stability

**Service Architecture:**
```
┌─────────────────────────────────────────┐
│     Control Panel Server (orchd)       │
│  ┌──────────────────────────────────┐  │
│  │   UI + API (Control Panel)       │  │
│  │   orchd (Orchestration Daemon)   │  │
│  │   authd (Authentication Service) │  │
│  │   filerd (File Manager Service)  │  │
│  └──────────────────────────────────┘  │
└─────────────────────────────────────────┘
           │
           ├─── Port 50000 ────► Application Server (appcd)
           │                      ├─ Web Server (Apache/Nginx/LiteSpeed/OLS)
           │                      ├─ PHP-FPM Containers
           │                      └─ ModSecurity (optional)
           │
           ├─── Port 50000 ────► Database Server
           │                      ├─ MySQL 8.0
           │                      ├─ MariaDB 10.11
           │                      └─ PostgreSQL 16
           │
           ├─── Port 50000 ────► Email Server
           │                      ├─ Postfix (SMTP)
           │                      ├─ Dovecot (IMAP/POP3)
           │                      └─ Rspamd (Anti-spam)
           │
           ├─── Port 50000 ────► DNS Server
           │                      └─ PowerDNS
           │
           └─── Port 50000 ────► Backup Server
                                 └─ Incremental backup service
```

### 3.3 Internal Communication

**Network Ports:**
- **Port 2087** - Control Panel Web UI (HTTPS)
- **Port 50000** - Internal RPC communication between all cluster members
- **Port 50001-50003** - Additional internal services
- **Port 50004** - File management service
- **Port 22** - SSH access for administration
- **Port 80/443** - Web traffic (Application servers)
- **Port 21** - FTP (Application servers)
- **Port 25/587/465** - SMTP (Email servers)
- **Port 143/993** - IMAP (Email servers)
- **Port 110/995** - POP3 (Email servers)
- **Port 53** - DNS queries (DNS servers)

---

## 4. Installation & Setup

### 4.1 Prerequisites

**Hardware Requirements:**
- **Minimum Specifications:**
  - 2GB RAM
  - 20GB storage (preferably on /var partition)
  - 2 CPU cores or vCPUs
  - x86_64/amd64 architecture
  
**Operating System:**
- Ubuntu 24.04 LTS Server (clean installation required)
- Full virtual machine or bare metal server
- **NOT SUPPORTED:**
  - ARM architecture
  - Containerized environments
  - Paravirtualized environments (LXD, Virtuozzo)

**Additional Requirements:**
- Valid Enhance license key
- Root SSH access or sudo privileges
- Internet connectivity for installation

**Scalability Considerations:**
- For mass hosting: Substantially more resources required
- Recommended: 100-200MB RAM per website
- Resource requirements vary by installed roles
- Most customer data stored in `/var` directory

### 4.2 Installation Process

**One-Command Installation:**
```bash
bash -c "$(curl https://install.enhance.com/install.sh)"
```

**Installation Steps:**
1. **Server Preparation:**
   - Log in as root via SSH or use `sudo -i`
   - Ensure firewall allows port 2087
   
2. **Execute Installation Command:**
   - Paste installation command into terminal
   - System installs packages via APT package manager
   - Automatic Docker and dependency installation
   
3. **Initial Access:**
   - One-time setup link displayed in terminal
   - Visit link to complete setup in browser
   - Configure admin credentials via profile button
   
4. **License Activation:**
   - Navigate to Settings → License
   - Enter license key obtained from my.enhance.com
   
5. **Control Panel Domain (Optional):**
   - Settings → Platform
   - Configure custom domain instead of IP access
   
6. **Server Addition:**
   - Servers → Add Server
   - Copy command specific to installation
   - Execute on target server as root
   - Verify server shows online (green status)

### 4.3 Post-Installation Configuration

**Essential Setup Tasks:**
- **System Generated Emails:**
  - Default: Uses local SMTP server
  - Custom SMTP can be configured
  - Required for password resets, user invites
  
- **Service Websites Configuration:**
  - Control Panel domain
  - phpMyAdmin domain
  - Webmail (Roundcube) domain
  
- **Firewall Configuration:**
  - Internal: Automatic UFW configuration
  - External: Manual configuration required
  - Ensure port 50000 allowed between cluster members

**SSO Access:**
- If admin credentials forgotten: `ecp sso` command
- Generates single sign-on link from command line

### 4.4 Cluster Expansion

**Adding Additional Servers:**
1. Navigate to Servers section
2. Click "Add Server"
3. Copy installation command
4. Execute on new server as root
5. Server automatically joins cluster

**Server Connectivity Checks:**
- Verify green status in panel
- Check external firewall permits port 50000 from control panel IP
- Test connectivity between all cluster members

---

## 5. Server Management

### 5.1 Server Roles System

**Available Roles:**
1. **Control Panel Role:**
   - Central management interface
   - Single instance per cluster
   - Cannot be removed or uninstalled
   - Automatic failover: Upcoming feature
   
2. **Application Role:**
   - Web server hosting
   - PHP processing
   - Multiple instances supported
   - Webserver options: Apache, LiteSpeed, OpenLiteSpeed, Nginx
   
3. **Database Role:**
   - Database hosting
   - Options: MySQL 8.0, MariaDB 10.11, PostgreSQL 16
   - Multiple instances supported
   - Custom my.cnf configuration
   
4. **Email Role:**
   - Complete mail server stack
   - Postfix (SMTP) + Dovecot (IMAP/POP3)
   - Rspamd for anti-spam
   - Multiple instances supported
   
5. **DNS Role:**
   - PowerDNS nameserver
   - Minimum 2 instances recommended
   - Automatic zone replication
   - Supports all cluster domains
   
6. **Backup Role:**
   - Incremental backup service
   - Efficient hard-link based storage
   - Multiple instances supported
   - Recommended: Separate datacenter

### 5.2 Role Installation

**Installation Process:**
1. **Navigate:** Servers → Select Server → Manage
2. **Add Role:** Scroll to Roles → Add Role tab
3. **Select Services:** Check desired roles
4. **Provision:** Click "Add services"
5. **Background Processing:** Notification on completion

**Important Notes:**
- New roles don't auto-map to existing websites
- Use "Move Server" option for mapping
- Installation can take several minutes
- Safe to navigate away during installation

### 5.3 Role Configuration

**Global Service Settings:**
- Default: All roles inherit global settings
- Location: Settings → Service Settings
- Configurable per service type

**Per-Server Overrides:**

**Email Role Overrides:**
- SMTP Settings
- Smart host configuration
- Relay configuration

**Application Role Overrides:**
- PHP-FPM settings (per-website override available)
- PHP INI settings (per-website override available)
- ModSecurity configuration
- Webserver-specific settings

**Database Role Overrides:**
- my.cnf directives
- Custom configuration
- Performance tuning

**Override Configuration Steps:**
1. Open Servers in sidebar
2. Select "Manage" on target server
3. Navigate to desired Role tab
4. Click "Settings"
5. Apply custom configuration

### 5.4 Role Lifecycle Management

**Disabling Roles:**
- Prevents new website placement
- Existing websites unaffected
- Can be re-enabled anytime

**Deleting Roles:**
- Removes all containers and data
- Cannot delete if websites mapped
- Must migrate websites first
- Irreversible action

**Restarting Roles:**

**Graceful Restart:**
- Waits for processes to finish
- Brief downtime during restart
- Notifications on start/finish
- Recommended for production

**Forceful Restart:**
- Immediate process termination
- Brief downtime during restart
- Use when graceful fails
- More disruptive

### 5.5 Webserver Management

**Supported Webservers:**

**Apache:**
- Full .htaccess support
- ModSecurity compatible
- Traditional hosting choice
- Proven stability

**Nginx:**
- High performance
- No .htaccess support
- ModSecurity compatible
- Modern infrastructure choice
- Recommended for personal hosting

**LiteSpeed (Commercial):**
- Drop-in Apache replacement
- Full .htaccess support
- LSCache plugin support
- ModSecurity rules
- ~10ms TTFB with cache hits
- Superior WordPress performance

**OpenLiteSpeed (Free):**
- LiteSpeed alternative
- LSCache plugin support
- No .htaccess support
- No ModSecurity support
- Good performance

**Switching Webservers:**
- No configuration changes required
- Zero-downtime switch
- Steps:
  1. Servers → Manage Server
  2. Roles → Application tab
  3. Switch webserver option
  4. Select new type
  5. Confirm change

### 5.6 Server Metadata

**Server Hostname:**
- Friendly name for management
- Email server identification
- Alternative to IP addresses
- Format: hostname.yourdomain.com
- Requires valid DNS A record
- PTR record recommended

**Server Groups:**
- Logical organization of servers
- Filter and manage by group
- Custom group creation
- Multi-server operations

**Server Screenshots:**
- Visual identification
- Thumbnail generation
- screenshotd service handles generation

### 5.7 Server SSL Certificates

**Internal Communication:**
- mTLS certificates for inter-server communication
- Automatic certificate management
- Secure RPC calls

**Public SSL:**
- Let's Encrypt for public services
- Automatic renewal
- Per-website certificates

### 5.8 IPv6 Support

**Configuration:**
- Full IPv6 stack support
- Dual-stack operation (IPv4 + IPv6)
- Automatic AAAA record creation
- Per-server IPv6 assignment

---

## 6. Application Role

### 6.1 PHP Management

**PHP Versions:**
- Multiple concurrent PHP versions
- Per-website PHP version selection
- Supported versions vary by release
- PHP 8.3 referenced in documentation

**PHP-FPM Configuration:**

**Global PHP-FPM Settings:**
- Master configuration template
- Applied to all websites by default
- Configurable parameters:
  - pm (process manager)
  - pm.max_children
  - pm.start_servers
  - pm.min_spare_servers
  - pm.max_spare_servers
  - pm.max_requests

**Per-Server PHP-FPM Override:**
- Server-specific PHP-FPM configuration
- Overrides global settings
- Applied to all websites on server

**Per-Website PHP-FPM Override:**
- Ultimate granular control
- Website-specific tuning
- Performance optimization

**PHP INI Configuration:**

**Three-Tier Configuration:**
1. **Global PHP INI:**
   - Master configuration
   - Default for all websites
   
2. **Per-Server Override:**
   - Server-specific PHP settings
   
3. **Per-Website Override:**
   - Available to:
     - Master organization
     - Resellers
     - End users (if permitted)

**Common PHP INI Directives:**
- memory_limit
- upload_max_filesize
- post_max_size
- max_execution_time
- max_input_vars
- display_errors
- error_reporting

### 6.2 Application Logs

**Access Logs:**
- Per-website access logs
- Standard Apache/Nginx log format
- Rotation and retention policies

**Error Logs:**
- PHP error logging
- Web server error logs
- Application debugging

### 6.3 Screenshotd Service

**Website Screenshots:**
- Automatic thumbnail generation
- Visual website identification
- Preview functionality
- Background processing

---

## 7. Database Management

### 7.1 Database Engines

**Supported Database Systems:**

**MySQL 8.0:**
- Industry standard
- Strong compatibility
- Modern features
- Authentication: caching_sha2_password

**MariaDB 10.11:**
- MySQL fork
- Query cache support (recommended)
- Enhanced performance
- Authentication: mysql_native_password
- Better performance for hosting

**PostgreSQL 16:**
- Advanced features
- ACID compliance
- Complex query optimization

### 7.2 Database Configuration

**my.cnf Customization:**
- Custom directives supported
- Per-server configuration
- Warning: Incorrect settings cause data corruption
- Validation before application
- MySQL restart verification

**Configuration Access:**
1. Servers → Manage Server
2. Roles → Database tab
3. Settings → my.cnf
4. Apply custom directives

### 7.3 phpMyAdmin Integration

**Automatic Installation:**
- Fetches and installs automatically
- Service website type
- Control panel server placement
- Cannot be moved to other servers

**Features:**
- Single sign-on (SSO) capability
- Direct access from customer dashboard
- No credential entry required
- Automatic server selection
- Full database management

**Domain Configuration:**
1. Settings → Platform
2. Control panel website domains
3. Edit phpMyAdmin domain
4. Save configuration
5. Domain added as alias if changing

### 7.4 Database User Management

**Authentication Methods:**
- MySQL: caching_sha2_password
- MariaDB: mysql_native_password
- Migration considerations between engines

**Per-Website Database:**
- Isolated database access
- Automatic user creation
- Permission management
- Security isolation

---

## 8. Email Services

### 8.1 Email Stack Architecture

**Components:**

**Postfix (SMTP Server):**
- Incoming mail reception (Port 25)
- Outgoing mail delivery (Port 587/465)
- Smart host support
- Relay configuration

**Dovecot (IMAP/POP3):**
- IMAP: Ports 143 (plain), 993 (SSL)
- POP3: Ports 110 (plain), 995 (SSL)
- Mailbox management
- Mail delivery agent

**Rspamd:**
- Anti-spam engine
- DKIM signing
- SPF checking
- Bayesian filtering

### 8.2 Email Security

**SPF (Sender Policy Framework):**
- Automatic SPF record creation
- Per-domain configuration
- Enhance DNS integration
- External DNS support

**DKIM (DomainKeys Identified Mail):**
- Domain authentication
- Message signing
- Per-domain enablement
- Automatic validation with Enhance DNS
- Manual configuration for external DNS

**SSL/TLS:**
- Let's Encrypt automatic certificates
- Secure SMTP (STARTTLS)
- IMAPS/POP3S support

### 8.3 Email Accounts

**Account Management:**
- Unlimited email accounts (package-dependent)
- Quota management
- Password management
- Mailbox statistics

**Features:**
- Catch-all addresses
- Email forwarding
- Auto-responders (if package allows)
- Mailbox size limits

### 8.4 Webmail (Roundcube)

**Automatic Installation:**
- Downloads and installs Roundcube
- Service website type
- Control panel server placement
- Integrated authentication

**Access Methods:**
1. Direct: https://webmail.yourdomain.com
2. Per-domain: https://mail.{customer_domain}
3. SSO link from customer dashboard

**Customization:**
- Plugin support
- Plugins persist through updates
- Installation via control panel file manager
- Configuration: public_html/config/config.inc.php

**Domain Configuration:**
1. Settings → Platform
2. Control panel website domains
3. Edit Roundcube domain
4. Save configuration
5. Domain added as alias if changing

### 8.5 Smart Host Configuration

**Purpose:**
- Outbound SMTP relay
- Filtering/scrubbing services
- Third-party email gateways
- Spam prevention services

**Configuration:**
- Per-server Email role setting
- Override global SMTP settings
- Authentication support
- Port configuration

**Common Use Cases:**
- SendGrid relay
- Mailgun SMTP
- AWS SES
- Office 365 relay

### 8.6 Email Backup

**Automatic Backup:**
- Email accounts backed up by default
- If backup role installed
- Incremental backup strategy
- Per-mailbox restoration

### 8.7 Email Logs

**Log Access:**
- **Control Panel Server:**
  ```bash
  journalctl -u orchd
  ```

- **Email Server:**
  ```bash
  journalctl -u appcd
  ```

**Troubleshooting:**
- Inbound issues: Check MX records, mail IP
- Outbound issues: SPF/DKIM, PTR records
- Port 25 blocking: Test with `telnet smtp.gmail.com 25`
- Smart host fallback if port 25 blocked

---

## 9. DNS Management

### 9.1 DNS Role Architecture

**PowerDNS Integration:**
- Industry-standard DNS server
- Automatic zone replication
- Multi-server redundancy
- DNSSEC support (if enabled)

**Cluster Configuration:**
- Minimum 2 DNS servers recommended
- Automatic synchronization
- Load distribution
- High availability

**Installation Requirements:**
- Can be installed on any server
- Multiple instances supported
- Automatic zone serving
- No manual zone transfers

### 9.2 DNS Zone Management

**Automatic Zone Creation:**
- Zone created on website/domain addition
- Standard records auto-generated:
  - A record (IPv4)
  - AAAA record (IPv6, if configured)
  - MX record (if email enabled)
  - SPF record (if email enabled)
  - www CNAME (implicit alias)

**Zone Editor:**
- Available to customers (if package allows)
- Record types supported:
  - A
  - AAAA
  - CNAME
  - MX
  - TXT
  - SRV
  - NS
  - CAA

**DNS Zone Templating:**

**Purpose:**
- Pre-configure default DNS records
- External service integration
- Consistent record deployment
- Automation of common configurations

**Use Cases:**
1. **Third-party email gateways:**
   - Smart host SPF records
   - DKIM records
   - Custom MX records

2. **External WAF/CDN:**
   - Override root A record
   - Point to WAF/CDN IP
   - Traffic routing through proxy

3. **Custom services:**
   - SaaS integrations
   - API endpoints
   - Verification records

**Template Configuration:**
1. Settings → Platform
2. DNS zone templating section
3. Add record button
4. Configure record type and value
5. Save template

**Variable Support:**
- **$$origin$$** - Replaced with customer domain
- Example: `$$origin$$.yourservice.com`
- Dynamic record generation

**Template Application:**
- Applied to new websites/domains only
- Existing zones not affected
- No retroactive application

### 9.3 Cloudflare Integration

**Native Integration:**
- Direct Cloudflare API connection
- Automatic zone management
- Enhance as DNS provider
- Cloudflare as proxy/CDN

**Configuration Requirements:**
- Cloudflare API key
- Domain verification
- Nameserver delegation
- Zone management permissions

### 9.4 External DNS

**Support:**
- Customer can use external DNS
- Manual A/AAAA record configuration required
- MX record setup for email
- SPF/DKIM manual configuration
- CAA records for SSL

### 9.5 Gmail DNS Auto-Configuration

**Feature:**
- Automatic Google Workspace integration
- MX record configuration
- SPF record setup
- DKIM key exchange
- Package permission required

### 9.6 DNS Logs

**Log Access:**
- PowerDNS query logs
- Zone update logs
- DNSSEC signing logs (if enabled)
- Troubleshooting information

---

## 10. Backup System

### 10.1 Backup Architecture

**Incremental Backup System:**
- Hard-link based storage
- Space-efficient design
- Fast backup creation
- Quick restoration process

**Backup Storage Structure:**
```
/[backup_mount]/[website_uuid]/
├── snapshot-1678048114772/  (Unix timestamp)
│   ├── files/
│   ├── databases/
│   └── emails/
├── snapshot-1678134514772/
└── current → snapshot-[latest]/  (symlink)
```

### 10.2 Backup Role Installation

**Installation Process:**
1. Servers → Add Server or select existing
2. Manage → Roles → Add Role
3. Select Backup checkbox
4. Choose directory with adequate space
5. Click "Add Role"

**Recommendations:**
- **Dedicated server:** Standalone backup server
- **Separate datacenter:** Enhanced redundancy
- **Storage sizing:**
  - Accommodate multiple snapshots
  - Account for growth
  - Formula: (websites × 30) × average_db_size

**Firewall Requirements:**
- Port 50000 TCP from all cluster members
- Port 22 TCP from all cluster members

### 10.3 Global Backup Settings

**Configuration Location:**
- Settings → Service Settings → Backups

**Backup Scheduling:**
- **Delay first backup:** Grace period for new websites
- **Minimum backup age:** Interval between backups
- **Maximum backup age:** Target maximum interval
- **Allowed backup hours:** Scheduled backup windows
  - Uses control panel server timezone
  - Specify hours when backups can run

**Backup Retention:**
- **Maximum backup retention:** Days to keep backups
- Automatic cleanup of old backups
- Configurable per installation

**Performance Settings:**
- **Maximum concurrent backups:** Per-server limit
- Prevent resource exhaustion
- Balance backup speed vs. system load

### 10.4 Backup Operations

**Manual Backups:**
- On-demand backup creation
- Immediate execution
- Package permission: "Allow manual backups"

**Automatic Backups:**
- Scheduled execution
- Based on retention policies
- Background operation
- No customer intervention

**Backup Contents:**

**Website Backups:**
- All website files
- Document root
- User files
- Configuration files

**Database Backups:**
- All website databases
- SQL dumps
- Consistent snapshots

**Email Backups:**
- All email accounts
- Mailbox contents
- Email configuration
- Default: Enabled if backup role installed

### 10.5 Restoration Process

**Restoration Types:**

**1. Email Only:**
- Restores all email accounts
- Mailbox contents
- Email configuration
- Excludes files and databases

**2. Website Only:**
- All files
- All databases
- Excludes email accounts

**3. Custom Restoration:**
- Selective restoration
- Options:
  - All files
  - Specific databases
  - All email accounts
- Maximum flexibility

**Restoration Steps:**
1. Website dashboard
2. Advanced dropdown → Backups
3. Select backup snapshot
4. Click "Restore"
5. Choose restoration type
6. Confirm action

**Important Warnings:**
- Changes after backup will be lost
- May not be recoverable
- Downtime during restoration
- Test in staging if possible

### 10.6 Manual Backup Restoration

**Command-Line Access:**
- SSH to backup server
- Navigate to backup directory
- Restore specific files or databases

**Backup Location:**
```
/[backup_mount]/[website_uuid]/snapshot-[timestamp]/
```

**Use Cases:**
- Granular file restoration
- Partial database recovery
- Bypass UI limitations
- Root-owned file issues

**SSH Key Setup:**
- Generate key on backup server
- Add public key to target website
- Secure file transfer
- Manual rsync operations

### 10.7 S3 Backup Integration (Beta)

**S3-Compatible Storage:**
- Third-party S3 providers
- Backblaze B2
- AWS S3
- Wasabi
- DigitalOcean Spaces
- Any S3-compatible storage

**Configuration:**
1. Integrations → Backups tab
2. Click "Connect"
3. Enter S3 credentials:
   - Access Key ID
   - Secret Access Key
   - Endpoint URL
   - Bucket name
   - Region

**Backup Strategy:**
- Full tar.gz archives
- More storage consumption vs. incremental
- Can run alongside Enhance backup role
- Inherits backup role settings

**Advantages:**
- Off-site storage
- Provider redundancy
- Compliance requirements
- External disaster recovery

### 10.8 Backup Logs

**Log Locations:**
- **Control Panel:**
  ```bash
  journalctl -u orchd
  ```

- **Source Server:**
  ```bash
  journalctl -u appcd
  ```

**Troubleshooting:**
- Failed backups: File permission issues
- Check for root-owned files
- Verify backup user permissions
- Website home directory ownership

---

## 11. Security Features

### 11.1 Container Isolation

**Containerized Architecture:**
- All roles in separate containers
- No access to website files
- Process isolation
- Resource isolation

**Security Benefits:**
- Email services isolated from web
- Database isolated from application
- Limited attack surface
- Compromised container doesn't affect others

### 11.2 SSL/TLS Management

**Let's Encrypt Integration:**
- Automatic certificate provisioning
- Auto-renewal before expiration
- All websites SSL-protected by default
- Zero manual intervention

**Certificate Management:**
- Per-website certificates
- Wildcard support (if DNS managed by Enhance)
- SNI (Server Name Indication)
- HTTP to HTTPS redirection

**Third-Party SSL:**
- Custom certificate upload
- Private key management
- Intermediate certificate chain
- Manual renewal process

### 11.3 ModSecurity WAF

**Web Application Firewall:**
- Apache and Nginx support
- Per-domain enable/disable
- OWASP Core Rule Set (CRS)
- Custom rule support

**Installation:**
1. Servers → Select server
2. Server management → Roles
3. Application → Settings
4. ModSecurity section
5. Enable ModSecurity

**Global Configuration:**
- Per-server configuration
- Custom rules addition
- Rule override capability
- OWASP CRS disable option
- Include external rule files

**Configuration File:**
- Location: Custom config directory
- Include path: `/etc/modsecurity.d`
- Warning: Invalid syntax breaks web server

**Per-Domain Control:**
1. Websites → Select website
2. Website dashboard → Security
3. Enable/Disable ModSecurity
4. Prerequisites: Server-level enablement

**Important Notes:**
- Enabled by default on all domains
- Apache and Nginx only
- Not supported on LiteSpeed
- Not supported on OpenLiteSpeed

### 11.4 Two-Factor Authentication (2FA)

**Support:**
- Admin account 2FA
- TOTP (Time-based One-Time Password)
- Authenticator app integration
- Backup codes

### 11.5 Brute Force Protection

**Features:**
- Login attempt limiting
- IP-based blocking
- Automatic unlock after time period
- Configurable thresholds

**Configuration:**
- Settings → Platform → Brute Force Protection
- Max attempts before lockout
- Lockout duration
- Whitelist IPs

### 11.6 Admin Lockdown

**IP Restriction:**
- Restrict control panel access by IP
- Multiple IP whitelist
- CIDR notation support
- Enhanced security for sensitive installations

### 11.7 Server SSL Certificates

**Internal mTLS:**
- Mutual TLS for inter-server communication
- Automatic certificate generation
- Certificate rotation
- Secure RPC calls

### 11.8 UFW Firewall Integration

**Automatic Management:**
- UFW (Uncomplicated Firewall)
- Automatic rule generation
- Port management per role
- Safe default configuration

**Manual Overrides:**
- Custom UFW rules supported
- External firewall compatibility
- Required ports documented

---

## 12. WordPress Management

### 12.1 WordPress Toolkit Overview

**Native Integration:**
- Built-in WordPress installer
- No third-party plugins required
- Supported versions: 5.59 - 6.4+
- Free bundled toolkit

**Management Interface:**
- Manage without leaving control panel
- Central dashboard for all installs
- Bulk operations support
- Package-level feature control

### 12.2 WordPress Installation

**Installation Methods:**

**1. New Website with WordPress:**
- Websites → Add website
- Select "Install an app"
- Complete form with WP details
- Click Add

**2. Existing Website:**
- Select website
- WordPress section
- Install WordPress button
- Configure settings

**Installation Options:**
- Latest WordPress version
- Admin username and password
- Database auto-creation
- Instant deployment

**Domain Support:**
- Primary domain
- Addon domains
- Alias domains
- Subdomains

**Prerequisites:**
- Package permission: "Allow software installer"
- Available disk space
- PHP compatibility
- Database availability

### 12.3 WordPress Management Features

**Plugin Manager:**
- View installed plugins
- Search and install new plugins
- Bulk plugin updates
- Activate/Deactivate plugins
- Delete plugins
- Security recommendations

**Theme Manager:**
- Installed themes view
- Theme marketplace access
- Theme installation
- Theme activation
- Theme updates
- Theme deletion

**Core Auto-Updates:**
- Enable/Disable per site
- Major version updates
- Minor security updates
- Manual update option
- Backup before update

**User Manager:**
- WordPress user accounts
- Role assignment
- Password management
- User creation/deletion
- Bulk user operations

**Debug Mode:**
- Enable/Disable WP_DEBUG
- Error logging
- Development mode
- Troubleshooting tool

**Restrict Admin Access:**
- IP whitelist for /wp-admin
- Enhanced security
- Multiple IP support
- Bypass for authorized users

### 12.4 WordPress Discovery

**Discover Existing Installations:**
- Scan website for WordPress
- Automatic detection
- Import to management interface
- Bulk discovery

### 12.5 WordPress Optimization

**Performance Recommendations:**

**Infrastructure:**
- **Storage:** NVMe SSD preferred
- **CPU:** Modern CPU, high clock speed
- **RAM:** Adequate per-site allocation

**Opcode Caching:**
- Enable OPcache
- Reduce script compilation
- Significant performance improvement

**LiteSpeed + LSCache:**
- Switch to LiteSpeed/OpenLiteSpeed
- Install LSCache plugin
- ~10ms TTFB for cache hits
- Dramatic performance improvement

**Server Distribution:**
- Avoid server overload
- Distribute sites across servers
- No per-server license cost
- Prevent "bad neighbor" effect

**Package-Level Controls:**
- Enable/Disable toolkit features per package
- Control customer access
- Feature restrictions
- Security policies

---

## 13. Website Management

### 13.1 Website Creation

**Creation Methods:**

**1. Blank Website:**
- Empty document root
- Manual file upload
- Complete control

**2. Clone Existing:**
- Duplicate existing website
- Copy files and databases
- Optionally copy email accounts
- Fast deployment

**3. Install Application:**
- WordPress installation
- Other CMS support
- Pre-configured environment

**4. Import from cPanel:**
- cPanel backup import
- Account migration
- Full transfer utility

**5. Import from Plesk:**
- Plesk backup import
- Migration tool
- Streamlined process

### 13.2 Website Settings

**Basic Configuration:**
- Primary domain
- Document root
- PHP version
- Database assignment
- Email accounts
- Disk quota
- Bandwidth limit

**Server Placement:**
- Application server selection
- Database server selection
- Email server selection
- DNS server selection
- Backup server selection

### 13.3 Domain Management

**Domain Types:**

**Primary Domain:**
- Main website domain
- Document root association
- Cannot be deleted

**Addon Domains:**
- Additional fully-functional domains
- Separate document root option
- Own email accounts
- Subdomain creation capability

**Alias Domains:**
- Mirror primary domain content
- Same document root
- Alternative access URLs
- SEO considerations

**Subdomains:**
- Subdomain creation
- Separate or shared document root
- Unlimited (package-dependent)
- Full functionality

**Implicit www Alias:**
- Automatic www subdomain
- All websites include by default
- No configuration needed
- cPanel/Plesk compatibility

**Prerequisites:**
- Package: "Number of domain aliases"
- Package: "Allow DNS zone editor"
- Package: "Allow Gmail DNS Auto configuration"

### 13.4 Domain Redirects

**Redirect Types:**
- 301 Permanent
- 302 Temporary
- Domain forwarding
- HTTPS enforcement

**Configuration:**
- Per-domain settings
- Wildcard support
- Path preservation
- Query string handling

### 13.5 Cloning Websites

**Cloning Process:**
1. Websites → Source website
2. Advanced → Clone
3. Select destination
4. Choose cloning options:
   - Files
   - Databases
   - Email accounts
5. Execute clone

**Use Cases:**
- Testing environments
- Development copies
- Customer migrations
- Rapid deployment

### 13.6 Staging Websites

**Staging Infrastructure:**

**Configuration Requirements:**
- **Settings → Platform → Platform domains**
- Staging domain configuration
- Nameserver delegation to cluster
- DNS A record to application server

**Staging Domain Format:**
- `[random].stagingdomain.com`
- Automatic subdomain generation
- No customer domain consumption

**Features:**
- Full website functionality
- Database access
- PHP execution
- No email accounts
- No domain mapping

**Creation:**
1. Websites → Add website
2. Select creation method
3. Check "Make this a staging site"
4. Complete configuration

**Identification:**
- STG tag in website listing
- Visual distinction
- Grouped with production site

**Push to Production:**
- Replace production content
- Database migration
- File synchronization
- DNS switchover

**Reseller Capability:**
- Resellers set own staging domain
- Independent from master staging
- Customer isolation

### 13.7 Preview Domain

**Functionality:**
- Preview website before DNS propagation
- Test configuration
- Development access
- No DNS required

### 13.8 Service Websites

**Definition:**
- Dedicated internal service sites
- Control panel infrastructure
- Cannot be moved from control panel server
- Tag-prefixed domains

**Types:**

**1. Control Panel Website (CTL tag):**
- Customer login interface
- Master org configured during install
- Reseller configures on first login

**2. phpMyAdmin Website (PMA tag):**
- Database management interface
- Automatic software installation
- SSO from dashboard
- Server-auto-selection

**3. Webmail Website (WML tag):**
- Roundcube interface
- Email management
- Auto-login capability
- Unified access for all servers

**Domain Management:**
1. Settings → Platform
2. Control panel website domains
3. Edit respective domain
4. Save configuration
5. Old domain remains as alias

### 13.9 Custom Virtualhosts

**Availability:**
- Master Organization only
- Resellers (with permission)
- Not available to end users

**Functionality:**
- Inject custom directives
- Apache and Nginx support
- Below generated config
- Advanced customization

**Configuration:**
- Per-website basis
- Webserver-specific syntax
- Validated before application

### 13.10 Redis Support

**In-Memory Data Store:**
- Object caching
- Session storage
- WordPress optimization
- Application acceleration

**Configuration:**
- Per-website enablement
- Connection details provided
- PHP Redis extension
- No customer configuration needed

### 13.11 IonCube Loader

**PHP Encryption Support:**
- Decode encrypted PHP
- Licensed software execution
- Per-PHP version availability
- Automatic installation option

### 13.12 Website Deletion

**Deletion Process:**
- All data permanently removed
- Files deleted
- Databases dropped
- Email accounts removed
- DNS zones deleted

**Safety Features:**
- Confirmation required
- No accidental deletion
- Irreversible action warning

### 13.13 File Manager

**Built-in File Management:**
- Web-based interface
- Upload/Download files
- Edit files
- Permission management
- Extract archives
- Compress files

**Features:**
- Syntax highlighting
- Search functionality
- Bulk operations
- Trash/Restore

### 13.14 Website Statistics

**Metrics:**
- Disk usage
- Bandwidth usage
- Inode count
- Database size
- Email account count

### 13.15 Website Logs

**Available Logs:**
- Access logs
- Error logs
- PHP error logs
- Email logs

**Access Methods:**
- Download via UI
- Real-time viewing
- Log rotation
- Retention period

### 13.16 Node.js Support

**Features:**
- Node.js application hosting
- Ghost CMS installation
- Version management
- PM2 process management

**Ghost CMS:**
- One-click installation
- Automatic configuration
- Nginx proxy setup
- SSL integration

---

## 14. Customer & Reseller Management

### 14.1 Customer Hierarchy

**Organizational Structure:**
```
Master Organization
├── Direct Customers
│   ├── Websites
│   ├── Databases
│   └── Email
└── Resellers
    └── Reseller Customers
        ├── Websites
        ├── Databases
        └── Email
```

### 14.2 Master Organization

**Capabilities:**
- Full platform control
- All server management
- Package creation
- Reseller creation
- Global settings management
- Direct customer management
- Billing integration

**Unique Access:**
- Server role management
- Cluster configuration
- Update management
- Platform settings

### 14.3 Resellers

**Reseller Features:**
- White-label capability
- Custom branding
- Own customer base
- Package creation
- Service website management
- Impersonation capability

**Reseller Setup:**
1. Customers → Add Reseller
2. Configure reseller details
3. Set resource allocations
4. Assign packages (if any)
5. Send login credentials

**Reseller Branding:**
- Custom control panel domain
- Logo upload
- Color scheme customization
- Custom email templates
- Unique phpMyAdmin domain
- Unique webmail domain

**Reseller Limitations:**
- No server management
- No cluster configuration
- No core updates
- No platform settings
- Package quantity limits (if set)

### 14.4 End Customer Management

**Customer Creation:**
1. Customers → Add Customer
2. Enter customer details:
   - Name
   - Email
   - Username
   - Password
   - Contact info
3. Assign to reseller (if applicable)
4. Subscribe to package
5. Send credentials

**Customer Details:**
- Personal information
- Contact details
- Subscription status
- Resource usage
- Website list
- Package assignment

### 14.5 Impersonation

**Master Organization:**
- Impersonate any customer
- Impersonate any reseller
- See exact customer view
- Troubleshooting tool

**Reseller:**
- Impersonate their customers only
- Cannot impersonate other resellers
- See customer perspective

**Security:**
- Audit trail
- No password access
- Session-based
- Automatic logout

### 14.6 Subscription Management

**Package Subscription:**
- Subscribe customer to package
- Multiple packages per customer
- Upgrade/Downgrade capability
- Package switching

**Resource Allocation:**
- Disk space
- Bandwidth
- Websites
- Databases
- Email accounts
- Domains/Subdomains

### 14.7 Customer Portal

**Customer Dashboard:**
- Website listing
- Resource usage graphs
- Package details
- Support access
- Billing info (if integrated)

**Customer Actions:**
- Create websites (if package allows)
- Manage databases
- Configure email
- Install WordPress
- File management
- Backup/Restore
- Domain management

---

## 15. Package Management

### 15.1 Package Concept

**Definition:**
- Collection of resources and tools
- Customer subscription unit
- Multiple packages per customer
- Flexible allocation

**Package Types:**

**Shared Hosting:**
- Multiple websites per package
- Shared server resources
- Resource limits
- Standard offering

**Dedicated Hosting:**
- Full server allocation
- No resource sharing
- High-performance
- Premium pricing

**VPS Packages:**
- Virtual private server
- Isolated resources
- Scalable
- Custom configuration

**Location-Based Packages:**
- Geographic server selection
- Compliance requirements
- Latency optimization
- Regional presence

**WordPress Packages:**
- Pre-installed WordPress
- Optimized configuration
- WordPress toolkit enabled
- Themed packages

### 15.2 Package Creation

**Creation Process:**
1. Packages → Add Hosting Package
2. Configure basic details:
   - Package name
   - Description
   - Billing cycle
3. Set resource limits
4. Enable/Disable features
5. Configure restrictions
6. Save package

### 15.3 Resource Limits

**System Resources:**

**CPU Limiting:**
- Per-website CPU percentage
- Prevents CPU hogging
- Consistent performance
- Fair resource distribution

**I/O Bandwidth:**
- Disk I/O limiting
- MB/s restrictions
- Prevent disk thrashing
- Server stability

**IOPS (I/O Operations Per Second):**
- Read/Write operation limits
- Fine-grained control
- Performance consistency
- Database optimization

**Process Count (nproc):**
- Maximum processes per website
- Prevent fork bombs
- Resource protection
- Stability assurance

**Memory Limiting:**
- RAM allocation per website
- OOM protection
- Shared hosting fairness
- Prevent memory leaks

**Disk Space:**
- Storage quota
- Includes:
  - Website files
  - Databases
  - Email
  - Backups (if counted)

**Bandwidth:**
- Monthly transfer limit
- Incoming + Outgoing
- Overage handling
- Unlimited option

### 15.4 Website Allowances

**Quantity Restrictions:**
- Number of websites
- Number of subdomains
- Number of domain aliases
- Number of addon domains

**Database Limits:**
- Number of databases
- Database size limits
- MySQL/MariaDB/PostgreSQL selection

**Email Limits:**
- Number of email accounts
- Mailbox size per account
- Total email storage

### 15.5 Feature Permissions

**Website Features:**
- [ ] Allow software installer (WordPress, etc.)
- [ ] Allow staging websites
- [ ] Allow website cloning
- [ ] Allow cron jobs
- [ ] Allow SSH access
- [ ] Allow FTP access
- [ ] Allow Git deployment
- [ ] Allow Node.js applications

**WordPress Toolkit:**
- [ ] Enable WordPress toolkit
- [ ] Allow plugin management
- [ ] Allow theme management
- [ ] Allow core updates
- [ ] Allow user management
- [ ] Enable debug mode
- [ ] Restrict admin access

**Database Features:**
- [ ] Allow phpMyAdmin access
- [ ] Allow database creation
- [ ] Allow database deletion
- [ ] Remote database access

**DNS Features:**
- [ ] Allow DNS zone editor
- [ ] Gmail DNS auto-configuration
- [ ] Custom MX records
- [ ] DNSSEC management

**Email Features:**
- [ ] Email services enabled
- [ ] Webmail access
- [ ] Email forwarding
- [ ] Auto-responders
- [ ] IMAP/POP3 access
- [ ] Spam filter control

**Backup Features:**
- [ ] Allow manual backups
- [ ] Access backup restoration
- [ ] Download backups
- [ ] Scheduled backups

**Security Features:**
- [ ] Allow SSL management
- [ ] Third-party SSL upload
- [ ] ModSecurity per-domain
- [ ] IP whitelisting

**File Management:**
- [ ] File manager access
- [ ] FTP accounts
- [ ] SSH/SFTP access
- [ ] Archive extraction

### 15.6 Package Assignment

**Subscription:**
1. Customers → Select customer
2. Subscriptions tab
3. Add subscription
4. Select package
5. Confirm assignment

**Multiple Packages:**
- Customer can have multiple subscriptions
- Different packages simultaneously
- Resource aggregation
- Flexible offerings

### 15.7 Package Upgrades/Downgrades

**Upgrade Process:**
1. Customer → Subscriptions
2. Select subscription
3. Upgrade/Downgrade button
4. Choose new package
5. Confirm change

**Handling:**
- Resource adjustment
- Website reassignment
- Automatic enforcement
- No website interruption

**Prorated Billing:**
- Integration dependent
- WHMCS calculation
- Immediate effect
- Credit/Charge adjustment

### 15.8 Package-to-Website Mapping

**Moving Between Packages:**
1. Websites → Select website
2. Move server option
3. Package selection
4. New package assignment
5. Confirm move

---

## 16. Billing Integration

### 16.1 Supported Platforms

**WHMCS:**
- Industry standard
- Product provisioning
- Automated billing
- Domain reselling
- Client management
- Support ticket system

**Upmind:**
- Client management
- Product management
- Automated billing
- Small-medium businesses
- Modern interface

**HostBill:**
- Automation platform
- Client management
- Help desk system
- Domain management
- Automated billing

**Blesta:**
- Billing platform
- Web hosting focused
- Order management
- Client portal

### 16.2 WHMCS Integration

**Module Features:**
- Hosting subscription creation
- Suspension automation
- Termination handling
- Package synchronization
- Multi-subscription support
- Website auto-creation

**Installation:**
1. Download WHMCS module
2. Extract to `modules/servers/enhance/`
3. Run `composer install` in api directory
4. Create SuperAdmin access token in Enhance
5. Configure WHMCS server connection:
   - Name: Any description
   - Hostname: Control panel domain
   - Username: Enhance orgId
   - Access Hash: Access token

**Custom Fields:**
- `enhanceSubscriptionId`: Per-product subscription tracking
- `enhOrgId`: Per-client organization ID
- Automatic creation on first subscription

**Workflow:**
1. WHMCS order created
2. Customer doesn't exist: Create in Enhance
3. Custom `enhOrgId` field created for client
4. Package subscription created in Enhance
5. Website automatically created (optional)
6. WHMCS product provisioned

**Suspension/Termination:**
- Automatic suspension on non-payment
- Website disabled
- Email continues (configurable)
- Termination removes all data

**Multiple Subscriptions:**
- Customer can have multiple Enhance packages
- Each WHMCS product maps to Enhance package
- Separate subscription tracking

**Website Creation Behavior:**
- Module attempts automatic website creation
- Domain from WHMCS order
- Package from product configuration
- Error handling for failures

**Upgrading Module:**
- Custom field rename: `subscriptionId` → `enhanceSubscriptionId`
- Manual update for existing subscriptions
- Composer dependency updates

### 16.3 API Access

**SuperAdmin Token:**
- Created in Enhance control panel
- Full API access
- Secure token storage
- Rotation capability

**API Endpoints:**
- RESTful API
- JSON responses
- Authentication via token
- Comprehensive operations

---

## 17. Platform Settings

### 17.1 Platform Configuration

**Settings Location:**
- Settings → Platform

**Configuration Categories:**

**Control Panel Domains:**
- Control panel domain
- phpMyAdmin domain
- Webmail domain
- Staging domain
- Preview domain

**Branding Settings:**
- Control panel logo
- Favicon
- Color scheme
- Login page customization
- Email templates

**Platform Domains:**
- Service domain configuration
- Subdomain management
- DNS requirements

### 17.2 System Generated Emails

**Email Types:**
- Password reset emails
- User invitation emails
- System notifications
- Welcome emails

**SMTP Configuration:**
- Default: Local SMTP server
- Custom SMTP support:
  - SMTP host
  - SMTP port
  - Authentication
  - Encryption (SSL/TLS)

**Requirements:**
- Valid sender domain
- SPF record
- DKIM configuration (recommended)

### 17.3 Brute Force Protection

**Configuration:**
- Maximum login attempts
- Lockout duration
- Whitelist IPs
- Alert notifications

### 17.4 Prohibited Domains

**Purpose:**
- Prevent specific domain registration
- Trademark protection
- Reserved names
- Blacklist management

**Configuration:**
- Domain list management
- Wildcard support
- Reason notes

### 17.5 Language Selection

**Interface Languages:**
- Multi-language support
- User preference
- Per-customer language
- Locale management

### 17.6 Website Server Placement

**Automatic Placement:**
- Algorithm-based assignment
- Load balancing
- Resource availability
- Geographic considerations

**Manual Placement:**
- Override automation
- Specific server selection
- Customer requirements
- Compliance needs

### 17.7 Admin Lockdown

**IP Restrictions:**
- Control panel access by IP
- Whitelist management
- Multiple IP support
- CIDR notation
- VPN considerations

---

## 18. Monitoring & Logs

### 18.1 System Logs

**Orchestration Logs (orchd):**
- Control panel operations
- Cluster management
- Job execution
- Error tracking

**Access:**
```bash
journalctl -u orchd
```

**Application Daemon Logs (appcd):**
- Website operations
- PHP processing
- Web server events
- Backup operations

**Access:**
```bash
journalctl -u appcd
```

### 18.2 Service-Specific Logs

**Email Logs:**
- Postfix logs
- Dovecot logs
- Rspamd logs
- Delivery tracking

**DNS Logs:**
- PowerDNS query logs
- Zone updates
- DNSSEC operations

**Database Logs:**
- MySQL/MariaDB error log
- Query logs
- Slow query log

**Backup Logs:**
- Backup job status
- Success/Failure tracking
- Duration statistics
- Error details

**Web Server Logs:**
- Access logs (per website)
- Error logs (per website)
- ModSecurity logs

### 18.3 Monitoring Dashboards

**Control Panel Metrics:**
- Server status
- Resource usage
- Active websites
- Backup status
- Update availability

**Server Metrics:**
- CPU usage
- Memory usage
- Disk usage
- Network traffic
- Load average

**Website Metrics:**
- Disk space
- Bandwidth usage
- Database size
- Email storage
- Backup size

### 18.4 Notifications

**Email Notifications:**
- Update availability
- Backup completion/failure
- Server offline alerts
- Disk space warnings
- Resource limit exceeded

**Notification Channels:**
- Email
- Webhook (API integration)
- In-panel notifications

---

## 19. Migration Tools

### 19.1 cPanel Importer

**Functionality:**
- Import cPanel accounts
- Preserve:
  - Websites
  - Databases
  - Email accounts
  - Passwords
  - DNS zones
  - Cron jobs

**Migration Process:**
1. Prepare cPanel backup
2. Upload to Enhance
3. Select import option
4. Map to customer/package
5. Execute import
6. Verify data

**Supported Backup Types:**
- Full cPanel backups
- Account-level backups
- WHM backup format

### 19.2 Plesk Backup Importer

**Functionality:**
- Import Plesk backups
- Account migration
- Website preservation
- Database import
- Email migration

**Process:**
1. Generate Plesk backup
2. Upload to Enhance
3. Import wizard
4. Customer assignment
5. Verification

### 19.3 Control Panel Migration

**Purpose:**
- Move control panel role to new hardware
- Upgrade infrastructure
- Disaster recovery
- Performance improvement

**Requirements:**
- Enhance v12 or later
- Clean Ubuntu 24.04 server
- No existing cluster membership
- Latest Enhance version

**Prerequisites:**
1. Update all cluster servers
2. Migrate customer websites to other servers
3. Remove backup role from control panel server
4. Delete phpMyAdmin and webmail service websites
5. Uninstall DNS, Database, Email roles
6. Verify backups are intact

**Migration Process:**

**1. Setup New Server:**
```bash
curl https://install.enhance.com/install.sh | bash
```

**2. Stop Services on New Server:**
```bash
systemctl stop orchd
systemctl stop appcd
```

**3. Stop Services on Old Server:**
```bash
systemctl stop orchd
systemctl stop appcd
systemctl disable orchd
systemctl disable appcd
```

**Critical Warning:** Never restart these services on old server

**4. Database Backup on Old Server:**
```bash
sudo -u orchd pg_dump -O -d orchd > /var/orchd/orchd.sql
sudo -u orchd pg_dump -O -d authd > /var/orchd/authd.sql
```

**5. Transfer Data to New Server:**
```bash
scp /var/orchd/orchd.sql root@[new_server]:/var/orchd/
scp /var/orchd/authd.sql root@[new_server]:/var/orchd/
scp -r /etc/ssl/certs/enhance root@[new_server]:/etc/ssl/certs/
scp -r /etc/ssl/private/enhance root@[new_server]:/etc/ssl/private/
scp -r /var/local/enhance/orchd/private root@[new_server]:/var/local/enhance/orchd/
scp /var/local/enhance/rca.pw root@[new_server]:/var/local/enhance/
scp -r /var/www/control-panel/assets root@[new_server]:/var/www/control-panel/
```

**6. Import on New Server:**
```bash
sudo -u postgres psql -c "DROP DATABASE orchd;"
sudo -u postgres psql -c "DROP DATABASE authd;"
sudo -u postgres psql -c "CREATE DATABASE orchd;"
sudo -u postgres psql -c "CREATE DATABASE authd;"
sudo -u orchd psql -d orchd < /var/orchd/orchd.sql
sudo -u orchd psql -d authd < /var/orchd/authd.sql
```

**7. Fix Permissions:**
```bash
chown orchd:orchd /var/local/enhance/orchd/private/orchd.key
chown orchd:orchd /etc/ssl/certs/enhance/orchd.crt
chown -R orchd:orchd /var/www/control-panel/screenshots
chown -R orchd:orchd /var/www/control-panel/assets
```

**8. Reinstate Control Panel Websites:**
- Run database query to regenerate website configurations
- Update primary IP addresses
- Restart services

**9. Cleanup:**
- Decommission old server
- Update DNS records
- Verify all services operational

**Troubleshooting:**
- Contact support for issues
- Document guide for Enhance v12+

---

## 20. API & Automation

### 20.1 API Architecture

**RESTful API:**
- JSON request/response
- HTTPS only
- Token authentication
- Comprehensive endpoints

**API Documentation:**
- Available in control panel
- Swagger/OpenAPI format
- Interactive testing
- Code examples

### 20.2 Authentication

**Access Tokens:**
- Created in control panel
- SuperAdmin level access
- API key management
- Secure storage required

**Token Types:**
- **SuperAdmin:** Full cluster access
- **Organization:** Limited to organization
- **User:** Limited user access

**Security:**
- HTTPS required
- Token rotation
- IP whitelisting option
- Rate limiting

### 20.3 API Endpoints

**Organization Management:**
- Create organization
- List organizations
- Get organization details
- Update organization
- Delete organization

**Customer Management:**
- Create customer
- List customers
- Get customer details
- Update customer
- Delete customer
- Impersonate customer

**Package Management:**
- Create package
- List packages
- Get package details
- Update package
- Delete package

**Website Management:**
- Create website
- List websites
- Get website details
- Update website
- Delete website
- Clone website

**Subscription Management:**
- Create subscription
- List subscriptions
- Get subscription details
- Update subscription
- Cancel subscription

**Server Management:**
- List servers
- Get server details
- Add server role
- Remove server role
- Restart role

**Backup Operations:**
- Create backup
- List backups
- Restore backup
- Download backup

**Database Operations:**
- Create database
- List databases
- Get database details
- Delete database

**Domain Operations:**
- Add domain
- List domains
- Update domain
- Delete domain

**Email Operations:**
- Create email account
- List email accounts
- Update email account
- Delete email account

### 20.4 CORS Headers

**Custom CORS Configuration:**
- Custom CORS header for API
- Cross-origin resource sharing
- Browser-based API access
- Security configuration

**Configuration:**
- Advanced → Custom CORS header for API
- Domain whitelist
- Method restrictions
- Header configuration

### 20.5 Webhooks

**Event Notifications:**
- Website created
- Website deleted
- Backup completed
- Server offline
- Update available

**Webhook Configuration:**
- URL endpoint
- Authentication
- Payload format
- Retry logic

---

## 21. Performance Optimization

### 21.1 Cluster Configuration Best Practices

**Small Personal Cluster:**
- Single server: Control Panel + Application + Database
- S3 backups (external)
- Nginx webserver (no .htaccess needed)
- Memory: 100-200MB per website

**Small Hosting Provider:**
- Control Panel: Dedicated server
- Application + Database: Combined servers
- Separate datacenter application servers
- Apache or LiteSpeed (full .htaccess)
- MariaDB for query cache
- 2+ DNS servers
- Dedicated backup server (separate datacenter)
- Email on standalone server

**Scaling Considerations:**
- Choose providers allowing server resizing
- Easy website migration between servers
- No per-server license cost
- Distribute to prevent "bad neighbor" effect

### 21.2 PHP Optimization

**OPcache Configuration:**
- Enable opcode caching
- Reduce script compilation
- Memory allocation
- Revalidation frequency

**PHP-FPM Tuning:**
- Process manager settings
- Worker processes
- Request limits
- Memory per process

### 21.3 Database Optimization

**MariaDB Query Cache:**
- Enable query caching
- Cache size allocation
- Improved read performance
- Repeated query optimization

**my.cnf Tuning:**
- InnoDB buffer pool size
- Max connections
- Query cache size
- Table cache

### 21.4 Web Server Optimization

**LiteSpeed Benefits:**
- Fastest option available
- LSCache integration
- ~10ms TTFB cache hits
- Drop-in Apache replacement
- ModSecurity support

**Nginx Benefits:**
- High performance
- Low memory footprint
- Good for static content
- Reverse proxy capability

**Apache:**
- Mature and stable
- .htaccess support
- Module ecosystem
- Familiar configuration

### 21.5 Resource Limiting

**CPU Limits:**
- Per-website CPU percentage
- Fair resource distribution
- Prevent CPU monopolization

**I/O Limits:**
- Bandwidth (MB/s)
- IOPS limits
- Consistent performance

**Memory Limits:**
- Per-website RAM allocation
- OOM protection
- Fair memory distribution

**Process Limits:**
- nproc limits
- Prevent fork bombs
- System stability

### 21.6 Caching Strategies

**OPcache:**
- Compiled PHP bytecode caching
- Reduces CPU usage
- Faster script execution

**Redis:**
- Object caching
- Session storage
- Database query cache
- WordPress optimization

**LiteSpeed Cache:**
- Page caching
- Browser cache
- Object cache
- Image optimization

### 21.7 Server Hardware Recommendations

**Storage:**
- NVMe SSD strongly recommended
- High IOPS for databases
- Adequate space for growth

**CPU:**
- Modern CPU architecture
- High clock speed
- Multiple cores for parallel processing

**Memory:**
- 100-200MB per website guideline
- Overhead for services
- Buffer for peak loads

**Network:**
- Low latency connections
- Adequate bandwidth
- DDoS protection

---

## 22. Technical Requirements

### 22.1 System Requirements

**Operating System:**
- **Required:** Ubuntu 24.04 LTS Server
- **Not Supported:**
  - Other Linux distributions
  - Ubuntu Desktop
  - Older Ubuntu versions
  - ARM architecture
  - Containers (Docker, LXC, etc.)
  - Paravirtualized environments

**Hardware Minimum:**
- **RAM:** 2GB
- **Storage:** 20GB
- **CPU:** 2 cores (physical or virtual)
- **Architecture:** x86_64/amd64

**Hardware Recommendations:**
- **Control Panel:** 4GB+ RAM, 50GB+ storage
- **Application Server:** 4GB+ RAM per 20-40 websites
- **Database Server:** 4GB+ RAM, high IOPS storage
- **Email Server:** 2GB+ RAM
- **DNS Server:** 1GB+ RAM
- **Backup Server:** Large storage volume

**Storage Considerations:**
- Most data in `/var` directory
- Separate `/var` partition recommended
- High IOPS for databases
- NVMe SSD for best performance

### 22.2 Network Requirements

**Bandwidth:**
- Adequate for hosted content
- Consider peak traffic
- DDoS mitigation

**Ports (Inbound):**
- **2087:** Control Panel HTTPS
- **80/443:** HTTP/HTTPS (Application)
- **21:** FTP (Application)
- **22:** SSH (All servers)
- **25:** SMTP (Email)
- **143/993:** IMAP (Email)
- **110/995:** POP3 (Email)
- **587/465:** Submission (Email)
- **53:** DNS queries (DNS servers)

**Ports (Internal - between cluster members):**
- **50000:** Primary RPC
- **50001-50003:** Additional services
- **50004:** File management
- **22:** SSH for management

**Firewall Requirements:**
- Allow internal cluster communication
- External firewall: Manual port configuration
- UFW: Automatic internal management

### 22.3 Software Dependencies

**Automatic Installation:**
- Docker Engine
- Docker Compose
- Required system packages
- Enhance services
- Database engines
- Web servers
- Mail servers

**Package Manager:**
- APT (Advanced Package Tool)
- Automatic updates via `apt`
- System-level management

### 22.4 Domain Requirements

**Control Panel:**
- Valid domain name recommended
- DNS A record to control panel IP
- SSL certificate (Let's Encrypt auto)

**Hosting Domains:**
- DNS management (Enhance or external)
- Proper A/AAAA records
- MX records for email
- SPF/DKIM for email authentication

**Service Domains:**
- Staging domain (if using staging)
- phpMyAdmin domain
- Webmail domain
- Preview domain (automatic)

### 22.5 License Requirements

**License Types:**
- Single installation license
- Unlimited servers per installation
- No per-server fees
- Valid license for updates

**License Activation:**
- Required for operation
- Entered in Settings → License
- Verification with license server
- Trial options available

### 22.6 Update System

**Core Updates:**
- Platform-wide updates
- API and UI updates
- Manual trigger required
- Release notes provided
- Notifications sent

**Appcd Updates:**
- Per-server updates
- PHP container updates
- Security patches
- Performance improvements
- Manual trigger per server

**Update Process:**
1. Notification received
2. Review release notes
3. Trigger update (off-peak recommended)
4. Orchestration layer updates
5. Service restarts (brief downtime)
6. Verification

**APT Updates:**
- System package updates
- `apt update && apt upgrade`
- Kernel updates
- Security patches

### 22.7 Backup Requirements

**Backup Server:**
- Adequate storage (Formula: websites × 30 × avg DB size)
- Separate datacenter recommended
- High-speed connection to cluster
- Reliable storage (RAID recommended)

**S3 Backup (Optional):**
- S3-compatible storage account
- API credentials
- Bucket configuration
- Retention policy

---

## Appendix A: Quick Reference Commands

### Common CLI Commands

**SSO Access:**
```bash
ecp sso
```

**Update System:**
```bash
apt update && apt upgrade
```

**View Orchestration Logs:**
```bash
journalctl -u orchd
```

**View Application Logs:**
```bash
journalctl -u appcd
```

**Check Port 25:**
```bash
telnet smtp.gmail.com 25
```

**Service Management:**
```bash
systemctl status orchd
systemctl restart orchd
systemctl status appcd
systemctl restart appcd
```

---

## Appendix B: Port Reference Table

| Port | Protocol | Service | Servers |
|------|----------|---------|---------|
| 2087 | TCP | Control Panel HTTPS | Control Panel |
| 80 | TCP | HTTP | Application |
| 443 | TCP | HTTPS | Application |
| 21 | TCP | FTP | Application |
| 22 | TCP | SSH | All |
| 25 | TCP | SMTP | Email |
| 143 | TCP | IMAP | Email |
| 993 | TCP | IMAPS | Email |
| 110 | TCP | POP3 | Email |
| 995 | TCP | POP3S | Email |
| 587 | TCP | Submission | Email |
| 465 | TCP | SMTPS | Email |
| 53 | TCP/UDP | DNS | DNS |
| 50000 | TCP | Internal RPC | All (cluster) |
| 50001-50003 | TCP | Internal Services | All (cluster) |
| 50004 | TCP | File Management | All (cluster) |

---

## Appendix C: Webserver Comparison Matrix

| Feature | Apache | Nginx | LiteSpeed | OpenLiteSpeed |
|---------|--------|-------|-----------|---------------|
| .htaccess Support | ✓ | ✗ | ✓ | ✗ |
| ModSecurity | ✓ | ✓ | ✓ | ✗ |
| Performance | Good | Excellent | Excellent | Excellent |
| LSCache | ✗ | ✗ | ✓ | ✓ |
| WordPress Optimized | Good | Good | Excellent | Excellent |
| License | Free | Free | Commercial | Free |
| Drop-in Replacement | - | - | Apache | - |
| Recommended For | Legacy/Compatibility | Modern/Personal | Professional Hosting | Cost-conscious |

---

## Appendix D: Database Engine Comparison

| Feature | MySQL 8.0 | MariaDB 10.11 | PostgreSQL 16 |
|---------|-----------|---------------|---------------|
| Query Cache | ✗ | ✓ | ✗ |
| Authentication | caching_sha2_password | mysql_native_password | Multiple |
| WordPress Compatible | ✓ | ✓ | Limited |
| Performance | Excellent | Excellent | Excellent |
| ACID Compliance | ✓ | ✓ | ✓ |
| Recommended For | General Use | Hosting Providers | Advanced Applications |

---

## Document Change History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | November 2, 2025 | Initial comprehensive documentation |

---

## Conclusion

This comprehensive PRD document provides a complete analysis of the Enhance Panel hosting control panel system. The documentation covers all major features, functions, and functionalities discovered through exhaustive research of the official documentation.

**Key Takeaways:**

1. **Modern Architecture:** Containerized, scalable, multi-server cluster design
2. **Flexible Deployment:** All roles on one server or distributed across many
3. **Cost-Effective:** Single license, unlimited servers
4. **Security-First:** Containerization, automatic SSL, ModSecurity, 2FA
5. **Performance-Oriented:** Resource limits, multiple webserver options, optimization tools
6. **Feature-Rich:** WordPress toolkit, backups, email, DNS, databases
7. **Developer-Friendly:** RESTful API, CLI tools, automation support
8. **White-Label Ready:** Reseller capabilities with full branding
9. **Migration-Friendly:** cPanel and Plesk import tools

This document serves as a foundation for creating a Product Requirements Document for a new hosting control panel project or for understanding the complete feature set of the Enhance Panel system.

---

**Document Prepared By:** AI Research Assistant  
**Source Data:** https://enhance.com/docs/  
**Research Date:** November 2, 2025  
**Total Pages:** This comprehensive document  
**Format:** Markdown for easy conversion to other formats
