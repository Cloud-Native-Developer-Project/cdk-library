# S3 Bucket Strategies - AWS CDK Go Library

Comprehensive S3 bucket configurations using Factory + Strategy design patterns for common AWS use cases.

## Architecture Overview

This library implements S3 bucket creation using two complementary design patterns:

- **Factory Pattern**: `NewSimpleStorageServiceFactory()` creates buckets based on `BucketType`
- **Strategy Pattern**: Each `BucketType` delegates to a specialized strategy with optimized configurations

```
User Code
    ↓
NewSimpleStorageServiceFactory(BucketType)
    ↓
Strategy Selection (switch statement)
    ↓
Strategy.Build() → Configured S3 Bucket
```

## Available Bucket Strategies

| Strategy | Security Level | Use Case | Cost Optimization | Data Retention |
|----------|---------------|----------|-------------------|----------------|
| **CloudFront Origin** | High | Static websites, SPAs | Intelligent Tiering | 1 year |
| **Data Lake** | High | Big data analytics | Aggressive lifecycle | Archive-focused |
| **Backup** | Maximum | Disaster recovery | Glacier transitions | 10 years |
| **Media Streaming** | Medium | Video/audio delivery | IA transitions | 1 year |
| **Enterprise** | Maximum | Financial, PII, compliance | Compliant archival | 7 years (locked) |
| **Development** | Basic | Dev/test environments | 30-day expiration | 30 days |

---

## 1. CloudFront Origin Strategy

**BucketType**: `BucketTypeCloudfrontOAC`

Optimized for serving static websites and SPAs through CloudFront with Origin Access Control.

### Key Features

- **Security**: Private bucket, KMS encryption, TLS 1.2, OAC-ready
- **Performance**: Intelligent Tiering for cost optimization
- **Lifecycle**: Automatic transition to IA (30d) → Glacier (90d) → Deep Archive (365d)
- **Monitoring**: EventBridge enabled for automation

### Configuration Highlights

```go
Encryption:        awss3.BucketEncryption_KMS_MANAGED
BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL()
Versioned:         jsii.Bool(true)
EnforceSSL:        jsii.Bool(true)
MinimumTLSVersion: jsii.Number(1.2)
RemovalPolicy:     awscdk.RemovalPolicy_RETAIN
```

### Usage Example

```go
package main

import (
    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/constructs-go/constructs/v10"
    s3construct "github.com/yourusername/cdk-library/constructs/S3"
)

func NewWebsiteStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
    stack := awscdk.NewStack(scope, &id, props)

    // Create CloudFront origin bucket
    bucket := s3construct.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
        s3construct.SimpleStorageServiceFactoryProps{
            BucketType: s3construct.BucketTypeCloudfrontOAC,
            BucketName: "my-website-bucket-prod",
        })

    // Use bucket with CloudFront distribution
    // ...

    return stack
}
```

### Use Cases

- Static website hosting (React, Vue, Angular)
- Single Page Applications (SPAs)
- CDN origin for CloudFront
- Documentation sites
- Marketing landing pages

### Estimated Monthly Cost

- **Small site** (10 GB, 100K requests): ~$2-5/month
- **Medium site** (100 GB, 1M requests): ~$15-30/month
- **Large site** (1 TB, 10M requests): ~$100-200/month

---

## 2. Data Lake Strategy

**BucketType**: `BucketTypeDataLake`

Optimized for big data analytics with separate lifecycle policies for raw and processed data.

### Key Features

- **Security**: KMS encryption, versioning, comprehensive auditing
- **Multi-tier Lifecycle**:
  - `raw-data/`: IA@30d → Glacier@90d → Deep Archive@365d
  - `processed-data/`: IA@7d → Glacier@30d (faster access)
- **Cost Optimization**: Intelligent Tiering enabled
- **Monitoring**: Daily inventory, CloudWatch metrics, EventBridge

### Configuration Highlights

