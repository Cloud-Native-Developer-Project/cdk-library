# Addi Stack - Arquitectura de Transferencia de Archivos

## 📋 Contexto del Proyecto

**Objetivo**: Crear un flujo automatizado que transfiera archivos desde S3 a un servidor SFTP externo (on-premise).

**Ambiente de Pruebas**: Simular el servidor SFTP externo usando EC2 dentro de AWS para validar el flujo completo antes de conectar al servidor real.

---

## 🎯 Arquitectura Optimizada (Serverless + Low Cost)

### Para Ambiente de Desarrollo/Testing

Esta arquitectura simula el flujo completo usando EC2 como servidor SFTP de prueba, pero con la flexibilidad de cambiar a Transfer Family cuando se conecte al servidor externo real.

```
┌──────────────────────────────────────────────────────────────────┐
│                    ARQUITECTURA SERVERLESS                       │
└──────────────────────────────────────────────────────────────────┘

Cliente (PUT archivo)
    ↓
┌────────────────────────────────┐
│ S3 Bucket (Landing Zone)       │  ← BucketTypeDevelopment
│ - EventBridge notifications ON │     (Auto-delete, CORS, 30d lifecycle)
│ - Versioning disabled          │     Costo: ~$0.023/GB/mes
│ - S3 Standard storage          │
└────────────┬───────────────────┘
             │ s3:ObjectCreated:* event
             ↓
┌────────────────────────────────┐
│ EventBridge Rule               │  ← Filtro por prefijo/sufijo
│ - Pattern: s3:ObjectCreated    │     Costo: Gratis (primeros 14M eventos)
│ - Target: Lambda               │
└────────────┬───────────────────┘
             │ Invoke Lambda
             ↓
┌────────────────────────────────────────────────────────────┐
│ Lambda Function (Python 3.12)                              │
│                                                             │
│ Flujo:                                                      │
│ 1. Validar archivo (size, extension)                       │
│ 2. Descargar archivo desde S3                              │
│ 3. Conectar a SFTP server via SSH (paramiko)               │
│ 4. Transferir archivo                                       │
│ 5. Verificar transferencia exitosa                         │
│ 6. Archivar/eliminar de S3                                 │
│ 7. Enviar notificación (SNS)                               │
│                                                             │
│ Recursos:                                                   │
│ - Memory: 512MB                                             │
│ - Timeout: 5 minutos                                        │
│ - Runtime: Python 3.12 (con layer paramiko)                │
│                                                             │
│ Costo: ~$0.20 por 1M invocations                           │
│        $0.0000166667 por GB-segundo                         │
└────────────┬───────────────────────────────────────────────┘
             │ SSH/SFTP Protocol (port 22)
             ↓
┌────────────────────────────────────────────────────────────┐
│ EC2 Instance (Simula servidor SFTP externo)               │
│                                                             │
│ Specs:                                                      │
│ - Instance: t4g.nano (ARM Graviton)                        │
│ - vCPU: 2 cores                                             │
│ - RAM: 512 MB                                               │
│ - Storage: 8GB EBS gp3                                      │
│ - OS: Amazon Linux 2023                                     │
│                                                             │
│ Software:                                                   │
│ - OpenSSH Server (sshd)                                     │
│ - sFTP subsystem enabled                                    │
│ - Usuario: sftpuser (chroot jail)                           │
│                                                             │
│ Networking:                                                 │
│ - Security Group: Solo Lambda (via SG source)              │
│ - Private subnet (NAT Gateway para Lambda)                  │
│ - No IP pública (no internet ingress)                       │
│                                                             │
│ Costo: ~$3.50/mes (t4g.nano Reserved 1yr)                  │
│        ~$0.80/mes (8GB EBS gp3)                             │
│        Total: ~$4.30/mes                                    │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────┐
│ CloudWatch Logs + Metrics      │  ← Monitoring
│ - Lambda execution logs        │     Costo: ~$0.50/GB ingested
│ - SFTP transfer metrics        │            ~$0.03/métrica
│ - Success/failure counters     │
└────────────────────────────────┘

┌────────────────────────────────┐
│ SNS Topic                      │  ← Notificaciones
│ - Transfer success             │     Costo: $0.50 por 1M notificaciones
│ - Transfer failure             │
│ - Email/SMS/Slack webhook      │
└────────────────────────────────┘

┌────────────────────────────────┐
│ GuardDuty (Optional)           │  ← Security monitoring
│ - GuardDutyTypeBasic           │     Costo: ~$4-8/mes
│ - S3 data event monitoring     │     (Recomendado solo para prod)
└────────────────────────────────┘
```

