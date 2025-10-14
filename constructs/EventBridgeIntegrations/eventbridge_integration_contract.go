package eventbridgeintegrations

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/constructs-go/constructs/v10"
)

// EventBridgeIntegrationStrategy defines the contract for EventBridge integration strategies
// Each strategy implements a specific integration pattern (S3→Lambda, S3→SQS, etc.)
type EventBridgeIntegrationStrategy interface {
	Build(scope constructs.Construct, id string, props EventBridgeIntegrationFactoryProps) awsevents.Rule
}
