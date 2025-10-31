#!/bin/bash

# VIP Hosting Panel - Nginx Configuration Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Configuring Nginx for VIP Hosting Panel...${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Get server IP
SERVER_IP=$(hostname -I | awk '{print $1}')

echo ""
echo "Enter your domain name (or press Enter to use IP: $SERVER_IP):"
read -r DOMAIN

if [ -z "$DOMAIN" ]; then
    DOMAIN=$SERVER_IP
    SSL_ENABLED=false
else
    echo "Do you want to setup SSL with Let's Encrypt? (y/n)"
    read -r SSL_CHOICE
    if [[ $SSL_CHOICE =~ ^[Yy]$ ]]; then
        SSL_ENABLED=true
    else
        SSL_ENABLED=false
    fi
fi

# Create Nginx configuration
cat > /etc/nginx/sites-available/vip-panel <<EOF
# VIP Hosting Panel - Nginx Configuration

# Upstream to VIP Panel API
upstream vip_panel_api {
    server 127.0.0.1:3000;
    keepalive 32;
}

# Rate limiting zones
limit_req_zone \$binary_remote_addr zone=panel_limit:10m rate=10r/s;
limit_conn_zone \$binary_remote_addr zone=conn_limit:10m;

# HTTP Server (redirect to HTTPS if SSL enabled)
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;

    # Access and error logs
    access_log /var/log/nginx/vip-panel-access.log;
    error_log /var/log/nginx/vip-panel-error.log;

    # Let's Encrypt ACME challenge
    location ^~ /.well-known/acme-challenge/ {
        root /var/www/html;
    }

EOF

if [ "$SSL_ENABLED" = true ]; then
    cat >> /etc/nginx/sites-available/vip-panel <<EOF
    # Redirect HTTP to HTTPS
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

# HTTPS Server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $DOMAIN;

    # SSL Configuration (will be updated by certbot)
    ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/$DOMAIN/chain.pem;

    # SSL Security Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

EOF
else
    cat >> /etc/nginx/sites-available/vip-panel <<EOF
    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

EOF
fi

cat >> /etc/nginx/sites-available/vip-panel <<EOF
    # Access and error logs
    access_log /var/log/nginx/vip-panel-ssl-access.log;
    error_log /var/log/nginx/vip-panel-ssl-error.log;

    # Rate limiting
    limit_req zone=panel_limit burst=20 nodelay;
    limit_conn conn_limit 10;

    # Max upload size
    client_max_body_size 100M;
    client_body_buffer_size 128k;

    # Timeouts
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
    send_timeout 60s;

    # Static files
    location /static/ {
        alias /opt/vip-panel/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # WebSocket support
    location /ws {
        proxy_pass http://vip_panel_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 86400;
    }

    # API and application
    location / {
        proxy_pass http://vip_panel_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
        
        # Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
        
        # Caching
        proxy_cache_bypass \$http_upgrade;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}
EOF

# Enable the site
ln -sf /etc/nginx/sites-available/vip-panel /etc/nginx/sites-enabled/

# Remove default site if exists
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
echo ""
echo -e "${GREEN}Testing Nginx configuration...${NC}"
nginx -t

if [ $? -eq 0 ]; then
    # Reload Nginx
    echo -e "${GREEN}Reloading Nginx...${NC}"
    systemctl reload nginx
    
    echo ""
    echo -e "${GREEN}✓ Nginx configured successfully!${NC}"
    echo ""
    
    if [ "$SSL_ENABLED" = true ]; then
        echo -e "${YELLOW}Installing SSL certificate...${NC}"
        
        # Install certbot if not present
        if ! command -v certbot &> /dev/null; then
            apt install -y certbot python3-certbot-nginx
        fi
        
        # Get SSL certificate
        certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --register-unsafely-without-email
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ SSL certificate installed!${NC}"
            echo ""
            echo "Panel URL: https://$DOMAIN"
        else
            echo -e "${RED}✗ SSL certificate installation failed${NC}"
            echo "Panel URL: http://$DOMAIN"
        fi
    else
        echo "Panel URL: http://$DOMAIN"
    fi
    
    echo ""
    echo -e "${YELLOW}Nginx Logs:${NC}"
    echo "  Access: tail -f /var/log/nginx/vip-panel-access.log"
    echo "  Error: tail -f /var/log/nginx/vip-panel-error.log"
    
else
    echo -e "${RED}✗ Nginx configuration test failed${NC}"
    exit 1
fi
