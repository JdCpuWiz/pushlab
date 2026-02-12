package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pushlab/backend/internal/db"
)

type HealthHandler struct {
	db *db.DB
}

func NewHealthHandler(database *db.DB) *HealthHandler {
	return &HealthHandler{db: database}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:   "ok",
		Database: "ok",
	}

	if err := h.db.Health(r.Context()); err != nil {
		response.Database = "error"
		response.Status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
