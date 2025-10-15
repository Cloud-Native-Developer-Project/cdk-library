# Addi - S3 to SFTP Pipeline

Sistema de transferencia automatizada de archivos CSV desde AWS S3 hacia un servidor SFTP on-premise, utilizando arquitectura serverless y event-driven.

---

## 📋 Tabla de Contenidos

- [Arquitectura General](#-arquitectura-general)
- [1. S3 Bucket - Configuración y Accesibilidad](#1-s3-bucket---configuración-y-accesibilidad)
- [2. EventBridge - Configuración y Funcionalidad](#2-eventbridge---configuración-y-funcionalidad)
- [3. Lambda Function - Configuración y Petición de Salida](#3-lambda-function---configuración-y-petición-de-salida)
- [4. Backend API - Receptor y Procesamiento](#4-backend-api---receptor-y-procesamiento)
- [5. ngrok - Túnel Público (Opcional)](#5-ngrok---túnel-público-opcional)
- [6. SFTP - Destino Final](#6-sftp---destino-final)
- [Diagramas de Secuencia Completos](#-diagramas-de-secuencia-completos)
- [Seguridad y Observabilidad](#-seguridad-y-observabilidad)
- [Costos Estimados](#-costos-estimados)
- [Infraestructura Congelada](#-infraestructura-congelada)

---

## 🏗 Arquitectura General

```mermaid
graph LR
    Client[Cliente con IAM User]
    S3[S3 Bucket<br/>addi-landing-zone-dev]
    EB[EventBridge Rule]
    Lambda[Lambda Function<br/>Go ARM64]
    SM[Secrets Manager]
    Ngrok[ngrok Tunnel<br/>OPCIONAL]
    Backend[Backend API<br/>Go Docker]
    SFTP[SFTP Server<br/>atmoz/sftp]
    DLQ[SQS DLQ]
    GD[GuardDuty]

    Client -->|PUT Object<br/>IAM Credentials| S3
    S3 -->|ObjectCreated Event| EB
    EB -->|Trigger| Lambda
    Lambda -->|Read Secrets| SM
    Lambda -->|Generate Presigned URL| S3
    Lambda -->|POST Webhook<br/>with Presigned URL| Ngrok
    Ngrok -->|Forward Request| Backend
    Backend -->|HTTP GET<br/>No AWS Credentials| S3
    Backend -->|SFTP Upload| SFTP
    Lambda -.->|On Failure| DLQ
    GD -.->|Monitor| S3
    GD -.->|Monitor| Lambda

    style S3 fill:#FF9900
    style Lambda fill:#FF9900
    style Backend fill:#4CAF50
    style SFTP fill:#4CAF50
    style Ngrok fill:#FFC107
    style DLQ fill:#F44336
    style GD fill:#2196F3
```

### Características Clave

- **Event-Driven:** Disparado automáticamente al subir archivos
- **Serverless:** Lambda ARM64 con Go runtime custom
- **Secure:** Presigned URLs (sin credenciales AWS en backend)
- **Resilient:** SQS DLQ con 4 reintentos exponenciales
- **Observable:** CloudWatch Logs + X-Ray tracing
- **Cost-Optimized:** S3 Development + ARM64 Lambda

---

## 1. S3 Bucket - Configuración y Accesibilidad

### 📦 Configuración Técnica

**Bucket:** `addi-landing-zone-dev`
**Estrategia:** Development (ver `constructs/S3/simple_storage_service_development.go`)

```mermaid
graph TB
    subgraph "S3 Bucket Configuration"
        Config[Bucket Configuration]
        Config --> Encryption[S3_MANAGED AES-256]
        Config --> TLS[TLS 1.2 Minimum]
        Config --> Public[Public Access BLOCKED]
        Config --> Versioning[Versioning: Disabled]
        Config --> Lifecycle[30-day Expiration]
        Config --> EventBridge[EventBridge Enabled]
    end

    subgraph "Security Features"
        IAM[IAM Bucket Policy]
        IAM --> LambdaRead[Lambda: s3:GetObject uploads/*]
        IAM --> ClientPut[Client IAM User: s3:PutObject uploads/*]
    end

    Config --> IAM
```

### 🔐 Acceso al Bucket (Cliente)

El bucket **NO es público**. El cliente debe autenticarse con credenciales IAM para subir archivos.

#### Opción 1: IAM User con Permisos Limitados (RECOMENDADO)

**Paso 1: Crear IAM User**

```bash
# Crear usuario dedicado
aws iam create-user --user-name addi-uploader

# Crear política con permisos mínimos (solo PUT en uploads/)
aws iam put-user-policy \
  --user-name addi-uploader \
  --policy-name S3UploadOnlyPolicy \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": "arn:aws:s3:::addi-landing-zone-dev/uploads/*"
    }]
  }'

# Generar Access Keys
aws iam create-access-key --user-name addi-uploader
```

**Salida esperada:**
```json
{
  "AccessKey": {
    "UserName": "addi-uploader",
    "AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
    "SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "Status": "Active",
    "CreateDate": "2025-10-15T00:00:00Z"
  }
}
```

**Paso 2: Cliente Configura Credenciales**

```bash
# Opción A: Variables de entorno
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_DEFAULT_REGION="us-east-1"

# Opción B: Archivo de configuración
aws configure --profile addi-uploader
# AWS Access Key ID: AKIAIOSFODNN7EXAMPLE
# AWS Secret Access Key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
# Default region name: us-east-1
# Default output format: json
```

**Paso 3: Cliente Sube Archivos**

```bash
# Usando AWS CLI
aws s3 cp archivo.csv s3://addi-landing-zone-dev/uploads/

# Usando SDK (Python)
import boto3
s3 = boto3.client('s3')
s3.upload_file('archivo.csv', 'addi-landing-zone-dev', 'uploads/archivo.csv')

# Usando SDK (Go)
import "github.com/aws/aws-sdk-go-v2/service/s3"
s3Client := s3.NewFromConfig(cfg)
s3Client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: aws.String("addi-landing-zone-dev"),
    Key:    aws.String("uploads/archivo.csv"),
    Body:   file,
})
```

### 📊 Estructura del Bucket

```
s3://addi-landing-zone-dev/
└── uploads/
    ├── archivo1.csv      ← Cliente sube aquí (dispara pipeline)
    ├── archivo2.csv
    └── data-2025.csv
```

**⚠️ IMPORTANTE:**
- Solo archivos en `uploads/` disparan EventBridge
- Otros prefijos son ignorados
- Archivos expiran automáticamente después de 30 días (lifecycle)

---

## 2. EventBridge - Configuración y Funcionalidad

### ⚡ Diagrama de Flujo EventBridge

```mermaid
sequenceDiagram
    participant S3 as S3 Bucket
    participant EB as EventBridge
    participant Lambda as Lambda Function
    participant DLQ as Dead Letter Queue

    S3->>EB: ObjectCreated Event
    Note over EB: Filter: prefix="uploads/"

    alt Event matches filter
        EB->>Lambda: Invoke (Attempt 1)
        alt Lambda Success
            Lambda-->>EB: 200 OK
        else Lambda Failure
            Lambda-->>EB: Error 500
            Note over EB: Wait 1s (2^0)
            EB->>Lambda: Invoke (Attempt 2)
            alt Lambda Success
                Lambda-->>EB: 200 OK
            else Lambda Failure
                Lambda-->>EB: Error 500
                Note over EB: Wait 2s (2^1)
                EB->>Lambda: Invoke (Attempt 3)
                alt Lambda Success
                    Lambda-->>EB: 200 OK
                else Lambda Failure
                    Lambda-->>EB: Error 500
                    Note over EB: Wait 4s (2^2)
                    EB->>Lambda: Invoke (Attempt 4)
                    alt Lambda Success
                        Lambda-->>EB: 200 OK
                    else Lambda Failure (Final)
                        Lambda-->>EB: Error 500
                        Note over EB: Max retries exceeded
                        EB->>DLQ: Send event to DLQ
                    end
                end
            end
        end
    else Event ignored
        Note over EB: Prefix != "uploads/"
    end
```

### 🔧 Configuración Técnica

```go
// En constructs/EventBridgeIntegrations/eventbridge_s3_to_lambda.go
EventPattern: {
    "source": ["aws.s3"],
    "detail-type": ["Object Created"],
    "detail": {
        "bucket": {
            "name": ["addi-landing-zone-dev"]
        },
        "object": {
            "key": [{
                "prefix": "uploads/"
            }]
        }
    }
}

RetryPolicy: {
    MaximumRetryAttempts: 4,
    MaximumEventAge: 900 (15 minutos),
    BackoffRate: 2 (exponencial)
}
```

### 📋 Payload del Evento S3

EventBridge transforma el evento S3 y lo envía a Lambda:

```json
{
    "version": "0",
    "id": "c12f2c5d-70a5-7053-7c9f-443011568456",
    "detail-type": "Object Created",
    "source": "aws.s3",
    "account": "123456789012",
    "time": "2025-10-15T00:35:29Z",
    "region": "us-east-1",
    "resources": [
        "arn:aws:s3:::addi-landing-zone-dev"
    ],
    "detail": {
        "version": "0",
        "bucket": {
            "name": "addi-landing-zone-dev"
        },
        "object": {
            "key": "uploads/test-1760488526.csv",
            "size": 114,
            "etag": "6f44aae4dd43bbc94690f6d1646d4b8e",
            "sequencer": "00635F0B3A..."
        },
        "request-id": "ABC123DEF456",
        "requester": "123456789012",
        "source-ip-address": "192.0.2.1"
    }
}
```

---

## 3. Lambda Function - Configuración y Petición de Salida

### 🚀 Configuración de Lambda

**Función:** `addi-webhook-notifier`
**Runtime:** `provided.al2` (Go Custom Runtime)
**Architecture:** ARM64
**Handler:** `bootstrap` (ejecutable compilado)

```mermaid
graph TB
    subgraph "Lambda Configuration"
        Runtime[Go 1.x Custom Runtime]
        Arch[ARM64 Graviton2]
        Memory[512 MB]
        Timeout[30 segundos]
        Tracing[X-Ray Enabled]
    end

    subgraph "Environment Variables"
        ENV1[BUCKET_NAME]
        ENV2[WEBHOOK_SECRET_ARN]
        ENV3[PRESIGNED_URL_EXPIRES: 900]
        ENV4[MAX_RETRY_ATTEMPTS: 4]
        ENV5[WEBHOOK_URL_OVERRIDE]
    end

    subgraph "IAM Permissions"
        S3Read[s3:GetObject uploads/*]
        SMRead[secretsmanager:GetSecretValue]
        Logs[logs:CreateLogStream/PutLogEvents]
        XRay[xray:PutTraceSegments]
    end

    Runtime --> Arch
    Arch --> Memory
    Memory --> Timeout
    Timeout --> Tracing
```

### 📤 Proceso de Lambda (Diagrama Detallado)

```mermaid
sequenceDiagram
    participant EB as EventBridge
    participant Lambda as Lambda Function
    participant S3 as S3 API
    participant SM as Secrets Manager
    participant Webhook as Webhook Endpoint

    EB->>Lambda: Invoke with S3 Event
    activate Lambda

    Note over Lambda: Parse Event<br/>{bucket, key, size, etag}

    Lambda->>S3: GeneratePresignedURL()
    Note over Lambda,S3: bucket: addi-landing-zone-dev<br/>key: uploads/file.csv<br/>expires: 900s
    S3-->>Lambda: Presigned URL (valid 15min)

    Lambda->>SM: GetSecretValue()
    Note over Lambda,SM: Secret: addi/webhook-credentials
    SM-->>Lambda: {webhookUrl, apiKey, hmacSecret}

    Note over Lambda: Build Webhook Payload<br/>{eventId, bucket, key,<br/>presignedUrl, expiresAt}

    Lambda->>Webhook: POST /webhook/addi-csv
    Note over Lambda,Webhook: Headers:<br/>Content-Type: application/json<br/>X-API-Key: addi_prod_ak_...<br/>X-HMAC-Signature: sha256(body)

    alt Webhook Success
        Webhook-->>Lambda: 200 OK
        Note over Lambda: Log success & return
    else Webhook Failure
        Webhook-->>Lambda: 500 Error
        Note over Lambda: Lambda throws error<br/>EventBridge retries
    end

    deactivate Lambda
```

### 📨 Payload de Salida (Webhook Request)

Lambda genera el siguiente payload HTTP POST:

**URL:** `https://[ngrok-url]/webhook/addi-csv` (desde Secrets Manager o env var override)

**Headers:**
```http
POST /webhook/addi-csv HTTP/1.1
Host: 360b6fbd06bc.ngrok-free.app
Content-Type: application/json
X-API-Key: addi_prod_ak_placeholder
X-HMAC-Signature: sha256=a1b2c3d4e5f6...
User-Agent: AWS-Lambda-Webhook-Notifier/1.0
```

**Body:**
```json
{
  "eventId": "c12f2c5d-70a5-7053-7c9f-443011568456",
  "timestamp": "2025-10-15T00:35:29Z",
  "bucket": "addi-landing-zone-dev",
  "key": "uploads/test-1760488526.csv",
  "size": 114,
  "etag": "6f44aae4dd43bbc94690f6d1646d4b8e",
  "presignedUrl": "https://addi-landing-zone-dev.s3.us-east-1.amazonaws.com/uploads/test-1760488526.csv?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20251015%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20251015T003529Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=abcdef1234567890...",
  "expiresAt": "2025-10-15T00:50:29Z"
}
```

**Campos clave:**
- `presignedUrl`: URL firmada válida por 15 minutos (no requiere credenciales AWS)
- `expiresAt`: Timestamp ISO 8601 de expiración
- `eventId`: UUID único para tracking

---

## 4. Backend API - Receptor y Procesamiento

### 🖥️ Arquitectura del Backend

```mermaid
graph TB
    subgraph "Docker Compose Stack"
        API[Backend API<br/>Go :8080]
        SFTP[SFTP Server<br/>atmoz/sftp :22]
        Ngrok[ngrok Tunnel<br/>:4040]
    end

    subgraph "Backend API Components"
        Handler[Webhook Handler<br/>/webhook/addi-csv]
        S3Service[S3 Service<br/>HTTP Client]
        SFTPService[SFTP Service<br/>ssh/sftp client]
        Processor[Webhook Processor<br/>Orchestrator]
    end

    Ngrok --> API
    API --> Handler
    Handler --> Processor
    Processor --> S3Service
    Processor --> SFTPService
    SFTPService --> SFTP

    style API fill:#4CAF50
    style SFTP fill:#4CAF50
    style Ngrok fill:#FFC107
```

### 🔄 Proceso de Backend (Detallado)

```mermaid
sequenceDiagram
    participant Lambda as Lambda Function
    participant Ngrok as ngrok Tunnel
    participant API as Backend API
    participant S3 as S3 (Presigned URL)
    participant SFTP as SFTP Server

    Lambda->>Ngrok: POST /webhook/addi-csv
    Note over Lambda,Ngrok: Headers: X-API-Key, X-HMAC-Signature<br/>Body: {presignedUrl, bucket, key, ...}

    Ngrok->>API: Forward Request
    activate API

    Note over API: 1. Validate Headers<br/>Check X-API-Key<br/>Verify HMAC signature

    alt Invalid Authentication
        API-->>Ngrok: 401 Unauthorized
        Ngrok-->>Lambda: 401 Unauthorized
    else Valid Authentication
        Note over API: 2. Parse Webhook Payload<br/>Extract presignedUrl, bucket, key

        API->>S3: HTTP GET (Presigned URL)
        Note over API,S3: NO AWS credentials needed<br/>URL includes signature
        S3-->>API: File data (streaming)

        Note over API: 3. Store in memory<br/>or temp file

        API->>SFTP: Connect(host:22, user, pass)
        SFTP-->>API: SSH Connection OK

        Note over API: 4. Generate remote path<br/>/uploads/YYYY/MM/DD/file.csv

        API->>SFTP: Upload file
        Note over API,SFTP: Stream file data<br/>Progress tracking
        SFTP-->>API: Upload Success

        API->>SFTP: Close Connection

        Note over API: 5. Log success metrics<br/>Duration: 9.7ms<br/>Bytes: 114

        API-->>Ngrok: 200 OK
        Note over API,Ngrok: Body: {status: "success",<br/>file: "uploads/...",<br/>processedAt: "..."}
        Ngrok-->>Lambda: 200 OK
    end

    deactivate API
```

### 📝 Backend API - Endpoints

#### POST /webhook/addi-csv

**Responsabilidad:** Procesar eventos S3 y transferir archivos a SFTP

**Request Headers:**
```http
Content-Type: application/json
X-API-Key: addi_prod_ak_placeholder
X-HMAC-Signature: sha256=abc123...
```

**Request Body:** (Ver sección Lambda - Payload de Salida)

**Response Success (200 OK):**
```json
{
  "status": "success",
  "message": "File transferred to SFTP successfully in 9.726083ms",
  "file": "uploads/test-1760488526.csv",
  "processed_at": "2025-10-15T00:35:30Z"
}
```

**Response Error (500 Internal Server Error):**
```json
{
  "status": "error",
  "message": "failed to download file from S3: failed to download file from presigned URL: connection timeout"
}
```

#### GET /health

**Responsabilidad:** Health check del servicio

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-15T00:35:00Z",
  "services": {
    "sftp": "connected",
    "s3": "available"
  }
}
```

### 🔐 Descarga desde S3 (Sin Credenciales AWS)

**Implementación (`internal/services/s3_service.go`):**

```go
// DownloadFileFromPresignedURL - HTTP GET puro, sin AWS SDK
func (s *S3ServiceImpl) DownloadFileFromPresignedURL(ctx context.Context, presignedURL string) (io.ReadCloser, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, presignedURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request: %w", err)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to download file: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        resp.Body.Close()
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
    }

    return resp.Body, nil
}
```

**Ventajas:**
- ✅ No requiere AWS SDK en backend
- ✅ No requiere credenciales AWS
- ✅ Reduce dependencias y tamaño del binario
- ✅ Funciona desde cualquier red (sin VPC peering)

---

## 5. ngrok - Túnel Público (Opcional)

### 🌐 Configuración de ngrok

ngrok expone el backend local (puerto 8080) a internet con HTTPS.

```mermaid
graph LR
    Internet[Internet<br/>AWS Lambda]
    NgrokCloud[ngrok Cloud<br/>*.ngrok-free.app]
    NgrokAgent[ngrok Agent<br/>Docker Container]
    Backend[Backend API<br/>localhost:8080]

    Internet -->|HTTPS Request| NgrokCloud
    NgrokCloud -->|Encrypted Tunnel| NgrokAgent
    NgrokAgent -->|HTTP| Backend

    style NgrokCloud fill:#FFC107
    style NgrokAgent fill:#FFC107
```

### 📋 Docker Compose Configuration

```yaml
services:
  ngrok:
    image: ngrok/ngrok:latest
    restart: unless-stopped
    command:
      - "http"
      - "api:8080"
    environment:
      - NGROK_AUTHTOKEN=${NGROK_AUTHTOKEN}
    ports:
      - "4040:4040"  # ngrok Web UI
    depends_on:
      - api
```

### 🔗 Obtener URL de ngrok

```bash
# Opción 1: Web UI
open http://localhost:4040

# Opción 2: API
curl http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url'

# Opción 3: Script automatizado
./stacks/addi/backend/get-ngrok-url.sh
```

**Output ejemplo:**
```
https://360b6fbd06bc.ngrok-free.app
```

### ⚠️ Limitaciones de ngrok Free Tier

- URL pública cambia cada vez que reinicia el contenedor
- Requiere actualizar `WEBHOOK_URL_OVERRIDE` en Lambda
- Límite de 40 requests/minuto (suficiente para dev/test)

### 🔄 Alternativas para Producción

**No usar ngrok en producción.** Alternativas recomendadas:

1. **VPN Site-to-Site:** AWS VPN → On-premise
2. **Direct Connect:** Conexión dedicada AWS ↔ Datacenter
3. **AWS PrivateLink:** Exponer backend como VPC Endpoint Service
4. **Reverse Proxy público:** Nginx/HAProxy con IP estática
5. **API Gateway + VPC Link:** Backend en VPC privado

---

## 6. SFTP - Destino Final

### 📁 Configuración del Servidor SFTP

**Imagen Docker:** `atmoz/sftp`
**Puerto:** 2222 (mapeado desde 22)
**Usuario:** `addiuser`
**Password:** `addipass`
**Home Directory:** `/home/addiuser/uploads`

```mermaid
graph TB
    subgraph "SFTP Server Configuration"
        Container[Docker Container<br/>atmoz/sftp]
        User[User: addiuser<br/>UID: 1001, GID: 100]
        HomeDir[Home: /home/addiuser]
        Uploads[uploads/<br/>chroot directory]
        Volume[Docker Volume<br/>./sftp-data/uploads]
    end

    Container --> User
    User --> HomeDir
    HomeDir --> Uploads
    Uploads --> Volume

    subgraph "File Organization"
        Root[/home/addiuser/uploads/]
        Root --> Year[2025/]
        Year --> Month[10/]
        Month --> Day[15/]
        Day --> File[test-1760488526.csv]
    end

    Uploads --> Root

    style Container fill:#4CAF50
    style Volume fill:#2196F3
```

### 🔐 Autenticación SFTP

**Método:** Password authentication
**Credenciales:** Configuradas en `docker-compose.yml`

```yaml
services:
  sftp:
    image: atmoz/sftp:latest
    ports:
      - "2222:22"
    volumes:
      - ./sftp-data/uploads:/home/addiuser/uploads
    command: addiuser:addipass:1001:100:uploads
    #         └─user    └─pass  └─uid └─gid └─chroot dir
```

### 📂 Estructura de Archivos

Los archivos se organizan automáticamente por fecha (año/mes/día):

```
/home/addiuser/uploads/
└── 2025/
    └── 10/
        └── 15/
            ├── test-1760488526.csv (114 bytes)
            ├── data-20251015-001.csv
            └── transaction-batch.csv
```

**Lógica de path** (en `internal/services/webhook_processor.go`):

```go
func generateRemotePath(originalKey string, timestamp time.Time) string {
    fileName := filepath.Base(originalKey)              // "test-1760488526.csv"
    datePrefix := timestamp.Format("2006/01/02")       // "2025/10/15"
    return filepath.Join(datePrefix, fileName)         // "2025/10/15/test-1760488526.csv"
}
```

### 🔌 Conexión desde Backend API

```go
// internal/services/sftp_service.go
func (s *SFTPServiceImpl) Connect(ctx context.Context) error {
    config := &ssh.ClientConfig{
        User: "addiuser",
        Auth: []ssh.AuthMethod{
            ssh.Password("addipass"),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Dev only
        Timeout:         10 * time.Second,
    }

    conn, err := ssh.Dial("tcp", "sftp:22", config)
    if err != nil {
        return fmt.Errorf("failed to connect to SFTP: %w", err)
    }

    client, err := sftp.NewClient(conn)
    if err != nil {
        conn.Close()
        return fmt.Errorf("failed to create SFTP client: %w", err)
    }

    s.conn = conn
    s.client = client
    return nil
}
```

### 🚀 Upload Process

```go
func (s *SFTPServiceImpl) UploadFile(ctx context.Context, reader io.Reader, remotePath string, size int64) (*SFTPTransferResult, error) {
    start := time.Now()

    // Create directory structure (2025/10/15/)
    dir := filepath.Dir(remotePath)
    s.client.MkdirAll(dir)

    // Create remote file
    remoteFile, err := s.client.Create(remotePath)
    if err != nil {
        return nil, fmt.Errorf("failed to create remote file: %w", err)
    }
    defer remoteFile.Close()

    // Stream file content
    written, err := io.Copy(remoteFile, reader)
    if err != nil {
        return nil, fmt.Errorf("failed to upload file: %w", err)
    }

    return &SFTPTransferResult{
        Success:           true,
        RemotePath:        filepath.Join("/uploads", remotePath),
        BytesTransferred:  written,
        Duration:          time.Since(start),
    }, nil
}
```

### 🔍 Verificar Archivos SFTP

```bash
# Opción 1: Desde el host (Docker volume)
ls -lh stacks/addi/backend/sftp-data/uploads/2025/10/15/

# Opción 2: Cliente SFTP
sftp -P 2222 addiuser@localhost
sftp> cd uploads/2025/10/15
sftp> ls -lh
sftp> get test-1760488526.csv
sftp> bye

# Opción 3: Docker exec
docker exec addi-backend-sftp-1 ls -lh /home/addiuser/uploads/2025/10/15/
```

---

## 📊 Diagramas de Secuencia Completos

### 1. Flujo End-to-End Completo

```mermaid
sequenceDiagram
    actor Client as Cliente
    participant S3 as S3 Bucket
    participant EB as EventBridge
    participant Lambda as Lambda Function
    participant SM as Secrets Manager
    participant Ngrok as ngrok Tunnel
    participant Backend as Backend API
    participant SFTP as SFTP Server

    Note over Client: Configure IAM credentials
    Client->>S3: PUT uploads/file.csv
    Note over Client,S3: Authentication: IAM User<br/>Access Key + Secret Key

    S3->>EB: ObjectCreated Event
    Note over EB: Filter: prefix="uploads/"

    EB->>Lambda: Invoke Function
    activate Lambda

    Lambda->>S3: GeneratePresignedURL(bucket, key, 900s)
    S3-->>Lambda: Presigned URL (valid 15 min)

    Lambda->>SM: GetSecretValue("addi/webhook-credentials")
    SM-->>Lambda: {webhookUrl, apiKey, hmacSecret}

    Lambda->>Ngrok: POST /webhook/addi-csv
    Note over Lambda,Ngrok: Headers: X-API-Key, X-HMAC-Signature<br/>Body: {presignedUrl, bucket, key, size, etag}

    Ngrok->>Backend: Forward Request (HTTP)
    activate Backend

    Backend->>Backend: Validate HMAC signature

    Backend->>S3: HTTP GET (Presigned URL)
    Note over Backend,S3: No AWS credentials needed
    S3-->>Backend: File data (streaming)

    Backend->>SFTP: Connect (SSH)
    SFTP-->>Backend: Connection OK

    Backend->>SFTP: MkdirAll("/uploads/2025/10/15")
    Backend->>SFTP: Create file & Upload data
    SFTP-->>Backend: Upload Success (114 bytes, 9.7ms)

    Backend->>SFTP: Close connection

    Backend-->>Ngrok: 200 OK {status: "success"}
    deactivate Backend

    Ngrok-->>Lambda: 200 OK
    deactivate Lambda

    Lambda->>EB: Success
    EB->>EB: Event completed
```

### 2. Flujo con Reintentos y DLQ

```mermaid
sequenceDiagram
    participant S3 as S3 Bucket
    participant EB as EventBridge
    participant Lambda as Lambda Function
    participant Backend as Backend API
    participant DLQ as SQS Dead Letter Queue
    participant CW as CloudWatch Logs

    S3->>EB: ObjectCreated Event

    loop Retry Loop (max 4 attempts)
        EB->>Lambda: Invoke (Attempt N)
        activate Lambda

        alt Backend Available
            Lambda->>Backend: POST /webhook/addi-csv
            Backend-->>Lambda: 200 OK
            deactivate Lambda
            Lambda->>CW: Log success
            Note over EB: Exit retry loop
        else Backend Unavailable/Error
            Lambda->>Backend: POST /webhook/addi-csv
            Backend--xLambda: 500 Error / Timeout
            deactivate Lambda
            Lambda->>CW: Log error (attempt N)

            alt Attempts < 4
                Note over EB: Wait 2^(N-1) seconds<br/>Attempt 1: 1s<br/>Attempt 2: 2s<br/>Attempt 3: 4s<br/>Attempt 4: 8s
            else Attempts >= 4
                Note over EB: Max retries exceeded
                EB->>DLQ: Send event to Dead Letter Queue
                EB->>CW: Log DLQ event
                Note over DLQ: Event stored for 14 days<br/>Manual intervention required
            end
        end
    end
```

### 3. Autenticación y Seguridad

```mermaid
sequenceDiagram
    actor Client as Cliente
    participant IAM as AWS IAM
    participant S3 as S3 Bucket
    participant Lambda as Lambda Function
    participant SM as Secrets Manager
    participant Backend as Backend API

    Note over Client,Backend: === Client Upload Authentication ===

    Client->>IAM: Request STS credentials
    IAM-->>Client: Access Key + Secret Key + Session Token

    Client->>S3: PUT Object (signed request)
    Note over Client,S3: Signature: AWS4-HMAC-SHA256
    S3->>IAM: Validate credentials
    IAM-->>S3: Authorized
    S3-->>Client: 200 OK

    Note over Client,Backend: === Lambda → Backend Authentication ===

    S3->>Lambda: Trigger via EventBridge
    Lambda->>SM: GetSecretValue()
    SM-->>Lambda: hmacSecret

    Lambda->>Lambda: Generate HMAC-SHA256<br/>HMAC(body, hmacSecret)

    Lambda->>Backend: POST /webhook/addi-csv
    Note over Lambda,Backend: Header: X-HMAC-Signature: sha256=abc123...

    Backend->>Backend: Compute HMAC(body, stored_secret)
    Backend->>Backend: Compare signatures<br/>(constant-time comparison)

    alt Valid Signature
        Backend->>S3: HTTP GET (Presigned URL)
        Note over Backend,S3: URL signature validated by S3
        S3-->>Backend: File data
    else Invalid Signature
        Backend-->>Lambda: 401 Unauthorized
    end
```

---

## 🔐 Seguridad y Observabilidad

### Capas de Seguridad

#### 1. Cliente → S3
- ✅ IAM User con política de least privilege
- ✅ TLS 1.2+ enforcement
- ✅ Public access blocked
- ✅ S3_MANAGED encryption (AES-256)

#### 2. Lambda → Backend
- ✅ HMAC-SHA256 signature validation
- ✅ API Key authentication
- ✅ Secrets Manager para credenciales
- ✅ X-Ray tracing habilitado

#### 3. Backend → S3
- ✅ Presigned URL (sin credenciales AWS)
- ✅ URL expiration (15 minutos)
- ✅ HTTP-only (no AWS SDK)

#### 4. Backend → SFTP
- ✅ SSH authentication
- ✅ chroot jail (`/home/addiuser/uploads`)
- ✅ Password protection

### Monitoreo

**CloudWatch Logs:**
- `/aws/lambda/addi-webhook-notifier` - Lambda execution logs
- Docker Compose logs - Backend API logs

**CloudWatch Metrics:**
- Lambda: Invocations, Errors, Duration, Throttles
- EventBridge: TriggeredRules, Invocations, FailedInvocations

**X-Ray Tracing:**
- End-to-end latency tracking
- Service map visualization
- Error analysis

**GuardDuty Findings:**
- S3 bucket anomaly detection
- Lambda network activity monitoring
- IAM credential misuse detection

---

## 💰 Costos Estimados

### Mensual (100 archivos/día ≈ 3,000/mes)

| Servicio | Costo |
|----------|-------|
| S3 Storage (10 GB) | $0.23 |
| S3 Requests (6,000 total) | $0.02 |
| Lambda (3,000 invocations) | $0.17 |
| EventBridge (3,000 events) | $0.00 |
| Secrets Manager | $0.40 |
| GuardDuty | $15-50 |
| CloudWatch Logs (1 GB) | $0.50 |
| SQS DLQ (~50 msgs) | $0.00 |
| **TOTAL** | **$16-51/mes** |

**Nota:** GuardDuty es opcional y representa ~90% del costo total.

---

## 🧊 Infraestructura Congelada

### Bucket: `addi-landing-zone-prod`

**Estado:** 🔒 BLOQUEADO (Object Lock COMPLIANCE - 7 años)

Este bucket fue creado con Object Lock COMPLIANCE durante pruebas y quedó permanentemente bloqueado hasta 2032-10-15.

**Configuración:**
- ObjectLockEnabled: true
- RetentionPeriod: 7 years (2032-10-15)
- Encryption: aws:kms
- Lifecycle: Transition to Glacier after 30 days

**Contenido:**
- `uploads/test-1760483459.csv` (114 bytes)
- **Costo acumulado 7 años:** ~$0.0003 (menos de un centavo)

**Recomendaciones:**
1. ✅ NO intentar borrar (causará errores)
2. ✅ Usar `addi-landing-zone-dev` en su lugar
3. 📅 Eliminar en 2032 cuando expire la retención

**Lección aprendida:**
- ⚠️ Object Lock COMPLIANCE es **permanente e irreversible**
- ✅ Usar Development Strategy para dev/test (sin Object Lock)

---

## 📚 Referencias

**Código Fuente:**
- CDK Stack: `stacks/addi/addi_stack_example.go`
- Lambda: `stacks/addi/lambda/webhook-notifier/main.go`
- Backend: `stacks/addi/backend/api/`
- Constructs: `constructs/{S3,Lambda,EventBridgeIntegrations,GuardDuty}/`

**Comandos Útiles:**
```bash
# Deploy
cdk deploy AddiStack --require-approval never

# Logs Lambda
aws logs tail /aws/lambda/addi-webhook-notifier --follow

# Upload test file
aws s3 cp test.csv s3://addi-landing-zone-dev/uploads/

# Verify SFTP
ls -lh stacks/addi/backend/sftp-data/uploads/

# Check GuardDuty
aws guardduty list-findings --detector-id $(aws guardduty list-detectors --query 'DetectorIds[0]' --output text)
```

---

**Última actualización:** 2025-10-15
**Estado:** ✅ Producción
**Versión:** 2.0.0
