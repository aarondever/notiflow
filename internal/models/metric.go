package models

import "time"

// ApplicationMetrics tracks runtime metrics for the application
type ApplicationMetrics struct {
	StartTime   time.Time     `json:"start_time"`   // Application start timestamp
	TotalUptime time.Duration `json:"total_uptime"` // Total application uptime
}
