# 📌 Changelog

Todos los cambios relevantes de este proyecto se documentarán en este archivo.

El formato sigue las recomendaciones de [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/)  
y este proyecto se adhiere a [Semantic Versioning](https://semver.org/lang/es/).

---

## [0.2.0] - 2025-09-19

### Added

- 🏗️ Constructo **API Gateway**:
  - Creación de APIs REST
  - Configuración de recursos y métodos
  - Soporte para CORS
  - Definición de stages (`dev`, `prod`)
  - Logs y métricas en CloudWatch
  - Configuración de throttling básico

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
