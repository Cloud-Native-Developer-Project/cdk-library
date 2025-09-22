# 🏗️ CDK Constructs Library for Go

Bienvenido a **cdk-library**, una colección de _constructs reutilizables_ en **Go** para simplificar la creación de componentes de AWS usando [AWS CDK](https://docs.aws.amazon.com/cdk/latest/guide/home.html).

El objetivo es ofrecer **constructos de alto nivel** que implementen buenas prácticas de seguridad, optimización de costos y rendimiento, listos para integrarse en proyectos de infraestructura como código.

---

## 📦 Constructos disponibles

- **S3** → Bucket de Amazon S3 con configuración extendida:

  - Seguridad (SSL, cifrado, versionado, bloqueo de acceso público, Object Lock)
  - Optimización de costos (tiering, reglas de ciclo de vida)
  - Replicación cross-region
  - Logs, inventarios y métricas
  - Transfer Acceleration y CORS
  - Hosting para sitios estáticos

- **CloudFront** → Distribución global con seguridad y caching optimizado:

  - Soporte para múltiples orígenes (S3, S3 Website, HTTP, ALB)
  - Configuración de caching avanzado (cache policies, response headers, request policies)
  - SSL/TLS con certificados de ACM
  - Restricciones geográficas
  - Integración con WAF
  - Edge Functions (Lambda\@Edge, Function)
  - Logging y métricas avanzadas

---

## 🛠️ Roadmap

- [x] Constructo `S3`
- [x] Constructo `CloudFront`
- [ ] Constructo `Lambda` (integración con CloudFront u orígenes personalizados)
- [ ] Constructo `VPC` con subnets, NAT y seguridad integrada
- [ ] Constructo `Aurora Serverless v2` con monitoreo automático

---

## 📢 Cambios

Este proyecto mantiene un historial de versiones en el archivo [CHANGELOG.md](./CHANGELOG.md), siguiendo el formato [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/) y la convención de [Semantic Versioning](https://semver.org/lang/es/).

---
