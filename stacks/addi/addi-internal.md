# Addi Stack - Transferencia a SFTP Interno (EC2 en AWS)

## 📋 Resumen

Arquitectura serverless para transferir archivos desde S3 a un servidor SFTP **dentro de AWS** usando EC2 privado. Ideal para testing o casos donde el destino SFTP está en AWS.

**Costo estimado**: ~$7/mes (desarrollo) | ~$15/mes (producción)

---

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────────────────────────────┐
│              ARQUITECTURA INTERNA (SFTP EN AWS)                 │
└─────────────────────────────────────────────────────────────────┘

Cliente/Aplicación
    ↓ (PUT Object)
┌────────────────────────────────────┐
│ S3 Bucket (Landing Zone)           │
│ - EventBridge notifications: ON    │  Estrategia: BucketTypeDevelopment
│ - Encryption: S3_MANAGED           │  Costo: ~$0.24/mes (10GB)
│ - Lifecycle: 30 días               │
└────────────┬───────────────────────┘
             │ s3:ObjectCreated:* event
             ↓
┌────────────────────────────────────┐
│ EventBridge Rule                   │
│ - Source: aws.s3                   │  Costo: Free Tier
│ - Target: Lambda Function          │
└────────────┬───────────────────────┘
             │ Trigger
             ↓
┌─────────────────────────────────────────────────────────────────┐
│ Lambda Function (Go 1.x)                                        │
│                                                                  │
│ Flow:                                                            │
│ 1. Load config from Secrets Manager (.env pattern)              │
│ 2. Validate file (size, extension)                              │
│ 3. Download file from S3                                         │
│ 4. Connect to SFTP via SSH (golang.org/x/crypto/ssh)            │
│ 5. Transfer file using SFTP protocol                             │
│ 6. Verify transfer (checksum/size)                               │
│ 7. Archive to S3 /processed/ folder                              │
│ 8. Send notification (SNS)                                       │
│                                                                  │
│ Config:                                                          │
│ - Memory: 512MB                                                  │
│ - Timeout: 5 min                                                 │
│ - VPC: Enabled (mismo VPC que EC2)                              │
│ - Security Group: Permite egress SSH (port 22)                  │
│                                                                  │
│ Costo: Free Tier (1M requests/mes)                              │
└────────────┬────────────────────────────────────────────────────┘
             │ SSH/SFTP (port 22)
             │ VPC Internal (no internet)
             ↓
┌─────────────────────────────────────────────────────────────────┐
│ EC2 Instance (SFTP Server)                                      │
│                                                                  │
│ Network:                                                         │
│ - VPC: Private subnet (NO public IP)                            │
│ - Security Group: Solo permite SSH desde Lambda SG              │
│ - Accessibility: SOLO desde VPC interno                         │
│                                                                  │
│ Specs:                                                           │
│ - Instance: t4g.nano (ARM Graviton)                             │
│ - vCPU: 2 cores                                                  │
│ - RAM: 512 MB                                                    │
│ - Storage: 8GB EBS gp3                                           │
│ - OS: Amazon Linux 2023                                          │
│                                                                  │
│ Software:                                                        │
│ - OpenSSH Server (sshd)                                          │
│ - SFTP subsystem enabled                                         │
│ - User: sftpuser (chroot jail)                                   │
│ - Auth: SSH key-based (no password)                              │
│                                                                  │
│ Security:                                                        │
│ - NO acceso desde Internet                                       │
│ - Solo Lambda puede conectarse                                   │
│ - Chroot jail (usuario aislado)                                  │
│ - SSH key rotation via Secrets Manager                           │
│                                                                  │
│ Costo: ~$3.50/mes (Reserved 1yr)                                │
│        ~$0.80/mes (8GB EBS gp3)                                  │
│        Total: ~$4.30/mes                                         │
└─────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────┐
│ AWS Secrets Manager                │
│                                     │  Costo: $0.40/mes por secret
│ Secret: addi/sftp/config            │        $0.05 por 10k API calls
│                                     │
│ Estructura .env:                    │
│ {                                   │
│   "SFTP_HOST": "10.0.1.50",         │
│   "SFTP_PORT": "22",                │
│   "SFTP_USER": "sftpuser",          │
│   "SFTP_PRIVATE_KEY": "-----...",   │
│   "SFTP_REMOTE_PATH": "/uploads",   │
│   "MAX_FILE_SIZE_MB": "100",        │
│   "SNS_TOPIC_ARN": "arn:..."        │
│ }                                   │
└────────────────────────────────────┘

