# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Go-based AWS CDK library** (`cdk-library`) that provides reusable, high-level constructs for building AWS infrastructure. The library implements best practices for security, cost optimization, and performance.

**Primary Language:** Go (Go 1.23+)
**CDK Version:** AWS CDK v2 for Go
**Purpose:** Reusable infrastructure constructs following AWS best practices

## Available Constructs

### S3 Construct (`constructs/S3/s3.go`)
Comprehensive S3 bucket construct with pre-configured templates for different use cases:
- `GetDefaultProperties()` - Secure defaults
- `GetEnterpriseDataProperties()` - Financial/PII data with compliance (KMS, Object Lock, 7-year retention)
- `GetCloudFrontOriginProperties()` - **Recommended for static websites** (private bucket with OAC)
- `GetDataLakeProperties()` - Analytics workloads with lifecycle policies
- `GetBackupProperties()` - Backup/DR with retention and replication
- `GetMediaStreamingProperties()` - Media/CDN origin
- `GetDevelopmentProperties()` - Dev/test environments with auto-cleanup

### CloudFront Construct (`constructs/Cloudfront/`)
**Architecture:** Factory + Strategy pattern
- `cloudfront_factory.go` - Entry point (`NewDistributionV2`)
- `cloudfront_contract.go` - Strategy interface
- `cloudfront_s3.go` - S3 origin implementation with OAC (Origin Access Control)
- Supports: S3, S3 Website, HTTP, ALB origins (only S3 fully implemented)
- Features: Cache policies, SSL/TLS, WAF, error responses for SPAs

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

### Factory + Strategy Pattern (CloudFront)
The CloudFront construct uses a **Factory pattern** to select the appropriate **Strategy** based on origin type:

1. **Factory** (`NewDistributionV2` in `cloudfront_factory.go`):
   - Takes `CloudFrontPropertiesV2` with `OriginType` enum
   - Selects appropriate strategy (S3, API, ALB)
   - Returns configured CloudFront distribution

2. **Strategy Interface** (`cloudfront_contract.go`):
   ```go
   type CloudFrontStrategy interface {
       Build(scope, id, props) awscloudfront.Distribution
   }
   ```

3. **Concrete Strategies**:
   - `S3CloudFrontStrategy` - Implements S3 origin with OAC

**When adding new origin types:**
- Create new strategy file (e.g., `cloudfront_alb.go`)
- Implement `CloudFrontStrategy` interface
- Add case to factory's switch statement
- Update `OriginType` enum

### Stack Structure
**Location:** `stacks/website/StaticWebSite.go`

Static website stack demonstrates proper construct composition:
1. Creates S3 bucket using `s3.GetCloudFrontOriginProperties()`
2. Creates CloudFront distribution via factory
3. Deploys content with `BucketDeployment`
4. Exports outputs (bucket name, CloudFront domain, URL)

**Key pattern:** Use pre-configured property functions (e.g., `GetCloudFrontOriginProperties()`) as starting points, then customize.

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

### S3 + CloudFront Best Practices
- **DO:** Use `GetCloudFrontOriginProperties()` for static websites (private bucket + OAC)
- **DON'T:** Use S3 website hosting directly (commented out `GetStaticWebsiteProperties()`)
- **WHY:** OAC (Origin Access Control) provides better security than legacy OAI

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
- **Constructs:** PascalCase packages (`S3`, `Cloudfront`)
- **Files:** snake_case with descriptive names (`cloudfront_s3.go`, `cloudfront_factory.go`)
- **Stacks:** PascalCase with "Stack" suffix (`StaticWebsiteStack`)
- **Resources:** Descriptive IDs in construct calls (`"WebsiteBucket"`, `"WebsiteDistribution"`)

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
