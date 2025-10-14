# Lambda Construct

Simplified Lambda construct optimized for **Go runtime** with sensible production defaults.

## Philosophy

Unlike other constructs in this library (S3, CloudFront, WAF, GuardDuty), Lambda **does not use Factory + Strategy pattern** because:

1. Lambda configuration is mostly **parameter variation**, not different creation strategies
2. The business logic lives in **external code files** (`lambda/*/main.go`)
3. Most Lambdas share the same structure - only memory, timeout, env vars differ

This construct provides a **single, flexible function** with smart defaults.

## Usage

### Minimal Configuration (Uses all defaults)

```go
import (
    golambda "cdk-library/constructs/Lambda"
)

lambda := golambda.NewGoLambda(stack, "WebhookHandler",
    golambda.GoLambdaProps{
        FunctionName: "webhook-notifier",
        CodePath:     "lambda/webhook-notifier",
    })
```

**Defaults applied:**
- Runtime: Go 1.x
- Architecture: ARM64 (Graviton2)
- Memory: 512 MB
- Timeout: 30 seconds
- Handler: "main"
- Tracing: Active (X-Ray)
- Retry attempts: 2

### Custom Configuration

```go
lambda := golambda.NewGoLambda(stack, "HeavyProcessor",
    golambda.GoLambdaProps{
        FunctionName: "document-processor",
        CodePath:     "lambda/processor",
        Description:  jsii.String("Processes large PDF documents"),
        MemorySize:   jsii.Number(2048),        // Override: 2GB RAM
        Timeout:      awscdk.Duration_Minutes(jsii.Number(5)), // Override: 5 min
        Environment: &map[string]*string{
            "BATCH_SIZE":    jsii.String("100"),
            "MAX_FILE_SIZE": jsii.String("10485760"), // 10MB
        },
    })
```

### Advanced Configuration (All options)

```go
dlq := awssqs.NewQueue(stack, jsii.String("DLQ"), &awssqs.QueueProps{
    QueueName: jsii.String("processor-dlq"),
})

layer := awslambda.LayerVersion_FromLayerVersionArn(
    stack,
    jsii.String("SharedLayer"),
    jsii.String("arn:aws:lambda:us-east-1:123456789:layer:shared:1"),
)

lambda := golambda.NewGoLambda(stack, "AdvancedProcessor",
    golambda.GoLambdaProps{
        FunctionName: "advanced-processor",
        CodePath:     "lambda/advanced",
        Description:  jsii.String("Advanced document processing"),
        Handler:      jsii.String("bootstrap"), // Custom handler for AL2023
        Architecture: awslambda.Architecture_X86_64(), // Override to x86
        MemorySize:   jsii.Number(4096),
        Timeout:      awscdk.Duration_Minutes(jsii.Number(15)),
        ReservedConcurrentExecutions: jsii.Number(10), // Limit concurrency
        RetryAttempts: jsii.Number(0), // Disable retries
        DeadLetterQueue: dlq,
        Layers: &[]awslambda.ILayerVersion{layer},
        Tracing: awslambda.Tracing_PASS_THROUGH,
        Environment: &map[string]*string{
            "LOG_LEVEL": jsii.String("debug"),
        },
    })
```

## Configuration Options

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `FunctionName` | `string` | Lambda function name (must be unique per region) |
| `CodePath` | `string` | Path to Lambda code directory (e.g., `"lambda/handler"`) |

### Optional Fields (with defaults)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Description` | `*string` | `nil` | Function description |
| `Handler` | `*string` | `"main"` | Executable name (standard for Go) |
| `Architecture` | `Architecture` | `ARM64` | CPU architecture (ARM64 = 20% cheaper) |
| `MemorySize` | `*float64` | `512` | Memory in MB (128-10240) |
| `Timeout` | `Duration` | `30 seconds` | Max execution time |
| `Environment` | `*map[string]*string` | `{}` | Environment variables |
| `ReservedConcurrentExecutions` | `*float64` | `nil` | Limit concurrent executions |
| `RetryAttempts` | `*float64` | `2` | Async invocation retries |
| `DeadLetterQueue` | `IQueue` | `nil` | SQS queue for failed invocations |
| `Layers` | `*[]ILayerVersion` | `[]` | Lambda layers |
| `Tracing` | `Tracing` | `Active` | X-Ray tracing mode |

## Defaults Explained

### Why ARM64 (Graviton2)?
- **20% cost savings** vs x86_64
- Equal or better performance for most workloads
- Go has excellent ARM64 support (native compilation)

**When to use x86_64:**
- Third-party libraries without ARM support
- Legacy dependencies
- Performance testing shows x86 is faster for your use case

### Why 512 MB Memory?
- Good balance between cost and performance
- Sufficient for most API/webhook handlers
- Go's low memory footprint makes this viable

**When to increase:**
- Processing large files (2-4 GB)
- Heavy CPU operations (more memory = more CPU)
- Concurrent processing within one invocation

### Why 30 Second Timeout?
- Suitable for webhook/API operations
- Prevents runaway processes
- Forces proper async design