┌────────────────────────────────────┐
│ CloudWatch Logs + Metrics          │  Costo: ~$2.80/mes
│ - Lambda execution logs            │
│ - Transfer success/failure metrics │
└────────────────────────────────────┘

┌────────────────────────────────────┐
│ SNS Topic (Notifications)          │  Costo: Free Tier
│ - Email alerts                     │
│ - Slack webhooks (opcional)        │
└────────────────────────────────────┘
```

---

## 💰 Análisis de Costos Detallado

### Ambiente de Desarrollo/Testing

| Servicio | Configuración | Cálculo | Costo Mensual |
|----------|---------------|---------|---------------|
| **S3** | 10GB storage | 10GB × $0.023 | $0.23 |
| **S3** | 1000 PUT requests | 1000 × $0.005/1000 | $0.01 |
| **S3** | 5000 GET requests | 5000 × $0.0004/1000 | $0.00 |
| **Lambda** | 1000 invocations, 512MB, 30s avg | Free Tier | $0.00 |
| **EC2** | t4g.nano Reserved 1yr | 730h × $0.0048 | $3.50 |
| **EBS** | 8GB gp3 | 8GB × $0.10 | $0.80 |
| **Secrets Manager** | 1 secret, 1000 retrievals | $0.40 + ($0.05/10k × 0.1) | $0.40 |
| **EventBridge** | 1000 eventos/mes | Free Tier | $0.00 |
| **CloudWatch Logs** | 5GB ingestion | 5GB × $0.50 | $2.50 |
| **CloudWatch Metrics** | 10 custom metrics | 10 × $0.03 | $0.30 |
| **SNS** | 100 notifications | Free Tier | $0.00 |
| **VPC** | Subnets, SGs, IGW | Incluido | $0.00 |
| | **TOTAL MENSUAL** | | **$7.74** |

### Optimizaciones Aplicadas

✅ **EC2 Reserved Instance 1 año**: 62% descuento vs On-Demand
✅ **t4g.nano ARM Graviton**: 40% más barato que t3.nano x86
✅ **EBS gp3**: 20% más barato que gp2
✅ **Sin NAT Gateway**: Ahorro de $32/mes (EC2 en subnet privada)
✅ **Sin Transfer Family**: Ahorro de ~$150/mes
✅ **Lambda Free Tier**: 1M requests gratis
✅ **S3 Development strategy**: Sin versioning ni KMS

---

## 🔧 Implementación CDK

### 1. Estructura del Proyecto

```
stacks/addi/
├── addi-internal.md                    # Este documento
├── addi-internal-stack.go              # Stack CDK
├── lambda/
│   └── sftp-transfer/
│       ├── go.mod
│       ├── go.sum
│       ├── main.go                     # Handler Lambda
│       ├── sftp/
│       │   └── client.go               # Cliente SFTP
│       └── config/
│           └── secrets.go              # Carga desde Secrets Manager
├── scripts/
│   ├── generate-ssh-keys.sh            # Genera par de llaves SSH
│   └── upload-secrets.sh               # Sube .env a Secrets Manager
└── .env.example                        # Template de configuración
```

### 2. Archivo .env (Template)

**`.env.example`**:

```bash
# SFTP Server Configuration
SFTP_HOST=10.0.1.50
SFTP_PORT=22
SFTP_USER=sftpuser
SFTP_PRIVATE_KEY=-----BEGIN OPENSSH PRIVATE KEY-----\nxxxxxx\n-----END OPENSSH PRIVATE KEY-----
SFTP_REMOTE_PATH=/home/sftpuser/uploads

# Transfer Configuration
MAX_FILE_SIZE_MB=100
ARCHIVE_PROCESSED_FILES=true

# AWS Resources
SNS_TOPIC_ARN=arn:aws:sns:us-east-1:123456789012:addi-sftp-notifications

