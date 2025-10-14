package guardduty

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/constructs-go/constructs/v10"
)

// GuardDutyStrategy defines the interface that all GuardDuty strategies must implement.
// This contract ensures consistent detector creation across different threat detection strategies.
type GuardDutyStrategy interface {
	// Build creates and configures a GuardDuty detector with strategy-specific settings.
	// Returns the created CfnDetector construct.
	Build(scope constructs.Construct, id string, props GuardDutyFactoryProps) awsguardduty.CfnDetector
}
