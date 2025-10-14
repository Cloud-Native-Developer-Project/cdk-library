package guardduty

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// GuardDutyCustomStrategy implements flexible, user-defined threat detection.
// Allows granular control over individual GuardDuty features for custom security requirements.
//
// Use Cases:
// - Organizations with specific compliance requirements (enable only required features)
// - Cost optimization (enable only S3 + EKS protection, disable malware scanning)
// - Phased rollout (start with foundational, gradually enable advanced features)
// - Hybrid environments (enable Lambda monitoring only for serverless workloads)
//
// Customizable Features:
// - S3 data event monitoring
// - EKS audit logs and runtime monitoring
// - EC2 runtime monitoring with agent management
// - EBS malware protection
// - RDS login event monitoring
// - Lambda network logs
// - Custom finding publication frequency
//
// Usage Example:
//   guardduty.NewGuardDutyDetector(stack, "CustomDetector", guardduty.GuardDutyFactoryProps{
//       DetectorType: guardduty.GuardDutyTypeCustom,
//       EnableS3Protection: jsii.Bool(true),
//       EnableEKSProtection: jsii.Bool(true),
//       EnableMalwareProtection: jsii.Bool(false),  // Disabled to reduce cost
//       FindingPublishingFrequency: jsii.String("ONE_HOUR"),
//   })
type GuardDutyCustomStrategy struct{}

// Build creates a GuardDuty detector with user-specified feature configuration.
// Only explicitly enabled features will be activated, providing maximum flexibility.
func (s *GuardDutyCustomStrategy) Build(scope constructs.Construct, id string, props GuardDutyFactoryProps) awsguardduty.CfnDetector {

	// Set default finding frequency if not specified
	findingFrequency := props.FindingPublishingFrequency
	if findingFrequency == nil {
		findingFrequency = jsii.String("SIX_HOURS")
	}

	// Build data sources configuration based on user input
	dataSourcesConfig := &awsguardduty.CfnDetector_CFNDataSourceConfigurationsProperty{}

	// Configure S3 protection if requested
	if props.EnableS3Protection != nil {
		dataSourcesConfig.S3Logs = &awsguardduty.CfnDetector_CFNS3LogsConfigurationProperty{
			Enable: props.EnableS3Protection,
		}
	}

	// Configure EKS protection if requested
	if props.EnableEKSProtection != nil {
		dataSourcesConfig.Kubernetes = &awsguardduty.CfnDetector_CFNKubernetesConfigurationProperty{
			AuditLogs: &awsguardduty.CfnDetector_CFNKubernetesAuditLogsConfigurationProperty{
				Enable: props.EnableEKSProtection,
			},
		}
	}

	// Configure malware protection if requested
	if props.EnableMalwareProtection != nil {
		dataSourcesConfig.MalwareProtection = &awsguardduty.CfnDetector_CFNMalwareProtectionConfigurationProperty{
			ScanEc2InstanceWithFindings: &awsguardduty.CfnDetector_CFNScanEc2InstanceWithFindingsConfigurationProperty{
				EbsVolumes: props.EnableMalwareProtection,
			},
		}
	}

	// Build features configuration based on user input
	features := &[]interface{}{}

	// Add S3 data events if enabled
	if props.EnableS3Protection != nil && *props.EnableS3Protection {
		*features = append(*features, &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
			Name:   jsii.String("S3_DATA_EVENTS"),
			Status: jsii.String("ENABLED"),
		})
	}

	// Add EKS features if enabled
	if props.EnableEKSProtection != nil && *props.EnableEKSProtection {
		*features = append(*features, &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
			Name:   jsii.String("EKS_AUDIT_LOGS"),
			Status: jsii.String("ENABLED"),
		})

		// Add EKS runtime monitoring if specified
		if props.EnableEKSRuntimeMonitoring != nil && *props.EnableEKSRuntimeMonitoring {
			*features = append(*features, &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
				Name:   jsii.String("EKS_RUNTIME_MONITORING"),
				Status: jsii.String("ENABLED"),
				AdditionalConfiguration: &[]interface{}{
					&awsguardduty.CfnDetector_CFNFeatureAdditionalConfigurationProperty{
						Name:   jsii.String("EKS_ADDON_MANAGEMENT"),
						Status: jsii.String("ENABLED"),
					},
				},
			})
		}
	}

	// Add EBS malware protection if enabled
	if props.EnableMalwareProtection != nil && *props.EnableMalwareProtection {
		*features = append(*features, &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
			Name:   jsii.String("EBS_MALWARE_PROTECTION"),
			Status: jsii.String("ENABLED"),
		})
	}

	// Add RDS login events if enabled
	if props.EnableRDSProtection != nil && *props.EnableRDSProtection {
		*features = append(*features, &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
			Name:   jsii.String("RDS_LOGIN_EVENTS"),
			Status: jsii.String("ENABLED"),
		})
	}

	// Add Lambda network logs if enabled
	if props.EnableLambdaProtection != nil && *props.EnableLambdaProtection {
		*features = append(*features, &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
			Name:   jsii.String("LAMBDA_NETWORK_LOGS"),
			Status: jsii.String("ENABLED"),
		})
	}

	// Add Runtime Monitoring if enabled
	if props.EnableRuntimeMonitoring != nil && *props.EnableRuntimeMonitoring {
		runtimeConfig := &awsguardduty.CfnDetector_CFNFeatureConfigurationProperty{
			Name:   jsii.String("RUNTIME_MONITORING"),
			Status: jsii.String("ENABLED"),
		}

		// Add agent management configurations if specified
		additionalConfig := &[]interface{}{}

		if props.EnableEC2AgentManagement != nil && *props.EnableEC2AgentManagement {
			*additionalConfig = append(*additionalConfig, &awsguardduty.CfnDetector_CFNFeatureAdditionalConfigurationProperty{
				Name:   jsii.String("EC2_AGENT_MANAGEMENT"),
				Status: jsii.String("ENABLED"),
			})
		}

		if props.EnableFargateAgentManagement != nil && *props.EnableFargateAgentManagement {
			*additionalConfig = append(*additionalConfig, &awsguardduty.CfnDetector_CFNFeatureAdditionalConfigurationProperty{
				Name:   jsii.String("ECS_FARGATE_AGENT_MANAGEMENT"),
				Status: jsii.String("ENABLED"),
			})
		}

		if len(*additionalConfig) > 0 {
			runtimeConfig.AdditionalConfiguration = additionalConfig
		}

		*features = append(*features, runtimeConfig)
	}

	// Create detector with custom configuration
	detectorProps := &awsguardduty.CfnDetectorProps{
		Enable:                     jsii.Bool(true),
		FindingPublishingFrequency: findingFrequency,
		Tags:                       props.Tags,
	}

	// Only set DataSources if at least one data source was configured
	if dataSourcesConfig.S3Logs != nil || dataSourcesConfig.Kubernetes != nil || dataSourcesConfig.MalwareProtection != nil {
		detectorProps.DataSources = dataSourcesConfig
	}

	// Only set Features if at least one feature was configured
	if len(*features) > 0 {
		detectorProps.Features = features
	}

	detector := awsguardduty.NewCfnDetector(scope, jsii.String(id), detectorProps)

	return detector
}
