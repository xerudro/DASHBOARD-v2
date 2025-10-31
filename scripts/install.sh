#!/bin/bash

# VIP Hosting Panel v2 - Automated Installation Script
# This script installs all dependencies and sets up the panel

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/vip-panel"
CONFIG_DIR="/etc/vip-panel"
LOG_DIR="/var/log/vip-panel"
DATA_DIR="/var/lib/vip-panel"

echo -e "${BLUE}"
echo "═══════════════════════════════════════════════════════"
echo "   VIP Hosting Panel v2 - Installation Script"
echo "═══════════════════════════════════════════════════════"
echo -e "${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VER=$VERSION_ID
else
    echo -e "${RED}Cannot detect OS version${NC}"
    exit 1
fi

echo -e "${GREEN}Detected OS: $OS $VER${NC}"

# Check if supported OS
if [[ "$OS" != "ubuntu" && "$OS" != "debian" ]]; then
    echo -e "${RED}This script only supports Ubuntu and Debian${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}This script will install:${NC}"
echo "  - PostgreSQL 15"
echo "  - Redis 7"
echo "  - Nginx"
echo "  - Go 1.21"
echo "  - Node.js 18"
echo "  - Python 3 & Ansible"
echo "  - VIP Hosting Panel"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

echo ""
echo -e "${BLUE}Step 1: Updating system...${NC}"
apt update
apt upgrade -y

echo ""
echo -e "${BLUE}Step 2: Installing system dependencies...${NC}"
apt install -y \
    curl \
    wget \
    git \
    build-essential \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release

echo ""
echo -e "${BLUE}Step 3: Installing PostgreSQL...${NC}"
if ! command -v psql &> /dev/null; then
    # Add PostgreSQL repository
    curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list
    
    apt update
    apt install -y postgresql-15 postgresql-contrib-15
    
    # Start and enable PostgreSQL
    systemctl enable postgresql
    systemctl start postgresql
    
    echo -e "${GREEN}✓ PostgreSQL installed${NC}"
else
    echo -e "${GREEN}✓ PostgreSQL already installed${NC}"
fi

echo ""
echo -e "${BLUE}Step 4: Installing Redis...${NC}"
if ! command -v redis-server &> /dev/null; then
    apt install -y redis-server
    
    # Configure Redis
    sed -i 's/supervised no/supervised systemd/' /etc/redis/redis.conf
    
    # Start and enable Redis
    systemctl enable redis-server
    systemctl restart redis-server
    
    echo -e "${GREEN}✓ Redis installed${NC}"
else
    echo -e "${GREEN}✓ Redis already installed${NC}"
fi

echo ""
echo -e "${BLUE}Step 5: Installing Nginx...${NC}"
if ! command -v nginx &> /dev/null; then
    apt install -y nginx
    
    # Start and enable Nginx
    systemctl enable nginx
    systemctl start nginx
    
    echo -e "${GREEN}✓ Nginx installed${NC}"
else
    echo -e "${GREEN}✓ Nginx already installed${NC}"
fi

echo ""
echo -e "${BLUE}Step 6: Installing Go...${NC}"
if ! command -v go &> /dev/null; then
    cd /tmp
    wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    
    echo -e "${GREEN}✓ Go installed${NC}"
else
    echo -e "${GREEN}✓ Go already installed${NC}"
fi

echo ""
echo -e "${BLUE}Step 7: Installing Node.js...${NC}"
if ! command -v node &> /dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
    apt install -y nodejs
    
    echo -e "${GREEN}✓ Node.js installed${NC}"
else
    echo -e "${GREEN}✓ Node.js already installed${NC}"
fi

echo ""
echo -e "${BLUE}Step 8: Installing Python and Ansible...${NC}"
apt install -y python3 python3-pip ansible

echo ""
echo -e "${BLUE}Step 9: Creating database...${NC}"
DB_PASSWORD=$(openssl rand -base64 32)

sudo -u postgres psql <<EOF
CREATE DATABASE vip_hosting;
CREATE USER vip_panel WITH PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE vip_hosting TO vip_panel;
\c vip_hosting
GRANT ALL ON SCHEMA public TO vip_panel;
EOF

