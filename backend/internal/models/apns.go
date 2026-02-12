package models

import (
	"time"

	"github.com/google/uuid"
)

type APNsCredential struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	TeamID         string    `json:"team_id" db:"team_id"`
	KeyID          string    `json:"key_id" db:"key_id"`
	BundleID       string    `json:"bundle_id" db:"bundle_id"`
	Environment    string    `json:"environment" db:"environment"`
	PrivateKeyPath string    `json:"-" db:"private_key_path"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	IsActive       bool      `json:"is_active" db:"is_active"`
}

type CreateAPNsCredentialRequest struct {
	TeamID      string `json:"team_id"`
	KeyID       string `json:"key_id"`
	BundleID    string `json:"bundle_id"`
	Environment string `json:"environment"`
	PrivateKey  string `json:"private_key"`
}
