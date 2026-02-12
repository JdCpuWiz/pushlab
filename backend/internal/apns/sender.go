package apns

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sideshow/apns2"
)

type SendResult struct {
	DeviceTokenID uuid.UUID
	Success       bool
	StatusCode    int
	Reason        string
	Timestamp     time.Time
}

type Sender struct {
	client *Client
}

func NewSender(client *Client) *Sender {
	return &Sender{client: client}
}

func (s *Sender) Send(ctx context.Context, deviceToken, bundleID, environment, keyPath, keyID, teamID string, notif *apns2.Notification) (*SendResult, error) {
	apnsClient, err := s.client.GetClient(keyPath, keyID, teamID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get APNs client: %w", err)
	}

	notif.Topic = bundleID

	res, err := apnsClient.PushWithContext(ctx, notif)
	if err != nil {
		return nil, fmt.Errorf("failed to send push notification: %w", err)
	}

	result := &SendResult{
		StatusCode: res.StatusCode,
		Timestamp:  time.Now(),
	}

	if res.StatusCode == 200 {
		result.Success = true
		log.Printf("Successfully sent notification to device token: %s", deviceToken[:16]+"...")
	} else {
		result.Success = false
		result.Reason = res.Reason
		log.Printf("Failed to send notification: status=%d, reason=%s, apns_id=%s",
			res.StatusCode, res.Reason, res.ApnsID)
	}

	return result, nil
}

func (s *Sender) SendWithRetry(ctx context.Context, deviceToken, bundleID, environment, keyPath, keyID, teamID string, notif *apns2.Notification, maxRetries int) (*SendResult, error) {
	var lastErr error
	var result *SendResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt*attempt) * time.Second
			log.Printf("Retry attempt %d after %v delay", attempt, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, lastErr = s.Send(ctx, deviceToken, bundleID, environment, keyPath, keyID, teamID, notif)
		if lastErr == nil {
			if result.Success {
				return result, nil
			}

			// Don't retry on certain error codes
			if shouldNotRetry(result.StatusCode) {
				return result, nil
			}
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
	}

	return result, nil
}

func shouldNotRetry(statusCode int) bool {
	switch statusCode {
	case 400, // Bad request
		403,  // Invalid topic
		405,  // Bad method
		410,  // Device token inactive
		413:  // Payload too large
		return true
	default:
		return false
	}
}
