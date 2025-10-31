# VIP Hosting Panel v2 - Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Step 1: Download & Extract
```bash
# Extract the archive
tar -xzf vip-hosting-panel-v2.tar.gz
cd vip-hosting-panel-v2
```

### Step 2: Install (Choose One)

#### Option A: Automated (Recommended for Production)
```bash
# Run the automated installer
sudo bash scripts/install.sh

# This will:
# ✓ Install PostgreSQL, Redis, Nginx
# ✓ Build the application
# ✓ Setup systemd services
# ✓ Create database
# ✓ Generate SSL certificate
# ✓ Start all services
```

#### Option B: Development Setup
```bash
# Install dependencies
sudo apt install postgresql redis-server nginx golang nodejs npm

# Setup database
sudo -u postgres createdb vip_hosting
sudo -u postgres createuser vip_panel

# Install packages
go mod download
npm install

# Build CSS
npm run build:css

# Build app
make build

# Run migrations
make migrate

# Start dev server
make dev
```

### Step 3: Access the Panel
```
URL: http://your-server-ip
  or https://your-domain.com (if SSL configured)

Default Login:
Email: admin@localhost
Password: (check /etc/vip-panel/credentials.txt)
```

### Step 4: Configure
```bash
# Edit main config
sudo nano /etc/vip-panel/config.yaml

# Edit provider API keys
sudo nano /etc/vip-panel/providers.yaml

# Restart services
sudo systemctl restart vip-panel-api
sudo systemctl restart vip-panel-worker
```

---

## 🎨 Design Features

✅ **Dark Theme by Default** (navy blue background)
✅ **Orange Accent Color** (#FF6B35)
✅ **Light Mode Toggle** (top right)
✅ **Sidebar Navigation** (16 menu items)
✅ **Real-time Updates** (via HTMX + SSE)
✅ **Responsive Design** (mobile-friendly)

---

## 📋 Features Included

### Infrastructure
- Multi-PHP (5.6 - 8.3)
- Multi-Node.js (14-21)
- Multi-Database (MySQL, PostgreSQL, MongoDB, Redis)
- Multi-Webserver (Nginx, Apache)

### Security
- WAF (ModSecurity)
- Antivirus (ClamAV)
- Fail2ban
- Firewall Manager
- SSL/TLS (Let's Encrypt)
- 2FA Support

### Email
- Mail Server (Postfix + Dovecot)
- Webmail (Roundcube)
- Email Accounts
- Aliases & Forwarding
- Spam Protection

### Management
- Server Provisioning
- Site Deployment
- DNS Management
- Backup System
- Monitoring & Alerts
- Billing (Stripe)

---

## 🛠️ Common Commands

### Service Management
```bash
# Check status
make status

# View logs
make logs

# Restart services
make restart

# Stop services
make stop

# Start services
make start
```

### Development
```bash
# Build application
make build

# Run tests
make test

# Generate templates
templ generate

# Build CSS
npm run build:css

# Start dev mode
make dev
```

### Database
```bash
# Run migrations
make migrate

# Rollback last migration
make rollback

# Seed test data
make seed
```

### Maintenance
```bash
# Create backup
make backup

# Update to latest version
make update

# Check system health
make health
```

---

## 📁 Project Structure

```
vip-hosting-panel-v2/
├── cmd/                  # Go applications
│   ├── api/             # Web server
│   ├── worker/          # Background jobs
│   └── cli/             # CLI tool
│
├── web/                 # Frontend
│   ├── templates/       # HTMX/Templ templates
│   └── static/          # CSS, JS, images
│
├── automation/          # Ansible + Python
│   ├── playbooks/       # Server automation
│   └── scripts/         # Management scripts
│
├── configs/             # Configuration files
├── migrations/          # Database migrations
└── scripts/             # Build & deploy scripts
```

---

## 🔐 Security Checklist

After installation:

- [ ] Change default admin password
- [ ] Update database password in config
- [ ] Setup firewall (UFW)
- [ ] Install SSL certificate
- [ ] Enable 2FA for admin
- [ ] Review security settings
- [ ] Setup backup schedule
- [ ] Configure monitoring alerts

---

## 📞 Getting Help

### Documentation
- `PROJECT_OVERVIEW.md` - Complete project overview
- `IMPLEMENTATION_ROADMAP.md` - Development roadmap
- `README.md` - Full documentation

### Logs
```bash
# API logs
sudo journalctl -u vip-panel-api -f

# Worker logs
sudo journalctl -u vip-panel-worker -f

# Nginx logs
sudo tail -f /var/log/nginx/vip-panel-error.log
```

### Health Check
```bash
make health
```

---

## 🎯 Next Steps

1. **Configure Providers**
   - Add Hetzner API token
   - Add Cloudflare API token (for DNS)
   - Add Stripe keys (for billing)

2. **Create First Server**
   - Go to Servers → Add Physical Server
   - Enter server details
   - Wait for provisioning

3. **Deploy First Site**
   - Go to Websites → New Website
   - Select server
   - Enter domain
   - Choose PHP version
   - Deploy!

4. **Setup DNS**
   - Go to DNS Settings
   - Add your domain
   - Configure records

5. **Enable SSL**
   - Go to Security → SSL Management
   - Select domain
   - Generate certificate

---

## 💡 Pro Tips

1. **Use systemd for reliability**
   - Services auto-restart on failure
   - Logs in journalctl
   - Easy management with systemctl

2. **Monitor everything**
   - Check dashboard regularly
   - Setup email alerts
   - Review logs weekly

3. **Backup regularly**
   - Automated daily backups
   - Test restore process
   - Store offsite (S3)

4. **Keep updated**
   - Update panel monthly
   - Update server packages weekly
   - Review security advisories

5. **Document changes**
   - Keep config backups
   - Document custom rules
   - Track server changes

---

## 🌟 Features Roadmap

### v1.0 (Current)
- Core infrastructure ✅
- Basic features ✅
- Dark theme UI ✅

### v1.1 (Next)
- Container support (Docker) ⏳
- Advanced WAF rules ⏳
- Mobile app ⏳

### v1.2 (Future)
- White-labeling ⏳
- Plugin system ⏳
- AI optimization ⏳

---

## 🎉 You're Ready!

Your VIP Hosting Panel v2 is ready to use. Start by:

1. Logging in to the panel
2. Changing the admin password
3. Adding your first provider
4. Creating your first server

**Happy hosting!** 🚀