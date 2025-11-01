package models

// ServerWithMetrics combines server and its latest metrics
type ServerWithMetrics struct {
	Server       *Server        `json:"server"`
	Metrics      *ServerMetrics `json:"metrics,omitempty"`
	ProviderName string         `json:"provider_name,omitempty"`
	ProviderType string         `json:"provider_type,omitempty"`
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

// GetProviderDisplay returns provider display name with fallback
func (swm *ServerWithMetrics) GetProviderDisplay() string {
	if swm.ProviderName != "" {
		return swm.ProviderName
	}
	if swm.ProviderType != "" {
		return swm.ProviderType
	}
	return "N/A"
}
