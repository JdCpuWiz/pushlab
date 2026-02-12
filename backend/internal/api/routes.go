package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/pushlab/backend/internal/api/handlers"
	"github.com/pushlab/backend/internal/api/middleware"
	"github.com/pushlab/backend/internal/auth"
	"github.com/pushlab/backend/internal/db"
	"github.com/pushlab/backend/internal/queue"
	"github.com/pushlab/backend/internal/repository"
)

type Server struct {
	router         *chi.Mux
	authHandler    *handlers.AuthHandler
	deviceHandler  *handlers.DeviceHandler
	notifHandler   *handlers.NotificationHandler
	apnsHandler    *handlers.APNsHandler
	healthHandler  *handlers.HealthHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewServer(database *db.DB, jwtService *auth.JWTService, publisher *queue.Publisher, certsDir string) *Server {
	userRepo := repository.NewUserRepository(database.Pool)
	deviceRepo := repository.NewDeviceRepository(database.Pool)
	notifRepo := repository.NewNotificationRepository(database.Pool)
	apnsRepo := repository.NewAPNsRepository(database.Pool)

	s := &Server{
		router:         chi.NewRouter(),
		authHandler:    handlers.NewAuthHandler(userRepo, jwtService),
		deviceHandler:  handlers.NewDeviceHandler(deviceRepo),
		notifHandler:   handlers.NewNotificationHandler(notifRepo, deviceRepo, publisher),
		apnsHandler:    handlers.NewAPNsHandler(apnsRepo, certsDir),
		healthHandler:  handlers.NewHealthHandler(database),
		authMiddleware: middleware.NewAuthMiddleware(jwtService, userRepo),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Middleware
	s.router.Use(middleware.Logging)
	s.router.Use(middleware.CORS)

	// Health check
	s.router.Get("/health", s.healthHandler.Check)

	// Public routes
	s.router.Post("/api/v1/auth/register", s.authHandler.Register)
	s.router.Post("/api/v1/auth/login", s.authHandler.Login)

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(s.authMiddleware.Authenticate)

		// Auth
		r.Get("/api/v1/auth/apikey", s.authHandler.GenerateAPIKey)

		// Devices
		r.Post("/api/v1/devices", s.deviceHandler.Register)
		r.Get("/api/v1/devices", s.deviceHandler.List)
		r.Get("/api/v1/devices/{id}", s.deviceHandler.Get)
		r.Put("/api/v1/devices/{id}", s.deviceHandler.Update)
		r.Delete("/api/v1/devices/{id}", s.deviceHandler.Delete)
		r.Put("/api/v1/devices/{id}/token", s.deviceHandler.UpdateToken)

		// Notifications
		r.Post("/api/v1/notify", s.notifHandler.Send)
		r.Post("/api/v1/notify/device/{device_id}", s.notifHandler.SendToDevice)
		r.Get("/api/v1/notifications", s.notifHandler.List)
		r.Get("/api/v1/notifications/{id}", s.notifHandler.Get)

		// APNs Credentials
		r.Post("/api/v1/credentials/apns", s.apnsHandler.Create)
		r.Get("/api/v1/credentials/apns", s.apnsHandler.List)
		r.Delete("/api/v1/credentials/apns/{id}", s.apnsHandler.Delete)
	})
}

func (s *Server) Router() *chi.Mux {
	return s.router
}
