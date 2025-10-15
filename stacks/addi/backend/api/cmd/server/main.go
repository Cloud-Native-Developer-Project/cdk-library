package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"

	appConfig "addi-backend/internal/config"
	"addi-backend/internal/handlers"
	"addi-backend/internal/middleware"
	"addi-backend/internal/services"
)

func main() {
	log.Println("🚀 Starting Addi Backend API...")

	// Load configuration
	cfg, err := appConfig.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("📋 Configuration loaded:")
	log.Printf("   Server: %s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("   SFTP: %s:%d", cfg.SFTP.Host, cfg.SFTP.Port)
	log.Printf("   AWS Region: %s", cfg.AWS.Region)

	// Initialize AWS SDK
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Initialize services
	s3Service := services.NewS3Service(awsConfig)
	sftpService := services.NewSFTPService(&cfg.SFTP)
	webhookProcessor := services.NewWebhookProcessor(s3Service, sftpService)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(webhookProcessor)
	healthHandler := handlers.NewHealthHandler(sftpService)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/addi-csv", webhookHandler.HandleWebhook)
	mux.HandleFunc("/health", healthHandler.HandleHealth)

	// Apply middleware
	handler := middleware.Logger(middleware.Recovery(mux))

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🌐 Server listening on http://%s", server.Addr)
		log.Printf("📍 Endpoints:")
		log.Printf("   POST /webhook/addi-csv - Process CSV upload notifications")
		log.Printf("   GET  /health           - Health check")
		log.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server gracefully stopped")
}
