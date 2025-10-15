package domain

import "time"

// S3EventPayload represents the webhook payload from Lambda
// This matches the WebhookPayload structure from the Lambda function
type S3EventPayload struct {
	EventID      string `json:"eventId"`
	Timestamp    string `json:"timestamp"`
	Bucket       string `json:"bucket"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ETag         string `json:"etag"`
	PresignedURL string `json:"presignedUrl"`
	ExpiresAt    string `json:"expiresAt"`
}

// WebhookResponse represents the API response after processing webhook
type WebhookResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	File      string    `json:"file"`
	ProcessedAt time.Time `json:"processed_at"`
}

// SFTPTransferResult represents the result of an SFTP transfer operation
type SFTPTransferResult struct {
	Success       bool
	RemotePath    string
	BytesTransferred int64
	Duration      time.Duration
	Error         error
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}
