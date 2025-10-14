package eventbridgeintegrations

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// S3ToLambdaConfig defines configuration specific to S3→Lambda integration
type S3ToLambdaConfig struct {
	// Source S3 bucket that emits events (REQUIRED)
	SourceBucket awss3.IBucket

	// Target Lambda function to invoke (REQUIRED)
	TargetLambda awslambda.IFunction

	// Object key prefix filter (e.g., "uploads/", "data/raw/")
	// Only objects with keys starting with this prefix will trigger the Lambda
	// Optional: if nil, all objects in the bucket will match
	ObjectKeyPrefix *string

	// Object key suffix filter (e.g., ".pdf", ".json")
	// Only objects with keys ending with this suffix will trigger the Lambda
	// Optional: if nil, all file types will match
	ObjectKeySuffix *string

	// S3 event types to monitor
	// Common values: "Object Created", "Object Removed", "Object Restore Completed"
	// Optional: defaults to ["Object Created"] if nil or empty
	EventTypes []string

	// Maximum number of retry attempts for failed Lambda invocations
	// Optional: defaults to EventBridge default (185 attempts over 24 hours) if nil
	MaxRetryAttempts *float64

	// Maximum time to retain events for retry
	// Optional: defaults to EventBridge default (24 hours) if nil
	MaxEventAge awscdk.Duration

	// Enable Dead Letter Queue for failed events
	// If true, creates an SQS queue to capture events that fail after all retries
	// Optional: defaults to false if nil
	EnableDLQ *bool
}

// EventBridgeS3ToLambdaStrategy implements the strategy for S3→Lambda integration
type EventBridgeS3ToLambdaStrategy struct{}

// Build creates a complete S3 → EventBridge → Lambda integration
//
// This strategy encapsulates:
// - EventBridge Rule with S3 event pattern filtering by bucket name and object key
// - Lambda function as target with configurable retry policy
// - Dead Letter Queue for failed invocations (optional)
// - IAM permissions (automatically configured by CDK)
// - EventBridge notifications enabled on S3 bucket
//
// Architecture:
//
//	S3 Bucket (uploads/*) → EventBridge Rule (filter) → Lambda Function
//	                                                      ↓ (on failure)
//	                                                   SQS DLQ (optional)
func (s *EventBridgeS3ToLambdaStrategy) Build(
	scope constructs.Construct,
	id string,
	props EventBridgeIntegrationFactoryProps,
) awsevents.Rule {

	config := props.S3ToLambdaConfig

	// Validate required configuration
	if config.SourceBucket == nil {
		panic("S3ToLambdaConfig.SourceBucket is required")
	}
	if config.TargetLambda == nil {
		panic("S3ToLambdaConfig.TargetLambda is required")
	}

	// 1. Create Dead Letter Queue (if enabled)
	var dlq awssqs.Queue
	if config.EnableDLQ != nil && *config.EnableDLQ {
		dlq = awssqs.NewQueue(scope, jsii.String(id+"-DLQ"), &awssqs.QueueProps{
			QueueName:       jsii.String(id + "-dlq"),
			RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
		})
	}

	// 2. Build event pattern detail configuration
	detailConfig := make(map[string]interface{})

	// Filter by specific bucket name (extracted from instance)
	detailConfig["bucket"] = map[string]interface{}{
		"name": []interface{}{*config.SourceBucket.BucketName()},
	}

	// Filter by object key prefix/suffix (if provided)
	objectFilter := make(map[string]interface{})
	if config.ObjectKeyPrefix != nil {
		objectFilter["prefix"] = *config.ObjectKeyPrefix
	}
	if config.ObjectKeySuffix != nil {
		objectFilter["suffix"] = *config.ObjectKeySuffix
	}

	if len(objectFilter) > 0 {
		detailConfig["object"] = map[string]interface{}{
			"key": []interface{}{objectFilter},
		}
	}

	// 3. Configure event types (default to "Object Created")
	eventTypes := config.EventTypes
	if eventTypes == nil || len(eventTypes) == 0 {
		eventTypes = []string{"Object Created"}
	}

	detailTypes := make([]*string, len(eventTypes))
	for i, et := range eventTypes {
		detailTypes[i] = jsii.String(et)
	}

	// 4. Create EventBridge Rule with S3 event pattern
	rule := awsevents.NewRule(scope, jsii.String(id+"-Rule"), &awsevents.RuleProps{
		RuleName:    jsii.String(id + "-rule"),
		Description: jsii.String("Routes S3 events from " + *config.SourceBucket.BucketName() + " to Lambda " + *config.TargetLambda.FunctionName()),
		EventPattern: &awsevents.EventPattern{
			Source:     jsii.Strings("aws.s3"),
			DetailType: &detailTypes,
			Detail:     &detailConfig,
		},
	})

	// 5. Configure Lambda target with retry policy
	targetProps := &awseventstargets.LambdaFunctionProps{}

	if config.MaxRetryAttempts != nil {
		targetProps.RetryAttempts = jsii.Number(*config.MaxRetryAttempts)
	}

	if config.MaxEventAge != nil {
		targetProps.MaxEventAge = config.MaxEventAge
	}

	if dlq != nil {
		targetProps.DeadLetterQueue = dlq
	}

	// 6. Add Lambda as target to the rule
	// CDK automatically configures IAM permissions for EventBridge to invoke Lambda
	rule.AddTarget(awseventstargets.NewLambdaFunction(config.TargetLambda, targetProps))

	// 7. Enable EventBridge notifications on the S3 bucket
	// This is CRITICAL - without this, S3 won't emit events to EventBridge
	config.SourceBucket.EnableEventBridgeNotification()

	return rule
}
