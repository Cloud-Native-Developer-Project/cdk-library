package services

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"addi-backend/internal/domain"
)

// WebhookProcessorImpl implements the WebhookProcessor interface
type WebhookProcessorImpl struct {
	s3Service   domain.S3Service
	sftpService domain.SFTPService
}

// NewWebhookProcessor creates a new webhook processor instance
func NewWebhookProcessor(s3Service domain.S3Service, sftpService domain.SFTPService) domain.WebhookProcessor {
	return &WebhookProcessorImpl{
		s3Service:   s3Service,
		sftpService: sftpService,
	}
}

// ProcessS3Event processes an S3 event notification
func (w *WebhookProcessorImpl) ProcessS3Event(ctx context.Context, event *domain.S3EventPayload) (*domain.WebhookResponse, error) {
	log.Printf("📥 Processing S3 event:")
	log.Printf("   Event ID: %s", event.EventID)
	log.Printf("   Bucket: %s", event.Bucket)
	log.Printf("   Key: %s", event.Key)
	log.Printf("   Size: %d bytes", event.Size)
	log.Printf("   ETag: %s", event.ETag)
	log.Printf("   Timestamp: %s", event.Timestamp)
	log.Printf("   Presigned URL expires: %s", event.ExpiresAt)

	// Step 1: Download file from S3 using presigned URL (no AWS credentials needed)
	log.Printf("⬇️  Downloading file from S3 using presigned URL...")
	fileReader, err := w.s3Service.DownloadFileFromPresignedURL(ctx, event.PresignedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer fileReader.Close()

	// Step 2: Connect to SFTP server
	log.Printf("🔌 Connecting to SFTP server...")
	if err := w.sftpService.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to SFTP server: %w", err)
	}
	defer w.sftpService.Close()

	// Step 3: Upload file to SFTP
	log.Printf("⬆️  Uploading file to SFTP server...")
	timestamp, _ := time.Parse(time.RFC3339, event.Timestamp)
	remotePath := generateRemotePath(event.Key, timestamp)
	result, err := w.sftpService.UploadFile(ctx, fileReader, remotePath, event.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to SFTP: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("SFTP upload failed: %v", result.Error)
	}

	log.Printf("✅ File successfully transferred to SFTP:")
	log.Printf("   Remote Path: %s", result.RemotePath)
	log.Printf("   Bytes Transferred: %d", result.BytesTransferred)
	log.Printf("   Duration: %s", result.Duration)

	return &domain.WebhookResponse{
		Status:      "success",
		Message:     fmt.Sprintf("File transferred to SFTP successfully in %s", result.Duration),
		File:        event.Key,
		ProcessedAt: time.Now(),
	}, nil
}

// generateRemotePath generates a remote path with timestamp organization
// Example: uploads/2025/10/14/file.csv
func generateRemotePath(originalKey string, timestamp time.Time) string {
	fileName := filepath.Base(originalKey)
	datePrefix := timestamp.Format("2006/01/02")
	return filepath.Join(datePrefix, fileName)
}
