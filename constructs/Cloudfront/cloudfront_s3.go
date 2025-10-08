package cloudfront

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type S3CloudFrontStrategy struct{}

func (s *S3CloudFrontStrategy) Build(scope constructs.Construct, id string, props CloudFrontPropertiesV2) awscloudfront.Distribution {
	// =============================================================================
	// 1. VALIDACIÓN BÁSICA
	// =============================================================================
	if props.S3Bucket == nil {
		panic("S3CloudFrontStrategy requiere un bucket S3 (props.S3Bucket no puede ser nil)")
	}

	// =============================================================================
	// 2. CREAR OAC (Origin Access Control)
	// =============================================================================
	oac := awscloudfront.NewS3OriginAccessControl(scope, jsii.String(fmt.Sprintf("%s-OAC", id)), &awscloudfront.S3OriginAccessControlProps{
		Description: jsii.String(fmt.Sprintf("OAC for %s", id)),
	})

	// =============================================================================
	// 3. CONFIGURAR DISTRIBUTION PROPS
	// =============================================================================
	distributionProps := &awscloudfront.DistributionProps{
		Comment:           jsii.String(props.Comment),
		DefaultRootObject: jsii.String("index.html"),
		HttpVersion:       awscloudfront.HttpVersion_HTTP2_AND_3,
		EnableIpv6:        jsii.Bool(true),
		EnableLogging:     jsii.Bool(props.EnableAccessLogging),
		PriceClass:        awscloudfront.PriceClass_PRICE_CLASS_100,
	}

	// SSL/TLS (opcional)
	if props.CertificateArn != "" {
		cert := awscertificatemanager.Certificate_FromCertificateArn(
			scope,
			jsii.String(fmt.Sprintf("%s-Cert", id)),
			jsii.String(props.CertificateArn),
		)
		distributionProps.Certificate = cert
		distributionProps.MinimumProtocolVersion = awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
		distributionProps.SslSupportMethod = awscloudfront.SSLMethod_SNI
	}

	// Dominio personalizado (opcional)
	if len(props.DomainNames) > 0 {
		var domains []*string
		for _, d := range props.DomainNames {
			domains = append(domains, jsii.String(d))
		}
		distributionProps.DomainNames = &domains
	}

	// =============================================================================
	// 4. DEFAULT BEHAVIOR
	// =============================================================================
	distributionProps.DefaultBehavior = &awscloudfront.BehaviorOptions{
		Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(
			props.S3Bucket,
			&awscloudfrontorigins.S3BucketOriginWithOACProps{
				OriginAccessControl: oac,
			},
		),
		ViewerProtocolPolicy:  awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
		AllowedMethods:        awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS(),
		CachedMethods:         awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS(),
		CachePolicy:           awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
		ResponseHeadersPolicy: awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS(),
		Compress:              jsii.Bool(true),
	}

	// =============================================================================
	// 5. ERRORES PERSONALIZADOS PARA SPA (index.html)
	// =============================================================================
	distributionProps.ErrorResponses = &[]*awscloudfront.ErrorResponse{
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
	}

	// =============================================================================
	// 6. WAF (si aplica)
	// =============================================================================
	if props.WebAclArn != "" {
		distributionProps.WebAclId = jsii.String(props.WebAclArn)
	}

	// =============================================================================
	// 7. CREAR DISTRIBUTION
	// =============================================================================
	distribution := awscloudfront.NewDistribution(scope, jsii.String(fmt.Sprintf("%s-Distribution", id)), distributionProps)

	// =============================================================================
	// 8. POLÍTICA S3 PARA OAC
	// =============================================================================
	if props.AutoConfigureS3BucketPolicy {
		props.S3Bucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Sid:    jsii.String("AllowCloudFrontServicePrincipal"),
			Effect: awsiam.Effect_ALLOW,
			Principals: &[]awsiam.IPrincipal{
				awsiam.NewServicePrincipal(jsii.String("cloudfront.amazonaws.com"), nil),
			},
			Actions: jsii.Strings("s3:GetObject"),
			Resources: jsii.Strings(
				fmt.Sprintf("%s/*", *props.S3Bucket.BucketArn()),
			),
			Conditions: &map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"AWS:SourceArn": *distribution.DistributionArn(),
				},
			},
		}))
	}

	return distribution
}