# Monitoring
LOG_LEVEL=INFO
ENABLE_METRICS=true
```

### 3. Lambda Function en Go

**`lambda/sftp-transfer/main.go`**:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"sftp-transfer/sftp"
	appConfig "sftp-transfer/config"
)

// S3Event representa el evento de EventBridge para S3
type S3Event struct {
	Detail struct {
		Bucket struct {
			Name string `json:"name"`
		} `json:"bucket"`
		Object struct {
			Key  string `json:"key"`
			Size int64  `json:"size"`
		} `json:"object"`
	} `json:"detail"`
}

type Handler struct {
	s3Client     *s3.Client
	snsClient    *sns.Client
	sftpClient   *sftp.Client
	cfg          *appConfig.Config
}

func NewHandler(ctx context.Context) (*Handler, error) {
	// Cargar configuración de AWS
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Cargar configuración desde Secrets Manager
	secretsClient := secretsmanager.NewFromConfig(awsCfg)
	appCfg, err := appConfig.LoadFromSecretsManager(ctx, secretsClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from Secrets Manager: %w", err)
	}

	// Inicializar cliente SFTP
	sftpClient, err := sftp.NewClient(appCfg.SFTP)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &Handler{
		s3Client:   s3.NewFromConfig(awsCfg),
		snsClient:  sns.NewFromConfig(awsCfg),
		sftpClient: sftpClient,
		cfg:        appCfg,
	}, nil
}

func (h *Handler) HandleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	log.Printf("Received event: %s", event.DetailType)

	// Parse S3 event from EventBridge
	var s3Event S3Event
	if err := json.Unmarshal(event.Detail, &s3Event); err != nil {
		return fmt.Errorf("failed to parse S3 event: %w", err)
	}

	bucket := s3Event.Detail.Bucket.Name
	key := s3Event.Detail.Object.Key
	size := s3Event.Detail.Object.Size

	log.Printf("Processing file: s3://%s/%s (size: %d bytes)", bucket, key, size)

	// Validar tamaño del archivo
	maxSize := int64(h.cfg.MaxFileSizeMB) * 1024 * 1024
	if size > maxSize {
		return h.handleError(ctx, bucket, key, fmt.Errorf("file too large: %d bytes (max: %d MB)", size, h.cfg.MaxFileSizeMB))
	}

	// Descargar archivo desde S3
	log.Printf("Downloading file from S3...")
	fileData, err := h.downloadFromS3(ctx, bucket, key)
	if err != nil {
		return h.handleError(ctx, bucket, key, fmt.Errorf("failed to download from S3: %w", err))
	}

	// Conectar a SFTP server
	log.Printf("Connecting to SFTP server: %s@%s:%s", h.cfg.SFTP.User, h.cfg.SFTP.Host, h.cfg.SFTP.Port)
	if err := h.sftpClient.Connect(); err != nil {
		return h.handleError(ctx, bucket, key, fmt.Errorf("failed to connect to SFTP: %w", err))
	}
	defer h.sftpClient.Close()

	// Transferir archivo
	remoteFilename := filepath.Base(key)
	remotePath := filepath.Join(h.cfg.SFTP.RemotePath, remoteFilename)

	log.Printf("Transferring file to: %s", remotePath)
	if err := h.sftpClient.Upload(fileData, remotePath); err != nil {
		return h.handleError(ctx, bucket, key, fmt.Errorf("failed to upload to SFTP: %w", err))
	}

	// Verificar transferencia
	remoteSize, err := h.sftpClient.Stat(remotePath)
	if err != nil {
		return h.handleError(ctx, bucket, key, fmt.Errorf("failed to verify upload: %w", err))
	}

	if remoteSize != size {
		return h.handleError(ctx, bucket, key, fmt.Errorf("size mismatch: local=%d, remote=%d", size, remoteSize))
	}

	log.Printf("✅ Transfer successful: %s", remotePath)

	// Archivar archivo en S3 (opcional)
	if h.cfg.ArchiveProcessedFiles {
		processedKey := fmt.Sprintf("processed/%s", key)
		if err := h.archiveInS3(ctx, bucket, key, processedKey); err != nil {
			log.Printf("Warning: failed to archive file: %v", err)
		} else {
			log.Printf("File archived to: s3://%s/%s", bucket, processedKey)
		}
	}

	// Enviar notificación de éxito
	return h.notifySuccess(ctx, bucket, key, remotePath, size)
}

func (h *Handler) downloadFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	result, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (h *Handler) archiveInS3(ctx context.Context, bucket, sourceKey, destKey string) error {
	// Copiar archivo
	copySource := fmt.Sprintf("%s/%s", bucket, sourceKey)
	_, err := h.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &bucket,
		CopySource: &copySource,
		Key:        &destKey,
	})
	if err != nil {
		return err
	}

	// Eliminar original
	_, err = h.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &sourceKey,
	})
	return err
}

func (h *Handler) notifySuccess(ctx context.Context, bucket, key, remotePath string, size int64) error {
	message := fmt.Sprintf(`✅ SFTP Transfer Successful

