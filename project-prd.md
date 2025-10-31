# VIP Hosting Panel - Product Requirements Document

### TL;DR

VIP Hosting Panel centralizes provisioning, management, and billing of web hosting across cost-efficient infrastructure providers like Hetzner. It delivers a simple, secure, multi-tenant control plane for superadmins, admins, resellers, and clients to create servers, deploy sites, manage DNS/SSL, monitor health, and automate backups. Primary audience: hosting providers, MSPs, agencies, and resellers seeking lower costs, faster operations, and stronger security without cPanel/Plesk licensing overhead.

---

## Goals

### Business Goals

* Reduce infrastructure and licensing costs by 30–50% vs. legacy panels within 6 months of adoption.

* Decrease average time-to-provision (server + site + SSL) to under 8 minutes.

* Achieve 60% reseller adoption among existing agency/MSP customers by end of quarter two post-launch.

* Decrease L1 support tickets per site by 25% via guided flows and automation.

* Reach >95% provisioning success rate and <1% month-over-month churn among active resellers.

### User Goals

* Create and manage sites and servers with minimal steps and clear defaults.

* Save operational time via one-click SSL/DNS, automated backups, and template-based deployments.

* Maintain strong security through least-privilege roles, 2FA/SSO, and immutable audit logs.

* Gain visibility with real-time health, usage, and cost insights per account, reseller, and client.

* Scale smoothly from a handful to thousands of sites without vendor lock-in.

### Non-Goals

* White-labeling and custom theming beyond basic logo/color in v1.

* App marketplace or plugin ecosystem in v1.

* Native mobile applications and in-panel full email server management in v1.

---

## User Stories

* Superadmin (Platform Owner)

  * As a Superadmin, I want to define global plans, quotas, and pricing, so that resellers and admins operate within cost and resource guardrails.

  * As a Superadmin, I want to connect and manage infrastructure providers (e.g., Hetzner), so that all tenants can provision reliably.

  * As a Superadmin, I want to view platform-wide health, costs, and usage, so that I can optimize margins and SLAs.

  * As a Superadmin, I want to enforce RBAC, SSO/2FA, and audit policies, so that the platform remains secure and compliant.

* Admin (Organization/Workspace Owner)

  * As an Admin, I want to spin up servers and deploy sites quickly, so that my team meets client deadlines.

  * As an Admin, I want to set quotas for teams/clients, so that resources and costs stay controlled.

  * As an Admin, I want to restore from backups and roll back deployments, so that I can recover from incidents fast.

  * As an Admin, I want to invite teammates with specific permissions, so that I can delegate safely.

* Reseller (Agency/MSP)

  * As a Reseller, I want to create client accounts with isolated billing and usage, so that I can offer managed hosting as a service.

  * As a Reseller, I want to apply templates (stack, PHP version, caching, WAF), so that I can deploy consistent environments.

  * As a Reseller, I want to see per-client margin and cost reports, so that I can price plans profitably.

  * As a Reseller, I want to brand invoices and emails lightly (logo/colors), so that my clients receive a cohesive experience.

* Client (End Customer)

  * As a Client, I want to add my domain and get SSL automatically, so that my site is secure without extra steps.

  * As a Client, I want to monitor uptime and resource usage, so that I know when to upgrade my plan.

  * As a Client, I want to create on-demand backups and restore, so that I can protect my data.

  * As a Client, I want to submit tickets or view runbooks, so that I can resolve issues quickly.

---

## Functional Requirements

* Authentication & Access Control (Priority: P0) -- RBAC: Role-based permissions for Superadmin/Admin/Reseller/Client; multi-tenant isolation. Competitive: Parity; simpler, opinionated defaults than cPanel/WHM role sprawl. -- SSO/OIDC + 2FA: Optional org-level enforcement. Competitive: Advantage; modern SSO not first-class in legacy panels. -- Audit Logging: Immutable event trails with export. Competitive: Advantage; built-in compliance focus.

* Provider & Server Management (Priority: P0) -- Provider Integration: Connect Hetzner via API tokens; manage regions, images, sizes. Competitive: Advantage on cost-awareness and direct cloud integration. -- Provisioning: One-click server setup with hardened baseline (firewall, users, packages). Competitive: Parity; faster defaults, modern hardening. -- Lifecycle: Reboot, resize, snapshot, terminate with safety checks. Competitive: Parity.

* Website & Application Management (Priority: P0) -- Site Creation: Deploy common stacks (e.g., PHP-FPM/Nginx, static) with presets. Competitive: Parity with reduced complexity. -- Domains & SSL: Attach domains; auto-issue/renew Let’s Encrypt. Competitive: Parity; streamlined 1-step flow. -- Deployments: Git-based or archive uploads; zero-downtime deploy and rollback. Competitive: Advantage vs. manual workflows.