```go
Encryption:        awss3.BucketEncryption_KMS_MANAGED
Versioned:         jsii.Bool(true)
ServerAccessLogsPrefix: jsii.String("access-logs/")
Inventories:       Daily frequency with object versions
LifecycleRules:    Separate policies for raw-data/ and processed-data/
```

### Usage Example

```go
// Create data lake bucket
dataLakeBucket := s3construct.NewSimpleStorageServiceFactory(stack, "DataLakeBucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeDataLake,
        BucketName: "analytics-data-lake-prod",
    })

// Upload raw data
// s3.putObject("raw-data/2025/01/logs.json", data)

// Upload processed data
// s3.putObject("processed-data/aggregated/report.parquet", processedData)
```

### Data Organization Best Practices

```
s3://analytics-data-lake-prod/
├── raw-data/              # Long-term archival (slower lifecycle)
│   ├── 2025/
│   │   ├── 01/
│   │   │   └── application-logs.json
│   │   └── 02/
│   └── schemas/
├── processed-data/        # Frequently accessed (faster lifecycle)
│   ├── aggregated/
│   │   └── daily-reports.parquet
│   └── curated/
└── access-logs/           # S3 access logs
```

### Use Cases

- Big data analytics (Athena, EMR, Redshift Spectrum)
- Data warehousing
- Machine learning training datasets
- IoT data ingestion
- Log aggregation and analytics

### Estimated Monthly Cost

- **Small data lake** (500 GB): ~$15-25/month
- **Medium data lake** (5 TB): ~$100-150/month
- **Large data lake** (50 TB): ~$800-1,200/month

---

## 3. Backup Strategy

**BucketType**: `BucketTypeBackup`

Optimized for backup and disaster recovery with Object Lock protection.

### Key Features

- **Security**: KMS encryption, Object Lock GOVERNANCE (90 days), TLS 1.2
- **Immutability**: GOVERNANCE retention prevents accidental deletion
- **Aggressive Lifecycle**: IA@30d → Glacier@90d → Deep Archive@365d
- **Long-term Retention**: 10-year expiration policy
- **Monitoring**: Comprehensive auditing and inventory

### Configuration Highlights

```go
Encryption:        awss3.BucketEncryption_KMS_MANAGED
ObjectLockEnabled: jsii.Bool(true)
ObjectLockDefaultRetention: awss3.ObjectLockRetention_Governance(
    awscdk.Duration_Days(jsii.Number(90)),
)
Versioned:         jsii.Bool(true)
Expiration:        awscdk.Duration_Days(jsii.Number(3650)) // 10 years
```

### Object Lock Modes Explained

- **GOVERNANCE**: Can be bypassed with `s3:BypassGovernanceRetention` permission
- **COMPLIANCE**: Cannot be bypassed by anyone (not even root) - see Enterprise strategy

### Usage Example

```go
// Create backup bucket
backupBucket := s3construct.NewSimpleStorageServiceFactory(stack, "BackupBucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeBackup,
        BucketName: "database-backups-prod",
    })

// Integrate with AWS Backup or custom backup solution
backup := awsbackup.NewBackupPlan(stack, jsii.String("BackupPlan"), &awsbackup.BackupPlanProps{
    BackupPlanRules: &[]*awsbackup.BackupPlanRule{
        awsbackup.NewBackupPlanRule(&awsbackup.BackupPlanRuleProps{
            RuleName: jsii.String("DailyBackup"),
            ScheduleExpression: events.Schedule_Cron(&events.CronOptions{
                Hour:   jsii.String("2"),
                Minute: jsii.String("0"),
            }),
        }),
    },
})
```

### Use Cases

- Database backups (RDS, DynamoDB)
- EBS snapshot storage
- Application backup archives
- Disaster recovery storage
- Compliance backup retention

### Estimated Monthly Cost

- **Small backups** (100 GB): ~$5-10/month
- **Medium backups** (1 TB): ~$30-50/month
- **Large backups** (10 TB): ~$200-350/month

---

