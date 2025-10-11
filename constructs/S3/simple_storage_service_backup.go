package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// SimpleStorageServiceBackupStrategy implements S3 bucket optimized for backup and disaster recovery
// This strategy is designed for database backups, application backups, and DR scenarios
//
// Security Model:
// - Private bucket with KMS encryption
// - Object Lock with GOVERNANCE retention (90 days)
// - TLS 1.2 enforced
//
// Use Cases:
// - Database backups (RDS, DynamoDB, etc.)
// - Application state backups
// - Disaster recovery storage
// - Compliance archival
//
// Cost Optimization:
// - Aggressive lifecycle policies to Glacier/Deep Archive
// - Intelligent Tiering enabled
// - 10-year expiration policy
type SimpleStorageServiceBackupStrategy struct{}

// Build creates an S3 bucket configured for backup and disaster recovery
func (s *SimpleStorageServiceBackupStrategy) Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {

	bucketProps := &awss3.BucketProps{
		// Basic Configuration
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,  // Backups should be retained
		AutoDeleteObjects: jsii.Bool(false),

		// Enhanced Security for Backups
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_KMS_MANAGED,  // Enhanced security
		BucketKeyEnabled:  jsii.Bool(true),
		EnforceSSL:        jsii.Bool(true),
		MinimumTLSVersion: jsii.Number(1.2),

		// Object Ownership
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_ENFORCED,

		// Data Protection & Compliance
		Versioned:         jsii.Bool(true),
		ObjectLockEnabled: jsii.Bool(true),
		ObjectLockDefaultRetention: awss3.ObjectLockRetention_Governance(
			awscdk.Duration_Days(jsii.Number(90)),  // 3 months minimum retention
		),

		// Aggressive Cost Optimization for Backups
		IntelligentTieringConfigurations: &[]*awss3.IntelligentTieringConfiguration{
			{
				Name:                      jsii.String("EntireBucket"),
				ArchiveAccessTierTime:     awscdk.Duration_Days(jsii.Number(90)),
				DeepArchiveAccessTierTime: awscdk.Duration_Days(jsii.Number(180)),
			},
		},

		// Lifecycle Management for Backup Retention
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:      jsii.String("BackupRetention"),
				Enabled: jsii.Bool(true),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)),  // Move to IA after 1 month
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)),  // Archive after 3 months
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),  // Deep archive after 1 year
					},
				},
				Expiration: awscdk.Duration_Days(jsii.Number(3650)),  // 10 years retention
			},
		},

		// Comprehensive Monitoring
		ServerAccessLogsPrefix: jsii.String("access-logs/"),
		Inventories: &[]*awss3.Inventory{
			{
				Enabled:               jsii.Bool(true),
				IncludeObjectVersions: awss3.InventoryObjectVersion_CURRENT,
				Frequency:             awss3.InventoryFrequency_DAILY,
			},
		},
		Metrics: &[]*awss3.BucketMetrics{
			{
				Id: jsii.String("EntireBucket"),
			},
		},
		EventBridgeEnabled: jsii.Bool(true),  // For backup automation workflows

		// Performance
		TransferAcceleration: jsii.Bool(false),  // Not typically needed for backups
	}

	// Apply custom overrides
	if props.RemovalPolicy != "" {
		switch props.RemovalPolicy {
		case "retain":
			bucketProps.RemovalPolicy = awscdk.RemovalPolicy_RETAIN
		case "destroy":
			bucketProps.RemovalPolicy = awscdk.RemovalPolicy_DESTROY
		case "retain_on_update_or_delete":
			bucketProps.RemovalPolicy = awscdk.RemovalPolicy_RETAIN_ON_UPDATE_OR_DELETE
		}
	}

	if props.AutoDeleteObjects != nil {
		bucketProps.AutoDeleteObjects = jsii.Bool(*props.AutoDeleteObjects)
	}

	bucket := awss3.NewBucket(scope, jsii.String(id), bucketProps)

	return bucket
}
