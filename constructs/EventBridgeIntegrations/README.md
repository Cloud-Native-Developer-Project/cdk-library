# EventBridge Integrations Construct

EventBridge integration constructs using **Factory + Strategy pattern** for consistent, reusable event-driven architectures.

## Architecture

```
User Code → Factory (Type Selection) → Strategy Interface → Concrete Strategy → AWS Resources
```

## Available Strategies

### 1. S3 → Lambda Integration

Connects S3 bucket events to Lambda function invocation with configurable filtering and retry policies.

**Use cases:**
- Process uploaded files (document parsing, image transformation)
- Send webhook notifications when files are created
- Trigger workflows on S3 object lifecycle events

**Configuration:**
- Filter by bucket, object key prefix/suffix
- Multiple event types (Created, Removed, Restored)
- Configurable retry attempts and max event age
- Optional Dead Letter Queue for failed invocations

## Usage Examples

### Basic S3 → Lambda Integration

```go
package main

import (
    "github.com/aws/aws-cdk-go/awscdk/v2"
    "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
    "cdk-library/constructs/S3"
    "cdk-library/constructs/EventBridgeIntegrations"
    "github.com/aws/jsii-runtime-go"
)

// Create S3 bucket
bucket := s3.NewSimpleStorageServiceFactory(stack, "Bucket",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeEnterprise,
        BucketName: "my-data-bucket",
    })

// Create Lambda function
lambda := awslambda.NewFunction(stack, jsii.String("Handler"), &awslambda.FunctionProps{
    Runtime:      awslambda.Runtime_GO_1_X(),
    Architecture: awslambda.Architecture_ARM_64(),
    Code:         awslambda.Code_FromAsset(jsii.String("lambda/handler"), nil),
    Handler:      jsii.String("main"),
})

// Create integration using Factory pattern
rule := eventbridgeintegrations.NewEventBridgeIntegrationFactory(
    stack,
    "S3ToLambda",
    eventbridgeintegrations.EventBridgeIntegrationFactoryProps{
        IntegrationType: eventbridgeintegrations.IntegrationTypeS3ToLambda,
        S3ToLambdaConfig: &eventbridgeintegrations.S3ToLambdaConfig{
            SourceBucket:     bucket,
            TargetLambda:     lambda,
            ObjectKeyPrefix:  jsii.String("uploads/"),
            MaxRetryAttempts: jsii.Number(4),
            EnableDLQ:        jsii.Bool(true),
        },
    })
```

### Advanced: Multiple Integrations (Same Bucket, Different Lambdas)

```go
// Scenario: One bucket, multiple event handlers

bucket := s3.NewSimpleStorageServiceFactory(stack, "LandingZone",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeEnterprise,
        BucketName: "addi-landing-zone-prod",
    })

lambdaWebhook := awslambda.NewFunction(stack, jsii.String("WebhookNotifier"), &awslambda.FunctionProps{...})
lambdaProcessor := awslambda.NewFunction(stack, jsii.String("Processor"), &awslambda.FunctionProps{...})

// Integration 1: uploads/ → Webhook Lambda
eventbridgeintegrations.NewEventBridgeIntegrationFactory(
    stack,
    "UploadsToWebhook",
    eventbridgeintegrations.EventBridgeIntegrationFactoryProps{
        IntegrationType: eventbridgeintegrations.IntegrationTypeS3ToLambda,
        S3ToLambdaConfig: &eventbridgeintegrations.S3ToLambdaConfig{
            SourceBucket:     bucket,
            TargetLambda:     lambdaWebhook,
            ObjectKeyPrefix:  jsii.String("uploads/"),
            EventTypes:       []string{"Object Created"},
            MaxRetryAttempts: jsii.Number(4),
            EnableDLQ:        jsii.Bool(true),
        },
    })

// Integration 2: pending/*.pdf → Processor Lambda
eventbridgeintegrations.NewEventBridgeIntegrationFactory(
    stack,
    "PendingToProcessor",
    eventbridgeintegrations.EventBridgeIntegrationFactoryProps{
        IntegrationType: eventbridgeintegrations.IntegrationTypeS3ToLambda,
        S3ToLambdaConfig: &eventbridgeintegrations.S3ToLambdaConfig{
            SourceBucket:     bucket,
            TargetLambda:     lambdaProcessor,
            ObjectKeyPrefix:  jsii.String("pending/"),
            ObjectKeySuffix:  jsii.String(".pdf"),
            EventTypes:       []string{"Object Created"},
            MaxRetryAttempts: jsii.Number(2),
            EnableDLQ:        jsii.Bool(false),
        },
    })
```