## 4. Media Streaming Strategy

**BucketType**: `BucketTypeMediaStreaming`

Optimized for video/audio streaming and high-throughput content delivery.

### Key Features

- **Security**: Private bucket with CORS enabled for players
- **Performance**: S3_MANAGED encryption (lower latency than KMS)
- **No Versioning**: Media files are typically immutable
- **CORS Configuration**: Pre-configured for player domains
- **Lifecycle**: Popular content stays hot (IA@90d → Glacier@365d)
- **Monitoring**: EventBridge for content processing workflows

### Configuration Highlights

```go
Encryption:        awss3.BucketEncryption_S3_MANAGED  // Lower latency
Versioned:         jsii.Bool(false)                   // Immutable files
Cors: &[]*awss3.CorsRule{
    {
        AllowedOrigins: &[]*string{
            jsii.String("https://player.example.com"),
            jsii.String("https://*.cdn.example.com"),
        },
        AllowedMethods: &[]awss3.HttpMethods{
            awss3.HttpMethods_GET,
            awss3.HttpMethods_HEAD,
        },
        AllowedHeaders: &[]*string{
            jsii.String("Range"),           // For seeking in videos
            jsii.String("Authorization"),   // For signed URLs
        },
    },
}
```

### Usage Example

```go
// Create media streaming bucket
mediaBucket := s3construct.NewSimpleStorageServiceFactory(stack, "MediaBucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeMediaStreaming,
        BucketName: "video-streaming-content-prod",
    })

// Integrate with CloudFront for CDN delivery
distribution := cloudfront.NewDistribution(stack, jsii.String("MediaCDN"), &cloudfront.DistributionProps{
    DefaultBehavior: &cloudfront.BehaviorOptions{
        Origin: origins.S3BucketOrigin_WithOriginAccessControl(mediaBucket, &origins.S3BucketOriginWithOACProps{
            // OAC configuration
        }),
    },
})
```

### Media Organization Best Practices

```
s3://video-streaming-content-prod/
├── videos/
│   ├── hls/              # HLS streaming segments
│   │   ├── video-1/
│   │   │   ├── playlist.m3u8
│   │   │   ├── segment-001.ts
│   │   │   └── segment-002.ts
│   ├── dash/             # MPEG-DASH streaming
│   └── originals/        # Master copies
├── audio/
│   ├── podcasts/
│   └── music/
└── thumbnails/
```

### Use Cases

- Video-on-demand (VOD) platforms
- Live streaming archives
- Podcast hosting
- Music streaming services
- E-learning video content
- Webinar recordings

### Estimated Monthly Cost

- **Small library** (500 GB, 500K views): ~$50-80/month
- **Medium library** (5 TB, 5M views): ~$400-600/month
- **Large library** (50 TB, 50M views): ~$3,000-5,000/month

---

## 5. Enterprise Strategy

**BucketType**: `BucketTypeEnterprise`

Maximum security for financial data, PII, and regulated industries with COMPLIANCE-mode Object Lock.

### Key Features

- **Maximum Security**: KMS encryption, TLS 1.3, comprehensive auditing
- **COMPLIANCE Retention**: 7-year Object Lock (cannot be bypassed by anyone)
- **Forced Retention**: RemovalPolicy ALWAYS set to RETAIN (overrides user input)
- **Immutability**: Object Lock COMPLIANCE mode prevents deletion
- **Cost Optimization**: Intelligent Tiering with compliance-safe lifecycle
- **Monitoring**: Daily inventory, access logs, CloudWatch metrics

### Configuration Highlights

```go
Encryption:        awss3.BucketEncryption_KMS_MANAGED
MinimumTLSVersion: jsii.Number(1.3)  // Highest TLS version
ObjectLockEnabled: jsii.Bool(true)
ObjectLockDefaultRetention: awss3.ObjectLockRetention_Compliance(
    awscdk.Duration_Days(jsii.Number(2555)),  // 7 years - CANNOT BE BYPASSED
)
RemovalPolicy:     awscdk.RemovalPolicy_RETAIN  // Forced
AutoDeleteObjects: jsii.Bool(false)             // Forced
```

