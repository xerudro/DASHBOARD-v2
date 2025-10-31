package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// MetricsCollector collects and exposes application metrics
type MetricsCollector struct {
	startTime     time.Time
	requestCount  int64
	errorCount    int64
	responseTime  time.Duration
	activeUsers   int64
	systemMetrics *SystemMetrics
}

// SystemMetrics holds system-level metrics
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	Goroutines  int     `json:"goroutines"`
	HeapSize    int64   `json:"heap_size"`
	GCCount     uint32  `json:"gc_count"`
}

// ApplicationMetrics holds application-level metrics
type ApplicationMetrics struct {
	Uptime       string         `json:"uptime"`
	RequestCount int64          `json:"request_count"`
	ErrorCount   int64          `json:"error_count"`
	ErrorRate    float64        `json:"error_rate"`
	AvgResponse  string         `json:"avg_response_time"`
	ActiveUsers  int64          `json:"active_users"`
	System       *SystemMetrics `json:"system"`
}

// HealthStatus represents the health status of the application
type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	collector := &MetricsCollector{
		startTime:     time.Now(),
		systemMetrics: &SystemMetrics{},
	}
	
	// Start metrics collection goroutine
	go collector.collectSystemMetrics()
	
	return collector
}

// collectSystemMetrics collects system metrics periodically
func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		mc.systemMetrics = &SystemMetrics{
			CPUUsage:    getCPUUsage(),
			MemoryUsage: int64(m.Alloc),
			Goroutines:  runtime.NumGoroutine(),
			HeapSize:    int64(m.HeapAlloc),
			GCCount:     m.NumGC,
		}
	}
}

// getCPUUsage returns CPU usage percentage (simplified)
func getCPUUsage() float64 {
	// This is a simplified CPU usage calculation
	// In production, use a proper CPU monitoring library
	return 0.0 // Placeholder
}

// RecordRequest records a request metric
func (mc *MetricsCollector) RecordRequest(duration time.Duration, statusCode int) {
	mc.requestCount++
	mc.responseTime += duration
	
	if statusCode >= 400 {
		mc.errorCount++
	}
}

// GetMetrics returns current application metrics
func (mc *MetricsCollector) GetMetrics() *ApplicationMetrics {
	uptime := time.Since(mc.startTime)
	errorRate := 0.0
	if mc.requestCount > 0 {
		errorRate = float64(mc.errorCount) / float64(mc.requestCount) * 100
	}
	
	avgResponse := time.Duration(0)
	if mc.requestCount > 0 {
		avgResponse = mc.responseTime / time.Duration(mc.requestCount)
	}
	
	return &ApplicationMetrics{
		Uptime:       uptime.String(),
		RequestCount: mc.requestCount,
		ErrorCount:   mc.errorCount,
		ErrorRate:    errorRate,
		AvgResponse:  avgResponse.String(),
		ActiveUsers:  mc.activeUsers,
		System:       mc.systemMetrics,
	}
}

// MetricsMiddleware provides metrics collection middleware
func (mc *MetricsCollector) MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		err := c.Next()
		
		duration := time.Since(start)
		mc.RecordRequest(duration, c.Response().StatusCode())
		
		return err
	}
}

// HealthChecker performs health checks
type HealthChecker struct {
	version     string
	environment string
	checks      map[string]func() HealthCheck
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version, environment string) *HealthChecker {
	hc := &HealthChecker{
		version:     version,
		environment: environment,
		checks:      make(map[string]func() HealthCheck),
	}
	
	// Register default health checks
	hc.RegisterCheck("database", hc.checkDatabase)
	hc.RegisterCheck("redis", hc.checkRedis)
	hc.RegisterCheck("memory", hc.checkMemory)
	hc.RegisterCheck("disk", hc.checkDisk)
	
	return hc
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, check func() HealthCheck) {
	hc.checks[name] = check
}

// CheckHealth performs all health checks
func (hc *HealthChecker) CheckHealth(startTime time.Time) *HealthStatus {
	status := &HealthStatus{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     hc.version,
		Environment: hc.environment,
		Uptime:      time.Since(startTime).String(),
		Checks:      make(map[string]HealthCheck),
	}
	
	// Run all health checks
	for name, check := range hc.checks {
		result := check()
		status.Checks[name] = result
		
		// If any check fails, mark overall status as unhealthy
		if result.Status != "healthy" {
			status.Status = "unhealthy"
		}
	}
	
	return status
}

// checkDatabase checks database connectivity
func (hc *HealthChecker) checkDatabase() HealthCheck {
	// Placeholder - implement actual database health check
	return HealthCheck{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Time:    time.Now(),
	}
}

// checkRedis checks Redis connectivity
func (hc *HealthChecker) checkRedis() HealthCheck {
	// Placeholder - implement actual Redis health check
	return HealthCheck{
		Status:  "healthy",
		Message: "Redis connection is healthy",
		Time:    time.Now(),
	}
}

