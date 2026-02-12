package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pushlab/backend/internal/api/middleware"
	"github.com/pushlab/backend/internal/models"
	"github.com/pushlab/backend/internal/repository"
)

type APNsHandler struct {
	apnsRepo *repository.APNsRepository
	certsDir string
}

func NewAPNsHandler(apnsRepo *repository.APNsRepository, certsDir string) *APNsHandler {
	// Create certs directory if it doesn't exist
	os.MkdirAll(certsDir, 0700)
	return &APNsHandler{
		apnsRepo: apnsRepo,
		certsDir: certsDir,
	}
}

func (h *APNsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	var req models.CreateAPNsCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TeamID == "" || req.KeyID == "" || req.BundleID == "" || req.Environment == "" || req.PrivateKey == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if req.Environment != "sandbox" && req.Environment != "production" {
		http.Error(w, "Environment must be 'sandbox' or 'production'", http.StatusBadRequest)
		return
	}

	// Save private key to file
	filename := fmt.Sprintf("%s_%s_%s_%s.p8", user.ID, req.BundleID, req.Environment, req.KeyID)
	keyPath := filepath.Join(h.certsDir, filename)

	if err := os.WriteFile(keyPath, []byte(req.PrivateKey), 0600); err != nil {
		http.Error(w, "Failed to save private key", http.StatusInternalServerError)
		return
	}

	cred := &models.APNsCredential{
		UserID:         user.ID,
		TeamID:         req.TeamID,
		KeyID:          req.KeyID,
		BundleID:       req.BundleID,
		Environment:    req.Environment,
		PrivateKeyPath: keyPath,
		IsActive:       true,
	}

	if err := h.apnsRepo.Create(r.Context(), cred); err != nil {
		// Clean up file on error
		os.Remove(keyPath)
		http.Error(w, "Failed to create credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cred)
}

func (h *APNsHandler) List(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	credentials, err := h.apnsRepo.GetByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch credentials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credentials)
}

func (h *APNsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	credID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	// Get credential to verify ownership
	credentials, err := h.apnsRepo.GetByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch credentials", http.StatusInternalServerError)
		return
	}

	found := false
	var keyPath string
	for _, cred := range credentials {
		if cred.ID == credID {
			found = true
			keyPath = cred.PrivateKeyPath
			break
		}
	}

	if !found {
		http.Error(w, "Credential not found", http.StatusNotFound)
		return
	}

	if err := h.apnsRepo.Delete(r.Context(), credID); err != nil {
		http.Error(w, "Failed to delete credential", http.StatusInternalServerError)
		return
	}

	// Delete the private key file
	if keyPath != "" {
		os.Remove(keyPath)
	}

	w.WriteHeader(http.StatusNoContent)
}
