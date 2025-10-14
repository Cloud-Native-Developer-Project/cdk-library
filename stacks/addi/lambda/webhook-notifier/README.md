# Webhook Notifier Lambda

Lambda function que procesa eventos de S3 (vía EventBridge), genera Presigned URLs y envía notificaciones webhook a un servidor on-premise.

## Funcionalidad

### Flujo de Ejecución

1. **Recibe evento de EventBridge** con metadata de S3 (bucket, key, size, etag)
2. **Genera Presigned URL** temporal (15 minutos) para el objeto S3
3. **Lee credenciales** desde Secrets Manager (webhook URL, API Key, HMAC secret)
4. **Construye payload JSON** con toda la información necesaria
5. **Calcula firma HMAC-SHA256** del payload para autenticación
6. **Envía HTTP POST** al servidor on-premise con retry exponencial (4 intentos)
7. **Si falla** después de todos los reintentos → EventBridge envía evento a DLQ

### Características de Seguridad

- ✅ **Presigned URLs**: URLs temporales con expiración de 15 minutos
- ✅ **HMAC Signature**: Firma criptográfica HMAC-SHA256 del payload
- ✅ **API Key**: Validación adicional mediante header `X-API-Key`
- ✅ **Timestamp**: Anti-replay attack con timestamp en header `X-Timestamp`
- ✅ **Secrets Manager**: Credenciales seguras, rotación automática
- ✅ **TLS**: Comunicación HTTPS con el servidor on-premise

### Retry Logic

**Estrategia: Exponential Backoff**
- Intento 1: Inmediato
- Intento 2: +2 segundos
- Intento 3: +4 segundos
- Intento 4: +8 segundos
- **Total:** 4 intentos en ~14 segundos

Si todos fallan → EventBridge captura el evento en Dead Letter Queue (SQS)

## Configuración

### Variables de Entorno (configuradas por CDK)

| Variable | Descripción | Default |
|----------|-------------|---------|
| `BUCKET_NAME` | Nombre del bucket S3 | (requerido) |
| `WEBHOOK_SECRET_ARN` | ARN del secret en Secrets Manager | (requerido) |
| `PRESIGNED_URL_EXPIRES` | Expiración de Presigned URL (segundos) | 900 (15 min) |
| `MAX_RETRY_ATTEMPTS` | Número máximo de reintentos | 4 |
| `RETRY_EXPONENTIAL_BASE` | Base para backoff exponencial | 2 |

### Secrets Manager Structure

```json
{
  "webhookUrl": "https://on-premise.addi.com/api/s3-events",
  "apiKey": "addi_prod_ak_xxxxxxxxxxxxx",
  "hmacSecret": "base64-encoded-256-bit-secret"
}
```

## Payload Webhook

### Request Headers

```
Content-Type: application/json
X-API-Key: addi_prod_ak_xxxxxxxxxxxxx
X-Signature: hmac-sha256-hex-encoded
X-Timestamp: 2025-10-13T10:30:00Z
User-Agent: AWS-Lambda-Webhook-Notifier/1.0
```

### Request Body

```json
{
  "eventId": "12345678-1234-1234-1234-123456789abc",
  "timestamp": "2025-10-13T10:30:00Z",
  "bucket": "addi-landing-zone-prod",
  "key": "uploads/documento.pdf",
  "size": 2456789,
  "etag": "d41d8cd98f00b204e9800998ecf8427e",
  "presignedUrl": "https://s3.amazonaws.com/bucket/key?X-Amz-Signature=...",
  "expiresAt": "2025-10-13T10:45:00Z"
}
```

### Expected Response

**Success:**
```
HTTP/1.1 202 Accepted
Content-Type: application/json

{
  "status": "accepted",
  "jobId": "job-12345"
}
```

**Error:**
```
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": "Internal server error"
}
```

## Build & Deploy

### Build for ARM64 (Graviton2)

```bash
# Dentro del directorio lambda/webhook-notifier/
cd stacks/addi/lambda/webhook-notifier

# Descargar dependencias
go mod download

# Build para ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o main main.go

# Verificar binario
file main
# Output: main: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV)
```

### Deploy Stack

```bash
# Desde el root del proyecto
cd /Users/andressepulveda/Documents/cdk-library

# Build Lambda
cd stacks/addi/lambda/webhook-notifier
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o main main.go
cd ../../../..

# Deploy CDK stack
cdk deploy AddiS3ToSFTPStack --require-approval never
```

## Testing

### Test Local (sin desplegar)

```bash
# Instalar SAM CLI
brew install aws-sam-cli

# Invocar Lambda localmente
sam local invoke WebhookNotifier -e test-event.json
```

**test-event.json:**
```json
{
  "version": "0",
  "id": "test-event-id",
  "detail-type": "Object Created",
  "source": "aws.s3",
  "detail": {
    "bucket": {
      "name": "addi-landing-zone-prod"
    },
    "object": {
      "key": "uploads/test.pdf",
      "size": 1024,
      "etag": "abc123"
    }
  }
}
```

### Test en AWS

