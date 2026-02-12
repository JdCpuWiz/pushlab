package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pushlab/backend/internal/apns"
	"github.com/pushlab/backend/internal/config"
	"github.com/pushlab/backend/internal/db"
	"github.com/pushlab/backend/internal/queue"
	"github.com/pushlab/backend/internal/worker"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	log.Println("Configuration loaded successfully")

	// Connect to database
	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database")

	// Connect to RabbitMQ
	rmq, err := queue.NewRabbitMQ(cfg.RabbitMQ.URL, cfg.RabbitMQ.QueueName, cfg.RabbitMQ.ReconnectDelay)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()

	log.Println("Connected to RabbitMQ")

	// Create APNs client
	apnsClient := apns.NewClient()
	defer apnsClient.Close()

	// Create processor
	processor := worker.NewProcessor(database.Pool, apnsClient)

	// Create consumer
	consumer := queue.NewConsumer(rmq, processor.ProcessNotification, cfg.RabbitMQ.PrefetchCount)

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	log.Printf("Worker started with %d concurrent workers", cfg.Server.WorkerCount)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	cancel()

	log.Println("Worker stopped")
}
