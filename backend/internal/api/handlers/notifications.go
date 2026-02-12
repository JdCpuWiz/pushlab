package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pushlab/backend/internal/api/middleware"
	"github.com/pushlab/backend/internal/models"
	"github.com/pushlab/backend/internal/queue"
	"github.com/pushlab/backend/internal/repository"
)

type NotificationHandler struct {
	notifRepo  *repository.NotificationRepository
	deviceRepo *repository.DeviceRepository
	publisher  *queue.Publisher
}

func NewNotificationHandler(
	notifRepo *repository.NotificationRepository,
	deviceRepo *repository.DeviceRepository,
	publisher *queue.Publisher,
) *NotificationHandler {
	return &NotificationHandler{
		notifRepo:  notifRepo,
		deviceRepo: deviceRepo,
		publisher:  publisher,
	}
}

func (h *NotificationHandler) Send(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	var req models.SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Body == "" {
		http.Error(w, "Body is required", http.StatusBadRequest)
		return
	}

	if req.Priority == "" {
		req.Priority = "normal"
	}

	if req.Sound == "" {
		req.Sound = "default"
	}

	// Create notification record
	dataJSON, _ := json.Marshal(req.Data)
	notification := &models.Notification{
		UserID:   user.ID,
		Title:    req.Title,
		Body:     req.Body,
		Data:     dataJSON,
		Badge:    req.Badge,
		Sound:    req.Sound,
		Category: req.Category,
		Priority: req.Priority,
		Tags:     req.Tags,
		Status:   "queued",
	}

	if err := h.notifRepo.Create(r.Context(), notification); err != nil {
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	// Get device tokens to send to
	var deviceTokens []models.DeviceToken
	var err error

	if len(req.Tags) > 0 {
		deviceTokens, err = h.deviceRepo.GetTokensByUserAndTags(r.Context(), user.ID, req.Tags)
	} else {
		deviceTokens, err = h.deviceRepo.GetTokensByUserID(r.Context(), user.ID)
	}

	if err != nil {
		http.Error(w, "Failed to get device tokens", http.StatusInternalServerError)
		return
	}

	if len(deviceTokens) == 0 {
		http.Error(w, "No valid devices found", http.StatusNotFound)
		return
	}

	// Extract device token IDs
	tokenIDs := make([]uuid.UUID, len(deviceTokens))
	for i, token := range deviceTokens {
		tokenIDs[i] = token.ID
	}

	// Create notification job
	job := &models.NotificationJob{
		NotificationID: notification.ID,
		UserID:         user.ID,
		DeviceTokenIDs: tokenIDs,
		Payload: models.NotificationPayload{
			Title:    req.Title,
			Body:     req.Body,
			Badge:    req.Badge,
			Sound:    req.Sound,
			Category: req.Category,
			Priority: req.Priority,
			Data:     req.Data,
		},
	}

	// Publish to queue
	if err := h.publisher.PublishNotification(r.Context(), job); err != nil {
		http.Error(w, "Failed to queue notification", http.StatusInternalServerError)
		return
	}

	response := models.SendNotificationResponse{
		NotificationID: notification.ID,
		TargetDevices:  len(deviceTokens),
		Status:         "queued",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *NotificationHandler) SendToDevice(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	deviceID, err := uuid.Parse(chi.URLParam(r, "device_id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req models.SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Body == "" {
		http.Error(w, "Body is required", http.StatusBadRequest)
		return
	}

	// Verify device belongs to user
	device, err := h.deviceRepo.GetByID(r.Context(), deviceID)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	if device.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get device token
	deviceToken, err := h.deviceRepo.GetTokenByDeviceID(r.Context(), deviceID)
	if err != nil {
		http.Error(w, "Device token not found", http.StatusNotFound)
		return
	}

	if req.Priority == "" {
		req.Priority = "normal"
	}

	if req.Sound == "" {
		req.Sound = "default"
	}

	// Create notification record
	dataJSON, _ := json.Marshal(req.Data)
	notification := &models.Notification{
		UserID:   user.ID,
		Title:    req.Title,
		Body:     req.Body,
		Data:     dataJSON,
		Badge:    req.Badge,
		Sound:    req.Sound,
		Category: req.Category,
		Priority: req.Priority,
		Status:   "queued",
	}

	if err := h.notifRepo.Create(r.Context(), notification); err != nil {
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	// Create notification job
	job := &models.NotificationJob{
		NotificationID: notification.ID,
		UserID:         user.ID,
		DeviceTokenIDs: []uuid.UUID{deviceToken.ID},
		Payload: models.NotificationPayload{
			Title:    req.Title,
			Body:     req.Body,
			Badge:    req.Badge,
			Sound:    req.Sound,
			Category: req.Category,
			Priority: req.Priority,
			Data:     req.Data,
		},
	}

	// Publish to queue
	if err := h.publisher.PublishNotification(r.Context(), job); err != nil {
		http.Error(w, "Failed to queue notification", http.StatusInternalServerError)
		return
	}

	response := models.SendNotificationResponse{
		NotificationID: notification.ID,
		TargetDevices:  1,
		Status:         "queued",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.notifRepo.GetByUserID(r.Context(), user.ID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func (h *NotificationHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	notifID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	notification, err := h.notifRepo.GetByID(r.Context(), notifID)
	if err != nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	if notification.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	deliveries, err := h.notifRepo.GetDeliveriesByNotificationID(r.Context(), notifID)
	if err != nil {
		http.Error(w, "Failed to fetch deliveries", http.StatusInternalServerError)
		return
	}

	response := models.NotificationDetail{
		Notification: *notification,
		Deliveries:   deliveries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
