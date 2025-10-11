# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Go-based AWS CDK library** (`cdk-library`) that provides reusable, high-level constructs for building AWS infrastructure. The library implements best practices for security, cost optimization, and performance.

**Primary Language:** Go (Go 1.23+)
**CDK Version:** AWS CDK v2 for Go
**Purpose:** Reusable infrastructure constructs following AWS best practices

## Available Constructs

All constructs follow **Factory + Strategy pattern** for maximum extensibility and maintainability.

### S3 Construct (`constructs/S3/`)
**Architecture:** Factory + Strategy pattern (6 specialized strategies)

**Entry Point:** `NewSimpleStorageServiceFactory(scope, id, props)`

**Available Strategies:**
- `BucketTypeCloudfrontOAC` → CloudFront origin with OAC (KMS, TLS 1.2, versioning)
- `BucketTypeDataLake` → Big data analytics (multi-tier lifecycle: raw-data, processed-data)
- `BucketTypeBackup` → Disaster recovery (Object Lock GOVERNANCE, 10-year retention)
- `BucketTypeMediaStreaming` → Video/audio streaming (S3_MANAGED, CORS, low latency)
- `BucketTypeEnterprise` → Maximum security (Object Lock COMPLIANCE, TLS 1.3, 7-year immutable)
- `BucketTypeDevelopment` → Dev/test (auto-delete, 30-day expiration, minimal cost)

**Files:**
- `simple_storage_service_factory.go` - Factory entry point
- `simple_storage_service_contract.go` - Strategy interface
- `simple_storage_service_cloudfront_origin.go` - CloudFront origin strategy
- `simple_storage_service_data_lake.go` - Data lake strategy
- `simple_storage_service_backup.go` - Backup strategy
- `simple_storage_service_media_streaming.go` - Media streaming strategy
- `simple_storage_service_enterprise.go` - Enterprise compliance strategy
- `simple_storage_service_development.go` - Development strategy

**Usage Example:**
```go
bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeCloudfrontOAC,
        BucketName: "my-website-prod",
    })
```

### CloudFront Construct (`constructs/Cloudfront/`)
**Architecture:** Factory + Strategy pattern (1 strategy implemented, 3 planned)

**Entry Point:** `NewDistributionV2(scope, id, props)`

**Available Strategies:**
- `OriginTypeS3` → S3 origin with OAC (fully implemented)

**Planned Strategies:**
- `OriginTypeAPIGateway` → API Gateway integration
- `OriginTypeALB` → Application Load Balancer origin
- `OriginTypeCustomHTTP` → Custom HTTP origins

**Files:**
- `cloudfront_factory.go` - Factory entry point
- `cloudfront_contract.go` - Strategy interface
- `cloudfront_s3.go` - S3 origin implementation with OAC

**Features:** Cache policies, SSL/TLS (ACM), WAF integration, SPA support, security headers

**Usage Example:**
```go
distribution := cloudfront.NewDistributionV2(stack, "CDN",
    cloudfront.CloudFrontPropertiesV2{
        OriginType: cloudfront.OriginTypeS3,
        S3Bucket:   bucket,
        WebACLId:   webacl.Arn(),
        AutoConfigureS3BucketPolicy: true,
    })
```

### WAF Construct (`constructs/WAF/`)
**Architecture:** Factory + Strategy pattern (3 specialized strategies)

**Entry Point:** `NewWebApplicationFirewallV2(scope, id, props)`

**Available Strategies:**
- `WafTypeWebApplication` → Web apps (OWASP Top 10, rate limit 2000 req/5min, XSS, SQL injection)
- `WafTypeAPI` → APIs (rate limit 10000 req/5min, token validation, bot protection)
- `WafTypeOWASP` → OWASP Core Rule Set (comprehensive vulnerability protection)

**Files:**
- `web_application_firewall_factory.go` - Factory entry point
- `web_application_firewall_contract.go` - Strategy interface
- `web_application_firewall_for_web_application.go` - Web application strategy
- `web_application_firewall_for_api.go` - API protection strategy
- `web_application_firewall_for_owasp.go` - OWASP strategy

**Features:** AWS Managed Rules, custom rate limiting, geo-blocking, CloudWatch metrics

**Usage Example:**
```go
webacl := waf.NewWebApplicationFirewallV2(stack, "WAF",
    waf.WebApplicationFirewallFactoryProps{
        WafType: waf.WafTypeWebApplication,
        Scope:   waf.WafScopeCloudfront,
        Name:    "my-app-waf",
    })
```

## Common Commands

### Development Workflow

**Bootstrap CDK (first time only):**
```bash
cdk bootstrap --profile <aws-profile> --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
```

**Synthesize CloudFormation templates:**
```bash
cdk synth
# Or for specific stack:
cdk synth DevStaticWebsite
```

