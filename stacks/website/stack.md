# Static Website Stack - Production-Grade Architecture

## 📋 Overview

The `StaticWebsiteStack` implements a **production-ready, secure, and cost-optimized architecture** for hosting static websites on AWS using the CloudFront + S3 pattern with Origin Access Control (OAC). This stack follows AWS Well-Architected Framework best practices and leverages custom L3 constructs for consistent, reusable infrastructure patterns.

**Architecture Pattern:** CloudFront (CDN) → S3 (Private Origin) + Automated Deployment
**Security Model:** Zero direct S3 access, CloudFront-only via OAC, HTTPS enforcement
**Deployment:** Automated content sync with cache invalidation

---

## 🏗️ Architecture Components

### 1. **S3 Bucket (Private Origin)**
- **Purpose:** Stores static website assets (HTML, CSS, JS, images)
- **Access Model:** Private bucket, accessible only via CloudFront OAC
- **Security:** Block all public access, enforce SSL/TLS
- **Lifecycle:** Auto-delete on stack removal (configurable)

### 2. **CloudFront Distribution**
- **Purpose:** Global content delivery with edge caching
- **Origin:** S3 bucket via Origin Access Control (OAC)
- **Security:** HTTPS redirection, security headers, optional WAF
- **Performance:** HTTP/2 and HTTP/3 support, compression enabled
- **SPA Support:** 403/404 errors redirect to `index.html` for client-side routing

### 3. **S3 Deployment (BucketDeployment)**
- **Purpose:** Automated content upload and cache invalidation
- **Features:** Prune old files, invalidate CloudFront cache, set cache headers
- **Cache Strategy:** 365-day max-age for immutable assets
- **Logging:** CloudWatch Logs with 1-month retention

### 4. **Stack Outputs**
- **BucketName:** S3 bucket identifier
- **DistributionDomain:** CloudFront domain (e.g., `d123abc.cloudfront.net`)
- **WebsiteURL:** Full HTTPS URL to access the site

---

## 📦 Stack Properties (`StaticWebsiteStackProps`)

```go
type StaticWebsiteStackProps struct {
    awscdk.StackProps           // Standard CDK stack properties (env, tags, etc.)

    // Required Configuration
    BucketName  string          // Globally unique S3 bucket name
    WebsiteName string          // Logical name for outputs (e.g., "my-app-dev")
    SourcePath  string          // Local path to website files (e.g., "./dist")

    // Optional: Custom Domain
    DomainNames    []string     // Custom domains (requires DNS setup)
    CertificateArn string        // ACM certificate ARN for HTTPS (must be in us-east-1)

    // Optional: CloudFront Configuration
    PriceClass string           // "100", "200", or "ALL" (geographic coverage)
    EnableWAF  bool             // Enable Web Application Firewall
    WebAclArn  string           // WAF Web ACL ARN (if EnableWAF is true)
}
```

### Property Details

| Property | Required | Default | Description |
|----------|----------|---------|-------------|
| `BucketName` | ✅ | - | Must be globally unique, lowercase, DNS-compliant |
| `WebsiteName` | ✅ | - | Used in CloudFormation export names |
| `SourcePath` | ✅ | - | Must contain `index.html` |
| `DomainNames` | ❌ | `[]` | Requires valid ACM certificate in `us-east-1` |
| `CertificateArn` | ❌ | `""` | Only required if using custom domains |
| `PriceClass` | ❌ | `"100"` | `100` = North America/Europe, `200` = +Asia, `ALL` = Global |
| `EnableWAF` | ❌ | `false` | Requires existing WAF Web ACL |
| `WebAclArn` | ❌ | `""` | Must be in `us-east-1` region |

---

## 🔧 Implementation Walkthrough

### Step 1: S3 Bucket Creation (Lines 38-43)

```go
s3Props := s3.GetCloudFrontOriginProperties()
s3Props.BucketName = props.BucketName
s3Props.RemovalPolicy = "destroy"
s3Props.AutoDeleteObjects = true

bucket := s3.NewBucket(stack, "WebsiteBucket", s3Props)
```

**What's Happening:**
1. Uses the custom S3 construct with pre-configured CloudFront origin properties
2. `GetCloudFrontOriginProperties()` returns:
   - `PublicAccess: false` (private bucket)
   - `Encryption: S3_MANAGED`
   - `Versioned: true` (for deployment rollbacks)
   - `EnforceSSL: true`
   - `EventBridgeEnabled: true` (for automated workflows)

**Well-Architected Alignment:**
- **Security:** Private bucket, no public access, encryption at rest
- **Reliability:** Versioning enabled for rollback capability
- **Cost Optimization:** Destroy policy prevents orphaned resources in dev/test

