# 🛡️ AWS WAF Construct - Factory + Strategy Pattern

AWS Web Application Firewall (WAF) construct con múltiples perfiles de seguridad usando Factory + Strategy pattern.

## 📋 Tabla de Contenidos

- [Características](#-características)
- [Arquitectura](#-arquitectura)
- [Perfiles Disponibles](#-perfiles-disponibles)
- [Uso Básico](#-uso-básico)
- [Ejemplos Completos](#-ejemplos-completos)
- [Costos](#-costos)
- [Reglas Incluidas](#-reglas-incluidas)

---

## ✨ Características

✅ **Factory + Strategy Pattern**: Arquitectura modular y extensible
✅ **3 Perfiles Pre-configurados**: Web Application, API Protection, Bot Control
✅ **AWS Managed Rules**: OWASP Top 10, SQL injection, IP reputation, etc.
✅ **Rate Limiting**: Configurable por IP
✅ **Geo-blocking**: Bloquea/permite países específicos
✅ **IP Allow/Block Lists**: Whitelist y blacklist de IPs
✅ **CloudWatch Metrics**: Monitoreo completo habilitado
✅ **Scope Flexible**: CloudFront (global) o Regional (ALB, API Gateway)

---

## 🏗️ Arquitectura

```
WAF Factory (waf_factory.go)
      │
      ├─ ProfileType: WEB_APPLICATION
      │  └─ WAFWebApplicationStrategy (waf_web_application.go)
      │     └─ Core Rule Set + Known Bad Inputs + IP Reputation
      │
      ├─ ProfileType: API_PROTECTION
      │  └─ WAFAPIProtectionStrategy (waf_api_protection.go)
      │     └─ SQL Injection + Body Size Limits + Higher Rate Limits
      │
      └─ ProfileType: BOT_CONTROL
         └─ WAFBotControlStrategy (waf_bot_control.go)
            └─ Bot Control ML + CAPTCHA + All Baseline Rules
```

---

## 🎯 Perfiles Disponibles

### 1. Web Application (`ProfileTypeWebApplication`)

**Recomendado para:**
- Sitios web estáticos (React, Vue, Angular)
- Single Page Applications (SPAs)
- JAMstack sites

**Reglas incluidas:**
- ✅ AWS Managed Rules - Core Rule Set (OWASP Top 10)
- ✅ AWS Managed Rules - Known Bad Inputs
- ✅ AWS Managed Rules - IP Reputation List
- ✅ AWS Managed Rules - Anonymous IP List (VPNs, Tor, proxies)
- ✅ Rate Limiting (configurable)
- ✅ Geo-blocking (opcional)

**Costo estimado:** ~$9/mes + $0.60 per 1M requests

### 2. API Protection (`ProfileTypeAPIProtection`)

**Recomendado para:**
- REST APIs (API Gateway, ALB)
- GraphQL APIs
- Backend APIs para SPAs

**Reglas incluidas:**
- ✅ Todas las reglas de Web Application
- ✅ AWS Managed Rules - SQL Database Protection
- ✅ Request Body Size Constraints (8KB limit)
- ✅ Higher Rate Limits (10,000 req/5min default)

**Costo estimado:** ~$10/mes + $0.60 per 1M requests

### 3. Bot Control (`ProfileTypeBotControl`) 🚀 PREMIUM

**Recomendado para:**
- E-commerce (prevenir inventory hoarding)
- Sites con problemas de scraping
- Aplicaciones de alto valor

**Reglas incluidas:**
- ✅ Todas las reglas de API Protection
- ✅ AWS Bot Control (Machine Learning detection)
- ✅ CAPTCHA/Challenge configuration
- ✅ Stricter Rate Limits (500 req/5min default)

**Costo estimado:** ~$20/mes + $1.60 per 1M requests

---

## 🚀 Uso Básico

### Ejemplo 1: Web Application Protection (Sitio Web Estático)

```go
package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	waf "cdk-library/constructs/WAF"
)

func NewMyStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// Crear WAF Web ACL
	webACL := waf.NewWebApplicationFirewallFactory(stack, "WebsiteWAF",
		waf.WAFFactoryProps{
			Scope:       waf.ScopeCloudFront,  // CloudFront distributions
			ProfileType: waf.ProfileTypeWebApplication,
			Name:        "MyWebsite-WAF",

			// Opcional: Rate limiting
			RateLimitRequests: jsii.Int64(2000), // 2000 req/5min per IP
		})

	// El WebACL ARN está disponible en: webACL.AttrArn()

	return stack
}
```

### Ejemplo 2: Con Geo-blocking e IP Lists

```go
webACL := waf.NewWebApplicationFirewallFactory(stack, "WebsiteWAF",
	waf.WAFFactoryProps{
		Scope:       waf.ScopeCloudFront,
		ProfileType: waf.ProfileTypeWebApplication,

		// Rate limiting
		RateLimitRequests: jsii.Int64(1000),

		// Geo-blocking: Bloquear estos países
		GeoBlockCountries: []string{"CN", "RU", "KP"},

		// IP Blocklist
		BlockedIPs: []string{
			"192.0.2.0/24",
			"198.51.100.0/24",
		},

		// IP Whitelist (siempre permitir)
		AllowedIPs: []string{
			"203.0.113.0/24", // Office IP
		},
	})
```

### Ejemplo 3: API Protection

```go
webACL := waf.NewWebApplicationFirewallFactory(stack, "APIWAF",
	waf.WAFFactoryProps{
		Scope:       waf.ScopeRegional,  // Para ALB/API Gateway
		ProfileType: waf.ProfileTypeAPIProtection,

		// APIs normalmente soportan más tráfico
		RateLimitRequests: jsii.Int64(10000),
	})
```

### Ejemplo 4: Bot Control (Premium)

```go
webACL := waf.NewWebApplicationFirewallFactory(stack, "EcommerceWAF",
	waf.WAFFactoryProps{
		Scope:       waf.ScopeCloudFront,
		ProfileType: waf.ProfileTypeBotControl,

		// Rate limit más estricto
		RateLimitRequests: jsii.Int64(500),

		// Bloquear proxies anónimos (incluido en strategy)
		GeoBlockCountries: []string{"CN", "RU"},
	})
```

---

## 🔗 Ejemplos Completos

### Stack Completo: S3 + CloudFront + WAF

```go
package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"

	cloudfront "cdk-library/constructs/Cloudfront"
	s3 "cdk-library/constructs/S3"
	waf "cdk-library/constructs/WAF"
)

type SecureWebsiteStackProps struct {
	awscdk.StackProps
	BucketName  string
	WebsiteName string
	EnableWAF   bool
}

func NewSecureWebsiteStack(scope constructs.Construct, id string, props *SecureWebsiteStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// =============================================================================
	// 1. CREATE S3 BUCKET (Private, for CloudFront Origin)
	// =============================================================================
	autoDelete := true
	bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
		s3.SimpleStorageServiceFactoryProps{
			BucketType:        s3.BucketTypeCloudfrontOAC,
			BucketName:        props.BucketName,
			RemovalPolicy:     "destroy",
			AutoDeleteObjects: &autoDelete,
		})

	// =============================================================================
	// 2. CREATE WAF WEB ACL (if enabled)
	// =============================================================================
	var webACLArn *string
	if props.EnableWAF {
		webACL := waf.NewWebApplicationFirewallFactory(stack, "WebsiteWAF",
			waf.WAFFactoryProps{
				Scope:             waf.ScopeCloudFront,
				ProfileType:       waf.ProfileTypeWebApplication,
				Name:              props.WebsiteName + "-WAF",
				RateLimitRequests: jsii.Int64(2000),
			})
		webACLArn = webACL.AttrArn()
	}

	// =============================================================================
	// 3. CREATE CLOUDFRONT DISTRIBUTION with WAF
	// =============================================================================
	distribution := cloudfront.NewDistributionV2(stack, "WebsiteDistribution",
		cloudfront.CloudFrontPropertiesV2{
			OriginType:                  cloudfront.OriginTypeS3,
			S3Bucket:                    bucket,
			WebAclArn:                   *webACLArn,  // ← WAF integration
			Comment:                     props.WebsiteName + " - Secure Distribution",
			EnableAccessLogging:         false,
			AutoConfigureS3BucketPolicy: true,
		})

	// Stack outputs
	awscdk.NewCfnOutput(stack, jsii.String("DistributionDomain"), &awscdk.CfnOutputProps{
		Value:       distribution.DomainName(),
		Description: jsii.String("CloudFront domain (protected by WAF)"),
	})

	if webACLArn != nil {
		awscdk.NewCfnOutput(stack, jsii.String("WAFArn"), &awscdk.CfnOutputProps{
			Value:       webACLArn,
			Description: jsii.String("WAF Web ACL ARN"),
		})
	}

	return stack
}
```

---

## 💰 Costos

### Pricing (us-east-1)

| Componente | Costo |
|------------|-------|
| Web ACL (por distribución) | $5.00/mes |
| AWS Managed Rule Group | $1.00/mes cada uno |
| Custom Rule | $1.00/mes cada una |
| Requests procesados | $0.60 por 1M requests |
| Bot Control (premium) | $10.00/mes + $1.00/1M req |

### Ejemplos de Costos Mensuales

**Sitio web pequeño (1M requests/mes):**
```
Web ACL:                     $5.00
Core Rule Set:               $1.00
Known Bad Inputs:            $1.00
IP Reputation:               $1.00
Anonymous IP List:           $1.00
Rate Limit Rule:             $1.00
Requests (1M):               $0.60
────────────────────────────────
Total:                      $10.60/mes
```

**API con tráfico medio (10M requests/mes):**
```
Web ACL:                     $5.00
Core Rule Set:               $1.00
SQL Database:                $1.00
Known Bad Inputs:            $1.00
IP Reputation:               $1.00
Rate Limit Rule:             $1.00
Requests (10M):              $6.00
────────────────────────────────
Total:                      $16.00/mes
```

**E-commerce con Bot Control (10M requests/mes):**
```
Web ACL:                     $5.00
Bot Control:                $10.00
Core Rule Set:               $1.00
SQL Database:                $1.00
Known Bad Inputs:            $1.00
IP Reputation:               $1.00
Anonymous IP List:           $1.00
Requests (10M @ $0.60):      $6.00
Bot Control Reqs (10M @ $1): $10.00
────────────────────────────────
Total:                      $36.00/mes
```

---

## 📊 Reglas Incluidas por Perfil

### Web Application Profile

| Prioridad | Regla | Acción | Descripción |
|-----------|-------|--------|-------------|
| 0 | Rate Limit | Block (429) | Límite de requests por IP |
| 10 | Geo Blocking | Block | Bloquea países especificados |
| 20 | IP Blocklist | Block | Bloquea IPs maliciosas |
| 30 | IP Allowlist | Allow | Whitelist de IPs confiables |
| 40 | AWS Core Rule Set | Managed | OWASP Top 10 protection |
| 50 | Known Bad Inputs | Managed | Payloads maliciosos conocidos |
| 60 | IP Reputation | Managed | IPs con historial de ataques |
| 70 | Anonymous IP List | Managed | Bloquea VPNs, Tor, proxies |

### API Protection Profile

Todas las reglas de Web Application **MÁS**:

| Prioridad | Regla | Acción | Descripción |
|-----------|-------|--------|-------------|
| 35 | Body Size Limit | Block (413) | Rechaza payloads > 8KB |
| 45 | SQL Database Rules | Managed | SQL injection avanzado |

### Bot Control Profile

Todas las reglas de API Protection **MÁS**:

| Prioridad | Regla | Acción | Descripción |
|-----------|-------|--------|-------------|
| 40 | Bot Control (ML) | Challenge/Block | Detección con Machine Learning |

---

## 🔍 Monitoreo y Logs

WAF provee métricas automáticas en CloudWatch:

### Métricas Disponibles

- `AllowedRequests`: Requests permitidos
- `BlockedRequests`: Requests bloqueados
- `CountedRequests`: Requests en modo Count
- `SampledRequests`: Sample de requests para análisis

### Ver Requests Bloqueados

```bash
# AWS CLI
aws wafv2 get-sampled-requests \
  --scope CLOUDFRONT \
  --web-acl-id <web-acl-id> \
  --rule-metric-name <rule-name> \
  --time-window StartTime=<timestamp>,EndTime=<timestamp> \
  --max-items 100
```

---

## 🛠️ Próximas Implementaciones

- [ ] `ProfileTypeWordPress`: Protección específica para WordPress
- [ ] `ProfileTypeCustom`: Reglas completamente personalizadas
- [ ] Logging a S3 / CloudWatch Logs / Firehose
- [ ] Rate limiting por URI path
- [ ] CAPTCHA configuration personalizada
- [ ] Account Takeover Prevention (ATP)
- [ ] Account Creation Fraud Prevention (ACFP)

---

## 📚 Referencias

- [AWS WAF Developer Guide](https://docs.aws.amazon.com/waf/latest/developerguide/)
- [AWS Managed Rules for WAF](https://docs.aws.amazon.com/waf/latest/developerguide/aws-managed-rule-groups-list.html)
- [WAF Pricing](https://aws.amazon.com/waf/pricing/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

**Arquitectura:** Factory + Strategy Pattern
**Versión CDK:** AWS CDK Go v2
**Mantenedor:** cdk-library
