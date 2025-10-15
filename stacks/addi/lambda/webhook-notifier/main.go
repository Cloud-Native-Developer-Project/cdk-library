package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// WebhookCredentials stores the webhook configuration from Secrets Manager
type WebhookCredentials struct {
	WebhookURL string `json:"webhookUrl"`
	APIKey     string `json:"apiKey"`
	HMACSecret string `json:"hmacSecret"`
}

// WebhookPayload is the JSON structure sent to the on-premise webhook
type WebhookPayload struct {
	EventID      string `json:"eventId"`
	Timestamp    string `json:"timestamp"`
	Bucket       string `json:"bucket"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ETag         string `json:"etag"`
	PresignedURL string `json:"presignedUrl"`
	ExpiresAt    string `json:"expiresAt"`
}

// Config holds Lambda configuration from environment variables
type Config struct {
	BucketName           string
	WebhookSecretARN     string
	WebhookURLOverride   string // For local development with ngrok
	PresignedURLExpires  int
	MaxRetryAttempts     int
	RetryExponentialBase int
}

var (
	cfg         Config
	s3Client    *s3.Client
	smClient    *secretsmanager.Client
	httpClient  *http.Client
	credentials *WebhookCredentials
)

// init initializes AWS clients and loads configuration (runs once per Lambda container)
func init() {
	// Load configuration from environment variables
	cfg = Config{
		BucketName:           getEnv("BUCKET_NAME", ""),
		WebhookSecretARN:     getEnv("WEBHOOK_SECRET_ARN", ""),
		WebhookURLOverride:   getEnv("WEBHOOK_URL_OVERRIDE", ""), // For local dev with ngrok
		PresignedURLExpires:  getEnvInt("PRESIGNED_URL_EXPIRES", 900),  // 15 minutes
		MaxRetryAttempts:     getEnvInt("MAX_RETRY_ATTEMPTS", 4),
		RetryExponentialBase: getEnvInt("RETRY_EXPONENTIAL_BASE", 2),
	}

	// Validate required configuration
	if cfg.BucketName == "" {
		log.Fatal("BUCKET_NAME environment variable is required")
	}

	// WebhookSecretARN is required only if override is not provided (production mode)
	if cfg.WebhookSecretARN == "" && cfg.WebhookURLOverride == "" {
		log.Fatal("Either WEBHOOK_SECRET_ARN or WEBHOOK_URL_OVERRIDE is required")
	}

	// Initialize AWS SDK clients
	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	s3Client = s3.NewFromConfig(awsConfig)
	smClient = secretsmanager.NewFromConfig(awsConfig)

	// Initialize HTTP client with timeout
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Load webhook credentials
	// In development mode (WEBHOOK_URL_OVERRIDE set), use override and skip Secrets Manager
	if cfg.WebhookURLOverride != "" {
		log.Printf("⚠️  DEVELOPMENT MODE: Using WEBHOOK_URL_OVERRIDE=%s", cfg.WebhookURLOverride)
		credentials = &WebhookCredentials{
			WebhookURL: cfg.WebhookURLOverride,
			APIKey:     "dev-mode", // Not validated in dev mode
			HMACSecret: "dev-mode", // Not validated in dev mode
		}
	} else {
		// Production mode: Load credentials from Secrets Manager (cached in Lambda container)
		credentials, err = loadWebhookCredentials(ctx)
		if err != nil {
			log.Fatalf("Failed to load webhook credentials: %v", err)
		}
		log.Println("✅ Loaded credentials from Secrets Manager")
	}

	log.Println("Lambda initialized successfully")
}

// handler processes S3 events from EventBridge and sends webhooks to on-premise server
func handler(ctx context.Context, event events.CloudWatchEvent) error {
	log.Printf("Processing event: %s", event.ID)

	// Parse S3 event detail from EventBridge
	var detail map[string]interface{}
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		return fmt.Errorf("failed to parse event detail: %w", err)
	}

	// Extract S3 object metadata
	bucket := detail["bucket"].(map[string]interface{})["name"].(string)
	object := detail["object"].(map[string]interface{})
	key := object["key"].(string)
	size := int64(object["size"].(float64))
	etag := object["etag"].(string)

	log.Printf("S3 object: s3://%s/%s (size: %d bytes, etag: %s)", bucket, key, size, etag)

	// Generate Presigned URL for the S3 object
	presignedURL, expiresAt, err := generatePresignedURL(ctx, bucket, key)
	if err != nil {
		return fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	log.Printf("Generated presigned URL (expires at: %s)", expiresAt)

	// Build webhook payload
	payload := WebhookPayload{
		EventID:      event.ID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Bucket:       bucket,
		Key:          key,
		Size:         size,
		ETag:         etag,
		PresignedURL: presignedURL,
		ExpiresAt:    expiresAt,
	}

	// Send webhook with retry logic
	if err := sendWebhookWithRetry(ctx, payload); err != nil {
		return fmt.Errorf("failed to send webhook after retries: %w", err)
	}

	log.Printf("Successfully processed event: %s", event.ID)
	return nil
}

// loadWebhookCredentials retrieves webhook credentials from Secrets Manager
func loadWebhookCredentials(ctx context.Context) (*WebhookCredentials, error) {
	result, err := smClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &cfg.WebhookSecretARN,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	var creds WebhookCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// generatePresignedURL creates a temporary signed URL for S3 object access
func generatePresignedURL(ctx context.Context, bucket, key string) (string, string, error) {
	presignClient := s3.NewPresignClient(s3Client)

	expiresIn := time.Duration(cfg.PresignedURLExpires) * time.Second
	expiresAt := time.Now().UTC().Add(expiresIn).Format(time.RFC3339)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", "", err
	}

	return presignResult.URL, expiresAt, nil
}

// sendWebhookWithRetry sends webhook with exponential backoff retry logic
func sendWebhookWithRetry(ctx context.Context, payload WebhookPayload) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt seconds (2s, 4s, 8s, 16s)
			backoff := time.Duration(math.Pow(float64(cfg.RetryExponentialBase), float64(attempt))) * time.Second
			log.Printf("Retry attempt %d/%d after %v", attempt, cfg.MaxRetryAttempts, backoff)
			time.Sleep(backoff)
		}

		err := sendWebhook(ctx, payload)
		if err == nil {
			if attempt > 0 {
				log.Printf("Webhook succeeded after %d retries", attempt)
			}
			return nil
		}

		lastErr = err
		log.Printf("Webhook attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("all retry attempts exhausted: %w", lastErr)
}

// sendWebhook sends a single webhook request to on-premise server
func sendWebhook(ctx context.Context, payload WebhookPayload) error {
	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Calculate HMAC signature
	signature := calculateHMAC(payloadJSON, credentials.HMACSecret)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", credentials.WebhookURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", credentials.APIKey)
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("User-Agent", "AWS-Lambda-Webhook-Notifier/1.0")

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	log.Printf("Webhook sent successfully (status: %d)", resp.StatusCode)
	return nil
}

// calculateHMAC computes HMAC-SHA256 signature for payload
func calculateHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// getEnv retrieves environment variable with fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves environment variable as integer with fallback default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func main() {
	lambda.Start(handler)
}
