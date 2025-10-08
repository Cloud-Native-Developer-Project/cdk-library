package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	cloudfront "cdk-library/constructs/Cloudfront"
	s3 "cdk-library/constructs/S3"
)

type StaticWebsiteStackProps struct {
	awscdk.StackProps

	// Website Configuration
	BucketName  string
	WebsiteName string
	SourcePath  string

	// Optional: Custom Domain
	DomainNames    []string
	CertificateArn string

	// Optional: CloudFront Configuration
	PriceClass string
	EnableWAF  bool
	WebAclArn  string
}

func NewStaticWebsiteStack(scope constructs.Construct, id string, props *StaticWebsiteStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// =============================================================================
	// 1. CREATE S3 BUCKET (Private, for CloudFront Origin)
	// =============================================================================
	s3Props := s3.GetCloudFrontOriginProperties()
	s3Props.BucketName = props.BucketName
	s3Props.RemovalPolicy = "destroy"
	s3Props.AutoDeleteObjects = true

	bucket := s3.NewBucket(stack, "WebsiteBucket", s3Props)

	// =============================================================================
	// 2. CREATE CLOUDFRONT DISTRIBUTION USING FACTORY
	// =============================================================================
	distribution := cloudfront.NewDistributionV2(stack, "WebsiteDistribution", cloudfront.CloudFrontPropertiesV2{
		OriginType:                  cloudfront.OriginTypeS3,
		S3Bucket:                    bucket,
		DomainNames:                 props.DomainNames,
		CertificateArn:              props.CertificateArn,
		WebAclArn:                   props.WebAclArn,
		Comment:                     props.WebsiteName + " - Static Website Distribution",
		EnableAccessLogging:         false,
		AutoConfigureS3BucketPolicy: true,
	})

	// =============================================================================
	// 3. DEPLOY CONTENT TO S3
	// =============================================================================
	deployment := awss3deployment.NewBucketDeployment(stack, jsii.String("WebsiteDeployment"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(jsii.String(props.SourcePath), nil),
		},
		DestinationBucket:    bucket,
		DestinationKeyPrefix: jsii.String(""),
		Distribution:         distribution,
		DistributionPaths: &[]*string{
			jsii.String("/*"),
		},
		CacheControl: &[]awss3deployment.CacheControl{
			awss3deployment.CacheControl_MaxAge(awscdk.Duration_Days(jsii.Number(365))),
			awss3deployment.CacheControl_Immutable(),
		},
		Prune:          jsii.Bool(true),
		RetainOnDelete: jsii.Bool(false),
		LogRetention:   awslogs.RetentionDays_ONE_MONTH,
	})

	// Ensure deployment happens after distribution is created
	deployment.Node().AddDependency(distribution)

	// =============================================================================
	// 4. STACK OUTPUTS
	// =============================================================================
	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 bucket name"),
		ExportName:  jsii.String(props.WebsiteName + "-BucketName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DistributionDomain"), &awscdk.CfnOutputProps{
		Value:       distribution.DomainName(),
		Description: jsii.String("CloudFront domain"),
		ExportName:  jsii.String(props.WebsiteName + "-DistributionDomain"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebsiteURL"), &awscdk.CfnOutputProps{
		Value:       jsii.String("https://" + *distribution.DomainName()),
		Description: jsii.String("Website URL"),
		ExportName:  jsii.String(props.WebsiteName + "-WebsiteURL"),
	})

	return stack
}