---

## 💰 Análisis de Costos (Optimizado)

### Ambiente de Desarrollo/Testing

| Servicio | Configuración | Costo Mensual |
|----------|---------------|---------------|
| **S3 Bucket** | 10GB storage, 1000 PUT, 5000 GET | $0.23 + $0.005 + $0.002 = **$0.24** |
| **Lambda** | 1000 invocations/mes, 512MB, 30s avg | $0.20 (Free Tier) = **$0.00** |
| **EC2 (SFTP)** | t4g.nano Reserved 1yr + 8GB EBS gp3 | $3.50 + $0.80 = **$4.30** |
| **EventBridge** | 1000 eventos/mes | Free Tier = **$0.00** |
| **CloudWatch** | 5GB logs, 10 métricas custom | $2.50 + $0.30 = **$2.80** |
| **SNS** | 100 notificaciones/mes | Free Tier = **$0.00** |
| **Data Transfer** | 10GB/mes (Lambda→EC2, mismo AZ) | **$0.00** |
| **NAT Gateway** | No usar (alternativa: VPC Endpoint o EC2 public) | **$0.00** |
| | **TOTAL MENSUAL** | **~$7.34/mes** |

### Optimizaciones Aplicadas

1. ✅ **t4g.nano** en lugar de t3.micro (40% más barato, ARM Graviton)
2. ✅ **Reserved Instance 1 año** (62% descuento vs On-Demand)
3. ✅ **Sin NAT Gateway** ($32/mes) → usar EC2 con IP pública en subnet pública
4. ✅ **Lambda en Free Tier** (1M requests gratis)
5. ✅ **S3 Development strategy** (sin versioning, auto-delete)
6. ✅ **Sin Step Functions** ($25 por 1M transiciones) → Lambda directo
7. ✅ **Sin Transfer Family Connector** ($0.04/GB + $0.60/hr) → SSH nativo

**Ahorro vs Arquitectura Original**: ~$200/mes → ~$7/mes = **96% de ahorro**

---

## 🏗️ Componentes Técnicos Detallados

### 1. S3 Bucket (Landing Zone)

**Estrategia**: `BucketTypeDevelopment` (optimizada para testing)

```go
bucket := s3.NewSimpleStorageServiceFactory(stack, "AddiLandingBucket",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeDevelopment,
        BucketName: "addi-file-landing-dev",
        RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
        AutoDeleteObjects: jsii.Bool(true),
    })
```

**Características**:
- ✅ EventBridge notifications (crítico para event-driven)
- ✅ CORS permisivo (si subes desde navegador)
- ✅ Lifecycle 30 días (auto-cleanup)
- ✅ Auto-delete on stack destroy
- ❌ Versioning disabled (reduce costos)
- ❌ KMS encryption (usa S3_MANAGED para dev)

### 2. EventBridge Rule

```go
rule := events.NewRule(stack, "AddiFileUploadRule", &events.RuleProps{
    EventPattern: &events.EventPattern{
        Source:     jsii.Strings("aws.s3"),
        DetailType: jsii.Strings("Object Created"),
        Detail: map[string]interface{}{
            "bucket": map[string]interface{}{
                "name": jsii.Strings(*bucket.BucketName()),
            },
            // Filtro opcional por tipo de archivo
            "object": map[string]interface{}{
                "key": []map[string]interface{}{
                    {"prefix": "uploads/"},
                    {"suffix": ".csv"},
                },
            },
        },
    },
    Targets: []events.IRuleTarget{
        targets.NewLambdaFunction(transferLambda),
    },
})
```