**Trade-offs:**
- ✅ **Pro:** Secure by default, no accidental public exposure
- ✅ **Pro:** Automatic cleanup on stack deletion
- ⚠️ **Con:** Destroy policy = data loss if stack is deleted (use `retain` in production)

---

### Step 2: CloudFront Distribution (Lines 48-57)

```go
distribution := cloudfront.NewDistributionV2(stack, "WebsiteDistribution",
    cloudfront.CloudFrontPropertiesV2{
        OriginType:                  cloudfront.OriginTypeS3,
        S3Bucket:                    bucket,
        DomainNames:                 props.DomainNames,
        CertificateArn:              props.CertificateArn,
        WebAclArn:                   props.WebAclArn,
        Comment:                     props.WebsiteName + " - Static Website Distribution",
        EnableAccessLogging:         false,
        AutoConfigureS3BucketPolicy: true,
    })
```

**What's Happening:**
1. Uses the **Factory Pattern** (`NewDistributionV2`) to select the appropriate CloudFront strategy
2. `OriginTypeS3` triggers the `S3CloudFrontStrategy` implementation
3. `AutoConfigureS3BucketPolicy: true` automatically creates the S3 bucket policy for OAC

**Under the Hood (from `cloudfront_s3.go`):**
- Creates an **Origin Access Control (OAC)** resource
- Configures default behavior:
  - **Cache Policy:** `CACHING_OPTIMIZED` (CloudFront managed policy)
  - **Response Headers:** `SECURITY_HEADERS` (HSTS, X-Frame-Options, etc.)
  - **Viewer Protocol:** Redirect HTTP → HTTPS
  - **Compression:** Brotli/Gzip enabled
- Sets up SPA error responses (403/404 → `index.html`)
- Adds S3 bucket policy allowing CloudFront service principal access

**Well-Architected Alignment:**
- **Security:** HTTPS enforcement, security headers, OAC (not deprecated OAI)
- **Performance:** Edge caching, compression, HTTP/3 support
- **Cost Optimization:** Efficient caching reduces S3 GET requests
- **Reliability:** Global edge network with automatic failover

**Trade-offs:**
- ✅ **Pro:** OAC is the modern, recommended approach (replaces OAI)
- ✅ **Pro:** Factory pattern allows easy extension for API/ALB origins
- ⚠️ **Con:** Access logging disabled by default (enable in production for compliance)

---

### Step 3: Content Deployment (Lines 62-82)

```go
deployment := awss3deployment.NewBucketDeployment(stack, jsii.String("WebsiteDeployment"),
    &awss3deployment.BucketDeploymentProps{
        Sources: &[]awss3deployment.ISource{
            awss3deployment.Source_Asset(jsii.String(props.SourcePath), nil),
        },
        DestinationBucket:    bucket,
        DestinationKeyPrefix: jsii.String(""),
        Distribution:         distribution,
        DistributionPaths:    &[]*string{jsii.String("/*")},
        CacheControl: &[]awss3deployment.CacheControl{
            awss3deployment.CacheControl_MaxAge(awscdk.Duration_Days(jsii.Number(365))),
            awss3deployment.CacheControl_Immutable(),
        },
        Prune:          jsii.Bool(true),
        RetainOnDelete: jsii.Bool(false),
        LogRetention:   awslogs.RetentionDays_ONE_MONTH,
    })

deployment.Node().AddDependency(distribution)
```

**What's Happening:**
1. **Source_Asset:** Packages local files (`SourcePath`) into a CDK asset
2. **Distribution + DistributionPaths:** Automatically invalidates CloudFront cache on deployment
3. **CacheControl:** Sets HTTP headers for browser/CDN caching (365 days + immutable)
4. **Prune:** Removes old files from S3 that don't exist in the source
5. **RetainOnDelete:** Deletes deployment Lambda on stack removal
6. **AddDependency:** Ensures CloudFront distribution exists before deployment

**BucketDeployment Implementation (CDK Magic):**
- Creates a **Lambda function** that runs during deployment
- Lambda zips source files and uploads to S3
- Calls CloudFront `CreateInvalidation` API for cache purging
- Logs to CloudWatch with 1-month retention

**Well-Architected Alignment:**
- **Operational Excellence:** Automated deployment, no manual S3 uploads
- **Performance:** Aggressive caching (365 days) for static assets
- **Cost Optimization:** Prune removes unused objects (saves storage costs)
- **Reliability:** Atomic deployments with automatic cache invalidation