// checkMemory checks memory usage
func (hc *HealthChecker) checkMemory() HealthCheck {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Convert bytes to MB
	allocMB := m.Alloc / 1024 / 1024
	
	// Check if memory usage is too high (>500MB as example)
	if allocMB > 500 {
		return HealthCheck{
			Status:  "warning",
			Message: fmt.Sprintf("High memory usage: %d MB", allocMB),
			Time:    time.Now(),
		}
	}
	
	return HealthCheck{
		Status:  "healthy",
		Message: fmt.Sprintf("Memory usage: %d MB", allocMB),
		Time:    time.Now(),
	}
}

// checkDisk checks disk space
func (hc *HealthChecker) checkDisk() HealthCheck {
	// Placeholder - implement actual disk space check
	return HealthCheck{
		Status:  "healthy",
		Message: "Disk space is sufficient",
		Time:    time.Now(),
	}
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	alerts        []Alert
	thresholds    map[string]float64
	notifications []NotificationChannel
}

// Alert represents an alert
type Alert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// NotificationChannel represents a notification channel
type NotificationChannel interface {
	Send(alert Alert) error
}

// EmailNotification sends alerts via email
type EmailNotification struct {
	SMTPServer string
	Username   string
	Password   string
	Recipients []string
}

// Send sends an alert via email
func (en *EmailNotification) Send(alert Alert) error {
	// Placeholder - implement actual email sending
	log.Info().
		Str("alert_id", alert.ID).
		Str("type", alert.Type).
		Str("severity", alert.Severity).
		Str("message", alert.Message).
		Msg("Alert sent via email")
	return nil
}

// WebhookNotification sends alerts via webhook
type WebhookNotification struct {
	URL     string
	Headers map[string]string
}

// Send sends an alert via webhook
func (wn *WebhookNotification) Send(alert Alert) error {
	// Placeholder - implement actual webhook sending
	log.Info().
		Str("alert_id", alert.ID).
		Str("webhook_url", wn.URL).
		Msg("Alert sent via webhook")
	return nil
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	am := &AlertManager{
		alerts:     make([]Alert, 0),
		thresholds: make(map[string]float64),
	}
	
	// Set default thresholds
	am.thresholds["error_rate"] = 5.0     // 5% error rate
	am.thresholds["response_time"] = 1000 // 1000ms response time
	am.thresholds["memory_usage"] = 80.0  // 80% memory usage
	am.thresholds["cpu_usage"] = 80.0     // 80% CPU usage
	
	return am
}

// AddNotificationChannel adds a notification channel
func (am *AlertManager) AddNotificationChannel(channel NotificationChannel) {
	am.notifications = append(am.notifications, channel)
}

// CheckThresholds checks metrics against thresholds and creates alerts
func (am *AlertManager) CheckThresholds(metrics *ApplicationMetrics) {
	// Check error rate
	if metrics.ErrorRate > am.thresholds["error_rate"] {
		alert := Alert{
			ID:        fmt.Sprintf("error_rate_%d", time.Now().Unix()),
			Type:      "error_rate",
			Severity:  "warning",
			Message:   fmt.Sprintf("High error rate: %.2f%%", metrics.ErrorRate),
			Timestamp: time.Now(),
			Resolved:  false,
		}
		am.triggerAlert(alert)
	}
	
	// Check memory usage
	memoryUsageMB := float64(metrics.System.MemoryUsage) / 1024 / 1024
	if memoryUsageMB > 500 { // 500MB threshold
		alert := Alert{
			ID:        fmt.Sprintf("memory_usage_%d", time.Now().Unix()),
			Type:      "memory_usage",
			Severity:  "warning",
			Message:   fmt.Sprintf("High memory usage: %.0f MB", memoryUsageMB),
			Timestamp: time.Now(),
			Resolved:  false,
		}
		am.triggerAlert(alert)
	}
	
	// Check goroutine count
	if metrics.System.Goroutines > 1000 {
		alert := Alert{
			ID:        fmt.Sprintf("goroutines_%d", time.Now().Unix()),
			Type:      "goroutines",
			Severity:  "critical",
			Message:   fmt.Sprintf("High goroutine count: %d", metrics.System.Goroutines),
			Timestamp: time.Now(),
			Resolved:  false,
		}
		am.triggerAlert(alert)
	}
}

// triggerAlert triggers an alert and sends notifications
func (am *AlertManager) triggerAlert(alert Alert) {
	am.alerts = append(am.alerts, alert)
	
	// Send notifications
	for _, channel := range am.notifications {
		if err := channel.Send(alert); err != nil {
			log.Error().
				Err(err).
				Str("alert_id", alert.ID).
				Msg("Failed to send alert notification")
		}
	}
	
	log.Warn().
		Str("alert_id", alert.ID).
		Str("type", alert.Type).
		Str("severity", alert.Severity).
		Str("message", alert.Message).
		Msg("Alert triggered")
}

// GetActiveAlerts returns all active (unresolved) alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	var activeAlerts []Alert
	for _, alert := range am.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	return activeAlerts
}

// ResolveAlert resolves an alert by ID
func (am *AlertManager) ResolveAlert(alertID string) {
	for i, alert := range am.alerts {
		if alert.ID == alertID && !alert.Resolved {
			now := time.Now()
			am.alerts[i].Resolved = true
			am.alerts[i].ResolvedAt = &now
			
			log.Info().
				Str("alert_id", alertID).
				Msg("Alert resolved")
			break
		}
	}
}

