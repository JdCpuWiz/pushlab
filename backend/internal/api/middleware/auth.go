package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pushlab/backend/internal/auth"
	"github.com/pushlab/backend/internal/repository"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	jwtService *auth.JWTService
	userRepo   *repository.UserRepository
}

func NewAuthMiddleware(jwtService *auth.JWTService, userRepo *repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for API key first
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			user, err := m.userRepo.GetByAPIKey(r.Context(), apiKey)
			if err == nil && user != nil {
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Check for JWT token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtService.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		user, err := m.userRepo.GetByID(r.Context(), claims.UserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		if !user.IsActive {
			http.Error(w, "User account is inactive", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