**Trade-offs:**
- ✅ **Pro:** Fully automated, idempotent deployments
- ✅ **Pro:** 365-day cache reduces bandwidth costs
- ⚠️ **Con:** Lambda cold starts add 5-10s to deployment time
- ⚠️ **Con:** `immutable` cache requires versioned filenames for updates (e.g., `app.abc123.js`)

**Cache Strategy Best Practice:**
```
index.html           → max-age=300 (5 minutes, not immutable)
app.[hash].js        → max-age=31536000, immutable
style.[hash].css     → max-age=31536000, immutable
logo.png             → max-age=86400 (1 day, if not versioned)
```

**Current Implementation Caveat:**
This stack applies the same cache policy to ALL files. For production, consider:
- Separate deployment for versioned vs. non-versioned assets
- CloudFront Functions to modify Cache-Control headers dynamically

---

### Step 4: Stack Outputs (Lines 87-103)

```go
awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
    Value:       bucket.BucketName(),
    Description: jsii.String("S3 bucket name"),
    ExportName:  jsii.String(props.WebsiteName + "-BucketName"),
})
// ... (similar for DistributionDomain, WebsiteURL)
```

**What's Happening:**
1. Creates CloudFormation **Stack Outputs** visible in the AWS Console
2. **ExportName:** Allows cross-stack references (e.g., for CI/CD pipelines)
3. Outputs are displayed after `cdk deploy` completes

**Output Usage Examples:**

```bash
# Get website URL programmatically
aws cloudformation describe-stacks \
  --stack-name dev-static-website \
  --query 'Stacks[0].Outputs[?OutputKey==`WebsiteURL`].OutputValue' \
  --output text

# Use in another CDK stack
const bucketName = Fn.importValue('my-website-dev-BucketName');
```

**Well-Architected Alignment:**
- **Operational Excellence:** Provides deployment feedback and cross-stack integration
- **Reliability:** Exports enable dependent stack orchestration

---

## 🔐 Security Deep Dive

### Origin Access Control (OAC) Flow

```
User Request → CloudFront Edge → OAC Signature → S3 Bucket Policy Check → Content
```

**Bucket Policy (Auto-generated):**
```json
{
  "Effect": "Allow",
  "Principal": {
    "Service": "cloudfront.amazonaws.com"
  },
  "Action": "s3:GetObject",
  "Resource": "arn:aws:s3:::my-bucket/*",
  "Condition": {
    "StringEquals": {
      "AWS:SourceArn": "arn:aws:cloudfront::123456789012:distribution/EDFDVBD6EXAMPLE"
    }
  }
}
```

**Security Benefits:**
- ✅ S3 bucket is **never publicly accessible**
- ✅ Only the specific CloudFront distribution can read objects
- ✅ Prevents bucket enumeration and direct S3 access
- ✅ Supports S3 server-side encryption (SSE-S3, SSE-KMS)

### Comparison: OAC vs. OAI (Legacy)

| Feature | OAC (Current) | OAI (Deprecated) |
|---------|---------------|------------------|
| IAM Principal | `Service: cloudfront.amazonaws.com` | Specific OAI identity |
| S3 Encryption | Supports SSE-KMS | SSE-S3 only |
| Recommended | ✅ Yes | ❌ No (maintenance mode) |
| CloudFormation | Native support | Legacy |

**Migration Note:** AWS recommends migrating from OAI to OAC for new deployments.

---

## 💰 Cost Analysis

### Monthly Cost Estimate (Example: 10,000 page views/month)

**Assumptions:**
- Average page size: 500 KB (HTML + assets)
- Geographic distribution: US/Europe (PriceClass 100)
- Cache hit ratio: 85%

| Service | Usage | Monthly Cost |
|---------|-------|--------------|
| **S3 Storage** | 1 GB | $0.023 |
| **S3 Requests** | 1,500 GET (15% cache misses) | $0.001 |
| **CloudFront Requests** | 10,000 HTTPS requests | $0.01 |
| **CloudFront Data Transfer** | 5 GB (10k × 500 KB) | $0.425 |
| **Lambda (Deployment)** | 1 invocation/deploy | ~$0.00 |
| **Total** | | **~$0.46/month** |

**Cost Optimization Tips:**
1. **Enable CloudFront compression** (reduces data transfer by ~70%)
2. **Use S3 Intelligent-Tiering** for infrequently accessed assets
3. **Set longer TTLs** for static assets (reduce S3 requests)
4. **Price Class 100** covers 90% of global users at lower cost

**Scaling:**
- At 1M page views/month: ~$25-30/month (mostly data transfer)
- At 10M page views/month: ~$200-250/month

---

## 🚀 Deployment Workflow