* DNS Management (Priority: P1) -- DNS Providers: Connect Cloudflare or panel-managed DNS (PowerDNS). Competitive: Advantage; multi-provider vs. single-stack. -- Zone/Record Editor: A/AAAA/CNAME/TXT/MX; validation and propagation status. Competitive: Parity with better UX. -- Auto DNS: Auto-provision records on site creation. Competitive: Advantage in automation.

* Reseller & Tenant Management (Priority: P1) -- Client Accounts: Isolated billing, usage, permissions. Competitive: Parity with WHM; simpler flows. -- Plans & Quotas: Define CPU/RAM/Storage/Bandwidth caps per plan. Competitive: Parity. -- Branding (Light): Logo, colors on invoices/emails. Competitive: Minimal parity for v1.

* Billing & Invoicing (Priority: P1) -- Payments: Stripe integration for subscriptions and usage-based add-ons. Competitive: Advantage; integrated and transparent cost model. -- Invoices & Taxes: Automated invoices, VAT fields, downloadable PDFs. Competitive: Parity. -- Cost Insights: Provider cost pass-through and margin dashboards. Competitive: Advantage (cost-first design).

* Monitoring, Alerts & Logs (Priority: P1) -- Health Metrics: CPU/RAM/disk/network per server/site; thresholds. Competitive: Parity. -- Uptime Checks: HTTP/HTTPS checks with alerting. Competitive: Parity. -- Centralized Logs (Light): Access logs and error summaries with retention. Competitive: Advantage at panel level.

* Backups & Restore (Priority: P1) -- Policies: Scheduled full/incremental backups; retention rules. Competitive: Parity. -- Storage: Provider snapshots or S3-compatible offsite. Competitive: Advantage with multi-target support. -- Restore/Clone: Point-in-time restore and environment cloning. Competitive: Advantage.

* Analytics & Reporting (Priority: P2) -- Usage Reports: Per tenant, client, site, and server. Competitive: Parity. -- Cost & Margin: Trend and forecast. Competitive: Advantage. -- Export & Webhooks: CSV/JSON and outbound events. Competitive: Advantage.

* Automation & Public API (Priority: P2) -- REST API: CRUD for servers, sites, domains, backups. Competitive: Advantage vs. limited legacy APIs. -- Webhooks: Provisioning, SSL, backup, billing events. Competitive: Advantage. -- Templates: Save and share deployment templates. Competitive: Advantage.

* Migration Tools (Priority: P2) -- Importers: cPanel/Plesk archive imports (files + DB). Competitive: Parity. -- Domain & DNS Assist: Guided changes with verification. Competitive: Advantage (guided).

* Support & Help (Priority: P2) -- Guided Runbooks: Common fixes (SSL fail, DNS misconfig). Competitive: Advantage. -- Ticketing Integration: Link to external support tools. Competitive: Parity.

---

## User Experience

* Overall principles

  * Opinionated defaults; expert options tucked behind “Advanced”.

  * Clear, irreversible action warnings; inline validation and helpful copy.

  * Accessible, responsive UI with keyboard navigation and high contrast.

**Entry Point & First-Time User Experience**

* Discovery: Users invited by email or sign up via landing page with clear persona picker (Provider/Reseller/Client).

* Onboarding Wizard:

  * Step 1: Verify email, set 2FA (recommended), choose role.

  * Step 2: Connect provider (Hetzner API token) and test permissions.

  * Step 3: Create workspace, set currency, taxes, and billing (Stripe).

  * Step 4: Create first server or import existing site; empty states show walkthroughs.

  * Tooltips, checklists, and sample data available in sandbox mode.

**Core Experience**

* Step 1: Create Server

  * UI: “New Server” modal with provider/region/size/image presets and warning on cost.

  * Validation: API token/limits; SSH key presence; firewall rules.

  * Success: Provision job queued; status chip (Queued/Building/Ready) with live logs.

* Step 2: Create Site

  * UI: “New Site” with template presets (WordPress/PHP/Static), PHP version, cache/WAF toggles.

  * Validation: Disk space, quota, compatible PHP/runtime.

  * Success: Site deployed with dashboard card; next prompt to add domain.

* Step 3: Add Domain & SSL

  * UI: Domain input; DNS provider connect or verify existing DNS.

  * Validation: Ownership checks (TXT/HTTP); DNS propagation checks.

  * Success: SSL issued; auto-renew enabled; HTTPS enforced toggle visible.

* Step 4: Configure Backups

  * UI: Choose policy (daily/weekly), retention, target (snapshot/S3).

  * Validation: Storage limits, credentials test.

  * Success: Schedule confirmation; first backup job queued.

* Step 5: Monitoring & Alerts

  * UI: Resource graph with thresholds and alert channels (email/webhook).

  * Validation: Contact methods, rate-limits.

  * Success: Alerts active; test alert option.

