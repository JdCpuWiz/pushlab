package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pushlab/backend/internal/models"
)

type APNsRepository struct {
	db *pgxpool.Pool
}

func NewAPNsRepository(db *pgxpool.Pool) *APNsRepository {
	return &APNsRepository{db: db}
}

func (r *APNsRepository) Create(ctx context.Context, cred *models.APNsCredential) error {
	query := `
		INSERT INTO apns_credentials (user_id, team_id, key_id, bundle_id, environment, private_key_path)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, is_active
	`
	return r.db.QueryRow(ctx, query,
		cred.UserID, cred.TeamID, cred.KeyID, cred.BundleID, cred.Environment, cred.PrivateKeyPath,
	).Scan(&cred.ID, &cred.CreatedAt, &cred.IsActive)
}

func (r *APNsRepository) GetByUserAndBundle(ctx context.Context, userID uuid.UUID, bundleID, environment string) (*models.APNsCredential, error) {
	var cred models.APNsCredential
	query := `
		SELECT id, user_id, team_id, key_id, bundle_id, environment, private_key_path, created_at, is_active
		FROM apns_credentials
		WHERE user_id = $1 AND bundle_id = $2 AND environment = $3 AND is_active = true
	`
	err := r.db.QueryRow(ctx, query, userID, bundleID, environment).Scan(
		&cred.ID, &cred.UserID, &cred.TeamID, &cred.KeyID, &cred.BundleID,
		&cred.Environment, &cred.PrivateKeyPath, &cred.CreatedAt, &cred.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get APNs credential: %w", err)
	}
	return &cred, nil
}

func (r *APNsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.APNsCredential, error) {
	query := `
		SELECT id, user_id, team_id, key_id, bundle_id, environment, private_key_path, created_at, is_active
		FROM apns_credentials
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query APNs credentials: %w", err)
	}
	defer rows.Close()

	var credentials []models.APNsCredential
	for rows.Next() {
		var cred models.APNsCredential
		if err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.TeamID, &cred.KeyID, &cred.BundleID,
			&cred.Environment, &cred.PrivateKeyPath, &cred.CreatedAt, &cred.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan APNs credential: %w", err)
		}
		credentials = append(credentials, cred)
	}

	return credentials, nil
}

func (r *APNsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE apns_credentials SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