**Deploy infrastructure:**
```bash
# Using deploy script (recommended):
./deploy.sh

# Or manually:
export CDK_DEFAULT_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
export CDK_DEFAULT_REGION=$(aws configure get region)
cdk deploy DevStaticWebsite --require-approval never
```

**Destroy stack:**
```bash
cdk destroy DevStaticWebsite
```

**Clean CDK output:**
```bash
./clear.sh
```

### Go Commands

**Update dependencies:**
```bash
go mod tidy
go mod download
```

**Run tests:**
```bash
go test -v ./...
```

**Build/validate Go code:**
```bash
go build
```

## Architecture & Patterns

### Factory + Strategy Pattern (All Constructs)
All constructs in this library use **Factory + Strategy pattern** for consistency and extensibility:

**Pattern Structure:**
```
User Code → Factory (Type Selection) → Strategy Interface → Concrete Strategy → AWS Resources
```

**Benefits:**
- ✅ **Open/Closed Principle**: Add new strategies without modifying existing code
- ✅ **Single Responsibility**: Each strategy handles one specific use case (~100-150 lines)
- ✅ **Testability**: Mock strategies for unit tests
- ✅ **Maintainability**: Clear separation of concerns
- ✅ **Type Safety**: Compile-time validation of strategy types

### Implementation Pattern (All Constructs)

**1. Factory Function** (Entry point):
```go
func NewXxxFactory(scope constructs.Construct, id string, props XxxFactoryProps) Resource {
    var strategy XxxStrategy

    switch props.Type {
    case TypeA:
        strategy = &TypeAStrategy{}
    case TypeB:
        strategy = &TypeBStrategy{}
    default:
        panic(fmt.Sprintf("Unsupported type: %s", props.Type))
    }

    return strategy.Build(scope, id, props)
}
```

**2. Strategy Interface** (Contract):
```go
type XxxStrategy interface {
    Build(scope constructs.Construct, id string, props XxxFactoryProps) Resource
}
```

**3. Concrete Strategies** (Implementations):
```go
type TypeAStrategy struct{}

func (s *TypeAStrategy) Build(scope constructs.Construct, id string, props XxxFactoryProps) Resource {
    // Specific configuration for Type A use case
    return awsxxx.NewResource(scope, jsii.String(id), &awsxxx.ResourceProps{
        // Optimized properties for Type A
    })
}
```

### Adding New Strategies

To add a new strategy to any construct:

1. **Create strategy file**: `construct_new_strategy.go`
2. **Implement interface**: `Build(scope, id, props) Resource`
3. **Add type constant**: Update type enum
4. **Register in factory**: Add case to switch statement
5. **Document**: Update README with use cases and examples

**Example - Adding new S3 strategy:**
```go
// 1. Add constant
const BucketTypeReplication BucketType = "REPLICATION"

// 2. Create simple_storage_service_replication.go
type SimpleStorageServiceReplicationStrategy struct{}

func (s *SimpleStorageServiceReplicationStrategy) Build(...) awss3.Bucket {
    // Implementation
}

// 3. Register in factory
case BucketTypeReplication:
    strategy = &SimpleStorageServiceReplicationStrategy{}
```

### Stack Structure
**Location:** `stacks/website/StaticWebSite.go`

Static website stack demonstrates proper construct composition using Factory pattern:

1. **Create S3 bucket** using `NewSimpleStorageServiceFactory`:
   ```go
   bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
       s3.SimpleStorageServiceFactoryProps{
           BucketType: s3.BucketTypeCloudfrontOAC,
           BucketName: "my-website-prod",
       })
   ```

2. **Create WAF** using `NewWebApplicationFirewallV2`:
   ```go
   webacl := waf.NewWebApplicationFirewallV2(stack, "WAF",
       waf.WebApplicationFirewallFactoryProps{
           WafType: waf.WafTypeWebApplication,
           Scope:   waf.WafScopeCloudfront,
           Name:    "website-waf",
       })
   ```

3. **Create CloudFront** using `NewDistributionV2`:
   ```go
   distribution := cloudfront.NewDistributionV2(stack, "CDN",
       cloudfront.CloudFrontPropertiesV2{
           OriginType: cloudfront.OriginTypeS3,
           S3Bucket:   bucket,
           WebACLId:   webacl.Arn(),
           AutoConfigureS3BucketPolicy: true,
       })
   ```

4. **Deploy content** with `BucketDeployment`
5. **Export outputs** (bucket name, CloudFront domain, URL)

**Key Pattern:** All constructs use Factory pattern with type-based strategy selection.

### CDK Application Entry Point
**Location:** `main.go`

- Reads AWS account/region from environment (`CDK_DEFAULT_ACCOUNT`, `CDK_DEFAULT_REGION`)
- Creates stacks with consistent naming and tagging
- Currently deploys: `DevStaticWebsite` stack