S3 Source: s3://%s/%s
SFTP Destination: %s@%s:%s
File Size: %d bytes
Status: SUCCESS
Timestamp: %s
`, bucket, key, h.cfg.SFTP.User, h.cfg.SFTP.Host, remotePath, size,
		os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME"))

	_, err := h.snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: &h.cfg.SNSTopicArn,
		Subject:  stringPtr("✅ SFTP Transfer Success"),
		Message:  &message,
	})
	return err
}

func (h *Handler) handleError(ctx context.Context, bucket, key string, err error) error {
	log.Printf("❌ Transfer failed: %v", err)

	message := fmt.Sprintf(`❌ SFTP Transfer Failed

S3 Source: s3://%s/%s
Error: %v
Status: FAILED
Timestamp: %s
`, bucket, key, err, os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME"))

	h.snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: &h.cfg.SNSTopicArn,
		Subject:  stringPtr("❌ SFTP Transfer Failure"),
		Message:  &message,
	})

	return err
}

func stringPtr(s string) *string {
	return &s
}

func main() {
	ctx := context.Background()
	handler, err := NewHandler(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	lambda.Start(handler.HandleRequest)
}
```

**`lambda/sftp-transfer/config/secrets.go`**:

```go
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Config struct {
	SFTP                  SFTPConfig
	MaxFileSizeMB         int
	ArchiveProcessedFiles bool
	SNSTopicArn           string
	LogLevel              string
	EnableMetrics         bool
}

type SFTPConfig struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
	RemotePath string
}

// Estructura que coincide con el JSON en Secrets Manager
type SecretsPayload struct {
	SFTPHost              string `json:"SFTP_HOST"`
	SFTPPort              string `json:"SFTP_PORT"`
	SFTPUser              string `json:"SFTP_USER"`
	SFTPPrivateKey        string `json:"SFTP_PRIVATE_KEY"`
	SFTPRemotePath        string `json:"SFTP_REMOTE_PATH"`
	MaxFileSizeMB         string `json:"MAX_FILE_SIZE_MB"`
	ArchiveProcessedFiles string `json:"ARCHIVE_PROCESSED_FILES"`
	SNSTopicArn           string `json:"SNS_TOPIC_ARN"`
	LogLevel              string `json:"LOG_LEVEL"`
	EnableMetrics         string `json:"ENABLE_METRICS"`
}

func LoadFromSecretsManager(ctx context.Context, client *secretsmanager.Client) (*Config, error) {
	// Obtener el ARN del secret desde variable de entorno
	secretARN := os.Getenv("CONFIG_SECRET_ARN")
	if secretARN == "" {
		return nil, fmt.Errorf("CONFIG_SECRET_ARN environment variable not set")
	}

	// Obtener el secret
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Parse JSON
	var payload SecretsPayload
	if err := json.Unmarshal([]byte(*result.SecretString), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	// Convertir tipos
	maxSize, _ := strconv.Atoi(payload.MaxFileSizeMB)
	if maxSize == 0 {
		maxSize = 100 // default
	}

	archiveFiles, _ := strconv.ParseBool(payload.ArchiveProcessedFiles)
	enableMetrics, _ := strconv.ParseBool(payload.EnableMetrics)

	return &Config{
		SFTP: SFTPConfig{
			Host:       payload.SFTPHost,
			Port:       payload.SFTPPort,
			User:       payload.SFTPUser,
			PrivateKey: payload.SFTPPrivateKey,
			RemotePath: payload.SFTPRemotePath,
		},
		MaxFileSizeMB:         maxSize,
		ArchiveProcessedFiles: archiveFiles,
		SNSTopicArn:           payload.SNSTopicArn,
		LogLevel:              payload.LogLevel,
		EnableMetrics:         enableMetrics,
	}, nil
}
```

