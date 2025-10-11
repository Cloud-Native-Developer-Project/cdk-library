# Static Website Stack - Usage Examples

Este documento muestra ejemplos de uso del `StaticWebsiteStack` con diferentes configuraciones, incluyendo WAF.

## Ejemplo 1: Website Básico (Sin WAF)

```go
stacks.NewStaticWebsiteStack(app, "DevStaticWebsite", &stacks.StaticWebsiteStackProps{
    StackProps: awscdk.StackProps{
        Env: &awscdk.Environment{
            Account: jsii.String(account),
            Region:  jsii.String(region),
        },
        StackName:   jsii.String("dev-static-website"),
        Description: jsii.String("Development static website with S3 + CloudFront"),
    },
    BucketName:  "my-website-dev-bucket",
    WebsiteName: "my-website-dev",
    SourcePath:  "stacks/website/dist",
    PriceClass:  "100",
    EnableWAF:   false, // Sin WAF
})
```

## Ejemplo 2: Website con WAF (Protección Web Application)

```go
stacks.NewStaticWebsiteStack(app, "ProdStaticWebsite", &stacks.StaticWebsiteStackProps{
    StackProps: awscdk.StackProps{
        Env: &awscdk.Environment{
            Account: jsii.String(account),
            Region:  jsii.String(region),
        },
        StackName:   jsii.String("prod-static-website"),
        Description: jsii.String("Production website with WAF protection"),
        Tags: &map[string]*string{
            "Environment": jsii.String("Production"),
            "Project":     jsii.String("StaticWebsite"),
            "Security":    jsii.String("WAF-Enabled"),
        },
    },
    BucketName:     "my-website-prod-bucket",
    WebsiteName:    "my-website-prod",
    SourcePath:     "stacks/website/dist",
    PriceClass:     "100",

    // ✅ Habilita WAF con protección para Web Applications
    EnableWAF:      true,
    WafProfileType: waf.ProfileTypeWebApplication, // Protección OWASP Top 10
})
```

**Protección incluida (ProfileTypeWebApplication):**
- OWASP Top 10 vulnerabilities
- SQL Injection protection
- XSS (Cross-Site Scripting) protection
- Rate limiting: 2000 requests/5min por IP
- AWS IP reputation lists

## Ejemplo 3: Website con WAF (Protección API)

```go
import (
    waf "cdk-library/constructs/WAF"
)

stacks.NewStaticWebsiteStack(app, "APIWebsite", &stacks.StaticWebsiteStackProps{
    StackProps: awscdk.StackProps{
        Env: &awscdk.Environment{
            Account: jsii.String(account),
            Region:  jsii.String(region),
        },
        StackName:   jsii.String("api-website"),
        Description: jsii.String("API frontend with aggressive rate limiting"),
    },
    BucketName:     "api-website-bucket",
    WebsiteName:    "api-website",
    SourcePath:     "stacks/website/dist",

    // ✅ Habilita WAF con protección para APIs
    EnableWAF:      true,
    WafProfileType: waf.ProfileTypeAPIProtection, // Protección agresiva para APIs
})
```

**Protección incluida (ProfileTypeAPIProtection):**
- Rate limiting: 10000 requests/5min por IP (más permisivo para APIs legítimas)
- SQL Injection protection
- AWS IP reputation lists
- Known bad inputs blocking

## Ejemplo 4: Website con WAF (Bot Control)

```go
stacks.NewStaticWebsiteStack(app, "EcommerceWebsite", &stacks.StaticWebsiteStackProps{
    StackProps: awscdk.StackProps{
        Env: &awscdk.Environment{
            Account: jsii.String(account),
            Region:  jsii.String(region),
        },
        StackName:   jsii.String("ecommerce-website"),
        Description: jsii.String("E-commerce website with bot protection"),
    },
    BucketName:     "ecommerce-website-bucket",
    WebsiteName:    "ecommerce-website",
    SourcePath:     "stacks/website/dist",

    // ✅ Habilita WAF con protección contra bots
    EnableWAF:      true,
    WafProfileType: waf.ProfileTypeBotControl, // Protección avanzada contra bots
})
```

**Protección incluida (ProfileTypeBotControl):**
- AWS Managed Bot Control (detecta bots maliciosos)
- Rate limiting: 5000 requests/5min por IP
- AWS IP reputation lists
- Known bad inputs blocking
- Bot detection patterns

## Ejemplo 5: Website con Custom Domain + Certificate + WAF

