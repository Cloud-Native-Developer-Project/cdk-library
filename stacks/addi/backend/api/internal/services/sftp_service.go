package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"addi-backend/internal/config"
	"addi-backend/internal/domain"
)

// SFTPServiceImpl implements the SFTPService interface
type SFTPServiceImpl struct {
	config     *config.SFTPConfig
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// NewSFTPService creates a new SFTP service instance
func NewSFTPService(cfg *config.SFTPConfig) domain.SFTPService {
	return &SFTPServiceImpl{
		config: cfg,
	}
}

// Connect establishes connection to SFTP server
func (s *SFTPServiceImpl) Connect(ctx context.Context) error {
	// Configure SSH client
	sshConfig := &ssh.ClientConfig{
		User: s.config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // ⚠️ Only for development
		Timeout:         10 * time.Second,
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server %s: %w", addr, err)
	}
	s.sshClient = sshClient

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		s.sshClient.Close()
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	s.sftpClient = sftpClient

	return nil
}

// UploadFile uploads a file to SFTP server
func (s *SFTPServiceImpl) UploadFile(ctx context.Context, reader io.Reader, remotePath string, size int64) (*domain.SFTPTransferResult, error) {
	if s.sftpClient == nil {
		return nil, fmt.Errorf("SFTP client not connected")
	}

	startTime := time.Now()
	result := &domain.SFTPTransferResult{
		RemotePath: remotePath,
	}

	// Ensure upload directory exists
	uploadDir := filepath.Join(s.config.UploadDir, filepath.Dir(remotePath))
	if err := s.sftpClient.MkdirAll(uploadDir); err != nil {
		result.Error = fmt.Errorf("failed to create remote directory %s: %w", uploadDir, err)
		return result, result.Error
	}

	// Create remote file
	fullPath := filepath.Join(s.config.UploadDir, remotePath)
	remoteFile, err := s.sftpClient.Create(fullPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create remote file %s: %w", fullPath, err)
		return result, result.Error
	}
	defer remoteFile.Close()

	// Copy data from reader to remote file
	bytesWritten, err := io.Copy(remoteFile, reader)
	if err != nil {
		result.Error = fmt.Errorf("failed to upload file to %s: %w", fullPath, err)
		return result, result.Error
	}

	result.Success = true
	result.BytesTransferred = bytesWritten
	result.Duration = time.Since(startTime)
	result.RemotePath = fullPath

	return result, nil
}

// Close closes the SFTP connection
func (s *SFTPServiceImpl) Close() error {
	var errors []error

	if s.sftpClient != nil {
		if err := s.sftpClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close SFTP client: %w", err))
		}
		s.sftpClient = nil
	}

	if s.sshClient != nil {
		if err := s.sshClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close SSH client: %w", err))
		}
		s.sshClient = nil
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing SFTP connection: %v", errors)
	}

	return nil
}

// HealthCheck verifies SFTP connection is alive
func (s *SFTPServiceImpl) HealthCheck(ctx context.Context) error {
	if s.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	// Try to stat the upload directory
	_, err := s.sftpClient.Stat(s.config.UploadDir)
	if err != nil {
		return fmt.Errorf("SFTP health check failed: %w", err)
	}

	return nil
}
