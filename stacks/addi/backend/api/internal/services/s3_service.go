package services

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"addi-backend/internal/domain"
)

// S3ServiceImpl implements the S3Service interface
type S3ServiceImpl struct {
	client *s3.Client
}

// NewS3Service creates a new S3 service instance
func NewS3Service(cfg aws.Config) domain.S3Service {
	return &S3ServiceImpl{
		client: s3.NewFromConfig(cfg),
	}
}

// DownloadFile downloads a file from S3 and returns a reader
func (s *S3ServiceImpl) DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3 %s/%s: %w", bucket, key, err)
	}

	return result.Body, nil
}

// DownloadFileFromPresignedURL downloads a file using a presigned URL (no AWS credentials needed)
func (s *S3ServiceImpl) DownloadFileFromPresignedURL(ctx context.Context, presignedURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, presignedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from presigned URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// GetFileMetadata retrieves metadata about an S3 object
func (s *S3ServiceImpl) GetFileMetadata(ctx context.Context, bucket, key string) (map[string]string, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata for S3 object %s/%s: %w", bucket, key, err)
	}

	metadata := make(map[string]string)
	if result.ContentType != nil {
		metadata["content-type"] = *result.ContentType
	}
	if result.ContentLength != nil {
		metadata["content-length"] = fmt.Sprintf("%d", *result.ContentLength)
	}
	if result.LastModified != nil {
		metadata["last-modified"] = result.LastModified.String()
	}

	// Include custom metadata
	for key, value := range result.Metadata {
		metadata[key] = value
	}

	return metadata, nil
}