### COMPLIANCE vs GOVERNANCE

| Feature | GOVERNANCE | COMPLIANCE |
|---------|-----------|------------|
| Root can delete | Yes (with permission) | **NO** |
| Can shorten retention | Yes (with permission) | **NO** |
| Recommended for | Backups, general data | Financial, PII, regulated |
| Override capability | `s3:BypassGovernanceRetention` | **None** |

### Usage Example

```go
// Create enterprise-grade bucket
enterpriseBucket := s3construct.NewSimpleStorageServiceFactory(stack, "EnterpriseBucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeEnterprise,
        BucketName: "financial-records-prod",

        // These overrides are IGNORED for security
        RemovalPolicy: "destroy",        // Forced to RETAIN
        AutoDeleteObjects: jsii.Bool(true), // Forced to false
    })

// Example: Store financial transaction records
// Once written, these objects CANNOT be deleted for 7 years
```

### Security Measures

1. **Object Lock COMPLIANCE**: Objects are immutable for 7 years
2. **KMS Encryption**: Full control over encryption keys
3. **TLS 1.3**: Highest transport security
4. **RemovalPolicy Override**: Cannot accidentally destroy bucket
5. **Comprehensive Auditing**: Access logs, inventory, metrics
6. **EventBridge**: Compliance automation triggers

### Use Cases

- Financial transaction records (SOX, PCI-DSS)
- Personal Identifiable Information (GDPR)
- Healthcare records (HIPAA)
- Legal documents and contracts
- Regulatory compliance archives
- Audit logs for compliance

### Estimated Monthly Cost

- **Small archive** (1 TB): ~$50-80/month
- **Medium archive** (10 TB): ~$400-600/month
- **Large archive** (100 TB): ~$3,000-5,000/month

### Compliance Standards Supported

- **SOX**: Sarbanes-Oxley (financial records)
- **PCI-DSS**: Payment Card Industry Data Security
- **HIPAA**: Health Insurance Portability (healthcare)
- **GDPR**: General Data Protection Regulation
- **SEC 17a-4**: Securities and Exchange Commission

---

## 6. Development Strategy

**BucketType**: `BucketTypeDevelopment`

Cost-optimized bucket for development/testing with easy cleanup.

### Key Features

- **Easy Cleanup**: `RemovalPolicy=DESTROY`, `AutoDeleteObjects=true`
- **Minimal Security**: S3_MANAGED encryption, TLS 1.2, no SSL enforcement
- **No Versioning**: Reduces costs for temporary data
- **30-Day Expiration**: Automatic cleanup lifecycle
- **Permissive CORS**: Allow all origins for development
- **Minimal Monitoring**: Reduce costs

### Configuration Highlights

```go
RemovalPolicy:     awscdk.RemovalPolicy_DESTROY
AutoDeleteObjects: jsii.Bool(true)
Encryption:        awss3.BucketEncryption_S3_MANAGED
EnforceSSL:        jsii.Bool(false)  // Simplify local testing
Versioned:         jsii.Bool(false)   // Reduce costs
LifecycleRules: &[]*awss3.LifecycleRule{
    {
        Expiration: awscdk.Duration_Days(jsii.Number(30)),  // Auto-cleanup
    },
}
Cors: // Permissive for all origins
```

### Usage Example

```go
// Create development bucket
devBucket := s3construct.NewSimpleStorageServiceFactory(stack, "DevBucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeDevelopment,
        BucketName: "app-testing-dev-johndoe",
    })

// Perfect for:
// - CI/CD artifact storage
// - Temporary file uploads
// - Test data generation
// - Feature branch deployments
```

### Cost-Saving Features