### Prerequisites
1. AWS account with configured credentials
2. Go 1.23+ and AWS CDK CLI installed
3. Content ready in `SourcePath` directory with `index.html`

### Initial Deployment

```bash
# 1. Bootstrap CDK (first time only)
cdk bootstrap

# 2. Synthesize CloudFormation template
cdk synth DevStaticWebsite

# 3. Deploy stack
cdk deploy DevStaticWebsite --require-approval never

# 4. Access outputs
aws cloudformation describe-stacks \
  --stack-name dev-static-website \
  --query 'Stacks[0].Outputs'
```

### Continuous Deployment

```bash
# Update content in stacks/website/dist/
# Re-deploy (only content changes)
cdk deploy DevStaticWebsite

# CDK will:
# 1. Detect changed asset hash
# 2. Upload new files to S3
# 3. Invalidate CloudFront cache
# 4. Remove old files (prune)
```

### Rollback Strategy

```bash
# Option 1: Redeploy previous version
git checkout <previous-commit>
cdk deploy DevStaticWebsite

# Option 2: Use S3 versioning (if enabled)
aws s3api list-object-versions --bucket my-bucket
aws s3api copy-object --copy-source my-bucket/index.html?versionId=xxx
```

---

## 🎯 Use Cases & Variations

### 1. **Multi-Environment Setup**

```go
// main.go
for _, env := range []string{"dev", "staging", "prod"} {
    stacks.NewStaticWebsiteStack(app, fmt.Sprintf("%s-website", env),
        &stacks.StaticWebsiteStackProps{
            BucketName:  fmt.Sprintf("%s-website-%s", env, account),
            WebsiteName: fmt.Sprintf("my-app-%s", env),
            SourcePath:  "dist",
            // Prod-specific configs
            EnableWAF: env == "prod",
        })
}
```

### 2. **Custom Domain Setup**

**Requirements:**
- ACM certificate in `us-east-1` (CloudFront requirement)
- DNS provider (Route 53, Cloudflare, etc.)

```go
stacks.NewStaticWebsiteStack(app, "ProdWebsite", &stacks.StaticWebsiteStackProps{
    DomainNames:    []string{"www.example.com", "example.com"},
    CertificateArn: "arn:aws:acm:us-east-1:123456789012:certificate/abc123",
    // ... other props
})
```

**Post-Deployment:**
1. Get CloudFront domain from outputs
2. Create DNS CNAME record: `www.example.com → d123abc.cloudfront.net`
3. Wait for DNS propagation (5-30 minutes)

### 3. **WAF Protection (Production)**

```go
// Create WAF Web ACL first (separate stack or manual)
webAclArn := "arn:aws:wafv2:us-east-1:123456789012:global/webacl/my-acl/abc123"

stacks.NewStaticWebsiteStack(app, "ProdWebsite", &stacks.StaticWebsiteStackProps{
    EnableWAF:  true,
    WebAclArn:  webAclArn,
    // ... other props
})
```

**Recommended WAF Rules:**
- AWS Managed Rules: Core Rule Set (CRS)
- Rate limiting: 2,000 requests per 5 minutes per IP
- Geo-blocking (if applicable)
- SQL injection protection

---

## ⚠️ Production Considerations

### 1. **Modify for Production Environments**

```go
// Development (current implementation)
s3Props.RemovalPolicy = "destroy"       // ❌ Dangerous in prod
s3Props.AutoDeleteObjects = true        // ❌ Data loss risk

// Production (recommended)
s3Props.RemovalPolicy = "retain"        // ✅ Protects data
s3Props.AutoDeleteObjects = false       // ✅ Manual deletion required
```

### 2. **Enable Access Logging**

```go
distribution := cloudfront.NewDistributionV2(stack, "WebsiteDistribution",
    cloudfront.CloudFrontPropertiesV2{
        // ... existing props
        EnableAccessLogging: true,  // ✅ Enable for compliance/analytics
    })
```

**Log Analysis Use Cases:**
- Traffic patterns and geographic distribution
- Security incident investigation
- Cost attribution by content type
- User behavior analytics (page views, referrers)

### 3. **Monitoring & Alarms**

```go
// Add to stack (not currently implemented)
import "github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"

// CloudFront metrics
errorRateAlarm := awscloudwatch.NewAlarm(stack, jsii.String("HighErrorRate"),
    &awscloudwatch.AlarmProps{
        Metric: distribution.MetricErrorRate(),
        Threshold: jsii.Number(5),  // 5% error rate
        EvaluationPeriods: jsii.Number(2),
    })
```

**Key Metrics to Monitor:**
- `4xxErrorRate` / `5xxErrorRate`
- `BytesDownloaded` (bandwidth usage)
- `Requests` (traffic volume)
- S3 `4xxErrors` (missing objects)

