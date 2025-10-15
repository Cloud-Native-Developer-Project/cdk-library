package guardduty

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// GuardDutyDataProtectionStrategy implements threat detection focused on data services.
// Monitors S3, databases, serverless functions, and Kubernetes without runtime agents.
//
// Use Cases:
// - S3 data lakes and sensitive bucket monitoring
// - Serverless architectures (Lambda-heavy workloads)
// - Database security (RDS login monitoring)
// - Kubernetes API abuse detection (EKS audit logs only)
// - Compliance requirements without agent deployment
//
// Features Enabled:
// - Foundational: CloudTrail, VPC Flow Logs, DNS logs
// - S3 Data Events: Monitors unusual S3 access patterns and bucket policy changes
// - EKS Audit Logs: Kubernetes API monitoring (NO runtime agents)
// - EBS Malware Protection: Agentless snapshot scanning
// - RDS Login Events: Detects brute force and anomalous database logins
// - Lambda Network Logs: Monitors serverless function network activity
// - 15-minute finding frequency for rapid incident response
//
// NOT Included:
// - EC2 Runtime Monitoring (use GuardDutyTypeEC2Runtime when needed)
// - EKS Runtime Monitoring (use GuardDutyTypeEKSRuntime when needed)
// - ECS Runtime Monitoring (use GuardDutyTypeECSRuntime when needed)
//
// Estimated Cost: ~$15-50/month (varies with S3/Lambda/RDS usage)
type GuardDutyDataProtectionStrategy struct{}

// Build creates a GuardDuty detector focused on data protection services.
// Ideal for S3-centric architectures, serverless, and database workloads.
func (s *GuardDutyDataProtectionStrategy) Build(scope constructs.Construct, id string, props GuardDutyFactoryProps) awsguardduty.CfnDetector {

	// Set aggressive finding frequency for production environments
	findingFrequency := props.FindingPublishingFrequency
	if findingFrequency == nil {
		findingFrequency = jsii.String("FIFTEEN_MINUTES")
	}

	// Create detector with comprehensive protection enabled
	detector := awsguardduty.NewCfnDetector(scope, jsii.String(id), &awsguardduty.CfnDetectorProps{
		Enable: jsii.Bool(true),

		// Enable advanced protection features using the Features API
		Features: &[]interface{}{
			// S3 data events monitoring (detects unusual S3 access patterns)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("S3_DATA_EVENTS"),
				Status: jsii.String("ENABLED"),
			},

			// EKS audit logs (Kubernetes API monitoring without runtime agents)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("EKS_AUDIT_LOGS"),
				Status: jsii.String("ENABLED"),
			},

			// EBS malware protection (agentless snapshot scanning)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("EBS_MALWARE_PROTECTION"),
				Status: jsii.String("ENABLED"),
			},

			// RDS login activity monitoring (detects brute force, anomalous logins)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("RDS_LOGIN_EVENTS"),
				Status: jsii.String("ENABLED"),
			},

			// Lambda network activity monitoring (detects C2 communication)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("LAMBDA_NETWORK_LOGS"),
				Status: jsii.String("ENABLED"),
			},

			// Note: Runtime monitoring (EC2_RUNTIME_MONITORING, EKS_RUNTIME_MONITORING)
			// excluded to avoid conflicts. Use specific strategies when needed.
		},

		// Rapid finding publication for quick incident response
		FindingPublishingFrequency: findingFrequency,

		// Add resource tags if provided
		Tags: props.Tags,
	})

	return detector
}