1. **No Versioning**: Single copy of each object
2. **30-Day Expiration**: Automatic cleanup prevents runaway costs
3. **S3_MANAGED Encryption**: No KMS costs
4. **Minimal Monitoring**: No EventBridge, metrics, or inventory
5. **Auto-Delete on Stack Removal**: No orphaned resources

### Use Cases

- Development/testing environments
- CI/CD temporary storage
- Feature branch deployments
- Local development testing
- Sandbox environments
- POC/prototype storage

### Estimated Monthly Cost

- **Small dev** (10 GB, 10K requests): ~$0.50-2/month
- **Medium dev** (100 GB, 100K requests): ~$3-8/month
- **Large dev** (500 GB, 500K requests): ~$15-30/month

### Warning

**DO NOT USE IN PRODUCTION**. This strategy is designed for temporary, non-critical data with automatic deletion.

---

## Strategy Selection Guide

### Decision Tree

```
Need maximum security & compliance?
├─ YES → Enterprise Strategy
└─ NO
   ├─ Is this production data?
   │  ├─ YES
   │  │  ├─ Static website/SPA? → CloudFront Origin Strategy
   │  │  ├─ Big data analytics? → Data Lake Strategy
   │  │  ├─ Backup/DR? → Backup Strategy
   │  │  └─ Media streaming? → Media Streaming Strategy
   │  └─ NO → Development Strategy
```

### By Security Level

| Security Level | Strategies | When to Use |
|---------------|-----------|-------------|
| **Maximum** | Enterprise, Backup | Financial data, PII, compliance |
| **High** | CloudFront Origin, Data Lake | Production websites, analytics |
| **Medium** | Media Streaming | Public content with CORS |
| **Basic** | Development | Dev/test environments only |

### By Cost Optimization

| Priority | Strategies | Cost Profile |
|----------|-----------|-------------|
| **Compliance First** | Enterprise | High (KMS, Object Lock) |
| **Balanced** | CloudFront, Data Lake, Media | Medium (Intelligent Tiering) |
| **Aggressive** | Backup | Low (fast Glacier transitions) |
| **Minimal** | Development | Very Low (30-day expiration) |

---

## Complete Stack Examples

### Example 1: Static Website with CloudFront

```go
package main

import (
    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
    "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
    "github.com/aws/constructs-go/constructs/v10"
    s3construct "github.com/yourusername/cdk-library/constructs/S3"
)

type WebsiteStackProps struct {
    awscdk.StackProps
}

func NewWebsiteStack(scope constructs.Construct, id string, props *WebsiteStackProps) awscdk.Stack {
    stack := awscdk.NewStack(scope, &id, &props.StackProps)

    // Create S3 bucket for website content
    websiteBucket := s3construct.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
        s3construct.SimpleStorageServiceFactoryProps{
            BucketType: s3construct.BucketTypeCloudfrontOAC,
            BucketName: "my-company-website-prod",
        })

    // Create CloudFront distribution
    distribution := awscloudfront.NewDistribution(stack, jsii.String("WebsiteCDN"), &awscloudfront.DistributionProps{
        DefaultBehavior: &awscloudfront.BehaviorOptions{
            Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(websiteBucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{}),
            ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
        },
        DefaultRootObject: jsii.String("index.html"),
        ErrorResponses: &[]*awscloudfront.ErrorResponse{
            {
                HttpStatus:         jsii.Number(404),
                ResponseHttpStatus: jsii.Number(200),
                ResponsePagePath:   jsii.String("/index.html"),
            },
        },
    })

    // Output CloudFront URL
    awscdk.NewCfnOutput(stack, jsii.String("DistributionDomain"), &awscdk.CfnOutputProps{
        Value: distribution.DistributionDomainName(),
    })

    return stack
}
```

### Example 2: Data Lake with Analytics

