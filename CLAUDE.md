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

### GuardDuty Construct (`constructs/GuardDuty/`)
**Architecture:** Factory + Strategy pattern (4 specialized strategies)

**Entry Point:** `NewGuardDutyDetector(scope, id, props)`

**Available Strategies:**
- `GuardDutyTypeBasic` → Foundational detection (CloudTrail, VPC Flow, DNS) - Cost: ~$4-8/month
- `GuardDutyTypeComprehensive` → Full protection (S3, EKS, Malware, RDS, Lambda, Runtime) - Cost: ~$30-100/month
- `GuardDutyTypeCustom` → Granular control over individual features - Cost: Variable
- `GuardDutyTypeDataProtection` → S3 data protection with malware scanning - Cost: ~$15-25/month

**Files:**
- `guardduty_factory.go` - Factory entry point
- `guardduty_contract.go` - Strategy interface
- `guardduty_basic.go` - Basic foundational strategy
- `guardduty_comprehensive.go` - Comprehensive protection strategy
- `guardduty_custom.go` - Custom configuration strategy
- `guardduty_data_protection.go` - Data protection strategy

**Features:** Threat intelligence, ML-based anomaly detection, multi-stage attack correlation, runtime monitoring

**Usage Example:**
```go
// Basic (Development/Testing)
detector := guardduty.NewGuardDutyDetector(stack, "BasicDetector",
    guardduty.GuardDutyFactoryProps{
        DetectorType: guardduty.GuardDutyTypeBasic,
    })

// Comprehensive (Production)
detector := guardduty.NewGuardDutyDetector(stack, "ProdDetector",
    guardduty.GuardDutyFactoryProps{
        DetectorType: guardduty.GuardDutyTypeComprehensive,
        FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
    })

// Custom (S3 + EKS only)
detector := guardduty.NewGuardDutyDetector(stack, "CustomDetector",
    guardduty.GuardDutyFactoryProps{
        DetectorType: guardduty.GuardDutyTypeCustom,
        EnableS3Protection: jsii.Bool(true),
        EnableEKSProtection: jsii.Bool(true),
        EnableEKSRuntimeMonitoring: jsii.Bool(true),
    })

// Data Protection (S3 focused)
detector := guardduty.NewGuardDutyDetector(stack, "DataProtection",
    guardduty.GuardDutyFactoryProps{
        DetectorType: guardduty.GuardDutyTypeDataProtection,
        FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
    })
```

### Lambda Construct (`constructs/Lambda/`)
**Architecture:** Factory + Strategy pattern (1 strategy implemented)

**Entry Point:** `NewGoLambdaFunction(scope, id, props)`

**Available Strategies:**
- `GoLambda` → Go runtime with ARM64 support (custom runtime `provided.al2`)

**Files:**
- `lambda_factory.go` - Factory entry point
- `lambda_contract.go` - Strategy interface
- `go_lambda.go` - Go Lambda ARM64 implementation

**Features:** ARM64 (Graviton2), environment variables, IAM permissions, Dead Letter Queue, retry logic

**Usage Example:**
```go
lambdaFn := lambda.NewGoLambdaFunction(stack, "WebhookNotifier",
    lambda.GoLambdaFactoryProps{
        FunctionName: "s3-to-sftp-webhook",
        CodePath:     "stacks/addi/lambda/webhook-notifier",
        Handler:      "bootstrap",
        Environment: map[string]*string{
            "WEBHOOK_URL": jsii.String("https://api.example.com/webhook"),
        },
        Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
        MemorySize:   jsii.Number(256),
    })
```

**Compilation (Required):**
```bash
cd stacks/addi/lambda/webhook-notifier/
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc main.go
```

### EventBridge Integration
**Architecture:** Event-driven patterns (3 strategies: 1 implemented, 2 planned)

**Available Patterns:**
- `S3-to-Lambda` → S3 ObjectCreated events trigger Lambda (✅ Implemented)
- `SNS-to-Lambda` → SNS message triggers Lambda (⏳ Planned)
- `SQS-to-Lambda` → SQS message triggers Lambda (⏳ Planned)

