package guardduty

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// GuardDutyBasicStrategy implements foundational threat detection.
// Monitors CloudTrail management events, VPC Flow Logs, and DNS logs for basic security monitoring.
//
// Use Cases:
// - Small to medium workloads with standard security requirements
// - Development and testing environments requiring baseline protection
// - Cost-conscious deployments needing essential threat detection
//
// Features:
// - CloudTrail event monitoring (unauthorized API calls, IAM changes)
// - VPC Flow Logs analysis (unusual network traffic patterns)
// - DNS query monitoring (domain reputation, command and control detection)
// - 6-hour finding publication frequency (balanced cost/visibility)
//
// Estimated Cost: ~$4-8/month (based on 10-50 CloudTrail events/min)
type GuardDutyBasicStrategy struct{}

// Build creates a GuardDuty detector with foundational data source monitoring.
// No additional features (S3, EKS, Malware) are enabled to minimize cost.
func (s *GuardDutyBasicStrategy) Build(scope constructs.Construct, id string, props GuardDutyFactoryProps) awsguardduty.CfnDetector {

	// Set default finding frequency if not specified
	findingFrequency := props.FindingPublishingFrequency
	if findingFrequency == nil {
		findingFrequency = jsii.String("SIX_HOURS")
	}

	// Create detector with only foundational data sources (no additional features)
	detector := awsguardduty.NewCfnDetector(scope, jsii.String(id), &awsguardduty.CfnDetectorProps{
		Enable: jsii.Bool(true),

		// Foundational data sources are enabled by default:
		// - CloudTrail Management Events
		// - VPC Flow Logs
		// - DNS Logs
		DataSources: &awsguardduty.CfnDetector_CFNDataSourceConfigurationsProperty{
			// S3 Logs explicitly disabled (use Comprehensive strategy for S3 protection)
			S3Logs: &awsguardduty.CfnDetector_CFNS3LogsConfigurationProperty{
				Enable: jsii.Bool(false),
			},
			// Kubernetes audit logs disabled (use Comprehensive strategy for EKS protection)
			Kubernetes: &awsguardduty.CfnDetector_CFNKubernetesConfigurationProperty{
				AuditLogs: &awsguardduty.CfnDetector_CFNKubernetesAuditLogsConfigurationProperty{
					Enable: jsii.Bool(false),
				},
			},
			// Malware protection disabled (use Comprehensive strategy for malware scanning)
			MalwareProtection: &awsguardduty.CfnDetector_CFNMalwareProtectionConfigurationProperty{
				ScanEc2InstanceWithFindings: &awsguardduty.CfnDetector_CFNScanEc2InstanceWithFindingsConfigurationProperty{
					EbsVolumes: jsii.Bool(false),
				},
			},
		},

		// Export findings every 6 hours (cost-effective for basic monitoring)
		FindingPublishingFrequency: findingFrequency,

		// Add resource tags if provided
		Tags: props.Tags,
	})

	return detector
}
