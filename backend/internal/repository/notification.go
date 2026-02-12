package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pushlab/backend/internal/models"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (user_id, title, body, data, badge, sound, category, priority, tags, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		notification.UserID, notification.Title, notification.Body, notification.Data,
		notification.Badge, notification.Sound, notification.Category, notification.Priority,
		notification.Tags, notification.Status,
	).Scan(&notification.ID, &notification.CreatedAt)
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var notif models.Notification
	query := `
		SELECT id, user_id, title, body, data, badge, sound, category, priority, tags, created_at, status
		FROM notifications WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notif.ID, &notif.UserID, &notif.Title, &notif.Body, &notif.Data,
		&notif.Badge, &notif.Sound, &notif.Category, &notif.Priority,
		&notif.Tags, &notif.CreatedAt, &notif.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return &notif, nil
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, title, body, data, badge, sound, category, priority, tags, created_at, status
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notif models.Notification
		if err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Title, &notif.Body, &notif.Data,
			&notif.Badge, &notif.Sound, &notif.Category, &notif.Priority,
			&notif.Tags, &notif.CreatedAt, &notif.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE notifications SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

// Delivery operations

func (r *NotificationRepository) CreateDelivery(ctx context.Context, delivery *models.NotificationDelivery) error {
	query := `
		INSERT INTO notification_deliveries (notification_id, device_token_id, delivery_status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, delivery.NotificationID, delivery.DeviceTokenID, delivery.DeliveryStatus).
		Scan(&delivery.ID, &delivery.CreatedAt, &delivery.UpdatedAt)
}

func (r *NotificationRepository) GetDeliveriesByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]models.NotificationDelivery, error) {
	query := `
		SELECT id, notification_id, device_token_id, delivery_status, attempt_count,
		       apns_response_code, apns_error_reason, delivered_at, created_at, updated_at
		FROM notification_deliveries
		WHERE notification_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []models.NotificationDelivery
	for rows.Next() {
		var delivery models.NotificationDelivery
		if err := rows.Scan(
			&delivery.ID, &delivery.NotificationID, &delivery.DeviceTokenID,
			&delivery.DeliveryStatus, &delivery.AttemptCount, &delivery.APNsResponseCode,
			&delivery.APNsErrorReason, &delivery.DeliveredAt, &delivery.CreatedAt, &delivery.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delivery: %w", err)
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, nil
}

func (r *NotificationRepository) UpdateDeliveryStatus(ctx context.Context, delivery *models.NotificationDelivery) error {
	query := `
		UPDATE notification_deliveries
		SET delivery_status = $2, attempt_count = $3, apns_response_code = $4,
		    apns_error_reason = $5, delivered_at = $6
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		delivery.ID, delivery.DeliveryStatus, delivery.AttemptCount,
		delivery.APNsResponseCode, delivery.APNsErrorReason, delivery.DeliveredAt,
	).Scan(&delivery.UpdatedAt)
}
