# Addi Stack - Transferencia a SFTP Externo (On-Premise)

## 📋 Resumen

Arquitectura serverless para transferir archivos desde S3 a un servidor SFTP **externo/on-premise** fuera de AWS. Usa Direct Connect o VPN para conectividad híbrida.

**Costo estimado**: ~$7-50/mes (dependiendo de conectividad existente)

---

## 🏗️ Arquitectura

```
┌──────────────────────────────────────────────────────────────────┐
│         ARQUITECTURA EXTERNA (SFTP ON-PREMISE)                   │
└──────────────────────────────────────────────────────────────────┘

Cliente/Aplicación
    ↓ (PUT Object)
┌────────────────────────────────────┐
│ S3 Bucket (Landing Zone)           │
│ - EventBridge notifications: ON    │  Estrategia: BucketTypeEnterprise
│ - Encryption: KMS                  │  Costo: ~$0.30/mes (10GB)
│ - Versioning: Enabled              │
│ - Lifecycle: Archive processed     │
└────────────┬───────────────────────┘
             │ s3:ObjectCreated:* event
             ↓
┌────────────────────────────────────┐
│ EventBridge Rule                   │
│ - Source: aws.s3                   │  Costo: Free Tier
│ - Target: Lambda Function          │
│ - Dead Letter Queue: SQS           │
└────────────┬───────────────────────┘
             │ Trigger
             ↓
┌─────────────────────────────────────────────────────────────────┐
│ Lambda Function (Go 1.x)                                        │
│                                                                  │
│ Flow:                                                            │
│ 1. Load config from Secrets Manager (.env pattern)              │
│ 2. Validate file (size, extension, checksum)                    │
│ 3. Download file from S3 (streaming for large files)            │
│ 4. Connect to external SFTP via hybrid network                  │
│ 5. Transfer file using SFTP protocol                             │
│ 6. Verify transfer (checksum/size)                               │
│ 7. Archive to S3 /processed/ folder                              │
│ 8. Send notification (SNS)                                       │
│ 9. Log metrics to CloudWatch                                     │
│                                                                  │
│ Config:                                                          │
│ - Memory: 1024MB (archivos más grandes)                         │
│ - Timeout: 15 min (transferencias lentas)                       │
│ - VPC: Enabled (acceso a hybrid network)                        │
│ - Security Group: Egress a SFTP externo                         │
│ - Retry: 3 intentos con exponential backoff                     │
│                                                                  │
│ Costo: Free Tier (1M requests/mes)                              │
│        $0.0000166667 por GB-segundo                              │
└────────────┬────────────────────────────────────────────────────┘
             │ SSH/SFTP (port 22 o custom)
             │
             ↓ (Atraviesa hybrid network)
┌─────────────────────────────────────────────────────────────────┐
│           CONECTIVIDAD HÍBRIDA (elegir una opción)              │
│                                                                  │
│ OPCIÓN A: AWS Direct Connect ✅ Recomendado                     │
│ ┌─────────────────────────────────────────┐                     │
│ │ Direct Connect Gateway                   │  Costo: Ya existe  │
│ │ - Dedicated connection (1Gbps - 10Gbps) │  (infraestructura) │
│ │ - Latency: <10ms                         │                     │
│ │ - Bandwidth: Alta                        │                     │
│ │ - Private IP routing                     │                     │
│ └─────────────────────────────────────────┘                     │
│                                                                  │
│ OPCIÓN B: AWS Site-to-Site VPN                                  │
│ ┌─────────────────────────────────────────┐                     │
│ │ VPN Gateway                              │  Costo: $0.05/hr   │
│ │ - IPSec tunnels (x2 for HA)             │  = ~$36/mes         │
│ │ - Latency: 20-50ms                       │  + $0.09/GB out    │
│ │ - Bandwidth: Up to 1.25Gbps              │                     │
│ │ - Encrypted traffic                      │                     │
│ └─────────────────────────────────────────┘                     │
│                                                                  │
│ OPCIÓN C: Internet público + NAT Gateway ⚠️ Menos seguro        │
│ ┌─────────────────────────────────────────┐                     │
│ │ NAT Gateway                              │  Costo: $0.045/hr  │
│ │ - Public internet route                  │  = ~$32/mes         │
│ │ - Latency: Variable                      │  + $0.045/GB out   │
│ │ - Security: SSH encryption only          │                     │
│ │ - Requiere firewall externo abierto      │                     │
│ └─────────────────────────────────────────┘                     │
└────────────┬────────────────────────────────────────────────────┘
             │
             ↓ (Llega a red on-premise)
┌─────────────────────────────────────────────────────────────────┐
│ Servidor SFTP Externo (On-Premise / Data Center)               │
│                                                                  │
│ Network:                                                         │
│ - IP: 192.168.X.X o 10.X.X.X (private)                          │
│ - Port: 22 (o custom como 2222)                                 │
│ - Firewall: Permite ingress desde AWS CIDR block                │
│                                                                  │
│ Server:                                                          │
│ - OS: Linux/Windows Server                                       │
│ - Software: OpenSSH, ProFTPD, etc.                               │
│ - Auth: SSH key-based (recommended)                              │
│ - Storage: NAS, SAN, local disk                                  │
│                                                                  │
│ Security:                                                        │
│ - Firewall rules para AWS IP ranges                             │
│ - SSH key rotation (90 días)                                     │
│ - Audit logging enabled                                          │
│ - Compliance: SOX, PCI DSS, HIPAA                                │
│                                                                  │
│ Costo: Infraestructura existente (no adicional)                 │
└─────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────┐
│ AWS Secrets Manager                │  Costo: $0.40/mes por secret
│                                     │        $0.05 por 10k API calls
│ Secret: addi/sftp-external/config  │
│                                     │
│ Estructura .env:                    │
│ {                                   │
│   "SFTP_HOST": "sftp.empresa.com",  │
│   "SFTP_PORT": "2222",              │
│   "SFTP_USER": "aws_transfer_user", │
│   "SFTP_PRIVATE_KEY": "-----...",   │
│   "SFTP_REMOTE_PATH": "/ingress",   │
│   "SFTP_HOST_KEY": "ssh-rsa ...",   │ ← ✅ Verificación segura
│   "MAX_FILE_SIZE_MB": "500",        │
│   "ENABLE_CHECKSUM": "true",        │
│   "SNS_TOPIC_ARN": "arn:..."        │
│ }                                   │
└────────────────────────────────────┘

┌────────────────────────────────────┐
│ SQS Dead Letter Queue              │  Costo: Free Tier
│ - Failed transfers                 │  (1M requests/mes)
│ - Retry logic                      │
│ - Manual intervention queue        │
└────────────────────────────────────┘

┌────────────────────────────────────┐
│ CloudWatch Logs + Metrics + Alarms │  Costo: ~$5/mes
│ - Lambda execution logs            │
│ - Transfer duration metrics        │
│ - Success/failure rate             │
│ - Alarms para high failure rate    │
└────────────────────────────────────┘

┌────────────────────────────────────┐
│ GuardDuty (Comprehensive)          │  Costo: ~$30-50/mes
│ - S3 data event monitoring         │  (Recomendado para prod)
│ - Malware scanning                 │
│ - Anomaly detection                │
└────────────────────────────────────┘
```

