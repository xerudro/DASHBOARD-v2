# AI-Assisted Development - Quick Reference Card (REFINED v2.0)
## For Claude & GitHub Copilot Usage - systemd, HTMX, Ansible

**Print this and keep at your desk!**

---

## QUICK PROMPTING TEMPLATES (REFINED)

### Template 1: systemd Service Implementation
```
[Paste relevant section from ai-dev-system-prompts-v2-refined.md]

Implement systemd service for [feature]:
- Service name: [name]
- Start command: [command]
- Dependencies: [services]
- Security: [requirements]
- Restart policy: [policy]

Include: Service file, health check script, Ansible deployment
```

### Template 2: HTMX Frontend Component
```
[Paste FRONTEND DEVELOPMENT section]

Build HTMX component for [feature]:
- Server-side template: Tera/Maud
- HTMX attributes: [interactions]
- Styling: TailwindCSS
- Form validation: [server-side]

Include: HTML template, Rust handler, HTMX attributes
```

### Template 3: Ansible Automation
```
[Paste INFRASTRUCTURE section]

Create Ansible playbook for [task]:
- Target: [servers]
- Actions: [list actions]
- Error handling: [strategy]
- Idempotent: [yes/no]

Include: Playbook, handlers, variables, documentation
```

### Template 4: Bash Automation Script
```
[Paste BASH SCRIPT SECURITY section]

Write bash script for [task]:
- Input validation: [required]
- Error handling: [set -euo pipefail]
- Logging: [/var/log/hosting-panel.log]
- Permissions: [owner:group mode]

Include: Script, comments, logging
```

### Template 5: Python Automation
```
[Paste PYTHON AUTOMATION section]

Write Python script for [task]:
- Type hints: [yes]
- Error handling: [comprehensive]
- Logging: [configured]
- Subprocess: [subprocess.run, not os.system]

Include: Script, type hints, error handling, logging
```

---

## TECH STACK CHECKLIST (REFINED)

### Backend
- [ ] Rust 1.75+ with Tokio
- [ ] Actix-web 4.x
- [ ] PostgreSQL 14+
- [ ] Redis 7.x

### Frontend
- [ ] HTMX 1.9.x
- [ ] Tera or Maud templates
- [ ] TailwindCSS 3.x
- [ ] Minimal vanilla JavaScript

### Infrastructure
- [x] **systemd** (NOT Docker/Kubernetes)
- [ ] Ansible 2.13+
- [ ] Bash 4.4+
- [ ] Python 3.10+

### Automation
- [ ] GitHub Actions (CI/CD)
- [ ] Prometheus + Grafana (monitoring)
- [ ] systemd timers (scheduling)
- [ ] Cron jobs (via systemd)

---

## SYSTEMD SERVICE CHECKLIST

### Creating systemd Service

```ini
[Unit]
Description=...
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/binary
User=service-user
Group=service-user
Restart=on-failure
RestartSec=5s

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/service

# Resources
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### Service Management Commands

```bash
# Start/stop/restart
sudo systemctl start service-name
sudo systemctl stop service-name
sudo systemctl restart service-name

# Status and logs
sudo systemctl status service-name
sudo journalctl -u service-name -f

# Enable on boot
sudo systemctl enable service-name

# Timers
sudo systemctl list-timers
sudo systemctl status timer-name.timer
```

---

## HTMX FRONTEND CHECKLIST

### HTMX Basics

```html
<!-- GET request -->
<button hx-get="/api/data" hx-target="#container">Load</button>

<!-- POST request with form data -->
<form hx-post="/api/create" hx-target="#result" hx-swap="outerHTML">
    <input type="text" name="domain" required>
    <button type="submit">Create</button>
</form>

<!-- Real-time validation -->
<input type="email" name="email"
       hx-post="/api/validate/email"
       hx-target="#email-error"
       hx-trigger="change">

<!-- Polling (auto-refresh)-->
<div hx-get="/api/status"
     hx-trigger="every 5s"
     hx-swap="innerHTML">
</div>

<!-- Server-Sent Events -->
<div hx-sse="connect:/api/events"
     hx-trigger="sse:update"
     hx-get="/api/data"
     hx-target="#container">
</div>
```

### Rust Handler Template

```rust
#[get("/api/data")]
async fn get_data(user: AuthUser, db: web::Data<Database>) 
    -> Result<HttpResponse> {
    let data = db.get_user_data(user.id).await?;
    let html = render_template("fragment.html", &data)?;
    Ok(HttpResponse::Ok().content_type("text/html").body(html))
}

