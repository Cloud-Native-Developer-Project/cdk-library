```go
package stacks

import (
	s3construct "cdk-library/constructs/S3"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
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
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// =============================================================================
	// 1. CREATE S3 BUCKET USING YOUR CUSTOM MODULE
	// =============================================================================
	s3Props := s3construct.GetCloudFrontOriginProperties()
	s3Props.BucketName = props.BucketName
	s3Props.RemovalPolicy = "destroy"
	s3Props.AutoDeleteObjects = true

	bucket := s3construct.NewBucket(stack, "WebsiteBucket", s3Props)

	// =============================================================================
	// 2. CREATE CLOUDFRONT DISTRIBUTION DIRECTLY (NOT USING YOUR MODULE)
	// =============================================================================
	// Esta es la forma correcta de crear la distribución para evitar problemas con IDs

	// Configure price class
	priceClass := awscloudfront.PriceClass_PRICE_CLASS_100
	if props.PriceClass == "200" {
		priceClass = awscloudfront.PriceClass_PRICE_CLASS_200
	} else if props.PriceClass == "ALL" {
		priceClass = awscloudfront.PriceClass_PRICE_CLASS_ALL
	}

	// Build distribution props
	distributionProps := &awscloudfront.DistributionProps{
		Comment:           jsii.String(props.WebsiteName + " - Static Website Distribution"),
		DefaultRootObject: jsii.String("index.html"),
		HttpVersion:       awscloudfront.HttpVersion_HTTP2_AND_3,
		EnableIpv6:        jsii.Bool(true),
		PriceClass:        priceClass,
		EnableLogging:     jsii.Bool(true),

		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:                awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, nil),
			ViewerProtocolPolicy:  awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			AllowedMethods:        awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS(),
			CachedMethods:         awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS(),
			CachePolicy:           awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
			ResponseHeadersPolicy: awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS(),
			Compress:              jsii.Bool(true),
		},

		// Error responses for SPA routing
		ErrorResponses: &[]*awscloudfront.ErrorResponse{
			{
				HttpStatus:         jsii.Number(403),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Minutes(jsii.Number(5)),
			},
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Minutes(jsii.Number(5)),
			},
		},

		// Additional behaviors for static assets
		AdditionalBehaviors: &map[string]*awscloudfront.BehaviorOptions{
			"/assets/*": {
				Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, nil),
				ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
				CachePolicy:          awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
				Compress:             jsii.Bool(true),
			},
			"/js/*": {
				Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, nil),
				ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
				CachePolicy:          awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
				Compress:             jsii.Bool(true),
			},
			"/css/*": {
				Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, nil),
				ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
				CachePolicy:          awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
				Compress:             jsii.Bool(true),
			},
		},
	}

	// Custom domain configuration
	if len(props.DomainNames) > 0 && props.CertificateArn != "" {
		domainPointers := make([]*string, len(props.DomainNames))
		for i, domain := range props.DomainNames {
			domainPointers[i] = jsii.String(domain)
		}
		distributionProps.DomainNames = &domainPointers

		certificate := awscertificatemanager.Certificate_FromCertificateArn(
			stack,
			jsii.String("Certificate"),
			jsii.String(props.CertificateArn),
		)
		distributionProps.Certificate = certificate
		distributionProps.MinimumProtocolVersion = awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
		distributionProps.SslSupportMethod = awscloudfront.SSLMethod_SNI
	}

	// WAF configuration
	if props.EnableWAF && props.WebAclArn != "" {
		distributionProps.WebAclId = jsii.String(props.WebAclArn)
	}

	// Create distribution
	distribution := awscloudfront.NewDistribution(stack, jsii.String("WebsiteDistribution"), distributionProps)

	// =============================================================================
	// 3. BUCKET POLICY FOR OAC
	// =============================================================================
	bucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:GetObject"),
		Resources: jsii.Strings(*bucket.BucketArn() + "/*"),
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewServicePrincipal(jsii.String("cloudfront.amazonaws.com"), nil),
		},
		Conditions: &map[string]interface{}{
			"StringEquals": map[string]interface{}{
				"AWS:SourceArn": distribution.DistributionArn(),
			},
		},
	}))

	// =============================================================================
	// 4. DEPLOY CONTENT
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

	deployment.Node().AddDependency(distribution)

	// =============================================================================
	// 5. OUTPUTS
	// =============================================================================
	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 bucket name"),
		ExportName:  jsii.String(props.WebsiteName + "-BucketName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DistributionId"), &awscdk.CfnOutputProps{
		Value:       distribution.DistributionId(),
		Description: jsii.String("CloudFront distribution ID"),
		ExportName:  jsii.String(props.WebsiteName + "-DistributionId"),
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

	if len(props.DomainNames) > 0 {
		awscdk.NewCfnOutput(stack, jsii.String("CustomDomainURL"), &awscdk.CfnOutputProps{
			Value:       jsii.String("https://" + props.DomainNames[0]),
			Description: jsii.String("Custom domain URL"),
			ExportName:  jsii.String(props.WebsiteName + "-CustomDomainURL"),
		})
	}

	return stack
}

```