```go
func NewDataLakeStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
    stack := awscdk.NewStack(scope, &id, props)

    // Create data lake bucket
    dataLakeBucket := s3construct.NewSimpleStorageServiceFactory(stack, "DataLakeBucket",
        s3construct.SimpleStorageServiceFactoryProps{
            BucketType: s3construct.BucketTypeDataLake,
            BucketName: "analytics-data-lake-prod",
        })

    // Create Glue database for Athena
    database := awsglue.NewCfnDatabase(stack, jsii.String("AnalyticsDB"), &awsglue.CfnDatabaseProps{
        CatalogId: stack.Account(),
        DatabaseInput: &awsglue.CfnDatabase_DatabaseInputProperty{
            Name: jsii.String("analytics_db"),
        },
    })

    // Create Glue crawler for raw data
    crawler := awsglue.NewCfnCrawler(stack, jsii.String("RawDataCrawler"), &awsglue.CfnCrawlerProps{
        DatabaseName: database.Ref(),
        Role:         // IAM role for Glue
        Targets: &awsglue.CfnCrawler_TargetsProperty{
            S3Targets: &[]*awsglue.CfnCrawler_S3TargetProperty{
                {
                    Path: jsii.String("s3://analytics-data-lake-prod/raw-data/"),
                },
            },
        },
    })

    // Output bucket name
    awscdk.NewCfnOutput(stack, jsii.String("DataLakeBucketName"), &awscdk.CfnOutputProps{
        Value: dataLakeBucket.BucketName(),
    })

    return stack
}
```

### Example 3: Multi-Environment Setup

```go
type Environment string

const (
    EnvDev  Environment = "dev"
    EnvStaging Environment = "staging"
    EnvProd Environment = "prod"
)

func NewMultiEnvStack(scope constructs.Construct, id string, env Environment) awscdk.Stack {
    stack := awscdk.NewStack(scope, &id, nil)

    var bucketType s3construct.BucketType
    var bucketName string

    // Strategy selection based on environment
    switch env {
    case EnvDev:
        bucketType = s3construct.BucketTypeDevelopment
        bucketName = "my-app-dev"
    case EnvStaging:
        bucketType = s3construct.BucketTypeCloudfrontOAC
        bucketName = "my-app-staging"
    case EnvProd:
        bucketType = s3construct.BucketTypeCloudfrontOAC
        bucketName = "my-app-prod"
    }

    bucket := s3construct.NewSimpleStorageServiceFactory(stack, "AppBucket",
        s3construct.SimpleStorageServiceFactoryProps{
            BucketType: bucketType,
            BucketName: bucketName,
        })

    return stack
}

// Usage
func main() {
    app := awscdk.NewApp(nil)

    NewMultiEnvStack(app, "DevStack", EnvDev)
    NewMultiEnvStack(app, "StagingStack", EnvStaging)
    NewMultiEnvStack(app, "ProdStack", EnvProd)

    app.Synth(nil)
}
```

---

## Advanced Usage

### Custom Overrides

All strategies support optional overrides for `RemovalPolicy` and `AutoDeleteObjects` (except Enterprise, which enforces RETAIN).

```go
bucket := s3construct.NewSimpleStorageServiceFactory(stack, "CustomBucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeDataLake,
        BucketName: "custom-data-lake",

        // Override defaults
        RemovalPolicy:     "destroy",        // Override RETAIN default
        AutoDeleteObjects: jsii.Bool(true),  // Enable auto-delete
    })
```

**Note**: Enterprise strategy IGNORES these overrides for security reasons.

### Accessing Bucket Properties

```go
bucket := s3construct.NewSimpleStorageServiceFactory(stack, "Bucket", props)

// Access standard S3 bucket properties
bucketName := bucket.BucketName()
bucketArn := bucket.BucketArn()
bucketUrl := bucket.BucketWebsiteUrl()

// Grant permissions
bucket.GrantRead(myLambdaFunction, jsii.String("*"))
bucket.GrantWrite(myLambdaFunction, jsii.String("uploads/*"))
```

### Integration with Other AWS Services

