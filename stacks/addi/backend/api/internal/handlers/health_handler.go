package handlers

import (
	"net/http"
	"time"

	"addi-backend/internal/domain"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	sftpService domain.SFTPService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(sftpService domain.SFTPService) *HealthHandler {
	return &HealthHandler{
		sftpService: sftpService,
	}
}

// HandleHealth returns the health status of the application
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := &domain.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
	}

	// Check SFTP connection (optional - may not be connected yet)
	if err := h.sftpService.HealthCheck(r.Context()); err != nil {
		status.Services["sftp"] = "disconnected"
	} else {
		status.Services["sftp"] = "connected"
	}

	status.Services["api"] = "running"

	respondWithJSON(w, http.StatusOK, status)
}