* Step 6: Reseller Setup (if applicable)

  * UI: Create client account, assign plan/quota, light branding.

  * Success: Invitation sent; client sees simplified dashboard.

* Error Handling

  * If provisioning fails: Show actionable error (quota exceeded, token invalid) and one-click retry.

  * If SSL fails: Provide ACME challenges, DNS record guidance, and retry with rollback.

  * If backup fails: Credential validator and last-known-good status; auto-fallback to snapshot.

**Advanced Features & Edge Cases**

* Blue/Green Deployments: Optional two-slot deployment for zero-downtime.

* Server Resize/Scale-Up: Pre-check disk resizing and downtime estimate.

* API Rate Limits: Exponential backoff with user-visible queue status.

* DNS Provider Lock-In: Offer “Panel DNS” fallback if third-party limits reached.

* Billing Failures: Grace period and partial feature limitation with clear notices.

* Suspicious Activity: Auto-lock account after repeated failed logins; require admin unlock.

**UI/UX Highlights**

* Status indicators: Color-coded chips (Ready/Degraded/Action Required) with tooltips.

* Empty States: Educational prompts and quick actions.

* Global Search: Command palette to jump to servers, sites, domains.

* Responsive Layout: Breakpoints for desktop/tablet/mobile; table virtualization for large lists.

* Accessibility: WCAG AA targets, focus states, semantic markup, and ARIA attributes.

* Internationalization-ready copy and currency formats.

---

## Narrative

Sofia runs a mid-sized agency that hosts 300 client sites. Her current stack is split across multiple cPanel servers with rising license fees and inconsistent performance. Provisioning new servers takes hours, DNS changes are error-prone, and SSL renewals occasionally fail, leading to after-hours fire drills. Her team spends too much time on chores instead of client work.

Sofia adopts VIP Hosting Panel. In one onboarding flow, she connects her Hetzner account, sets default server templates, and creates a workspace with Stripe billing. She launches a new server in minutes with hardening and monitoring pre-configured. Deploying a client site is a guided, three-step experience: choose template, add domain, auto-issue SSL. DNS is handled through a connected provider; where that’s not possible, the panel gives precise guided records and real-time verification.

Within a week, Sofia moves 50 key sites using the migration assistant. Backups run on schedule with offsite storage, and uptime checks trigger proactive alerts. Her resellers have isolated client accounts with quotas and light branding. The cost dashboard shows provider spend, margins by client, and opportunities to consolidate under-used servers. Support tickets drop, time-to-provision shrinks, and the team reclaims hours each week.

For the business, VIP Hosting Panel reduces licensing costs and churn while increasing reseller adoption. For Sofia, it removes friction, standardizes operations, and restores confidence in her hosting practice.

---

## Success Metrics

* Activation: % of workspaces that complete provider connection and first server creation within 48 hours.

* Time-to-Value: Median time from signup to first SSL-secured site live under 8 minutes.

* Reliability: Provisioning success rate ≥ 95%; SSL issuance success ≥ 98%.

* Efficiency: 25% reduction in L1 hosting tickets per 100 sites.

* Cost: 30–50% decrease in combined infra + licensing cost per site vs. legacy panel baseline.

* Security: 100% audit coverage of privileged actions; 0 critical unpatched CVEs in supported images.

### User-Centric Metrics

* Onboarding completion rate and first-week retention.

* NPS/CSAT after first deployment and at 30 days.

* Average admin tasks completed per session; error rate per flow.

### Business Metrics

* Reseller adoption rate and ARPU.

* Gross margin per tenant; infrastructure cost per site trend.

* Churn rate and expansion revenue from plan upgrades.

### Technical Metrics

* API p95 latency < 300ms; job queue wait time p95 < 30s.

* Control plane uptime ≥ 99.9%; agent connectivity ≥ 99.5%.

* Backup success rate ≥ 97%; restore success rate ≥ 99%.

### Tracking Plan

* Account events: sign_up, complete_onboarding, connect_provider_success/fail.

* Provisioning: create_server_requested/succeeded/failed; resize/reboot/snapshot events.

* Site: create_site_requested/succeeded/failed; deploy_started/completed/rolled_back.

* Domain/SSL: domain_added, dns_verified, acme_issued/failed, ssl_renewed.

* Backups: policy_created, backup_started/completed/failed, restore_started/completed/failed.

* Monitoring: uptime_check_failed/recovered; alert_created/acknowledged.

* Billing: payment_method_added, invoice_created/paid/failed, plan_changed.

* Security: login_success/fail, 2FA_enabled, role_changed, audit_log_exported.

---

## Technical Considerations

### Technical Needs

* Control Plane API: Multi-tenant REST API; RBAC enforcement; audit logging; idempotent jobs.

