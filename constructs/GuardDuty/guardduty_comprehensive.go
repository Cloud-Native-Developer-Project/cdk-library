package guardduty

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// GuardDutyComprehensiveStrategy implements full-spectrum threat detection.
// Enables all GuardDuty protection features for maximum security coverage across AWS services.
//
// Use Cases:
// - Production workloads requiring maximum security posture
// - Compliance requirements (PCI DSS, HIPAA, SOC 2)
// - Environments with S3 data lakes, EKS clusters, or sensitive EC2 workloads
// - Organizations requiring runtime threat detection and malware scanning
//
// Features Enabled:
// - Foundational: CloudTrail, VPC Flow Logs, DNS logs
// - S3 Protection: Monitors S3 data access patterns and bucket policy changes
// - EKS Protection: Kubernetes audit log monitoring for container threats
// - Malware Protection: Scans EBS volumes and EC2 instances for malware
// - RDS Protection: Detects suspicious database login activity
// - Lambda Protection: Monitors serverless function network activity
// - Runtime Monitoring: Deep visibility into EC2 and container runtime behavior
// - 15-minute finding frequency for rapid incident response
//
// Estimated Cost: ~$30-100/month (varies with workload size and data volume)
type GuardDutyComprehensiveStrategy struct{}

// Build creates a GuardDuty detector with all protection features enabled.
// Provides enterprise-grade threat detection across all supported AWS services.
func (s *GuardDutyComprehensiveStrategy) Build(scope constructs.Construct, id string, props GuardDutyFactoryProps) awsguardduty.CfnDetector {

	// Set aggressive finding frequency for production environments
	findingFrequency := props.FindingPublishingFrequency
	if findingFrequency == nil {
		findingFrequency = jsii.String("FIFTEEN_MINUTES")
	}

	// Create detector with comprehensive protection enabled
	detector := awsguardduty.NewCfnDetector(scope, jsii.String(id), &awsguardduty.CfnDetectorProps{
		Enable: jsii.Bool(true),

		// Enable all data sources for maximum coverage
		DataSources: &awsguardduty.CfnDetector_CFNDataSourceConfigurationsProperty{
			// S3 data event monitoring (detects unusual bucket access patterns)
			S3Logs: &awsguardduty.CfnDetector_CFNS3LogsConfigurationProperty{
				Enable: jsii.Bool(true),
			},

			// EKS audit log monitoring (detects Kubernetes API abuse, privilege escalation)
			Kubernetes: &awsguardduty.CfnDetector_CFNKubernetesConfigurationProperty{
				AuditLogs: &awsguardduty.CfnDetector_CFNKubernetesAuditLogsConfigurationProperty{
					Enable: jsii.Bool(true),
				},
			},

			// Malware protection for EC2 instances (agentless EBS volume scanning)
			MalwareProtection: &awsguardduty.CfnDetector_CFNMalwareProtectionConfigurationProperty{
				ScanEc2InstanceWithFindings: &awsguardduty.CfnDetector_CFNScanEc2InstanceWithFindingsConfigurationProperty{
					EbsVolumes: jsii.Bool(true),
				},
			},
		},

		// Enable advanced protection features using the Features API
		Features: &[]interface{}{
			// S3 malware protection (scans newly uploaded objects)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("S3_DATA_EVENTS"),
				Status: jsii.String("ENABLED"),
			},

			// EKS runtime monitoring (detects container runtime threats)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("EKS_AUDIT_LOGS"),
				Status: jsii.String("ENABLED"),
			},

			// EKS runtime monitoring with addon deployment
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("EKS_RUNTIME_MONITORING"),
				Status: jsii.String("ENABLED"),
				AdditionalConfiguration: &[]interface{}{
					&awsguardduty.CfnDetector_CFNFeatureAdditionalConfigurationProperty{
						Name:   jsii.String("EKS_ADDON_MANAGEMENT"),
						Status: jsii.String("ENABLED"),
					},
				},
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

			// EC2 runtime monitoring (detects file and process threats)
			&awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("RUNTIME_MONITORING"),
				Status: jsii.String("ENABLED"),
				AdditionalConfiguration: &[]interface{}{
					// Enable GuardDuty agent auto-management
					&awsguardduty.CfnDetector_CFNFeatureAdditionalConfigurationProperty{
						Name:   jsii.String("EC2_AGENT_MANAGEMENT"),
						Status: jsii.String("ENABLED"),
					},
					// Enable ECS Fargate agent deployment
					&awsguardduty.CfnDetector_CFNFeatureAdditionalConfigurationProperty{
						Name:   jsii.String("ECS_FARGATE_AGENT_MANAGEMENT"),
						Status: jsii.String("ENABLED"),
					},
				},
			},
		},

		// Rapid finding publication for quick incident response
		FindingPublishingFrequency: findingFrequency,

		// Add resource tags if provided
		Tags: props.Tags,
	})

	return detector
}
