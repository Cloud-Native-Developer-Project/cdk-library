# 📌 Changelog

Todos los cambios relevantes de este proyecto se documentarán en este archivo.

El formato sigue las recomendaciones de [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/)  
y este proyecto se adhiere a [Semantic Versioning](https://semver.org/lang/es/).

---

## [0.2.0] - 2025-09-22

### Added

- 🏗️ Constructo **Cloudfront**:
  - Soporte para múltiples orígenes (S3, S3 Website, HTTP, ALB)
  - Configuración de caching avanzado (cache policies, response headers, request policies)
  - SSL/TLS con certificados de ACM
  - Restricciones geográficas
  - Integración con WAF
  - Edge Functions (Lambda\@Edge, Function)
  - Logging y métricas avanzadas

### Changed

- 🔧 Documentación de **S3** mejorada en el README.

---

## [0.1.0] - 2025-09-18

### Added

- 🚀 Primer release de la librería.
- ✨ Constructo **S3** con configuraciones avanzadas:
  - Seguridad: enforce SSL, cifrado y bloqueo de acceso público.
  - Versionado y Object Lock.
  - Reglas de ciclo de vida e Intelligent Tiering.
  - Replicación cross-region.
  - Logging, métricas e inventarios.
  - Hosting de sitios estáticos.
