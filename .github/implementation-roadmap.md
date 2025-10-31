# Implementation Roadmap - VIP Hosting Panel v2

## 🎯 Ce Am Creat Până Acum

### ✅ Infrastructure & Setup (COMPLET)
- [x] Structură completă de directoare
- [x] Makefile cu toate comenzile
- [x] systemd service files
- [x] Script automat de instalare
- [x] Nginx setup automation
- [x] go.mod cu toate dependencies
- [x] Tailwind CSS configuration
- [x] package.json pentru frontend

### ✅ Frontend Design (COMPLET)
- [x] Design system identic cu screenshot-uri
- [x] Base layout cu sidebar
- [x] Dashboard page template
- [x] Custom CSS cu toate componentele
- [x] Dark/Light theme support
- [x] Responsive design

### ✅ Configuration (COMPLET)
- [x] config.yaml.example (comprehensive)
- [x] providers.yaml.example
- [x] Toate path-urile și setările

---

## 📝 Ce Trebuie Implementat

### 1. Core Go Application (Prioritate: CRITICAL)

#### A. Main Entry Points

**`cmd/api/main.go`** - Web Server
```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/vip-hosting/panel-v2/internal/config"
    "github.com/vip-hosting/panel-v2/internal/database"
    "github.com/vip-hosting/panel-v2/internal/handlers"
    "github.com/vip-hosting/panel-v2/internal/middleware"
)

func main() {
    // Load config
    cfg := config.Load()
    
    // Connect to database
    db := database.Connect(cfg)
    
    // Connect to Redis
    redis := database.ConnectRedis(cfg)
    
    // Initialize Fiber app
    app := fiber.New(fiber.Config{
        ServerHeader: "VIP-Panel",
        AppName:      "VIP Hosting Panel v2",
    })
    
    // Middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recover())
    app.Use(middleware.CORS())
    
    // Static files
    app.Static("/static", "./web/static")
    
    // Initialize handlers
    handlers.Setup(app, db, redis)
    
    // Start server
    log.Fatal(app.Listen(cfg.Server.Host + ":" + cfg.Server.Port))
}
```

**`cmd/worker/main.go`** - Background Worker
```go
package main

import (
    "log"
    "github.com/hibiken/asynq"
    "github.com/vip-hosting/panel-v2/internal/config"
    "github.com/vip-hosting/panel-v2/internal/jobs"
)

func main() {
    cfg := config.Load()
    
    // Create Asynq server
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: cfg.Redis.Host + ":" + cfg.Redis.Port},
        asynq.Config{
            Concurrency: cfg.Worker.Concurrency,
            Queues: map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            },
        },
    )
    
    // Register job handlers
    mux := asynq.NewServeMux()
    jobs.Register(mux)
    
    // Start worker
    if err := srv.Run(mux); err != nil {
        log.Fatalf("could not run server: %v", err)
    }
}
```

**`cmd/cli/main.go`** - CLI Tool
```go
package main

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
    "github.com/vip-hosting/panel-v2/internal/cli"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "vip-panel-cli",
        Short: "VIP Hosting Panel CLI",
    }
    
    // Add commands
    rootCmd.AddCommand(cli.MigrateCmd())
    rootCmd.AddCommand(cli.SeedCmd())
    rootCmd.AddCommand(cli.CreateAdminCmd())
    rootCmd.AddCommand(cli.VersionCmd())
    
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

#### B. Database Models (`internal/models/`)

**`user.go`**
```go
package models

import "time"

