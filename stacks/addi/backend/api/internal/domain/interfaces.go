package domain

import (
	"context"
	"io"
)

// S3Service defines the contract for S3 operations
type S3Service interface {
	// DownloadFile downloads a file from S3 and returns a reader
	DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// DownloadFileFromPresignedURL downloads a file using a presigned URL (no AWS credentials needed)
	DownloadFileFromPresignedURL(ctx context.Context, presignedURL string) (io.ReadCloser, error)

	// GetFileMetadata retrieves metadata about an S3 object
	GetFileMetadata(ctx context.Context, bucket, key string) (map[string]string, error)
}

// SFTPService defines the contract for SFTP operations
type SFTPService interface {
	// Connect establishes connection to SFTP server
	Connect(ctx context.Context) error

	// UploadFile uploads a file to SFTP server
	UploadFile(ctx context.Context, reader io.Reader, remotePath string, size int64) (*SFTPTransferResult, error)

	// Close closes the SFTP connection
	Close() error

	// HealthCheck verifies SFTP connection is alive
	HealthCheck(ctx context.Context) error
}

// WebhookProcessor defines the contract for processing webhook events
type WebhookProcessor interface {
	// ProcessS3Event processes an S3 event notification
	ProcessS3Event(ctx context.Context, event *S3EventPayload) (*WebhookResponse, error)
}