* Worker/Orchestrator: Queues provisioning tasks, monitors progress, handles retries/backoff.

* Server Bootstrap: Cloud-init or bootstrap scripts; agentless via SSH with optional lightweight agent for telemetry/logs.

* Front-End App: SPA with responsive layout; real-time status via websockets or SSE.

* ACME Service: Let’s Encrypt integration for issuance/renewal with DNS-01/HTTP-01 challenges.

* Monitoring Subsystem: Metrics collection (exporters), uptime pings, alert rules.

* Billing Module: Subscriptions, usage metering, invoicing, proration, taxes.

* Template Engine: Save/apply stack templates, environment variables, and hooks.

### Integration Points

* Infrastructure: Hetzner Cloud API (servers, snapshots, firewall, networking).

* DNS: Cloudflare API; panel-managed DNS (e.g., PowerDNS) for fallback; registrar-agnostic guidance.

* SSL: Let’s Encrypt/ACME directory for issuance and renewal.

* Billing: Stripe for payments, invoicing, and webhooks.

* Notifications: Email (transactional provider), webhooks to Slack/MS Teams if configured.

* Identity: OIDC/SAML SSO (optional enterprise), 2FA (TOTP).

### Data Storage & Privacy

* Data Model: Tenants, Users, Roles, Providers, Servers, Sites, Domains, Plans, Invoices, Metrics, Backups, AuditLogs.

* Storage Strategy: Relational DB for core entities and audit logs; object storage (S3-compatible) for backup archives/exports; timeseries DB or columnar store for metrics.

* Privacy/Compliance: Data minimization; PII encryption at rest; secrets vaulted; TLS in transit; GDPR-aligned data processing; regional data residency configurable where possible.

* Retention: Configurable retention for logs/metrics/backups; secure deletion routines.

### Scalability & Performance

* Expected Load v1: 100 tenants, 1,000 servers, 10,000 sites.

* Throughput: Burst provisioning up to 50 concurrent jobs; rate-limited to provider quotas.

* Caching: Provider metadata and cost catalogs cache with sensible TTLs.

* Horizontal Scale: Stateless API/FE; workers scaled by job backlog; backpressure on user actions when safe.

### Potential Challenges

* Provider Quotas/Rate Limits: Implement adaptive backoff and job batching.

* Heterogeneous Environments: Standardize baselines; detect drift and remediate.

* DNS Propagation: Offer verification timeouts and staging certificates.

* Security Risks: Key management, privilege escalation; enforce least-privilege SSH and rotate credentials.

* Operational Complexity: Clear observability (tracing/logging/metrics) and runbooks for incident response.

---

## Milestones & Sequencing

Lean roadmap focused on rapid, safe delivery with iterative value.

### Project Estimate

* Large: 4–8 weeks (target \~6 weeks to beta)

### Team Size & Composition

* Small Team: 2 total people

  * Founding Engineer: Backend, integrations, infra, basic frontend.

  * Product/Designer: UX flows, UI components, copy, QA/UAT, light frontend.

### Suggested Phases

* Discovery & Foundations (2–3 days)

  * Key Deliverables: Product/Designer—personas, core flows, low-fi wireframes; Engineer—data model, job orchestration design, integration stubs.

  * Dependencies: Access to Hetzner sandbox, Stripe test account, DNS test domain.

* Control Plane & Provider Integration (1.5 weeks)

  * Key Deliverables: Engineer—multi-tenant API, RBAC, Hetzner provisioning (create/reboot/snapshot/terminate), audit logs; Product—hi-fi prototypes for server flows.

  * Dependencies: Hetzner API credentials, bootstrap images.

* Sites, Domains & SSL (1.5 weeks)

  * Key Deliverables: Engineer—site deployment templates, domain attach, ACME issuance/renewal, guided DNS verification; Product—UX for domain/SSL and error states.

  * Dependencies: ACME directory access, DNS provider API keys or panel DNS.

* Reseller, Billing & Quotas (1 week)

  * Key Deliverables: Engineer—tenant/reseller accounts, plans/quotas, Stripe subscriptions/invoices; Product—billing screens, margin insights MVP.

  * Dependencies: Stripe webhook endpoints, tax settings.

* Monitoring, Backups & Alerts (1 week)

  * Key Deliverables: Engineer—health metrics, uptime checks, backup policies (snapshot + S3), alerting; Product—dashboards and backup UI.

  * Dependencies: S3-compatible storage, email provider.

* Hardening, Docs & Beta Launch (3–5 days)

  * Key Deliverables: Security review, rate limiting, onboarding wizard polish, tracking plan, runbooks, initial analytics; Closed beta with select resellers.

  * Dependencies: Domain/SSL for panel, telemetry pipeline, support channel setup.