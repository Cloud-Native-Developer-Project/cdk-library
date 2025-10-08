package cloudfront

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/constructs-go/constructs/v10"
)

// CloudFrontStrategy define el contrato que deben implementar todos los Strategies de CloudFront
type CloudFrontStrategy interface {
	Build(scope constructs.Construct, id string, props CloudFrontPropertiesV2) awscloudfront.Distribution
}
