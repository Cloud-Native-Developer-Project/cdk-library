# Documentación del Paquete AWS CloudFront CDK

## Resumen

Este paquete Go proporciona una implementación integral de AWS CDK para crear y configurar distribuciones de Amazon CloudFront con optimización de rendimiento, seguridad avanzada y configuraciones de caching inteligentes. Sigue las mejores prácticas de AWS y soporta múltiples tipos de orígenes, edge functions y configuraciones de comportamiento.

## Tabla de Contenidos

- [Instalación](#instalación)
- [Inicio Rápido](#inicio-rápido)
- [Opciones de Configuración](#opciones-de-configuración)
- [Tipos de Orígenes](#tipos-de-orígenes)
- [Configuración de Caching](#configuración-de-caching)
- [Seguridad y SSL/TLS](#seguridad-y-ssltls)
- [Edge Functions](#edge-functions)
- [Monitoreo y Logging](#monitoreo-y-logging)
- [Casos de Uso](#casos-de-uso)
- [Referencia de API](#referencia-de-api)
- [Mejores Prácticas](#mejores-prácticas)
- [Ejemplos](#ejemplos)

## Instalación

```bash
go get github.com/aws/aws-cdk-go/awscdk/v2
go get github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront
go get github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins
go get github.com/aws/constructs-go/constructs/v10
```

## Inicio Rápido

### Uso Básico - Sitio Web Estático

```go
package main

import (
    "tu-proyecto/cloudfront"
    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
)

func main() {
    app := awscdk.NewApp(nil)
    stack := awscdk.NewStack(app, jsii.String("MiStack"), nil)

    // Distribución básica con origen S3
    props := cloudfront.CloudFrontProperties{
        Comment:          "Mi sitio web",
        Enabled:          true,
        OriginType:       "S3",
        S3BucketName:     "mi-sitio-web-bucket",
        UseOriginAccessControl: true,
        ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
        CachePolicy:      "MANAGED_CACHING_OPTIMIZED",
        PriceClass:       "100", // Solo US, Canadá y Europa
    }

    distribution := cloudfront.NewDistribution(stack, "MiDistribucion", props)

    app.Synth(nil)
}
```

### Configuración Lista para Producción

```go
props := cloudfront.CloudFrontProperties{
    // Configuración Básica
    Comment:           "Aplicación Web Producción",
    Enabled:           true,
    DefaultRootObject: "index.html",
    DomainNames:       []string{"www.miapp.com", "miapp.com"},
    PriceClass:        "ALL", // Distribución global completa
    HttpVersion:       "HTTP2_AND_3",
    EnableIPv6:        true,

    // Origen S3 con OAC (Recomendado)
    OriginType:             "S3",
    S3BucketName:           "mi-app-origin-bucket",
    UseOriginAccessControl: true,
    OriginShield:           true,
    OriginShieldRegion:     "us-east-1",

    // SSL/TLS Seguro
    CertificateArn:         "arn:aws:acm:us-east-1:123456789012:certificate/abc123",
    MinimumProtocolVersion: "TLS_V1_2_2021",
    SSLSupportMethod:       "SNI_ONLY",

    // Seguridad
    WebAclArn:               "arn:aws:wafv2:us-east-1:123456789012:global/webacl/MyWebACL",
    GeoRestrictionType:      "DENY",
    GeoRestrictionCountries: []string{"CN", "RU"}, // Bloquear países específicos

    // Caching Optimizado
    CachePolicy:           "MANAGED_CACHING_OPTIMIZED",
    OriginRequestPolicy:   "MANAGED_CORS_S3",
    ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",
    ViewerProtocolPolicy:  "REDIRECT_TO_HTTPS",
    CompressResponse:      true,

    // Páginas de Error Personalizadas
    EnableErrorPages: true,
    ErrorPageConfigs: []cloudfront.ErrorPageConfig{
        {
            ErrorCode:        404,
            ResponseCode:     200,
            ResponsePagePath: "/index.html", // SPA routing
        },
        {
            ErrorCode:        403,
            ResponseCode:     200,
            ResponsePagePath: "/index.html",
        },
    },

    // Logging y Monitoreo
    EnableAccessLogging:       true,
    LoggingBucket:             "mi-cloudfront-logs-bucket",
    LoggingPrefix:             "access-logs/",
    EnableAdditionalMetrics:   true,
    MonitoringRealtimeMetrics: true,
}

distribution := cloudfront.NewDistribution(stack, "DistribucionProduccion", props)
```

## Opciones de Configuración

### Estructura CloudFrontProperties

#### Configuración Básica

- **Comment**: Descripción de la distribución
- **Enabled**: Si la distribución está habilitada
- **DefaultRootObject**: Objeto por defecto (ej. "index.html")
- **DomainNames**: Dominios personalizados (CNAMEs)
- **PriceClass**: Clase de precio ("ALL", "100", "200")
- **HttpVersion**: Versión HTTP ("HTTP1_1", "HTTP2", "HTTP2_AND_3")
- **EnableIPv6**: Soporte IPv6

#### Configuración de Origen

- **OriginType**: Tipo de origen ("S3", "S3_WEBSITE", "HTTP", "LOAD_BALANCER")
- **OriginDomainName**: Nombre de dominio del origen
- **OriginPath**: Prefijo de ruta para solicitudes al origen
- **OriginShield**: Habilitar Origin Shield para optimización
- **OriginShieldRegion**: Región de Origin Shield

#### Origen S3 Específico

- **S3BucketName**: Nombre del bucket S3
- **UseOriginAccessControl**: Usar OAC en lugar de OAI (recomendado)

#### Origen HTTP/Personalizado

- **OriginProtocolPolicy**: Política de protocolo ("HTTP_ONLY", "HTTPS_ONLY", "MATCH_VIEWER")
- **OriginPort**: Puerto personalizado
- **OriginSSLProtocols**: Protocolos SSL para orígenes personalizados
- **OriginReadTimeout**: Timeout de lectura (1-180 segundos)
- **OriginKeepaliveTimeout**: Timeout keep-alive (1-60 segundos)

#### Configuración SSL/TLS

- **CertificateArn**: ARN del certificado ACM (debe estar en us-east-1)
- **MinimumProtocolVersion**: Versión mínima TLS
- **SSLSupportMethod**: Método de soporte SSL ("SNI_ONLY", "VIP")

#### Configuración de Seguridad

- **WebAclArn**: ARN del WebACL de AWS WAF
- **GeoRestrictionType**: Restricción geográfica ("ALLOW", "DENY", "NONE")
- **GeoRestrictionCountries**: Lista de códigos de país

#### Configuración de Caching

- **CachePolicy**: Política de cache ("MANAGED_CACHING_OPTIMIZED", "CUSTOM")
- **OriginRequestPolicy**: Política de solicitudes al origen
- **ResponseHeadersPolicy**: Política de headers de respuesta
- **ViewerProtocolPolicy**: Política de protocolo del viewer
- **AllowedMethods**: Métodos HTTP permitidos
- **CachedMethods**: Métodos para cachear respuestas

## Tipos de Orígenes

### 1. Origen S3 con OAC (Recomendado)

```go
props := cloudfront.CloudFrontProperties{
    OriginType:             "S3",
    S3BucketName:           "mi-bucket-origen",
    UseOriginAccessControl: true, // Más seguro que OAI
    OriginPath:             "/static", // Opcional: prefijo de ruta
}
```

**Características:**

- Acceso privado al bucket S3
- No requiere políticas de bucket públicas
- Mejor seguridad que Origin Access Identity (OAI)
- Soporte para todos los métodos HTTP

### 2. Origen S3 Website

```go
props := cloudfront.CloudFrontProperties{
    OriginType:   "S3_WEBSITE",
    S3BucketName: "mi-bucket-website",
    // El bucket debe estar configurado para website hosting
}
```

**Casos de uso:**

- Redirecciones personalizadas
- Páginas de error específicas de S3
- Compatibilidad con aplicaciones legacy

### 3. Origen HTTP/HTTPS Personalizado

```go
props := cloudfront.CloudFrontProperties{
    OriginType:           "HTTP",
    OriginDomainName:     "api.miapp.com",
    OriginProtocolPolicy: "HTTPS_ONLY",
    OriginPort:           443,
    OriginSSLProtocols:   []string{"TLSv1.2"},
    OriginReadTimeout:    30,
    OriginKeepaliveTimeout: 5,
}
```

**Casos de uso:**

- APIs REST
- Servidores web personalizados
- Microservicios

### 4. Application Load Balancer

```go
props := cloudfront.CloudFrontProperties{
    OriginType:           "LOAD_BALANCER",
    OriginDomainName:     "mi-alb-123456789.us-east-1.elb.amazonaws.com",
    OriginProtocolPolicy: "HTTPS_ONLY",
}
```

**Beneficios:**

- Distribución de carga automática
- Health checks integrados
- Escalabilidad automática

## Configuración de Caching

### Políticas de Cache Administradas

#### MANAGED_CACHING_OPTIMIZED (Recomendada)

```go
props.CachePolicy = "MANAGED_CACHING_OPTIMIZED"
```

- TTL optimizado para contenido web
- Cachea basado en query strings selectivos
- Compresión automática habilitada

#### MANAGED_CACHING_DISABLED

```go
props.CachePolicy = "MANAGED_CACHING_DISABLED"
```

- Para contenido dinámico
- APIs que cambian frecuentemente
- Aplicaciones en tiempo real

#### MANAGED_AMPLIFY

```go
props.CachePolicy = "MANAGED_AMPLIFY"
```

- Optimizada para aplicaciones Amplify
- Manejo especial de SPAs
- Cache inteligente de assets

### Política de Cache Personalizada

```go
props := cloudfront.CloudFrontProperties{
    CachePolicy:               "CUSTOM",
    CustomCachePolicyName:     "MiPoliticaPersonalizada",
    CustomCacheTTLDefault:     86400,  // 1 día
    CustomCacheTTLMin:         0,      // Sin mínimo
    CustomCacheTTLMax:         31536000, // 1 año
    CustomCacheQueryStrings:   []string{"version", "lang"},
    CustomCacheHeaders:        []string{"Accept-Language", "CloudFront-Viewer-Country"},
    CustomCacheCookies:        []string{"session-id"},
}
```

**Configuración avanzada:**

- Control granular de TTL
- Headers específicos en cache key
- Query strings selectivos
- Cookies específicos

## Seguridad y SSL/TLS

### Certificados SSL

```go
props := cloudfront.CloudFrontProperties{
    CertificateArn:         "arn:aws:acm:us-east-1:123456789012:certificate/abc123",
    MinimumProtocolVersion: "TLS_V1_2_2021", // Más seguro
    SSLSupportMethod:       "SNI_ONLY",      // Más económico
}
```

**Versiones TLS disponibles:**

- `TLS_V1_2016`: Compatibilidad máxima
- `TLS_V1_1_2016`: Balance seguridad/compatibilidad
- `TLS_V1_2_2019`: Recomendado para la mayoría
- `TLS_V1_2_2021`: Máxima seguridad

### AWS WAF Integration

```go
props := cloudfront.CloudFrontProperties{
    WebAclArn: "arn:aws:wafv2:us-east-1:123456789012:global/webacl/MyWebACL",
}
```

**Protecciones típicas:**

- Rate limiting
- SQL injection protection
- XSS protection
- Bot detection
- Geo-blocking avanzado

### Restricciones Geográficas

```go
// Permitir solo países específicos
props := cloudfront.CloudFrontProperties{
    GeoRestrictionType:      "ALLOW",
    GeoRestrictionCountries: []string{"US", "CA", "MX"},
}

// Bloquear países específicos
props := cloudfront.CloudFrontProperties{
    GeoRestrictionType:      "DENY",
    GeoRestrictionCountries: []string{"CN", "RU", "KP"},
}
```

## Edge Functions

### CloudFront Functions (Recomendadas)

```go
props := cloudfront.CloudFrontProperties{
    EnableEdgeFunctions: true,
    CloudFrontFunctions: []cloudfront.CloudFrontFunctionConfig{
        {
            FunctionName: "url-rewrite",
            EventType:    "VIEWER_REQUEST",
            FunctionCode: `
                function handler(event) {
                    var request = event.request;
                    var uri = request.uri;

                    // Redirect /old-path to /new-path
                    if (uri === '/old-path') {
                        request.uri = '/new-path';
                    }

                    return request;
                }
            `,
        },
    },
}
```

**Casos de uso:**

- URL rewrites simples
- Headers manipulation
- Redirects básicos
- A/B testing simple

### Lambda@Edge

```go
props := cloudfront.CloudFrontProperties{
    EnableEdgeFunctions: true,
    LambdaEdgeFunctions: []cloudfront.LambdaEdgeConfig{
        {
            FunctionArn: "arn:aws:lambda:us-east-1:123456789012:function:MyEdgeFunction:1",
            EventType:   "ORIGIN_REQUEST",
            IncludeBody: true,
        },
    },
}
```

**Casos de uso:**

- Lógica compleja de routing
- Autenticación/autorización
- Manipulación de contenido
- Integración con bases de datos

## Monitoreo y Logging

### Access Logs

```go
props := cloudfront.CloudFrontProperties{
    EnableAccessLogging:   true,
    LoggingBucket:         "mi-cloudfront-logs",
    LoggingPrefix:         "access-logs/año/mes/día/",
    LoggingIncludeCookies: false, // Por privacidad
}
```

**Información capturada:**

- Timestamp de la solicitud
- Edge location que sirvió la solicitud
- Bytes transferidos
- IP del cliente
- Método HTTP y URI
- Status code
- Referrer y User-Agent
- Cache hit/miss status

### CloudWatch Metrics

```go
props := cloudfront.CloudFrontProperties{
    EnableAdditionalMetrics:   true,
    MonitoringRealtimeMetrics: true, // Costo adicional
}
```

**Métricas estándar:**

- Requests
- BytesDownloaded
- BytesUploaded
- 4xxErrorRate
- 5xxErrorRate

**Métricas adicionales:**

- CacheHitRate
- OriginLatency
- ErrorRate por edge location

### Real-time Logs

```go
props := cloudfront.CloudFrontProperties{
    EnableRealtimeLogging: true,
    RealtimeLogArn:        "arn:aws:kinesis:us-east-1:123456789012:stream/cloudfront-realtime-logs",
}
```

**Casos de uso:**

- Monitoreo en tiempo real
- Detección de anomalías
- Analytics en vivo
- Alertas inmediatas

## Casos de Uso

### 1. Sitio Web Estático (SPA)

```go
props := cloudfront.CloudFrontProperties{
    Comment:           "Single Page Application",
    OriginType:        "S3",
    S3BucketName:      "mi-spa-bucket",
    UseOriginAccessControl: true,
    DefaultRootObject: "index.html",

    // Manejo de routing de SPA
    EnableErrorPages: true,
    ErrorPageConfigs: []cloudfront.ErrorPageConfig{
        {
            ErrorCode:        404,
            ResponseCode:     200,
            ResponsePagePath: "/index.html",
        },
        {
            ErrorCode:        403,
            ResponseCode:     200,
            ResponsePagePath: "/index.html",
        },
    },

    CachePolicy:          "MANAGED_CACHING_OPTIMIZED",
    ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
    CompressResponse:     true,
}
```

### 2. API Gateway con Cache

```go
props := cloudfront.CloudFrontProperties{
    Comment:          "API con cache inteligente",
    OriginType:       "HTTP",
    OriginDomainName: "api.miapp.com",

    // Comportamientos múltiples
    AdditionalBehaviors: []cloudfront.BehaviorConfig{
        {
            PathPattern:          "/api/static/*",
            CachePolicy:          "MANAGED_CACHING_OPTIMIZED",
            ViewerProtocolPolicy: "HTTPS_ONLY",
        },
        {
            PathPattern:          "/api/dynamic/*",
            CachePolicy:          "MANAGED_CACHING_DISABLED",
            ViewerProtocolPolicy: "HTTPS_ONLY",
        },
    },
}
```

### 3. Streaming de Media

```go
props := cloudfront.CloudFrontProperties{
    Comment:      "Streaming de video",
    OriginType:   "S3",
    S3BucketName: "mi-media-bucket",

    // Optimizaciones para media
    AllowedMethods:       []string{"GET", "HEAD", "OPTIONS"},
    CachedMethods:        []string{"GET", "HEAD"},
    CompressResponse:     false, // No comprimir video
    SmoothStreaming:      true,  // Microsoft Smooth Streaming

    // Cache largo para contenido multimedia
    CachePolicy: "CUSTOM",
    CustomCacheTTLDefault: 86400 * 7, // 1 semana
    CustomCacheTTLMax:     86400 * 30, // 1 mes
}
```

### 4. Aplicación Global con Múltiples Orígenes

```go
props := cloudfront.CloudFrontProperties{
    Comment:           "Aplicación multi-origen",
    OriginType:        "S3", // Origen por defecto
    S3BucketName:      "mi-app-static",

    AdditionalBehaviors: []cloudfront.BehaviorConfig{
        {
            PathPattern:      "/api/*",
            OriginType:       "LOAD_BALANCER",
            OriginDomainName: "api-alb.us-east-1.elb.amazonaws.com",
            CachePolicy:      "MANAGED_CACHING_DISABLED",
        },
        {
            PathPattern:      "/images/*",
            OriginType:       "S3",
            OriginDomainName: "mi-images-bucket.s3.amazonaws.com",
            CachePolicy:      "MANAGED_CACHING_OPTIMIZED",
        },
        {
            PathPattern:      "/videos/*",
            OriginType:       "S3",
            OriginDomainName: "mi-videos-bucket.s3.amazonaws.com",
            CompressResponse: false,
        },
    },
}
```

## Referencia de API

### Función Principal

#### `NewDistribution(scope constructs.Construct, id string, props CloudFrontProperties) awscloudfront.Distribution`

Crea una nueva distribución CloudFront con la configuración especificada.

### Estructuras de Configuración

#### `CloudFrontProperties`

Estructura principal con todas las opciones de configuración.

#### `ErrorPageConfig`

Configuración de páginas de error personalizadas.

#### `CloudFrontFunctionConfig`

Configuración de CloudFront Functions.

#### `LambdaEdgeConfig`

Configuración de Lambda@Edge.

#### `BehaviorConfig`

Configuración de comportamientos adicionales para rutas específicas.

### Funciones de Configuración

#### `configurePriceClass(priceClass string) awscloudfront.PriceClass`

Convierte string a enum de clase de precio.

#### `configureHttpVersion(httpVersion string) awscloudfront.HttpVersion`

Configura la versión HTTP soportada.

#### `configureViewerProtocolPolicy(policy string) awscloudfront.ViewerProtocolPolicy`

Configura la política de protocolo del viewer.

## Mejores Prácticas

### Seguridad

1. **Siempre usar HTTPS**

   ```go
   props.ViewerProtocolPolicy = "REDIRECT_TO_HTTPS"
   props.MinimumProtocolVersion = "TLS_V1_2_2021"
   ```

2. **Origin Access Control para S3**

   ```go
   props.UseOriginAccessControl = true // En lugar de OAI
   ```

3. **WAF para protección**
   ```go
   props.WebAclArn = "arn:aws:wafv2:us-east-1:123456789012:global/webacl/MyWebACL"
   ```

### Rendimiento

1. **Origin Shield para orígenes costosos**

   ```go
   props.OriginShield = true
   props.OriginShieldRegion = "us-east-1" // Cerca del origen
   ```

2. **Compresión automática**

   ```go
   props.CompressResponse = true
   ```

3. **HTTP/2 y HTTP/3**
   ```go
   props.HttpVersion = "HTTP2_AND_3"
   ```

### Costos

1. **Clase de precio apropiada**

   ```go
   // Para audiencia regional
   props.PriceClass = "100" // US, Canadá, Europa

   // Para audiencia global
   props.PriceClass = "ALL"
   ```

2. **Cache policies optimizadas**
   ```go
   props.CachePolicy = "MANAGED_CACHING_OPTIMIZED"
   ```

### Monitoreo

1. **Logs estructurados**

   ```go
   props.EnableAccessLogging = true
   props.LoggingPrefix = "cloudfront-logs/año/mes/día/"
   ```

2. **Métricas adicionales para producción**
   ```go
   props.EnableAdditionalMetrics = true
   ```

## Ejemplos Avanzados

### Configuración Multi-Ambiente

```go
func CreateDistributionForEnvironment(env string) cloudfront.CloudFrontProperties {
    base := cloudfront.CloudFrontProperties{
        Comment:           fmt.Sprintf("App %s environment", env),
        Enabled:           true,
        OriginType:        "S3",
        UseOriginAccessControl: true,
        ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
        CompressResponse:     true,
    }

    switch env {
    case "prod":
        base.S3BucketName = "mi-app-prod-bucket"
        base.PriceClass = "ALL"
        base.EnableAdditionalMetrics = true
        base.MonitoringRealtimeMetrics = true
        base.CachePolicy = "MANAGED_CACHING_OPTIMIZED"

    case "staging":
        base.S3BucketName = "mi-app-staging-bucket"
        base.PriceClass = "100"
        base.CachePolicy = "MANAGED_CACHING_OPTIMIZED"

    case "dev":
        base.S3BucketName = "mi-app-dev-bucket"
        base.PriceClass = "100"
        base.CachePolicy = "MANAGED_CACHING_DISABLED" // Para testing
    }

    return base
}
```

### Pipeline de Deployment con Invalidación

```go
// En tu pipeline de CI/CD, después del deployment
func InvalidateDistribution(distributionId string, paths []string) {
    // Usar AWS SDK para crear invalidación
    // Ejemplo conceptual - implementar según tu pipeline

    invalidationPaths := []string{
        "/index.html",
        "/static/js/*",
        "/static/css/*",
    }

    // CreateInvalidation API call
}
```

---

## Conclusión

Esta implementación de CloudFront CDK proporciona una solución completa para distribución de contenido global con configuraciones optimizadas para diferentes casos de uso. La estructura modular permite desde configuraciones simples hasta setups empresariales complejos con múltiples orígenes, edge functions y políticas de cache personalizadas.

Las mejores prácticas integradas aseguran rendimiento óptimo, seguridad robusta y costos controlados, mientras que la flexibilidad de configuración permite adaptarse a requisitos específicos de cada aplicación.

```go
package cloudfront

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ========================================================================
// ENHANCED CLOUDFRONT PROPERTIES
// ========================================================================

type CloudFrontProperties struct {
	// Basic Configuration
	Comment           string
	Enabled           bool
	DefaultRootObject string
	DomainNames       []string
	PriceClass        string
	HttpVersion       string
	EnableIPv6        bool

	// Origin Configuration
	OriginType         string
	OriginDomainName   string
	OriginPath         string
	OriginShield       bool
	OriginShieldRegion string

	// S3 Origin Specific - ENHANCED: Now prioritizes direct bucket reference
	S3Bucket     awss3.IBucket // PRIMARY: Direct bucket reference (RECOMMENDED)
	S3BucketName string        // FALLBACK: Bucket name for existing buckets

	// HTTP/Custom Origin Specific
	OriginProtocolPolicy   string
	OriginPort             int32
	OriginSSLProtocols     []string
	OriginReadTimeout      int32
	OriginKeepaliveTimeout int32

	// SSL/TLS Configuration
	CertificateArn         string
	MinimumProtocolVersion string
	SSLSupportMethod       string

	// Security Configuration
	WebAclArn               string
	GeoRestrictionType      string
	GeoRestrictionCountries []string

	// Caching Configuration
	CachePolicy           string
	OriginRequestPolicy   string
	ResponseHeadersPolicy string
	CompressResponse      bool
	ViewerProtocolPolicy  string
	AllowedMethods        []string
	CachedMethods         []string

	// Custom Cache Policy
	CustomCachePolicyName   string
	CustomCacheTTLDefault   int32
	CustomCacheTTLMin       int32
	CustomCacheTTLMax       int32
	CustomCacheQueryStrings []string
	CustomCacheHeaders      []string
	CustomCacheCookies      []string

	// Error Pages Configuration
	EnableErrorPages bool
	ErrorPageConfigs []ErrorPageConfig

	// Logging Configuration
	EnableAccessLogging   bool
	LoggingBucket         string
	LoggingPrefix         string
	LoggingIncludeCookies bool
	EnableRealtimeLogging bool
	RealtimeLogArn        string

	// Monitoring Configuration
	EnableAdditionalMetrics   bool
	MonitoringRealtimeMetrics bool

	// Edge Functions Configuration
	EnableEdgeFunctions bool
	CloudFrontFunctions []CloudFrontFunctionConfig
	LambdaEdgeFunctions []LambdaEdgeConfig

	// Additional Behaviors
	AdditionalBehaviors []BehaviorConfig

	// Performance Optimization
	EnableGRPC      bool
	SmoothStreaming bool

	// Trusted Signers
	TrustedSigners   []string
	TrustedKeyGroups []string

	// NEW: Auto-configure S3 bucket policy for OAC
	AutoConfigureS3BucketPolicy bool // Default: true for S3 origins
}

type ErrorPageConfig struct {
	ErrorCode          int32
	ResponseCode       int32
	ResponsePagePath   string
	ErrorCachingMinTTL int32
}

type CloudFrontFunctionConfig struct {
	FunctionName string
	EventType    string
	FunctionCode string
}

type LambdaEdgeConfig struct {
	FunctionArn string
	EventType   string
	IncludeBody bool
}

type BehaviorConfig struct {
	PathPattern           string
	UseDefaultOrigin      bool
	OriginType            string
	OriginDomainName      string
	CachePolicy           string
	OriginRequestPolicy   string
	ResponseHeadersPolicy string
	ViewerProtocolPolicy  string
	AllowedMethods        []string
	CachedMethods         []string
	CompressResponse      bool
	TrustedSigners        []string
	TrustedKeyGroups      []string
}

// ========================================================================
// ENHANCED: NewDistribution with automatic S3 bucket policy configuration
// ========================================================================

func NewDistribution(scope constructs.Construct, id string, props CloudFrontProperties) awscloudfront.Distribution {
	// Set default for auto-configuration if not explicitly set
	if props.OriginType == "S3" {
		// Default to true if not explicitly set
		props.AutoConfigureS3BucketPolicy = true
	}

	distributionProps := &awscloudfront.DistributionProps{
		Comment:           jsii.String(props.Comment),
		Enabled:           jsii.Bool(props.Enabled),
		DefaultRootObject: configureDefaultRootObject(props.DefaultRootObject),
		DomainNames:       convertToStringPointers(props.DomainNames),
		PriceClass:        configurePriceClass(props.PriceClass),
		HttpVersion:       configureHttpVersion(props.HttpVersion),
		EnableIpv6:        jsii.Bool(props.EnableIPv6),
	}

	// Configure default behavior and capture origin
	defaultBehavior, defaultOrigin := configureDefaultBehavior(scope, props)
	distributionProps.DefaultBehavior = defaultBehavior

	configureSSLSettings(scope, distributionProps, props)
	configureSecurity(distributionProps, props)
	configureErrorPages(distributionProps, props)
	configureLogging(scope, distributionProps, props)
	configureMonitoring(distributionProps, props)
	configureAdditionalBehaviors(scope, distributionProps, props, defaultOrigin)

	// Create distribution
	distribution := awscloudfront.NewDistribution(scope, jsii.String(id), distributionProps)

	// ========================================================================
	// CRITICAL: Configure S3 bucket policy for OAC AFTER distribution creation
	// ========================================================================
	if props.OriginType == "S3" && props.S3Bucket != nil {
		configureS3BucketPolicyForOAC(props.S3Bucket, distribution)
	}

	return distribution
}

// ========================================================================
// NEW: Configure S3 Bucket Policy for OAC
// ========================================================================

func configureS3BucketPolicyForOAC(bucket awss3.IBucket, distribution awscloudfront.Distribution) {
	// Add bucket policy to allow CloudFront OAC access
	bucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:GetObject"),
		Resources: jsii.Strings(*bucket.BucketArn() + "/*"),
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewServicePrincipal(jsii.String("cloudfront.amazonaws.com"), nil),
		},
		Conditions: &map[string]interface{}{
			"StringEquals": map[string]interface{}{
				"AWS:SourceArn": distribution.DistributionArn(),
			},
		},
	}))
}

// ========================================================================
// DEFAULT PROPERTY FACTORIES - ENHANCED
// ========================================================================

// DefaultS3PrivateOACProps returns defaults for serving private S3 content via OAC
// USAGE: Pass your S3 bucket instance to props.S3Bucket
func DefaultS3PrivateOACProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:     "S3 Private Content with OAC",
		Enabled:     true,
		PriceClass:  "ALL",
		HttpVersion: "HTTP2_AND_3",
		EnableIPv6:  true,

		OriginType:                  "S3",
		S3Bucket:                    nil,  // SET THIS: Pass your bucket instance
		AutoConfigureS3BucketPolicy: true, // Automatically configures bucket policy
		DefaultRootObject:           "index.html",
		ViewerProtocolPolicy:        "REDIRECT_TO_HTTPS",
		AllowedMethods:              []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:               []string{"GET", "HEAD"},
		CompressResponse:            true,
		CachePolicy:                 "MANAGED_CACHING_OPTIMIZED",
		OriginRequestPolicy:         "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy:       "MANAGED_SECURITY_HEADERS",
		EnableErrorPages:            true,

		ErrorPageConfigs: []ErrorPageConfig{
			{ErrorCode: 403, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
			{ErrorCode: 404, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
		},
	}
}

// DefaultS3StaticWebsiteProps returns defaults for S3 static website
func DefaultS3StaticWebsiteProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:           "S3 Static Website Distribution",
		Enabled:           true,
		DefaultRootObject: "index.html",
		PriceClass:        "ALL",
		HttpVersion:       "HTTP2_AND_3",
		EnableIPv6:        true,

		OriginType:                  "S3_WEBSITE",
		S3Bucket:                    nil,   // SET THIS
		AutoConfigureS3BucketPolicy: false, // S3 website endpoints don't use OAC

		ViewerProtocolPolicy:  "REDIRECT_TO_HTTPS",
		AllowedMethods:        []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:         []string{"GET", "HEAD"},
		CompressResponse:      true,
		CachePolicy:           "MANAGED_CACHING_OPTIMIZED",
		OriginRequestPolicy:   "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",

		EnableErrorPages: true,
		ErrorPageConfigs: []ErrorPageConfig{
			{ErrorCode: 403, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
			{ErrorCode: 404, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
		},
	}
}

// DefaultSpaProps returns defaults for Single Page Applications
func DefaultSpaProps() CloudFrontProperties {
	props := DefaultS3PrivateOACProps()
	props.Comment = "SPA Distribution with OAC"
	return props
}

// ========================================================================
// EXISTING HELPER FUNCTIONS (keeping all previous implementations)
// ========================================================================

func configureDefaultBehavior(scope constructs.Construct, props CloudFrontProperties) (*awscloudfront.BehaviorOptions, awscloudfront.IOrigin) {
	origin := createOrigin(scope, "DefaultOrigin", props)

	behaviorOptions := &awscloudfront.BehaviorOptions{
		Origin:                origin,
		ViewerProtocolPolicy:  configureViewerProtocolPolicy(props.ViewerProtocolPolicy),
		AllowedMethods:        configureAllowedMethods(props.AllowedMethods),
		CachedMethods:         configureCachedMethods(props.CachedMethods),
		Compress:              jsii.Bool(props.CompressResponse),
		CachePolicy:           configureCachePolicy(scope, props),
		OriginRequestPolicy:   configureOriginRequestPolicy(props.OriginRequestPolicy),
		ResponseHeadersPolicy: configureResponseHeadersPolicy(props.ResponseHeadersPolicy),
	}

	if props.EnableGRPC {
		behaviorOptions.EnableGrpc = jsii.Bool(true)
	}

	if props.SmoothStreaming {
		behaviorOptions.SmoothStreaming = jsii.Bool(true)
	}

	configureEdgeFunctions(scope, behaviorOptions, props)
	return behaviorOptions, origin
}

func createOrigin(scope constructs.Construct, idPrefix string, props CloudFrontProperties) awscloudfront.IOrigin {
	switch props.OriginType {
	case "S3":
		return createS3Origin(scope, idPrefix, props)
	case "S3_WEBSITE":
		return createS3WebsiteOrigin(scope, idPrefix, props)
	case "HTTP", "HTTPS":
		return createHttpOrigin(props)
	case "LOAD_BALANCER":
		return createLoadBalancerOrigin(props)
	default:
		return createHttpOrigin(props)
	}
}

func createS3Origin(scope constructs.Construct, idPrefix string, props CloudFrontProperties) awscloudfront.IOrigin {
	var bucket awss3.IBucket

	// Priority 1: Direct bucket reference (RECOMMENDED)
	if props.S3Bucket != nil {
		bucket = props.S3Bucket
	} else if props.S3BucketName != "" {
		// Priority 2: Import by name (for existing buckets)
		bucket = awss3.Bucket_FromBucketName(
			scope,
			jsii.String(idPrefix+"BucketRef"),
			jsii.String(props.S3BucketName),
		)
	} else {
		panic("Either S3Bucket or S3BucketName must be provided for S3 origin")
	}

	s3OriginProps := &awscloudfrontorigins.S3BucketOriginWithOACProps{
		OriginPath: jsii.String(props.OriginPath),
	}

	if props.OriginShield && props.OriginShieldRegion != "" {
		s3OriginProps.OriginShieldEnabled = jsii.Bool(true)
		s3OriginProps.OriginShieldRegion = jsii.String(props.OriginShieldRegion)
	}

	return awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, s3OriginProps)
}

func createS3WebsiteOrigin(scope constructs.Construct, idPrefix string, props CloudFrontProperties) awscloudfront.IOrigin {
	var bucket awss3.IBucket

	if props.S3Bucket != nil {
		bucket = props.S3Bucket
	} else if props.S3BucketName != "" {
		bucket = awss3.Bucket_FromBucketName(
			scope,
			jsii.String(idPrefix+"WebsiteBucketRef"),
			jsii.String(props.S3BucketName),
		)
	} else {
		panic("Either S3Bucket or S3BucketName must be provided for S3 website origin")
	}

	s3WebsiteOriginProps := &awscloudfrontorigins.S3StaticWebsiteOriginProps{
		OriginPath: jsii.String(props.OriginPath),
	}

	if props.OriginShield && props.OriginShieldRegion != "" {
		s3WebsiteOriginProps.OriginShieldEnabled = jsii.Bool(true)
		s3WebsiteOriginProps.OriginShieldRegion = jsii.String(props.OriginShieldRegion)
	}

	return awscloudfrontorigins.NewS3StaticWebsiteOrigin(bucket, s3WebsiteOriginProps)
}

func createHttpOrigin(props CloudFrontProperties) awscloudfront.IOrigin {
	httpOriginProps := &awscloudfrontorigins.HttpOriginProps{
		OriginPath:       jsii.String(props.OriginPath),
		ProtocolPolicy:   configureOriginProtocolPolicy(props.OriginProtocolPolicy),
		HttpPort:         jsii.Number(getPortOrDefault(props.OriginPort, 80)),
		HttpsPort:        jsii.Number(getPortOrDefault(props.OriginPort, 443)),
		ReadTimeout:      awscdk.Duration_Seconds(jsii.Number(getTimeoutOrDefault(props.OriginReadTimeout, 30))),
		KeepaliveTimeout: awscdk.Duration_Seconds(jsii.Number(getTimeoutOrDefault(props.OriginKeepaliveTimeout, 5))),
	}

	if len(props.OriginSSLProtocols) > 0 {
		httpOriginProps.OriginSslProtocols = convertToOriginSslProtocols(props.OriginSSLProtocols)
	}

	if props.OriginShield && props.OriginShieldRegion != "" {
		httpOriginProps.OriginShieldEnabled = jsii.Bool(true)
		httpOriginProps.OriginShieldRegion = jsii.String(props.OriginShieldRegion)
	}

	return awscloudfrontorigins.NewHttpOrigin(jsii.String(props.OriginDomainName), httpOriginProps)
}

func createLoadBalancerOrigin(props CloudFrontProperties) awscloudfront.IOrigin {
	return createHttpOrigin(props)
}

// [Continue with all other helper functions from your original code...]
// [Including: configureSSLSettings, configureSecurity, configureErrorPages, etc.]

func configureSSLSettings(scope constructs.Construct, distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.CertificateArn != "" {
		certificate := awscertificatemanager.Certificate_FromCertificateArn(
			scope, jsii.String("Certificate"), jsii.String(props.CertificateArn))
		distributionProps.Certificate = certificate
		distributionProps.MinimumProtocolVersion = configureMinimumProtocolVersion(props.MinimumProtocolVersion)
		distributionProps.SslSupportMethod = configureSSLSupportMethod(props.SSLSupportMethod)
	}
}

func configureSecurity(distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.WebAclArn != "" {
		distributionProps.WebAclId = jsii.String(props.WebAclArn)
	}

	if props.GeoRestrictionType != "NONE" && len(props.GeoRestrictionCountries) > 0 {
		distributionProps.GeoRestriction = configureGeoRestriction(
			props.GeoRestrictionType, props.GeoRestrictionCountries)
	}
}

func configureErrorPages(distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.EnableErrorPages && len(props.ErrorPageConfigs) > 0 {
		errorResponses := make([]*awscloudfront.ErrorResponse, 0, len(props.ErrorPageConfigs))

		for _, config := range props.ErrorPageConfigs {
			errorResponse := &awscloudfront.ErrorResponse{
				HttpStatus:         jsii.Number(config.ErrorCode),
				ResponseHttpStatus: jsii.Number(config.ResponseCode),
				ResponsePagePath:   jsii.String(config.ResponsePagePath),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(config.ErrorCachingMinTTL)),
			}
			errorResponses = append(errorResponses, errorResponse)
		}

		distributionProps.ErrorResponses = &errorResponses
	}
}

func configureLogging(scope constructs.Construct, distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.EnableAccessLogging {
		distributionProps.EnableLogging = jsii.Bool(true)

		if props.LoggingBucket != "" {
			logBucket := awss3.Bucket_FromBucketName(scope, jsii.String("LogBucket"), jsii.String(props.LoggingBucket))
			distributionProps.LogBucket = logBucket
		}

		if props.LoggingPrefix != "" {
			distributionProps.LogFilePrefix = jsii.String(props.LoggingPrefix)
		}

		distributionProps.LogIncludesCookies = jsii.Bool(props.LoggingIncludeCookies)
	}
}

func configureMonitoring(distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.EnableAdditionalMetrics {
		distributionProps.PublishAdditionalMetrics = jsii.Bool(true)
	}
}

func configureAdditionalBehaviors(scope constructs.Construct, distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties, defaultOrigin awscloudfront.IOrigin) {
	if len(props.AdditionalBehaviors) > 0 {
		additionalBehaviors := make(map[string]*awscloudfront.BehaviorOptions)

		for _, behaviorConfig := range props.AdditionalBehaviors {
			behaviorOptions := createBehaviorFromConfig(scope, behaviorConfig, props, defaultOrigin)
			additionalBehaviors[behaviorConfig.PathPattern] = behaviorOptions
		}

		distributionProps.AdditionalBehaviors = &additionalBehaviors
	}
}

func createBehaviorFromConfig(scope constructs.Construct, config BehaviorConfig, defaultProps CloudFrontProperties, defaultOrigin awscloudfront.IOrigin) *awscloudfront.BehaviorOptions {
	var origin awscloudfront.IOrigin

	if config.UseDefaultOrigin {
		origin = defaultOrigin
	} else {
		tempProps := defaultProps
		if config.OriginType != "" {
			tempProps.OriginType = config.OriginType
		}
		if config.OriginDomainName != "" {
			tempProps.OriginDomainName = config.OriginDomainName
		}

		behaviorOriginId := "BehaviorOrigin-" + sanitizeID(config.PathPattern)
		origin = createOrigin(scope, behaviorOriginId, tempProps)
	}

	behaviorOptions := &awscloudfront.BehaviorOptions{
		Origin:               origin,
		ViewerProtocolPolicy: configureViewerProtocolPolicy(getStringOrDefault(config.ViewerProtocolPolicy, defaultProps.ViewerProtocolPolicy)),
		AllowedMethods:       configureAllowedMethods(getStringSliceOrDefault(config.AllowedMethods, defaultProps.AllowedMethods)),
		CachedMethods:        configureCachedMethods(getStringSliceOrDefault(config.CachedMethods, defaultProps.CachedMethods)),
		Compress:             jsii.Bool(config.CompressResponse),
	}

	if config.CachePolicy != "" {
		tempProps := defaultProps
		tempProps.CachePolicy = config.CachePolicy
		behaviorOptions.CachePolicy = configureCachePolicy(scope, tempProps)
	}

	if config.OriginRequestPolicy != "" {
		behaviorOptions.OriginRequestPolicy = configureOriginRequestPolicy(config.OriginRequestPolicy)
	}

	if config.ResponseHeadersPolicy != "" {
		behaviorOptions.ResponseHeadersPolicy = configureResponseHeadersPolicy(config.ResponseHeadersPolicy)
	}

	return behaviorOptions
}

func configureEdgeFunctions(scope constructs.Construct, behaviorOptions *awscloudfront.BehaviorOptions, props CloudFrontProperties) {
	if !props.EnableEdgeFunctions {
		return
	}

	if len(props.CloudFrontFunctions) > 0 {
		functionAssociations := make([]*awscloudfront.FunctionAssociation, 0, len(props.CloudFrontFunctions))
		for _, funcConfig := range props.CloudFrontFunctions {
			functionAssociation := &awscloudfront.FunctionAssociation{
				EventType: configureFunctionEventType(funcConfig.EventType),
			}
			functionAssociations = append(functionAssociations, functionAssociation)
		}
		behaviorOptions.FunctionAssociations = &functionAssociations
	}

	if len(props.LambdaEdgeFunctions) > 0 {
		edgeLambdas := make([]*awscloudfront.EdgeLambda, 0, len(props.LambdaEdgeFunctions))
		for _, lambdaConfig := range props.LambdaEdgeFunctions {
			functionVersion := awslambda.Version_FromVersionArn(scope, jsii.String("LambdaEdgeVersion"), jsii.String(lambdaConfig.FunctionArn))
			edgeLambda := &awscloudfront.EdgeLambda{
				EventType:       configureLambdaEventType(lambdaConfig.EventType),
				FunctionVersion: functionVersion,
				IncludeBody:     jsii.Bool(lambdaConfig.IncludeBody),
			}
			edgeLambdas = append(edgeLambdas, edgeLambda)
		}
		behaviorOptions.EdgeLambdas = &edgeLambdas
	}
}

// ========================================================================
// CONFIGURATION HELPER FUNCTIONS
// ========================================================================

func configureDefaultRootObject(rootObject string) *string {
	if rootObject == "" {
		return nil
	}
	return jsii.String(rootObject)
}

func configurePriceClass(priceClass string) awscloudfront.PriceClass {
	switch priceClass {
	case "ALL":
		return awscloudfront.PriceClass_PRICE_CLASS_ALL
	case "200":
		return awscloudfront.PriceClass_PRICE_CLASS_200
	case "100":
		return awscloudfront.PriceClass_PRICE_CLASS_100
	default:
		return awscloudfront.PriceClass_PRICE_CLASS_ALL
	}
}

func configureHttpVersion(httpVersion string) awscloudfront.HttpVersion {
	switch httpVersion {
	case "HTTP1_1":
		return awscloudfront.HttpVersion_HTTP1_1
	case "HTTP2":
		return awscloudfront.HttpVersion_HTTP2
	case "HTTP2_AND_3":
		return awscloudfront.HttpVersion_HTTP2_AND_3
	default:
		return awscloudfront.HttpVersion_HTTP2
	}
}

func configureViewerProtocolPolicy(policy string) awscloudfront.ViewerProtocolPolicy {
	switch policy {
	case "ALLOW_ALL":
		return awscloudfront.ViewerProtocolPolicy_ALLOW_ALL
	case "REDIRECT_TO_HTTPS":
		return awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS
	case "HTTPS_ONLY":
		return awscloudfront.ViewerProtocolPolicy_HTTPS_ONLY
	default:
		return awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS
	}
}

func configureAllowedMethods(methods []string) awscloudfront.AllowedMethods {
	if len(methods) == 0 {
		return awscloudfront.AllowedMethods_ALLOW_GET_HEAD()
	}

	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] &&
		methodSet["PUT"] && methodSet["POST"] && methodSet["PATCH"] && methodSet["DELETE"] {
		return awscloudfront.AllowedMethods_ALLOW_ALL()
	}

	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] {
		return awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS()
	}

	return awscloudfront.AllowedMethods_ALLOW_GET_HEAD()
}

func configureCachedMethods(methods []string) awscloudfront.CachedMethods {
	if len(methods) == 0 {
		return awscloudfront.CachedMethods_CACHE_GET_HEAD()
	}

	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] {
		return awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS()
	}

	return awscloudfront.CachedMethods_CACHE_GET_HEAD()
}

func configureCachePolicy(scope constructs.Construct, props CloudFrontProperties) awscloudfront.ICachePolicy {
	switch props.CachePolicy {
	case "MANAGED_CACHING_OPTIMIZED":
		return awscloudfront.CachePolicy_CACHING_OPTIMIZED()
	case "MANAGED_CACHING_DISABLED":
		return awscloudfront.CachePolicy_CACHING_DISABLED()
	case "MANAGED_AMPLIFY":
		return awscloudfront.CachePolicy_AMPLIFY()
	case "CUSTOM":
		return createCustomCachePolicy(scope, props)
	default:
		return awscloudfront.CachePolicy_CACHING_OPTIMIZED()
	}
}

func createCustomCachePolicy(scope constructs.Construct, props CloudFrontProperties) awscloudfront.ICachePolicy {
	cachePolicyProps := &awscloudfront.CachePolicyProps{
		CachePolicyName: jsii.String(getStringOrDefault(props.CustomCachePolicyName, "CustomCachePolicy")),
		DefaultTtl:      awscdk.Duration_Seconds(jsii.Number(getInt32OrDefault(props.CustomCacheTTLDefault, 86400))),
		MinTtl:          awscdk.Duration_Seconds(jsii.Number(getInt32OrDefault(props.CustomCacheTTLMin, 0))),
		MaxTtl:          awscdk.Duration_Seconds(jsii.Number(getInt32OrDefault(props.CustomCacheTTLMax, 31536000))),
		Comment:         jsii.String("Custom cache policy"),
	}

	if len(props.CustomCacheQueryStrings) > 0 {
		cachePolicyProps.QueryStringBehavior = awscloudfront.CacheQueryStringBehavior_AllowList(
			*jsii.Strings(props.CustomCacheQueryStrings...)...)
	} else {
		cachePolicyProps.QueryStringBehavior = awscloudfront.CacheQueryStringBehavior_None()
	}

	if len(props.CustomCacheHeaders) > 0 {
		cachePolicyProps.HeaderBehavior = awscloudfront.CacheHeaderBehavior_AllowList(
			*jsii.Strings(props.CustomCacheHeaders...)...)
	} else {
		cachePolicyProps.HeaderBehavior = awscloudfront.CacheHeaderBehavior_None()
	}

	if len(props.CustomCacheCookies) > 0 {
		cachePolicyProps.CookieBehavior = awscloudfront.CacheCookieBehavior_AllowList(
			*jsii.Strings(props.CustomCacheCookies...)...)
	} else {
		cachePolicyProps.CookieBehavior = awscloudfront.CacheCookieBehavior_None()
	}

	cachePolicyProps.EnableAcceptEncodingGzip = jsii.Bool(true)
	cachePolicyProps.EnableAcceptEncodingBrotli = jsii.Bool(true)

	return awscloudfront.NewCachePolicy(scope, jsii.String("CustomCachePolicy"), cachePolicyProps)
}

func configureOriginRequestPolicy(policy string) awscloudfront.IOriginRequestPolicy {
	switch policy {
	case "MANAGED_ALL_VIEWER":
		return awscloudfront.OriginRequestPolicy_ALL_VIEWER()
	case "MANAGED_CORS_S3":
		return awscloudfront.OriginRequestPolicy_CORS_S3_ORIGIN()
	case "MANAGED_ELEMENT_CAPTURE":
		return awscloudfront.OriginRequestPolicy_ELEMENTAL_MEDIA_TAILOR()
	case "CUSTOM":
		return awscloudfront.OriginRequestPolicy_ALL_VIEWER()
	default:
		return nil
	}
}

func configureResponseHeadersPolicy(policy string) awscloudfront.IResponseHeadersPolicy {
	switch policy {
	case "MANAGED_CORS_ALLOW_ALL":
		return awscloudfront.ResponseHeadersPolicy_CORS_ALLOW_ALL_ORIGINS()
	case "MANAGED_SECURITY_HEADERS":
		return awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS()
	case "CUSTOM":
		return awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS()
	default:
		return nil
	}
}

func configureOriginProtocolPolicy(policy string) awscloudfront.OriginProtocolPolicy {
	switch policy {
	case "HTTP_ONLY":
		return awscloudfront.OriginProtocolPolicy_HTTP_ONLY
	case "HTTPS_ONLY":
		return awscloudfront.OriginProtocolPolicy_HTTPS_ONLY
	case "MATCH_VIEWER":
		return awscloudfront.OriginProtocolPolicy_MATCH_VIEWER
	default:
		return awscloudfront.OriginProtocolPolicy_HTTPS_ONLY
	}
}

func configureMinimumProtocolVersion(version string) awscloudfront.SecurityPolicyProtocol {
	switch version {
	case "TLS_V1_2016":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2016
	case "TLS_V1_1_2016":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_1_2016
	case "TLS_V1_2_2019":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2019
	case "TLS_V1_2_2021":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
	default:
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
	}
}

func configureSSLSupportMethod(method string) awscloudfront.SSLMethod {
	switch method {
	case "SNI_ONLY":
		return awscloudfront.SSLMethod_SNI
	case "VIP":
		return awscloudfront.SSLMethod_VIP
	default:
		return awscloudfront.SSLMethod_SNI
	}
}

func configureGeoRestriction(restrictionType string, countries []string) awscloudfront.GeoRestriction {
	converted := jsii.Strings(countries...)

	switch restrictionType {
	case "ALLOW":
		return awscloudfront.GeoRestriction_Allowlist((*converted)...)
	case "DENY":
		return awscloudfront.GeoRestriction_Denylist((*converted)...)
	default:
		return awscloudfront.GeoRestriction_Allowlist((*converted)...)
	}
}

func configureFunctionEventType(eventType string) awscloudfront.FunctionEventType {
	switch eventType {
	case "VIEWER_REQUEST":
		return awscloudfront.FunctionEventType_VIEWER_REQUEST
	case "VIEWER_RESPONSE":
		return awscloudfront.FunctionEventType_VIEWER_RESPONSE
	default:
		return awscloudfront.FunctionEventType_VIEWER_REQUEST
	}
}

func configureLambdaEventType(eventType string) awscloudfront.LambdaEdgeEventType {
	switch eventType {
	case "ORIGIN_REQUEST":
		return awscloudfront.LambdaEdgeEventType_ORIGIN_REQUEST
	case "ORIGIN_RESPONSE":
		return awscloudfront.LambdaEdgeEventType_ORIGIN_RESPONSE
	case "VIEWER_REQUEST":
		return awscloudfront.LambdaEdgeEventType_VIEWER_REQUEST
	case "VIEWER_RESPONSE":
		return awscloudfront.LambdaEdgeEventType_VIEWER_RESPONSE
	default:
		return awscloudfront.LambdaEdgeEventType_VIEWER_REQUEST
	}
}

func convertToOriginSslProtocols(protocols []string) *[]awscloudfront.OriginSslPolicy {
	sslPolicies := make([]awscloudfront.OriginSslPolicy, 0, len(protocols))

	for _, protocol := range protocols {
		switch protocol {
		case "SSLv3":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_SSL_V3)
		case "TLSv1":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_TLS_V1)
		case "TLSv1.1":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_TLS_V1_1)
		case "TLSv1.2":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_TLS_V1_2)
		}
	}

	return &sslPolicies
}

// ========================================================================
// UTILITY HELPER FUNCTIONS
// ========================================================================

func convertToStringPointers(strings []string) *[]*string {
	if len(strings) == 0 {
		return nil
	}
	pointers := make([]*string, len(strings))
	for i, s := range strings {
		pointers[i] = jsii.String(s)
	}
	return &pointers
}

func getPortOrDefault(port int32, defaultPort int32) int32 {
	if port == 0 {
		return defaultPort
	}
	return port
}

func getTimeoutOrDefault(timeout int32, defaultTimeout int32) int32 {
	if timeout == 0 {
		return defaultTimeout
	}
	return timeout
}

func getStringOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func getStringSliceOrDefault(value, defaultValue []string) []string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getInt32OrDefault(value int32, defaultValue int32) int32 {
	if value == 0 {
		return defaultValue
	}
	return value
}

func sanitizeID(s string) string {
	s = strings.ReplaceAll(s, "/*", "All")
	s = strings.ReplaceAll(s, "*", "Any")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}

```
