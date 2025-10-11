package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// SimpleStorageServiceCloudfrontOriginStrategy implements S3 bucket optimized for CloudFront origin
// This is the RECOMMENDED approach for static websites instead of S3 website hosting
//
// Architecture: Private S3 bucket → CloudFront with OAC (Origin Access Control)
//
// Security Model:
// - Bucket is completely private (no public access)
// - CloudFront accesses bucket via OAC (modern replacement for OAI)
// - HTTPS enforcement via CloudFront
//
// Use Cases:
// - Static websites (React, Vue, Angular, etc.)
// - Single Page Applications (SPAs)
// - JAMstack sites
// - Static asset hosting for web applications
type SimpleStorageServiceCloudfrontOriginStrategy struct{}

// Build creates an S3 bucket configured as a CloudFront origin with OAC
func (s *SimpleStorageServiceCloudfrontOriginStrategy) Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {

	// =============================================================================
	// BUCKET CONFIGURATION - CloudFront Origin Optimized
	// =============================================================================

	bucketProps := &awss3.BucketProps{
		// Basic Configuration
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY, // Easy cleanup for dev/test
		AutoDeleteObjects: jsii.Bool(true),              // Auto-delete on stack removal

		// Security - Private bucket, CloudFront handles public access
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(), // Completely private
		Encryption:        awss3.BucketEncryption_S3_MANAGED,    // SSE-S3 encryption
		BucketKeyEnabled:  jsii.Bool(true),                      // Reduce KMS costs if using KMS
		EnforceSSL:        jsii.Bool(true),                      // HTTPS only
		MinimumTLSVersion: jsii.Number(1.2),                     // TLS 1.2 minimum

		// Object Ownership - Bucket owner enforced (recommended)
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_ENFORCED,

		// Versioning - Enabled for deployment rollbacks
		// With 1-day lifecycle expiration, storage cost impact is minimal
		Versioned: jsii.Bool(true),

		// Lifecycle Management - Clean up old versions
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:      jsii.String("WebContentCleanup"),
				Enabled: jsii.Bool(true),
				// Delete non-current versions after 1 day
				NoncurrentVersionExpiration: awscdk.Duration_Days(jsii.Number(1)),
			},
		},

		// Monitoring - EventBridge for automated workflows
		EventBridgeEnabled: jsii.Bool(true), // Trigger Lambda on object create/delete

		// Performance - CloudFront handles these, so disabled
		TransferAcceleration: jsii.Bool(false), // CloudFront provides global acceleration

		// Website Hosting - DISABLED (CloudFront handles routing)
		// Never enable S3 website hosting when using CloudFront with OAC
		WebsiteIndexDocument: nil,
		WebsiteErrorDocument: nil,
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
