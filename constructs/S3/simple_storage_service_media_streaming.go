package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// SimpleStorageServiceMediaStreamingStrategy implements S3 bucket optimized for media streaming
// This strategy is designed for video/audio streaming, CDN origin, and high-throughput content delivery
//
// Security Model:
// - Private bucket (use CloudFront with OAC)
// - S3_MANAGED encryption (KMS adds latency for streaming)
// - TLS 1.2 enforced
// - CORS enabled for player applications
//
// Use Cases:
// - Video/audio streaming (HLS, DASH)
// - Media CDN origin
// - High-throughput content delivery
// - Video-on-demand platforms
//
// Performance Optimization:
// - CORS configured for specific player domains
// - Intelligent Tiering for cost optimization
// - Lifecycle policies for content archival
type SimpleStorageServiceMediaStreamingStrategy struct{}

// Build creates an S3 bucket configured for media streaming
func (s *SimpleStorageServiceMediaStreamingStrategy) Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {

	bucketProps := &awss3.BucketProps{
		// Basic Configuration
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
		AutoDeleteObjects: jsii.Bool(false),

		// Security - Balanced for content delivery
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),  // Use CloudFront with OAC
		Encryption:        awss3.BucketEncryption_S3_MANAGED,    // KMS adds latency for streaming
		BucketKeyEnabled:  jsii.Bool(true),
		EnforceSSL:        jsii.Bool(true),
		MinimumTLSVersion: jsii.Number(1.2),

		// Object Ownership
		ObjectOwnership: awss3.ObjectOwnership_BUCKET_OWNER_ENFORCED,

		// Data Protection - Media files are typically immutable
		Versioned: jsii.Bool(false),

		// Cost Optimization for Media
		IntelligentTieringConfigurations: &[]*awss3.IntelligentTieringConfiguration{
			{
				Name:                      jsii.String("EntireBucket"),
				ArchiveAccessTierTime:     awscdk.Duration_Days(jsii.Number(90)),
				DeepArchiveAccessTierTime: awscdk.Duration_Days(jsii.Number(180)),
			},
		},

		// Lifecycle Management for Media Content
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:      jsii.String("MediaContentLifecycle"),
				Enabled: jsii.Bool(true),
				Prefix:  jsii.String("videos/"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)),  // Popular content stays hot longer
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),  // Archive old content
					},
				},
			},
		},

		// Performance Optimization for Streaming
		TransferAcceleration: jsii.Bool(false),  // Use CloudFront instead
		Cors: &[]*awss3.CorsRule{
			{
				AllowedOrigins: &[]*string{
					jsii.String("https://player.example.com"),  // Specific player domains
					jsii.String("https://*.cdn.example.com"),   // CDN subdomains
				},
				AllowedMethods: &[]awss3.HttpMethods{
					awss3.HttpMethods_GET,
					awss3.HttpMethods_HEAD,
				},
				AllowedHeaders: &[]*string{
					jsii.String("Range"),
					jsii.String("Authorization"),
					jsii.String("Content-Type"),
				},
				MaxAge: jsii.Number(3000),
			},
		},

		// Monitoring for Performance
		Metrics: &[]*awss3.BucketMetrics{
			{
				Id:     jsii.String("VideosMetrics"),
				Prefix: jsii.String("videos/"),  // Monitor video performance
			},
		},
		EventBridgeEnabled: jsii.Bool(true),  // For content processing workflows
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
