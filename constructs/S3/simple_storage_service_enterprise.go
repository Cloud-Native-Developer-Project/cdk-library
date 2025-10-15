package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// SimpleStorageServiceEnterpriseStrategy implements S3 bucket for secure enterprise data
// This strategy is designed for financial data, PII, regulated industries, and compliance
//
// Security Model:
// - Private bucket with KMS encryption
// - Object Lock with COMPLIANCE retention (7 years)
// - TLS 1.3 enforced (highest security)
// - Comprehensive monitoring and auditing
//
// Use Cases:
// - Financial data storage
// - Personal Identifiable Information (PII)
// - Healthcare records (HIPAA compliance)
// - Regulated industries
// - Compliance archival
//
// Compliance:
// - Object Lock COMPLIANCE mode (cannot be bypassed)
// - 7-year retention policy
// - Comprehensive audit logs
// - Encryption at rest and in transit
type SimpleStorageServiceEnterpriseStrategy struct{}

// Build creates an S3 bucket configured for enterprise data with maximum security
func (s *SimpleStorageServiceEnterpriseStrategy) Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {

	bucketProps := &awss3.BucketProps{
		// Basic Configuration
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,  // Enterprise data should never be auto-deleted
		AutoDeleteObjects: jsii.Bool(false),

		// Enhanced Security Configuration
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),  // Never allow public access
		Encryption:        awss3.BucketEncryption_KMS_MANAGED,   // Use KMS for maximum control
		BucketKeyEnabled:  jsii.Bool(true),                      // Reduce KMS costs
		EnforceSSL:        jsii.Bool(true),                      // Force HTTPS
		MinimumTLSVersion: jsii.Number(1.3),                     // Highest TLS version

		// Object Ownership
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_ENFORCED,

		// Data Protection & Compliance - CRITICAL
		Versioned:         jsii.Bool(true),
		ObjectLockEnabled: jsii.Bool(true),
		ObjectLockDefaultRetention: awss3.ObjectLockRetention_Compliance(
			awscdk.Duration_Days(jsii.Number(2555)),  // 7 years retention (cannot be bypassed)
		),

		// Cost Optimization (while maintaining compliance)
		IntelligentTieringConfigurations: &[]*awss3.IntelligentTieringConfiguration{
			{
				Name:                      jsii.String("EntireBucket"),
				ArchiveAccessTierTime:     awscdk.Duration_Days(jsii.Number(90)),
				DeepArchiveAccessTierTime: awscdk.Duration_Days(jsii.Number(180)),
			},
		},
		TransitionDefaultMinimumObjectSize: awss3.TransitionDefaultMinimumObjectSize_ALL_STORAGE_CLASSES_128_K,

		// Lifecycle Management for Compliance
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:      jsii.String("ComplianceArchival"),
				Enabled: jsii.Bool(true),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),  // 1 year to Glacier
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(1095)),  // 3 years to Deep Archive
					},
				},
			},
		},

		// Comprehensive Monitoring & Auditing - REQUIRED for compliance
		// ServerAccessLogs commented out temporarily - KMS encryption conflicts with logging target
		// For production, create a separate logging bucket with S3_MANAGED encryption
		// ServerAccessLogsPrefix: jsii.String("access-logs/"),
		// Inventories commented out temporarily - requires additional destination bucket setup
		// Inventories: &[]*awss3.Inventory{
		// 	{
		// 		Enabled:               jsii.Bool(true),
		// 		IncludeObjectVersions: awss3.InventoryObjectVersion_CURRENT,
		// 		Frequency:             awss3.InventoryFrequency_DAILY,
		// 	},
		// },
		Metrics: &[]*awss3.BucketMetrics{
			{
				Id: jsii.String("EntireBucket"),
			},
		},
		EventBridgeEnabled: jsii.Bool(true),  // For compliance automation

		// Performance - Not priority for enterprise data
		TransferAcceleration: jsii.Bool(false),
	}

	// Apply custom overrides (limited for enterprise buckets)
	if props.RemovalPolicy != "" {
		// Only allow RETAIN for enterprise buckets (security measure)
		if props.RemovalPolicy == "retain" || props.RemovalPolicy == "retain_on_update_or_delete" {
			bucketProps.RemovalPolicy = awscdk.RemovalPolicy_RETAIN
		}
		// Ignore "destroy" for enterprise buckets
	}

	// AutoDeleteObjects should NEVER be true for enterprise buckets
	// Ignore this override for safety

	bucket := awss3.NewBucket(scope, jsii.String(id), bucketProps)

	return bucket
}
