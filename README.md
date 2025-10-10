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

- **S3** → Bucket de Amazon S3 con configuración extendida:

  - Seguridad (SSL, cifrado, versionado, bloqueo de acceso público, Object Lock)
  - Optimización de costos (tiering, reglas de ciclo de vida)
  - Replicación cross-region
  - Logs, inventarios y métricas
  - Transfer Acceleration y CORS
  - Hosting para sitios estáticos

- **CloudFront** → Distribución global con seguridad y caching optimizado (**Factory + Strategy Pattern** ✨):

  - **Arquitectura modular**: Factory Pattern para selección de estrategia según tipo de origen
  - **Estrategias implementadas**:
    - `S3CloudFrontStrategy`: Origin Access Control (OAC), cache optimizado, security headers
  - **Estrategias futuras**: API Gateway, ALB, Custom Origins
  - **Features**:
    - Configuración de caching avanzado (cache policies, response headers, request policies)
    - SSL/TLS con certificados de ACM
    - Restricciones geográficas
    - Integración con WAF
    - SPA support (error redirects a index.html)
    - Auto-configuración de bucket policies para OAC

  **Uso:**
  ```go
  distribution := cloudfront.NewDistributionV2(stack, "Distribution",
      cloudfront.CloudFrontPropertiesV2{
          OriginType: cloudfront.OriginTypeS3,
          S3Bucket:   bucket,
          AutoConfigureS3BucketPolicy: true,
      })
  ```

---

## 🛠️ Roadmap

### Fase 1: Arquitectura Foundation ✅
- [x] ~~Constructo `S3` (Monolítico)~~
- [x] **CloudFront con Factory + Strategy Pattern** (Caso piloto) ✨

### Fase 2: Refactoring & Patterns (En progreso)
- [ ] **S3 Construct Refactoring** → Factory + Strategy Pattern
  - Strategies: `CloudFrontOrigin`, `DataLake`, `Backup`, `MediaStreaming`, `Enterprise`, `Development`
  - Eliminar helper functions monolíticas
- [ ] **WAF Construct** → Factory + Strategy Pattern
  - Strategies: `WebApplication`, `API`, `OWASP`, `Custom`
  - Integración automática con CloudFront

### Fase 3: Extensiones (Futuro)
- [ ] CloudFront Strategies adicionales: `API Gateway`, `ALB`, `Custom HTTP`
- [ ] Constructo `Lambda` con Factory Pattern
- [ ] Constructo `VPC` con subnets, NAT y seguridad integrada
- [ ] Constructo `Aurora Serverless v2` con monitoreo automático
- [ ] Template/Scaffold para crear nuevos constructos siguiendo el patrón

**Meta**: Todos los constructos nuevos seguirán el patrón Factory + Strategy a partir de la versión 0.3.0

---

## 📢 Cambios

Este proyecto mantiene un historial de versiones en el archivo [CHANGELOG.md](./CHANGELOG.md), siguiendo el formato [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/) y la convención de [Semantic Versioning](https://semver.org/lang/es/).

---