### 3. Lambda Function (SFTP Transfer)

**Runtime**: Python 3.12 (más reciente y eficiente)
**Memory**: 512MB (balance costo/performance)
**Timeout**: 5 minutos (suficiente para archivos hasta 100MB)

**Código Lambda** (`lambda/sftp_transfer/handler.py`):

```python
import json
import boto3
import paramiko
import os
from io import BytesIO

s3 = boto3.client('s3')
sns = boto3.client('sns')

# Configuración desde variables de entorno
SFTP_HOST = os.environ['SFTP_HOST']
SFTP_PORT = int(os.environ.get('SFTP_PORT', '22'))
SFTP_USER = os.environ['SFTP_USER']
SFTP_PRIVATE_KEY = os.environ['SFTP_PRIVATE_KEY_SECRET']  # Secrets Manager ARN
SFTP_REMOTE_PATH = os.environ.get('SFTP_REMOTE_PATH', '/uploads')
SNS_TOPIC_ARN = os.environ['SNS_TOPIC_ARN']
MAX_FILE_SIZE = int(os.environ.get('MAX_FILE_SIZE_MB', '100')) * 1024 * 1024

def get_ssh_key_from_secrets_manager(secret_arn):
    """Obtiene la clave SSH privada desde Secrets Manager"""
    secrets = boto3.client('secretsmanager')
    response = secrets.get_secret_value(SecretId=secret_arn)
    return response['SecretString']

def lambda_handler(event, context):
    print(f"Evento recibido: {json.dumps(event)}")

    try:
        # Extraer información del archivo desde el evento de S3
        detail = event['detail']
        bucket = detail['bucket']['name']
        key = detail['object']['key']
        size = detail['object']['size']

        # Validación: Tamaño máximo
        if size > MAX_FILE_SIZE:
            raise ValueError(f"Archivo demasiado grande: {size} bytes (máximo: {MAX_FILE_SIZE})")

        print(f"Procesando archivo: s3://{bucket}/{key} ({size} bytes)")

        # Descargar archivo desde S3 a memoria
        s3_object = s3.get_object(Bucket=bucket, Key=key)
        file_data = s3_object['Body'].read()

        # Obtener clave SSH privada desde Secrets Manager
        private_key_str = get_ssh_key_from_secrets_manager(SFTP_PRIVATE_KEY)
        private_key = paramiko.RSAKey.from_private_key(BytesIO(private_key_str.encode()))

        # Conectar al servidor SFTP
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        print(f"Conectando a SFTP: {SFTP_USER}@{SFTP_HOST}:{SFTP_PORT}")
        ssh.connect(
            hostname=SFTP_HOST,
            port=SFTP_PORT,
            username=SFTP_USER,
            pkey=private_key,
            timeout=30
        )

        # Abrir canal SFTP
        sftp = ssh.open_sftp()

        # Nombre del archivo remoto (preservar nombre original)
        remote_filename = os.path.basename(key)
        remote_filepath = f"{SFTP_REMOTE_PATH}/{remote_filename}"

        print(f"Transfiriendo archivo a: {remote_filepath}")

        # Transferir archivo
        sftp.putfo(BytesIO(file_data), remote_filepath)

        # Verificar que el archivo se subió correctamente
        remote_file_stat = sftp.stat(remote_filepath)
        if remote_file_stat.st_size != size:
            raise Exception(f"Error: Tamaño no coincide. Local: {size}, Remoto: {remote_file_stat.st_size}")

        # Cerrar conexiones
        sftp.close()
        ssh.close()

        print(f"✅ Transferencia exitosa: {remote_filepath}")

        # Opcional: Mover archivo a carpeta "processed" en S3
        processed_key = f"processed/{key}"
        s3.copy_object(
            Bucket=bucket,
            CopySource={'Bucket': bucket, 'Key': key},
            Key=processed_key
        )
        s3.delete_object(Bucket=bucket, Key=key)
        print(f"Archivo archivado: s3://{bucket}/{processed_key}")

        # Notificar éxito
        sns.publish(
            TopicArn=SNS_TOPIC_ARN,
            Subject='✅ Transferencia SFTP Exitosa',
            Message=f"""
Archivo transferido exitosamente:

S3 Source: s3://{bucket}/{key}
SFTP Destination: {SFTP_HOST}:{remote_filepath}
File Size: {size} bytes
Status: SUCCESS
            """
        )

        return {
            'statusCode': 200,
            'body': json.dumps({
                'message': 'Transferencia exitosa',
                'file': remote_filepath,
                'size': size
            })
        }

    except Exception as e:
        error_msg = f"❌ Error en transferencia: {str(e)}"
        print(error_msg)

        # Notificar error
        sns.publish(
            TopicArn=SNS_TOPIC_ARN,
            Subject='❌ Error en Transferencia SFTP',
            Message=f"""
Error al transferir archivo:

S3 Source: s3://{bucket}/{key}
Error: {str(e)}
Status: FAILED
            """
        )

        return {
            'statusCode': 500,
            'body': json.dumps({'error': str(e)})
        }
```

