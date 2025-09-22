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