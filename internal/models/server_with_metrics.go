package models

// ServerWithMetrics combines server and its latest metrics
type ServerWithMetrics struct {
	Server  *Server        `json:"server"`
	Metrics *ServerMetrics `json:"metrics,omitempty"`
}

// GetCPUDisplay returns CPU usage with N/A fallback
func (swm *ServerWithMetrics) GetCPUDisplay() string {
	if swm.Metrics == nil {
		return "N/A"
	}
	return swm.Metrics.GetCPUDisplay()
}

// GetMemoryDisplay returns memory usage with N/A fallback
func (swm *ServerWithMetrics) GetMemoryDisplay() string {
	if swm.Metrics == nil {
		return "N/A"
	}
	return swm.Metrics.GetMemoryDisplay()
}

// GetDiskDisplay returns disk usage with N/A fallback
func (swm *ServerWithMetrics) GetDiskDisplay() string {
	if swm.Metrics == nil {
		return "N/A"
	}
	return swm.Metrics.GetDiskDisplay()
}

// GetStatusDisplay returns server status with proper fallback
func (swm *ServerWithMetrics) GetStatusDisplay() string {
	if swm.Server.Status == "" {
		return "Unknown"
	}
	return swm.Server.Status
}

// GetHealthStatus returns overall health status
func (swm *ServerWithMetrics) GetHealthStatus() string {
	if swm.Metrics == nil {
		return "Unknown"
	}
	return swm.Metrics.GetHealthStatus()
}