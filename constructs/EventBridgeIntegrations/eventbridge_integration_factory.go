package eventbridgeintegrations

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/constructs-go/constructs/v10"
)

// IntegrationType defines the type of EventBridge integration to create
type IntegrationType string

const (
	// IntegrationTypeS3ToLambda creates an integration from S3 bucket events to Lambda function
	IntegrationTypeS3ToLambda IntegrationType = "S3_TO_LAMBDA"
)

// EventBridgeIntegrationFactoryProps defines properties for creating an EventBridge integration via Factory
type EventBridgeIntegrationFactoryProps struct {
	// Type of integration to create (REQUIRED)
	IntegrationType IntegrationType

	// Configuration specific to S3ToLambda integration
	S3ToLambdaConfig *S3ToLambdaConfig
}

// NewEventBridgeIntegrationFactory creates an EventBridge integration using the Factory + Strategy pattern
//
// This factory selects the appropriate strategy based on IntegrationType and delegates
// integration creation to the specialized strategy implementation.
//
// Example usage:
//
//	integration := eventbridgeintegrations.NewEventBridgeIntegrationFactory(
//	    stack,
//	    "S3ToLambda",
//	    eventbridgeintegrations.EventBridgeIntegrationFactoryProps{
//	        IntegrationType: eventbridgeintegrations.IntegrationTypeS3ToLambda,
//	        S3ToLambdaConfig: &eventbridgeintegrations.S3ToLambdaConfig{
//	            SourceBucket:     bucket,
//	            TargetLambda:     lambda,
//	            ObjectKeyPrefix:  jsii.String("uploads/"),
//	            MaxRetryAttempts: jsii.Number(4),
//	            EnableDLQ:        jsii.Bool(true),
//	        },
//	    })
func NewEventBridgeIntegrationFactory(
	scope constructs.Construct,
	id string,
	props EventBridgeIntegrationFactoryProps,
) awsevents.Rule {
	var strategy EventBridgeIntegrationStrategy

	// Select strategy based on integration type
	switch props.IntegrationType {
	case IntegrationTypeS3ToLambda:
		if props.S3ToLambdaConfig == nil {
			panic("S3ToLambdaConfig is required when IntegrationType is S3_TO_LAMBDA")
		}
		strategy = &EventBridgeS3ToLambdaStrategy{}

	default:
		panic(fmt.Sprintf("Unsupported IntegrationType: %s", props.IntegrationType))
	}

	// Delegate integration creation to selected strategy
	return strategy.Build(scope, id, props)
}