**`lambda/sftp-transfer/sftp/client.go`**:

```go
package sftp

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"github.com/pkg/sftp"
)

type Config struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
	RemotePath string
}

type Client struct {
	cfg        Config
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewClient(cfg Config) (*Client, error) {
	return &Client{
		cfg: cfg,
	}, nil
}

func (c *Client) Connect() error {
	// Parse private key
	signer, err := ssh.ParsePrivateKey([]byte(c.cfg.PrivateKey))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// SSH client config
	config := &ssh.ClientConfig{
		User: c.cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // ⚠️ Para producción, usar known_hosts
	}

	// Connect SSH
	address := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)
	sshClient, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect SSH: %w", err)
	}
	c.sshClient = sshClient

	// Open SFTP session
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		c.sshClient.Close()
		return fmt.Errorf("failed to open SFTP session: %w", err)
	}
	c.sftpClient = sftpClient

	return nil
}

func (c *Client) Upload(data []byte, remotePath string) error {
	if c.sftpClient == nil {
		return fmt.Errorf("not connected")
	}

	// Crear directorios padre si no existen
	remoteDir := filepath.Dir(remotePath)
	if err := c.sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Crear archivo remoto
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Escribir datos
	_, err = io.Copy(remoteFile, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func (c *Client) Stat(remotePath string) (int64, error) {
	if c.sftpClient == nil {
		return 0, fmt.Errorf("not connected")
	}

	stat, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat remote file: %w", err)
	}

	return stat.Size(), nil
}

func (c *Client) Close() error {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
	return nil
}
```

**`lambda/sftp-transfer/go.mod`**:

```go
module sftp-transfer

go 1.21

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.30.3
	github.com/aws/aws-sdk-go-v2/config v1.27.27
	github.com/aws/aws-sdk-go-v2/service/s3 v1.58.3
	github.com/aws/aws-sdk-go-v2/service/sns v1.31.3
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.32.4
	github.com/pkg/sftp v1.13.6
	golang.org/x/crypto v0.25.0
)
```

### 4. Stack CDK en Go

**`stacks/addi/addi-internal-stack.go`**:

```go
package addi

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk-library/constructs/S3"
)

type AddiInternalStackProps struct {
	awscdk.StackProps
	SSHPublicKey string
	AlertEmail   string
}

func NewAddiInternalStack(scope constructs.Construct, id string, props *AddiInternalStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// =================================================================
	// 1. VPC (Simple, sin NAT Gateway para reducir costos)
	// =================================================================
	vpc := awsec2.NewVpc(stack, jsii.String("AddiVPC"), &awsec2.VpcProps{
		MaxAzs:       jsii.Number(2),
		NatGateways:  jsii.Number(0), // Sin NAT - EC2 en subnet privada
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// =================================================================
	// 2. S3 Bucket (Landing Zone)
	// =================================================================
	bucket := s3.NewSimpleStorageServiceFactory(stack, "AddiLandingBucket",
		s3.SimpleStorageServiceFactoryProps{
			BucketType:        s3.BucketTypeDevelopment,
			BucketName:        jsii.String("addi-file-landing-internal-dev"),
			RemovalPolicy:     jsii.String("destroy"),
			AutoDeleteObjects: jsii.Bool(true),
		})

	// =================================================================
	// 3. SNS Topic para notificaciones
	// =================================================================
	snsTopic := awssns.NewTopic(stack, jsii.String("SFTPTransferNotifications"), &awssns.TopicProps{
		DisplayName: jsii.String("Addi SFTP Transfer Notifications"),
	})

	if props.AlertEmail != "" {
		snsTopic.AddSubscription(
			awssnssubscriptions.NewEmailSubscription(jsii.String(props.AlertEmail), nil),
		)
	}

	// =================================================================
	// 4. Secrets Manager - Configuración SFTP
	// =================================================================

	// Leer .env file y crear JSON para Secrets Manager
	envContent := loadEnvFile("stacks/addi/.env")

	sftpSecret := awssecretsmanager.NewSecret(stack, jsii.String("SFTPConfig"), &awssecretsmanager.SecretProps{
		SecretName:  jsii.String("addi/sftp/config"),
		Description: jsii.String("SFTP configuration for Addi transfer Lambda"),
		SecretStringValue: awscdk.SecretValue_UnsafePlainText(jsii.String(envContent)),
	})

	// =================================================================
	// 5. Security Groups
	// =================================================================

	// Security Group para Lambda
	lambdaSG := awsec2.NewSecurityGroup(stack, jsii.String("LambdaSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		Description:       jsii.String("Security group for SFTP transfer Lambda"),
		AllowAllOutbound:  jsii.Bool(true),
	})

	// Security Group para SFTP Server
	sftpSG := awsec2.NewSecurityGroup(stack, jsii.String("SFTPServerSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Security group for SFTP server"),
		AllowAllOutbound: jsii.Bool(true),
	})

	// Permitir SSH desde Lambda
	sftpSG.AddIngressRule(
		lambdaSG,
		awsec2.Port_Tcp(jsii.Number(22)),
		jsii.String("Allow SSH from Lambda"),
		jsii.Bool(false),
	)

	// =================================================================
	// 6. EC2 Instance (SFTP Server)
	// =================================================================

	// User Data Script
	userDataScript := fmt.Sprintf(`#!/bin/bash
