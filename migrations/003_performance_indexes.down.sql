-- =====================================================
-- ROLLBACK PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================
-- Drops the performance indexes added in 003_performance_indexes.up.sql

DROP INDEX CONCURRENTLY IF EXISTS idx_server_metrics_covering;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_logs_tenant_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_sites_tenant_server_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_tenant_email;
DROP INDEX CONCURRENTLY IF EXISTS idx_servers_tenant_status;