```bash
# Subir archivo de prueba a S3
aws s3 cp test.pdf s3://addi-landing-zone-prod/uploads/test.pdf

# Monitorear logs de Lambda
aws logs tail /aws/lambda/addi-webhook-notifier --follow

# Ver eventos en DLQ (si hubo fallos)
aws sqs receive-message --queue-url https://sqs.us-east-1.amazonaws.com/123456789/S3ToLambdaIntegration-dlq
```

## Monitoring

### CloudWatch Metrics

- **Invocations**: Número total de invocaciones
- **Errors**: Errores no manejados (exceptions)
- **Duration**: Tiempo de ejecución (ms)
- **Throttles**: Invocaciones limitadas por concurrencia

### CloudWatch Logs

Logs estructurados con prefijo `[INFO]`, `[ERROR]`, etc.:

```
2025-10-13T10:30:00Z [INFO] Processing event: abc-123
2025-10-13T10:30:01Z [INFO] S3 object: s3://bucket/key (size: 1024 bytes)
2025-10-13T10:30:02Z [INFO] Generated presigned URL (expires at: 2025-10-13T10:45:00Z)
2025-10-13T10:30:03Z [INFO] Webhook sent successfully (status: 202)
2025-10-13T10:30:03Z [INFO] Successfully processed event: abc-123
```

### X-Ray Tracing

La Lambda tiene X-Ray habilitado por defecto. Visualiza traces en:
- AWS X-Ray Console
- CloudWatch ServiceLens

**Segmentos típicos:**
1. Lambda initialization (cold start)
2. Secrets Manager GetSecretValue
3. S3 PresignGetObject
4. HTTP POST to webhook endpoint

## Error Handling

### Errores Retryables (se reintenta)

- HTTP 5xx del servidor on-premise
- Timeout de red (> 10 segundos)
- Connection refused

### Errores No Retryables (falla inmediatamente)

- HTTP 4xx del servidor on-premise (excepto 429)
- Presigned URL generation failure
- Secrets Manager access denied

### Dead Letter Queue

Eventos que fallan después de todos los reintentos van a SQS DLQ:

```bash
# Leer mensajes de DLQ
aws sqs receive-message \
  --queue-url https://sqs.us-east-1.amazonaws.com/123456789/S3ToLambdaIntegration-dlq \
  --max-number-of-messages 10

# Reprocesar manualmente desde DLQ
# (Script de reintento manual)
```

## Performance

### Cold Start

- **ARM64 Go runtime**: ~200ms
- **Secrets Manager cache**: Primera invocación +50ms, luego 0ms (cached)
- **Total cold start**: ~250ms

### Warm Invocation

- **Parse event**: ~1ms
- **Generate Presigned URL**: ~10ms
- **Send webhook**: ~50-200ms (depende de latencia on-premise)
- **Total warm execution**: ~60-210ms

### Memory Usage

- **Allocated**: 512 MB
- **Peak usage**: ~80 MB
- **Efficiency**: 15% usage (sufficient headroom)

## Cost Estimation

**Escenario: 10,000 archivos/mes**

| Componente | Costo Mensual |
|------------|---------------|
| Lambda invocations (10,000 × ARM64 × 512MB × 0.2s) | $0.17 |
| Secrets Manager API calls (10,000 × $0.05/10k) | $0.05 |
| Data transfer OUT (10,000 × 1KB payload) | $0.00 |
| **Total Lambda** | **$0.22** |

**Nota:** Costo de transferencia de datos desde S3 (Presigned URL) no está incluido aquí (se cobra cuando el servidor on-premise descarga el archivo).

## Troubleshooting

### Lambda no recibe eventos

```bash
# Verificar EventBridge Rule
aws events describe-rule --name S3ToLambdaIntegration-rule

# Verificar que S3 tenga EventBridge habilitado
aws s3api get-bucket-notification-configuration --bucket addi-landing-zone-prod
```

### Webhook falla con 401 Unauthorized

- Verificar API Key en Secrets Manager
- Verificar firma HMAC (secret debe coincidir)
- Revisar timestamp (debe estar dentro de 5 minutos)

### Presigned URL expirada

- Default: 15 minutos
- Si el servidor on-premise tarda más → aumentar `PRESIGNED_URL_EXPIRES`
- Máximo recomendado: 1 hora (3600 segundos)

## Dependencies

```
github.com/aws/aws-lambda-go v1.47.0           # Lambda runtime
github.com/aws/aws-sdk-go-v2 v1.32.4           # AWS SDK v2 core
github.com/aws/aws-sdk-go-v2/config v1.28.3    # SDK config
github.com/aws/aws-sdk-go-v2/service/s3        # S3 client (Presigned URLs)
github.com/aws/aws-sdk-go-v2/service/secretsmanager # Secrets Manager client
```

## Referencias

- [Lambda Go Runtime](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html)
- [S3 Presigned URLs](https://docs.aws.amazon.com/AmazonS3/latest/userguide/PresignedUrlUploadObject.html)
- [EventBridge S3 Events](https://docs.aws.amazon.com/AmazonS3/latest/userguide/EventBridge.html)
- [AWS SDK Go v2](https://aws.github.io/aws-sdk-go-v2/docs/)
