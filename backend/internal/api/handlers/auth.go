package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pushlab/backend/internal/auth"
	"github.com/pushlab/backend/internal/models"
	"github.com/pushlab/backend/internal/repository"
	"github.com/pushlab/backend/internal/api/middleware"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtService *auth.JWTService
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	apiKey, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		APIKey:       apiKey,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = "" // Don't send password hash
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		http.Error(w, "User account is inactive", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	apiKey, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdateAPIKey(r.Context(), user.ID, apiKey); err != nil {
		http.Error(w, "Failed to update API key", http.StatusInternalServerError)
		return
	}

	response := models.APIKeyResponse{APIKey: apiKey}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
