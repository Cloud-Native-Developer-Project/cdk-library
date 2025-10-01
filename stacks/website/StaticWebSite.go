package stacks

import (
	cloudfront "cdk-library/constructs/Cloudfront"
	s3 "cdk-library/constructs/S3"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	// Importa tus módulos personalizados
)

type StaticWebsiteStackProps struct {
	awscdk.StackProps

	// Website Configuration
	BucketName  string
	WebsiteName string
	SourcePath  string // Path to your built frontend (e.g., "./dist" or "./build")

	// Optional: Custom Domain
	DomainNames    []string
	CertificateArn string

	// Optional: CloudFront Configuration
	PriceClass string // "100", "200", "ALL"
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
	// 1. CREATE S3 BUCKET USING CUSTOM CONSTRUCT
	// =============================================================================

	// Get optimized properties for CloudFront origin
	s3Props := s3.GetCloudFrontOriginProperties()

	// Override with custom values
	s3Props.BucketName = props.BucketName
	s3Props.RemovalPolicy = "destroy" // Change to "retain" for production
	s3Props.AutoDeleteObjects = true  // Change to false for production

	// Create the bucket
	bucket := s3.NewBucket(stack, "WebsiteBucket", s3Props)

	// =============================================================================
	// 2. CREATE CLOUDFRONT DISTRIBUTION USING CUSTOM CONSTRUCT
	// =============================================================================

	// Get default properties for S3 + OAC configuration
	cfProps := cloudfront.DefaultS3PrivateOACProps()

	// Override with custom values
	cfProps.Comment = props.WebsiteName + " - Static Website Distribution"
	cfProps.S3BucketName = props.BucketName
	cfProps.S3Bucket = bucket

	// Set price class
	if props.PriceClass != "" {
		cfProps.PriceClass = props.PriceClass
	} else {
		cfProps.PriceClass = "100" // Cost-optimized for most use cases
	}

	// Configure custom domain if provided
	if len(props.DomainNames) > 0 && props.CertificateArn != "" {
		cfProps.DomainNames = props.DomainNames
		cfProps.CertificateArn = props.CertificateArn
		cfProps.MinimumProtocolVersion = "TLS_V1_2_2021"
		cfProps.SSLSupportMethod = "SNI_ONLY"
	}

	// Configure WAF if enabled
	if props.EnableWAF && props.WebAclArn != "" {
		cfProps.WebAclArn = props.WebAclArn
	}

	// Add SPA error page handling (for Vue.js, React, Angular, etc.)
	cfProps.EnableErrorPages = true
	cfProps.ErrorPageConfigs = []cloudfront.ErrorPageConfig{
		{
			ErrorCode:          403,
			ResponseCode:       200,
			ResponsePagePath:   "/index.html",
			ErrorCachingMinTTL: 0, // Don't cache 404s for SPAs during development
		},
		{
			ErrorCode:          404,
			ResponseCode:       200,
			ResponsePagePath:   "/index.html",
			ErrorCachingMinTTL: 0,
		},
	}

	// Configure additional behaviors for static assets with long-term caching
	/*
		cfProps.AdditionalBehaviors = []cloudfront.BehaviorConfig{
			{
				PathPattern:          "/assets/*",
				UseDefaultOrigin:     true, // ← NUEVO: Reutilizar origen por defecto
				CachePolicy:          "MANAGED_CACHING_OPTIMIZED",
				ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
				CompressResponse:     true,
			},
			{
				PathPattern:          "/static/*",
				UseDefaultOrigin:     true, // ← NUEVO
				CachePolicy:          "MANAGED_CACHING_OPTIMIZED",
				ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
				CompressResponse:     true,
			},
			{
				PathPattern:          "*.js",
				UseDefaultOrigin:     true, // ← NUEVO
				CachePolicy:          "MANAGED_CACHING_OPTIMIZED",
				ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
				CompressResponse:     true,
			},
			{
				PathPattern:          "*.css",
				UseDefaultOrigin:     true, // ← NUEVO
				CachePolicy:          "MANAGED_CACHING_OPTIMIZED",
				ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
				CompressResponse:     true,
			},
		}
	*/

	// Enable logging for production monitoring
	cfProps.EnableAccessLogging = true
	cfProps.LoggingPrefix = "cloudfront-logs/"

	// Create CloudFront distribution
	distribution := cloudfront.NewDistribution(stack, "WebsiteDistribution", cfProps)

	// =============================================================================
	// 3. APLICAR PERMISOS DE LECTURA (OAC) AL BUCKET (CORRECCIÓN)
	// =============================================================================
	// Reemplaza la llamada a GrantRead fallida con la adición manual de la política.
	// Esto otorga permiso al servicio CloudFront (restringido por el ARN de esta distribución)
	// para leer objetos del bucket S3, que es la forma correcta de usar OAC.
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
	// 4. DEPLOY WEBSITE CONTENT TO S3
	// =============================================================================

	deployment := awss3deployment.NewBucketDeployment(stack, jsii.String("WebsiteDeployment"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(jsii.String(props.SourcePath), nil),
		},
		DestinationBucket:    bucket,
		DestinationKeyPrefix: jsii.String(""),

		// Invalidate CloudFront cache on deployment
		Distribution: distribution,
		DistributionPaths: &[]*string{
			jsii.String("/*"),
		},

		// Set cache control headers for optimal caching
		CacheControl: &[]awss3deployment.CacheControl{
			awss3deployment.CacheControl_MaxAge(awscdk.Duration_Days(jsii.Number(365))),
			awss3deployment.CacheControl_Immutable(),
		},

		// Prune old files on deployment (clean bucket before deploying)
		Prune: jsii.Bool(true),

		// Retain deployment on stack deletion
		RetainOnDelete: jsii.Bool(false),

		LogRetention: awslogs.RetentionDays_ONE_MONTH,
	})

	// Ensure deployment happens after distribution is created
	deployment.Node().AddDependency(distribution)

	// =============================================================================
	// 5. STACK OUTPUTS
	// =============================================================================

	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 bucket name for website content"),
		ExportName:  jsii.String(props.WebsiteName + "-BucketName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("BucketArn"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketArn(),
		Description: jsii.String("S3 bucket ARN"),
		ExportName:  jsii.String(props.WebsiteName + "-BucketArn"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DistributionId"), &awscdk.CfnOutputProps{
		Value:       distribution.DistributionId(),
		Description: jsii.String("CloudFront distribution ID"),
		ExportName:  jsii.String(props.WebsiteName + "-DistributionId"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DistributionDomain"), &awscdk.CfnOutputProps{
		Value:       distribution.DomainName(),
		Description: jsii.String("CloudFront distribution domain name"),
		ExportName:  jsii.String(props.WebsiteName + "-DistributionDomain"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebsiteURL"), &awscdk.CfnOutputProps{
		Value:       jsii.String("https://" + *distribution.DomainName()),
		Description: jsii.String("Website URL"),
		ExportName:  jsii.String(props.WebsiteName + "-WebsiteURL"),
	})

	// If custom domain is configured, output it
	if len(props.DomainNames) > 0 {
		awscdk.NewCfnOutput(stack, jsii.String("CustomDomainURL"), &awscdk.CfnOutputProps{
			Value:       jsii.String("https://" + props.DomainNames[0]),
			Description: jsii.String("Custom domain URL"),
			ExportName:  jsii.String(props.WebsiteName + "-CustomDomainURL"),
		})
	}

	return stack
}