type User struct {
    ID           int64     `db:"id" json:"id"`
    Email        string    `db:"email" json:"email"`
    Password     string    `db:"password" json:"-"`
    Name         string    `db:"name" json:"name"`
    Role         string    `db:"role" json:"role"` // superadmin, admin, reseller, client
    TwoFactorKey string    `db:"two_factor_key" json:"-"`
    Active       bool      `db:"active" json:"active"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Session struct {
    ID        string    `db:"id" json:"id"`
    UserID    int64     `db:"user_id" json:"user_id"`
    Token     string    `db:"token" json:"token"`
    ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}
```

**`server.go`**
```go
package models

type Server struct {
    ID          int64   `db:"id" json:"id"`
    UserID      int64   `db:"user_id" json:"user_id"`
    Name        string  `db:"name" json:"name"`
    Provider    string  `db:"provider" json:"provider"` // hetzner, digitalocean, etc
    ProviderID  string  `db:"provider_id" json:"provider_id"`
    IPAddress   string  `db:"ip_address" json:"ip_address"`
    Region      string  `db:"region" json:"region"`
    Size        string  `db:"size" json:"size"`
    Status      string  `db:"status" json:"status"` // provisioning, ready, failed, etc
    CPUUsage    float64 `db:"cpu_usage" json:"cpu_usage"`
    RAMUsage    float64 `db:"ram_usage" json:"ram_usage"`
    DiskUsage   float64 `db:"disk_usage" json:"disk_usage"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
```

**`site.go`**
```go
package models

type Site struct {
    ID           int64     `db:"id" json:"id"`
    ServerID     int64     `db:"server_id" json:"server_id"`
    UserID       int64     `db:"user_id" json:"user_id"`
    Domain       string    `db:"domain" json:"domain"`
    PHPVersion   string    `db:"php_version" json:"php_version"`
    NodeVersion  string    `db:"node_version" json:"node_version"`
    Webserver    string    `db:"webserver" json:"webserver"` // nginx, apache
    RootPath     string    `db:"root_path" json:"root_path"`
    Status       string    `db:"status" json:"status"`
    SSLEnabled   bool      `db:"ssl_enabled" json:"ssl_enabled"`
    SSLProvider  string    `db:"ssl_provider" json:"ssl_provider"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
```

#### C. Database Migrations (`migrations/`)

**`001_initial_schema.sql`**
```sql
-- Users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'client',
    two_factor_key VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Servers table
CREATE TABLE servers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255),
    ip_address INET,
    region VARCHAR(100),
    size VARCHAR(100),
    status VARCHAR(50) DEFAULT 'provisioning',
    cpu_usage DECIMAL(5,2) DEFAULT 0,
    ram_usage DECIMAL(5,2) DEFAULT 0,
    disk_usage DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sites table
CREATE TABLE sites (
    id BIGSERIAL PRIMARY KEY,
    server_id BIGINT REFERENCES servers(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    domain VARCHAR(255) UNIQUE NOT NULL,
    php_version VARCHAR(20),
    node_version VARCHAR(20),
    webserver VARCHAR(50) DEFAULT 'nginx',
    root_path VARCHAR(500),
    status VARCHAR(50) DEFAULT 'active',
    ssl_enabled BOOLEAN DEFAULT false,
    ssl_provider VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_servers_user_id ON servers(user_id);
CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_sites_server_id ON sites(server_id);
CREATE INDEX idx_sites_user_id ON sites(user_id);
CREATE INDEX idx_sites_domain ON sites(domain);
```

### 2. HTMX Pages (Prioritate: HIGH)

Trebuie create template-uri pentru:

- [ ] **Servers Page** (`web/templates/pages/servers/list.templ`)
- [ ] **Server Detail** (`web/templates/pages/servers/detail.templ`)
- [ ] **Create Server Modal** (`web/templates/pages/servers/create.templ`)
- [ ] **Automation Page** (`web/templates/pages/automation.templ`)
- [ ] **Analytics Page** (`web/templates/pages/analytics.templ`)
- [ ] **Security Page** (`web/templates/pages/security.templ`)
- [ ] **Email Page** (`web/templates/pages/email.templ`)
- [ ] **DNS Page** (`web/templates/pages/dns.templ`)

### 3. Ansible Playbooks (Prioritate: HIGH)

Playbook-urile principale:

#### **`automation/playbooks/provision-server.yml`**
```yaml
---
- name: Provision new server
  hosts: all
  become: yes
  
  tasks:
    - name: Update system
      apt:
        update_cache: yes
        upgrade: dist
        
    - name: Install base packages
      apt:
        name:
          - curl
          - wget
          - git
          - ufw
          - fail2ban
        state: present
        
    - name: Configure firewall
      ufw:
        rule: allow
        port: "{{ item }}"
      loop:
        - "22"
        - "80"
        - "443"
```

#### **`automation/playbooks/install-php.yml`**
```yaml
---
- name: Install multiple PHP versions
  hosts: all
  become: yes
  
  vars:
    php_versions:
      - "7.4"
      - "8.0"
      - "8.1"
      - "8.2"
      - "8.3"
  
  tasks:
    - name: Add PHP repository
      apt_repository:
        repo: ppa:ondrej/php
        
    - name: Install PHP versions
      apt:
        name: "php{{ item }}-fpm"
        state: present
      loop: "{{ php_versions }}"
```

### 4. Python Scripts (Prioritate: MEDIUM)

#### **`automation/scripts/server_manager.py`**
```python
#!/usr/bin/env python3
"""
Server management script
Called by Go application via subprocess
"""

import sys
import json
import subprocess

def provision_server(server_data):
    """Run Ansible playbook to provision server"""
    cmd = [
        'ansible-playbook',
        'automation/playbooks/provision-server.yml',
        '-i', f"{server_data['ip_address']},",
        '--private-key', server_data['ssh_key'],
        '-e', f"server_id={server_data['id']}"
    ]
    
    result = subprocess.run(cmd, capture_output=True)
    return result.returncode == 0

if __name__ == "__main__":
    action = sys.argv[1]
    data = json.loads(sys.argv[2])
    
    if action == "provision":
        success = provision_server(data)
        print(json.dumps({"success": success}))
```

---

## 🎯 Priority Order

### Sprint 1 (Week 1-2): Foundation
1. ✅ Implement `cmd/api/main.go`
2. ✅ Create database models
3. ✅ Write migrations
4. ✅ Implement authentication handlers
5. ✅ Create dashboard handler

### Sprint 2 (Week 3-4): Core Features
1. ⏳ Server management (create, list, delete)
2. ⏳ Hetzner API integration
3. ⏳ Basic Ansible playbooks
4. ⏳ Server provisioning flow
5. ⏳ Real-time status updates (SSE)

### Sprint 3 (Week 5-6): Advanced Features
1. ⏳ Site deployment
2. ⏳ DNS management
3. ⏳ SSL automation
4. ⏳ Email configuration
5. ⏳ Backup system

### Sprint 4 (Week 7-8): Polish & Launch
1. ⏳ Security features (WAF, Fail2ban)
2. ⏳ Monitoring & alerts
3. ⏳ Billing integration
4. ⏳ Documentation
5. ⏳ Beta launch

---

## 📚 Learning Resources

### Go
- Official Tour: https://go.dev/tour/
- Go by Example: https://gobyexample.com/
- Fiber docs: https://docs.gofiber.io/

### HTMX
- Official docs: https://htmx.org/docs/
- Examples: https://htmx.org/examples/
- Essays: https://htmx.org/essays/

### Templ
- Documentation: https://templ.guide/
- GitHub: https://github.com/a-h/templ

### Ansible
- Getting Started: https://docs.ansible.com/ansible/latest/getting_started/
- Playbook Guide: https://docs.ansible.com/ansible/latest/playbook_guide/

---

## 🐛 Common Issues & Solutions

### Issue: Templ generate fails
```bash
# Solution: Install templ CLI
go install github.com/a-h/templ/cmd/templ@latest
```

### Issue: CSS not building
```bash
# Solution: Install node modules
npm install
npm run build:css
```

### Issue: Database connection refused
```bash
# Solution: Check PostgreSQL is running
sudo systemctl status postgresql
sudo systemctl start postgresql
```

---

## ✅ Definition of Done

Un feature este "Done" când:

1. ✅ Codul este scris și testat
2. ✅ Template-urile HTMX sunt create
3. ✅ Există migrări de database (dacă aplicabil)
4. ✅ Ansible playbook-urile funcționează
5. ✅ Documentația este actualizată
6. ✅ Nu există bug-uri critice
7. ✅ UI matches design screenshots
8. ✅ Funcționează cu systemd în production

---

## 🚀 Ready to Code!

Ai acum:
- ✅ Structura completă
- ✅ Design system implementat
- ✅ Infrastructure code
- ✅ Clear roadmap
- ✅ Priority order

**Next step: Implementează `cmd/api/main.go` și pornește development!**

Mult succes! 🎉
