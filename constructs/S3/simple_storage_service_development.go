package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// SimpleStorageServiceDevelopmentStrategy implements S3 bucket optimized for development/testing
// This strategy is designed for dev/test environments with cost efficiency and easy cleanup
//
// Security Model:
// - Private bucket with basic security
// - S3_MANAGED encryption
// - Minimal monitoring for cost reduction
//
// Use Cases:
// - Development/testing environments
// - Temporary storage
// - CI/CD artifact storage
// - Sandbox environments
//
// Cost Optimization:
// - Auto-delete on stack removal
// - 30-day expiration lifecycle
// - No versioning (reduces costs)
// - Minimal monitoring
type SimpleStorageServiceDevelopmentStrategy struct{}

// Build creates an S3 bucket configured for development/testing
func (s *SimpleStorageServiceDevelopmentStrategy) Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {

	bucketProps := &awss3.BucketProps{
		// Basic Configuration - Easy cleanup
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,  // Easy cleanup for dev
		AutoDeleteObjects: jsii.Bool(true),

		// Minimal Security for Development
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BucketKeyEnabled:  jsii.Bool(false),  // Reduce complexity
		EnforceSSL:        jsii.Bool(true),   // Required for TLS version
		MinimumTLSVersion: jsii.Number(1.2),

		// Object Ownership
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_ENFORCED,

		// Basic Data Protection - No versioning to reduce costs
		Versioned: jsii.Bool(false),

		// Simple Lifecycle for Cost Control
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:         jsii.String("DevCleanup"),
				Enabled:    jsii.Bool(true),
				Expiration: awscdk.Duration_Days(jsii.Number(30)),  // Auto-cleanup after 30 days
			},
		},

		// Minimal Monitoring - Reduce costs
		EventBridgeEnabled: jsii.Bool(false),

		// Development Features
		TransferAcceleration: jsii.Bool(false),
		Cors: &[]*awss3.CorsRule{
			{
				AllowedOrigins: &[]*string{
					jsii.String("*"),  // Permissive for development
				},
				AllowedMethods: &[]awss3.HttpMethods{
					awss3.HttpMethods_GET,
					awss3.HttpMethods_POST,
					awss3.HttpMethods_PUT,
					awss3.HttpMethods_DELETE,
					awss3.HttpMethods_HEAD,
				},
				AllowedHeaders: &[]*string{
					jsii.String("*"),
				},
				MaxAge: jsii.Number(3000),
			},
		},
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