set -e

# Update system
dnf update -y

# Install OpenSSH Server
dnf install -y openssh-server

# Create SFTP user
useradd -m -d /home/sftpuser -s /bin/bash sftpuser

# Create uploads directory
mkdir -p /home/sftpuser/uploads
chown sftpuser:sftpuser /home/sftpuser/uploads

# Configure chroot for SFTP
cat >> /etc/ssh/sshd_config <<'EOF'

# SFTP Chroot Configuration
Match User sftpuser
    ForceCommand internal-sftp
    ChrootDirectory /home/sftpuser
    PermitTunnel no
    AllowAgentForwarding no
    AllowTcpForwarding no
    X11Forwarding no
EOF

# Set chroot permissions
chown root:root /home/sftpuser
chmod 755 /home/sftpuser

# Configure SSH key authentication
mkdir -p /home/sftpuser/.ssh
echo "%s" > /home/sftpuser/.ssh/authorized_keys
chown -R sftpuser:sftpuser /home/sftpuser/.ssh
chmod 700 /home/sftpuser/.ssh
chmod 600 /home/sftpuser/.ssh/authorized_keys

# Restart SSH daemon
systemctl restart sshd
systemctl enable sshd

# Setup complete
echo "SFTP Server configured successfully" > /var/log/sftp-setup-complete.log
`, props.SSHPublicKey)

	// EC2 Instance
	instance := awsec2.NewInstance(stack, jsii.String("SFTPServer"), &awsec2.InstanceProps{
		Vpc: vpc,
		InstanceType: awsec2.InstanceType_Of(
			awsec2.InstanceClass_BURSTABLE4_GRAVITON,
			awsec2.InstanceSize_NANO,
		),
		MachineImage: awsec2.MachineImage_LatestAmazonLinux2023(&awsec2.AmazonLinux2023ImageSsmParameterProps{
			CpuType: awsec2.AmazonLinuxCpuType_ARM_64,
		}),
		BlockDevices: &[]*awsec2.BlockDevice{
			{
				DeviceName: jsii.String("/dev/xvda"),
				Volume: awsec2.BlockDeviceVolume_Ebs(jsii.Number(8), &awsec2.EbsDeviceOptions{
					VolumeType:          awsec2.EbsDeviceVolumeType_GP3,
					DeleteOnTermination: jsii.Bool(true),
				}),
			},
		},
		SecurityGroup: sftpSG,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		UserData: awsec2.UserData_Custom(jsii.String(userDataScript)),
	})

	// =================================================================
	// 7. Lambda Function (Go runtime)
	// =================================================================

	transferLambda := awslambda.NewFunction(stack, jsii.String("SFTPTransferFunction"), &awslambda.FunctionProps{
		Runtime:     awslambda.Runtime_PROVIDED_AL2023(),
		Handler:     jsii.String("bootstrap"),
		Code:        awslambda.Code_FromAsset(jsii.String("lambda/sftp-transfer/build"), nil),
		MemorySize:  jsii.Number(512),
		Timeout:     awscdk.Duration_Minutes(jsii.Number(5)),
		Environment: &map[string]*string{
			"CONFIG_SECRET_ARN": sftpSecret.SecretArn(),
		},
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
		},
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	// Permisos IAM
	bucket.GrantRead(transferLambda, nil)
	bucket.GrantDelete(transferLambda, nil)
	bucket.GrantPut(transferLambda, nil)
	sftpSecret.GrantRead(transferLambda, nil)
	snsTopic.GrantPublish(transferLambda)

	// =================================================================
	// 8. EventBridge Rule
	// =================================================================

	rule := awsevents.NewRule(stack, jsii.String("S3FileUploadRule"), &awsevents.RuleProps{
		Description: jsii.String("Trigger Lambda when file uploaded to S3"),
		EventPattern: &awsevents.EventPattern{
			Source:     jsii.Strings("aws.s3"),
			DetailType: jsii.Strings("Object Created"),
			Detail: &map[string]interface{}{
				"bucket": map[string]interface{}{
					"name": jsii.Strings(*bucket.BucketName()),
				},
			},
		},
	})

	rule.AddTarget(awseventstargets.NewLambdaFunction(transferLambda, nil))

	// =================================================================
	// 9. Outputs
	// =================================================================

	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 Landing Bucket Name"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("SFTPServerPrivateIP"), &awscdk.CfnOutputProps{
		Value:       instance.InstancePrivateIp(),
		Description: jsii.String("SFTP Server Private IP Address"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LambdaFunctionName"), &awscdk.CfnOutputProps{
		Value:       transferLambda.FunctionName(),
		Description: jsii.String("SFTP Transfer Lambda Function Name"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("SecretARN"), &awscdk.CfnOutputProps{
		Value:       sftpSecret.SecretArn(),
		Description: jsii.String("Secrets Manager ARN for SFTP config"),
	})

	return stack
}

// Helper function para cargar .env y convertir a JSON
func loadEnvFile(path string) string {
	// Leer archivo .env
	content, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read .env file: %v", err))
	}

	// Aquí deberías implementar parser real de .env a JSON
	// Por simplicidad, asumimos que el .env ya está en formato JSON
	// O usar library como github.com/joho/godotenv

	return string(content)
}
```

