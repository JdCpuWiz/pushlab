package apns

import (
	"github.com/pushlab/backend/internal/models"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
)

func BuildNotification(deviceToken string, notif *models.NotificationPayload) *apns2.Notification {
	p := payload.NewPayload()

	if notif.Title != nil {
		p.AlertTitle(*notif.Title)
	}
	p.AlertBody(notif.Body)

	if notif.Badge != nil {
		p.Badge(*notif.Badge)
	}

	if notif.Sound != "" {
		p.Sound(notif.Sound)
	}

	if notif.Category != nil {
		p.Category(*notif.Category)
	}

	// Add custom data
	if notif.Data != nil {
		for key, value := range notif.Data {
			p.Custom(key, value)
		}
	}

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Payload:     p,
	}

	// Set priority
	if notif.Priority == "high" {
		notification.Priority = apns2.PriorityHigh
	} else {
		notification.Priority = apns2.PriorityLow
	}

	return notification
}
