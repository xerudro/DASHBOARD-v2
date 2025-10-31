-- VIP Hosting Panel - Seed Data
-- Migration: 002_seed_data

-- Insert default tenant (superadmin tenant)
INSERT INTO tenants (id, name, slug, plan, status) VALUES
('00000000-0000-0000-0000-000000000001', 'SuperAdmin', 'superadmin', 'unlimited', 'active');

-- Insert default superadmin user
-- Password: admin123 (bcrypt hash)
INSERT INTO users (id, tenant_id, email, password_hash, name, role, status, email_verified) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'admin@example.com',
 '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
 'System Administrator', 'superadmin', 'active', TRUE);

-- Insert default plans
INSERT INTO plans (id, name, slug, description, price_monthly, price_yearly, limits, features, status) VALUES
('00000000-0000-0000-0000-000000000001', 'Starter', 'starter', 'Perfect for small projects', 9.99, 99.00,
 '{"servers": 1, "sites": 5, "domains": 5, "databases": 5, "storage_gb": 10, "bandwidth_gb": 100}'::jsonb,
 ARRAY['1 Server', '5 Websites', '10GB Storage', '100GB Bandwidth', 'SSL Certificates', 'Daily Backups'],
 'active'),

('00000000-0000-0000-0000-000000000002', 'Professional', 'professional', 'For growing businesses', 29.99, 299.00,
 '{"servers": 3, "sites": 20, "domains": 20, "databases": 20, "storage_gb": 50, "bandwidth_gb": 500}'::jsonb,
 ARRAY['3 Servers', '20 Websites', '50GB Storage', '500GB Bandwidth', 'SSL Certificates', 'Hourly Backups', 'Priority Support'],
 'active'),

('00000000-0000-0000-0000-000000000003', 'Enterprise', 'enterprise', 'For large scale operations', 99.99, 999.00,
 '{"servers": 20, "sites": 100, "domains": 100, "databases": 100, "storage_gb": 500, "bandwidth_gb": 5000}'::jsonb,
 ARRAY['20 Servers', '100 Websites', '500GB Storage', '5TB Bandwidth', 'SSL Certificates', 'Real-time Backups', '24/7 Support', 'White Label'],
 'active');

-- Insert sample audit log
INSERT INTO audit_logs (tenant_id, user_id, action, resource_type, details) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'system.installed', 'system',
 '{"version": "2.0.0", "installer": "setup script"}'::jsonb);