```go
stacks.NewStaticWebsiteStack(app, "CustomDomainWebsite", &stacks.StaticWebsiteStackProps{
    StackProps: awscdk.StackProps{
        Env: &awscdk.Environment{
            Account: jsii.String(account),
            Region:  jsii.String(region),
        },
        StackName:   jsii.String("custom-domain-website"),
        Description: jsii.String("Production website with custom domain and WAF"),
        Tags: &map[string]*string{
            "Environment": jsii.String("Production"),
            "Security":    jsii.String("High"),
        },
    },
    BucketName:     "mycompany-website-prod",
    WebsiteName:    "mycompany-website",
    SourcePath:     "stacks/website/dist",
    PriceClass:     "100",

    // Custom Domain Configuration
    DomainNames: []string{
        "www.mycompany.com",
        "mycompany.com",
    },
    CertificateArn: "arn:aws:acm:us-east-1:123456789012:certificate/xxxxxx",

    // WAF Protection
    EnableWAF:      true,
    WafProfileType: waf.ProfileTypeWebApplication,
})
```

## Ejemplo 6: Multi-Environment Setup

```go
package main

import (
    stacks "cdk-library/stacks/website"
    waf "cdk-library/constructs/WAF"
    "fmt"
    "os"

    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/jsii-runtime-go"
)

func main() {
    defer jsii.Close()

    app := awscdk.NewApp(nil)
    account := os.Getenv("CDK_DEFAULT_ACCOUNT")
    region := os.Getenv("CDK_DEFAULT_REGION")

    // DEVELOPMENT - Sin WAF (ahorro de costos)
    stacks.NewStaticWebsiteStack(app, "DevWebsite", &stacks.StaticWebsiteStackProps{
        StackProps: awscdk.StackProps{
            Env: &awscdk.Environment{
                Account: jsii.String(account),
                Region:  jsii.String(region),
            },
            StackName: jsii.String("dev-website"),
        },
        BucketName:  fmt.Sprintf("dev-website-%s", account),
        WebsiteName: "dev-website",
        SourcePath:  "stacks/website/dist",
        EnableWAF:   false, // Sin WAF en dev
    })

    // STAGING - WAF con protección básica
    stacks.NewStaticWebsiteStack(app, "StagingWebsite", &stacks.StaticWebsiteStackProps{
        StackProps: awscdk.StackProps{
            Env: &awscdk.Environment{
                Account: jsii.String(account),
                Region:  jsii.String(region),
            },
            StackName: jsii.String("staging-website"),
        },
        BucketName:     fmt.Sprintf("staging-website-%s", account),
        WebsiteName:    "staging-website",
        SourcePath:     "stacks/website/dist",
        EnableWAF:      true,
        WafProfileType: waf.ProfileTypeWebApplication, // Protección estándar
    })

    // PRODUCTION - WAF con protección contra bots
    stacks.NewStaticWebsiteStack(app, "ProdWebsite", &stacks.StaticWebsiteStackProps{
        StackProps: awscdk.StackProps{
            Env: &awscdk.Environment{
                Account: jsii.String(account),
                Region:  jsii.String(region),
            },
            StackName: jsii.String("prod-website"),
            Tags: &map[string]*string{
                "Environment": jsii.String("Production"),
                "Security":    jsii.String("Maximum"),
                "Compliance":  jsii.String("Required"),
            },
        },
        BucketName:     fmt.Sprintf("prod-website-%s", account),
        WebsiteName:    "prod-website",
        SourcePath:     "stacks/website/dist",
        DomainNames:    []string{"www.mycompany.com"},
        CertificateArn: "arn:aws:acm:us-east-1:123456789012:certificate/xxxxxx",
        EnableWAF:      true,
        WafProfileType: waf.ProfileTypeBotControl, // Protección máxima
    })

    app.Synth(nil)
}
```

## Comparación de Perfiles WAF

| Característica | ProfileTypeWebApplication | ProfileTypeAPIProtection | ProfileTypeBotControl |
|----------------|---------------------------|--------------------------|----------------------|
| **Rate Limiting** | 2000 req/5min | 10000 req/5min | 5000 req/5min |
| **OWASP Top 10** | ✅ | ✅ | ✅ |
| **SQL Injection** | ✅ | ✅ | ✅ |
| **XSS Protection** | ✅ | ❌ | ✅ |
| **Bot Detection** | ❌ | ❌ | ✅ AWS Managed |
| **AWS IP Reputation** | ✅ | ✅ | ✅ |
| **Known Bad Inputs** | ✅ | ✅ | ✅ |
| **Cost** | ~$5-10/month | ~$5-10/month | ~$10-20/month |
| **Uso Recomendado** | Sitios web públicos | REST APIs | E-commerce, Alta seguridad |