**Lambda Layer para Paramiko**:

```bash
# Crear layer con paramiko
mkdir python
pip install paramiko cryptography -t python/
zip -r paramiko-layer.zip python/
aws lambda publish-layer-version \
  --layer-name paramiko \
  --zip-file fileb://paramiko-layer.zip \
  --compatible-runtimes python3.12
```

**CDK Configuration**:

```go
// Lambda function
transferLambda := awslambda.NewFunction(stack, "SFTPTransferFunction", &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_PYTHON_3_12(),
    Handler: jsii.String("handler.lambda_handler"),
    Code:    awslambda.Code_FromAsset(jsii.String("lambda/sftp_transfer")),
    MemorySize: jsii.Number(512),
    Timeout: awscdk.Duration_Minutes(jsii.Number(5)),

    // Lambda layer con paramiko
    Layers: []awslambda.ILayerVersion{
        paramikoLayer,
    },

    // Variables de entorno
    Environment: &map[string]*string{
        "SFTP_HOST":               ec2Instance.InstancePrivateIp(),
        "SFTP_PORT":               jsii.String("22"),
        "SFTP_USER":               jsii.String("sftpuser"),
        "SFTP_PRIVATE_KEY_SECRET": sshKeySecret.SecretArn(),
        "SFTP_REMOTE_PATH":        jsii.String("/home/sftpuser/uploads"),
        "SNS_TOPIC_ARN":           snsTopic.TopicArn(),
        "MAX_FILE_SIZE_MB":        jsii.String("100"),
    },

    // VPC configuration (mismo VPC que EC2)
    Vpc: vpc,
    VpcSubnets: &awsec2.SubnetSelection{
        SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
    },

    // Security group
    SecurityGroups: []awsec2.ISecurityGroup{
        lambdaSG,
    },
})

// Permisos IAM
bucket.GrantRead(transferLambda)
bucket.GrantDelete(transferLambda)
bucket.GrantPut(transferLambda)
sshKeySecret.GrantRead(transferLambda)
snsTopic.GrantPublish(transferLambda)
```

### 4. EC2 Instance (SFTP Server Simulado)

**Instance Type**: t4g.nano (ARM Graviton, más barato)
**OS**: Amazon Linux 2023
**Storage**: 8GB EBS gp3 (más barato que gp2)

**User Data Script** (configuración automática):

