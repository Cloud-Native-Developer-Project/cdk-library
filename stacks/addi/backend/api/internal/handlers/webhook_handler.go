package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"addi-backend/internal/domain"
)

// WebhookHandler handles webhook requests from Lambda
type WebhookHandler struct {
	processor domain.WebhookProcessor
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(processor domain.WebhookProcessor) *WebhookHandler {
	return &WebhookHandler{
		processor: processor,
	}
}

// HandleWebhook processes incoming webhook requests
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Parse JSON payload
	var event domain.S3EventPayload
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate payload
	if err := validateS3Event(&event); err != nil {
		log.Printf("❌ Invalid event payload: %v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Process event
	response, err := h.processor.ProcessS3Event(r.Context(), &event)
	if err != nil {
		log.Printf("❌ Error processing event: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process event")
		return
	}

	// Send successful response
	respondWithJSON(w, http.StatusOK, response)
}

// validateS3Event validates the S3 event payload
func validateS3Event(event *domain.S3EventPayload) error {
	if event.EventID == "" {
		return &ValidationError{Field: "eventId", Message: "eventId is required"}
	}

	if event.Bucket == "" {
		return &ValidationError{Field: "bucket", Message: "bucket is required"}
	}

	if event.Key == "" {
		return &ValidationError{Field: "key", Message: "key is required"}
	}

	if event.Size <= 0 {
		return &ValidationError{Field: "size", Message: "size must be greater than 0"}
	}

	if event.PresignedURL == "" {
		return &ValidationError{Field: "presignedUrl", Message: "presignedUrl is required"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