## Outputs del Stack (Con WAF)

Cuando `EnableWAF: true`, el stack genera outputs adicionales:

```bash
# CloudFormation Outputs
Outputs:
  BucketName = my-website-prod-bucket
  DistributionDomain = d1234567890abc.cloudfront.net
  WebsiteURL = https://d1234567890abc.cloudfront.net

  # WAF Outputs (solo si EnableWAF=true)
  WAFEnabled = Yes - WEB_APPLICATION
  WAFWebACLArn = arn:aws:wafv2:us-east-1:123456789012:global/webacl/my-website-prod-waf/xxxxx
```

## Costos Estimados

### Sin WAF:
- **S3**: ~$0.023/GB + $0.0004/1000 requests
- **CloudFront**: ~$0.085/GB (primeros 10TB)
- **Total mensual (sitio pequeño, 10GB, 100K requests)**: ~$2-5/mes

### Con WAF:
- **WAF Base**: ~$5/mes (fijo)
- **WAF Requests**: ~$0.60/millón de requests
- **AWS Managed Rules**: ~$1-2/mes por regla
- **Total mensual (sitio pequeño con WAF)**: ~$10-15/mes

### Con WAF Bot Control:
- **Bot Control adicional**: ~$10/mes
- **Total mensual (sitio con Bot Control)**: ~$20-25/mes

## Deployment

```bash
# 1. Configurar credenciales AWS
export CDK_DEFAULT_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
export CDK_DEFAULT_REGION=$(aws configure get region)

# 2. Bootstrap CDK (primera vez solamente)
cdk bootstrap

# 3. Preparar contenido del sitio
# Asegúrate de tener index.html en stacks/website/dist/

# 4. Deploy del stack
cdk deploy DevStaticWebsite

# 5. Verificar outputs
# CloudFormation mostrará:
# - BucketName
# - DistributionDomain
# - WebsiteURL
# - WAFEnabled (si aplica)
# - WAFWebACLArn (si aplica)
```

## Verificar WAF en AWS Console

1. **AWS WAF Console**: https://console.aws.amazon.com/wafv2/
2. Navega a "Web ACLs" → Región "Global (CloudFront)"
3. Busca tu Web ACL: `<WebsiteName>-waf`
4. Revisa métricas en CloudWatch:
   - Requests bloqueados
   - Rate limit triggers
   - Rule matches

## Testing WAF Protection

### Test Rate Limiting:
```bash
# Envía múltiples requests para activar rate limiting
for i in {1..2500}; do
  curl -s https://your-cloudfront-domain.cloudfront.net/ > /dev/null
  echo "Request $i"
done

# Deberías recibir 403 Forbidden después de 2000 requests (WebApplication profile)
```

### Test SQL Injection (Debe ser bloqueado):
```bash
curl "https://your-cloudfront-domain.cloudfront.net/?id=1' OR '1'='1"
# Response: 403 Forbidden
```

### Test XSS (Debe ser bloqueado):
```bash
curl "https://your-cloudfront-domain.cloudfront.net/?name=<script>alert('XSS')</script>"
# Response: 403 Forbidden
```

## Best Practices

1. **Desarrollo**: No uses WAF (ahorro de costos)
2. **Staging**: Usa `ProfileTypeWebApplication` para testing
3. **Producción**:
   - Sitios públicos → `ProfileTypeWebApplication`
   - APIs → `ProfileTypeAPIProtection`
   - E-commerce/High-value → `ProfileTypeBotControl`
4. **Monitoreo**: Configura CloudWatch Alarms para requests bloqueados
5. **Testing**: Siempre prueba WAF en staging antes de producción

## Troubleshooting

### Error: "WAF must be in us-east-1 for CloudFront"
**Solución**: Asegúrate que `CDK_DEFAULT_REGION=us-east-1` para CloudFront distributions.

### Error: "Rate limit blocking legitimate users"
**Solución**: Cambia a `ProfileTypeAPIProtection` (10000 req/5min) o ajusta manualmente el rate limit en el código del construct.

### Website bloqueado por WAF
**Solución**: Revisa logs en CloudWatch y ajusta reglas. Puedes temporalmente deshabilitar WAF con `EnableWAF: false`.

## Referencias

- [Construct S3 README](../../constructs/S3/README.md)
- [Construct WAF README](../../constructs/WAF/README.md)
- [Construct CloudFront README](../../constructs/CloudFront/README.md)
- [AWS WAF Documentation](https://docs.aws.amazon.com/waf/)
- [CloudFront + WAF Best Practices](https://docs.aws.amazon.com/waf/latest/developerguide/cloudfront-features.html)