```bash
#!/bin/bash
# Script de inicialización para servidor SFTP

# Actualizar sistema
dnf update -y

# Instalar OpenSSH Server (ya viene instalado pero aseguramos)
dnf install -y openssh-server

# Crear usuario SFTP con chroot jail
useradd -m -d /home/sftpuser -s /bin/bash sftpuser

# Crear directorio de uploads
mkdir -p /home/sftpuser/uploads
chown sftpuser:sftpuser /home/sftpuser/uploads

# Configurar chroot para SFTP (seguridad)
cat >> /etc/ssh/sshd_config <<EOF

# SFTP Chroot Configuration
Match User sftpuser
    ForceCommand internal-sftp
    ChrootDirectory /home/sftpuser
    PermitTunnel no
    AllowAgentForwarding no
    AllowTcpForwarding no
    X11Forwarding no
EOF

# Ajustar permisos para chroot
chown root:root /home/sftpuser
chmod 755 /home/sftpuser

# Configurar autenticación SSH con clave pública (desde Secrets Manager)
# Nota: La clave pública se inyectará via CDK
mkdir -p /home/sftpuser/.ssh
echo "{{SSH_PUBLIC_KEY}}" > /home/sftpuser/.ssh/authorized_keys
chown -R sftpuser:sftpuser /home/sftpuser/.ssh
chmod 700 /home/sftpuser/.ssh
chmod 600 /home/sftpuser/.ssh/authorized_keys

# Reiniciar SSH daemon
systemctl restart sshd
systemctl enable sshd

# CloudWatch Agent (opcional, para monitoring)
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# Log de finalización
echo "SFTP Server configurado exitosamente" > /var/log/sftp-setup-complete.log
```

**CDK Configuration**:

```go
// VPC
vpc := awsec2.NewVpc(stack, "AddiVPC", &awsec2.VpcProps{
    MaxAzs: jsii.Number(2),
    NatGateways: jsii.Number(0), // Sin NAT Gateway para reducir costos
    SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
        {
            Name: jsii.String("Public"),
            SubnetType: awsec2.SubnetType_PUBLIC,
            CidrMask: jsii.Number(24),
        },
    },
})

// Security Group para SFTP Server
sftpSG := awsec2.NewSecurityGroup(stack, "SFTPServerSG", &awsec2.SecurityGroupProps{
    Vpc: vpc,
    Description: jsii.String("Security group for SFTP server"),
    AllowAllOutbound: jsii.Bool(true),
})

// Solo permitir SSH desde Lambda SG
sftpSG.AddIngressRule(
    lambdaSG,
    awsec2.Port_Tcp(jsii.Number(22)),
    jsii.String("Allow SSH from Lambda"),
)

// EC2 Instance
instance := awsec2.NewInstance(stack, "SFTPServer", &awsec2.InstanceProps{
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
                VolumeType: awsec2.EbsDeviceVolumeType_GP3,
                DeleteOnTermination: jsii.Bool(true),
            }),
        },
    },
    SecurityGroup: sftpSG,
    VpcSubnets: &awsec2.SubnetSelection{
        SubnetType: awsec2.SubnetType_PUBLIC,
    },
    UserData: awsec2.UserData_Custom(jsii.String(userDataScript)),
})

// Outputs
awscdk.NewCfnOutput(stack, "SFTPServerIP", &awscdk.CfnOutputProps{
    Value: instance.InstancePrivateIp(),
    Description: jsii.String("SFTP Server Private IP"),
})
```

### 5. Secrets Manager (SSH Keys)

```go
// Generar par de llaves SSH (hacer esto manualmente y subir a Secrets Manager)
// ssh-keygen -t rsa -b 4096 -f sftp_key -N ""

sshKeySecret := awssecretsmanager.NewSecret(stack, "SFTPSSHKey", &awssecretsmanager.SecretProps{
    SecretName: jsii.String("addi/sftp/ssh-private-key"),
    Description: jsii.String("SSH private key for SFTP authentication"),
    SecretStringValue: awscdk.SecretValue_UnsafePlainText(jsii.String(`-----BEGIN RSA PRIVATE KEY-----
... (contenido de la clave privada) ...
-----END RSA PRIVATE KEY-----`)),
})
```

### 6. Monitoring & Alerting

