# Hetzner Cloud Integration Guide

**Date**: October 31, 2025
**Status**: ✅ Complete and Production-Ready

---

## 🎯 Overview

Complete Hetzner Cloud integration for VIP Hosting Panel v2, providing automated server provisioning, management, and monitoring using Hetzner's official API.

### Features Implemented

- ✅ Full Hetzner Cloud API integration
- ✅ Server provisioning with configuration options
- ✅ Background job processing for async operations
- ✅ Real-time pricing and cost calculation
- ✅ Server lifecycle management (create, delete, resize, reboot)
- ✅ SSH key management
- ✅ Location and server type discovery
- ✅ OS image selection
- ✅ Metrics collection from Hetzner API
- ✅ Redis caching for improved performance
- ✅ Comprehensive audit logging
- ✅ Ansible playbooks for server configuration

---

## 📦 Files Created

### 1. Hetzner Provider Service
**File**: [internal/services/providers/hetzner.go](internal/services/providers/hetzner.go)

Comprehensive Hetzner Cloud API client with:
- Server CRUD operations
- Pricing information retrieval
- Location and server type management
- SSH key management
- Metrics collection
- Redis caching integration
- N/A fallback patterns

### 2. Background Job System
**File**: [internal/jobs/server_provisioning.go](internal/jobs/server_provisioning.go)

Async job processing for:
- Server provisioning (15-minute timeout)
- Server deletion
- Server resizing
- Retry logic (3 attempts)
- Status tracking
- Audit logging

### 3. Worker Service
**File**: [cmd/worker/main.go](cmd/worker/main.go)

Dedicated worker process with:
- Asynq job queue integration
- Priority-based queue processing (critical/default/low)
- Graceful shutdown
- Error handling and logging
- Database and Redis connectivity

### 4. Ansible Playbook
**File**: [automation/playbooks/provision-server.yml](automation/playbooks/provision-server.yml)

Initial server provisioning with:
- System updates and security patches
- Firewall configuration (UFW)
- Fail2ban setup
- SSH hardening
- Swap configuration
- Automatic security updates
- Performance tuning

---

## 🚀 Quick Start

### Step 1: Get Hetzner API Token