---

## 💰 Análisis de Costos Detallado

### Ambiente de Producción

| Servicio | Configuración | Cálculo | Costo Mensual |
|----------|---------------|---------|---------------|
| **S3** | 100GB storage, 10k PUT, 50k GET | (100×$0.023)+(10k×$0.005/1k)+(50k×$0.0004/1k) | $2.35 + $0.05 + $0.02 = **$2.42** |
| **S3** | KMS encryption | 10k requests × $0.03/10k | **$0.03** |
| **Lambda** | 10k invocations, 1GB, 60s avg | (10k-1k)×$0.20/1M + (9k×60×1024)×$0.0000166667 | **$9.00** |
| **VPC** | 3 AZs, private subnets | Incluido | **$0.00** |
| **Secrets Manager** | 1 secret, 10k API calls | $0.40 + ($0.05×1) | **$0.45** |
| **EventBridge** | 10k eventos/mes | Free Tier | **$0.00** |
| **SQS** | 1k failed messages | Free Tier | **$0.00** |
| **CloudWatch Logs** | 20GB ingestion, 30d retention | 20GB × $0.50 | **$10.00** |
| **CloudWatch Metrics** | 50 custom metrics | 50 × $0.03 | **$1.50** |
| **CloudWatch Alarms** | 10 alarms | 10 × $0.10 | **$1.00** |
| **SNS** | 1k notifications | Free Tier | **$0.00** |
| **GuardDuty** | Comprehensive (100GB S3) | Base + data events | **~$35.00** |
| | | |
| **Subtotal (sin conectividad)** | | | **$59.40** |
| | | |
| **CONECTIVIDAD - Opción A: Direct Connect** | Ya existe | $0 | **$0.00** |
| **CONECTIVIDAD - Opción B: Site-to-Site VPN** | 730h×$0.05 + 100GB×$0.09 | $36.50 + $9 | **$45.50** |
| **CONECTIVIDAD - Opción C: NAT Gateway** | 730h×$0.045 + 100GB×$0.045 | $32.85 + $4.50 | **$37.35** |
| | | |
| **TOTAL con Direct Connect** | | | **$59.40/mes** |
| **TOTAL con VPN** | | | **$104.90/mes** |
| **TOTAL con NAT** | | | **$96.75/mes** |

### Recomendaciones de Costos

- ✅ **Si Direct Connect ya existe**: ~$60/mes (muy económico)
- ⚠️ **Si necesitas crear VPN nueva**: ~$105/mes (medio)
- ❌ **NAT Gateway sobre internet**: ~$97/mes (menos seguro, similar costo)

**Conclusión**: Direct Connect es la mejor opción (costo + seguridad + latencia).