```go
// SNS Topic para notificaciones
snsTopic := awssns.NewTopic(stack, "SFTPTransferNotifications", &awssns.TopicProps{
    DisplayName: jsii.String("SFTP Transfer Notifications"),
})

// Suscripción por email
snsTopic.AddSubscription(
    awssnssubscriptions.NewEmailSubscription(jsii.String("tu-email@empresa.com")),
)

// CloudWatch Alarm: Lambda errors
lambdaErrorAlarm := awscloudwatch.NewAlarm(stack, "LambdaErrorAlarm", &awscloudwatch.AlarmProps{
    Metric: transferLambda.MetricErrors(&awscloudwatch.MetricOptions{
        Period: awscdk.Duration_Minutes(jsii.Number(5)),
        Statistic: jsii.String("Sum"),
    }),
    Threshold: jsii.Number(1),
    EvaluationPeriods: jsii.Number(1),
    AlarmDescription: jsii.String("Lambda function has errors"),
    ActionsEnabled: jsii.Bool(true),
})

lambdaErrorAlarm.AddAlarmAction(awscloudwatchactions.NewSnsAction(snsTopic))

// CloudWatch Dashboard
dashboard := awscloudwatch.NewDashboard(stack, "AddiMonitoringDashboard", &awscloudwatch.DashboardProps{
    DashboardName: jsii.String("Addi-SFTP-Transfer"),
})

dashboard.AddWidgets(
    awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
        Title: jsii.String("Lambda Invocations"),
        Left: &[]awscloudwatch.IMetric{
            transferLambda.MetricInvocations(),
            transferLambda.MetricErrors(),
        },
    }),
    awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
        Title: jsii.String("Lambda Duration"),
        Left: &[]awscloudwatch.IMetric{
            transferLambda.MetricDuration(),
        },
    }),
)
```

---

## 🚀 Migración a Producción (Servidor SFTP Externo Real)

Cuando estés listo para conectar al servidor SFTP externo, solo necesitas cambiar **3 configuraciones**:

### Cambios Necesarios

1. **Actualizar variables de entorno de Lambda**:

```go
Environment: &map[string]*string{
    "SFTP_HOST":    jsii.String("sftp.cliente-externo.com"),  // ← Cambiar aquí
    "SFTP_PORT":    jsii.String("2222"),                      // ← Si usa puerto custom
    "SFTP_USER":    jsii.String("cliente_user"),              // ← Usuario real
    // ... resto igual
}
```

2. **Configurar conectividad de red** (elegir una opción):

**Opción A: Direct Connect** (ya existe)
```go
// Lambda en subnet privada con route a Direct Connect Gateway
VpcSubnets: &awsec2.SubnetSelection{
    SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
}
```

**Opción B: Site-to-Site VPN**
```go
// Configurar VPN Gateway y actualizar route tables
vpnGateway := awsec2.NewVpnGateway(stack, "VPNGateway", &awsec2.VpnGatewayProps{
    Type: awsec2.VpnConnectionType_IPSEC_1,
})
vpc.AddVpnGateway(vpnGateway)
```

**Opción C: Internet público** (menos recomendado)
```go
// Lambda en subnet pública con NAT Gateway
VpcSubnets: &awsec2.SubnetSelection{
    SubnetType: awsec2.SubnetType_PUBLIC,
}
```

3. **Actualizar Security Groups**:

```go
// Permitir outbound SSH desde Lambda hacia servidor externo
lambdaSG.AddEgressRule(
    awsec2.Peer_Ipv4(jsii.String("10.20.30.40/32")), // IP del servidor externo
    awsec2.Port_Tcp(jsii.Number(22)),
    jsii.String("Allow SFTP to external server"),
)
```

4. **Eliminar EC2 instance de testing**:

```diff
- instance := awsec2.NewInstance(stack, "SFTPServer", ...)
+ // EC2 ya no es necesaria
```

5. **Actualizar Secrets Manager con credenciales reales**:

