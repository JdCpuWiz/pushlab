package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pushlab/backend/internal/models"
)

type Publisher struct {
	rmq *RabbitMQ
}

func NewPublisher(rmq *RabbitMQ) *Publisher {
	return &Publisher{rmq: rmq}
}

func (p *Publisher) PublishNotification(ctx context.Context, job *models.NotificationJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal notification job: %w", err)
	}

	if err := p.rmq.Publish(ctx, body); err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	return nil
}
