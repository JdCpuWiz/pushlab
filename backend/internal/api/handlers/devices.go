package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pushlab/backend/internal/api/middleware"
	"github.com/pushlab/backend/internal/models"
	"github.com/pushlab/backend/internal/repository"
)

type DeviceHandler struct {
	deviceRepo *repository.DeviceRepository
}

func NewDeviceHandler(deviceRepo *repository.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{deviceRepo: deviceRepo}
}

func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	var req models.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceName == "" || req.DeviceIdentifier == "" || req.DeviceToken == "" || req.BundleID == "" {
		http.Error(w, "Device name, identifier, token, and bundle ID are required", http.StatusBadRequest)
		return
	}

	if req.Environment == "" {
		req.Environment = "production"
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	// Check if device already exists
	existingDevice, err := h.deviceRepo.GetByUserAndIdentifier(r.Context(), user.ID, req.DeviceIdentifier)
	if err == nil && existingDevice != nil {
		// Update existing device
		existingDevice.DeviceName = req.DeviceName
		existingDevice.Tags = req.Tags

		if err := h.deviceRepo.Update(r.Context(), existingDevice); err != nil {
			http.Error(w, "Failed to update device", http.StatusInternalServerError)
			return
		}

		// Update device token
		if err := h.deviceRepo.UpdateToken(r.Context(), existingDevice.ID, req.DeviceToken, req.Environment, req.BundleID); err != nil {
			http.Error(w, "Failed to update device token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existingDevice)
		return
	}

	// Create new device
	device := &models.Device{
		UserID:           user.ID,
		DeviceName:       req.DeviceName,
		DeviceIdentifier: req.DeviceIdentifier,
		Tags:             req.Tags,
	}

	if err := h.deviceRepo.Create(r.Context(), device); err != nil {
		http.Error(w, "Failed to create device", http.StatusInternalServerError)
		return
	}

	// Create device token
	token := &models.DeviceToken{
		DeviceID:    device.ID,
		Token:       req.DeviceToken,
		Environment: req.Environment,
		BundleID:    req.BundleID,
		IsValid:     true,
	}

	if err := h.deviceRepo.CreateToken(r.Context(), token); err != nil {
		http.Error(w, "Failed to create device token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	devices, err := h.deviceRepo.GetByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func (h *DeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := h.deviceRepo.GetByID(r.Context(), deviceID)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	if device.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := h.deviceRepo.GetByID(r.Context(), deviceID)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	if device.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceName != "" {
		device.DeviceName = req.DeviceName
	}
	if req.Tags != nil {
		device.Tags = req.Tags
	}

	if err := h.deviceRepo.Update(r.Context(), device); err != nil {
		http.Error(w, "Failed to update device", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := h.deviceRepo.GetByID(r.Context(), deviceID)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	if device.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.deviceRepo.Delete(r.Context(), deviceID); err != nil {
		http.Error(w, "Failed to delete device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeviceHandler) UpdateToken(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := h.deviceRepo.GetByID(r.Context(), deviceID)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	if device.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.UpdateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceToken == "" || req.Environment == "" || req.BundleID == "" {
		http.Error(w, "Device token, environment, and bundle ID are required", http.StatusBadRequest)
		return
	}

	if err := h.deviceRepo.UpdateToken(r.Context(), deviceID, req.DeviceToken, req.Environment, req.BundleID); err != nil {
		http.Error(w, "Failed to update token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Token updated successfully"}`))
}
