package s3

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
)

// BucketType defines the type of S3 bucket to create
type BucketType string

const (
	// BucketTypeCloudfrontOAC creates a bucket optimized for CloudFront origin with OAC
	BucketTypeCloudfrontOAC BucketType = "CLOUDFRONT_OAC"

	// BucketTypeDataLake creates a bucket optimized for data lake analytics
	BucketTypeDataLake BucketType = "DATA_LAKE"

	// BucketTypeBackup creates a bucket optimized for backup and disaster recovery
	BucketTypeBackup BucketType = "BACKUP"

	// BucketTypeMediaStreaming creates a bucket optimized for media streaming
	BucketTypeMediaStreaming BucketType = "MEDIA_STREAMING"

	// BucketTypeEnterprise creates a bucket with maximum security for enterprise data
	BucketTypeEnterprise BucketType = "ENTERPRISE"

	// BucketTypeDevelopment creates a bucket optimized for development/testing
	BucketTypeDevelopment BucketType = "DEVELOPMENT"
)

// SimpleStorageServiceFactoryProps defines properties for creating an S3 bucket via Factory
type SimpleStorageServiceFactoryProps struct {
	BucketType BucketType

	// Required: Globally unique bucket name
	BucketName string

	// Optional: Override defaults
	RemovalPolicy     string // "retain", "destroy", "retain_on_update_or_delete"
	AutoDeleteObjects *bool  // Override auto-delete setting
}

// NewSimpleStorageServiceFactory creates an S3 bucket using the Factory + Strategy pattern
//
// This factory selects the appropriate strategy based on BucketType and delegates
// bucket creation to the specialized strategy implementation.
//
// Example usage:
//
//	bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
//	    s3.SimpleStorageServiceFactoryProps{
//	        BucketType: s3.BucketTypeCloudfrontOAC,
//	        BucketName: "my-website-bucket",
//	    })
func NewSimpleStorageServiceFactory(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket {
	var strategy SimpleStorageServiceStrategy

	// Select strategy based on bucket type
	switch props.BucketType {
	case BucketTypeCloudfrontOAC:
		strategy = &SimpleStorageServiceCloudfrontOriginStrategy{}

	case BucketTypeDataLake:
		strategy = &SimpleStorageServiceDataLakeStrategy{}

	case BucketTypeBackup:
		strategy = &SimpleStorageServiceBackupStrategy{}

	case BucketTypeMediaStreaming:
		strategy = &SimpleStorageServiceMediaStreamingStrategy{}

	case BucketTypeEnterprise:
		strategy = &SimpleStorageServiceEnterpriseStrategy{}

	case BucketTypeDevelopment:
		strategy = &SimpleStorageServiceDevelopmentStrategy{}

	default:
		panic(fmt.Sprintf("Unsupported BucketType: %s", props.BucketType))
	}

	// Delegate bucket creation to selected strategy
	return strategy.Build(scope, id, props)
}