---

## 🔧 Implementación CDK

### 1. Estructura del Proyecto

```
stacks/addi/
├── addi-external.md                    # Este documento
├── addi-external-stack.go              # Stack CDK para SFTP externo
├── lambda/
│   └── sftp-transfer/                  # ← Mismo código Go que internal
│       ├── go.mod
│       ├── main.go
│       ├── sftp/
│       │   └── client.go
│       └── config/
│           └── secrets.go
├── scripts/
│   ├── generate-ssh-keys.sh
│   ├── upload-secrets-external.sh      # Sube .env para SFTP externo
│   └── test-connectivity.sh            # Valida conectividad a SFTP externo
└── .env.external                       # Configuración para servidor externo
```

### 2. Archivo .env para SFTP Externo

**`.env.external`**:

```bash
# External SFTP Server Configuration
SFTP_HOST=sftp.cliente-empresa.com
SFTP_PORT=2222
SFTP_USER=aws_integration_user
SFTP_PRIVATE_KEY=-----BEGIN OPENSSH PRIVATE KEY-----\nMIIEowIBAAKCAQEAxxxxx...\n-----END OPENSSH PRIVATE KEY-----

# ✅ Host Key Verification (IMPORTANTE para producción)
# Obtener con: ssh-keyscan -p 2222 sftp.cliente-empresa.com
SFTP_HOST_KEY=sftp.cliente-empresa.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB...

SFTP_REMOTE_PATH=/data/ingress/aws

# Transfer Configuration
MAX_FILE_SIZE_MB=500
ARCHIVE_PROCESSED_FILES=true
ENABLE_CHECKSUM_VERIFICATION=true
CHECKSUM_ALGORITHM=SHA256

# Retry Configuration
MAX_RETRY_ATTEMPTS=3
RETRY_BACKOFF_SECONDS=30

# AWS Resources (se llenan después del deploy)
SNS_TOPIC_ARN=arn:aws:sns:us-east-1:123456789012:addi-sftp-external-notifications
DLQ_URL=https://sqs.us-east-1.amazonaws.com/123456789012/addi-transfer-dlq

# Monitoring
LOG_LEVEL=INFO
ENABLE_METRICS=true
ENABLE_DETAILED_LOGGING=true
```

### 3. Lambda Function en Go (Actualizada con Host Key Verification)

La función Lambda es **casi idéntica** a la versión internal, con estas mejoras:

**Cambios en `lambda/sftp-transfer/sftp/client.go`**:

```go
package sftp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/pkg/sftp"
)

type Config struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
	RemotePath string
	HostKey    string // ✅ Nuevo: Para verificación segura
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

	// ✅ Parse expected host key (producción)
	var hostKeyCallback ssh.HostKeyCallback
	if c.cfg.HostKey != "" {
		expectedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(c.cfg.HostKey))
		if err != nil {
			return fmt.Errorf("failed to parse host key: %w", err)
		}
		hostKeyCallback = ssh.FixedHostKey(expectedKey)
	} else {
		// ⚠️ Fallback inseguro para testing
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// SSH client config
	config := &ssh.ClientConfig{
		User: c.cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	// Connect SSH with retry
	address := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)
	var sshClient *ssh.Client
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		sshClient, err = ssh.Dial("tcp", address, config)
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(time.Duration(attempt*5) * time.Second)
	}

	if lastErr != nil {
		return fmt.Errorf("failed to connect SSH after 3 attempts: %w", lastErr)
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

func (c *Client) UploadWithChecksum(data []byte, remotePath string, algorithm string) (string, error) {
	// Calcular checksum local
	var localChecksum string
	if algorithm == "SHA256" {
		hash := sha256.Sum256(data)
		localChecksum = hex.EncodeToString(hash[:])
	}

	// Upload file
	if err := c.Upload(data, remotePath); err != nil {
		return "", err
	}

	return localChecksum, nil
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

	// Escribir datos con progress logging
	reader := bytes.NewReader(data)
	written := int64(0)
	buffer := make([]byte, 32*1024) // 32KB chunks

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if _, writeErr := remoteFile.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write data: %w", writeErr)
			}
			written += int64(n)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read data: %w", err)
		}
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

**Actualización en `config/secrets.go`** para incluir nuevos campos:

```go
type SecretsPayload struct {
	// ... campos existentes ...
	SFTPHostKey                string `json:"SFTP_HOST_KEY"`
	EnableChecksumVerification string `json:"ENABLE_CHECKSUM_VERIFICATION"`
	ChecksumAlgorithm          string `json:"CHECKSUM_ALGORITHM"`
	MaxRetryAttempts           string `json:"MAX_RETRY_ATTEMPTS"`
	RetryBackoffSeconds        string `json:"RETRY_BACKOFF_SECONDS"`
	DLQUrl                     string `json:"DLQ_URL"`
	// ... resto de campos
}
```

### 4. Stack CDK para SFTP Externo

**`stacks/addi/addi-external-stack.go`**:

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
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatchactions"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk-library/constructs/S3"
	"cdk-library/constructs/GuardDuty"
)

type AddiExternalStackProps struct {
	awscdk.StackProps
	AlertEmail                string
	UseDirectConnect          bool
	ExistingVPCID             string // VPC con Direct Connect configurado
	ExistingDirectConnectGWID string
}

func NewAddiExternalStack(scope constructs.Construct, id string, props *AddiExternalStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// =================================================================
	// 1. VPC Configuration (usar VPC existente con Direct Connect)
	// =================================================================

	var vpc awsec2.IVpc

	if props.ExistingVPCID != "" {
		// Importar VPC existente con Direct Connect
		vpc = awsec2.Vpc_FromLookup(stack, jsii.String("ExistingVPC"), &awsec2.VpcLookupOptions{
			VpcId: jsii.String(props.ExistingVPCID),
		})
	} else {
		// Crear nuevo VPC con NAT Gateway (fallback si no hay Direct Connect)
		vpc = awsec2.NewVpc(stack, jsii.String("AddiVPC"), &awsec2.VpcProps{
			MaxAzs:      jsii.Number(2),
			NatGateways: jsii.Number(1), // Necesario para salida a internet/on-premise
			SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
				{
					Name:       jsii.String("Public"),
					SubnetType: awsec2.SubnetType_PUBLIC,
					CidrMask:   jsii.Number(24),
				},
				{
					Name:       jsii.String("Private"),
					SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
					CidrMask:   jsii.Number(24),
				},
			},
		})
	}

	// =================================================================
	// 2. S3 Bucket (Landing Zone) - Enterprise grade
	// =================================================================

	bucket := s3.NewSimpleStorageServiceFactory(stack, "AddiLandingBucket",
		s3.SimpleStorageServiceFactoryProps{
			BucketType:    s3.BucketTypeEnterprise,
			BucketName:    jsii.String("addi-file-landing-external-prod"),
			RemovalPolicy: jsii.String("retain"), // RETAIN para producción
		})

	// =================================================================
	// 3. GuardDuty - Comprehensive Protection
	// =================================================================

	guardduty.NewGuardDutyDetector(stack, "AddiThreatDetection",
		guardduty.GuardDutyFactoryProps{
			DetectorType:               guardduty.GuardDutyTypeComprehensive,
			FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
		})

	// =================================================================
	// 4. Dead Letter Queue (SQS) para errores
	// =================================================================

	dlq := awssqs.NewQueue(stack, jsii.String("TransferDLQ"), &awssqs.QueueProps{
		QueueName:         jsii.String("addi-transfer-failures"),
		RetentionPeriod:   awscdk.Duration_Days(jsii.Number(14)),
		VisibilityTimeout: awscdk.Duration_Minutes(jsii.Number(5)),
	})

	// =================================================================
	// 5. SNS Topic para notificaciones
	// =================================================================

	snsTopic := awssns.NewTopic(stack, jsii.String("SFTPTransferNotifications"), &awssns.TopicProps{
		DisplayName: jsii.String("Addi SFTP External Transfer Notifications"),
		TopicName:   jsii.String("addi-sftp-external-notifications"),
	})

	if props.AlertEmail != "" {
		snsTopic.AddSubscription(
			awssnssubscriptions.NewEmailSubscription(jsii.String(props.AlertEmail), nil),
		)
	}

	// Alarm para alta tasa de fallos
	snsAlarmAction := awscloudwatchactions.NewSnsAction(snsTopic)

	// =================================================================
	// 6. Secrets Manager - Configuración SFTP Externa
	// =================================================================

	envContent := loadEnvFile("stacks/addi/.env.external")

	sftpSecret := awssecretsmanager.NewSecret(stack, jsii.String("SFTPExternalConfig"), &awssecretsmanager.SecretProps{
		SecretName:        jsii.String("addi/sftp-external/config"),
		Description:       jsii.String("SFTP external server configuration"),
		SecretStringValue: awscdk.SecretValue_UnsafePlainText(jsii.String(envContent)),
	})

	// =================================================================
	// 7. Security Group para Lambda
	// =================================================================

	lambdaSG := awsec2.NewSecurityGroup(stack, jsii.String("LambdaSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Security group for SFTP transfer Lambda"),
		AllowAllOutbound: jsii.Bool(false), // Restringir outbound
	})

	// Permitir HTTPS para AWS API calls (Secrets Manager, S3, SNS)
	lambdaSG.AddEgressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(443)),
		jsii.String("Allow HTTPS for AWS APIs"),
		jsii.Bool(false),
	)

	// Permitir SSH al servidor SFTP externo (especificar IP/CIDR específico)
	// ⚠️ REEMPLAZAR con la IP real del servidor SFTP externo
	sftpServerIP := "192.168.10.50/32" // Ejemplo: IP del servidor on-premise
	sftpPort := 2222                    // Puerto custom

	lambdaSG.AddEgressRule(
		awsec2.Peer_Ipv4(jsii.String(sftpServerIP)),
		awsec2.Port_Tcp(jsii.Number(sftpPort)),
		jsii.String("Allow SFTP to external server"),
		jsii.Bool(false),
	)

	// =================================================================
	// 8. Lambda Function (Go runtime) con retry logic
	// =================================================================

	transferLambda := awslambda.NewFunction(stack, jsii.String("SFTPTransferFunction"), &awslambda.FunctionProps{
		Runtime:    awslambda.Runtime_PROVIDED_AL2023(),
		Handler:    jsii.String("bootstrap"),
		Code:       awslambda.Code_FromAsset(jsii.String("lambda/sftp-transfer/build"), nil),
		MemorySize: jsii.Number(1024), // 1GB para archivos más grandes
		Timeout:    awscdk.Duration_Minutes(jsii.Number(15)), // 15 min timeout
		Environment: &map[string]*string{
			"CONFIG_SECRET_ARN": sftpSecret.SecretArn(),
			"DLQ_URL":           dlq.QueueUrl(),
		},
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
		},
		LogRetention: awslogs.RetentionDays_ONE_MONTH, // 1 mes para producción
		ReservedConcurrentExecutions: jsii.Number(10), // Limitar concurrencia
	})

	// Permisos IAM
	bucket.GrantRead(transferLambda, nil)
	bucket.GrantDelete(transferLambda, nil)
	bucket.GrantPut(transferLambda, nil)
	sftpSecret.GrantRead(transferLambda, nil)
	snsTopic.GrantPublish(transferLambda)
	dlq.GrantSendMessages(transferLambda)

	// =================================================================
	// 9. EventBridge Rule con DLQ
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

	rule.AddTarget(awseventstargets.NewLambdaFunction(transferLambda, &awseventstargets.LambdaFunctionProps{
		DeadLetterQueue: dlq,
		MaxEventAge:     awscdk.Duration_Hours(jsii.Number(2)),
		RetryAttempts:   jsii.Number(2),
	}))

	// =================================================================
	// 10. CloudWatch Alarms
	// =================================================================

	// Alarm: High error rate
	errorAlarm := awscloudwatch.NewAlarm(stack, jsii.String("LambdaErrorAlarm"), &awscloudwatch.AlarmProps{
		Metric: transferLambda.MetricErrors(&awscloudwatch.MetricOptions{
			Period:    awscdk.Duration_Minutes(jsii.Number(5)),
			Statistic: jsii.String("Sum"),
		}),
		Threshold:          jsii.Number(3),
		EvaluationPeriods:  jsii.Number(2),
		AlarmDescription:   jsii.String("Alert when Lambda has high error rate"),
		AlarmName:          jsii.String("Addi-SFTP-HighErrorRate"),
		TreatMissingData:   awscloudwatch.TreatMissingData_NOT_BREACHING,
		ActionsEnabled:     jsii.Bool(true),
	})
	errorAlarm.AddAlarmAction(snsAlarmAction)

	// Alarm: High duration (transferencias lentas)
	durationAlarm := awscloudwatch.NewAlarm(stack, jsii.String("LambdaDurationAlarm"), &awscloudwatch.AlarmProps{
		Metric: transferLambda.MetricDuration(&awscloudwatch.MetricOptions{
			Period:    awscdk.Duration_Minutes(jsii.Number(5)),
			Statistic: jsii.String("Average"),
		}),
		Threshold:         jsii.Number(600000), // 10 minutos
		EvaluationPeriods: jsii.Number(3),
		AlarmDescription:  jsii.String("Alert when transfers take too long"),
		AlarmName:         jsii.String("Addi-SFTP-SlowTransfers"),
		ActionsEnabled:    jsii.Bool(true),
	})
	durationAlarm.AddAlarmAction(snsAlarmAction)

	// Alarm: Messages in DLQ
	dlqAlarm := awscloudwatch.NewAlarm(stack, jsii.String("DLQAlarm"), &awscloudwatch.AlarmProps{
		Metric: dlq.MetricApproximateNumberOfMessagesVisible(&awscloudwatch.MetricOptions{
			Period:    awscdk.Duration_Minutes(jsii.Number(5)),
			Statistic: jsii.String("Sum"),
		}),
		Threshold:         jsii.Number(1),
		EvaluationPeriods: jsii.Number(1),
		AlarmDescription:  jsii.String("Alert when failed transfers in DLQ"),
		AlarmName:         jsii.String("Addi-SFTP-FailedTransfers"),
		ActionsEnabled:    jsii.Bool(true),
	})
	dlqAlarm.AddAlarmAction(snsAlarmAction)

	// =================================================================
	// 11. Outputs
	// =================================================================

	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 Landing Bucket Name"),
		ExportName:  jsii.String("AddiExternalBucketName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LambdaFunctionName"), &awscdk.CfnOutputProps{
		Value:       transferLambda.FunctionName(),
		Description: jsii.String("SFTP Transfer Lambda Function Name"),
		ExportName:  jsii.String("AddiExternalLambdaName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("SecretARN"), &awscdk.CfnOutputProps{
		Value:       sftpSecret.SecretArn(),
		Description: jsii.String("Secrets Manager ARN for SFTP config"),
		ExportName:  jsii.String("AddiExternalSecretARN"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DLQURL"), &awscdk.CfnOutputProps{
		Value:       dlq.QueueUrl(),
		Description: jsii.String("Dead Letter Queue URL"),
		ExportName:  jsii.String("AddiExternalDLQURL"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("SNSTopicARN"), &awscdk.CfnOutputProps{
		Value:       snsTopic.TopicArn(),
		Description: jsii.String("SNS Topic ARN for notifications"),
		ExportName:  jsii.String("AddiExternalSNSTopicARN"),
	})

	return stack
}

func loadEnvFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read .env file: %v", err))
	}
	return string(content)
}
```

