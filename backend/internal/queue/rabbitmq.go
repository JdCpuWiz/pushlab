package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	url          string
	queueName    string
	reconnectDelay time.Duration
}

func NewRabbitMQ(url, queueName string, reconnectDelay time.Duration) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		url:          url,
		queueName:    queueName,
		reconnectDelay: reconnectDelay,
	}

	if err := rmq.connect(); err != nil {
		return nil, err
	}

	return rmq, nil
}

func (r *RabbitMQ) connect() error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare main queue
	_, err = ch.QueueDeclare(
		r.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "",
			"x-dead-letter-routing-key": r.queueName + ".dlq",
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Declare dead letter queue
	_, err = ch.QueueDeclare(
		r.queueName+".dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	r.conn = conn
	r.channel = ch

	// Setup reconnection handlers
	go r.handleReconnect()

	return nil
}

func (r *RabbitMQ) handleReconnect() {
	for {
		reason, ok := <-r.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			log.Println("RabbitMQ connection closed")
			return
		}

		log.Printf("RabbitMQ connection closed: %v. Reconnecting...", reason)

		for {
			time.Sleep(r.reconnectDelay)
			if err := r.connect(); err != nil {
				log.Printf("Failed to reconnect: %v", err)
				continue
			}
			log.Println("Reconnected to RabbitMQ")
			break
		}
	}
}

func (r *RabbitMQ) Channel() *amqp.Channel {
	return r.channel
}

func (r *RabbitMQ) QueueName() string {
	return r.queueName
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, body []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		"",           // exchange
		r.queueName,  // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
}
