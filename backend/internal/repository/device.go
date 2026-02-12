package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pushlab/backend/internal/models"
)

type DeviceRepository struct {
	db *pgxpool.Pool
}

func NewDeviceRepository(db *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(ctx context.Context, device *models.Device) error {
	query := `
		INSERT INTO devices (user_id, device_name, device_identifier, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, device.UserID, device.DeviceName, device.DeviceIdentifier, device.Tags).
		Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
}

func (r *DeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var device models.Device
	query := `
		SELECT id, user_id, device_name, device_identifier, tags, created_at, updated_at, last_seen_at
		FROM devices WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&device.ID, &device.UserID, &device.DeviceName, &device.DeviceIdentifier,
		&device.Tags, &device.CreatedAt, &device.UpdatedAt, &device.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return &device, nil
}

func (r *DeviceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Device, error) {
	query := `
		SELECT id, user_id, device_name, device_identifier, tags, created_at, updated_at, last_seen_at
		FROM devices WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(
			&device.ID, &device.UserID, &device.DeviceName, &device.DeviceIdentifier,
			&device.Tags, &device.CreatedAt, &device.UpdatedAt, &device.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *DeviceRepository) GetByUserAndIdentifier(ctx context.Context, userID uuid.UUID, identifier string) (*models.Device, error) {
	var device models.Device
	query := `
		SELECT id, user_id, device_name, device_identifier, tags, created_at, updated_at, last_seen_at
		FROM devices WHERE user_id = $1 AND device_identifier = $2
	`
	err := r.db.QueryRow(ctx, query, userID, identifier).Scan(
		&device.ID, &device.UserID, &device.DeviceName, &device.DeviceIdentifier,
		&device.Tags, &device.CreatedAt, &device.UpdatedAt, &device.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return &device, nil
}

func (r *DeviceRepository) Update(ctx context.Context, device *models.Device) error {
	query := `
		UPDATE devices
		SET device_name = $2, tags = $3
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query, device.ID, device.DeviceName, device.Tags).Scan(&device.UpdatedAt)
}

func (r *DeviceRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE devices SET last_seen_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

func (r *DeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM devices WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Device Token operations

func (r *DeviceRepository) CreateToken(ctx context.Context, token *models.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (device_id, token, environment, bundle_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, issued_at, is_valid, error_count, updated_at
	`
	return r.db.QueryRow(ctx, query, token.DeviceID, token.Token, token.Environment, token.BundleID).
		Scan(&token.ID, &token.IssuedAt, &token.IsValid, &token.ErrorCount, &token.UpdatedAt)
}

func (r *DeviceRepository) GetTokenByDeviceID(ctx context.Context, deviceID uuid.UUID) (*models.DeviceToken, error) {
	var token models.DeviceToken
	query := `
		SELECT id, device_id, token, environment, bundle_id, issued_at, is_valid,
		       last_used_at, error_count, last_error, updated_at
		FROM device_tokens WHERE device_id = $1 AND is_valid = true
		ORDER BY issued_at DESC LIMIT 1
	`
	err := r.db.QueryRow(ctx, query, deviceID).Scan(
		&token.ID, &token.DeviceID, &token.Token, &token.Environment, &token.BundleID,
		&token.IssuedAt, &token.IsValid, &token.LastUsedAt, &token.ErrorCount,
		&token.LastError, &token.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get device token: %w", err)
	}
	return &token, nil
}

func (r *DeviceRepository) GetTokensByUserAndTags(ctx context.Context, userID uuid.UUID, tags []string) ([]models.DeviceToken, error) {
	query := `
		SELECT dt.id, dt.device_id, dt.token, dt.environment, dt.bundle_id,
		       dt.issued_at, dt.is_valid, dt.last_used_at, dt.error_count,
		       dt.last_error, dt.updated_at
		FROM device_tokens dt
		JOIN devices d ON dt.device_id = d.id
		WHERE d.user_id = $1 AND dt.is_valid = true
		  AND d.tags && $2::text[]
	`
	rows, err := r.db.Query(ctx, query, userID, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to query device tokens: %w", err)
	}
	defer rows.Close()

	return r.scanTokens(rows)
}

func (r *DeviceRepository) GetTokensByUserID(ctx context.Context, userID uuid.UUID) ([]models.DeviceToken, error) {
	query := `
		SELECT dt.id, dt.device_id, dt.token, dt.environment, dt.bundle_id,
		       dt.issued_at, dt.is_valid, dt.last_used_at, dt.error_count,
		       dt.last_error, dt.updated_at
		FROM device_tokens dt
		JOIN devices d ON dt.device_id = d.id
		WHERE d.user_id = $1 AND dt.is_valid = true
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device tokens: %w", err)
	}
	defer rows.Close()

	return r.scanTokens(rows)
}

func (r *DeviceRepository) scanTokens(rows pgx.Rows) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	for rows.Next() {
		var token models.DeviceToken
		if err := rows.Scan(
			&token.ID, &token.DeviceID, &token.Token, &token.Environment, &token.BundleID,
			&token.IssuedAt, &token.IsValid, &token.LastUsedAt, &token.ErrorCount,
			&token.LastError, &token.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device token: %w", err)
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (r *DeviceRepository) UpdateToken(ctx context.Context, deviceID uuid.UUID, tokenStr, environment, bundleID string) error {
	// First, invalidate old tokens for this device
	invalidateQuery := `UPDATE device_tokens SET is_valid = false WHERE device_id = $1`
	if _, err := r.db.Exec(ctx, invalidateQuery, deviceID); err != nil {
		return fmt.Errorf("failed to invalidate old tokens: %w", err)
	}

	// Create new token
	token := &models.DeviceToken{
		DeviceID:    deviceID,
		Token:       tokenStr,
		Environment: environment,
		BundleID:    bundleID,
	}
	return r.CreateToken(ctx, token)
}

func (r *DeviceRepository) MarkTokenInvalid(ctx context.Context, tokenID uuid.UUID, errorReason string) error {
	query := `
		UPDATE device_tokens
		SET is_valid = false, error_count = error_count + 1, last_error = $2
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, tokenID, errorReason)
	return err
}

func (r *DeviceRepository) UpdateTokenLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	query := `UPDATE device_tokens SET last_used_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID, time.Now())
	return err
}