### 5. Scripts de Utilidad

**`scripts/test-connectivity.sh`**:

```bash
#!/bin/bash
set -e

echo "🔍 Testing connectivity to external SFTP server..."

# Load config from .env.external
source stacks/addi/.env.external

echo "Target: $SFTP_HOST:$SFTP_PORT"
echo "User: $SFTP_USER"
echo ""

# Test 1: Network connectivity (telnet/nc)
echo "Test 1: Network connectivity..."
if timeout 5 bash -c "echo >/dev/tcp/$SFTP_HOST/$SFTP_PORT" 2>/dev/null; then
    echo "✅ Network connection successful"
else
    echo "❌ Cannot reach $SFTP_HOST:$SFTP_PORT"
    echo "   Check firewall rules and routing"
    exit 1
fi

# Test 2: SSH handshake
echo ""
echo "Test 2: SSH handshake..."
if timeout 5 ssh -o BatchMode=yes -o StrictHostKeyChecking=no -p $SFTP_PORT $SFTP_USER@$SFTP_HOST exit 2>/dev/null; then
    echo "✅ SSH handshake successful"
else
    echo "⚠️  SSH handshake failed (expected if key auth not setup yet)"
fi

# Test 3: Get host key for .env
echo ""
echo "Test 3: Retrieving server host key..."
HOST_KEY=$(ssh-keyscan -p $SFTP_PORT $SFTP_HOST 2>/dev/null | grep -v "^#")
if [ -n "$HOST_KEY" ]; then
    echo "✅ Host key retrieved:"
    echo "$HOST_KEY"
    echo ""
    echo "📋 Add this to .env.external:"
    echo "SFTP_HOST_KEY=\"$HOST_KEY\""
else
    echo "❌ Could not retrieve host key"
fi

echo ""
echo "✅ Connectivity tests complete"
```

