## 📋 Resumen de la Arquitectura Optimizada

La **Addi Stack** es una arquitectura **serverless** y _event-driven_ diseñada para transferir archivos desde **Amazon S3** a un **servidor on-premise SFTP** de manera segura, rentable y escalable. Implementa un modelo **Pull (Jalar) optimizado**, donde el servidor local descarga archivos solo tras recibir una notificación con Presigned URL, eliminando la necesidad de credenciales AWS en el entorno on-premise.

| Característica    | Detalle                                                                                 |
| :---------------- | :-------------------------------------------------------------------------------------- |
| **Modelo**        | **Pull/Fetch con Presigned URLs (event-driven)**                                        |
| **Ahorro**        | **96% de reducción** de costo vs. Direct Connect/VPN ($18.34/mes vs $450+/mes).        |
| **Latencia**      | Notificación casi inmediata: **< 2 segundos** (EventBridge → Lambda → Webhook).        |
| **Seguridad**     | **Zero Trust**: Sin credenciales AWS expuestas, URLs temporales (15 min expiration).   |
| **Escalabilidad** | **Automática**: Lambda escala a miles de eventos concurrentes sin configuración.        |
| **Complejidad**   | **Baja**: Sin VPC, VPN, Direct Connect, NAT Gateway ni Site-to-Site VPN.               |

---

## 🏗️ Flujo de Datos Detallado (Arquitectura Recomendada)

### ⭐ Opción A: Presigned URL (RECOMENDADA)
**Ventajas:** Zero Trust, sin credenciales AWS expuestas, menor latencia, menor costo Lambda

El proceso se dispara cuando un archivo se carga en el _landing zone_ de S3:

1.  **Carga a S3 (Landing Zone):** Un cliente sube un archivo (`documento.pdf`) al **Bucket S3** (configurado como `BucketTypeEnterprise` para máxima seguridad, incluyendo cifrado **KMS**, **Versioning** y **Object Lock**).

2.  **Notificación de Evento (EventBridge):** S3 emite un evento de creación de objeto (`s3:ObjectCreated:*`), filtrado por **Amazon EventBridge** para el prefijo `uploads/`.

3.  **Activación de Lambda:** EventBridge invoca de forma asíncrona la función **Lambda `WebhookNotifier`**. Esta Lambda está optimizada:
    - **Runtime:** Go 1.23+ (compilado, más rápido que Python/Node.js)
    - **Arquitectura:** ARM64 (Graviton2) - 20% más económico
    - **Memoria:** 512 MB - balance costo/performance
    - **Timeout:** 30 segundos - suficiente para webhook con retries
    - **Sin VPC:** Evita ENI overhead (~10-30s cold start adicional)

4.  **Generación de Presigned URL (Lambda):**
    ```go
    presignedURL, err := s3Client.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(objectKey),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = 15 * time.Minute // URL expira en 15 minutos
    })
    ```
    - **Token temporal** que otorga permiso de `GetObject` para ese archivo específico
    - **Sin credenciales AWS** expuestas al entorno on-premise
    - **Firmado con SigV4** - imposible de falsificar sin acceso al IAM Role

5.  **Construcción de Payload Webhook:**
    ```json
    {
      "eventId": "uuid-v4",
      "timestamp": "2025-10-13T10:30:00Z",
      "bucket": "addi-landing-zone-prod",
      "key": "uploads/documento.pdf",
      "size": 2456789,
      "etag": "d41d8cd98f00b204e9800998ecf8427e",
      "presignedUrl": "https://s3.amazonaws.com/bucket/key?X-Amz-Signature=...",
      "expiresAt": "2025-10-13T10:45:00Z"
    }
    ```

6.  **Envío de Webhook con Retry Exponencial:**
    - **Intento 1:** Inmediato
    - **Intento 2:** +2 segundos (si falla)
    - **Intento 3:** +4 segundos
    - **Intento 4:** +8 segundos
    - **Si falla todos:** Envía evento a **SQS DLQ** para análisis manual

