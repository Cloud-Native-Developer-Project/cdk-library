package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	cloudfront "cdk-library/constructs/Cloudfront"
	s3 "cdk-library/constructs/S3"
	waf "cdk-library/constructs/WAF"
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

	// Optional: WAF Configuration
	EnableWAF      bool                // Set to true to create WAF WebACL
	WafProfileType waf.WAFProfileType  // ProfileTypeWebApplication, ProfileTypeAPIProtection, or ProfileTypeBotControl (default: ProfileTypeWebApplication)
}

func NewStaticWebsiteStack(scope constructs.Construct, id string, props *StaticWebsiteStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	// =============================================================================
	// 1. CREATE S3 BUCKET (Private, for CloudFront Origin)
	// Using Factory + Strategy Pattern
	// =============================================================================
	autoDelete := true
	bucket := s3.NewSimpleStorageServiceFactory(stack, "WebsiteBucket",
		s3.SimpleStorageServiceFactoryProps{
			BucketType:        s3.BucketTypeCloudfrontOAC,
			BucketName:        props.BucketName,
			RemovalPolicy:     "destroy",
			AutoDeleteObjects: &autoDelete,
		})

	// =============================================================================
	// 2. CREATE WAF WEB ACL (Optional)
	// Using Factory + Strategy Pattern
	// =============================================================================
	var webAclArn string
	if props.EnableWAF {
		// Default to WebApplication profile if not specified
		profileType := props.WafProfileType
		if profileType == "" {
			profileType = waf.ProfileTypeWebApplication
		}

		webacl := waf.NewWebApplicationFirewallFactory(stack, "WebsiteWAF",
			waf.WAFFactoryProps{
				Scope:       waf.ScopeCloudFront, // Must be CLOUDFRONT for CloudFront distributions
				ProfileType: profileType,
				Name:        props.WebsiteName + "-waf",
			})

		webAclArn = *webacl.AttrArn()
	}

	// =============================================================================
	// 3. CREATE CLOUDFRONT DISTRIBUTION USING FACTORY
	// =============================================================================
	distribution := cloudfront.NewDistributionV2(stack, "WebsiteDistribution", cloudfront.CloudFrontPropertiesV2{
		OriginType:                  cloudfront.OriginTypeS3,
		S3Bucket:                    bucket,
		DomainNames:                 props.DomainNames,
		CertificateArn:              props.CertificateArn,
		WebAclArn:                   webAclArn,
		Comment:                     props.WebsiteName + " - Static Website Distribution",
		EnableAccessLogging:         false,
		AutoConfigureS3BucketPolicy: true,
	})

	// =============================================================================
	// 4. DEPLOY CONTENT TO S3
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
	// 5. STACK OUTPUTS
	// =============================================================================
	if props.EnableWAF {
		profileType := props.WafProfileType
		if profileType == "" {
			profileType = waf.ProfileTypeWebApplication
		}

		awscdk.NewCfnOutput(stack, jsii.String("WAFEnabled"), &awscdk.CfnOutputProps{
			Value:       jsii.String("Yes - " + string(profileType)),
			Description: jsii.String("WAF protection enabled with profile"),
			ExportName:  jsii.String(props.WebsiteName + "-WAFEnabled"),
		})

		awscdk.NewCfnOutput(stack, jsii.String("WAFWebACLArn"), &awscdk.CfnOutputProps{
			Value:       jsii.String(webAclArn),
			Description: jsii.String("WAF WebACL ARN"),
			ExportName:  jsii.String(props.WebsiteName + "-WAFWebACLArn"),
		})
	}

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
