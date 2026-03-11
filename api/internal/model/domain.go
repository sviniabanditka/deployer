package model

import (
	"time"

	"github.com/google/uuid"
)

type DomainStatus string

const (
	DomainStatusPendingVerification DomainStatus = "pending_verification"
	DomainStatusVerified            DomainStatus = "verified"
	DomainStatusFailed              DomainStatus = "failed"
)

type CustomDomain struct {
	ID                uuid.UUID    `json:"id"`
	AppID             uuid.UUID    `json:"app_id"`
	Domain            string       `json:"domain"`
	VerificationToken string       `json:"verification_token"`
	Status            DomainStatus `json:"status"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}
