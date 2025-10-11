# 🏗️ CDK Constructs Library for Go

Bienvenido a **cdk-library**, una colección de _constructs reutilizables_ en **Go** para simplificar la creación de componentes de AWS usando [AWS CDK](https://docs.aws.amazon.com/cdk/latest/guide/home.html).

El objetivo es ofrecer **constructos de alto nivel** que implementen buenas prácticas de seguridad, optimización de costos y rendimiento, listos para integrarse en proyectos de infraestructura como código.

## 🎯 Arquitectura: Design Patterns Enterprise

A partir de la versión **0.3.0**, esta librería adopta una arquitectura modular basada en **Design Patterns empresariales** (Factory + Strategy), permitiendo:

- ✅ **Extensibilidad**: Agregar nuevos casos de uso sin modificar código existente
- ✅ **Mantenibilidad**: Archivos pequeños (~100-150 líneas) con responsabilidad única
- ✅ **Testabilidad**: Cada componente puede testearse de forma aislada
- ✅ **Escalabilidad**: Los stacks orquestan, los constructs implementan la lógica
- ✅ **Clean Code**: Separación de concerns, Open/Closed Principle, DRY

**Patrón implementado:**
```
Factory (Punto de entrada) → Strategy (Implementación específica) → AWS Resources
```

Ver [CHANGELOG.md](./CHANGELOG.md#030---2025-10-09) para detalles de la arquitectura.

---

## 📦 Constructos disponibles

### **S3** → 6 estrategias especializadas (**Factory + Strategy Pattern** ✨)

Buckets optimizados para casos de uso específicos con seguridad, costos y performance balanceados:

| Estrategia | Uso Principal | Seguridad | Retención |
|------------|---------------|-----------|-----------|
| **CloudFront Origin** | Sitios estáticos, SPAs | KMS, TLS 1.2 | 1 año |
| **Data Lake** | Analytics, Big Data | KMS, Versioning | Multi-tier |
| **Backup** | Disaster Recovery | KMS, Object Lock (GOVERNANCE) | 10 años |
| **Media Streaming** | Video/Audio CDN | S3_MANAGED, CORS | 1 año |
| **Enterprise** | PII, Compliance | KMS, Object Lock (COMPLIANCE), TLS 1.3 | 7 años |
| **Development** | Dev/Test | S3_MANAGED, Auto-delete | 30 días |

**Uso:**
```go
bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeCloudfrontOAC,
        BucketName: "my-website-prod",
    })
```

Ver [constructs/S3/README.md](./constructs/S3/README.md) para documentación completa.

---

### **CloudFront** → Distribución global con seguridad y caching (**Factory + Strategy Pattern** ✨)

Estrategias optimizadas por tipo de origen:

- **S3 Strategy**: Origin Access Control (OAC), cache optimizado, security headers
- **Futuras**: API Gateway, ALB, Custom HTTP Origins

**Features**: Cache policies, SSL/TLS (ACM), geo-restrictions, WAF integration, SPA support

**Uso:**
```go
distribution := cloudfront.NewDistributionV2(stack, "CDN",
    cloudfront.CloudFrontPropertiesV2{
        OriginType: cloudfront.OriginTypeS3,
        S3Bucket:   bucket,
        AutoConfigureS3BucketPolicy: true,
    })
```

Ver [constructs/CloudFront/README.md](./constructs/CloudFront/README.md) para detalles.

---

### **WAF** → Web Application Firewall con reglas gestionadas (**Factory + Strategy Pattern** ✨)

Protección contra amenazas comunes con estrategias pre-configuradas:

- **Web Application Strategy**: OWASP Top 10, rate limiting (2000 req/5min), SQL injection, XSS
- **API Strategy**: Rate limiting (10000 req/5min), token validation, bot protection
- **OWASP Strategy**: Core Rule Set completo, protección contra vulnerabilidades conocidas

**Features**: AWS Managed Rules, custom rules, CloudWatch metrics, geo-blocking

**Uso:**
```go
webacl := waf.NewWebApplicationFirewallV2(stack, "WAF",
    waf.WebApplicationFirewallFactoryProps{
        WafType: waf.WafTypeWebApplication,
        Scope:   waf.WafScopeCloudfront,
        Name:    "my-app-waf",
    })
```

Ver [constructs/WAF/README.md](./constructs/WAF/README.md) para casos de uso y reglas.

---

## 🛠️ Roadmap

### Fase 1: Arquitectura Foundation ✅
- [x] **CloudFront** → Factory + Strategy Pattern (Caso piloto)
- [x] **S3** → 6 estrategias especializadas (CloudFront Origin, Data Lake, Backup, Media, Enterprise, Dev)
- [x] **WAF** → 3 estrategias (Web Application, API, OWASP)

### Fase 2: Extensiones (En progreso)
- [ ] **CloudFront Strategies adicionales**: API Gateway, ALB, Custom HTTP Origins
- [ ] **Lambda Construct**: Factory + Strategy para casos de uso comunes
- [ ] **API Gateway Construct**: REST/HTTP APIs con autenticación integrada
- [ ] **DynamoDB Construct**: Estrategias para transaccional, analytics, time-series

### Fase 3: Enterprise Features (Futuro)
- [ ] **VPC Construct**: Subnets, NAT, security groups optimizados
- [ ] **Aurora Serverless v2**: Auto-scaling con monitoreo avanzado
- [ ] **ECS/Fargate**: Containers con service discovery y load balancing
- [ ] **Template/Scaffold**: CLI para generar nuevos constructos con el patrón

**Cobertura actual**: 10 estrategias implementadas | **Meta**: 95% de casos de uso AWS comunes

---

## 🎯 Casos de Uso Completos

Esta librería está diseñada para implementar arquitecturas completas de AWS siguiendo best practices:

### Stack 1: Static Website con CDN + WAF
```go
// Bucket S3 optimizado para CloudFront
bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeCloudfrontOAC,
        BucketName: "my-website-prod",
    })

// WAF con protección OWASP
webacl := waf.NewWebApplicationFirewallV2(stack, "WAF",
    waf.WebApplicationFirewallFactoryProps{
        WafType: waf.WafTypeWebApplication,
        Scope:   waf.WafScopeCloudfront,
        Name:    "website-waf",
    })

// CloudFront con S3 origin
distribution := cloudfront.NewDistributionV2(stack, "CDN",
    cloudfront.CloudFrontPropertiesV2{
        OriginType: cloudfront.OriginTypeS3,
        S3Bucket:   bucket,
        WebACLId:   webacl.Arn(),
        AutoConfigureS3BucketPolicy: true,
    })
```

### Stack 2: Data Lake Analytics
```go
// Bucket optimizado para big data
dataLake := s3.NewSimpleStorageServiceFactory(stack, "DataLake",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeDataLake,
        BucketName: "analytics-datalake",
    })
// Integra con: Glue, Athena, EMR, Redshift Spectrum
```

### Stack 3: Enterprise Compliance Archive
```go
// Bucket con Object Lock COMPLIANCE (7 años)
archive := s3.NewSimpleStorageServiceFactory(stack, "Archive",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeEnterprise,
        BucketName: "financial-records",
    })
// Cumple: SOX, HIPAA, GDPR, PCI-DSS, SEC 17a-4
```

---

## 📢 Cambios

Este proyecto mantiene un historial de versiones en el archivo [CHANGELOG.md](./CHANGELOG.md), siguiendo el formato [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/) y la convención de [Semantic Versioning](https://semver.org/lang/es/).

---
