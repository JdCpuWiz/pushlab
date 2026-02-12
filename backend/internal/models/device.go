package models

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	DeviceName       string     `json:"device_name" db:"device_name"`
	DeviceIdentifier string     `json:"device_identifier" db:"device_identifier"`
	Tags             []string   `json:"tags" db:"tags"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
}

type DeviceToken struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	DeviceID    uuid.UUID  `json:"device_id" db:"device_id"`
	Token       string     `json:"token" db:"token"`
	Environment string     `json:"environment" db:"environment"`
	BundleID    string     `json:"bundle_id" db:"bundle_id"`
	IssuedAt    time.Time  `json:"issued_at" db:"issued_at"`
	IsValid     bool       `json:"is_valid" db:"is_valid"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ErrorCount  int        `json:"error_count" db:"error_count"`
	LastError   *string    `json:"last_error,omitempty" db:"last_error"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type RegisterDeviceRequest struct {
	DeviceName       string   `json:"device_name"`
	DeviceIdentifier string   `json:"device_identifier"`
	DeviceToken      string   `json:"device_token"`
	BundleID         string   `json:"bundle_id"`
	Environment      string   `json:"environment"`
	Tags             []string `json:"tags"`
}

type UpdateDeviceRequest struct {
	DeviceName string   `json:"device_name,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type UpdateTokenRequest struct {
	DeviceToken string `json:"device_token"`
	Environment string `json:"environment"`
	BundleID    string `json:"bundle_id"`
}

type DeviceWithToken struct {
	Device      Device      `json:"device"`
	DeviceToken DeviceToken `json:"device_token"`
}
