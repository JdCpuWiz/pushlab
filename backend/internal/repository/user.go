package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pushlab/backend/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, api_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, is_active
	`
	return r.db.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.IsActive)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, api_key, created_at, updated_at, is_active
		FROM users WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.APIKey, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, api_key, created_at, updated_at, is_active
		FROM users WHERE username = $1
	`
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.APIKey, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, api_key, created_at, updated_at, is_active
		FROM users WHERE api_key = $1 AND is_active = true
	`
	err := r.db.QueryRow(ctx, query, apiKey).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.APIKey, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $2, is_active = $3
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query, user.ID, user.Email, user.IsActive).Scan(&user.UpdatedAt)
}

func (r *UserRepository) UpdateAPIKey(ctx context.Context, userID uuid.UUID, apiKey string) error {
	query := `UPDATE users SET api_key = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, apiKey)
	return err
}