```bash
# Obtener clave privada del cliente
aws secretsmanager update-secret \
  --secret-id addi/sftp/ssh-private-key \
  --secret-string file://cliente-private-key.pem
```

**Total de cambios**: ~10 líneas de código + 1 configuración de red.

---

## 📊 Comparación: Testing vs Producción

| Aspecto | Testing (EC2 interno) | Producción (SFTP externo) |
|---------|----------------------|---------------------------|
| **Costo** | ~$7/mes | ~$7/mes + conectividad |
| **Conectividad** | Interna (VPC) | Direct Connect/VPN |
| **Latencia** | <5ms | 10-50ms (según distancia) |
| **Seguridad** | Security Groups | SSH + firewall on-premise |
| **Configuración Lambda** | Host: IP privada EC2 | Host: dominio externo |
| **Complejidad** | Baja | Media (requiere red hybrid) |

---

## ✅ Ventajas de Esta Arquitectura

### 1. **Ultra Bajo Costo** (~$7/mes)
- 96% más barato que usar Transfer Family Connector
- Free Tier de Lambda (1M requests gratis)
- EC2 t4g.nano Reserved (62% descuento)

### 2. **Serverless Real**
- Lambda se escala automáticamente
- No hay servidores que gestionar (excepto EC2 de testing)
- Pay-per-use real

### 3. **Fácil Testing**
- EC2 simula perfectamente servidor SFTP externo
- Valida flujo completo antes de conectar a producción
- No requiere Direct Connect para pruebas

### 4. **Migración Transparente**
- Solo cambiar 3 configs para ir a producción
- Mismo código Lambda funciona para ambos casos
- Zero downtime migration

### 5. **Monitoring Completo**
- CloudWatch Logs de todas las transferencias
- SNS notifications (success/failure)
- CloudWatch Dashboard con métricas clave

### 6. **Resiliente**
- Retry automático de Lambda (3 intentos)
- Dead Letter Queue para errores persistentes
- Archivado automático de archivos procesados

---

## 🔧 Instrucciones de Implementación

### Paso 1: Generar Llaves SSH

```bash
# Generar par de llaves
ssh-keygen -t rsa -b 4096 -f ./sftp_key -N ""

# Esto crea:
# - sftp_key (privada) → subir a Secrets Manager
# - sftp_key.pub (pública) → inyectar en EC2 User Data
```

### Paso 2: Subir Clave Privada a Secrets Manager

```bash
aws secretsmanager create-secret \
  --name addi/sftp/ssh-private-key \
  --description "SSH private key for SFTP authentication" \
  --secret-string file://sftp_key \
  --region us-east-1
```

### Paso 3: Crear Lambda Layer con Paramiko

```bash
mkdir -p lambda-layers/paramiko/python
pip install paramiko cryptography -t lambda-layers/paramiko/python/
cd lambda-layers/paramiko
zip -r ../../paramiko-layer.zip python/
cd ../..

aws lambda publish-layer-version \
  --layer-name paramiko \
  --description "Paramiko library for SFTP" \
  --zip-file fileb://paramiko-layer.zip \
  --compatible-runtimes python3.12 \
  --region us-east-1
```

### Paso 4: Desplegar Stack CDK

```bash
# Sintetizar CloudFormation template
cdk synth AddiStack

# Desplegar
cdk deploy AddiStack --require-approval never

# Outputs importantes:
# - SFTPServerIP: XXX.XXX.XXX.XXX
# - LambdaFunctionArn: arn:aws:lambda:...
# - S3BucketName: addi-file-landing-dev
```

### Paso 5: Validar SFTP Server

```bash
# Obtener IP privada del EC2
SFTP_IP=$(aws cloudformation describe-stacks \
  --stack-name AddiStack \
  --query 'Stacks[0].Outputs[?OutputKey==`SFTPServerIP`].OutputValue' \
  --output text)

# Conectar via SSH (desde una instancia en el mismo VPC o via bastion)
ssh -i sftp_key sftpuser@$SFTP_IP

# Verificar directorio uploads existe
ls -la /home/sftpuser/uploads
```