#[post("/api/create")]
async fn create_item(
    user: AuthUser,
    body: web::Form<CreateRequest>,
    db: web::Data<Database>
) -> Result<HttpResponse> {
    let item = db.create(user.id, body.into_inner()).await?;
    let html = render_template("item_row.html", &item)?;
    Ok(HttpResponse::Ok().content_type("text/html").body(html))
}
```

---

## ANSIBLE PLAYBOOK CHECKLIST

### Basic Playbook Structure

```yaml
---
- name: Task name
  hosts: target_group
  become: yes  # Use sudo
  
  vars:
    var_name: value
  
  pre_tasks:
    - name: Pre-task
      debug: msg="Running pre-tasks"
  
  tasks:
    - name: Main task
      block:
        - name: Do something
          command: /bin/true
      rescue:
        - name: Handle error
          debug: msg="Error occurred"
      always:
        - name: Cleanup
          debug: msg="Always run"
  
  post_tasks:
    - name: Post-task
      debug: msg="Running post-tasks"
  
  handlers:
    - name: Handler
      systemd:
        name: service
        state: restarted
```

### Common Ansible Modules

```yaml
- name: Install packages
  apt:
    name: [package1, package2]
    state: present

- name: Create directory
  file:
    path: /path/to/dir
    state: directory
    owner: user
    group: group
    mode: '0755'

- name: Template file
  template:
    src: file.j2
    dest: /etc/file
    owner: root
    group: root
    mode: '0644'

- name: Systemd service
  systemd:
    name: service-name
    state: started
    daemon_reload: yes
    enabled: yes

- name: Run command
  command: /path/to/command
  register: result

- name: Copy file
  copy:
    src: local/file
    dest: /remote/file
    owner: user
    group: group
    mode: '0755'
```

---

## BASH SCRIPT CHECKLIST

### Secure Bash Template

```bash
#!/bin/bash
set -euo pipefail  # CRITICAL: Error handling

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LOG_FILE="/var/log/hosting-panel/script.log"

# Logging
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# Error handling
trap 'log "Error on line $LINENO"' ERR
trap 'log "Interrupted"; exit 130' INT TERM

# Input validation
validate_input() {
    local input="$1"
    # Whitelist allowed characters
    [[ "$input" =~ ^[a-zA-Z0-9._-]+$ ]] && return 0 || return 1
}

# Main logic
main() {
    log "Starting script"
    
    # Always use variables quoted
    local var="$1"
    
    # Validate before use
    if ! validate_input "$var"; then
        log "Invalid input: $var"
        return 1
    fi
    
    log "Task completed"
}

main "$@"
```

### Bash Best Practices

```bash
# DO:
set -euo pipefail              # Error handling
"$var"                         # Quote variables
if [[ condition ]]; then       # Use [[ ]]
readonly CONST="value"         # Use readonly
trap handle_error ERR          # Use trap
which command                  # Check existence

# DON'T:
set                            # No -euo
$var                          # Unquoted variables
if [ condition ]; then        # Use [ ] (old)
VAR="value"                   # Mutable globals
eval "command"                # Never with user input
command                       # Don't assume exists
```

---

## PYTHON AUTOMATION CHECKLIST

### Secure Python Template

```python
#!/usr/bin/env python3
"""Script description."""

