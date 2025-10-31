-- VIP Hosting Panel - Initial Schema Rollback
-- Migration: 001_initial_schema DOWN

-- Drop triggers
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS update_plans_updated_at ON plans;
DROP TRIGGER IF EXISTS update_databases_updated_at ON databases;
DROP TRIGGER IF EXISTS update_ssl_certificates_updated_at ON ssl_certificates;
DROP TRIGGER IF EXISTS update_dns_records_updated_at ON dns_records;
DROP TRIGGER IF EXISTS update_dns_zones_updated_at ON dns_zones;
DROP TRIGGER IF EXISTS update_sites_updated_at ON sites;
DROP TRIGGER IF EXISTS update_servers_updated_at ON servers;
DROP TRIGGER IF EXISTS update_providers_updated_at ON providers;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS jobs CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
DROP TABLE IF EXISTS uptime_checks CASCADE;
DROP TABLE IF EXISTS server_metrics CASCADE;
DROP TABLE IF EXISTS backups CASCADE;
DROP TABLE IF EXISTS databases CASCADE;
DROP TABLE IF EXISTS ssl_certificates CASCADE;
DROP TABLE IF EXISTS dns_records CASCADE;
DROP TABLE IF EXISTS dns_zones CASCADE;
DROP TABLE IF EXISTS sites CASCADE;
DROP TABLE IF EXISTS servers CASCADE;
DROP TABLE IF EXISTS providers CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
