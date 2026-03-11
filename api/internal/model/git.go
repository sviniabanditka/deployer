package model

import (
	"time"

	"github.com/google/uuid"
)

type GitConnection struct {
	ID            uuid.UUID `json:"id"`
	AppID         uuid.UUID `json:"app_id"`
	Provider      string    `json:"provider"`
	RepoURL       string    `json:"repo_url"`
	RepoOwner     string    `json:"repo_owner"`
	RepoName      string    `json:"repo_name"`
	Branch        string    `json:"branch"`
	AutoDeploy    bool      `json:"auto_deploy"`
	WebhookSecret string    `json:"-"`
	AccessToken   string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
