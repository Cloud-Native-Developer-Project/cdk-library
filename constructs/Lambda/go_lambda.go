package lambda

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// GoLambdaProps defines configuration for a Go Lambda function with sensible defaults
type GoLambdaProps struct {
	// Function name (REQUIRED)
	FunctionName string

	// Path to Lambda code directory (REQUIRED)
	// Example: "lambda/webhook-notifier" (relative to project root)
	CodePath string

	// Function description
	// Optional: defaults to empty string
	Description *string

	// Handler executable name
	// Optional: defaults to "main" (standard for Go lambdas)
	Handler *string

	// CPU architecture
	// Optional: defaults to ARM64 (Graviton2 - 20% cost savings)
	// Use awslambda.Architecture_X86_64() if you need x86
	Architecture awslambda.Architecture

	// Memory allocation in MB (128-10240)
	// Optional: defaults to 512 MB (good balance for most use cases)
	MemorySize *float64

	// Function timeout
	// Optional: defaults to 30 seconds
	Timeout awscdk.Duration

	// Environment variables
	// Optional: defaults to empty map
	Environment *map[string]*string

	// Reserved concurrent executions
	// Optional: defaults to nil (unreserved, auto-scales)
	// Set to limit concurrent executions (e.g., to protect downstream services)
	ReservedConcurrentExecutions *float64

	// Retry attempts for asynchronous invocations
	// Optional: defaults to 2
	RetryAttempts *float64

	// Dead Letter Queue for failed async invocations
	// Optional: defaults to nil (no DLQ)
	DeadLetterQueue awssqs.IQueue

	// Lambda Layers (e.g., shared libraries, extensions)
	// Optional: defaults to empty array
	Layers *[]awslambda.ILayerVersion

	// Tracing configuration (AWS X-Ray)
	// Optional: defaults to Active tracing
	Tracing awslambda.Tracing
}

// NewGoLambda creates a Go Lambda function with optimized defaults for production use
//
// Opinionated defaults optimized for Go runtime:
// - Runtime: Go 1.x (latest stable)
// - Architecture: ARM64 (Graviton2) - 20% cheaper, equal or better performance
// - Memory: 512 MB - good balance for most Go applications
// - Timeout: 30 seconds - reasonable for webhook/API operations
// - Handler: "main" - standard for compiled Go binaries
// - Retry: 2 attempts - balance between reliability and cost
// - Tracing: Active - enabled for observability
//
// Example usage (minimal):
//
//	lambda := golambda.NewGoLambda(stack, "WebhookHandler",
//	    golambda.GoLambdaProps{
//	        FunctionName: "webhook-notifier",
//	        CodePath:     "lambda/webhook-notifier",
//	    })
//
// Example usage (custom configuration):
//
//	lambda := golambda.NewGoLambda(stack, "HeavyProcessor",
//	    golambda.GoLambdaProps{
//	        FunctionName: "heavy-processor",
//	        CodePath:     "lambda/processor",
//	        Description:  jsii.String("Processes large files"),
//	        MemorySize:   jsii.Number(2048),
//	        Timeout:      awscdk.Duration_Minutes(jsii.Number(5)),
//	        Environment: &map[string]*string{
//	            "BATCH_SIZE": jsii.String("100"),
//	        },
//	    })
func NewGoLambda(scope constructs.Construct, id string, props GoLambdaProps) awslambda.Function {

	// Validate required fields
	if props.FunctionName == "" {
		panic("FunctionName is required")
	}
	if props.CodePath == "" {
		panic("CodePath is required")
	}

	// Apply defaults for optional fields
	handler := props.Handler
	if handler == nil {
		handler = jsii.String("main")
	}

	architecture := props.Architecture
	if architecture == nil {
		architecture = awslambda.Architecture_ARM_64() // Graviton2 - 20% cost savings
	}

	memorySize := props.MemorySize
	if memorySize == nil {
		memorySize = jsii.Number(512) // Good balance for most Go apps
	}

	timeout := props.Timeout
	if timeout == nil {
		timeout = awscdk.Duration_Seconds(jsii.Number(30))
	}

	environment := props.Environment
	if environment == nil {
		environment = &map[string]*string{}
	}

	retryAttempts := props.RetryAttempts
	if retryAttempts == nil {
		retryAttempts = jsii.Number(2)
	}

	tracing := props.Tracing
	if tracing == "" {
		tracing = awslambda.Tracing_ACTIVE // Enable X-Ray tracing
	}

	// Build function props
	functionProps := &awslambda.FunctionProps{
		FunctionName: jsii.String(props.FunctionName),
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: architecture,
		Code:         awslambda.Code_FromAsset(jsii.String(props.CodePath), nil),
		Handler:      handler,
		MemorySize:   memorySize,
		Timeout:      timeout,
		Environment:  environment,
		Tracing:      tracing,
	}

	// Add optional configurations
	if props.Description != nil {
		functionProps.Description = props.Description
	}

	if props.ReservedConcurrentExecutions != nil {
		functionProps.ReservedConcurrentExecutions = props.ReservedConcurrentExecutions
	}

	if props.DeadLetterQueue != nil {
		functionProps.DeadLetterQueue = props.DeadLetterQueue
	}

	if props.Layers != nil {
		functionProps.Layers = props.Layers
	}

	// Create Lambda function
	lambda := awslambda.NewFunction(scope, jsii.String(id), functionProps)

	// Configure retry attempts for async invocations
	lambda.ConfigureAsyncInvoke(&awslambda.EventInvokeConfigOptions{
		RetryAttempts: retryAttempts,
	})

	return lambda
}