// MonitoringService provides comprehensive monitoring
type MonitoringService struct {
	metrics      *MetricsCollector
	health       *HealthChecker
	alerts       *AlertManager
	startTime    time.Time
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(version, environment string) *MonitoringService {
	return &MonitoringService{
		metrics:   NewMetricsCollector(),
		health:    NewHealthChecker(version, environment),
		alerts:    NewAlertManager(),
		startTime: time.Now(),
	}
}

// Start starts the monitoring service
func (ms *MonitoringService) Start() {
	// Start periodic health checks and alerting
	go ms.periodicChecks()
	
	log.Info().Msg("Monitoring service started")
}

// periodicChecks runs periodic health checks and alerting
func (ms *MonitoringService) periodicChecks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// Get current metrics
		metrics := ms.metrics.GetMetrics()
		
		// Check thresholds and trigger alerts
		ms.alerts.CheckThresholds(metrics)
		
		// Log metrics for debugging
		if metrics.RequestCount > 0 {
			log.Debug().
				Int64("requests", metrics.RequestCount).
				Float64("error_rate", metrics.ErrorRate).
				Str("avg_response", metrics.AvgResponse).
				Int64("memory_mb", metrics.System.MemoryUsage/1024/1024).
				Int("goroutines", metrics.System.Goroutines).
				Msg("Metrics snapshot")
		}
	}
}

// GetMetrics returns current metrics
func (ms *MonitoringService) GetMetrics() *ApplicationMetrics {
	return ms.metrics.GetMetrics()
}

// GetHealth returns current health status
func (ms *MonitoringService) GetHealth() *HealthStatus {
	return ms.health.CheckHealth(ms.startTime)
}

// GetAlerts returns active alerts
func (ms *MonitoringService) GetAlerts() []Alert {
	return ms.alerts.GetActiveAlerts()
}

// Middleware returns the monitoring middleware
func (ms *MonitoringService) Middleware() fiber.Handler {
	return ms.metrics.MetricsMiddleware()
}

// SetupMonitoringRoutes sets up monitoring HTTP routes
func SetupMonitoringRoutes(app *fiber.App, service *MonitoringService) {
	monitoring := app.Group("/monitoring")
	
	// Metrics endpoint
	monitoring.Get("/metrics", func(c *fiber.Ctx) error {
		metrics := service.GetMetrics()
		return c.JSON(metrics)
	})
	
	// Health endpoint
	monitoring.Get("/health", func(c *fiber.Ctx) error {
		health := service.GetHealth()
		
		// Set appropriate status code
		statusCode := fiber.StatusOK
		if health.Status == "unhealthy" {
			statusCode = fiber.StatusServiceUnavailable
		} else if health.Status == "warning" {
			statusCode = fiber.StatusOK // Still considered healthy
		}
		
		return c.Status(statusCode).JSON(health)
	})
	
	// Alerts endpoint
	monitoring.Get("/alerts", func(c *fiber.Ctx) error {
		alerts := service.GetAlerts()
		return c.JSON(fiber.Map{
			"alerts": alerts,
			"count":  len(alerts),
		})
	})
	
	// Resolve alert endpoint
	monitoring.Post("/alerts/:id/resolve", func(c *fiber.Ctx) error {
		alertID := c.Params("id")
		service.alerts.ResolveAlert(alertID)
		return c.JSON(fiber.Map{
			"message": "Alert resolved",
			"id":      alertID,
		})
	})
}

// PrometheusExporter exports metrics in Prometheus format
type PrometheusExporter struct {
	service *MonitoringService
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(service *MonitoringService) *PrometheusExporter {
	return &PrometheusExporter{service: service}
}

// Export exports metrics in Prometheus format
func (pe *PrometheusExporter) Export() string {
	metrics := pe.service.GetMetrics()
	
	var output strings.Builder
	
	output.WriteString("# HELP vip_panel_requests_total Total number of requests\n")
	output.WriteString("# TYPE vip_panel_requests_total counter\n")
	output.WriteString(fmt.Sprintf("vip_panel_requests_total %d\n", metrics.RequestCount))
	
	output.WriteString("# HELP vip_panel_errors_total Total number of errors\n")
	output.WriteString("# TYPE vip_panel_errors_total counter\n")
	output.WriteString(fmt.Sprintf("vip_panel_errors_total %d\n", metrics.ErrorCount))
	
	output.WriteString("# HELP vip_panel_memory_bytes Memory usage in bytes\n")
	output.WriteString("# TYPE vip_panel_memory_bytes gauge\n")
	output.WriteString(fmt.Sprintf("vip_panel_memory_bytes %d\n", metrics.System.MemoryUsage))
	
	output.WriteString("# HELP vip_panel_goroutines Number of goroutines\n")
	output.WriteString("# TYPE vip_panel_goroutines gauge\n")
	output.WriteString(fmt.Sprintf("vip_panel_goroutines %d\n", metrics.System.Goroutines))
	
	return output.String()
}