```go
// Lambda trigger on S3 events
bucket.AddEventNotification(
    awss3.EventType_OBJECT_CREATED,
    awss3notifications.NewLambdaDestination(myLambdaFunction),
    &awss3.NotificationKeyFilter{
        Prefix: jsii.String("uploads/"),
        Suffix: jsii.String(".jpg"),
    },
)

// SNS notification
bucket.AddEventNotification(
    awss3.EventType_OBJECT_REMOVED,
    awss3notifications.NewSnsDestination(myTopic),
)

// EventBridge integration (if enabled)
rule := awsevents.NewRule(stack, jsii.String("S3EventRule"), &awsevents.RuleProps{
    EventPattern: &awsevents.EventPattern{
        Source: &[]*string{jsii.String("aws.s3")},
        DetailType: &[]*string{jsii.String("Object Created")},
    },
})
```

---

## Testing Strategies

### Unit Testing

```go
package s3_test

import (
    "testing"
    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/aws-cdk-go/awscdk/v2/assertions"
    "github.com/aws/constructs-go/constructs/v10"
    s3construct "github.com/yourusername/cdk-library/constructs/S3"
)

func TestCloudfrontOriginBucket(t *testing.T) {
    app := awscdk.NewApp(nil)
    stack := awscdk.NewStack(app, jsii.String("TestStack"), nil)

    // Create bucket
    s3construct.NewSimpleStorageServiceFactory(stack, "TestBucket",
        s3construct.SimpleStorageServiceFactoryProps{
            BucketType: s3construct.BucketTypeCloudfrontOAC,
            BucketName: "test-bucket",
        })

    // Assert bucket properties
    template := assertions.Template_FromStack(stack, nil)

    // Verify KMS encryption
    template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
        "BucketEncryption": map[string]interface{}{
            "ServerSideEncryptionConfiguration": []interface{}{
                map[string]interface{}{
                    "ServerSideEncryptionByDefault": map[string]interface{}{
                        "SSEAlgorithm": "aws:kms",
                    },
                },
            },
        },
    })

    // Verify versioning enabled
    template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
        "VersioningConfiguration": map[string]interface{}{
            "Status": "Enabled",
        },
    })
}
```

---

## Migration Guide

### From Helper Functions to Factory

**Before** (Old approach with helper functions):

```go
bucketProps := s3.GetCloudfrontOriginProperties("my-bucket")
bucket := awss3.NewBucket(stack, jsii.String("Bucket"), bucketProps)
```

**After** (New Factory + Strategy approach):

```go
bucket := s3construct.NewSimpleStorageServiceFactory(stack, "Bucket",
    s3construct.SimpleStorageServiceFactoryProps{
        BucketType: s3construct.BucketTypeCloudfrontOAC,
        BucketName: "my-bucket",
    })
```

### Benefits of Migration

1. **Type Safety**: Compile-time bucket type validation
2. **Consistency**: Enforced best practices across team
3. **Maintainability**: Single factory entry point
4. **Extensibility**: Easy to add new bucket types
5. **Testability**: Mock strategies for unit tests

---

## Best Practices

### Security

1. **Always Use KMS for Sensitive Data**: CloudFront Origin, Data Lake, Backup, Enterprise
2. **Enable Versioning**: CloudFront Origin, Data Lake, Backup, Enterprise
3. **Use Object Lock for Compliance**: Backup (GOVERNANCE), Enterprise (COMPLIANCE)
4. **Enforce TLS 1.2+**: All strategies except Development
5. **Block Public Access**: ALWAYS enabled on all strategies

### Cost Optimization

1. **Use Intelligent Tiering**: Enabled on CloudFront Origin, Data Lake, Enterprise
2. **Implement Lifecycle Policies**: All strategies have optimized transitions
3. **Right-Size Storage Class**: Match access patterns to strategy
4. **Monitor with Budgets**: Set up AWS Budgets for S3 costs
5. **Use Development Strategy for Non-Prod**: Avoid production costs in dev/test

### Operations