**When to increase:**
- File processing (1-5 minutes)
- External API calls with slow response
- Batch operations (up to 15 minutes)

### Why 2 Retry Attempts?
- Balance between reliability and cost
- Total attempts: 1 initial + 2 retries = 3 tries
- Exponential backoff between retries

**When to change:**
- Set to 0 for idempotent operations (e.g., EventBridge triggers)
- Increase to 3-4 for critical operations

## Integration with Other Constructs

### Example: S3 + EventBridge + Lambda (Addi Stack)

```go
import (
    "cdk-library/constructs/S3"
    "cdk-library/constructs/Lambda"
    "cdk-library/constructs/EventBridgeIntegrations"
)

// 1. Create S3 bucket
bucket := s3.NewSimpleStorageServiceFactory(stack, "LandingZone",
    s3.SimpleStorageServiceFactoryProps{
        BucketType: s3.BucketTypeEnterprise,
        BucketName: "addi-landing-zone-prod",
    })

// 2. Create Lambda with construct (clean & simple)
lambda := golambda.NewGoLambda(stack, "WebhookNotifier",
    golambda.GoLambdaProps{
        FunctionName: "webhook-notifier",
        CodePath:     "lambda/webhook-notifier",
        Environment: &map[string]*string{
            "BUCKET_NAME": bucket.BucketName(),
        },
    })

// 3. Grant permissions
bucket.GrantRead(lambda, jsii.String("uploads/*"))

// 4. Create EventBridge integration
eventbridgeintegrations.NewEventBridgeIntegrationFactory(
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

## Lambda Code Structure

Your Lambda code should be in a directory with a compiled Go binary:

```
project/
├── lambda/
│   ├── webhook-notifier/
│   │   ├── main.go           # Your handler code
│   │   └── go.mod            # Dependencies
│   └── processor/
│       ├── main.go
│       └── go.mod
├── constructs/
│   └── Lambda/
│       └── go_lambda.go      # This construct
└── stacks/
    └── addi/
        └── addi_stack.go     # Uses construct
```

**Build before deployment:**
```bash
# Build Lambda for ARM64 (Graviton2)
GOOS=linux GOARCH=arm64 go build -o lambda/webhook-notifier/main lambda/webhook-notifier/main.go

# Or use build script
./build-lambdas.sh
```

## Cost Optimization Tips

### 1. Use ARM64 Architecture
```go
// ✅ Default (ARM64) - 20% cheaper
lambda := golambda.NewGoLambda(stack, "Handler", golambda.GoLambdaProps{...})

// ❌ x86 - 20% more expensive
lambda := golambda.NewGoLambda(stack, "Handler", golambda.GoLambdaProps{
    Architecture: awslambda.Architecture_X86_64(),
    ...
})
```

**Savings example (10,000 invocations/month, 512MB, 1s average):**
- x86_64: $0.0000166667/GB-sec = $0.85/month
- ARM64: $0.0000133334/GB-sec = $0.68/month
- **Savings: $0.17/month (20%)**

### 2. Right-size Memory Allocation

More memory = more CPU, but diminishing returns:

```go
// Test different memory sizes:
MemorySize: jsii.Number(512)   // $0.68/month baseline
MemorySize: jsii.Number(1024)  // $1.36/month (2x cost, ~1.5x performance)
MemorySize: jsii.Number(2048)  // $2.72/month (4x cost, ~2x performance)
```

**Rule of thumb:**
- Start at 512 MB
- Monitor CloudWatch duration metrics
- Increase if consistently hitting timeout
- Decrease if memory usage < 50%

### 3. Limit Concurrent Executions (if needed)

```go
ReservedConcurrentExecutions: jsii.Number(10), // Prevent cost spikes
```

Protects against:
- Runaway event loops
- DDoS attacks triggering Lambda
- Downstream service overload

## Monitoring

The construct enables X-Ray tracing by default. View traces in:
- AWS X-Ray Console
- CloudWatch ServiceLens

**Key metrics to monitor:**
- Invocation count
- Error rate
- Duration (P50, P90, P99)
- Throttles
- Memory utilization

## Files

```
constructs/Lambda/
├── go_lambda.go    # Main construct
└── README.md       # This file
```

## Future Enhancements (Not Implemented)

If you need additional runtimes or configurations:

### Python Lambda
```go
// constructs/Lambda/python_lambda.go
func NewPythonLambda(scope constructs.Construct, id string, props PythonLambdaProps) awslambda.Function
```

### VPC Lambda
```go
// constructs/Lambda/vpc_lambda.go
func NewGoLambdaVPC(scope constructs.Construct, id string, props GoLambdaVPCProps) awslambda.Function
```

### Container Lambda
```go
// constructs/Lambda/container_lambda.go
func NewContainerLambda(scope constructs.Construct, id string, props ContainerLambdaProps) awslambda.Function
```

**Current recommendation:** Create these as needed, following the same pattern (simple function with defaults, not Factory + Strategy).
