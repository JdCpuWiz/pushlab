package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pushlab/backend/internal/api"
	"github.com/pushlab/backend/internal/auth"
	"github.com/pushlab/backend/internal/config"
	"github.com/pushlab/backend/internal/db"
	"github.com/pushlab/backend/internal/queue"
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

	// Create publisher
	publisher := queue.NewPublisher(rmq)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiryHours, cfg.JWT.Issuer)

	// Create certs directory
	certsDir := os.Getenv("CERTS_DIR")
	if certsDir == "" {
		certsDir = "/var/lib/pushlab/certs"
	}
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		log.Fatalf("Failed to create certs directory: %v", err)
	}

	// Create API server
	server := api.NewServer(database, jwtService, publisher, certsDir)

	// HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.APIPort)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting API server on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
