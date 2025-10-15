package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Server ServerConfig
	SFTP   SFTPConfig
	AWS    AWSConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
	Host string
}

// SFTPConfig holds SFTP server configuration
type SFTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	UploadDir string
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	Region string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		SFTP: SFTPConfig{
			Host:     getEnv("SFTP_HOST", "sftp"),
			Port:     getEnvAsInt("SFTP_PORT", 22),
			User:     getEnv("SFTP_USER", "addiuser"),
			Password: getEnv("SFTP_PASSWORD", "addipass"),
			UploadDir: getEnv("SFTP_UPLOAD_DIR", "/uploads"),
		},
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-east-1"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.SFTP.Host == "" {
		return fmt.Errorf("SFTP host is required")
	}

	if c.SFTP.User == "" {
		return fmt.Errorf("SFTP user is required")
	}

	if c.SFTP.Password == "" {
		return fmt.Errorf("SFTP password is required")
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
