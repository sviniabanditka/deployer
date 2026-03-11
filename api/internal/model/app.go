package model

import (
	"time"

	"github.com/google/uuid"
)

type AppStatus string

const (
	AppStatusCreated  AppStatus = "created"
	AppStatusBuilding AppStatus = "building"
	AppStatusRunning  AppStatus = "running"
	AppStatusStopped  AppStatus = "stopped"
	AppStatusFailed   AppStatus = "failed"
)

type App struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    AppStatus `json:"status"`
	Runtime   string    `json:"runtime"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Deployment struct {
	ID                uuid.UUID `json:"id"`
	AppID             uuid.UUID `json:"app_id"`
	Version           int       `json:"version"`
	Status            string    `json:"status"`
	ImageTag          string    `json:"image_tag"`
	BuildLog          string    `json:"build_log,omitempty"`
	IsPreview         bool      `json:"is_preview"`
	PreviewURL        string    `json:"preview_url,omitempty"`
	PullRequestNumber int       `json:"pull_request_number,omitempty"`
	PullRequestURL    string    `json:"pull_request_url,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
