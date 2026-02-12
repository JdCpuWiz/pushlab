package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	UserID    uuid.UUID       `json:"user_id" db:"user_id"`
	Title     *string         `json:"title,omitempty" db:"title"`
	Body      string          `json:"body" db:"body"`
	Data      json.RawMessage `json:"data,omitempty" db:"data"`
	Badge     *int            `json:"badge,omitempty" db:"badge"`
	Sound     string          `json:"sound" db:"sound"`
	Category  *string         `json:"category,omitempty" db:"category"`
	Priority  string          `json:"priority" db:"priority"`
	Tags      []string        `json:"tags,omitempty" db:"tags"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	Status    string          `json:"status" db:"status"`
}

type NotificationDelivery struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	NotificationID   uuid.UUID  `json:"notification_id" db:"notification_id"`
	DeviceTokenID    uuid.UUID  `json:"device_token_id" db:"device_token_id"`
	DeliveryStatus   string     `json:"delivery_status" db:"delivery_status"`
	AttemptCount     int        `json:"attempt_count" db:"attempt_count"`
	APNsResponseCode *int       `json:"apns_response_code,omitempty" db:"apns_response_code"`
	APNsErrorReason  *string    `json:"apns_error_reason,omitempty" db:"apns_error_reason"`
	DeliveredAt      *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type SendNotificationRequest struct {
	Title    *string                `json:"title,omitempty"`
	Body     string                 `json:"body"`
	Tags     []string               `json:"tags,omitempty"`
	Badge    *int                   `json:"badge,omitempty"`
	Sound    string                 `json:"sound,omitempty"`
	Category *string                `json:"category,omitempty"`
	Priority string                 `json:"priority,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type SendNotificationResponse struct {
	NotificationID uuid.UUID `json:"notification_id"`
	TargetDevices  int       `json:"target_devices"`
	Status         string    `json:"status"`
}

type NotificationDetail struct {
	Notification Notification           `json:"notification"`
	Deliveries   []NotificationDelivery `json:"deliveries"`
}

type NotificationJob struct {
	NotificationID uuid.UUID              `json:"notification_id"`
	UserID         uuid.UUID              `json:"user_id"`
	DeviceTokenIDs []uuid.UUID            `json:"device_token_ids"`
	Payload        NotificationPayload    `json:"payload"`
}

type NotificationPayload struct {
	Title    *string                `json:"title,omitempty"`
	Body     string                 `json:"body"`
	Badge    *int                   `json:"badge,omitempty"`
	Sound    string                 `json:"sound,omitempty"`
	Category *string                `json:"category,omitempty"`
	Priority string                 `json:"priority"`
	Data     map[string]interface{} `json:"data,omitempty"`
}
