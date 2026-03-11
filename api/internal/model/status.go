package model

import (
	"time"

	"github.com/google/uuid"
)

// SystemStatus represents the overall system health.
type SystemStatus struct {
	Overall    string            `json:"overall"` // operational, degraded, down
	Components []ComponentStatus `json:"components"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// ComponentStatus represents the health of an individual system component.
type ComponentStatus struct {
	Name    string        `json:"name"`
	Status  string        `json:"status"` // operational, degraded, down
	Latency time.Duration `json:"latency"`
	Message string        `json:"message,omitempty"`
}

// AppRuntimeStatus represents the runtime status of a deployed application.
type AppRuntimeStatus struct {
	Running      bool          `json:"running"`
	Healthy      bool          `json:"healthy"`
	Uptime       time.Duration `json:"uptime"`
	RestartCount int           `json:"restart_count"`
}

// Incident represents a recorded system incident.
type Incident struct {
	ID          uuid.UUID  `json:"id"`
	Component   string     `json:"component"`
	Description string     `json:"description"`
	Severity    string     `json:"severity"` // minor, major, critical
	Status      string     `json:"status"`   // investigating, identified, monitoring, resolved
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}