**S3-to-Lambda Features:**
- Event pattern: `s3:ObjectCreated:*` with bucket filter
- Lambda target with retry configuration (3 attempts, exponential backoff)
- Dead Letter Queue for failed events
- EventBridge metrics and CloudWatch integration

**Usage Example:**
```go
// EventBridge Rule for S3 events
rule := awsevents.NewRule(stack, "S3EventRule", &awsevents.RuleProps{
    EventPattern: &awsevents.EventPattern{
        Source:     jsii.Strings("aws.s3"),
        DetailType: jsii.Strings("Object Created"),
        Detail: map[string]interface{}{
            "bucket": map[string]interface{}{
                "name": []string{"addi-landing-zone-dev"},
            },
        },
    },
})

// Add Lambda target with DLQ
dlq := awssqs.NewQueue(stack, "DLQ", &awssqs.QueueProps{
    QueueName: jsii.String("s3-events-dlq"),
    RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
})

rule.AddTarget(awseventstargets.NewLambdaFunction(lambdaFn, &awseventstargets.LambdaFunctionProps{
    DeadLetterQueue: dlq,
    MaxEventAge:     awscdk.Duration_Hours(jsii.Number(2)),
    RetryAttempts:   jsii.Number(3),
}))
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

#### Static Website Stack
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

#### Addi S3 to SFTP Pipeline Stack
**Location:** `stacks/addi/addi_stack_example.go`

Production-ready event-driven pipeline that automates file transfer from S3 to SFTP servers:

**Architecture Flow:**
```
Client (IAM) → S3 Bucket → EventBridge → Lambda → Backend API → SFTP Server
```

**Stack Components:**

1. **S3 Bucket** (Development Strategy):
   ```go
   bucket := s3.NewSimpleStorageServiceFactory(stack, "LandingZone",
       s3.SimpleStorageServiceFactoryProps{
           BucketType: s3.BucketTypeDevelopment,
           BucketName: "addi-landing-zone-dev",
       })
   ```

2. **Lambda Function** (Go ARM64):
   ```go
   lambdaFn := lambda.NewGoLambdaFunction(stack, "WebhookNotifier",
       lambda.GoLambdaFactoryProps{
           FunctionName: "s3-to-sftp-webhook",
           CodePath:     "stacks/addi/lambda/webhook-notifier",
           Handler:      "bootstrap",
           Environment: map[string]*string{
               "WEBHOOK_URL": jsii.String(webhookURL),
           },
       })
   ```

3. **EventBridge Rule** (S3 to Lambda):
   ```go
   rule := awsevents.NewRule(stack, "S3EventRule", &awsevents.RuleProps{
       EventPattern: &awsevents.EventPattern{
           Source:     jsii.Strings("aws.s3"),
           DetailType: jsii.Strings("Object Created"),
           Detail: map[string]interface{}{
               "bucket": map[string]interface{}{
                   "name": []string{*bucket.BucketName()},
               },
           },
       },
   })
   ```

4. **Dead Letter Queue** (Failed Events):
   ```go
   dlq := awssqs.NewQueue(stack, "DLQ", &awssqs.QueueProps{
       QueueName: jsii.String("s3-events-dlq"),
       RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
   })

   rule.AddTarget(awseventstargets.NewLambdaFunction(lambdaFn, &awseventstargets.LambdaFunctionProps{
       DeadLetterQueue: dlq,
       MaxEventAge:     awscdk.Duration_Hours(jsii.Number(2)),
       RetryAttempts:   jsii.Number(3),
   }))
   ```

5. **GuardDuty** (Data Protection):
   ```go
   detector := guardduty.NewGuardDutyDetector(stack, "DataProtection",
       guardduty.GuardDutyFactoryProps{
           DetectorType: guardduty.GuardDutyTypeDataProtection,
           FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
       })
   ```

**Backend API (External):**
- Go API server running in Docker
- Receives webhook from Lambda with presigned URL
- Downloads file from S3 using HTTP (no AWS credentials needed)
- Uploads to SFTP server
- Located in: `stacks/addi/backend/`

**Key Features:**
- ✅ IAM-based S3 access (not public)
- ✅ Presigned URLs for secure downloads (1 hour expiration)
- ✅ Event-driven architecture with automatic retries
- ✅ Dead Letter Queue for failed events
- ✅ GuardDuty malware scanning
- ✅ Comprehensive documentation with Mermaid diagrams

**Documentation:** See `stacks/addi/README.md` for complete implementation details, testing procedures, and architecture diagrams.

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
- **DO:** Use factory functions for all construct creation (`NewSimpleStorageServiceFactory`, `NewDistributionV2`, `NewWebApplicationFirewallV2`, `NewGuardDutyDetector`)
- **DO:** Select appropriate strategy type based on use case (e.g., `BucketTypeCloudfrontOAC` for websites, `GuardDutyTypeComprehensive` for production)
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

### GuardDuty Best Practices
- **DO:** Use `GuardDutyTypeBasic` for dev/test environments to minimize cost
- **DO:** Use `GuardDutyTypeComprehensive` for production workloads requiring maximum security
- **DO:** Use `GuardDutyTypeCustom` for phased rollout or specific compliance requirements
- **DO:** Use `GuardDutyTypeDataProtection` for S3-focused workloads with sensitive data
- **DO:** Set `FindingPublishingFrequency: "FIFTEEN_MINUTES"` for production (rapid incident response)
- **DO:** Integrate findings with EventBridge for automated remediation workflows
- **DON'T:** Disable GuardDuty in production - it's a critical security baseline
- **WHY:** GuardDuty provides continuous threat detection using ML and threat intelligence without requiring agents

### Lambda Best Practices
- **DO:** Compile Go Lambda as `bootstrap` executable for custom runtime `provided.al2`
- **DO:** Use ARM64 architecture (Graviton2) for ~20% cost savings
- **DO:** Use compilation command: `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc main.go`
- **DO:** Set appropriate timeout based on operation (30s for network operations is reasonable)
- **DO:** Use environment variables for configuration (webhook URLs, API endpoints)
- **DO:** Implement Dead Letter Queue for failed invocations
- **DO:** Configure retry attempts (3 with exponential backoff is standard)
- **DO:** Grant minimal IAM permissions (S3 read-only + presigned URL generation)
- **DON'T:** Use x86_64 unless required by dependencies (ARM64 is cheaper)
- **DON'T:** Forget to recompile after code changes
- **DON'T:** Use AWS SDK in external services - use presigned URLs instead
- **WHY:** Presigned URLs avoid credential management outside AWS, reducing security risk

### EventBridge Best Practices
- **DO:** Use specific event patterns to filter events (bucket name, prefix, etc.)
- **DO:** Configure Dead Letter Queue for failed Lambda invocations
- **DO:** Set `MaxEventAge` to prevent processing stale events (2 hours is reasonable)
- **DO:** Use retry configuration with exponential backoff (3 attempts: 60s, 120s, 180s)
- **DO:** Monitor DLQ for failed events and investigate root causes
- **DO:** Use EventBridge metrics to track rule invocations and failures
- **DON'T:** Process every S3 event - filter by bucket/prefix to reduce costs
- **DON'T:** Ignore DLQ messages - they indicate systematic failures
- **WHY:** Proper retry logic ensures reliability while preventing infinite loops

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

**Completed (18 strategies across 6 constructs):**
- ✅ S3: 6 strategies (CloudFront Origin, Data Lake, Backup, Media, Enterprise, Development)
- ✅ CloudFront: 1 strategy (S3 origin with OAC)
- ✅ WAF: 3 strategies (Web Application, API, OWASP)
- ✅ GuardDuty: 4 strategies (Basic, Comprehensive, Custom, Data Protection)
- ✅ Lambda: 1 strategy (Go Lambda with ARM64 support)
- ✅ EventBridge: 3 strategies (S3-to-Lambda, SNS-to-Lambda, SQS-to-Lambda)

**Planned:**
- ⏳ CloudFront: 3 additional strategies (API Gateway, ALB, Custom HTTP)
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

6. **Lambda Runtime.InvalidEntrypoint:** If Lambda fails with `Couldn't find valid bootstrap(s)`:
   - Verify executable is named `bootstrap` (not `main`)
   - Verify compilation uses ARM64: `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc`
   - Check Lambda architecture is set to ARM64 in construct
   - File: `constructs/Lambda/go_lambda.go:57`