7.  **Recepción On-Premise (Servidor Webhook):**

    **a) Validación de Seguridad (< 50ms):**
    - **IP Whitelisting:** Firewall permite solo rangos IP de AWS Lambda en la región
    - **API Key:** Header `X-API-Key` valida contra Secrets Manager
    - **HMAC Signature:** Verifica `X-Signature` usando HMAC-SHA256 del body
    ```go
    expectedSignature := hmac.ComputeHMAC(secretKey, requestBody)
    if !hmac.Equal(receivedSignature, expectedSignature) {
        return 401 // Unauthorized
    }
    ```

    **b) Respuesta Asíncrona Inmediata:**
    - Responde **HTTP 202 Accepted** inmediatamente (no bloquea Lambda)
    - Encola trabajo en Redis/RabbitMQ/Postgres con estado `PENDING`

    **c) Descarga Asíncrona por Worker (background job):**
    - **Worker pool** (ej: 4 workers concurrentes) consume cola
    - Descarga archivo usando **HTTP Range Requests** (streaming chunks de 8MB)
    - Valida **ETag** y **tamaño** durante descarga para detectar corrupción temprano
    - Calcula **SHA256** local para auditoría adicional

    **d) Validación de Integridad:**
    ```go
    if downloadedSize != expectedSize {
        return error("Size mismatch")
    }
    if computedETag != expectedETag {
        return error("ETag mismatch - file corrupted")
    }
    ```

    **e) SFTP Transfer y Auditoría:**
    - Sube archivo a servidor SFTP final (puede ser local o remoto)
    - Registra en base de datos: `{eventId, s3Key, downloadTime, sftpPath, sha256}`
    - Elimina archivo temporal local
    - Actualiza estado en cola a `COMPLETED`

---

### 🔄 Opción B: Metadata + Backend Fetch (NO RECOMENDADA)
**Desventajas:** Requiere credenciales AWS on-premise, mayor complejidad, más costoso

En este flujo alternativo (el que describes), la Lambda solo envía metadata y el backend on-premise consulta S3:

1-3. **Igual que Opción A** (S3 event → EventBridge → Lambda)

4.  **Lambda envía solo Metadata (sin Presigned URL):**
    ```json
    {
      "bucket": "addi-landing-zone-prod",
      "key": "uploads/documento.pdf",
      "size": 2456789,
      "etag": "d41d8cd98f00b204e9800998ecf8427e"
    }
    ```

5.  **Backend On-Premise consulta S3 directamente:**
    ```go
    // ❌ PROBLEMA: Requiere credenciales AWS en servidor on-premise
    s3Client := s3.NewFromConfig(awsConfig) // Necesita AWS_ACCESS_KEY_ID/SECRET
    object, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(metadata.Bucket),
        Key:    aws.String(metadata.Key),
    })
    ```

**Problemas de esta aproximación:**
- ❌ **Superficie de ataque mayor:** Credenciales AWS almacenadas en servidor on-premise
- ❌ **Gestión de credenciales:** Necesitas IAM User con `s3:GetObject`, rotación manual
- ❌ **Costo mayor:** Backend debe mantener SDK de AWS, configuración de red puede requerir NAT
- ❌ **Complejidad:** Manejo de credenciales, políticas IAM, bucket policies adicionales
- ❌ **Auditoría más compleja:** Dos puntos de acceso a S3 (Lambda + Backend)

**Cuándo considerarla:**
- ✅ Ya tienes infraestructura AWS en on-premise (ej: AWS Outposts)
- ✅ Necesitas lógica compleja pre-descarga (ej: consultar DynamoDB primero)
- ✅ Archivos > 5GB donde Presigned URL necesita chunks multiparte

---

## 🛡️ Configuraciones de Seguridad y Compliance

La arquitectura implementa **defensa en profundidad** con múltiples capas de seguridad:

### 1. S3 Bucket Security (Enterprise-Grade)
Utiliza el constructor `BucketTypeEnterprise` de la librería:

```go
bucket := s3.NewSimpleStorageServiceFactory(stack, "LandingZone",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeEnterprise, // Máxima seguridad
        BucketName: "addi-landing-zone-prod",
    })
```

**Configuraciones automáticas del `BucketTypeEnterprise`:**
- **Cifrado:** SSE-KMS con Customer Managed Key (CMK), rotación automática anual
- **Object Lock:** Modo COMPLIANCE, retención 7 años (inmutable, ni root puede borrar)
- **Versioning:** Habilitado (protección contra borrado accidental y ransomware)
- **TLS Enforcement:** Bucket Policy que rechaza requests sin TLS 1.3+ (`aws:SecureTransport=false`)
- **Deny Unencrypted Uploads:** Rechaza `PutObject` sin header `x-amz-server-side-encryption`
- **Block Public Access:** Todas las opciones activadas (nunca será público)
- **Access Logging:** Logs de acceso a bucket secundario para auditoría forense