### 4. **Disaster Recovery**

**Backup Strategy:**
- Enable S3 versioning (already enabled via `GetCloudFrontOriginProperties`)
- Cross-region replication (for critical sites)
- Regular CloudFormation template backups

**RTO/RPO:**
- **RTO (Recovery Time Objective):** ~10 minutes (redeploy from Git + CloudFront propagation)
- **RPO (Recovery Point Objective):** 0 (source files in Git, S3 versioned)

---

## 🧪 Testing Strategy

### 1. **Pre-Deployment Validation**

```bash
# Validate CloudFormation template
cdk synth DevStaticWebsite > template.yaml
cfn-lint template.yaml

# Check for security issues
cdk synth | cfn_nag_scan --input-path /dev/stdin
```

### 2. **Post-Deployment Tests**

```bash
# Check website accessibility
curl -I https://$(aws cloudformation describe-stacks \
  --stack-name dev-static-website \
  --query 'Stacks[0].Outputs[?OutputKey==`DistributionDomain`].OutputValue' \
  --output text)

# Verify security headers
curl -I https://d123abc.cloudfront.net | grep -E 'X-Frame-Options|Strict-Transport-Security'

# Test cache behavior
curl -I https://d123abc.cloudfront.net/assets/logo.png | grep 'X-Cache'
```

### 3. **Load Testing**

```bash
# Using Artillery
artillery quick --count 100 --num 10 https://your-cloudfront-domain.net

# Monitor CloudFront cache hit ratio
aws cloudwatch get-metric-statistics \
  --namespace AWS/CloudFront \
  --metric-name CacheHitRate \
  --dimensions Name=DistributionId,Value=EDFDVBD6EXAMPLE \
  --start-time 2025-01-01T00:00:00Z \
  --end-time 2025-01-01T01:00:00Z \
  --period 300 \
  --statistics Average
```

---

## 🔄 Extension Points

### Add Custom CloudFront Functions (Edge Logic)

```go
// Create CloudFront Function for security headers
function := awscloudfront.NewFunction(stack, jsii.String("SecurityHeadersFunction"),
    &awscloudfront.FunctionProps{
        Code: awscloudfront.FunctionCode_FromInline(jsii.String(`
            function handler(event) {
                var response = event.response;
                response.headers['x-custom-header'] = {value: 'my-value'};
                return response;
            }
        `)),
    })

// Associate with distribution (requires modifying cloudfront_s3.go)
```

### Integrate with CI/CD (GitHub Actions Example)

```yaml
# .github/workflows/deploy.yml
name: Deploy Website
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Deploy CDK Stack
        run: |
          npm install -g aws-cdk
          cdk deploy DevStaticWebsite --require-approval never
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

---

## 📚 References

### AWS Documentation
- [CloudFront Developer Guide](https://docs.aws.amazon.com/cloudfront/)
- [Origin Access Control (OAC)](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html)
- [S3 Static Website Hosting Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/WebsiteHosting.html)
- [AWS CDK Go Documentation](https://docs.aws.amazon.com/cdk/api/v2/go/)

### Related Constructs
- `constructs/S3/s3.go` - Custom S3 bucket construct with pre-configured patterns
- `constructs/Cloudfront/cloudfront_factory.go` - Factory pattern for CloudFront distributions
- `constructs/Cloudfront/cloudfront_s3.go` - S3 origin strategy implementation

### Well-Architected Framework
- [Static Website Hosting Best Practices](https://aws.amazon.com/architecture/reference-architecture-diagrams/)
- [CloudFront Security Best Practices](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/security-best-practices.html)

---

## 🎓 Key Takeaways

1. **This stack implements the recommended AWS pattern** for static websites: CloudFront + private S3 with OAC
2. **Custom constructs** (`s3.GetCloudFrontOriginProperties`, `cloudfront.NewDistributionV2`) promote consistency and reduce boilerplate
3. **Factory + Strategy pattern** allows easy extension for different origin types (API, ALB)
4. **Automated deployment** via `BucketDeployment` handles S3 uploads and cache invalidation
5. **Production readiness requires modifications**: Retain policy, access logging, monitoring, WAF
6. **Cost-effective**: Aggressive caching and compression minimize bandwidth costs
7. **Secure by default**: Private bucket, HTTPS enforcement, security headers

**Next Steps:**
- Review `constructs/Cloudfront/cloudfront_s3.go` for CloudFront configuration details
- Implement monitoring and alarms for production deployments
- Add custom domain support via Route 53 integration
- Integrate with CI/CD pipeline for automated deployments