### Event Types

Common S3 event types:
- `"Object Created"` - Object uploaded via PUT, POST, COPY, or CompleteMultipartUpload
- `"Object Removed"` - Object deleted (DELETE API)
- `"Object Restore Initiated"` - Restore from Glacier initiated
- `"Object Restore Completed"` - Restore from Glacier completed

## Pattern Benefits

### Factory + Strategy Pattern
- **Open/Closed Principle**: Add new integration types without modifying existing code
- **Single Responsibility**: Each strategy handles one specific integration pattern
- **Type Safety**: Compile-time validation of configuration
- **Consistency**: Same API pattern as S3, CloudFront, WAF, GuardDuty constructs

### CDK Encapsulation
- **IAM Permissions**: Automatically configured (EventBridge → Lambda)
- **Event Pattern**: Dynamic filtering by bucket name, prefix, suffix
- **Dead Letter Queue**: Optional SQS queue for failed events
- **EventBridge Notification**: Automatically enabled on S3 bucket

## Cost Optimization

**EventBridge Pricing:**
- Custom events: $1.00 per million events
- Event replay: $0.10 per GB

**Example: 10,000 S3 events/month**
- EventBridge: 10,000 × $1.00/million = **$0.01/month**
- Lambda invocations: Depends on function configuration
- SQS DLQ (if enabled): ~$0.00 (minimal usage)

**Total EventBridge overhead: < $0.05/month** (negligible compared to data transfer)

## Implementation Status

### Completed (1 strategy)
- ✅ S3 → Lambda Integration

### Planned (future expansion)
- ⏳ S3 → SQS Integration (buffering for batch processing)
- ⏳ S3 → API Destination Integration (webhooks without Lambda)
- ⏳ Scheduled → Lambda Integration (cron/rate expressions)
- ⏳ Custom Event → Multiple Targets Integration (fan-out)

## Files Structure

```
constructs/EventBridgeIntegrations/
├── eventbridge_integration_factory.go      # Factory entry point
├── eventbridge_integration_contract.go     # Strategy interface
├── eventbridge_s3_to_lambda.go            # S3→Lambda strategy
└── README.md                               # Documentation
```

## Adding New Strategies

To add a new integration type (e.g., S3 → SQS):

1. **Add constant** in `eventbridge_integration_factory.go`:
   ```go
   const IntegrationTypeS3ToSQS IntegrationType = "S3_TO_SQS"
   ```

2. **Create config struct** in factory props:
   ```go
   S3ToSQSConfig *S3ToSQSConfig
   ```

3. **Implement strategy** in `eventbridge_s3_to_sqs.go`:
   ```go
   type EventBridgeS3ToSQSStrategy struct{}
   func (s *EventBridgeS3ToSQSStrategy) Build(...) awsevents.Rule { ... }
   ```

4. **Register in factory** switch statement:
   ```go
   case IntegrationTypeS3ToSQS:
       strategy = &EventBridgeS3ToSQSStrategy{}
   ```

## References

- [EventBridge Targets Documentation](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-targets.html)
- [S3 EventBridge Integration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/EventBridge.html)
- [CDK EventBridge Constructs](https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_events-readme.html)
