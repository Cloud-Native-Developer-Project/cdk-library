# Documentación: S3 CloudFront Strategy

## Tabla de Contenidos

- [Resumen Ejecutivo](#resumen-ejecutivo)
- [Arquitectura Técnica](#arquitectura-técnica)
- [Flujo de Construcción](#flujo-de-construcción)
- [Configuración Detallada](#configuración-detallada)
- [Casos de Uso](#casos-de-uso)
- [Seguridad](#seguridad)
- [Rendimiento](#rendimiento)
- [Troubleshooting](#troubleshooting)
- [Ejemplos Completos](#ejemplos-completos)
- [Mejores Prácticas](#mejores-prácticas)

---

## Resumen Ejecutivo

La **S3CloudFrontStrategy** es una implementación especializada del patrón Strategy que proporciona una configuración lista para producción de distribuciones CloudFront con origen en Amazon S3. Esta estrategia está optimizada para:

- ✅ **Single Page Applications (SPAs)** - React, Vue, Angular
- ✅ **Sitios web estáticos** - HTML/CSS/JS, generadores estáticos
- ✅ **Assets estáticos** - Imágenes, videos, documentos
- ✅ **Progressive Web Apps (PWAs)**

### Características Principales

| Característica  | Implementación              | Beneficio                                     |
| --------------- | --------------------------- | --------------------------------------------- |
| **Seguridad**   | Origin Access Control (OAC) | Acceso privado al S3 sin políticas públicas   |
| **Rendimiento** | HTTP/2 + HTTP/3, Compresión | Hasta 80% reducción en transferencia de datos |
| **SSL/TLS**     | TLS 1.2+ con SNI            | Máxima compatibilidad y seguridad             |
| **SPA Support** | Error handling inteligente  | Routing de lado del cliente funcional         |
| **IPv6**        | Habilitado por defecto      | Mejor alcance global                          |
| **Cache**       | Políticas optimizadas AWS   | Menor latencia, menor costo                   |

---

## Arquitectura Técnica

### Diagrama de Componentes

```
┌────────────────────────────────────────────────────────────────┐
│                    CloudFront Distribution                      │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │  Default Behavior                                         │ │
│  │  • HTTPS Only (Redirect)                                  │ │
│  │  • GET, HEAD, OPTIONS                                     │ │
│  │  • CACHING_OPTIMIZED Policy                               │ │
│  │  • SECURITY_HEADERS Policy                                │ │
│  │  • Compression Enabled                                    │ │
│  └──────────────────┬───────────────────────────────────────┘ │
│                     │                                           │
│                     ▼                                           │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │  Origin Access Control (OAC)                              │ │
│  │  • Service Principal: cloudfront.amazonaws.com            │ │
│  │  • Condition: SourceArn equals Distribution ARN           │ │
│  └──────────────────┬───────────────────────────────────────┘ │
│                     │                                           │
└─────────────────────┼───────────────────────────────────────────┘
                      │
                      ▼
            ┌──────────────────┐
            │   S3 Bucket      │
            │   (Private)      │
            │                  │
            │  Bucket Policy:  │
            │  Allow CloudFront│
            │  Service via OAC │
            └──────────────────┘
```

### Stack de Tecnologías

- **AWS CDK v2** - Infrastructure as Code
- **CloudFront** - CDN global con 450+ edge locations
- **S3** - Almacenamiento de objetos
- **ACM** - Certificados SSL/TLS administrados (opcional)
- **WAF v2** - Web Application Firewall (opcional)
- **IAM** - Gestión de permisos y políticas

---

## Flujo de Construcción

### Secuencia de Ejecución

```go
func (s *S3CloudFrontStrategy) Build(...) awscloudfront.Distribution {
    // 1. Validación
    // 2. Crear OAC
    // 3. Configurar Distribution Props
    // 4. Configurar Default Behavior
    // 5. Error Responses (SPA)
    // 6. WAF Integration
    // 7. Crear Distribution
    // 8. Configurar S3 Bucket Policy
    return distribution
}
```

### Paso 1: Validación Básica

```go
if props.S3Bucket == nil {
    panic("S3CloudFrontStrategy requiere un bucket S3")
}
```

**Validación en tiempo de ejecución** que garantiza que la estrategia recibe los recursos necesarios antes de proceder con la construcción.

### Paso 2: Origin Access Control (OAC)

```go
oac := awscloudfront.NewS3OriginAccessControl(scope,
    jsii.String(fmt.Sprintf("%s-OAC", id)),
    &awscloudfront.S3OriginAccessControlProps{
        Description: jsii.String(fmt.Sprintf("OAC for %s", id)),
    })
```

#### ¿Por qué OAC y no OAI?

| Característica      | OAI (Legacy)            | OAC (Moderno)             |
| ------------------- | ----------------------- | ------------------------- |
| Métodos HTTP        | Solo GET, HEAD          | Todos (PUT, DELETE, POST) |
| Seguridad           | Políticas S3 explícitas | Service Principal IAM     |
| SSE-KMS             | Limitado                | Soporte completo          |
| Estado              | Deprecated              | Recomendado por AWS       |
| Fecha recomendación | < 2022                  | 2022+                     |

### Paso 3: Configuración Base de Distribución

```go
distributionProps := &awscloudfront.DistributionProps{
    Comment:           jsii.String(props.Comment),
    DefaultRootObject: jsii.String("index.html"),
    HttpVersion:       awscloudfront.HttpVersion_HTTP2_AND_3,
    EnableIpv6:        jsii.Bool(true),
    EnableLogging:     jsii.Bool(props.EnableAccessLogging),
    PriceClass:        awscloudfront.PriceClass_PRICE_CLASS_100,
}
```

#### Desglose de Configuraciones

##### **DefaultRootObject: "index.html"**

- Sirve `/index.html` cuando se accede a la raíz del dominio
- Crítico para SPAs
- Evita errores 403 en la ruta base

##### **HttpVersion: HTTP2_AND_3**

**HTTP/2 Beneficios:**

- Multiplexing de solicitudes (múltiples recursos en una conexión)
- Server Push (envío proactivo de recursos)
- Compresión de headers (HPACK)
- Priorización de streams

**HTTP/3 Beneficios:**

- Protocolo QUIC (UDP en lugar de TCP)
- Eliminación de head-of-line blocking
- Conexiones más rápidas (0-RTT resumption)
- Mejor rendimiento en redes móviles/inestables

##### **EnableIpv6: true**

- Alcance a ~50% de usuarios en mobile
- Mejor enrutamiento en algunas geografías
- Sin costo adicional
- Requerido para algunos mercados (China, India)

##### **PriceClass: PRICE_CLASS_100**

| Price Class | Regiones                                  | Casos de Uso                    |
| ----------- | ----------------------------------------- | ------------------------------- |
| **100**     | US, Canadá, Europa                        | Audiencias occidentales         |
| **200**     | + Asia, África, Medio Oriente, Sudamérica | Audiencias globales balanceadas |
| **ALL**     | Todas (450+ edge locations)               | Máximo rendimiento global       |

**Decisión de diseño**: `PRICE_CLASS_100` es el default por **balance costo/rendimiento** para startups y medianas empresas.

### Paso 4: SSL/TLS (Opcional)

```go
if props.CertificateArn != "" {
    cert := awscertificatemanager.Certificate_FromCertificateArn(...)
    distributionProps.Certificate = cert
    distributionProps.MinimumProtocolVersion = awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
    distributionProps.SslSupportMethod = awscloudfront.SSLMethod_SNI
}
```

#### **MinimumProtocolVersion: TLS_V1_2_2021**

**Versiones disponibles:**

| Versión          | Compatibilidad | Seguridad     | Recomendación            |
| ---------------- | -------------- | ------------- | ------------------------ |
| TLS 1.0          | 99.9%          | ⚠️ Vulnerable | Nunca usar               |
| TLS 1.1          | 99%            | ⚠️ Inseguro   | Deprecated               |
| TLS 1.2_2019     | 95%            | ✅ Buena      | Mínimo aceptable         |
| **TLS 1.2_2021** | 90%            | ✅ Excelente  | **Recomendado**          |
| TLS 1.3          | 70%            | ✅ Máxima     | Para audiencias modernas |

**TLS_V1_2_2021** incluye:

- Cipher suites más fuertes
- Perfect Forward Secrecy (PFS)
- Sin soporte para ciphers débiles (RC4, 3DES)

#### **SSLSupportMethod: SNI_ONLY**

| Método  | Costo    | Compatibilidad               | Uso             |
| ------- | -------- | ---------------------------- | --------------- |
| **SNI** | Gratis   | Navegadores modernos (2010+) | **Recomendado** |
| VIP     | $600/mes | Navegadores muy antiguos     | Legacy only     |

**Server Name Indication (SNI)**:

- Múltiples certificados SSL en la misma IP
- Soportado por 99.5% de navegadores
- Sin costo adicional

### Paso 5: Dominios Personalizados (CNAMEs)

```go
if len(props.DomainNames) > 0 {
    var domains []*string
    for _, d := range props.DomainNames {
        domains = append(domains, jsii.String(d))
    }
    distributionProps.DomainNames = &domains
}
```

**Requisitos**:

1. Certificado ACM en **us-east-1** (CloudFront es global)
2. Validación de dominio (DNS o Email)
3. Record CNAME/ALIAS en Route 53 o DNS provider

**Ejemplo**:

```go
props := CloudFrontPropertiesV2{
    DomainNames: []string{"www.miapp.com", "miapp.com"},
    CertificateArn: "arn:aws:acm:us-east-1:123456789012:certificate/abc123",
}
```

### Paso 6: Default Behavior - Comportamiento por Defecto

```go
distributionProps.DefaultBehavior = &awscloudfront.BehaviorOptions{
    Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(...),
    ViewerProtocolPolicy:  awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
    AllowedMethods:        awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS(),
    CachedMethods:         awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS(),
    CachePolicy:           awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
    ResponseHeadersPolicy: awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS(),
    Compress:              jsii.Bool(true),
}
```

#### **ViewerProtocolPolicy: REDIRECT_TO_HTTPS**

| Política              | Comportamiento          | Uso             |
| --------------------- | ----------------------- | --------------- |
| ALLOW_ALL             | HTTP y HTTPS permitidos | ❌ Nunca usar   |
| **REDIRECT_TO_HTTPS** | HTTP → 301 → HTTPS      | **Recomendado** |
| HTTPS_ONLY            | HTTP bloqueado (error)  | APIs internas   |

#### **AllowedMethods: GET, HEAD, OPTIONS**

**Métodos HTTP**:

- **GET**: Recuperar recursos (la mayoría del tráfico)
- **HEAD**: Metadatos sin body (preflight checks)
- **OPTIONS**: CORS preflight requests

**No incluidos** (contenido estático):

- POST, PUT, DELETE, PATCH

Para APIs o formularios, usar `ALLOW_ALL` en behaviors específicos.

#### **CachedMethods: GET, HEAD, OPTIONS**

Solo estos métodos almacenan respuestas en cache. POST/PUT nunca se cachean por seguridad.

#### **CachePolicy: CACHING_OPTIMIZED**

**AWS Managed Policy** que incluye:

```
Cache Key Components:
├── Query Strings: None (ignora todos)
├── Headers:
│   └── Accept-Encoding (para compresión)
├── Cookies: None
└── TTL Settings:
    ├── Min: 1 segundo
    ├── Default: 86400 segundos (1 día)
    └── Max: 31536000 segundos (1 año)
```

**Comportamiento**:

1. CloudFront ignora query strings para cache key
2. Cachea por 1 día por defecto
3. Respeta `Cache-Control` headers del origen
4. Compresión automática basada en `Accept-Encoding`

**Alternativas**:

| Policy            | Query Strings | Cookies | Headers | Uso                  |
| ----------------- | ------------- | ------- | ------- | -------------------- |
| CACHING_OPTIMIZED | ❌            | ❌      | Mínimos | **Sitios estáticos** |
| CACHING_DISABLED  | ✅            | ✅      | Todos   | APIs dinámicas       |
| ELEMENTAL_MEDIA   | ✅            | ❌      | Media   | Video streaming      |

#### **ResponseHeadersPolicy: SECURITY_HEADERS**

**AWS Managed Policy** que agrega:

```http
Strict-Transport-Security: max-age=63072000; includeSubdomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

**Protecciones**:

- **HSTS**: Fuerza HTTPS en el navegador (2 años)
- **X-Content-Type-Options**: Previene MIME sniffing
- **X-Frame-Options**: Protege contra clickjacking
- **X-XSS-Protection**: Protección XSS en navegadores legacy
- **Referrer-Policy**: Control de información de referrer

#### **Compress: true**

**Compresión automática** para:

- `text/*` (HTML, CSS, JavaScript, XML)
- `application/json`
- `application/javascript`
- `application/xml`

**No comprime**:

- Imágenes (ya comprimidas)
- Videos
- Archivos ya comprimidos (gzip, br)

**Beneficios**:

- 70-80% reducción para texto
- Sin configuración en origen
- Soporte Gzip y Brotli

### Paso 7: Error Responses para SPAs

```go
distributionProps.ErrorResponses = &[]*awscloudfront.ErrorResponse{
    {
        HttpStatus:         jsii.Number(403),
        ResponseHttpStatus: jsii.Number(200),
        ResponsePagePath:   jsii.String("/index.html"),
        Ttl:                awscdk.Duration_Minutes(jsii.Number(5)),
    },
    {
        HttpStatus:         jsii.Number(404),
        ResponseHttpStatus: jsii.Number(200),
        ResponsePagePath:   jsii.String("/index.html"),
        Ttl:                awscdk.Duration_Minutes(jsii.Number(5)),
    },
}
```

#### ¿Por qué es Necesario?

**Problema de SPAs**:

```
Usuario → CloudFront → S3
GET /app/users/123

S3 busca: /app/users/123 (no existe)
S3 responde: 404 Not Found
```

**Solución**:

```
Usuario → CloudFront → S3
GET /app/users/123

S3: 404 → CloudFront intercepta
CloudFront: Sirve /index.html con 200 OK
React Router: Maneja /app/users/123
```

#### **403 vs 404**

| Error   | Causa            | Cuándo Ocurre                             |
| ------- | ---------------- | ----------------------------------------- |
| **403** | S3 denies access | Rutas sin archivo, prefijos no existentes |
| **404** | Object not found | URL directa a archivo inexistente         |

**Ambos deben redirigir a `/index.html`** para SPAs.

#### **TTL: 5 minutos**

```go
Ttl: awscdk.Duration_Minutes(jsii.Number(5))
```

- CloudFront cachea la respuesta de error
- Después de 5 minutos, reintenta al origen
- Balance entre rendimiento y actualización

**Consideraciones**:

- Muy bajo (1 min): Más requests al origen
- Muy alto (1 hora): Deployments tardan en reflejarse
- **5 minutos**: Sweet spot para la mayoría

### Paso 8: AWS WAF Integration

```go
if props.WebAclArn != "" {
    distributionProps.WebAclId = jsii.String(props.WebAclArn)
}
```

**Ejemplo de WebACL ARN**:

```
arn:aws:wafv2:us-east-1:123456789012:global/webacl/MyWebACL/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111
```

**Protecciones típicas**:

- Rate limiting (ejemplo: 2000 req/5min por IP)
- SQL injection patterns
- XSS patterns
- Known bad IPs (AWS Managed Rules)
- Geo-blocking (complementario a CloudFront Geo Restriction)
- Bot detection

### Paso 9: Crear Distribución

```go
distribution := awscloudfront.NewDistribution(
    scope,
    jsii.String(fmt.Sprintf("%s-Distribution", id)),
    distributionProps
)
```

CloudFront crea:

- Distribution ID: `E1ABCDEFGHIJK`
- Domain name: `d111111abcdef8.cloudfront.net`
- Provisiona edge locations globalmente (~5-10 minutos)

### Paso 10: Configuración Automática de S3 Bucket Policy

```go
if props.AutoConfigureS3BucketPolicy {
    props.S3Bucket.AddToResourcePolicy(awsiam.NewPolicyStatement({
        Sid:    jsii.String("AllowCloudFrontServicePrincipal"),
        Effect: awsiam.Effect_ALLOW,
        Principals: &[]awsiam.IPrincipal{
            awsiam.NewServicePrincipal(jsii.String("cloudfront.amazonaws.com"), nil),
        },
        Actions: jsii.Strings("s3:GetObject"),
        Resources: jsii.Strings(fmt.Sprintf("%s/*", *props.S3Bucket.BucketArn())),
        Conditions: &map[string]interface{}{
            "StringEquals": map[string]interface{}{
                "AWS:SourceArn": *distribution.DistributionArn(),
            },
        },
    }))
}
```

#### Política IAM Generada

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowCloudFrontServicePrincipal",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudfront.amazonaws.com"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::mi-bucket/*",
      "Condition": {
        "StringEquals": {
          "AWS:SourceArn": "arn:aws:cloudfront::123456789012:distribution/E1ABCDEFGHIJK"
        }
      }
    }
  ]
}
```

#### Seguridad de la Política

**Principio de Least Privilege**:

- ✅ Solo `s3:GetObject` (lectura)
- ✅ Solo desde **esta distribución específica** (SourceArn)
- ✅ Service Principal en lugar de IAM user/role
- ✅ No requiere bucket público

**Comparación con OAI (legacy)**:

```json
// OAI (legacy) - Menos seguro
{
  "Principal": {
    "AWS": "arn:aws:iam::cloudfront:user/CloudFront Origin Access Identity E1ABCDEF"
  }
}

// OAC (moderno) - Más seguro
{
  "Principal": {
    "Service": "cloudfront.amazonaws.com"
  },
  "Condition": {
    "StringEquals": {
      "AWS:SourceArn": "arn:aws:cloudfront::123456789012:distribution/E1ABCDEFGHIJK"
    }
  }
}
```

---

## Casos de Uso

### 1. Single Page Application (React/Vue/Angular)

```go
import (
    "mi-proyecto/cloudfront"
    "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
)

func DeploySPA(stack constructs.Construct) {
    // Crear bucket
    bucket := awss3.NewBucket(stack, jsii.String("SPABucket"), &awss3.BucketProps{
        BucketName: jsii.String("mi-spa-prod"),
        Versioned:  jsii.Bool(true),
        Encryption: awss3.BucketEncryption_S3_MANAGED,
    })

    // Distribución CloudFront
    props := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    bucket,
        Comment:                     "SPA Production",
        DomainNames:                 []string{"app.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/abc",
        EnableAccessLogging:         true,
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(
        stack,
        "SPADistribution",
        props,
    )
}
```

**Deployment típico**:

1. Build: `npm run build` → `/build` directory
2. Upload: `aws s3 sync build/ s3://mi-spa-prod/`
3. Invalidate: `aws cloudfront create-invalidation --distribution-id E1ABC --paths "/*"`

### 2. Sitio Web Estático con Blog (Hugo/Jekyll)

```go
func DeployStaticSite(stack constructs.Construct) {
    bucket := awss3.NewBucket(stack, jsii.String("BlogBucket"), &awss3.BucketProps{
        BucketName:       jsii.String("mi-blog-static"),
        WebsiteIndexDocument: jsii.String("index.html"),
        WebsiteErrorDocument: jsii.String("404.html"),
    })

    props := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    bucket,
        Comment:                     "Blog Estático",
        DomainNames:                 []string{"blog.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/xyz",
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(stack, "BlogDistribution", props)
}
```

**Ventajas**:

- Sin servidor (serverless)
- Costo: ~$1-5/mes para miles de visitas
- Performance: ~50ms tiempo de respuesta global
- 99.99% uptime SLA

### 3. Assets CDN para Aplicación (Imágenes/Videos)

```go
func DeployAssetsCDN(stack constructs.Construct) {
    assetsBucket := awss3.NewBucket(stack, jsii.String("AssetsBucket"), &awss3.BucketProps{
        BucketName: jsii.String("mi-app-assets"),
        Cors: &[]*awss3.CorsRule{
            {
                AllowedMethods: &[]awss3.HttpMethods{
                    awss3.HttpMethods_GET,
                    awss3.HttpMethods_HEAD,
                },
                AllowedOrigins: jsii.Strings("https://app.miempresa.com"),
                AllowedHeaders: jsii.Strings("*"),
            },
        },
    })

    props := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    assetsBucket,
        Comment:                     "Assets CDN",
        DomainNames:                 []string{"cdn.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/cdn",
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(stack, "AssetsCDN", props)
}
```

**Uso en frontend**:

```html
<img src="https://cdn.miempresa.com/images/logo.png" />
<video src="https://cdn.miempresa.com/videos/demo.mp4"></video>
```

### 4. Entorno Multi-Ambiente (Dev/Staging/Prod)

```go
func DeployMultiEnvironment(app awscdk.App) {
    environments := []struct {
        Name       string
        Domain     string
        BucketName string
        PriceClass awscloudfront.PriceClass
    }{
        {
            Name:       "dev",
            Domain:     "dev-app.miempresa.com",
            BucketName: "mi-app-dev",
            PriceClass: awscloudfront.PriceClass_PRICE_CLASS_100,
        },
        {
            Name:       "staging",
            Domain:     "staging-app.miempresa.com",
            BucketName: "mi-app-staging",
            PriceClass: awscloudfront.PriceClass_PRICE_CLASS_100,
        },
        {
            Name:       "prod",
            Domain:     "app.miempresa.com",
            BucketName: "mi-app-prod",
            PriceClass: awscloudfront.PriceClass_PRICE_CLASS_ALL,
        },
    }

    for _, env := range environments {
        stack := awscdk.NewStack(app, jsii.String(fmt.Sprintf("Stack-%s", env.Name)), nil)

        bucket := awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
            BucketName: jsii.String(env.BucketName),
        })

        props := cloudfront.CloudFrontPropertiesV2{
            OriginType:                  cloudfront.OriginTypeS3,
            S3Bucket:                    bucket,
            Comment:                     fmt.Sprintf("App %s", env.Name),
            DomainNames:                 []string{env.Domain},
            CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/multi-env",
            AutoConfigureS3BucketPolicy: true,
        }

        cloudfront.NewDistributionV2(stack, "Distribution", props)
    }
}
```

---

## Seguridad

### Modelo de Seguridad por Capas

```
Layer 1: Network (CloudFront)
├── HTTPS enforced (REDIRECT_TO_HTTPS)
├── TLS 1.2+ only
├── IPv6 enabled (no security impact)
└── WAF integration ready

Layer 2: Application (Headers)
├── HSTS (force HTTPS 2 years)
├── X-Content-Type-Options (prevent MIME sniffing)
├── X-Frame-Options (clickjacking protection)
└── Referrer-Policy (privacy)

Layer 3: Origin Access (OAC)
├── Service Principal authentication
├── Distribution-specific access
├── No public bucket access
└── IAM condition on SourceArn

Layer 4: Storage (S3)
├── Private bucket (no public access)
├── Versioning enabled (recommended)
├── Encryption at rest (optional)
└── MFA Delete (high security)
```

### Checklist de Seguridad

#### ✅ Implementado por Defecto

- [x] HTTPS redirect obligatorio
- [x] TLS 1.2 mínimo (cuando se usa certificado)
- [x] Origin Access Control (OAC)
- [x] Security headers (HSTS, X-Frame-Options, etc.)
- [x] Bucket S3 privado
- [x] Política IAM con Least Privilege

#### ⚠️ Configuración Adicional Recomendada

- [ ] AWS WAF (rate limiting, SQL injection, XSS)
- [ ] Shield Standard (DDoS - incluido gratis)
- [ ] Shield Advanced (DDoS avanzado - $3000/mes)
- [ ] Geo Restriction (bloquear países específicos)
- [ ] Field-level Encryption (datos sensibles)
- [ ] CloudWatch alarms (4xx/5xx rates)

### Configuración WAF Recomendada

```go
// Ejemplo conceptual - requiere crear WebACL por separado
wafArn := CreateWebACL(stack)

props := cloudfront.CloudFrontPropertiesV2{
    // ... otras props
    WebAclArn: wafArn,
}
```

**WebACL típico incluye**:

1. **AWS Managed Rules**:

   - Core Rule Set (OWASP Top 10)
   - Known Bad Inputs
   - SQL Database
   - Linux Operating System
   - POSIX Operating System

2. **Rate Limiting**:

   ```
   Rule: Rate limit 2000 requests per 5 minutes per IP
   Action: Block for 10 minutes
   ```

3. **Geo Blocking**:
   ```
   Rule: Block countries CN, RU, KP
   Action: Block with 403 response
   ```

---

## Rendimiento

### Métricas de Rendimiento Típicas

| Métrica                       | Sin CloudFront | Con CloudFront | Mejora       |
| ----------------------------- | -------------- | -------------- | ------------ |
| **TTFB** (Time to First Byte) | 200-500ms      | 20-50ms        | **90%**      |
| **Page Load Time**            | 3-5s           | 0.5-1.5s       | **70%**      |
| **Bandwidth Cost**            | $0.09/GB       | $0.085/GB\*    | **5-40%**    |
| **Cache Hit Rate**            | 0%             | 85-95%         | **Infinite** |
| **Origin Requests**           | 100%           | 5-15%          | **85-95%**   |

\*Varía por región y volumen

### Análisis de Cache Hit Rate

#### Cache Hit Rate Óptimo: 85-95%

**Fórmula**:

```
Cache Hit Rate = (CloudFront Requests - Origin Requests) / CloudFront Requests × 100
```

**Ejemplo**:

```
1,000,000 requests a CloudFront
50,000 requests al origen S3
Cache Hit Rate = (1,000,000 - 50,000) / 1,000,000 = 95%
```

#### Factores que Afectan el Cache Hit Rate

| Factor               | Impacto                    | Solución                             |
| -------------------- | -------------------------- | ------------------------------------ |
| Query strings únicos | ❌ Reduce hit rate         | Ignorar query strings innecesarios   |
| Headers variables    | ❌ Fragmenta cache         | Incluir solo headers necesarios      |
| Cookies únicos       | ❌ Cache per-user          | No incluir cookies en cache key      |
| TTL muy bajo         | ❌ Expiraciones frecuentes | Aumentar TTL para contenido estático |
| Contenido dinámico   | ❌ No cacheable            | Usar behaviors separados             |

#### Optimización de Cache

**Contenido por Tipo y TTL Recomendado**:

| Tipo de Contenido     | TTL Recomendado | Razón                           |
| --------------------- | --------------- | ------------------------------- |
| **HTML** (index.html) | 5-10 minutos    | Cambios frecuentes, deployments |
| **CSS/JS con hash**   | 1 año           | Inmutable (hash en nombre)      |
| **CSS/JS sin hash**   | 1 hora          | Balance actualización/cache     |
| **Imágenes**          | 1 mes           | Cambian raramente               |
| **Fuentes**           | 1 año           | Inmutables                      |
| **API responses**     | 0 (no cache)    | Datos dinámicos                 |
| **PDFs/Documentos**   | 1 semana        | Relativamente estáticos         |

### Compresión y Ahorro de Bandwidth

#### Ratios de Compresión Típicos

| Tipo de Archivo | Tamaño Original | Comprimido (Gzip) | Ahorro     |
| --------------- | --------------- | ----------------- | ---------- |
| HTML            | 100 KB          | 15-25 KB          | **75-85%** |
| CSS             | 50 KB           | 8-12 KB           | **75-85%** |
| JavaScript      | 200 KB          | 50-70 KB          | **65-75%** |
| JSON            | 80 KB           | 10-15 KB          | **80-85%** |
| SVG             | 30 KB           | 5-8 KB            | **75-85%** |

#### Ejemplo de Ahorro Real

**Aplicación SPA típica**:

```
Bundle sin comprimir:
├── index.html:        15 KB
├── main.js:          500 KB
├── vendor.js:        800 KB
├── styles.css:       100 KB
└── fonts/icons:      200 KB
    TOTAL:          1,615 KB

Bundle comprimido (Gzip):
├── index.html:         3 KB  (80% reducción)
├── main.js:          150 KB  (70% reducción)
├── vendor.js:        250 KB  (69% reducción)
├── styles.css:        20 KB  (80% reducción)
└── fonts/icons:      200 KB  (0% - ya comprimido)
    TOTAL:            623 KB  (61% reducción total)
```

**Ahorro mensual** (1M usuarios, 2 visitas/mes):

```
Sin compresión: 1,000,000 × 2 × 1.615 MB = 3,230 GB
Con compresión: 1,000,000 × 2 × 0.623 MB = 1,246 GB
Ahorro: 1,984 GB × $0.085/GB = $168.64/mes
```

### HTTP/2 y HTTP/3 Performance

#### Comparativa de Protocolos

| Característica            | HTTP/1.1      | HTTP/2           | HTTP/3           |
| ------------------------- | ------------- | ---------------- | ---------------- |
| **Conexiones**            | 6-8 paralelas | 1 multiplexada   | 1 multiplexada   |
| **Head-of-line blocking** | ✅ Sí (TCP)   | ⚠️ Parcial (TCP) | ❌ No (QUIC/UDP) |
| **Header compression**    | ❌ No         | ✅ HPACK         | ✅ QPACK         |
| **Server Push**           | ❌ No         | ✅ Sí            | ✅ Sí            |
| **0-RTT resumption**      | ❌ No         | ⚠️ TLS 1.3 only  | ✅ Sí            |
| **Latencia típica**       | 100-200ms     | 50-80ms          | 20-50ms          |

#### Casos de Uso por Protocolo

**HTTP/1.1** (Legacy):

- Navegadores muy antiguos (IE < 11)
- Proxies corporativos antiguos
- ~5% del tráfico web

**HTTP/2** (Estándar):

- Navegadores modernos (2015+)
- Mejora dramática vs HTTP/1.1
- ~90% del tráfico web

**HTTP/3** (Futuro):

- Chrome 87+, Firefox 88+, Safari 14+
- Óptimo para mobile y redes inestables
- ~15% del tráfico web (creciendo)

---

## Monitoreo y Observabilidad

### CloudWatch Metrics Disponibles

#### Métricas Estándar (Sin Costo Adicional)

```
Namespace: AWS/CloudFront
```

| Métrica             | Descripción                | Uso                      |
| ------------------- | -------------------------- | ------------------------ |
| **Requests**        | Total de requests          | Volumetría               |
| **BytesDownloaded** | Bytes enviados a viewers   | Bandwidth usage          |
| **BytesUploaded**   | Bytes recibidos de viewers | Upload traffic           |
| **4xxErrorRate**    | % de errores 4xx           | Client errors            |
| **5xxErrorRate**    | % de errores 5xx           | Origin/CloudFront errors |

#### Métricas Adicionales (Costo Extra)

```
EnableAdditionalMetrics: true
Costo: $0.01 por métrica por distribución por mes
```

| Métrica                  | Descripción                        | Valor           |
| ------------------------ | ---------------------------------- | --------------- |
| **CacheHitRate**         | % de requests servidos desde cache | 85-95% ideal    |
| **OriginLatency**        | Latencia al origen (P50, P95, P99) | < 100ms ideal   |
| **ErrorRate por código** | 401, 403, 404, 500, 502, 503, 504  | Troubleshooting |

### Dashboard CloudWatch Recomendado

```go
// Ejemplo conceptual para crear dashboard
func CreateCloudFrontDashboard(distribution awscloudfront.Distribution) {
    dashboard := cloudwatch.NewDashboard(scope, "CFDashboard", &cloudwatch.DashboardProps{
        DashboardName: jsii.String("CloudFront-Production"),
    })

    // Widget 1: Requests
    dashboard.AddWidgets(cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
        Title: jsii.String("Total Requests"),
        Left: &[]cloudwatch.IMetric{
            distribution.MetricRequests(),
        },
    }))

    // Widget 2: Cache Hit Rate
    dashboard.AddWidgets(cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
        Title: jsii.String("Cache Performance"),
        Left: &[]cloudwatch.IMetric{
            distribution.MetricCacheHitRate(),
        },
    }))

    // Widget 3: Error Rates
    dashboard.AddWidgets(cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
        Title: jsii.String("Error Rates"),
        Left: &[]cloudwatch.IMetric{
            distribution.Metric4xxErrorRate(),
            distribution.Metric5xxErrorRate(),
        },
    }))
}
```

### Alarmas Recomendadas

#### Alarma 1: High Error Rate

```go
distribution.Metric5xxErrorRate().CreateAlarm(scope, "HighErrorRate", &cloudwatch.CreateAlarmOptions{
    Threshold:          jsii.Number(5), // 5% de errores
    EvaluationPeriods:  jsii.Number(2),
    DatapointsToAlarm:  jsii.Number(2),
    TreatMissingData:   cloudwatch.TreatMissingData_NOT_BREACHING,
    AlarmDescription:   jsii.String("CloudFront 5xx error rate > 5%"),
    ComparisonOperator: cloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
})
```

**Acciones**:

- SNS notification → Email/Slack
- Lambda para auto-remediation
- PagerDuty para on-call

#### Alarma 2: Low Cache Hit Rate

```go
distribution.MetricCacheHitRate().CreateAlarm(scope, "LowCacheHit", &cloudwatch.CreateAlarmOptions{
    Threshold:          jsii.Number(70), // < 70% cache hit
    EvaluationPeriods:  jsii.Number(3),
    DatapointsToAlarm:  jsii.Number(3),
    TreatMissingData:   cloudwatch.TreatMissingData_IGNORE,
    AlarmDescription:   jsii.String("Cache hit rate dropped below 70%"),
    ComparisonOperator: cloudwatch.ComparisonOperator_LESS_THAN_THRESHOLD,
})
```

**Investigar**:

- Query strings invalidando cache
- Deployment reciente
- TTL muy bajo
- Cambios en headers

#### Alarma 3: Origin Latency

```go
distribution.MetricOriginLatency().CreateAlarm(scope, "HighOriginLatency", &cloudwatch.CreateAlarmOptions{
    Threshold:          jsii.Number(1000), // 1 segundo
    EvaluationPeriods:  jsii.Number(2),
    DatapointsToAlarm:  jsii.Number(2),
    Statistic:          "p99",
    AlarmDescription:   jsii.String("Origin latency P99 > 1s"),
    ComparisonOperator: cloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
})
```

**Causas comunes**:

- S3 throttling (rate limiting)
- Región del bucket lejana
- Objetos muy grandes
- Bucket policy incorrecta

### Access Logs

#### Formato de Logs

```go
props := CloudFrontPropertiesV2{
    EnableAccessLogging: true,
    // Logs se escriben automáticamente al bucket S3 de la distribución
    // o especificar bucket dedicado si se extiende la implementación
}
```

**Estructura de log típica**:

```
#Version: 1.0
#Fields: date time x-edge-location sc-bytes c-ip cs-method cs(Host) cs-uri-stem sc-status cs(Referer) cs(User-Agent) cs-uri-query cs(Cookie) x-edge-result-type x-edge-request-id x-host-header cs-protocol cs-bytes time-taken x-forwarded-for ssl-protocol ssl-cipher x-edge-response-result-type cs-protocol-version fle-status fle-encrypted-fields c-port time-to-first-byte x-edge-detailed-result-type sc-content-type sc-content-len sc-range-start sc-range-end
2024-01-15 12:30:45 LAX3 2847 203.0.113.5 GET d111111abcdef8.cloudfront.net /index.html 200 https://example.com Mozilla/5.0... - - Hit 3RTJ8X6KAF9Q1A== d111111abcdef8.cloudfront.net https 500 0.045 - TLSv1.3 TLS_AES_128_GCM_SHA256 Hit HTTP/2.0 - - 443 0.043 Hit text/html 2847 - -
```

**Campos clave**:

- `x-edge-result-type`: Hit, Miss, Error
- `sc-status`: HTTP status code
- `time-taken`: Request duration
- `x-edge-location`: Edge location que sirvió el request
- `cs-uri-stem`: Path solicitado
- `c-ip`: IP del cliente

#### Análisis de Logs con Athena

```sql
-- Query típico: Cache Hit Rate por hora
SELECT
    DATE_TRUNC('hour', from_iso8601_timestamp(CONCAT(date, 'T', time, 'Z'))) AS hour,
    COUNT(*) AS total_requests,
    SUM(CASE WHEN "x-edge-result-type" = 'Hit' THEN 1 ELSE 0 END) AS cache_hits,
    ROUND(100.0 * SUM(CASE WHEN "x-edge-result-type" = 'Hit' THEN 1 ELSE 0 END) / COUNT(*), 2) AS cache_hit_rate
FROM cloudfront_logs
WHERE date = '2024-01-15'
GROUP BY 1
ORDER BY 1;

-- Top 10 URLs más solicitadas
SELECT
    "cs-uri-stem" AS path,
    COUNT(*) AS requests,
    SUM("sc-bytes") AS total_bytes,
    AVG("time-taken") AS avg_latency
FROM cloudfront_logs
WHERE date = '2024-01-15'
GROUP BY 1
ORDER BY 2 DESC
LIMIT 10;

-- Errores 4xx/5xx
SELECT
    "sc-status" AS status_code,
    COUNT(*) AS count,
    "cs-uri-stem" AS path
FROM cloudfront_logs
WHERE date = '2024-01-15'
    AND ("sc-status" >= 400)
GROUP BY 1, 3
ORDER BY 2 DESC;
```

---

## Troubleshooting

### Problemas Comunes y Soluciones

#### Problema 1: Distribution No Sirve Contenido (403 Forbidden)

**Síntomas**:

```
curl https://d111111abcdef8.cloudfront.net/
<AccessDenied>
<Message>Access Denied</Message>
</AccessDenied>
```

**Diagnóstico**:

1. Verificar S3 bucket policy
2. Verificar OAC configuration
3. Verificar que los objetos existen en S3

**Solución**:

```go
// Asegurar AutoConfigureS3BucketPolicy: true
props := CloudFrontPropertiesV2{
    AutoConfigureS3BucketPolicy: true, // ✅ Esto configura automáticamente la política
}
```

**Verificación manual**:

```bash
# Verificar que la política existe
aws s3api get-bucket-policy --bucket mi-bucket

# Verificar que OAC está asociado
aws cloudfront get-distribution --id E1ABCDEFGHIJK | jq '.Distribution.DistributionConfig.Origins[0].S3OriginConfig'
```

#### Problema 2: SPA Routes Devuelven 404

**Síntomas**:

- `https://app.com/` funciona
- `https://app.com/users/123` devuelve 404
- Refresh en rutas internas causa error

**Causa**:
CloudFront busca `/users/123` como archivo en S3, no existe.

**Solución**:
La estrategia ya implementa esto automáticamente con:

```go
ErrorResponses: &[]*awscloudfront.ErrorResponse{
    {HttpStatus: 403, ResponseHttpStatus: 200, ResponsePagePath: "/index.html"},
    {HttpStatus: 404, ResponseHttpStatus: 200, ResponsePagePath: "/index.html"},
}
```

**Verificar**:

```bash
# Test directo
curl -I https://d111111abcdef8.cloudfront.net/users/123
# Debe devolver 200 y servir index.html
```

#### Problema 3: Cambios No Se Reflejan (Cache Stale)

**Síntomas**:

- Nuevo deployment realizado
- S3 tiene archivos actualizados
- CloudFront sirve versión antigua

**Causa**:
CloudFront cache no ha expirado (TTL no alcanzado).

**Solución Inmediata - Invalidación**:

```bash
# Invalidar todos los archivos
aws cloudfront create-invalidation \
    --distribution-id E1ABCDEFGHIJK \
    --paths "/*"

# Invalidar paths específicos
aws cloudfront create-invalidation \
    --distribution-id E1ABCDEFGHIJK \
    --paths "/index.html" "/static/js/*"
```

**Solución Permanente - Versioning**:

```javascript
// webpack.config.js - Cache busting
output: {
    filename: '[name].[contenthash].js',
    chunkFilename: '[name].[contenthash].chunk.js'
}

// Resultado:
// main.a3f8e2c4.js  ← Hash único por versión
// vendor.9f2d1b7e.js
```

**Mejores Prácticas**:

1. ✅ Usar hashes en nombres de archivo (JS/CSS)
2. ✅ Invalidar solo `index.html` en cada deploy
3. ❌ Evitar `/*` (cuenta contra límite de 1000 invalidaciones/mes)

#### Problema 4: Certificado SSL No Funciona

**Síntomas**:

```
curl https://app.miempresa.com
curl: (60) SSL certificate problem: certificate has expired
```

**Causas comunes**:

1. Certificado en región incorrecta (debe ser us-east-1)
2. Certificado no validado
3. CNAME no configurado en DNS

**Solución**:

```go
// 1. Certificado DEBE estar en us-east-1
// 2. Verificar estado
aws acm describe-certificate \
    --certificate-arn arn:aws:acm:us-east-1:123:certificate/abc \
    --region us-east-1

// Output debe mostrar:
// "Status": "ISSUED"
// "DomainValidationOptions": [ ... "ValidationStatus": "SUCCESS" ]

// 3. Configurar DNS
// Route 53 o DNS provider:
app.miempresa.com  CNAME  d111111abcdef8.cloudfront.net
```

#### Problema 5: Alto 5xx Error Rate

**Síntomas**:

- CloudWatch muestra 5xx error rate > 5%
- Requests fallan intermitentemente

**Diagnóstico**:

```bash
# Ver logs de CloudFront
aws s3 cp s3://mi-cloudfront-logs/E1ABC/ . --recursive

# Buscar errores 5xx
grep -E "50[0-9]" *.gz | gunzip

# Verificar salud del origen
aws s3api head-object --bucket mi-bucket --key index.html
```

**Causas comunes**:

| Error   | Causa               | Solución                                      |
| ------- | ------------------- | --------------------------------------------- |
| **502** | Origen inaccesible  | Verificar bucket policy, OAC                  |
| **503** | Origen sobrecargado | Aumentar capacity, usar Origin Shield         |
| **504** | Timeout origen      | Aumentar `OriginReadTimeout` (no aplica a S3) |

**Para S3 específicamente**:

```bash
# Verificar throttling
aws cloudwatch get-metric-statistics \
    --namespace AWS/S3 \
    --metric-name 4xxErrors \
    --dimensions Name=BucketName,Value=mi-bucket \
    --start-time 2024-01-15T00:00:00Z \
    --end-time 2024-01-15T23:59:59Z \
    --period 3600 \
    --statistics Sum

# Si hay throttling (rate limiting):
# Solución: Implementar Origin Shield o solicitar aumento de límites
```

#### Problema 6: Lentitud Intermitente

**Síntomas**:

- Algunos usuarios reportan lentitud
- Otros usuarios sin problemas
- CloudWatch metrics parecen normales

**Diagnóstico por Región**:

```bash
# Logs por edge location
grep "LAX" cloudfront-logs.gz | wc -l  # Los Angeles
grep "NRT" cloudfront-logs.gz | wc -l  # Tokyo
grep "GRU" cloudfront-logs.gz | wc -l  # São Paulo

# Verificar latency por edge
SELECT
    "x-edge-location",
    AVG("time-taken") AS avg_latency,
    COUNT(*) AS requests
FROM cloudfront_logs
GROUP BY 1
ORDER BY 2 DESC;
```

**Soluciones**:

1. **Problema: Edge locations lejanos**

```go
// Cambiar Price Class
props.PriceClass = awscloudfront.PriceClass_PRICE_CLASS_200 // Incluye más regiones
```

2. **Problema: Origen en región lejana**

```go
// Implementar Origin Shield (requiere extensión de la estrategia)
// Origin Shield en la región del bucket reduce latencia
```

3. **Problema: Objetos grandes**

```bash
# Optimizar assets
# - Imágenes: WebP format, responsive images
# - Videos: Adaptive bitrate streaming
# - JavaScript: Code splitting, lazy loading
```

---

## Costos y Optimización

### Estructura de Costos CloudFront

#### Componentes de Costo

```
CloudFront Cost =
    Data Transfer Out Cost +
    Request Cost +
    Additional Features Cost
```

#### 1. Data Transfer Out (Mayor costo - ~70-80%)

**Precios por Región (Tier 1 - primeros 10 TB/mes)**:

| Región                         | Costo/GB |
| ------------------------------ | -------- |
| **US, Europe, Canada**         | $0.085   |
| **Asia Pacific (excl. India)** | $0.140   |
| **South America**              | $0.250   |
| **India**                      | $0.170   |
| **Australia, New Zealand**     | $0.140   |

**Ejemplo cálculo**:

```
Sitio web SPA:
- 100,000 usuarios únicos/mes
- 2 MB bundle comprimido
- 3 páginas visitadas en promedio

Data transfer = 100,000 × 2 MB × 3 = 600 GB
Costo = 600 GB × $0.085 = $51.00/mes
```

#### 2. Request Cost (~20-30%)

**Precios (US/Europe)**:

| Tipo Request | Costo por 10,000 requests |
| ------------ | ------------------------- |
| **HTTP**     | $0.0075                   |
| **HTTPS**    | $0.0100                   |

**Ejemplo cálculo**:

```
100,000 usuarios × 50 requests = 5,000,000 requests
Costo HTTPS = (5,000,000 / 10,000) × $0.01 = $5.00/mes
```

#### 3. Features Adicionales

| Feature                    | Costo                                        |
| -------------------------- | -------------------------------------------- |
| **Invalidations**          | Primeras 1000/mes gratis, $0.005 c/u después |
| **Field-Level Encryption** | $0.02 por 10,000 requests                    |
| **Real-time Logs**         | $0.01 por 1,000,000 log lines                |
| **Dedicated IP (VIP SSL)** | $600/mes                                     |

### Cálculo de Costo Total - Ejemplo Real

#### Escenario: Startup SPA

```
Tráfico:
├── 500,000 usuarios únicos/mes
├── 5 páginas vistas/usuario
├── 1.5 MB bundle comprimido (gzip)
└── Región: US/Europe

Cálculos:
1. Data Transfer:
   500,000 × 5 × 1.5 MB = 3,750 GB
   3,750 GB × $0.085 = $318.75

2. Requests:
   500,000 × 5 × 20 resources = 50,000,000 requests
   (50,000,000 / 10,000) × $0.01 = $500.00

3. Invalidations:
   10 deploys/mes × 1 path (/index.html) = 10 invalidations
   Gratis (< 1000)

TOTAL: $318.75 + $500.00 = $818.75/mes
```

#### Escenario: Enterprise Application

```
Tráfico:
├── 5,000,000 usuarios únicos/mes
├── 20 páginas vistas/usuario
├── 2 MB bundle comprimido
├── Región: Global (multi-region)
└── WAF enabled

Cálculos:
1. Data Transfer (weighted average):
   100,000,000 páginas × 2 MB = 200,000 GB
   - US/EU (60%): 120,000 GB × $0.075 = $9,000
   - APAC (30%): 60,000 GB × $0.120 = $7,200
   - LATAM (10%): 20,000 GB × $0.200 = $4,000
   Subtotal: $20,200

2. Requests:
   100,000,000 × 30 resources = 3,000,000,000 requests
   (3,000,000,000 / 10,000) × $0.01 = $30,000

3. WAF:
   3,000,000,000 requests × $0.60 per million = $1,800

4. Additional Metrics:
   1 distribution × $0.01 × 10 metrics = $0.10/mes

TOTAL: $20,200 + $30,000 + $1,800 + $0.10 = $52,000.10/mes
```

### Optimización de Costos

#### Estrategia 1: Aumentar Cache Hit Rate

**Impacto**: Reducir requests y data transfer al origen.

```go
// Aumentar TTL para contenido estático
// (requiere custom cache policy - extensión futura)

// Resultado:
// Cache Hit Rate: 80% → 95%
// Origin Requests: 20% → 5%
// Ahorro: 15% de requests al origen
```

#### Estrategia 2: Comprimir Agresivamente

```javascript
// webpack.config.js
const CompressionPlugin = require("compression-webpack-plugin");

module.exports = {
  plugins: [
    new CompressionPlugin({
      algorithm: "brotli", // Mejor que gzip
      test: /\.(js|css|html|svg)$/,
      threshold: 10240, // Solo > 10KB
      minRatio: 0.8,
    }),
  ],
};

// Resultado:
// Bundle: 2 MB → 600 KB
// Ahorro: 70% en data transfer
```

#### Estrategia 3: Lazy Loading y Code Splitting

```javascript
// React Router lazy loading
const Dashboard = lazy(() => import("./Dashboard"));
const Profile = lazy(() => import("./Profile"));

// Resultado:
// Initial bundle: 2 MB → 500 KB
// On-demand: 4 × 375 KB
// First load ahorro: 75%
```

#### Estrategia 4: Image Optimization

```html
<!-- Usar WebP con fallback -->
<picture>
  <source srcset="image.webp" type="image/webp" />
  <img src="image.jpg" alt="..." />
</picture>

<!-- Responsive images -->
<img
  srcset="small.jpg 480w, medium.jpg 800w, large.jpg 1200w"
  sizes="(max-width: 600px) 480px, (max-width: 1000px) 800px, 1200px"
  src="medium.jpg"
  alt="..."
/>
```

**Ahorro típico**:

- JPEG → WebP: 25-35% reducción
- Responsive images: 40-60% reducción (mobile)

#### Estrategia 5: Price Class Optimization

```go
// Para audiencia regional (US/EU)
props.PriceClass = awscloudfront.PriceClass_PRICE_CLASS_100

// Ahorro vs PRICE_CLASS_ALL:
// - 30-40% en data transfer costs
// - Trade-off: Latency en otras regiones
```

**Decisión Matrix**:

| Audiencia         | Price Class | Ahorro | Latencia APAC |
| ----------------- | ----------- | ------ | ------------- |
| 90% US/EU         | 100         | 30-40% | +100-200ms    |
| Global balanceado | 200         | 15-20% | +50-100ms     |
| Global premium    | ALL         | 0%     | Óptima        |

### ROI Analysis: CloudFront vs Direct S3

#### Comparación de Costos

**Escenario**: 1,000,000 requests/mes, 2 MB average object size

| Métrica                 | S3 Direct | CloudFront | Diferencia   |
| ----------------------- | --------- | ---------- | ------------ |
| **Data Transfer**       | $180      | $170       | -$10 (-5.5%) |
| **Requests**            | $4        | $10        | +$6 (+150%)  |
| **Latency (avg)**       | 200ms     | 50ms       | **-75%**     |
| **Cache Hit (compute)** | N/A       | 90% cache  | -$162        |
| **TOTAL COST**          | $184/mes  | $18/mes\*  | **-90%**     |

\*90% cache hit = solo 10% de requests llegan a S3

**Conclusión**: CloudFront **siempre** es más económico para contenido servido múltiples veces.

---

## Mejores Prácticas

### 1. Deployment y CI/CD

#### Estrategia de Deployment Recomendada

```bash
#!/bin/bash
# deploy.sh

# 1. Build
npm run build

# 2. Upload a S3 (con cache headers)
aws s3 sync build/ s3://mi-bucket/ \
    --exclude "*.html" \
    --cache-control "public, max-age=31536000, immutable"

# 3. Upload HTML (sin cache)
aws s3 sync build/ s3://mi-bucket/ \
    --exclude "*" \
    --include "*.html" \
    --cache-control "public, max-age=0, must-revalidate"

# 4. Invalidar solo HTML
aws cloudfront create-invalidation \
    --distribution-id $DISTRIBUTION_ID \
    --paths "/
```

# Documentación: S3 CloudFront Strategy (Continuación)

---

## Mejores Prácticas (Continuación)

### 1. Deployment y CI/CD (Continuación)

#### Estrategia de Deployment Recomendada (Continuación)

```bash
#!/bin/bash
# deploy.sh (continuación)

# 4. Invalidar solo HTML
aws cloudfront create-invalidation \
    --distribution-id $DISTRIBUTION_ID \
    --paths "/index.html" "/404.html"

# 5. Esperar invalidación (opcional)
aws cloudfront wait invalidation-completed \
    --distribution-id $DISTRIBUTION_ID \
    --id $INVALIDATION_ID

echo "✅ Deployment completado"
```

#### GitHub Actions Pipeline

```yaml
# .github/workflows/deploy.yml
name: Deploy to CloudFront

on:
  push:
    branches: [main]

env:
  AWS_REGION: us-east-1
  S3_BUCKET: mi-app-prod
  DISTRIBUTION_ID: E1ABCDEFGHIJK

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: "18"
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build
        env:
          CI: true
          REACT_APP_VERSION: ${{ github.sha }}

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Sync assets to S3 (with long cache)
        run: |
          aws s3 sync build/ s3://${{ env.S3_BUCKET }}/ \
            --exclude "*.html" \
            --exclude "service-worker.js" \
            --cache-control "public, max-age=31536000, immutable" \
            --delete

      - name: Sync HTML to S3 (no cache)
        run: |
          aws s3 sync build/ s3://${{ env.S3_BUCKET }}/ \
            --exclude "*" \
            --include "*.html" \
            --include "service-worker.js" \
            --cache-control "public, max-age=0, must-revalidate"

      - name: Invalidate CloudFront
        run: |
          aws cloudfront create-invalidation \
            --distribution-id ${{ env.DISTRIBUTION_ID }} \
            --paths "/index.html" "/service-worker.js"

      - name: Deployment summary
        run: |
          echo "🚀 Deployment completed successfully"
          echo "📦 Build: ${{ github.sha }}"
          echo "🌐 URL: https://app.miempresa.com"
```

#### Blue-Green Deployment con CloudFront

```go
// Estrategia avanzada: Dos buckets para blue-green
func DeployBlueGreen(stack constructs.Construct) {
    // Bucket Blue (actual producción)
    blueBucket := awss3.NewBucket(stack, jsii.String("BlueBucket"), &awss3.BucketProps{
        BucketName: jsii.String("mi-app-blue"),
    })

    // Bucket Green (nuevo deployment)
    greenBucket := awss3.NewBucket(stack, jsii.String("GreenBucket"), &awss3.BucketProps{
        BucketName: jsii.String("mi-app-green"),
    })

    // Distribution apunta a Blue inicialmente
    props := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    blueBucket, // Switch a greenBucket para switch
        Comment:                     "Blue-Green Deployment",
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(stack, "Distribution", props)
}
```

**Proceso**:

1. Deploy a Green bucket
2. Test en Green (URL S3 directo o distribution de staging)
3. Switch distribution de Blue → Green (CDK update)
4. Rollback: Switch Green → Blue si hay problemas

### 2. Cache Strategy por Tipo de Contenido

#### Cache Headers Recomendados

```javascript
// S3 upload con cache headers específicos
const s3Client = new S3Client({ region: "us-east-1" });

// Assets con hash (inmutables)
await s3Client.send(
  new PutObjectCommand({
    Bucket: "mi-bucket",
    Key: "static/js/main.a3f8e2c4.js",
    Body: fileContent,
    ContentType: "application/javascript",
    CacheControl: "public, max-age=31536000, immutable", // 1 año
  })
);

// HTML (actualizable)
await s3Client.send(
  new PutObjectCommand({
    Bucket: "mi-bucket",
    Key: "index.html",
    Body: htmlContent,
    ContentType: "text/html",
    CacheControl: "public, max-age=0, must-revalidate", // No cache
  })
);

// Imágenes (rara vez cambian)
await s3Client.send(
  new PutObjectCommand({
    Bucket: "mi-bucket",
    Key: "images/logo.png",
    Body: imageContent,
    ContentType: "image/png",
    CacheControl: "public, max-age=2592000", // 30 días
  })
);
```

#### Cache Control Directives

| Directive           | Significado                      | Uso                            |
| ------------------- | -------------------------------- | ------------------------------ |
| **public**          | Cacheable por CDN y browser      | Contenido público              |
| **private**         | Solo browser cache               | Contenido con datos de usuario |
| **no-cache**        | Validar con origen antes de usar | APIs, datos dinámicos          |
| **no-store**        | Nunca cachear                    | Datos sensibles                |
| **max-age=N**       | TTL en segundos                  | Control de duración            |
| **immutable**       | Nunca revalidar antes de expirar | Assets con hash                |
| **must-revalidate** | Revalidar si expiró              | HTML, SPAs                     |

#### Matriz de Decisión

| Tipo Archivo          | Cache-Control                         | Razón                            |
| --------------------- | ------------------------------------- | -------------------------------- |
| **JS/CSS con hash**   | `public, max-age=31536000, immutable` | Nunca cambia (hash único)        |
| **index.html**        | `public, max-age=0, must-revalidate`  | Siempre actual                   |
| **service-worker.js** | `public, max-age=0`                   | Debe actualizarse inmediatamente |
| **manifest.json**     | `public, max-age=86400`               | 1 día (rara vez cambia)          |
| **Imágenes**          | `public, max-age=2592000`             | 30 días                          |
| **Fuentes**           | `public, max-age=31536000`            | 1 año (inmutables)               |
| **Videos**            | `public, max-age=604800`              | 1 semana                         |
| **API JSON**          | `private, max-age=0, no-cache`        | Siempre fresh                    |

### 3. Seguridad en Profundidad

#### Security Headers Completos

```javascript
// Lambda@Edge para headers adicionales (extensión futura)
// O configurar en origen S3 (metadata)

const securityHeaders = {
  "Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",
  "X-Content-Type-Options": "nosniff",
  "X-Frame-Options": "SAMEORIGIN",
  "X-XSS-Protection": "1; mode=block",
  "Referrer-Policy": "strict-origin-when-cross-origin",
  "Permissions-Policy": "geolocation=(), microphone=(), camera=()",
  "Content-Security-Policy":
    "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.example.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https://api.example.com;",
};
```

#### Content Security Policy (CSP) para SPAs

```javascript
// CSP estricto para React SPA
const csp = [
    "default-src 'none'", // Negar todo por defecto
    "script-src 'self' 'sha256-abc123...'", // Solo scripts propios + hash inline
    "style-src 'self' 'unsafe-inline'", // Estilos propios + inline (React emotion/styled)
    "img-src 'self' data: https:", // Imágenes propias + data URIs + HTTPS
    "font-src 'self' data:", // Fuentes propias + data URIs
    "connect-src 'self' https://api.miapp.com", // APIs permitidas
    "frame-ancestors 'none'", // No embeddable
    "base-uri 'self'", // Base tag solo propio dominio
    "form-action 'self'", // Formularios solo al mismo dominio
].join('; ');

// Subir a S3 con metadata
aws s3 cp index.html s3://mi-bucket/ \
    --metadata "Content-Security-Policy=${csp}"
```

#### Configuración WAF Detallada

```go
// Ejemplo conceptual - Requiere implementación separada de WebACL
func CreateProductionWebACL(stack constructs.Construct) string {
    webACL := awswafv2.NewCfnWebACL(stack, jsii.String("WebACL"), &awswafv2.CfnWebACLProps{
        Scope: jsii.String("CLOUDFRONT"),
        DefaultAction: &awswafv2.CfnWebACL_DefaultActionProperty{
            Allow: &awswafv2.CfnWebACL_AllowActionProperty{},
        },
        Rules: &[]interface{}{
            // Rule 1: Rate limiting
            &awswafv2.CfnWebACL_RuleProperty{
                Name:     jsii.String("RateLimit"),
                Priority: jsii.Number(1),
                Statement: &awswafv2.CfnWebACL_StatementProperty{
                    RateBasedStatement: &awswafv2.CfnWebACL_RateBasedStatementProperty{
                        Limit:              jsii.Number(2000),
                        AggregateKeyType:   jsii.String("IP"),
                    },
                },
                Action: &awswafv2.CfnWebACL_RuleActionProperty{
                    Block: &awswafv2.CfnWebACL_BlockActionProperty{},
                },
                VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
                    SampledRequestsEnabled:   jsii.Bool(true),
                    CloudWatchMetricsEnabled: jsii.Bool(true),
                    MetricName:              jsii.String("RateLimitRule"),
                },
            },

            // Rule 2: AWS Managed Rules - Core Rule Set
            &awswafv2.CfnWebACL_RuleProperty{
                Name:     jsii.String("AWSManagedRulesCommonRuleSet"),
                Priority: jsii.Number(2),
                Statement: &awswafv2.CfnWebACL_StatementProperty{
                    ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
                        VendorName: jsii.String("AWS"),
                        Name:       jsii.String("AWSManagedRulesCommonRuleSet"),
                    },
                },
                OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
                    None: &map[string]interface{}{},
                },
                VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
                    SampledRequestsEnabled:   jsii.Bool(true),
                    CloudWatchMetricsEnabled: jsii.Bool(true),
                    MetricName:              jsii.String("AWSManagedRulesCommonMetric"),
                },
            },

            // Rule 3: Geo blocking
            &awswafv2.CfnWebACL_RuleProperty{
                Name:     jsii.String("GeoBlock"),
                Priority: jsii.Number(3),
                Statement: &awswafv2.CfnWebACL_StatementProperty{
                    GeoMatchStatement: &awswafv2.CfnWebACL_GeoMatchStatementProperty{
                        CountryCodes: jsii.Strings("CN", "RU", "KP"),
                    },
                },
                Action: &awswafv2.CfnWebACL_RuleActionProperty{
                    Block: &awswafv2.CfnWebACL_BlockActionProperty{},
                },
                VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
                    SampledRequestsEnabled:   jsii.Bool(true),
                    CloudWatchMetricsEnabled: jsii.Bool(true),
                    MetricName:              jsii.String("GeoBlockRule"),
                },
            },
        },
        VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
            SampledRequestsEnabled:   jsii.Bool(true),
            CloudWatchMetricsEnabled: jsii.Bool(true),
            MetricName:              jsii.String("WebACLMetric"),
        },
    })

    return *webACL.AttrArn()
}
```

### 4. Monitoreo Proactivo

#### CloudWatch Dashboard Completo

```go
func CreateComprehensiveDashboard(
    scope constructs.Construct,
    distribution awscloudfront.Distribution,
) {
    dashboard := cloudwatch.NewDashboard(scope, jsii.String("CFDashboard"), &cloudwatch.DashboardProps{
        DashboardName: jsii.String("CloudFront-Production-Monitoring"),
    })

    // Row 1: Traffic Overview
    dashboard.AddWidgets(
        cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
            Title:  jsii.String("Total Requests"),
            Width:  jsii.Number(12),
            Height: jsii.Number(6),
            Left: &[]cloudwatch.IMetric{
                distribution.MetricRequests(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("Sum"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                }),
            },
        }),
        cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
            Title:  jsii.String("Bandwidth"),
            Width:  jsii.Number(12),
            Height: jsii.Number(6),
            Left: &[]cloudwatch.IMetric{
                distribution.MetricBytesDownloaded(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("Sum"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                    Unit:      cloudwatch.Unit_GIGABYTES,
                }),
            },
        }),
    )

    // Row 2: Performance
    dashboard.AddWidgets(
        cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
            Title:  jsii.String("Cache Performance"),
            Width:  jsii.Number(12),
            Height: jsii.Number(6),
            Left: &[]cloudwatch.IMetric{
                distribution.MetricCacheHitRate(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("Average"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                }),
            },
            LeftYAxis: &cloudwatch.YAxisProps{
                Min: jsii.Number(0),
                Max: jsii.Number(100),
            },
        }),
        cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
            Title:  jsii.String("Origin Latency (ms)"),
            Width:  jsii.Number(12),
            Height: jsii.Number(6),
            Left: &[]cloudwatch.IMetric{
                distribution.MetricOriginLatency(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("p50"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                }),
                distribution.MetricOriginLatency(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("p99"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                }),
            },
        }),
    )

    // Row 3: Errors
    dashboard.AddWidgets(
        cloudwatch.NewGraphWidget(&cloudwatch.GraphWidgetProps{
            Title:  jsii.String("Error Rates"),
            Width:  jsii.Number(24),
            Height: jsii.Number(6),
            Left: &[]cloudwatch.IMetric{
                distribution.Metric4xxErrorRate(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("Average"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                    Color:     cloudwatch.Color_ORANGE(),
                }),
                distribution.Metric5xxErrorRate(&cloudwatch.MetricOptions{
                    Statistic: jsii.String("Average"),
                    Period:    awscdk.Duration_Minutes(jsii.Number(5)),
                    Color:     cloudwatch.Color_RED(),
                }),
            },
        }),
    )
}
```

#### Alarmas de Producción

```go
func CreateProductionAlarms(
    scope constructs.Construct,
    distribution awscloudfront.Distribution,
    snsTopicArn string,
) {
    topic := awssns.Topic_FromTopicArn(
        scope,
        jsii.String("AlertTopic"),
        jsii.String(snsTopicArn),
    )

    // Alarma 1: Critical - High 5xx Rate
    alarm5xx := distribution.Metric5xxErrorRate().CreateAlarm(
        scope,
        jsii.String("Critical5xxRate"),
        &cloudwatch.CreateAlarmOptions{
            AlarmName:          jsii.String("CF-Critical-5xx-Error-Rate"),
            Threshold:          jsii.Number(5),
            EvaluationPeriods:  jsii.Number(2),
            DatapointsToAlarm:  jsii.Number(2),
            TreatMissingData:   cloudwatch.TreatMissingData_NOT_BREACHING,
            ComparisonOperator: cloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
            AlarmDescription:   jsii.String("CloudFront 5xx error rate exceeded 5%"),
        },
    )
    alarm5xx.AddAlarmAction(cloudwatch.NewSnsAction(topic))

    // Alarma 2: Warning - Cache Hit Rate Drop
    alarmCache := distribution.MetricCacheHitRate().CreateAlarm(
        scope,
        jsii.String("LowCacheHitRate"),
        &cloudwatch.CreateAlarmOptions{
            AlarmName:          jsii.String("CF-Warning-Low-Cache-Hit"),
            Threshold:          jsii.Number(70),
            EvaluationPeriods:  jsii.Number(3),
            DatapointsToAlarm:  jsii.Number(3),
            TreatMissingData:   cloudwatch.TreatMissingData_IGNORE,
            ComparisonOperator: cloudwatch.ComparisonOperator_LESS_THAN_THRESHOLD,
            AlarmDescription:   jsii.String("Cache hit rate dropped below 70%"),
        },
    )
    alarmCache.AddAlarmAction(cloudwatch.NewSnsAction(topic))

    // Alarma 3: Critical - Origin Unreachable
    alarmOrigin := distribution.Metric5xxErrorRate().CreateAlarm(
        scope,
        jsii.String("OriginUnreachable"),
        &cloudwatch.CreateAlarmOptions{
            AlarmName:          jsii.String("CF-Critical-Origin-Down"),
            Threshold:          jsii.Number(50),
            EvaluationPeriods:  jsii.Number(1),
            DatapointsToAlarm:  jsii.Number(1),
            TreatMissingData:   cloudwatch.TreatMissingData_BREACHING,
            ComparisonOperator: cloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
            AlarmDescription:   jsii.String("Origin appears to be down (>50% 5xx errors)"),
        },
    )
    alarmOrigin.AddAlarmAction(cloudwatch.NewSnsAction(topic))

    // Alarma 4: Info - High Traffic Spike
    alarmTraffic := distribution.MetricRequests().CreateAlarm(
        scope,
        jsii.String("HighTrafficSpike"),
        &cloudwatch.CreateAlarmOptions{
            AlarmName:          jsii.String("CF-Info-Traffic-Spike"),
            Threshold:          jsii.Number(100000),
            EvaluationPeriods:  jsii.Number(1),
            DatapointsToAlarm:  jsii.Number(1),
            TreatMissingData:   cloudwatch.TreatMissingData_NOT_BREACHING,
            ComparisonOperator: cloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
            AlarmDescription:   jsii.String("Unusual traffic spike detected (>100k req/5min)"),
        },
    )
    alarmTraffic.AddAlarmAction(cloudwatch.NewSnsAction(topic))
}
```

### 5. Testing y Validación

#### Pre-Deployment Checklist

```bash
#!/bin/bash
# pre-deploy-checks.sh

set -e

echo "🔍 Running pre-deployment checks..."

# 1. Verificar build exitoso
echo "✓ Checking build..."
if [ ! -d "build" ]; then
    echo "❌ Build directory not found"
    exit 1
fi

# 2. Verificar archivos esenciales
echo "✓ Checking essential files..."
for file in "index.html" "manifest.json"; do
    if [ ! -f "build/$file" ]; then
        echo "❌ Missing file: $file"
        exit 1
    fi
done

# 3. Verificar tamaño de bundle
echo "✓ Checking bundle size..."
BUNDLE_SIZE=$(du -sb build | awk '{print $1}')
MAX_SIZE=$((10 * 1024 * 1024)) # 10 MB
if [ $BUNDLE_SIZE -gt $MAX_SIZE ]; then
    echo "⚠️  Warning: Bundle size ${BUNDLE_SIZE} exceeds ${MAX_SIZE} bytes"
fi

# 4. Lighthouse CI (opcional)
echo "✓ Running Lighthouse..."
npm run lighthouse || echo "⚠️  Lighthouse checks failed"

# 5. Verificar CSP headers
echo "✓ Checking security headers..."
grep -q "Content-Security-Policy" build/index.html || \
    echo "⚠️  Warning: No CSP header found"

echo "✅ Pre-deployment checks completed"
```

#### Post-Deployment Validation

```bash
#!/bin/bash
# post-deploy-validation.sh

DOMAIN="https://app.miempresa.com"

echo "🧪 Running post-deployment validation..."

# 1. Health check
echo "✓ Health check..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" $DOMAIN)
if [ $STATUS -ne 200 ]; then
    echo "❌ Health check failed: HTTP $STATUS"
    exit 1
fi

# 2. Verificar HTTPS redirect
echo "✓ Checking HTTPS redirect..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://app.miempresa.com)
if [ $HTTP_STATUS -ne 301 ] && [ $HTTP_STATUS -ne 302 ]; then
    echo "⚠️  Warning: HTTP not redirecting to HTTPS"
fi

# 3. Verificar security headers
echo "✓ Checking security headers..."
HEADERS=$(curl -sI $DOMAIN)
echo "$HEADERS" | grep -q "Strict-Transport-Security" || \
    echo "⚠️  Missing: HSTS header"
echo "$HEADERS" | grep -q "X-Content-Type-Options" || \
    echo "⚠️  Missing: X-Content-Type-Options"

# 4. Verificar cache headers
echo "✓ Checking cache headers..."
CACHE_CONTROL=$(curl -sI $DOMAIN/static/js/main.*.js | grep -i "cache-control")
echo "JS Cache: $CACHE_CONTROL"

# 5. Smoke test de rutas principales
echo "✓ Smoke testing routes..."
for route in "/" "/dashboard" "/profile"; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${DOMAIN}${route}")
    if [ $STATUS -ne 200 ]; then
        echo "❌ Route $route failed: HTTP $STATUS"
        exit 1
    fi
done

# 6. Verificar CloudFront headers
echo "✓ Checking CloudFront headers..."
CF_HEADERS=$(curl -sI $DOMAIN | grep -i "x-cache")
echo "CloudFront Cache: $CF_HEADERS"

echo "✅ Post-deployment validation completed"
```

---

## Ejemplos Completos

### Ejemplo 1: Startup SaaS (React SPA)

```go
package main

import (
    "mi-proyecto/cloudfront"
    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
)

type SaaSStackProps struct {
    awscdk.StackProps
    Environment string
    Domain      string
    CertArn     string
}

func NewSaaSStack(scope constructs.Construct, id string, props *SaaSStackProps) awscdk.Stack {
    stack := awscdk.NewStack(scope, &id, &props.StackProps)

    // S3 Bucket para frontend
    bucket := awss3.NewBucket(stack, jsii.String("AppBucket"), &awss3.BucketProps{
        BucketName:        jsii.String(fmt.Sprintf("saas-app-%s", props.Environment)),
        Versioned:         jsii.Bool(true),
        Encryption:        awss3.BucketEncryption_S3_MANAGED,
        BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
        RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
        LifecycleRules: &[]*awss3.LifecycleRule{
            {
                // Limpieza de versiones antiguas
                NoncurrentVersionExpiration: awscdk.Duration_Days(jsii.Number(90)),
            },
        },
    })

    // CloudFront Distribution
    cfProps := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    bucket,
        Comment:                     fmt.Sprintf("SaaS App - %s", props.Environment),
        DomainNames:                 []string{props.Domain},
        CertificateArn:              props.CertArn,
        EnableAccessLogging:         props.Environment == "prod",
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(stack, "Distribution", cfProps)

    // Outputs
    awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
        Value:       bucket.BucketName(),
        Description: jsii.String("S3 Bucket for frontend assets"),
    })

    awscdk.NewCfnOutput(stack, jsii.String("DistributionId"), &awscdk.CfnOutputProps{
        Value:       distribution.DistributionId(),
        Description: jsii.String("CloudFront Distribution ID"),
    })

    awscdk.NewCfnOutput(stack, jsii.String("DomainName"), &awscdk.CfnOutputProps{
        Value:       distribution.DistributionDomainName(),
        Description: jsii.String("CloudFront domain name"),
    })

    awscdk.NewCfnOutput(stack, jsii.String("URL"), &awscdk.CfnOutputProps{
        Value:       jsii.String(fmt.Sprintf("https://%s", props.Domain)),
        Description: jsii.String("Application URL"),
    })

    return stack
}

func main() {
    app := awscdk.NewApp(nil)

    // Production
    NewSaaSStack(app, "SaaS-Prod", &SaaSStackProps{
        Environment: "prod",
        Domain:      "app.miempresa.com",
        CertArn:     "arn:aws:acm:us-east-1:123:certificate/prod",
        StackProps: awscdk.StackProps{
            Env: &awscdk.Environment{
                Region: jsii.String("us-east-1"),
            },
        },
    })

    // Staging
    NewSaaSStack(app, "SaaS-Staging", &SaaSStackProps{
        Environment: "staging",
        Domain:      "staging-app.miempresa.com",
        CertArn:     "arn:aws:acm:us-east-1:123:certificate/staging",
        StackProps: awscdk.StackProps{
            Env: &awscdk.Environment{
                Region: jsii.String("us-east-1"),
            },
        },
    })

    app.Synth(nil)
}
```

### Ejemplo 2: E-commerce con Assets CDN

# Documentación: S3 CloudFront Strategy (Continuación)

```go
func NewEcommerceStack(scope constructs.Construct, id string) awscdk.Stack {
    stack := awscdk.NewStack(scope, jsii.String(id), nil)

    // Bucket para frontend (SPA)
    frontendBucket := awss3.NewBucket(stack, jsii.String("FrontendBucket"), &awss3.BucketProps{
        BucketName: jsii.String("ecommerce-frontend"),
    })

    // Bucket para assets (imágenes productos)
    assetsBucket := awss3.NewBucket(stack, jsii.String("AssetsBucket"), &awss3.BucketProps{
        BucketName: jsii.String("ecommerce-assets"),
        Cors: &[]*awss3.CorsRule{
            {
                AllowedOrigins: jsii.Strings("https://shop.miempresa.com"),
                AllowedMethods: &[]awss3.HttpMethods{
                    awss3.HttpMethods_GET,
                    awss3.HttpMethods_HEAD,
                },
                AllowedHeaders: jsii.Strings("*"),
                MaxAge:         jsii.Number(3600),
            },
        },
    })

    // Distribution para frontend
    frontendProps := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    frontendBucket,
        Comment:                     "E-commerce Frontend",
        DomainNames:                 []string{"shop.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/shop",
        AutoConfigureS3BucketPolicy: true,
    }

    frontendDist := cloudfront.NewDistributionV2(
        stack,
        "FrontendDistribution",
        frontendProps,
    )

    // Distribution para assets (CDN)
    assetsProps := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    assetsBucket,
        Comment:                     "E-commerce Assets CDN",
        DomainNames:                 []string{"cdn.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/cdn",
        AutoConfigureS3BucketPolicy: true,
    }

    assetsDist := cloudfront.NewDistributionV2(
        stack,
        "AssetsDistribution",
        assetsProps,
    )

    // Outputs
    awscdk.NewCfnOutput(stack, jsii.String("ShopURL"), &awscdk.CfnOutputProps{
        Value: jsii.String("https://shop.miempresa.com"),
    })

    awscdk.NewCfnOutput(stack, jsii.String("CDNURL"), &awscdk.CfnOutputProps{
        Value: jsii.String("https://cdn.miempresa.com"),
    })

    return stack
}
```

**Uso en frontend**:

```jsx
// React component
const ProductImage = ({ productId, imageName }) => {
  const cdnUrl = "https://cdn.miempresa.com";

  return (
    <picture>
      {/* Responsive images */}
      <source
        srcSet={`${cdnUrl}/products/${productId}/${imageName}-small.webp 480w,
                 ${cdnUrl}/products/${productId}/${imageName}-medium.webp 800w,
                 ${cdnUrl}/products/${productId}/${imageName}-large.webp 1200w`}
        type="image/webp"
      />
      <img
        src={`${cdnUrl}/products/${productId}/${imageName}-medium.jpg`}
        alt="Product"
        loading="lazy"
      />
    </picture>
  );
};
```

### Ejemplo 3: Multi-Tenant SaaS

```go
type TenantConfig struct {
    TenantId    string
    Domain      string
    BucketName  string
}

func NewMultiTenantStack(scope constructs.Construct, id string) awscdk.Stack {
    stack := awscdk.NewStack(scope, jsii.String(id), nil)

    tenants := []TenantConfig{
        {
            TenantId:   "acme-corp",
            Domain:     "acme.app.miempresa.com",
            BucketName: "tenant-acme-corp",
        },
        {
            TenantId:   "globex",
            Domain:     "globex.app.miempresa.com",
            BucketName: "tenant-globex",
        },
        {
            TenantId:   "initech",
            Domain:     "initech.app.miempresa.com",
            BucketName: "tenant-initech",
        },
    }

    // Wildcard certificate para *.app.miempresa.com
    wildcardCertArn := "arn:aws:acm:us-east-1:123:certificate/wildcard"

    for _, tenant := range tenants {
        // Bucket por tenant
        bucket := awss3.NewBucket(
            stack,
            jsii.String(fmt.Sprintf("Bucket-%s", tenant.TenantId)),
            &awss3.BucketProps{
                BucketName: jsii.String(tenant.BucketName),
                Versioned:  jsii.Bool(true),
            },
        )

        // Distribution por tenant
        props := cloudfront.CloudFrontPropertiesV2{
            OriginType:                  cloudfront.OriginTypeS3,
            S3Bucket:                    bucket,
            Comment:                     fmt.Sprintf("Tenant: %s", tenant.TenantId),
            DomainNames:                 []string{tenant.Domain},
            CertificateArn:              wildcardCertArn,
            AutoConfigureS3BucketPolicy: true,
        }

        distribution := cloudfront.NewDistributionV2(
            stack,
            fmt.Sprintf("Distribution-%s", tenant.TenantId),
            props,
        )

        // Output por tenant
        awscdk.NewCfnOutput(
            stack,
            jsii.String(fmt.Sprintf("URL-%s", tenant.TenantId)),
            &awscdk.CfnOutputProps{
                Value:       jsii.String(fmt.Sprintf("https://%s", tenant.Domain)),
                Description: jsii.String(fmt.Sprintf("Tenant %s URL", tenant.TenantId)),
            },
        )
    }

    return stack
}
```

### Ejemplo 4: Progressive Web App (PWA)

```go
func NewPWAStack(scope constructs.Construct, id string) awscdk.Stack {
    stack := awscdk.NewStack(scope, jsii.String(id), nil)

    // Bucket para PWA
    bucket := awss3.NewBucket(stack, jsii.String("PWABucket"), &awss3.BucketProps{
        BucketName: jsii.String("mi-pwa-app"),
        Versioned:  jsii.Bool(true),
    })

    // Distribution
    props := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    bucket,
        Comment:                     "PWA Application",
        DomainNames:                 []string{"pwa.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/pwa",
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(stack, "PWADistribution", props)

    // NOTE: Para PWAs, service-worker.js debe tener cache especial
    // Implementar con custom cache policy en futuras versiones

    return stack
}
```

**Configuración especial de PWA**:

```javascript
// service-worker.js cache strategy
const CACHE_NAME = "pwa-v1";
const CDN_URL = "https://pwa.miempresa.com";

// Recursos para pre-cache
const urlsToCache = [
  "/",
  "/index.html",
  "/static/js/main.js",
  "/static/css/main.css",
  "/manifest.json",
  "/offline.html",
];

// Cache-first para assets estáticos
self.addEventListener("fetch", (event) => {
  if (event.request.url.startsWith(CDN_URL + "/static/")) {
    event.respondWith(
      caches.match(event.request).then((response) => {
        return response || fetch(event.request);
      })
    );
  }
});
```

---

## Migración y Actualización

### Migración desde OAI (Legacy) a OAC

Si tienes distribuciones existentes con Origin Access Identity (OAI), aquí está el proceso de migración:

#### Paso 1: Identificar Distribuciones con OAI

```bash
# Listar todas las distribuciones
aws cloudfront list-distributions \
    --query 'DistributionList.Items[*].[Id,Origins.Items[0].S3OriginConfig.OriginAccessIdentity]' \
    --output table

# Output mostrará distribuciones con OAI:
# E1ABC123 | origin-access-identity/cloudfront/E1XYZ789
```

#### Paso 2: Actualizar a la Nueva Estrategia

```go
// Código anterior (OAI - legacy)
// No usar - solo para referencia
/*
origin := awscloudfrontorigins.NewS3Origin(bucket, &awscloudfrontorigins.S3OriginProps{
    OriginAccessIdentity: oai, // DEPRECATED
})
*/

// Código nuevo (OAC - recomendado)
props := cloudfront.CloudFrontPropertiesV2{
    OriginType:                  cloudfront.OriginTypeS3,
    S3Bucket:                    existingBucket,
    AutoConfigureS3BucketPolicy: true, // Configura automáticamente OAC
    // ... otras props
}

distribution := cloudfront.NewDistributionV2(stack, "UpdatedDistribution", props)
```

#### Paso 3: Actualizar Bucket Policy

```bash
# La estrategia lo hace automáticamente con AutoConfigureS3BucketPolicy: true
# Pero si necesitas hacerlo manualmente:

cat > bucket-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowCloudFrontServicePrincipal",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudfront.amazonaws.com"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::mi-bucket/*",
      "Condition": {
        "StringEquals": {
          "AWS:SourceArn": "arn:aws:cloudfront::123456789012:distribution/E1NEWDIST"
        }
      }
    }
  ]
}
EOF

aws s3api put-bucket-policy --bucket mi-bucket --policy file://bucket-policy.json
```

#### Paso 4: Validación Post-Migración

```bash
# Test de acceso
curl -I https://d111111abcdef8.cloudfront.net/index.html
# Debe retornar 200 OK

# Verificar logs para OAC
aws cloudfront get-distribution --id E1NEWDIST \
    | jq '.Distribution.DistributionConfig.Origins[0].S3OriginConfig'
# Debe mostrar S3OriginAccessControlId en lugar de OriginAccessIdentity

# Test de acceso directo a S3 (debe fallar)
curl -I https://mi-bucket.s3.amazonaws.com/index.html
# Debe retornar 403 Forbidden (esto es correcto - bucket es privado)
```

### Actualización de Stack Existente

```go
// Scenario: Actualizar distribución existente sin downtime

func UpdateExistingDistribution(stack constructs.Construct) {
    // Importar bucket existente
    existingBucket := awss3.Bucket_FromBucketName(
        stack,
        jsii.String("ImportedBucket"),
        jsii.String("mi-bucket-existente"),
    )

    // Actualizar a nueva estrategia
    props := cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    existingBucket,
        Comment:                     "Actualizado a OAC",
        DomainNames:                 []string{"app.miempresa.com"},
        CertificateArn:              "arn:aws:acm:us-east-1:123:certificate/existing",
        AutoConfigureS3BucketPolicy: true,
    }

    distribution := cloudfront.NewDistributionV2(
        stack,
        "UpdatedDistribution",
        props,
    )

    // CDK actualizará la distribución in-place
    // No hay downtime durante la actualización
}
```

**Proceso de actualización**:

1. CDK detecta cambios en la distribución
2. CloudFront actualiza la configuración (2-5 minutos)
3. Los cambios se propagan a edge locations (10-15 minutos)
4. Durante la propagación, algunas edges tienen config nueva, otras antigua
5. Sin downtime para usuarios finales

---

## Limitaciones y Consideraciones

### Limitaciones de CloudFront

| Límite                           | Valor      | Ajustable           |
| -------------------------------- | ---------- | ------------------- |
| **Distribuciones por cuenta**    | 200        | ✅ Sí (via support) |
| **CNAMEs por distribución**      | 100        | ✅ Sí               |
| **Origins por distribución**     | 25         | ✅ Sí               |
| **Cache behaviors**              | 25         | ✅ Sí               |
| **Invalidaciones simultáneas**   | 3          | ❌ No               |
| **Paths por invalidación**       | 3,000      | ❌ No               |
| **Invalidaciones gratuitas/mes** | 1,000      | ❌ No               |
| **Request rate**                 | Sin límite | N/A                 |
| **Tamaño archivo (GET/HEAD)**    | 20 GB\*    | ❌ No               |
| **Tamaño archivo (POST/PUT)**    | 30 GB\*    | ❌ No               |

\*Nota: Para archivos > 20 GB, usar S3 Transfer Acceleration directamente

### Limitaciones de la Estrategia Actual

#### Lo que está Implementado ✅

- Origin Access Control (OAC)
- HTTPS enforcement
- HTTP/2 y HTTP/3
- Security headers básicos
- Error responses para SPAs
- Custom domains con ACM
- WAF integration
- Access logging
- IPv6 support
- Compression automática

#### Lo que Requiere Extensión Futura 🔄

**1. Custom Cache Policies**

```go
// Actualmente: Usa políticas AWS managed
// Futuro: Permitir custom cache policies

type CustomCachePolicyConfig struct {
    Name              string
    MinTTL            int
    MaxTTL            int
    DefaultTTL        int
    QueryStrings      []string
    Headers           []string
    Cookies           []string
}

// Extensión propuesta:
props := CloudFrontPropertiesV2{
    // ...
    CustomCachePolicy: &CustomCachePolicyConfig{
        Name:       "MyCustomPolicy",
        DefaultTTL: 3600,
        // ...
    },
}
```

**2. Multiple Behaviors (Path Patterns)**

```go
// Actualmente: Solo default behavior
// Futuro: Múltiples behaviors para diferentes paths

type BehaviorConfig struct {
    PathPattern   string
    CachePolicy   string
    AllowedMethods []string
}

// Extensión propuesta:
props := CloudFrontPropertiesV2{
    // ...
    AdditionalBehaviors: []BehaviorConfig{
        {
            PathPattern: "/api/*",
            CachePolicy: "CACHING_DISABLED",
        },
        {
            PathPattern: "/static/images/*",
            CachePolicy: "CACHING_OPTIMIZED",
        },
    },
}
```

**3. Origin Shield**

```go
// Actualmente: No implementado
// Futuro: Habilitar Origin Shield

props := CloudFrontPropertiesV2{
    // ...
    EnableOriginShield: true,
    OriginShieldRegion: "us-east-1",
}
```

**4. Lambda@Edge / CloudFront Functions**

```go
// Actualmente: No implementado
// Futuro: Soporte para edge functions

type EdgeFunctionConfig struct {
    FunctionARN string
    EventType   string // viewer-request, origin-request, etc.
}

props := CloudFrontPropertiesV2{
    // ...
    EdgeFunctions: []EdgeFunctionConfig{
        {
            FunctionARN: "arn:aws:lambda:us-east-1:123:function:MyEdgeFunction",
            EventType:   "viewer-request",
        },
    },
}
```

**5. Real-time Logs**

```go
// Actualmente: Solo standard access logs
// Futuro: Real-time logs a Kinesis

props := CloudFrontPropertiesV2{
    // ...
    EnableRealtimeLogs: true,
    KinesisStreamARN:   "arn:aws:kinesis:us-east-1:123:stream/cf-realtime",
}
```

**6. Geo Restriction**

```go
// Actualmente: Solo via WAF
// Futuro: Geo restriction nativa

props := CloudFrontPropertiesV2{
    // ...
    GeoRestrictionType:      "DENY",
    GeoRestrictionCountries: []string{"CN", "RU"},
}
```

### Consideraciones de Arquitectura

#### Cuándo NO usar CloudFront

| Escenario                           | Razón                        | Alternativa                     |
| ----------------------------------- | ---------------------------- | ------------------------------- |
| **Uploads grandes (>30GB)**         | CloudFront limita a 30GB     | S3 Transfer Acceleration        |
| **WebSockets bidireccionales**      | CloudFront es unidireccional | ALB directo con Sticky Sessions |
| **Streaming en vivo (sub-segundo)** | Latencia de cache            | AWS IVS o MediaLive             |
| **Contenido altamente dinámico**    | Cache inefectivo             | API Gateway + Lambda            |
| **Aplicación interna VPC-only**     | No necesita CDN público      | VPC endpoints directos          |

#### Cuándo usar CloudFront

| Escenario                          | Beneficio Principal                     |
| ---------------------------------- | --------------------------------------- |
| **SPAs (React/Vue/Angular)**       | Latencia ultra-baja global              |
| **Sitios web estáticos**           | Costo mínimo, máximo rendimiento        |
| **Assets media (imágenes/videos)** | Reducción masiva de bandwidth costs     |
| **APIs read-heavy**                | Cache reduce load en origen             |
| **PWAs**                           | Offline capabilities con service worker |
| **E-commerce**                     | Performance crítico para conversión     |
| **Global audience**                | 450+ edge locations                     |

---

## Glosario de Términos

### Términos AWS CloudFront

- **Distribution**: La configuración principal de CloudFront que define cómo se distribuye el contenido
- **Edge Location**: Datacenter donde CloudFront cachea contenido (450+ globalmente)
- **Origin**: Servidor fuente del contenido (S3, HTTP server, ALB, etc.)
- **OAC (Origin Access Control)**: Método moderno para dar acceso privado a S3
- **OAI (Origin Access Identity)**: Método legacy para acceso privado a S3 (deprecated)
- **Behavior**: Reglas de routing y cache para paths específicos
- **Cache Key**: Componentes usados para determinar unicidad de cache (URL, headers, cookies, query strings)
- **TTL (Time To Live)**: Duración que un objeto permanece en cache
- **Invalidation**: Proceso para remover objetos del cache antes de expiración
- **Price Class**: Selección de regiones geográficas para distribución

### Términos de Rendimiento

- **TTFB (Time To First Byte)**: Tiempo desde request hasta primer byte de respuesta
- **Cache Hit**: Request servido desde cache de CloudFront
- **Cache Miss**: Request que debe ir al origen
- **Origin Shield**: Capa adicional de cache entre edge locations y origen
- **Compression Ratio**: Porcentaje de reducción de tamaño por compresión
- **Cold Start**: Primera request a un objeto (siempre miss)

### Términos de Seguridad

- **SNI (Server Name Indication)**: Extensión TLS que permite múltiples certificados en misma IP
- **HSTS (HTTP Strict Transport Security)**: Header que fuerza HTTPS en navegador
- **CSP (Content Security Policy)**: Header que define fuentes permitidas de contenido
- **XSS (Cross-Site Scripting)**: Tipo de ataque web que CSP ayuda a prevenir
- **CORS (Cross-Origin Resource Sharing)**: Mecanismo para permitir requests entre dominios

---

## Recursos Adicionales

### Documentación AWS Oficial

- [CloudFront Developer Guide](https://docs.aws.amazon.com/cloudfront/)
- [Origin Access Control (OAC) Documentation](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html)
- [Cache Policies](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/controlling-the-cache-key.html)
- [Security Best Practices](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/security-best-practices.html)

### Herramientas Útiles

```bash
# AWS CLI CloudFront commands
aws cloudfront help

# CloudFront Invalidation
aws cloudfront create-invalidation --distribution-id E1ABC --paths "/*"

# Get distribution config
aws cloudfront get-distribution-config --id E1ABC

# List distributions
aws cloudfront list-distributions

# CloudWatch metrics
aws cloudwatch get-metric-statistics \
    --namespace AWS/CloudFront \
    --metric-name Requests \
    --dimensions Name=DistributionId,Value=E1ABC \
    --start-time 2024-01-15T00:00:00Z \
    --end-time 2024-01-15T23:59:59Z \
    --period 3600 \
    --statistics Sum
```

### Testing Tools

```bash
# Curl con headers detallados
curl -I -H "Accept-Encoding: gzip" https://app.miempresa.com

# Test cache
curl -I https://app.miempresa.com | grep -i "x-cache"
# Hit from cloudfront = Cache hit
# Miss from cloudfront = Cache miss

# Test compression
curl -H "Accept-Encoding: gzip" https://app.miempresa.com -o /dev/null -w "%{size_download}\n"

# Test HTTPS redirect
curl -I http://app.miempresa.com

# Lighthouse CI
npm install -g @lhci/cli
lhci autorun --collect.url=https://app.miempresa.com
```

---

## Conclusión

La **S3CloudFrontStrategy** proporciona una implementación robusta, segura y lista para producción de distribuciones CloudFront con origen en S3. Esta estrategia:

### ✅ Cumple con Mejores Prácticas

- **Seguridad**: OAC moderno, HTTPS enforced, TLS 1.2+, security headers
- **Rendimiento**: HTTP/2+3, compresión, políticas de cache optimizadas
- **Confiabilidad**: IPv6, error handling para SPAs, configuración automática
- **Costos**: Compresión automática, cache inteligente, price class configurable

### 🎯 Casos de Uso Ideales

- Single Page Applications (React, Vue, Angular)
- Sitios web estáticos (Hugo, Jekyll, Next.js export)
- Progressive Web Apps
- Assets CDN (imágenes, videos, documentos)
- Landing pages de marketing
- Portales corporativos

### 🚀 Próximos Pasos

1. **Implementar otras estrategias**: API Gateway, ALB, multi-origin
2. **Extender funcionalidad**: Custom cache policies, behaviors adicionales, Origin Shield
3. **Agregar observabilidad**: Real-time logs, dashboards avanzados
4. **Automatizar testing**: Integration tests, performance benchmarks
5. **Documentar edge cases**: Troubleshooting guides, runbooks operacionales

### 📚 Documentación Relacionada

- [cloudfront_contract.go](./cloudfront-contract.md) - Interface y contratos
- [cloudfront_factory.go](./cloudfront-factory.md) - Factory pattern implementation
- [cloudfront_api.go](./cloudfront-api-strategy.md) - API Strategy (futuro)
- [cloudfront_alb.go](./cloudfront-alb-strategy.md) - ALB Strategy (futuro)

---

**Versión**: 1.0.0  
**Última Actualización**: Enero 2024  
**Mantenedor**: Tu Equipo DevOps  
**Licencia**: MIT