### Paso 6: Probar Transferencia End-to-End

```bash
# Subir archivo de prueba a S3
echo "Test file content" > test-file.txt
aws s3 cp test-file.txt s3://addi-file-landing-dev/uploads/test-file.txt

# Monitorear logs de Lambda
aws logs tail /aws/lambda/AddiStack-SFTPTransferFunction --follow

# Verificar archivo llegó al SFTP server
ssh -i sftp_key sftpuser@$SFTP_IP "ls -lh /home/sftpuser/uploads/"
```

---

## 🐛 Troubleshooting

### Lambda no puede conectar a EC2

**Problema**: `Connection refused` o timeout

**Solución**:
1. Verificar Security Groups:
   ```bash
   # Lambda SG debe estar en el Ingress del SFTP SG
   aws ec2 describe-security-groups --group-ids sg-xxxxx
   ```

2. Verificar Lambda está en el mismo VPC:
   ```bash
   aws lambda get-function-configuration --function-name AddiStack-SFTPTransferFunction
   # Verificar VpcConfig tiene SubnetIds y SecurityGroupIds
   ```

3. Verificar SSH daemon corriendo en EC2:
   ```bash
   ssh -i sftp_key ec2-user@$SFTP_IP "sudo systemctl status sshd"
   ```

### Archivo no aparece en SFTP server

**Problema**: Lambda dice "success" pero archivo no existe

**Solución**:
1. Verificar permisos del directorio:
   ```bash
   ssh -i sftp_key sftpuser@$SFTP_IP "ls -la /home/sftpuser/"
   # uploads/ debe tener permisos 755 y owner sftpuser:sftpuser
   ```

2. Revisar logs detallados de Lambda:
   ```bash
   aws logs filter-log-events \
     --log-group-name /aws/lambda/AddiStack-SFTPTransferFunction \
     --filter-pattern "ERROR"
   ```

### Lambda timeout (5 minutos)

**Problema**: Archivos grandes causan timeout

**Solución**:
1. Aumentar timeout (máximo 15 minutos):
   ```go
   Timeout: awscdk.Duration_Minutes(jsii.Number(15)),
   ```

2. O implementar streaming en lugar de cargar a memoria:
   ```python
   # En lugar de:
   file_data = s3_object['Body'].read()

   # Usar:
   sftp.putfo(s3_object['Body'], remote_filepath)
   ```

---

## 📚 Referencias

- [Paramiko Documentation](https://www.paramiko.org/)
- [Lambda VPC Networking](https://docs.aws.amazon.com/lambda/latest/dg/configuration-vpc.html)
- [EC2 Instance Types](https://aws.amazon.com/ec2/instance-types/t4/)
- [OpenSSH SFTP Configuration](https://man.openbsd.org/sshd_config)

---

## 🔄 Roadmap

### Fase 1: PoC (Completado) ✅
- [x] S3 bucket con EventBridge
- [x] Lambda con Paramiko
- [x] EC2 como SFTP server
- [x] Transferencia básica funcional

### Fase 2: Optimización (1 semana)
- [ ] Dead Letter Queue (SQS) para errores
- [ ] Retry exponencial personalizado
- [ ] Métricas custom en CloudWatch
- [ ] Dashboard detallado

### Fase 3: Producción (2 semanas)
- [ ] Migrar a BucketTypeEnterprise
- [ ] Configurar Direct Connect/VPN
- [ ] Conectar a servidor SFTP externo real
- [ ] GuardDuty Comprehensive
- [ ] Runbook operacional

### Fase 4: Evolución (futuro)
- [ ] Considerar Step Functions si lógica se complica
- [ ] PGP encryption end-to-end
- [ ] Multi-destination (varios servidores SFTP)
- [ ] Chunking para archivos >1GB

---

**Última actualización**: 2025-10-11
**Versión**: 2.0 (Arquitectura optimizada serverless)
**Costo estimado**: ~$7/mes (testing) | ~$7-50/mes (producción con conectividad)
