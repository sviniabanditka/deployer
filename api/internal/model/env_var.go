package model

import (
	"time"

	"github.com/google/uuid"
)

type EnvVar struct {
	ID        uuid.UUID `json:"id"`
	AppID     uuid.UUID `json:"app_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}