7. **S3 EnforceSSL Error:** If deployment fails with `'enforceSSL' must be enabled for 'minimumTLSVersion'`:
   - `EnforceSSL` must be `true` when using `MinimumTLSVersion`
   - Fixed in: `constructs/S3/simple_storage_service_development.go:44`

8. **Backend AWS Credentials Error:** If backend shows `no EC2 IMDS role found`:
   - Backend should NOT use AWS SDK directly
   - Use presigned URLs from Lambda webhook payload
   - Use HTTP client (`http.DefaultClient.Do()`) instead of `s3.GetObject()`
   - File: `stacks/addi/backend/api/internal/services/s3_service.go:42-60`

9. **Object Lock Bucket Cannot Be Deleted:** Buckets with Object Lock COMPLIANCE cannot be deleted:
   - COMPLIANCE mode is immutable (cannot be bypassed)
   - Apply lifecycle policy to move to Glacier (reduce cost to near-zero)
   - Document as "frozen infrastructure" if needed
   - Consider using Development bucket type for testing

10. **Docker Changes Not Applied:** If backend code changes don't take effect:
    - Stop and remove specific container: `docker compose stop api && docker compose rm -f api`
    - Rebuild with no cache: `docker compose build --no-cache api`
    - Don't stop all services or ngrok URL will change
    - Restart: `docker compose up -d api`