## Configuration & Environment

**Required Environment Variables:**
- `CDK_DEFAULT_ACCOUNT` - AWS account ID
- `CDK_DEFAULT_REGION` - AWS region

**Set automatically by `deploy.sh`:**
```bash
CDK_DEFAULT_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
CDK_DEFAULT_REGION=$(aws configure get region)
```

**CDK Configuration (`cdk.json`):**
- App command: `go mod download && go run main.go`
- Watch mode configured for hot reload
- Feature flags: All modern CDK best practices enabled

## Important Development Notes

### Factory + Strategy Best Practices
- **DO:** Use factory functions for all construct creation (`NewSimpleStorageServiceFactory`, `NewDistributionV2`, `NewWebApplicationFirewallV2`)
- **DO:** Select appropriate strategy type based on use case (e.g., `BucketTypeCloudfrontOAC` for websites)
- **DON'T:** Create AWS resources directly - always use factory pattern
- **WHY:** Type safety, consistent configuration, maintainable code

### S3 + CloudFront Best Practices
- **DO:** Use `BucketTypeCloudfrontOAC` strategy for static websites (private bucket + OAC)
- **DO:** Enable `AutoConfigureS3BucketPolicy: true` in CloudFront props
- **DON'T:** Use S3 website hosting directly (no security, no HTTPS)
- **WHY:** OAC (Origin Access Control) provides better security than legacy OAI and S3 website hosting

### WAF Integration
- **DO:** Create WAF WebACL before CloudFront distribution
- **DO:** Use `WafScopeCloudfront` for CloudFront, `WafScopeRegional` for ALB/API Gateway
- **DO:** Pass `webacl.Arn()` to CloudFront `WebACLId` property
- **WHY:** WAF provides DDoS protection, rate limiting, and OWASP Top 10 defense

### Static Website Deployment
**Content location:** `stacks/website/dist/`
- Must contain `index.html`
- Assets in `assets/css/`, `assets/js/`, `assets/images/`
- Deployment handled by `BucketDeployment` construct

**SPA Support:** CloudFront error responses redirect 403/404 to `/index.html` (see `cloudfront_s3.go:86-99`)

### Testing
Tests are currently commented out in `cdk-library_test.go`. When writing tests:
- Use `assertions.Template_FromStack()` for CloudFormation assertions
- Test construct properties, not AWS API calls
- Example pattern available in commented code

### Naming Conventions
- **Constructs:** PascalCase packages (`S3`, `Cloudfront`, `WAF`)
- **Files:** snake_case with pattern `construct_strategy.go`
  - Factory: `construct_factory.go`
  - Contract: `construct_contract.go`
  - Strategies: `construct_specific_strategy.go`
- **Constants:** PascalCase with type prefix (`BucketTypeCloudfrontOAC`, `WafTypeWebApplication`)
- **Stacks:** PascalCase with "Stack" suffix (`StaticWebsiteStack`)
- **Resources:** Descriptive IDs in construct calls (`"WebsiteBucket"`, `"WebsiteDistribution"`)

### Current Implementation Status

**Completed (10 strategies across 3 constructs):**
- ✅ S3: 6 strategies (CloudFront Origin, Data Lake, Backup, Media, Enterprise, Development)
- ✅ CloudFront: 1 strategy (S3 origin with OAC)
- ✅ WAF: 3 strategies (Web Application, API, OWASP)

**Planned:**
- ⏳ CloudFront: 3 additional strategies (API Gateway, ALB, Custom HTTP)
- ⏳ Lambda: Factory + Strategy implementation
- ⏳ API Gateway: Factory + Strategy implementation
- ⏳ DynamoDB: Factory + Strategy implementation

## Deployment Script (`deploy.sh`)

The deployment script automates the full workflow:
1. Auto-configures AWS credentials (no hardcoding)
2. Validates CDK bootstrap status, bootstraps if needed
3. Prepares content directories
4. Synthesizes and deploys stack
5. Verifies S3 content and outputs URLs

**Key features:**
- Waits for CloudFormation stack completion
- Validates bucket has content (common source of errors)
- Outputs website URL and diagnostics

## Common Gotchas

1. **Empty S3 Bucket:** If CloudFront returns 404, verify:
   - `SourcePath` in stack props matches actual content location
   - `BucketDeployment` completed successfully
   - Use `aws s3 ls s3://<bucket-name> --recursive` to verify

2. **CloudFront Propagation:** Distribution changes take 5-10 minutes to propagate globally

3. **Bootstrap Required:** CDK requires bootstrapping once per account/region before deploying stacks that use assets

4. **Go Module Path:** Import paths use `cdk-library/constructs/...` (module name from `go.mod:1`)

5. **JSII Pointers:** AWS CDK Go requires `jsii.String()`, `jsii.Bool()`, etc. for all pointer parameters
