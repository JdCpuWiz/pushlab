package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pushlab/backend/internal/apns"
	"github.com/pushlab/backend/internal/models"
	"github.com/pushlab/backend/internal/repository"
)

type Processor struct {
	db             *pgxpool.Pool
	apnsClient     *apns.Client
	sender         *apns.Sender
	notifRepo      *repository.NotificationRepository
	deviceRepo     *repository.DeviceRepository
	apnsRepo       *repository.APNsRepository
}

func NewProcessor(
	db *pgxpool.Pool,
	apnsClient *apns.Client,
) *Processor {
	return &Processor{
		db:         db,
		apnsClient: apnsClient,
		sender:     apns.NewSender(apnsClient),
		notifRepo:  repository.NewNotificationRepository(db),
		deviceRepo: repository.NewDeviceRepository(db),
		apnsRepo:   repository.NewAPNsRepository(db),
	}
}

func (p *Processor) ProcessNotification(ctx context.Context, job *models.NotificationJob) error {
	log.Printf("Processing notification %s for user %s with %d device tokens",
		job.NotificationID, job.UserID, len(job.DeviceTokenIDs))

	// Update notification status to 'sent'
	if err := p.notifRepo.UpdateStatus(ctx, job.NotificationID, "sent"); err != nil {
		log.Printf("Failed to update notification status: %v", err)
	}

	successCount := 0
	failureCount := 0

	for _, tokenID := range job.DeviceTokenIDs {
		if err := p.processDeviceToken(ctx, job, tokenID); err != nil {
			log.Printf("Failed to process device token %s: %v", tokenID, err)
			failureCount++
		} else {
			successCount++
		}
	}

	// Update final status
	finalStatus := "delivered"
	if failureCount > 0 && successCount == 0 {
		finalStatus = "failed"
	}

	if err := p.notifRepo.UpdateStatus(ctx, job.NotificationID, finalStatus); err != nil {
		log.Printf("Failed to update final notification status: %v", err)
	}

	log.Printf("Notification %s processed: %d succeeded, %d failed",
		job.NotificationID, successCount, failureCount)

	return nil
}

func (p *Processor) processDeviceToken(ctx context.Context, job *models.NotificationJob, tokenID uuid.UUID) error {
	// Create delivery record
	delivery := &models.NotificationDelivery{
		NotificationID: job.NotificationID,
		DeviceTokenID:  tokenID,
		DeliveryStatus: "pending",
		AttemptCount:   0,
	}

	if err := p.notifRepo.CreateDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to create delivery record: %w", err)
	}

	// Get device token details
	deviceToken, err := p.getDeviceTokenByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get device token: %w", err)
	}

	if !deviceToken.IsValid {
		delivery.DeliveryStatus = "failed"
		delivery.APNsErrorReason = strPtr("Device token is invalid")
		p.notifRepo.UpdateDeliveryStatus(ctx, delivery)
		return fmt.Errorf("device token is invalid")
	}

	// Get device to find user and get APNs credentials
	device, err := p.deviceRepo.GetByID(ctx, deviceToken.DeviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	// Get APNs credentials
	cred, err := p.apnsRepo.GetByUserAndBundle(ctx, device.UserID, deviceToken.BundleID, deviceToken.Environment)
	if err != nil {
		delivery.DeliveryStatus = "failed"
		delivery.APNsErrorReason = strPtr("APNs credentials not found")
		p.notifRepo.UpdateDeliveryStatus(ctx, delivery)
		return fmt.Errorf("failed to get APNs credentials: %w", err)
	}

	// Build APNs notification
	notification := apns.BuildNotification(deviceToken.Token, &job.Payload)

	// Send notification with retries
	delivery.AttemptCount = 1
	result, err := p.sender.SendWithRetry(
		ctx,
		deviceToken.Token,
		deviceToken.BundleID,
		deviceToken.Environment,
		cred.PrivateKeyPath,
		cred.KeyID,
		cred.TeamID,
		notification,
		3, // max retries
	)

	if err != nil {
		delivery.DeliveryStatus = "failed"
		delivery.APNsErrorReason = strPtr(err.Error())
		p.notifRepo.UpdateDeliveryStatus(ctx, delivery)
		return err
	}

	// Update delivery status based on result
	delivery.APNsResponseCode = &result.StatusCode
	delivery.AttemptCount = 3 // Assuming max retries were used

	if result.Success {
		delivery.DeliveryStatus = "delivered"
		now := time.Now()
		delivery.DeliveredAt = &now
		p.deviceRepo.UpdateTokenLastUsed(ctx, tokenID)
	} else {
		delivery.DeliveryStatus = "failed"
		delivery.APNsErrorReason = &result.Reason

		// Handle invalid token (410 status)
		if result.StatusCode == 410 {
			p.deviceRepo.MarkTokenInvalid(ctx, tokenID, result.Reason)
		}
	}

	if err := p.notifRepo.UpdateDeliveryStatus(ctx, delivery); err != nil {
		log.Printf("Failed to update delivery status: %v", err)
	}

	return nil
}

func (p *Processor) getDeviceTokenByID(ctx context.Context, tokenID uuid.UUID) (*models.DeviceToken, error) {
	var token models.DeviceToken
	query := `
		SELECT id, device_id, token, environment, bundle_id, issued_at, is_valid,
		       last_used_at, error_count, last_error, updated_at
		FROM device_tokens WHERE id = $1
	`
	err := p.db.QueryRow(ctx, query, tokenID).Scan(
		&token.ID, &token.DeviceID, &token.Token, &token.Environment, &token.BundleID,
		&token.IssuedAt, &token.IsValid, &token.LastUsedAt, &token.ErrorCount,
		&token.LastError, &token.UpdatedAt,
	)
	return &token, err
}

func strPtr(s string) *string {
	return &s
}
