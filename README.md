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

- **API Gateway** → API REST escalable con buenas prácticas:
  - Definición de recursos y métodos
  - Soporte para CORS
  - Stages (`dev`, `prod`)
  - Logs y métricas en CloudWatch
  - Throttling básico

---

## 🛠️ Roadmap

- [x] Constructo `S3`
- [x] Constructo `API Gateway`
- [ ] Constructo `Lambda` (integración con API Gateway)
- [ ] Constructo `VPC` con subnets, NAT y seguridad integrada
- [ ] Constructo `Aurora Serverless v2` con monitoreo automático

---

## 📢 Cambios

Este proyecto mantiene un historial de versiones en el archivo [CHANGELOG.md](./CHANGELOG.md), siguiendo el formato [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/) y la convención de [Semantic Versioning](https://semver.org/lang/es/).