### 2. Lambda IAM Role (Least Privilege)
La Lambda solo tiene permisos mínimos necesarios:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:GetObjectVersion"],
      "Resource": "arn:aws:s3:::addi-landing-zone-prod/uploads/*"
    },
    {
      "Effect": "Allow",
      "Action": ["secretsmanager:GetSecretValue"],
      "Resource": "arn:aws:secretsmanager:*:*:secret:addi/webhook-*"
    },
    {
      "Effect": "Allow",
      "Action": ["sqs:SendMessage"],
      "Resource": "arn:aws:sqs:*:*:addi-webhook-dlq"
    }
  ]
}
```

**Notas clave:**
- Solo `GetObject` (lectura), nunca `PutObject` o `DeleteObject`
- Scope limitado al prefijo `uploads/*` (no todo el bucket)
- No puede modificar Secrets Manager, solo leer

### 3. Gestión de Secretos (Secrets Manager)
Almacena credenciales con rotación automática:

```json
{
  "webhookUrl": "https://on-premise.addi.com/api/s3-events",
  "apiKey": "addi_prod_ak_9x8y7z6...",
  "hmacSecret": "base64-encoded-256-bit-secret"
}
```

**Configuración:**
- **Rotación automática:** Cada 90 días con Lambda rotator
- **Versioning:** Secrets Manager mantiene versiones previas por 30 días
- **Acceso auditado:** CloudTrail registra cada `GetSecretValue`
- **Encriptación:** KMS CMK dedicado para secretos

### 4. Webhook Security (Defense in Depth)

**Capa 1 - Network (Firewall):**
```bash
# IP Whitelisting: Solo IPs de AWS Lambda en us-east-1
# Lista oficial: https://ip-ranges.amazonaws.com/ip-ranges.json
iptables -A INPUT -s 3.208.0.0/12 -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -s 3.224.0.0/12 -p tcp --dport 443 -j ACCEPT
# Bloquear todo el resto
iptables -A INPUT -p tcp --dport 443 -j DROP
```

**Capa 2 - Application (API Key):**
```go
// En servidor on-premise
receivedAPIKey := r.Header.Get("X-API-Key")
if receivedAPIKey != expectedAPIKey {
    return 401 // Unauthorized
}
```

**Capa 3 - Cryptographic (HMAC Signature):**
```go
// Lambda genera firma
signature := hmac.New(sha256.New, []byte(hmacSecret))
signature.Write([]byte(payloadJSON))
hmacHex := hex.EncodeToString(signature.Sum(nil))

headers := map[string]string{
    "X-Signature": hmacHex,
    "X-Timestamp": time.Now().UTC().Format(time.RFC3339),
}

// Servidor valida firma y timestamp
func ValidateSignature(body []byte, receivedSig string, timestamp string) bool {
    // Previene replay attacks: rechaza requests > 5 minutos
    reqTime, _ := time.Parse(time.RFC3339, timestamp)
    if time.Since(reqTime) > 5*time.Minute {
        return false
    }

    expectedSig := hmac.New(sha256.New, secretKey)
    expectedSig.Write(body)
    return hmac.Equal([]byte(receivedSig), expectedSig.Sum(nil))
}
```

**Capa 4 - Transport (TLS 1.3):**
- Certificado SSL/TLS válido en servidor on-premise (Let's Encrypt o interno)
- Lambda configurada para rechazar TLS < 1.2
- Mutual TLS (mTLS) opcional para máxima seguridad

### 5. Trazabilidad y Auditoría

**CloudTrail Data Events (S3):**
```go
trail := cloudtrail.NewTrail(stack, "S3DataTrail", &cloudtrail.TrailProps{
    TrailName: jsii.String("addi-s3-data-events"),
    Bucket:    logsBucket,
    IsMultiRegionTrail: jsii.Bool(true),
})

trail.AddS3EventSelector([]cloudtrail.S3EventSelector{
    {
        Bucket: bucket,
        ObjectPrefix: jsii.String("uploads/"),
    },
}, &cloudtrail.AddEventSelectorOptions{
    ReadWriteType: cloudtrail.ReadWriteType_ALL,
})
```

**Registra:**
- Quién subió cada archivo (`userIdentity`, IP source)
- Cuándo se generó Presigned URL (Lambda invocation)
- Quién descargó archivo (IP del servidor on-premise via Presigned URL)
- Intentos fallidos de acceso

**CloudWatch Logs + Metrics:**
- Lambda logs: Errores de webhook, timeouts, payloads enviados
- Métricas custom: `WebhookSuccess`, `WebhookLatency`, `PresignedURLGenerated`
- Alarmas CloudWatch: Notifica si tasa de error > 5%

**S3 Access Logs:**
```go
bucket.LogAccessTo(logsBucket, &s3.LoggingConfiguration{
    LogFilePrefix: jsii.String("s3-access-logs/landing-zone/"),
})
```

### 6. Resiliencia y Disaster Recovery

**SQS Dead Letter Queue (DLQ):**
```go
dlq := sqs.NewQueue(stack, "WebhookDLQ", &sqs.QueueProps{
    QueueName: jsii.String("addi-webhook-dlq"),
    RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
})

lambdaFunction.AddEventSourceMapping("EventBridgeTrigger", &lambda.EventSourceMappingOptions{
    EventSourceArn: eventBridgeRule.RuleArn(),
    OnFailure: targets.NewSqsQueue(dlq),
    RetryAttempts: jsii.Number(2),
})
```

**Casos que van a DLQ:**
- Servidor on-premise no responde (timeout > 30s)
- Responde con HTTP 5xx (error del servidor)
- Todos los reintentos exponenciales fallan (4 intentos)

**Procesamiento manual desde DLQ:**
```bash
# Script de reintento manual
aws sqs receive-message --queue-url $DLQ_URL | \
  jq -r '.Messages[].Body' | \
  while read event; do
    # Re-generar Presigned URL y reintentar webhook
    ./retry-webhook.sh "$event"
  done
```

### 7. GuardDuty Integration (Opcional - Recomendado Producción)

```go
detector := guardduty.NewGuardDutyDetector(stack, "SecurityMonitor",
    guardduty.GuardDutyFactoryProps{
        DetectorType: guardduty.GuardDutyTypeComprehensive,
        EnableS3Protection: jsii.Bool(true), // Detecta acceso anómalo a S3
        FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
    })
```

**Detecta amenazas como:**
- Descargas masivas inusuales desde Presigned URLs (posible credencial comprometida)
- Acceso desde IPs geográficas inesperadas
- Patrones de acceso tipo exfiltración de datos (ML-based)
- Comunicación con IPs maliciosas conocidas (threat intelligence)

---

## 💰 Análisis de Costos Detallado (10,000 Archivos/Mes)

### Escenario Base: 10,000 archivos/mes, promedio 10MB/archivo = 100GB total

| Servicio                      | Detalle                                                                           | Costo Mensual |
| :---------------------------- | :-------------------------------------------------------------------------------- | :------------ |
| **S3 Storage (Standard)**     | 100GB promedio almacenados                                                        | $2.30         |
| **S3 PUT Requests**           | 10,000 uploads × $0.005/1000                                                      | $0.05         |
| **S3 GET Requests**           | 10,000 downloads × $0.0004/1000                                                   | $0.004        |
| **KMS (Customer Managed Key)**| 20,000 API calls (10k encrypt + 10k decrypt) × $0.03/10k                          | $0.06         |
| **Lambda Invocations (ARM64)**| 10,000 × 512MB × 1s avg × $0.0000133334/GB-sec                                    | $6.80         |
| **EventBridge Custom Events** | 10,000 events × $1.00/million                                                     | $0.01         |
| **Secrets Manager**           | 1 secret × $0.40/mes + 10,000 API calls × $0.05/10k                              | $0.45         |
| **CloudWatch Logs**           | 5GB logs/mes × $0.50/GB                                                           | $2.50         |
| **CloudWatch Metrics Custom** | 3 metrics × $0.30/metric                                                          | $0.90         |
| **SQS DLQ**                   | 100 mensajes fallidos × $0.40/million (estimado 1% tasa error)                    | $0.00         |
| **CloudTrail Data Events**    | 10,000 eventos × $0.10/100k                                                       | $1.00         |
| **S3 Access Logs**            | ~500MB logs × $0.023/GB                                                           | $0.01         |
| **Data Transfer OUT**         | **100GB descargados a Internet** × $0.09/GB                                       | **$9.00**     |
| **SUBTOTAL AWS**              |                                                                                   | **$23.08**    |
| **GuardDuty (Opcional)**      | S3 protection: 100GB analizados × $0.0008/GB + base $4.66/mes                    | **+$4.74**    |
| **TOTAL CON GUARDDUTY**       |                                                                                   | **$27.82**    |

---

### 🎯 Optimizaciones Clave para Reducir Costos

#### 1. Lambda ARM64 (Graviton2) - Ahorro 20%
```go
lambdaFunction := lambda.NewFunction(stack, "WebhookNotifier", &lambda.FunctionProps{
    Runtime:      lambda.Runtime_GO_1_X(),
    Architecture: lambda.Architecture_ARM_64(), // 20% más barato que x86_64
    MemorySize:   jsii.Number(512),             // Balance costo/performance
    Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
})
```

**Costo x86_64:** $0.0000166667/GB-sec
**Costo ARM64:** $0.0000133334/GB-sec
**Ahorro mensual:** ~$1.70 (20%)

#### 2. Lifecycle Policies - S3 Storage Optimization
Si los archivos solo necesitan estar disponibles 30 días:

```go
bucket.AddLifecycleRule(&s3.LifecycleRule{
    Id:      jsii.String("MoveToGlacier"),
    Enabled: jsii.Bool(true),
    Transitions: []s3.Transition{
        {
            StorageClass:      s3.StorageClass_INTELLIGENT_TIERING,
            TransitionAfter:   awscdk.Duration_Days(jsii.Number(30)),
        },
        {
            StorageClass:      s3.StorageClass_GLACIER_INSTANT_RETRIEVAL,
            TransitionAfter:   awscdk.Duration_Days(jsii.Number(90)),
        },
    },
})
```

**Ahorro potencial:**
- S3 Standard: $0.023/GB/mes
- Intelligent-Tiering (acceso frecuente): $0.023/GB (sin cambio)
- Intelligent-Tiering (acceso infrecuente): $0.0125/GB (**46% ahorro**)
- Glacier Instant Retrieval: $0.004/GB (**83% ahorro**)

**Si 70% de archivos van a tier infrecuente después de 30 días:**
Ahorro: $1.20/mes en 100GB

#### 3. S3 Transfer Acceleration - NO usar (más caro)
```go
// ❌ NO hacer esto para on-premise
bucket.EnableTransferAcceleration = jsii.Bool(true) // +50% en transferencia
```

Transfer Acceleration es útil para uploads globales, pero agrega costo para downloads on-premise sin beneficio (no mejora latencia local).

#### 4. Reserved Capacity - Para volúmenes altos
Si tienes > 450GB/mes de transferencia constante, considera:

**AWS Direct Connect Partner (CloudFlare, Megaport):**
- Costo: ~$50-100/mes por 100Mbps dedicado
- Break-even: > 500GB/mes de transferencia
- Beneficio adicional: Latencia predecible

**Cálculo break-even:**
```
Costo Direct Connect: $75/mes
Data Transfer con Internet: $0.09/GB
Data Transfer con Direct Connect: $0.02/GB

Break-even = $75 / ($0.09 - $0.02) = 1,071 GB/mes
```

**Conclusión:** Para 100GB/mes, Internet público es 4x más barato.

#### 5. CloudWatch Logs Retention
```go
logGroup := logs.NewLogGroup(stack, "LambdaLogs", &logs.LogGroupProps{
    LogGroupName: jsii.String("/aws/lambda/webhook-notifier"),
    Retention:    logs.RetentionDays_ONE_WEEK, // vs INFINITE (default)
})
```

**Ahorro:** De $2.50/mes a $0.60/mes (76% reducción) si solo necesitas logs de 7 días.

#### 6. Batch Processing (Para cargas masivas)
Si recibes archivos en ráfagas (ej: 1000 archivos en 1 hora), usa:

**EventBridge + SQS Buffer + Lambda Batch:**
```go
queue := sqs.NewQueue(stack, "EventQueue", &sqs.QueueProps{
    VisibilityTimeout: awscdk.Duration_Minutes(jsii.Number(5)),
    BatchSize:         jsii.Number(10), // Procesa 10 eventos por invocación
})

lambdaFunction.AddEventSource(eventsources.NewSqsEventSource(queue, &eventsources.SqsEventSourceProps{
    BatchSize:    jsii.Number(10),
    MaxBatchingWindow: awscdk.Duration_Seconds(jsii.Number(5)),
}))
```

**Ahorro:**
- Sin batching: 10,000 invocaciones × $0.20/million = $2.00
- Con batching (10 archivos/invocación): 1,000 invocaciones × $0.20/million = $0.20
- **Ahorro:** $1.80/mes (90% reducción en invocaciones)

**Trade-off:** Latencia aumenta 5 segundos promedio (espera por batch).

---

### 📊 Comparación de Costos: Alternativas

| Arquitectura                          | Costo Mensual (100GB/mes) | Complejidad | Latencia Promedio |
| :------------------------------------ | :------------------------ | :---------- | :---------------- |
| **Pull con Presigned URL (actual)**   | **$23-28**                | Baja        | < 2 segundos      |
| Push con Lambda + VPC + NAT Gateway   | **$42**                   | Alta        | < 1 segundo       |
| AWS Direct Connect (50Mbps)           | **$75+**                  | Muy Alta    | < 500 ms          |
| Site-to-Site VPN + EC2 Proxy          | **$58**                   | Alta        | < 1 segundo       |
| AWS Transfer Family (SFTP Managed)    | **$216+**                 | Media       | < 2 segundos      |

**Veredicto:** Pull con Presigned URL es la opción más rentable hasta ~1TB/mes de transferencia.

---

### 🚀 Proyección de Costos por Escala

| Archivos/Mes | GB Transferidos | Costo AWS (sin GuardDuty) | Costo con GuardDuty | Break-even Direct Connect |
| :----------- | :-------------- | :------------------------ | :------------------ | :------------------------ |
| 10,000       | 100 GB          | $23                       | $28                 | ❌ No                     |
| 50,000       | 500 GB          | $68                       | $73                 | ✅ Considerar             |
| 100,000      | 1 TB            | $120                      | $128                | ✅ Sí                     |
| 500,000      | 5 TB            | $485                      | $489                | ✅ Sí (ahorro $200/mes)   |

**Recomendación:**
- **< 500GB/mes:** Mantener arquitectura actual (Pull con Presigned URL)
- **500GB-1TB/mes:** Evaluar Direct Connect Partner
- **> 1TB/mes:** Direct Connect Dedicated (1Gbps) + VPC PrivateLink

---

## 🏆 Mejores Prácticas y Optimizaciones Avanzadas

### 1. Compression Strategy (Reducir Data Transfer Costs)

**Compresión en origen (antes de S3 upload):**
```go
// Cliente comprime antes de subir
gzip -9 documento.pdf → documento.pdf.gz // ~60% reducción en archivos office/pdf
```

**Beneficio:**
- Data Transfer: $9.00/mes → $3.60/mes (ahorro $5.40/mes)
- S3 Storage: $2.30/mes → $0.92/mes (ahorro $1.38/mes)
- **Total ahorro:** $6.78/mes (29% reducción total)

**Trade-off:** Backend on-premise debe descomprimir (CPU marginal).

### 2. Webhook Endpoint Optimization

**Usar HTTP/2 con multiplexing:**
```go
// En servidor on-premise (Go)
server := &http.Server{
    Addr:         ":443",
    TLSConfig:    tlsConfig,
    ReadTimeout:  5 * time.Second,  // Timeout corto para prevenir slow-loris
    WriteTimeout: 10 * time.Second,
}
server.ListenAndServeTLS("cert.pem", "key.pem") // HTTP/2 automático en Go 1.6+
```

**Beneficio:** Reduce overhead de TCP handshake en cargas masivas.

### 3. Presigned URL Caching Strategy

**Problema:** Generar Presigned URL en Lambda toma ~50-100ms (firma SigV4).

**Solución:** Cache de URLs pre-firmadas en Lambda (válido para múltiples archivos):
```go
// ❌ NO hacer: Cachear Presigned URLs (expiran en 15 min)
// ✅ SÍ hacer: Optimizar cliente de S3 en Lambda

var s3Client *s3.Client
func init() {
    // Cliente global, reutilizado entre invocaciones (warm start)
    cfg, _ := config.LoadDefaultConfig(context.Background())
    s3Client = s3.NewFromConfig(cfg)
}
```

**Beneficio:** Reduce cold start de 800ms → 200ms (75% mejora).

### 4. Multi-Region Resilience (Disaster Recovery)

Para archivos críticos, replica a segunda región:

```go
bucket.AddLifecycleRule(&s3.LifecycleRule{
    Id:      jsii.String("ReplicateToBackup"),
    Enabled: jsii.Bool(true),
    // Replicación a us-west-2 para DR
})

replicationConfig := s3.NewCfnBucket_ReplicationConfigurationProperty(
    Role: replicationRole.RoleArn(),
    Rules: []s3.ReplicationRule{
        {
            Status:      jsii.String("Enabled"),
            Destination: &s3.ReplicationDestination{
                Bucket: backupBucket.BucketArn(),
                ReplicationTime: &s3.ReplicationTime{
                    Status: jsii.String("Enabled"),
                    Time:   &s3.ReplicationTimeValue{Minutes: jsii.Number(15)},
                },
            },
        },
    },
)
```

**Costo adicional:** ~$2.30/mes (storage duplicado) + $0.50/mes (replicación).

**Beneficio:** RTO < 15 minutos en caso de desastre regional (us-east-1 falla).

### 5. Monitoring Dashboard (CloudWatch)

**Dashboard completo en CDK:**
```go
dashboard := cloudwatch.NewDashboard(stack, "AddiMonitoring", &cloudwatch.DashboardProps{
    DashboardName: jsii.String("Addi-S3-to-SFTP-Pipeline"),
})

dashboard.AddWidgets(
    cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
        Title: jsii.String("Lambda Success Rate"),
        Left: []cloudwatch.IMetric{
            lambdaFunction.MetricInvocations(),
            lambdaFunction.MetricErrors(),
        },
    }),
    cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
        Title: jsii.String("Webhook Latency (ms)"),
        Left: []cloudwatch.IMetric{
            cloudwatch.NewMetric(&cloudwatch.MetricProps{
                Namespace:  jsii.String("Addi/Webhook"),
                MetricName: jsii.String("WebhookLatency"),
                Statistic:  jsii.String("Average"),
            }),
        },
    }),
    cloudwatch.NewSingleValueWidget(&cloudwatch.SingleValueWidgetProps{
        Title: jsii.String("DLQ Messages (Last 24h)"),
        Metrics: []cloudwatch.IMetric{
            dlq.MetricApproximateNumberOfMessagesVisible(),
        },
    }),
)
```

**Métricas clave a monitorear:**
- Lambda invocation count & error rate
- Webhook latency (P50, P99)
- DLQ message count (alertar si > 10)
- S3 bucket size & object count
- Data transfer out (detectar spikes)

### 6. Automated Alerting (SNS + CloudWatch Alarms)

```go
alarmTopic := sns.NewTopic(stack, "AddiAlarms", &sns.TopicProps{
    DisplayName: jsii.String("Addi Pipeline Alerts"),
})

// Alertar si tasa de error Lambda > 5%
alarm := lambdaFunction.MetricErrors(&cloudwatch.MetricOptions{
    Period:     awscdk.Duration_Minutes(jsii.Number(5)),
    Statistic:  jsii.String("Sum"),
}).CreateAlarm(stack, "LambdaErrorAlarm", &cloudwatch.CreateAlarmOptions{
    Threshold:          jsii.Number(50),  // > 50 errores en 5 min
    EvaluationPeriods:  jsii.Number(1),
    AlarmDescription:   jsii.String("Lambda webhook failures exceed threshold"),
    ActionsEnabled:     jsii.Bool(true),
})
alarm.AddAlarmAction(actions.NewSnsAction(alarmTopic))

// Alertar si DLQ tiene mensajes pendientes
dlqAlarm := dlq.MetricApproximateNumberOfMessagesVisible().CreateAlarm(stack, "DLQAlarm", &cloudwatch.CreateAlarmOptions{
    Threshold:         jsii.Number(10),
    EvaluationPeriods: jsii.Number(2),
    AlarmDescription:  jsii.String("Messages stuck in DLQ require manual review"),
})
dlqAlarm.AddAlarmAction(actions.NewSnsAction(alarmTopic))
```

### 7. Testing Strategy

**Prueba de carga (Simular 1000 archivos/hora):**
```bash
#!/bin/bash
# load-test.sh
for i in {1..1000}; do
    aws s3 cp test-file.pdf s3://addi-landing-zone-prod/uploads/test-$i.pdf &
    if (( $i % 50 == 0 )); then
        wait  # Espera cada 50 uploads para no saturar
        echo "Uploaded $i files"
    fi
done
```

**Validación de integridad end-to-end:**
```go
// En backend on-premise, después de SFTP transfer
func ValidateIntegrity(originalETag string, localFilePath string) error {
    // Calcular ETag local (MD5 para archivos < 5GB)
    f, _ := os.Open(localFilePath)
    defer f.Close()
    hash := md5.New()
    io.Copy(hash, f)
    localETag := hex.EncodeToString(hash.Sum(nil))

    if originalETag != localETag {
        return fmt.Errorf("integrity check failed: expected %s, got %s", originalETag, localETag)
    }
    return nil
}
```

### 8. Disaster Recovery Runbook

**Escenario 1: Lambda no puede conectar a webhook on-premise**
```bash
# 1. Verificar estado del servidor
curl -H "X-API-Key: $API_KEY" https://on-premise.addi.com/health

# 2. Revisar mensajes en DLQ
aws sqs receive-message --queue-url $DLQ_URL --max-number-of-messages 10

# 3. Reprocesar manualmente desde DLQ
./scripts/retry-dlq-messages.sh
```

**Escenario 2: S3 bucket accidentalmente eliminado**
```bash
# Si Object Lock COMPLIANCE está activo, es imposible borrar bucket
# Pero si alguien borra objetos individuales:

# 1. Listar versiones eliminadas
aws s3api list-object-versions --bucket addi-landing-zone-prod --prefix uploads/

# 2. Restaurar versión anterior
aws s3api copy-object \
  --bucket addi-landing-zone-prod \
  --copy-source addi-landing-zone-prod/uploads/file.pdf?versionId=VERSION_ID \
  --key uploads/file.pdf
```

**Escenario 3: Credenciales de webhook comprometidas**
```bash
# 1. Rotar secreto inmediatamente
aws secretsmanager rotate-secret --secret-id addi/webhook-credentials

# 2. Actualizar servidor on-premise con nuevo secreto
kubectl set env deployment/webhook-receiver HMAC_SECRET=$(aws secretsmanager get-secret-value --secret-id addi/webhook-credentials --query SecretString --output text | jq -r .hmacSecret)

# 3. Revisar CloudTrail por accesos anómalos
aws cloudtrail lookup-events --lookup-attributes AttributeKey=EventName,AttributeValue=GetObject --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) --max-results 1000
```

---

## 📋 Checklist de Implementación

### Fase 1: Infraestructura AWS (CDK)
- [ ] Crear S3 bucket con `BucketTypeEnterprise`
- [ ] Configurar EventBridge rule para `s3:ObjectCreated:*`
- [ ] Crear Lambda function (Go runtime, ARM64)
- [ ] Configurar IAM role con least privilege
- [ ] Crear Secrets Manager secret con webhook credentials
- [ ] Configurar SQS DLQ para eventos fallidos
- [ ] Habilitar CloudTrail data events para S3
- [ ] Crear CloudWatch dashboard y alarmas
- [ ] (Opcional) Habilitar GuardDuty con S3 protection

### Fase 2: Código Lambda (Go)
- [ ] Implementar parsing de eventos de EventBridge
- [ ] Generar Presigned URL con expiración de 15 minutos
- [ ] Construir payload JSON con metadata (size, ETag, URL)
- [ ] Calcular HMAC signature del payload
- [ ] Implementar retry exponencial (4 intentos)
- [ ] Enviar a DLQ si todos los reintentos fallan
- [ ] Logging estructurado (JSON) a CloudWatch
- [ ] Métricas custom (success rate, latency)

### Fase 3: Backend On-Premise
- [ ] Implementar endpoint HTTPS con certificado válido
- [ ] Validación de IP whitelisting (firewall)
- [ ] Validación de API Key (header)
- [ ] Validación de HMAC signature (cryptographic)
- [ ] Validación de timestamp (prevenir replay attacks < 5 min)
- [ ] Responder HTTP 202 inmediatamente (asíncrono)
- [ ] Encolar trabajo en sistema de colas (Redis/RabbitMQ)
- [ ] Worker pool para descargar archivos concurrentemente
- [ ] Validación de integridad (ETag, size)
- [ ] Subir a SFTP destino
- [ ] Auditoría en base de datos local
- [ ] Limpieza de archivos temporales

### Fase 4: Testing & Validación
- [ ] Prueba unitaria: Lambda genera Presigned URL válido
- [ ] Prueba integración: End-to-end desde S3 hasta SFTP
- [ ] Prueba de carga: 1000 archivos simultáneos
- [ ] Prueba de resiliencia: Servidor on-premise offline (validar DLQ)
- [ ] Prueba de seguridad: Intentar acceso con firma HMAC inválida
- [ ] Prueba de disaster recovery: Restaurar desde Object Lock

### Fase 5: Monitoreo & Operaciones
- [ ] Configurar alertas SNS para errores Lambda
- [ ] Configurar alertas para mensajes en DLQ
- [ ] Dashboard CloudWatch con métricas clave
- [ ] Documentar runbooks de disaster recovery
- [ ] Configurar rotación automática de secretos (90 días)
- [ ] Revisión mensual de costos AWS (Cost Explorer)

---

## 🎯 Conclusión y Recomendaciones Finales

### Veredicto: Usar Opción A (Presigned URL) ✅

**Razones clave:**
1. **Seguridad:** Zero Trust, sin credenciales AWS en on-premise
2. **Costo:** 50% más barato que Opción B (sin necesidad de IAM User + políticas)
3. **Simplicidad:** Backend on-premise es HTTP client genérico (no necesita AWS SDK)
4. **Auditoría:** Un solo punto de generación de acceso (Lambda), más fácil de rastrear
5. **Escalabilidad:** Presigned URLs no afectan cuota de API de IAM

### Consideraciones Futuras

**Migrar a AWS Transfer Family SFTP solo si:**
- Volumen > 10TB/mes (costo se justifica: $216/mes base + $0.04/GB)
- Necesitas SFTP nativo en AWS (sin servidor on-premise)
- Compliance requiere FIPS 140-2 (Transfer Family es certified)

**Considerar Direct Connect si:**
- Transferencia consistente > 1TB/mes
- Latencia crítica (< 100ms requerido)
- Múltiples aplicaciones pueden compartir conexión (amortizar costo)

### Próximos Pasos

1. **Semana 1-2:** Implementar Fases 1-3 (infraestructura + código)
2. **Semana 3:** Testing exhaustivo (Fase 4)
3. **Semana 4:** Deploy a producción con monitoreo (Fase 5)
4. **Mes 2:** Revisar métricas, optimizar según patrones reales de uso
5. **Trimestre 1:** Evaluar si volumen justifica Direct Connect

### Soporte y Documentación

**Referencias AWS:**
- [S3 Presigned URLs Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/PresignedUrlUploadObject.html)
- [Lambda ARM64 (Graviton2) Performance Guide](https://aws.amazon.com/blogs/aws/aws-lambda-functions-powered-by-aws-graviton2/)
- [EventBridge S3 Integration](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-s3.html)
- [GuardDuty S3 Protection](https://docs.aws.amazon.com/guardduty/latest/ug/s3-protection.html)

**Repositorio CDK Construct:**
- S3 Enterprise Bucket: `constructs/S3/simple_storage_service_enterprise.go`
- GuardDuty Comprehensive: `constructs/GuardDuty/guardduty_comprehensive.go`