### 5. Scripts de Utilidad

**`scripts/generate-ssh-keys.sh`**:

```bash
#!/bin/bash
set -e

echo "🔐 Generating SSH key pair for SFTP authentication..."

# Generar par de llaves SSH
ssh-keygen -t rsa -b 4096 -f ./sftp_key -N "" -C "addi-sftp-key"

echo "✅ SSH keys generated:"
echo "   Private key: ./sftp_key"
echo "   Public key: ./sftp_key.pub"
echo ""
echo "📋 Next steps:"
echo "1. Add private key to .env file (SFTP_PRIVATE_KEY)"
echo "2. Public key will be injected into EC2 via CDK"
echo ""
echo "Public key content:"
cat ./sftp_key.pub
```

**`scripts/upload-secrets.sh`**:

```bash
#!/bin/bash
set -e

# Cargar variables desde .env
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found"
    echo "Please create .env file from .env.example template"
    exit 1
fi

# Convertir .env a JSON
ENV_JSON=$(cat .env | grep -v '^#' | grep -v '^$' | jq -Rs 'split("\n") | map(select(length > 0) | split("=") | {key: .[0], value: .[1]}) | from_entries')

echo "🔐 Uploading configuration to AWS Secrets Manager..."

# Crear o actualizar secret
aws secretsmanager create-secret \
    --name addi/sftp/config \
    --description "SFTP configuration for Addi transfer Lambda" \
    --secret-string "$ENV_JSON" \
    2>/dev/null || \
aws secretsmanager update-secret \
    --secret-id addi/sftp/config \
    --secret-string "$ENV_JSON"

echo "✅ Secrets uploaded successfully"
echo "Secret ARN: $(aws secretsmanager describe-secret --secret-id addi/sftp/config --query 'ARN' --output text)"
```

**`scripts/build-lambda.sh`**:

```bash
#!/bin/bash
set -e

echo "🔨 Building Go Lambda function..."

cd lambda/sftp-transfer

# Download dependencies
go mod download

# Build for Lambda (Linux ARM64)
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go

# Create deployment package
mkdir -p build
cp bootstrap build/
chmod +x build/bootstrap

echo "✅ Lambda function built successfully"
echo "   Output: lambda/sftp-transfer/build/bootstrap"
```