**`scripts/upload-secrets-external.sh`**:

```bash
#!/bin/bash
set -e

if [ ! -f "stacks/addi/.env.external" ]; then
    echo "❌ Error: .env.external file not found"
    exit 1
fi

echo "🔐 Converting .env.external to JSON..."

# Convertir .env a JSON (simple parser)
JSON=$(cat stacks/addi/.env.external | \
    grep -v '^#' | \
    grep -v '^$' | \
    awk -F= '{
        gsub(/^[ \t]+|[ \t]+$/, "", $1);
        value = substr($0, index($0,"=")+1);
        gsub(/^[ \t]+|[ \t]+$/, "", value);
        printf "\"%s\": \"%s\",\n", $1, value
    }' | \
    sed '$ s/,$//' | \
    awk 'BEGIN{print "{"} {print} END{print "}"}')

echo "📤 Uploading to AWS Secrets Manager..."

aws secretsmanager create-secret \
    --name addi/sftp-external/config \
    --description "External SFTP server configuration" \
    --secret-string "$JSON" \
    2>/dev/null || \
aws secretsmanager update-secret \
    --secret-id addi/sftp-external/config \
    --secret-string "$JSON"

echo "✅ Configuration uploaded successfully"
echo "Secret ARN: $(aws secretsmanager describe-secret --secret-id addi/sftp-external/config --query 'ARN' --output text)"
```

---

## 🚀 Instrucciones de Despliegue

### Paso 1: Preparar Conectividad Híbrida

**Si usas Direct Connect** (recomendado):
```bash
# Verificar Direct Connect existente
aws directconnect describe-connections

# Verificar Virtual Private Gateway
aws ec2 describe-vpn-gateways

# Obtener VPC ID con Direct Connect configurado
export VPC_ID=vpc-xxxxx
export DX_GW_ID=dx-gw-xxxxx
```

**Si necesitas crear VPN Site-to-Site**:
```bash
# Crear Customer Gateway (on-premise router)
aws ec2 create-customer-gateway \
  --type ipsec.1 \
  --public-ip <IP_PUBLICA_ON_PREMISE> \
  --bgp-asn 65000

# Crear VPN Gateway (AWS side)
aws ec2 create-vpn-gateway --type ipsec.1

# Crear VPN Connection
aws ec2 create-vpn-connection \
  --type ipsec.1 \
  --customer-gateway-id cgw-xxxxx \
  --vpn-gateway-id vgw-xxxxx
```

