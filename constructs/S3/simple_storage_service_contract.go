package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
)

// SimpleStorageServiceStrategy defines the contract for S3 bucket creation strategies
// Each strategy implements a specific use case (CloudFront origin, Data Lake, Backup, etc.)
type SimpleStorageServiceStrategy interface {
	Build(scope constructs.Construct, id string, props SimpleStorageServiceFactoryProps) awss3.Bucket
}