echo -e "${GREEN}✓ Database created${NC}"
echo -e "${YELLOW}Database password: $DB_PASSWORD${NC}"
echo -e "${YELLOW}Save this password! It will be used in config.yaml${NC}"

echo ""
echo -e "${BLUE}Step 10: Building VIP Panel...${NC}"
cd "$(dirname "$0")/.."
make build

echo ""
echo -e "${BLUE}Step 11: Installing VIP Panel...${NC}"
make install

# Update config with database password
sed -i "s/password: postgres/password: $DB_PASSWORD/" $CONFIG_DIR/config.yaml

echo ""
echo -e "${BLUE}Step 12: Running migrations...${NC}"
$INSTALL_DIR/vip-panel-cli migrate up

echo ""
echo -e "${BLUE}Step 13: Creating admin user...${NC}"
ADMIN_PASSWORD=$(openssl rand -base64 16)
$INSTALL_DIR/vip-panel-cli create-admin \
    --email admin@localhost \
    --password "$ADMIN_PASSWORD" \
    --name "System Administrator"

echo -e "${GREEN}✓ Admin user created${NC}"
echo -e "${YELLOW}Email: admin@localhost${NC}"
echo -e "${YELLOW}Password: $ADMIN_PASSWORD${NC}"
echo -e "${YELLOW}Save this password!${NC}"

echo ""
echo -e "${BLUE}Step 14: Installing systemd services...${NC}"
make install-services

echo ""
echo -e "${BLUE}Step 15: Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    ufw allow 22/tcp   # SSH
    ufw allow 80/tcp   # HTTP
    ufw allow 443/tcp  # HTTPS
    ufw --force enable
    echo -e "${GREEN}✓ Firewall configured${NC}"
fi

echo ""
echo -e "${BLUE}Step 16: Configuring Nginx...${NC}"
make setup-nginx

echo ""
echo -e "${GREEN}"
echo "═══════════════════════════════════════════════════════"
echo "   Installation Complete!"
echo "═══════════════════════════════════════════════════════"
echo -e "${NC}"
echo ""
echo -e "${YELLOW}Important Information:${NC}"
echo ""
echo -e "${BLUE}Database:${NC}"
echo "  Password: $DB_PASSWORD"
echo ""
echo -e "${BLUE}Admin Account:${NC}"
echo "  Email: admin@localhost"
echo "  Password: $ADMIN_PASSWORD"
echo ""
echo -e "${BLUE}Access:${NC}"
echo "  URL: http://$(hostname -I | awk '{print $1}')"
echo "  Or setup domain in: $CONFIG_DIR/config.yaml"
echo ""
echo -e "${BLUE}Configuration:${NC}"
echo "  Main config: $CONFIG_DIR/config.yaml"
echo "  Providers: $CONFIG_DIR/providers.yaml"
echo ""
echo -e "${BLUE}Logs:${NC}"
echo "  View logs: make logs"
echo "  API logs: journalctl -u vip-panel-api -f"
echo "  Worker logs: journalctl -u vip-panel-worker -f"
echo ""
echo -e "${BLUE}Management:${NC}"
echo "  Status: make status"
echo "  Restart: make restart"
echo "  Stop: make stop"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Edit $CONFIG_DIR/config.yaml with your settings"
echo "2. Edit $CONFIG_DIR/providers.yaml with API keys"
echo "3. Setup SSL certificate (Let's Encrypt recommended)"
echo "4. Access the panel and change the admin password"
echo ""
echo -e "${GREEN}Thank you for using VIP Hosting Panel!${NC}"
echo ""

# Save credentials to file
cat > $CONFIG_DIR/credentials.txt <<EOF
VIP Hosting Panel - Installation Credentials
Generated: $(date)

Database:
  Name: vip_hosting
  User: vip_panel
  Password: $DB_PASSWORD

Admin Account:
  Email: admin@localhost
  Password: $ADMIN_PASSWORD

IMPORTANT: Delete this file after saving credentials securely!
EOF

chmod 600 $CONFIG_DIR/credentials.txt
echo -e "${YELLOW}Credentials saved to: $CONFIG_DIR/credentials.txt${NC}"
echo -e "${RED}Delete this file after saving the credentials!${NC}"