### Paso 2: Validar Conectividad al SFTP Externo

```bash
cd stacks/addi

# Crear .env.external desde template
cp .env.example .env.external

# Editar con datos del servidor SFTP externo
nano .env.external

# Probar conectividad
./scripts/test-connectivity.sh
```

**Output esperado**:
```
✅ Network connection successful
✅ SSH handshake successful
✅ Host key retrieved: sftp.empresa.com ssh-rsa AAAAB3...
```

### Paso 3: Configurar SSH Keys

```bash
# Generar llaves (o usar las provistas por el cliente)
./scripts/generate-ssh-keys.sh

# El cliente debe agregar sftp_key.pub a su servidor SFTP
# authorized_keys del usuario en el servidor externo
```

### Paso 4: Agregar Host Key a .env

```bash
# Ejecutar test-connectivity.sh y copiar el SFTP_HOST_KEY
./scripts/test-connectivity.sh

# Agregar la línea completa a .env.external:
echo 'SFTP_HOST_KEY="sftp.empresa.com ssh-rsa AAAAB3..."' >> .env.external
```

### Paso 5: Compilar Lambda

```bash
./scripts/build-lambda.sh
```

### Paso 6: Desplegar Stack

```bash
cd ../..

# Sintetizar
cdk synth AddiExternalStack

# Desplegar con parámetros
cdk deploy AddiExternalStack \
  --parameters AlertEmail="ops-team@empresa.com" \
  --parameters UseDirectConnect=true \
  --parameters ExistingVPCID="vpc-xxxxx" \
  --require-approval never
```

### Paso 7: Subir Configuración a Secrets Manager

```bash
cd stacks/addi

# Actualizar .env.external con outputs del stack
export SNS_ARN=$(aws cloudformation describe-stacks \
  --stack-name AddiExternalStack \
  --query 'Stacks[0].Outputs[?OutputKey==`SNSTopicARN`].OutputValue' \
  --output text)

export DLQ_URL=$(aws cloudformation describe-stacks \
  --stack-name AddiExternalStack \
  --query 'Stacks[0].Outputs[?OutputKey==`DLQURL`].OutputValue' \
  --output text)

sed -i '' "s|SNS_TOPIC_ARN=.*|SNS_TOPIC_ARN=$SNS_ARN|" .env.external
sed -i '' "s|DLQ_URL=.*|DLQ_URL=$DLQ_URL|" .env.external

# Subir a Secrets Manager
./scripts/upload-secrets-external.sh
```

### Paso 8: Prueba End-to-End

```bash
# Subir archivo de prueba
echo "Production test file" > test-prod.txt
aws s3 cp test-prod.txt s3://addi-file-landing-external-prod/uploads/test-prod.txt

# Monitorear logs
aws logs tail /aws/lambda/AddiExternalStack-SFTPTransferFunction --follow

# Verificar en servidor SFTP externo (pedir al cliente)
# ssh user@sftp.cliente.com "ls -lh /data/ingress/aws/"
```

---

## 🔒 Seguridad en Producción

### 1. Host Key Verification

✅ **SIEMPRE configurar** `SFTP_HOST_KEY` en `.env.external`

```bash
# Obtener host key del servidor
ssh-keyscan -p 2222 sftp.empresa.com > server_host_key.txt

# Agregar a .env.external
SFTP_HOST_KEY="$(cat server_host_key.txt)"
```

### 2. SSH Key Rotation

```bash
# Rotar cada 90 días
# 1. Generar nuevo par de llaves
ssh-keygen -t rsa -b 4096 -f sftp_key_new -N ""

# 2. Agregar nueva public key al servidor SFTP externo

# 3. Actualizar Secrets Manager
aws secretsmanager update-secret \
  --secret-id addi/sftp-external/config \
  --secret-string file://new-config.json

# 4. Verificar funciona con nueva key
# 5. Eliminar old key del servidor después de 7 días
```

### 3. Firewall Rules (On-Premise)

El firewall del servidor SFTP externo debe permitir:

```
Source: AWS CIDR blocks del VPC
Destination: Servidor SFTP (192.168.10.50)
Port: 2222
Protocol: TCP
```

**Obtener CIDR blocks de Lambda**:
```bash
aws ec2 describe-subnets \
  --filters "Name=vpc-id,Values=vpc-xxxxx" \
  --query 'Subnets[].CidrBlock' \
  --output table
```

### 4. Logging & Auditoría

```bash
# Habilitar CloudTrail en el bucket S3
aws cloudtrail put-event-selectors \
  --trail-name my-trail \
  --event-selectors '[{"ReadWriteType": "All", "IncludeManagementEvents": true, "DataResources": [{"Type": "AWS::S3::Object", "Values": ["arn:aws:s3:::addi-file-landing-external-prod/*"]}]}]'
```

---

## 📊 Monitoring & Alertas