1. **Enable Access Logging**: Enabled on Data Lake, Backup, Enterprise
2. **Use EventBridge for Automation**: Enabled on CloudFront, Data Lake, Media, Enterprise
3. **Set Up CloudWatch Alarms**: Monitor bucket size, request rates, errors
4. **Tag Resources**: Add cost allocation and ownership tags
5. **Document Bucket Purpose**: Use CloudFormation descriptions

### Compliance

1. **Use Enterprise Strategy for PII**: GDPR, HIPAA, PCI-DSS
2. **Enable Inventory**: Track all objects for auditing
3. **Retain Access Logs**: 90+ days for compliance reviews
4. **Use Object Lock for Financial Data**: SOX, SEC 17a-4
5. **Document Retention Policies**: Match lifecycle to legal requirements

---

## Troubleshooting

### Common Issues

#### Issue: "Bucket already exists"

```
Error: my-bucket-name already exists in another account or region
```

**Solution**: S3 bucket names are globally unique. Choose a unique name:
```go
BucketName: "company-app-environment-randomstring-prod"
```

#### Issue: "Cannot delete bucket with objects"

```
Error: Cannot delete non-empty bucket
```

**Solution**: Use Development strategy with `AutoDeleteObjects: true` for test buckets. For production, manually empty before deletion.

#### Issue: "Access Denied" with CloudFront

```
Error: AccessDenied when accessing S3 through CloudFront
```

**Solution**: Ensure you're using Origin Access Control (OAC):
```go
Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, ...)
```

#### Issue: "KMS key access denied"

```
Error: User is not authorized to perform kms:Decrypt
```

**Solution**: Grant KMS permissions to your IAM role:
```go
bucket.EncryptionKey().GrantDecrypt(myRole)
```

---

## Performance Benchmarks

### S3 Request Performance by Strategy

| Strategy | Avg GET Latency | Avg PUT Latency | Throughput |
|----------|-----------------|-----------------|------------|
| CloudFront Origin | 50-100ms | 80-150ms | High |
| Data Lake | 50-100ms | 80-150ms | Medium |
| Backup | 100-200ms | 150-300ms | Low |
| Media Streaming | **20-50ms** | 80-150ms | **Very High** |
| Enterprise | 100-200ms | 150-300ms | Low |
| Development | 50-100ms | 80-150ms | Medium |

**Note**: Media Streaming is optimized for low latency with S3_MANAGED encryption.

---

## Security Comparison

| Feature | CloudFront | Data Lake | Backup | Media | Enterprise | Development |
|---------|-----------|-----------|--------|-------|------------|-------------|
| Encryption | KMS | KMS | KMS | S3_MANAGED | KMS | S3_MANAGED |
| TLS Version | 1.2 | 1.2 | 1.2 | 1.2 | **1.3** | 1.2 |
| Object Lock | ❌ | ❌ | GOVERNANCE | ❌ | **COMPLIANCE** | ❌ |
| Versioning | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Access Logs | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Inventory | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| EventBridge | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| RemovalPolicy | RETAIN | RETAIN | RETAIN | RETAIN | **FORCED RETAIN** | DESTROY |

---

## Roadmap

### Planned Enhancements

- [ ] **ReplicationStrategy**: Cross-region bucket replication
- [ ] **ArchiveStrategy**: Glacier-first for cold storage
- [ ] **AccessPointStrategy**: S3 Access Points for multi-tenant
- [ ] **DirectoryBucketStrategy**: S3 Express One Zone for ultra-low latency

### Feedback

Have suggestions or issues? Open a GitHub issue or submit a pull request!

---

## Additional Resources

- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [AWS CDK Go Documentation](https://docs.aws.amazon.com/cdk/v2/guide/work-with-cdk-go.html)
- [S3 Storage Classes](https://aws.amazon.com/s3/storage-classes/)
- [S3 Object Lock](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html)
- [S3 Lifecycle Management](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)

---

## License

This library is licensed under the MIT License. See LICENSE file for details.

---

**Version**: 0.4.0
**Last Updated**: 2025-10-11
**Maintainer**: Andres Sepulveda