---

## 🚀 Instrucciones de Despliegue

### Paso 1: Preparar Credenciales SSH

```bash
cd stacks/addi

# Generar llaves SSH
./scripts/generate-ssh-keys.sh

# Esto crea:
# - sftp_key (privada)
# - sftp_key.pub (pública)
```

### Paso 2: Configurar .env

```bash
# Copiar template
cp .env.example .env

# Editar .env y agregar:
# - SFTP_HOST (se actualizará después del deploy con IP del EC2)
# - SFTP_PRIVATE_KEY (contenido de sftp_key)
# - SNS_TOPIC_ARN (se actualizará después del deploy)

# Ejemplo:
nano .env
```

### Paso 3: Compilar Lambda

```bash
./scripts/build-lambda.sh
```

### Paso 4: Desplegar Stack

```bash
cd ../..

# Sintetizar template
cdk synth AddiInternalStack

# Desplegar
cdk deploy AddiInternalStack \
  --parameters SSHPublicKey="$(cat stacks/addi/sftp_key.pub)" \
  --parameters AlertEmail="tu-email@empresa.com" \
  --require-approval never
```

### Paso 5: Actualizar .env con Outputs

```bash
# Obtener outputs del stack
export SFTP_IP=$(aws cloudformation describe-stacks \
  --stack-name AddiInternalStack \
  --query 'Stacks[0].Outputs[?OutputKey==`SFTPServerPrivateIP`].OutputValue' \
  --output text)

export SNS_ARN=$(aws cloudformation describe-stacks \
  --stack-name AddiInternalStack \
  --query 'Stacks[0].Outputs[?OutputKey==`SNSTopicArn`].OutputValue' \
  --output text)

# Actualizar .env
sed -i '' "s/SFTP_HOST=.*/SFTP_HOST=$SFTP_IP/" stacks/addi/.env
sed -i '' "s|SNS_TOPIC_ARN=.*|SNS_TOPIC_ARN=$SNS_ARN|" stacks/addi/.env

# Subir configuración a Secrets Manager
./stacks/addi/scripts/upload-secrets.sh
```

### Paso 6: Probar Transferencia

```bash
# Subir archivo de prueba
echo "Test file for SFTP transfer" > test.txt
aws s3 cp test.txt s3://addi-file-landing-internal-dev/uploads/test.txt

# Monitorear logs de Lambda
aws logs tail /aws/lambda/AddiInternalStack-SFTPTransferFunction --follow

# Verificar transferencia (necesitas acceso al VPC, usa bastion o SSM)
aws ssm start-session --target i-xxxxx  # ID de instancia EC2
# Dentro de la sesión:
sudo su - sftpuser
ls -la /home/sftpuser/uploads/
```

---

## 🐛 Troubleshooting

### Error: Lambda timeout

**Causa**: Archivo muy grande o conexión lenta

**Solución**:
```go
Timeout: awscdk.Duration_Minutes(jsii.Number(15)),  // Aumentar a 15 min
```

### Error: Connection refused (EC2)

**Causa**: Security Groups mal configurados

**Solución**:
```bash
# Verificar SGs
aws ec2 describe-security-groups --filters "Name=group-name,Values=*Lambda*"
aws ec2 describe-security-groups --filters "Name=group-name,Values=*SFTP*"

# Debe existir regla: Lambda SG → SFTP SG (port 22)
```

### Error: Secret not found

**Causa**: .env no subido a Secrets Manager

**Solución**:
```bash
./stacks/addi/scripts/upload-secrets.sh
```

### Warning: InsecureIgnoreHostKey

**Para producción**, reemplazar en `sftp/client.go`:

```go
// Desarrollo (inseguro):
HostKeyCallback: ssh.InsecureIgnoreHostKey(),

// Producción (seguro):
HostKeyCallback: ssh.FixedHostKey(serverPublicKey),
```

---

## 📚 Referencias

- [AWS Lambda Go Runtime](https://github.com/aws/aws-lambda-go)
- [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)
- [pkg/sftp](https://github.com/pkg/sftp)
- [AWS Secrets Manager Go SDK](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/secretsmanager)

---

**Última actualización**: 2025-10-11
**Versión**: 1.0
**Costo estimado**: ~$7.74/mes (development)
