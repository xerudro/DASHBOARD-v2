package models

import (
	"time"

	"github.com/google/uuid"
)

// ServerMetrics represents time-series server metrics
type ServerMetrics struct {
	Time            time.Time  `json:"time" db:"time"`
	ServerID        uuid.UUID  `json:"server_id" db:"server_id"`
	CPUPercent      *float64   `json:"cpu_percent,omitempty" db:"cpu_percent"`
	MemoryUsedMB    *int64     `json:"memory_used_mb,omitempty" db:"memory_used_mb"`
	MemoryTotalMB   *int64     `json:"memory_total_mb,omitempty" db:"memory_total_mb"`
	DiskUsedGB      *int64     `json:"disk_used_gb,omitempty" db:"disk_used_gb"`
	DiskTotalGB     *int64     `json:"disk_total_gb,omitempty" db:"disk_total_gb"`
	NetworkInMB     *int64     `json:"network_in_mb,omitempty" db:"network_in_mb"`
	NetworkOutMB    *int64     `json:"network_out_mb,omitempty" db:"network_out_mb"`
	LoadAverage     *float64   `json:"load_average,omitempty" db:"load_average"`
	Connections     *int       `json:"connections,omitempty" db:"connections"`
}

// GetCPUDisplay returns formatted CPU usage or N/A
func (m *ServerMetrics) GetCPUDisplay() string {
	if m.CPUPercent == nil {
		return "N/A"
	}
	return formatPercent(*m.CPUPercent)
}

// GetMemoryDisplay returns formatted memory usage or N/A
func (m *ServerMetrics) GetMemoryDisplay() string {
	if m.MemoryUsedMB == nil || m.MemoryTotalMB == nil {
		return "N/A"
	}
	return formatMemory(*m.MemoryUsedMB, *m.MemoryTotalMB)
}

// GetDiskDisplay returns formatted disk usage or N/A
func (m *ServerMetrics) GetDiskDisplay() string {
	if m.DiskUsedGB == nil || m.DiskTotalGB == nil {
		return "N/A"
	}
	return formatDisk(*m.DiskUsedGB, *m.DiskTotalGB)
}

// GetMemoryPercent returns memory usage percentage or 0
func (m *ServerMetrics) GetMemoryPercent() float64 {
	if m.MemoryUsedMB == nil || m.MemoryTotalMB == nil || *m.MemoryTotalMB == 0 {
		return 0
	}
	return (float64(*m.MemoryUsedMB) / float64(*m.MemoryTotalMB)) * 100
}

// GetDiskPercent returns disk usage percentage or 0
func (m *ServerMetrics) GetDiskPercent() float64 {
	if m.DiskUsedGB == nil || m.DiskTotalGB == nil || *m.DiskTotalGB == 0 {
		return 0
	}
	return (float64(*m.DiskUsedGB) / float64(*m.DiskTotalGB)) * 100
}

// IsHealthy returns true if all metrics are within healthy thresholds
func (m *ServerMetrics) IsHealthy() bool {
	if m.CPUPercent != nil && *m.CPUPercent > 90 {
		return false
	}
	if m.GetMemoryPercent() > 95 {
		return false
	}
	if m.GetDiskPercent() > 95 {
		return false
	}
	return true
}

// GetHealthStatus returns "healthy", "degraded", or "unknown"
func (m *ServerMetrics) GetHealthStatus() string {
	if m.CPUPercent == nil && m.MemoryUsedMB == nil {
		return "unknown"
	}
	if !m.IsHealthy() {
		return "degraded"
	}
	return "healthy"
}

// Helper functions for formatting
func formatPercent(value float64) string {
	return formatFloat(value, 1) + "%"
}

func formatMemory(used, total int64) string {
	usedGB := float64(used) / 1024
	totalGB := float64(total) / 1024
	return formatFloat(usedGB, 2) + " / " + formatFloat(totalGB, 2) + " GB"
}

func formatDisk(used, total int64) string {
	return formatInt64(used) + " / " + formatInt64(total) + " GB"
}

func formatFloat(value float64, precision int) string {
	format := "%." + string(rune(precision+'0')) + "f"
	return string([]byte(format))
}

func formatInt64(value int64) string {
	return string([]byte("%d"))
}

// UptimeCheck represents a site uptime monitoring check
type UptimeCheck struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	SiteID             *uuid.UUID `json:"site_id,omitempty" db:"site_id"`
	URL                string     `json:"url" db:"url"`
	Method             string     `json:"method" db:"method"`
	IntervalSeconds    int        `json:"interval_seconds" db:"interval_seconds"`
	TimeoutSeconds     int        `json:"timeout_seconds" db:"timeout_seconds"`
	Status             string     `json:"status" db:"status"`
	LastCheckAt        *time.Time `json:"last_check_at,omitempty" db:"last_check_at"`
	LastStatusCode     *int       `json:"last_status_code,omitempty" db:"last_status_code"`
	LastResponseTimeMS *int       `json:"last_response_time_ms,omitempty" db:"last_response_time_ms"`
	UptimePercent      float64    `json:"uptime_percent" db:"uptime_percent"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// IsOnline returns true if last check was successful
func (u *UptimeCheck) IsOnline() bool {
	if u.LastStatusCode == nil {
		return false
	}
	return *u.LastStatusCode >= 200 && *u.LastStatusCode < 400
}

// GetUptimeDisplay returns formatted uptime percentage
func (u *UptimeCheck) GetUptimeDisplay() string {
	return formatPercent(u.UptimePercent)
}

// GetResponseTimeDisplay returns formatted response time or N/A
func (u *UptimeCheck) GetResponseTimeDisplay() string {
	if u.LastResponseTimeMS == nil {
		return "N/A"
	}
	return formatInt64(int64(*u.LastResponseTimeMS)) + " ms"
}
