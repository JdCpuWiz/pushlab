package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/pushlab/backend/internal/models"
)

type MessageHandler func(context.Context, *models.NotificationJob) error

type Consumer struct {
	rmq           *RabbitMQ
	handler       MessageHandler
	prefetchCount int
}

func NewConsumer(rmq *RabbitMQ, handler MessageHandler, prefetchCount int) *Consumer {
	return &Consumer{
		rmq:           rmq,
		handler:       handler,
		prefetchCount: prefetchCount,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	ch := c.rmq.Channel()

	if err := ch.Qos(c.prefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		c.rmq.QueueName(),
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Consumer started, waiting for messages...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Consumer shutting down...")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}
				c.handleMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var job models.NotificationJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // Send to DLQ
		return
	}

	if err := c.handler(ctx, &job); err != nil {
		log.Printf("Failed to process notification %s: %v", job.NotificationID, err)

		// Check if we should retry
		if msg.Headers == nil {
			msg.Headers = amqp.Table{}
		}

		retryCount, _ := msg.Headers["x-retry-count"].(int32)
		if retryCount < 3 {
			msg.Headers["x-retry-count"] = retryCount + 1
			msg.Nack(false, true) // Requeue
		} else {
			log.Printf("Max retries exceeded for notification %s", job.NotificationID)
			msg.Nack(false, false) // Send to DLQ
		}
		return
	}

	msg.Ack(false)
}