## Troubleshooting Addi Pipeline

### Testing End-to-End Flow

**1. Verify S3 Bucket:**
```bash
aws s3 ls s3://addi-landing-zone-dev/uploads/ --recursive
```

**2. Check Lambda Logs:**
```bash
aws logs tail /aws/lambda/s3-to-sftp-webhook --follow
```

**3. Verify EventBridge Rule:**
```bash
aws events describe-rule --name S3EventRule
aws events list-targets-by-rule --rule S3EventRule
```

**4. Test Upload:**
```bash
# Upload test file
aws s3 cp test.csv s3://addi-landing-zone-dev/uploads/

# Verify SFTP received file
docker compose exec sftp ls -lh /home/uploader/2025/10/15/
```

**5. Check Dead Letter Queue:**
```bash
aws sqs receive-message --queue-url https://sqs.us-east-1.amazonaws.com/.../s3-events-dlq
```

### Common Pipeline Issues

**Lambda not triggered:**
- Check EventBridge rule is enabled
- Verify event pattern matches bucket name
- Check Lambda has permissions to be invoked by EventBridge

**Backend not receiving webhook:**
- Verify Lambda environment variable `WEBHOOK_URL` is correct
- Check ngrok tunnel is active (if using ngrok)
- Verify backend API is running: `docker compose ps`
- Check backend logs: `docker compose logs -f api`

**File not in SFTP:**
- Check backend logs for SFTP connection errors
- Verify SFTP credentials in docker-compose.yml
- Test SFTP connection: `sftp -P 2222 uploader@localhost`
- Check file permissions in SFTP container

**Presigned URL expired:**
- URLs expire after 1 hour by default
- Check Lambda execution time vs webhook processing time
- Consider increasing expiration if backend has delays

### IAM User Setup for S3 Upload

Clients need IAM credentials to upload to S3:

**1. Create IAM User:**
```bash
aws iam create-user --user-name addi-s3-uploader
```

**2. Create Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3:PutObject",
      "s3:PutObjectAcl"
    ],
    "Resource": "arn:aws:s3:::addi-landing-zone-dev/uploads/*"
  }]
}
```

**3. Attach Policy:**
```bash
aws iam put-user-policy --user-name addi-s3-uploader \
  --policy-name AddiS3UploadPolicy \
  --policy-document file://policy.json
```

**4. Generate Access Keys:**
```bash
aws iam create-access-key --user-name addi-s3-uploader
```

**5. Client Upload:**
```bash
export AWS_ACCESS_KEY_ID="AKIAxxxxx"
export AWS_SECRET_ACCESS_KEY="xxxxx"
aws s3 cp file.csv s3://addi-landing-zone-dev/uploads/
```