import os
import sys
import logging
import subprocess
from pathlib import Path
from typing import Optional, List

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('/var/log/hosting-panel/script.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Constants
PANEL_USER = "hosting-panel"
PANEL_DIR = Path("/var/lib/hosting-panel")

def run_command(cmd: List[str], user: Optional[str] = None) -> int:
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
        logger.error(f"Failed: {e.stderr}")
        return e.returncode

def main() -> int:
    """Main function."""
    try:
        logger.info("Starting")
        
        # Your logic here
        
        logger.info("Completed")
        return 0
    except Exception as e:
        logger.error(f"Error: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())
```

### Python Best Practices

```python
# DO:
subprocess.run(cmd, check=True)        # Use subprocess
from pathlib import Path               # Use pathlib
logger.info(message)                  # Use logging
try/except                            # Error handling
type hints                            # Always use hints
with open(file) as f:                 # Use context managers
input().strip()                       # Validate input

# DON'T:
os.system(command)                    # Never use os.system
exec(user_input)                      # Never execute user input
open(file).read()                     # No context manager
sys.exit(1)                          # Return from main instead
no type hints                         # Always use hints
hardcoded paths                       # Use pathlib
```

---

## SECURITY CHECKLIST

### Every systemd Service
- [ ] Run as non-root user
- [ ] NoNewPrivileges=true
- [ ] ProtectSystem=strict
- [ ] ProtectHome=true
- [ ] CapabilityBoundingSet configured
- [ ] RestrictAddressFamilies set
- [ ] MemoryDenyWriteExecute=yes

### Every Ansible Playbook
- [ ] No hardcoded secrets
- [ ] vars_files for credentials (gitignored)
- [ ] Use handlers not inline commands
- [ ] Idempotent operations
- [ ] Proper error handling (block/rescue)
- [ ] Audit logging enabled

### Every Bash Script
- [ ] set -euo pipefail
- [ ] Trap error handler
- [ ] Quote all variables: "$var"
- [ ] Validate all input
- [ ] Check command existence
- [ ] Log all operations
- [ ] Proper exit codes

### Every Python Script
- [ ] Type hints throughout
- [ ] subprocess.run not os.system
- [ ] Logging configured
- [ ] Try/except blocks
- [ ] Input validation
- [ ] Use pathlib not os.path
- [ ] Secure file permissions

---

## DEPLOYMENT FLOW

**1. Local Development**
```bash
cargo build --release
cargo test
cargo clippy
```

**2. CI/CD (GitHub Actions)**
```
- Build Rust binary
- Run all tests
- Security audit
- Upload artifact
```

**3. Deploy (Ansible)**
```bash
ansible-playbook playbooks/deploy.yml
# Ansible will:
# - Stop current service
# - Deploy new binary
# - Run migrations
# - Start service
# - Health check
```

**4. Verify (Monitoring)**
```bash
systemctl status hosting-panel
journalctl -u hosting-panel
curl http://127.0.0.1:8001/health
```

---

## TROUBLESHOOTING

### Service won't start
```bash
journalctl -u hosting-panel -n 50
systemctl status hosting-panel
/usr/local/bin/hosting-panel  # Run manually
```

### Deployment fails
```bash
ansible-playbook playbooks/deploy.yml -vvv  # Verbose
ansible-playbook playbooks/deploy.yml --check  # Dry-run
```

### Script errors
```bash
bash -x script.sh  # Debug mode
set -x            # In script
trap              # Error handler
```

### Ansible issues
```bash
ansible-lint playbooks/deploy.yml
ansible-playbook --syntax-check playbooks/deploy.yml
ansible -i inventory -m ping all
```

---

## COMMON PATTERNS

### systemd Service + Timer
```ini
# /etc/systemd/system/task.service
[Unit]
Description=Run task
[Service]
Type=oneshot
ExecStart=/usr/local/bin/task

# /etc/systemd/system/task.timer
[Unit]
Description=Task timer
[Timer]
OnBootSec=5min
OnUnitActiveSec=1h
[Install]
WantedBy=timers.target
```

### HTMX + Real-time
```html
<div hx-sse="connect:/api/events"
     hx-trigger="sse:update-website"
     hx-get="/api/websites/list"
     hx-target="#websites"
     hx-swap="innerHTML">
</div>
```

### Ansible Multi-Step
```yaml
- name: Deploy with rollback
  block:
    - name: Stop service
      systemd: name=app state=stopped
    - name: Deploy
      copy: src=app dest=/usr/local/bin/app
    - name: Start
      systemd: name=app state=started
  rescue:
    - name: Rollback
      copy: src=app.backup dest=/usr/local/bin/app
      notify: restart app
```

---

## KEYBOARD SHORTCUTS (Copilot)

**VSCode:**
- `Ctrl+I` - Inline chat
- `Ctrl+K` - Open chat panel
- `/explain` - Explain code
- `/fix` - Fix issues
- `/doc` - Add documentation

---

## QUICK LINKS

- Systemd manual: `man systemd.service`
- HTMX docs: https://htmx.org/docs/
- Ansible docs: https://docs.ansible.com/
- Rust book: https://doc.rust-lang.org/book/

---

**Remember:**
✨ Simplicity over complexity  
🔒 Security first  
🧪 Test everything  
📚 Document well  
👥 Review code  
📊 Monitor always  
🚀 Deploy confidently  

---

**Version:** 2.0 (Refined) | **Date:** November 2, 2025 | **Review:** May 2026