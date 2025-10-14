package guardduty

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/constructs-go/constructs/v10"
)

// GuardDutyType defines the available threat detection strategy types.
type GuardDutyType string

const (
	// GuardDutyTypeBasic provides foundational threat detection (CloudTrail, VPC Flow, DNS).
	// Cost: ~$4-8/month | Use: Dev/test, small workloads, cost-conscious deployments
	GuardDutyTypeBasic GuardDutyType = "BASIC"

	// GuardDutyTypeComprehensive enables all protection features (S3, EKS, Malware, RDS, Lambda).
	// Cost: ~$30-100/month | Use: Production, compliance requirements, maximum security
	GuardDutyTypeComprehensive GuardDutyType = "COMPREHENSIVE"

	// GuardDutyTypeCustom allows granular control over individual features.
	// Cost: Variable | Use: Specific compliance needs, phased rollout, custom requirements
	GuardDutyTypeCustom GuardDutyType = "CUSTOM"
)

// GuardDutyFactoryProps defines configuration properties for GuardDuty detector creation.
type GuardDutyFactoryProps struct {
	// DetectorType specifies which threat detection strategy to use.
	// Required. Options: GuardDutyTypeBasic, GuardDutyTypeComprehensive, GuardDutyTypeCustom
	DetectorType GuardDutyType

	// FindingPublishingFrequency controls how often findings are exported to CloudWatch/EventBridge.
	// Optional. Values: "FIFTEEN_MINUTES", "ONE_HOUR", "SIX_HOURS" (default varies by strategy)
	// Lower frequency = faster alerts but higher EventBridge costs
	FindingPublishingFrequency *string

	// Tags to apply to the GuardDuty detector resource.
	// Optional. Format: []*CfnDetector_TagItemProperty
	Tags *[]*awsguardduty.CfnDetector_TagItemProperty

	// --- Custom Strategy Options (only used when DetectorType = GuardDutyTypeCustom) ---

	// EnableS3Protection monitors S3 data access patterns and bucket policy changes.
	// Optional (Custom strategy only). Cost: ~$0.50 per million CloudTrail data events
	EnableS3Protection *bool

	// EnableEKSProtection monitors Kubernetes audit logs for container threats.
	// Optional (Custom strategy only). Cost: ~$0.50 per million audit log entries
	EnableEKSProtection *bool

	// EnableEKSRuntimeMonitoring enables deep runtime visibility for EKS workloads.
	// Optional (Custom strategy only). Requires EnableEKSProtection=true
	// Cost: ~$1.20 per node per month
	EnableEKSRuntimeMonitoring *bool

	// EnableMalwareProtection scans EBS volumes attached to EC2 instances.
	// Optional (Custom strategy only). Cost: ~$1.00 per GB scanned
	EnableMalwareProtection *bool

	// EnableRDSProtection detects anomalous database login activity.
	// Optional (Custom strategy only). Cost: ~$0.10 per million login events
	EnableRDSProtection *bool

	// EnableLambdaProtection monitors Lambda function network activity for C2 communication.
	// Optional (Custom strategy only). Cost: ~$0.40 per million invocations
	EnableLambdaProtection *bool

	// EnableRuntimeMonitoring enables EC2 and Fargate runtime threat detection.
	// Optional (Custom strategy only). Cost: ~$1.15 per node per month
	EnableRuntimeMonitoring *bool

	// EnableEC2AgentManagement automatically deploys GuardDuty security agent to EC2 instances.
	// Optional (Custom strategy only). Requires EnableRuntimeMonitoring=true
	EnableEC2AgentManagement *bool

	// EnableFargateAgentManagement automatically deploys GuardDuty agent to ECS Fargate tasks.
	// Optional (Custom strategy only). Requires EnableRuntimeMonitoring=true
	EnableFargateAgentManagement *bool
}

// NewGuardDutyDetector creates a GuardDuty threat detection detector using the Factory pattern.
// Automatically selects and applies the appropriate strategy based on DetectorType.
//
// Strategy Selection:
//   - GuardDutyTypeBasic: Foundational monitoring only (lowest cost)
//   - GuardDutyTypeComprehensive: All features enabled (maximum protection)
//   - GuardDutyTypeCustom: User-defined feature set (flexible configuration)
//
// Returns: awsguardduty.CfnDetector - The created detector construct
//
// Example (Basic):
//
//	detector := guardduty.NewGuardDutyDetector(stack, "BasicDetector",
//	    guardduty.GuardDutyFactoryProps{
//	        DetectorType: guardduty.GuardDutyTypeBasic,
//	    })
//
// Example (Comprehensive):
//
//	detector := guardduty.NewGuardDutyDetector(stack, "ProdDetector",
//	    guardduty.GuardDutyFactoryProps{
//	        DetectorType: guardduty.GuardDutyTypeComprehensive,
//	        FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
//	    })
//
// Example (Custom - S3 + EKS only):
//
//	detector := guardduty.NewGuardDutyDetector(stack, "CustomDetector",
//	    guardduty.GuardDutyFactoryProps{
//	        DetectorType: guardduty.GuardDutyTypeCustom,
//	        EnableS3Protection: jsii.Bool(true),
//	        EnableEKSProtection: jsii.Bool(true),
//	        EnableEKSRuntimeMonitoring: jsii.Bool(true),
//	        FindingPublishingFrequency: jsii.String("ONE_HOUR"),
//	    })
func NewGuardDutyDetector(scope constructs.Construct, id string, props GuardDutyFactoryProps) awsguardduty.CfnDetector {

	var strategy GuardDutyStrategy

	// Select strategy based on detector type
	switch props.DetectorType {
	case GuardDutyTypeBasic:
		strategy = &GuardDutyBasicStrategy{}

	case GuardDutyTypeComprehensive:
		strategy = &GuardDutyComprehensiveStrategy{}

	case GuardDutyTypeCustom:
		strategy = &GuardDutyCustomStrategy{}

	default:
		panic(fmt.Sprintf("Unsupported GuardDuty detector type: %s. Valid options: BASIC, COMPREHENSIVE, CUSTOM", props.DetectorType))
	}

	// Execute strategy to build detector
	return strategy.Build(scope, id, props)
}