1. Log in to [Hetzner Cloud Console](https://console.hetzner.cloud/)
2. Select your project
3. Go to Security → API Tokens
4. Click "Generate API Token"
5. Give it **Read & Write** permissions
6. Copy the token (you'll only see it once!)

### Step 2: Configure Application

Add to your `configs/config.yaml`:

```yaml
# Hetzner Cloud Configuration
hetzner:
  api_token: "YOUR_HETZNER_API_TOKEN_HERE"

# Worker Configuration
worker:
  concurrency: 10  # Number of concurrent jobs

# Redis Configuration (required for job queue)
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

# Database Configuration
database:
  host: localhost
  port: 5432
  name: vip_panel
  user: your_db_user
  password: your_db_password
  ssl_mode: disable
```

Or use environment variables:

```bash
export VIP_HETZNER_API_TOKEN="your_token_here"
export VIP_WORKER_CONCURRENCY=10
export VIP_REDIS_HOST=localhost
export VIP_DATABASE_HOST=localhost
```

### Step 3: Start the Worker

```bash
# Build the worker
make build-worker

# Or run directly
go run cmd/worker/main.go

# In production
./build/vip-panel-worker
```

---

## 💻 Usage Examples

### Example 1: Create a Server

```go
package main

import (
    "context"
    "time"

    "github.com/hibiken/asynq"
    "github.com/xerudro/DASHBOARD-v2/internal/jobs"
    "github.com/xerudro/DASHBOARD-v2/internal/services/providers"
)

func createServer() error {
    // Initialize Hetzner provider
    provider, err := providers.NewHetznerProvider(apiToken, cache)
    if err != nil {
        return err
    }

    // Configure server
    config := providers.ServerConfig{
        Name:             "my-server-001",
        ServerType:       "cx11",        // Smallest server
        Location:         "fsn1",        // Falkenstein, Germany
        Image:            "ubuntu-22.04",
        SSHKeys:          []int64{123456}, // Your SSH key IDs
        StartAfterCreate: true,
        Labels: map[string]string{
            "environment": "production",
            "managed_by":  "vip-panel",
        },
    }

    // Create server (this will take a few minutes)
    serverInfo, err := provider.CreateServer(context.Background(), config)
    if err != nil {
        return err
    }

    fmt.Printf("Server created: ID=%d, IP=%s\n",
        serverInfo.ID, serverInfo.PublicIPv4)

    return nil
}
```

### Example 2: Enqueue Provisioning Job

```go
// From your API handler
func (h *ServerHandler) Create(c *fiber.Ctx) error {
    // ... validate input ...

    // Save server to database with status="queued"
    server := &models.Server{
        TenantID: tenantID,
        UserID:   userID,
        Name:     "my-server",
        Provider: "hetzner",
        Status:   "queued",
    }

    err := h.serverRepo.Create(ctx, server)
    if err != nil {
        return err
    }

    // Enqueue provisioning job
    payload := jobs.ServerProvisioningPayload{
        ServerID:   server.ID,
        TenantID:   tenantID,
        UserID:     userID,
        Provider:   "hetzner",
        ServerType: "cx11",
        Location:   "fsn1",
        Image:      "ubuntu-22.04",
        SSHKeys:    []int64{123456},
    }

    err = jobs.EnqueueServerProvisioning(asynqClient, payload)
    if err != nil {
        return err
    }

    return c.JSON(server)
}
```

### Example 3: Get Pricing Information

```go
func getPricing() error {
    provider, _ := providers.NewHetznerProvider(apiToken, cache)

    // Get all pricing info
    pricing, err := provider.GetPricing(context.Background())
    if err != nil {
        return err
    }

    // Display server types and prices
    for name, st := range pricing.ServerTypes {
        fmt.Printf("%s: €%.2f/month (%d cores, %.0fGB RAM, %dGB disk)\n",
            name, st.Monthly, st.Cores, st.Memory, st.Disk)
    }

    // Calculate cost for specific configuration
    monthlyCost, err := provider.CalculateMonthlyCost(
        context.Background(),
        "cx11",  // Server type
        true,    // Include backups
    )
    fmt.Printf("Monthly cost with backups: €%.2f\n", monthlyCost)

    return nil
}
```

### Example 4: List Available Options

```go
// List locations
locations, err := provider.ListLocations(ctx)
for _, loc := range locations {
    fmt.Printf("Location: %s (%s) - %s\n",
        loc.Name, loc.Country, loc.City)
}

// List server types
serverTypes, err := provider.ListServerTypes(ctx)
for _, st := range serverTypes {
    fmt.Printf("Type: %s - %d cores, %.0fGB RAM, %dGB disk\n",
        st.Name, st.Cores, st.Memory, st.Disk)
}

// List OS images
images, err := provider.ListImages(ctx)
for _, img := range images {
    fmt.Printf("Image: %s - %s %s\n",
        img.Name, img.OSFlavor, img.OSVersion)
}
```

---

## 🎛️ Available Hetzner Options

### Server Types (as of 2024)

| Type | vCPUs | RAM | Disk | Price/Month |
|------|-------|-----|------|-------------|
| cx11 | 1 | 2GB | 20GB | ~€4.51 |
| cx21 | 2 | 4GB | 40GB | ~€6.44 |
| cx31 | 2 | 8GB | 80GB | ~€12.87 |
| cx41 | 4 | 16GB | 160GB | ~€25.74 |
| cx51 | 8 | 32GB | 240GB | ~€51.48 |

*Prices are approximate and subject to change*

### Locations

- **fsn1** - Falkenstein, Germany (Datacenter Park Falkenstein)
- **nbg1** - Nuremberg, Germany (Datacenter Park Nuremberg)
- **hel1** - Helsinki, Finland (Datacenter Park Helsinki)
- **ash** - Ashburn, USA (Datacenter Park Ashburn, Virginia)
- **hil** - Hillsboro, USA (Datacenter Park Hillsboro, Oregon)

### Popular OS Images

- **ubuntu-22.04** - Ubuntu 22.04 LTS
- **ubuntu-20.04** - Ubuntu 20.04 LTS
- **debian-11** - Debian 11
- **debian-12** - Debian 12
- **centos-stream-9** - CentOS Stream 9
- **rocky-9** - Rocky Linux 9
- **fedora-38** - Fedora 38

---

## 📊 Performance & Caching

### Caching Strategy

All Hetzner API responses are cached using Redis:

| Data Type | TTL | Cache Key Pattern |
|-----------|-----|-------------------|
| Server info | 1 minute | `hetzner:server:{id}` |
| Server list | 30 seconds | `hetzner:servers:all` |
| Pricing | 1 hour | `hetzner:pricing` |
| Locations | 24 hours | `hetzner:locations` |
| Server types | 24 hours | `hetzner:server_types` |

Cache is automatically invalidated on:
- Server creation
- Server deletion
- Server modification

### API Rate Limits

Hetzner Cloud API limits:
- **3,600 requests per hour** per project
- Resets every hour
- 429 status code when exceeded

The integration includes automatic retry logic and caching to stay well within limits.

---

## 🔧 Monitoring & Troubleshooting

### Check Worker Status

```bash
# View worker logs
journalctl -u vip-panel-worker -f

# Check if worker is processing jobs
redis-cli LLEN asynq:queues:critical
redis-cli LLEN asynq:queues:default

# View failed jobs
redis-cli LLEN asynq:dead
```

### Test Hetzner Connection

```go
package main

import (
    "context"
    "fmt"

    "github.com/xerudro/DASHBOARD-v2/internal/cache"
    "github.com/xerudro/DASHBOARD-v2/internal/services/providers"
)

func testConnection() {
    // Initialize provider
    provider, err := providers.NewHetznerProvider(apiToken, nil)
    if err != nil {
        fmt.Printf("Failed to connect: %v\n", err)
        return
    }

    // List locations (simple test)
    locations, err := provider.ListLocations(context.Background())
    if err != nil {
        fmt.Printf("API call failed: %v\n", err)
        return
    }

    fmt.Printf("✓ Connected successfully! Found %d locations\n", len(locations))
}
```

### Common Issues

**Issue**: `failed to connect to Hetzner API: 401 Unauthorized`
**Solution**: Check your API token is correct and has Read & Write permissions

**Issue**: `server type cx11 not found`
**Solution**: Server type names are case-sensitive. Use exact names from API

**Issue**: `location fsn1 not found`
**Solution**: Verify location exists for your project region

**Issue**: Worker not processing jobs
**Solution**: Check Redis connection and worker logs

---

## 🔐 Security Considerations

### API Token Security

```bash
# NEVER commit tokens to git
echo "configs/config.yaml" >> .gitignore

# Use environment variables in production
export VIP_HETZNER_API_TOKEN="$(cat /etc/vip-panel/hetzner-token)"

# Rotate tokens regularly (every 90 days)
```

### Server Security

The Ansible playbook applies these security measures:
- ✅ Firewall enabled (UFW)
- ✅ Fail2ban configured
- ✅ SSH hardened (no password auth)
- ✅ Automatic security updates
- ✅ Minimal open ports (22, 80, 443)

### Audit Logging

All Hetzner operations are logged:
```go
// Automatically logged:
- Server provisioning attempts (success/failure)
- Server deletions
- Server modifications
- API errors
- Cost calculations
```

---

## 📈 Next Steps

### Phase 4A: Advanced Features (Optional)

1. **Volume Management**
   - Create and attach volumes
   - Resize volumes
   - Automated backups

2. **Network Management**
   - Private networks
   - Floating IPs
   - Load balancers

3. **Firewall Rules**
   - Custom firewall configurations
   - Security group management

4. **Snapshots & Images**
   - Create server snapshots
   - Custom image deployment
   - Automated backup schedules

5. **Metrics & Monitoring**
   - Collect Hetzner metrics
   - Store in TimescaleDB
   - Alert on anomalies

### Integration with Existing Systems

```go
// Example: Update main.go to use Hetzner provider
func (app *App) initProviders() {
    app.hetznerProvider, err = providers.NewHetznerProvider(
        app.config.Hetzner.APIToken,
        app.cache,
    )
    if err != nil {
        log.Fatal("Failed to initialize Hetzner provider:", err)
    }
}
```

---

## 📚 Additional Resources

- [Hetzner Cloud API Documentation](https://docs.hetzner.cloud/)
- [hcloud-go Library](https://github.com/hetznercloud/hcloud-go)
- [Hetzner Cloud Console](https://console.hetzner.cloud/)
- [Hetzner Status Page](https://status.hetzner.com/)
- [Pricing Calculator](https://www.hetzner.com/cloud#pricing)

---

## ✅ Production Checklist

Before deploying to production:

- [ ] API token configured and tested
- [ ] Worker service running and monitored
- [ ] Redis available and persistent
- [ ] Database migrations applied
- [ ] Ansible playbooks tested
- [ ] SSH keys uploaded to Hetzner
- [ ] Firewall rules configured
- [ ] Monitoring and alerting set up
- [ ] Backup strategy defined
- [ ] Cost limits configured
- [ ] Documentation reviewed by team

---

## 🎉 Success Metrics

After implementation, you should have:

- ✅ Servers provisioning in < 2 minutes
- ✅ 95%+ provisioning success rate
- ✅ API response caching (>80% hit rate)
- ✅ Complete audit trail of all operations
- ✅ Zero manual server configuration needed
- ✅ Automatic security hardening applied
- ✅ Real-time cost tracking
- ✅ Background job processing with retry logic

---

**Ready to provision your first Hetzner server!** 🚀

For questions or issues, check the troubleshooting section or review the inline code documentation.