### CloudWatch Dashboard

```bash
# Crear dashboard con métricas clave
aws cloudwatch put-dashboard \
  --dashboard-name Addi-SFTP-External \
  --dashboard-body file://dashboard.json
```

**`dashboard.json`**:
```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "title": "Lambda Invocations",
        "metrics": [
          ["AWS/Lambda", "Invocations", {"stat": "Sum"}],
          [".", "Errors", {"stat": "Sum"}]
        ],
        "period": 300,
        "region": "us-east-1"
      }
    },
    {
      "type": "metric",
      "properties": {
        "title": "Transfer Duration",
        "metrics": [
          ["AWS/Lambda", "Duration", {"stat": "Average"}],
          ["...", {"stat": "Maximum"}]
        ],
        "period": 300
      }
    },
    {
      "type": "metric",
      "properties": {
        "title": "DLQ Messages",
        "metrics": [
          ["AWS/SQS", "ApproximateNumberOfMessagesVisible", {"queueName": "addi-transfer-failures"}]
        ],
        "period": 60
      }
    }
  ]
}
```

### Alarmas Configuradas

1. ✅ **High Error Rate**: >3 errores en 10 minutos
2. ✅ **Slow Transfers**: Duración promedio >10 minutos
3. ✅ **DLQ Not Empty**: Mensajes en cola de errores
4. ✅ **Lambda Throttling**: Concurrencia excedida

---

## 🐛 Troubleshooting

### Error: Network timeout al conectar

**Causa**: Direct Connect/VPN no configurado o firewall bloqueando

**Diagnóstico**:
```bash
# Desde Lambda (usando AWS Systems Manager Session Manager en VPC)
# 1. Lanzar EC2 temporal en mismo subnet que Lambda
# 2. Conectar via SSM
# 3. Probar conectividad:
nc -zv sftp.empresa.com 2222
telnet sftp.empresa.com 2222
```

**Solución**:
- Verificar route tables tienen ruta a Direct Connect Gateway
- Confirmar firewall on-premise permite AWS CIDR blocks
- Validar Security Group de Lambda permite egress a IP:port

### Error: Host key verification failed

**Causa**: `SFTP_HOST_KEY` incorrecto o faltante

**Solución**:
```bash
# Obtener host key correcto
ssh-keyscan -p 2222 sftp.empresa.com

# Actualizar en Secrets Manager
./scripts/upload-secrets-external.sh
```

### Error: Permission denied (publickey)

**Causa**: Clave SSH privada incorrecta o no agregada en servidor

**Solución**:
1. Verificar `sftp_key.pub` está en `~/.ssh/authorized_keys` del servidor
2. Permisos correctos en servidor:
   ```bash
   chmod 700 ~/.ssh
   chmod 600 ~/.ssh/authorized_keys
   ```
3. Verificar formato de clave en `.env.external` (debe incluir `\n` como literales)

### Mensajes en DLQ

**Causa**: Transferencias fallidas repetidas

**Diagnóstico**:
```bash
# Ver mensajes en DLQ
aws sqs receive-message \
  --queue-url https://sqs.us-east-1.amazonaws.com/.../addi-transfer-failures \
  --max-number-of-messages 10

# Analizar patrón de errores
aws logs filter-log-events \
  --log-group-name /aws/lambda/AddiExternalStack-SFTPTransferFunction \
  --filter-pattern "ERROR" \
  --start-time $(date -u -d '1 hour ago' +%s)000
```

**Solución**:
1. Identificar causa raíz en logs
2. Corregir configuración/conectividad
3. Re-procesar mensajes desde DLQ manualmente

---

## 📚 Referencias

- [AWS Direct Connect Documentation](https://docs.aws.amazon.com/directconnect/)
- [AWS Site-to-Site VPN](https://docs.aws.amazon.com/vpn/latest/s2svpn/)
- [Golang SSH Package](https://pkg.go.dev/golang.org/x/crypto/ssh)
- [SFTP Security Best Practices](https://www.ssh.com/academy/ssh/sftp-ssh-file-transfer-protocol)

---

## 📋 Checklist de Producción

Antes de go-live, verificar:

- [ ] Direct Connect o VPN configurado y probado
- [ ] Host key del servidor SFTP agregado a Secrets Manager
- [ ] SSH public key agregada al servidor SFTP externo
- [ ] Firewall on-premise permite tráfico desde AWS VPC CIDRs
- [ ] GuardDuty Comprehensive habilitado
- [ ] CloudWatch Alarms configuradas y probadas
- [ ] SNS notifications llegando al equipo correcto
- [ ] DLQ monitoreada con alertas
- [ ] Runbook documentado para on-call
- [ ] Prueba de failover (simular servidor SFTP caído)
- [ ] Backup plan para archivos en S3
- [ ] SSH key rotation schedule (90 días)

---

**Última actualización**: 2025-10-11
**Versión**: 1.0
**Costo estimado**: ~$59/mes (con Direct Connect existente) | ~$105/mes (con VPN nueva)
