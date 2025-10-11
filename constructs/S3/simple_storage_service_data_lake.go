package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// SimpleStorageServiceDataLakeStrategy implements S3 bucket optimized for data lake analytics
// This strategy is designed for big data analytics, data science, and batch processing workloads
//
// Architecture: Data Lake pattern with multi-tier storage and lifecycle management
//
// Security Model:
// - Private bucket with KMS encryption
// - TLS 1.2 enforced
// - Comprehensive monitoring and auditing
//
// Use Cases:
// - Big data analytics (Athena, EMR, Redshift Spectrum)
// - Data science workloads
// - Batch processing pipelines
// - Data warehousing
//
// Cost Optimization:
// - Intelligent Tiering enabled
// - Aggressive lifecycle policies (raw-data, processed-data)
// - Transition to Glacier/Deep Archive
type SimpleStorageServiceDataLakeStrategy struct{}

// Build creates an S3 bucket configured for data lake analytics
func (s *SimpleStorageServiceDataLakeStrategy) Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {

	// =============================================================================
	// BUCKET CONFIGURATION - Data Lake Optimized
	// =============================================================================

	bucketProps := &awss3.BucketProps{
		// Basic Configuration
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,  // Data lakes should be retained
		AutoDeleteObjects: jsii.Bool(false),             // Prevent accidental deletion

		// Security - Enhanced for analytics compliance
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_KMS_MANAGED,  // Better for analytics compliance
		BucketKeyEnabled:  jsii.Bool(true),                     // Reduce KMS costs
		EnforceSSL:        jsii.Bool(true),
		MinimumTLSVersion: jsii.Number(1.2),

		// Object Ownership
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_ENFORCED,

		// Data Protection
		Versioned: jsii.Bool(true),  // Track data changes

		// Cost Optimization - Critical for data lakes
		IntelligentTieringConfigurations: &[]*awss3.IntelligentTieringConfiguration{
			{
				Name:                      jsii.String("EntireBucket"),
				ArchiveAccessTierTime:     awscdk.Duration_Days(jsii.Number(90)),
				DeepArchiveAccessTierTime: awscdk.Duration_Days(jsii.Number(180)),
			},
		},
		TransitionDefaultMinimumObjectSize: awss3.TransitionDefaultMinimumObjectSize_ALL_STORAGE_CLASSES_128_K,

		// Comprehensive Lifecycle Management
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:      jsii.String("DataLakeLifecycle"),
				Enabled: jsii.Bool(true),
				Prefix:  jsii.String("raw-data/"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)),
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)),
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),
					},
				},
			},
			{
				Id:      jsii.String("ProcessedDataLifecycle"),
				Enabled: jsii.Bool(true),
				Prefix:  jsii.String("processed-data/"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(7)),  // Faster transition for processed data
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)),
					},
				},
			},
		},

		// Enhanced Monitoring & Analytics
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
				Id:     jsii.String("EntireBucket"),
				Prefix: jsii.String("analytics/"),
			},
		},
		EventBridgeEnabled: jsii.Bool(true),  // For data pipeline automation

		// Performance
		TransferAcceleration: jsii.Bool(false),  // Usually not needed for batch processing
	}

	// Apply custom overrides if provided
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

	// Create and return the bucket
	bucket := awss3.NewBucket(scope, jsii.String(id), bucketProps)

	return bucket
}
