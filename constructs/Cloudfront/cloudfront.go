package cloudfront

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// CloudFrontProperties defines all configurable properties for CloudFront distribution creation
// Includes performance optimization, security, caching, and monitoring configurations
type CloudFrontProperties struct {
	// Basic Configuration
	Comment           string   // Description for the distribution
	Enabled           bool     // Whether the distribution is enabled
	DefaultRootObject string   // Default object to serve at root (e.g., "index.html")
	DomainNames       []string // Custom domain names (CNAMEs)
	PriceClass        string   // Price class: "ALL", "100", "200" - controls geographic distribution
	HttpVersion       string   // HTTP version: "HTTP1_1", "HTTP2", "HTTP2_AND_3"
	EnableIPv6        bool     // Enable IPv6 support

	// Origin Configuration
	OriginType         string // Type of origin: "S3", "S3_WEBSITE", "HTTP", "LOAD_BALANCER"
	OriginDomainName   string // Origin domain name (S3 bucket, ALB DNS, custom domain)
	OriginPath         string // Path prefix to add to requests sent to origin
	OriginShield       bool   // Enable Origin Shield for cost optimization
	OriginShieldRegion string // Origin Shield region (required if OriginShield is true)

	// S3 Origin Specific (when OriginType is "S3")
	S3BucketName string        // S3 bucket name for S3 origins
	S3Bucket     awss3.IBucket // Direct bucket reference (takes precedence over S3BucketName)

	// HTTP/Custom Origin Specific
	OriginProtocolPolicy   string   // "HTTP_ONLY", "HTTPS_ONLY", "MATCH_VIEWER"
	OriginPort             int32    // Custom origin port (80, 443, or custom)
	OriginSSLProtocols     []string // SSL protocols for custom origins
	OriginReadTimeout      int32    // Timeout for origin requests (1-180 seconds)
	OriginKeepaliveTimeout int32    // Keep-alive timeout (1-60 seconds)

	// SSL/TLS Configuration
	CertificateArn         string // ACM certificate ARN (must be in us-east-1)
	MinimumProtocolVersion string // Minimum TLS version: "TLS_V1_2016", "TLS_V1_1_2016", "TLS_V1_2_2019", "TLS_V1_2_2021"
	SSLSupportMethod       string // SSL support method: "SNI_ONLY", "VIP"

	// Security Configuration
	WebAclArn               string   // AWS WAF WebACL ARN
	GeoRestrictionType      string   // Geographic restriction: "ALLOW", "DENY", "NONE"
	GeoRestrictionCountries []string // List of country codes for geo restriction

	// Caching Configuration
	CachePolicy           string   // Cache policy: "MANAGED_CACHING_OPTIMIZED", "MANAGED_CACHING_DISABLED", "MANAGED_AMPLIFY", "CUSTOM"
	OriginRequestPolicy   string   // Origin request policy: "MANAGED_ALL_VIEWER", "MANAGED_CORS_S3", "MANAGED_ELEMENT_CAPTURE", "CUSTOM"
	ResponseHeadersPolicy string   // Response headers policy: "MANAGED_CORS_ALLOW_ALL", "MANAGED_SECURITY_HEADERS", "CUSTOM"
	CompressResponse      bool     // Enable automatic compression
	ViewerProtocolPolicy  string   // Viewer protocol: "ALLOW_ALL", "REDIRECT_TO_HTTPS", "HTTPS_ONLY"
	AllowedMethods        []string // Allowed HTTP methods
	CachedMethods         []string // Methods to cache responses for

	// Custom Cache Policy (when CachePolicy is "CUSTOM")
	CustomCachePolicyName   string
	CustomCacheTTLDefault   int32    // Default TTL in seconds
	CustomCacheTTLMin       int32    // Minimum TTL in seconds
	CustomCacheTTLMax       int32    // Maximum TTL in seconds
	CustomCacheQueryStrings []string // Query string parameters to include in cache key
	CustomCacheHeaders      []string // Headers to include in cache key
	CustomCacheCookies      []string // Cookies to include in cache key

	// Error Pages Configuration
	EnableErrorPages bool              // Enable custom error pages
	ErrorPageConfigs []ErrorPageConfig // Custom error page configurations

	// Logging Configuration
	EnableAccessLogging   bool   // Enable standard access logging
	LoggingBucket         string // S3 bucket for access logs
	LoggingPrefix         string // Prefix for log files
	LoggingIncludeCookies bool   // Include cookies in access logs
	EnableRealtimeLogging bool   // Enable real-time logging
	RealtimeLogArn        string // Kinesis Data Stream ARN for real-time logs

	// Monitoring Configuration
	EnableAdditionalMetrics   bool // Enable additional CloudWatch metrics
	MonitoringRealtimeMetrics bool // Enable real-time metrics (additional cost)

	// Edge Functions Configuration
	EnableEdgeFunctions bool                       // Enable CloudFront Functions or Lambda@Edge
	CloudFrontFunctions []CloudFrontFunctionConfig // CloudFront Functions configuration
	LambdaEdgeFunctions []LambdaEdgeConfig         // Lambda@Edge functions configuration

	// Additional Behaviors (Path-based routing)
	AdditionalBehaviors []BehaviorConfig // Additional cache behaviors for specific paths

	// Performance Optimization
	EnableGRPC      bool // Enable gRPC support (requires HTTP/2)
	SmoothStreaming bool // Enable Microsoft Smooth Streaming support

	// Trusted Signers (for private content)
	TrustedSigners   []string // AWS account IDs for trusted signers
	TrustedKeyGroups []string // Key group names for signed URLs/cookies
}

// ErrorPageConfig defines custom error page configuration
type ErrorPageConfig struct {
	ErrorCode          int32  // HTTP error code (400, 403, 404, 500, etc.)
	ResponseCode       int32  // HTTP response code to return instead
	ResponsePagePath   string // Path to custom error page
	ErrorCachingMinTTL int32  // Minimum TTL for error responses
}

// CloudFrontFunctionConfig defines CloudFront Function configuration
type CloudFrontFunctionConfig struct {
	FunctionName string // Name of the CloudFront function
	EventType    string // Event type: "VIEWER_REQUEST", "VIEWER_RESPONSE"
	FunctionCode string // JavaScript code for the function (inline)
}

// LambdaEdgeConfig defines Lambda@Edge function configuration
type LambdaEdgeConfig struct {
	FunctionArn string // Lambda function ARN (must be in us-east-1)
	EventType   string // Event type: "ORIGIN_REQUEST", "ORIGIN_RESPONSE", "VIEWER_REQUEST", "VIEWER_RESPONSE"
	IncludeBody bool   // Include request body for Lambda@Edge
}

// BehaviorConfig defines additional cache behavior configuration
type BehaviorConfig struct {
	PathPattern           string   // Path pattern to match (e.g., "/api/*", "*.jpg")
	UseDefaultOrigin      bool     // Set to true to reuse default origin
	OriginType            string   // Override origin type for this path
	OriginDomainName      string   // Override origin domain for this path
	CachePolicy           string   // Override cache policy for this path
	OriginRequestPolicy   string   // Override origin request policy for this path
	ResponseHeadersPolicy string   // Override response headers policy for this path
	ViewerProtocolPolicy  string   // Override viewer protocol policy for this path
	AllowedMethods        []string // Override allowed methods for this path
	CachedMethods         []string // Override cached methods for this path
	CompressResponse      bool     // Override compression setting for this path
	TrustedSigners        []string // Override trusted signers for this path
	TrustedKeyGroups      []string // Override trusted key groups for this path
}

// NewDistribution creates a new CloudFront distribution with comprehensive configuration options
// This function applies AWS best practices for performance, security, and cost optimization
func NewDistribution(scope constructs.Construct, id string, props CloudFrontProperties) awscloudfront.Distribution {
	distributionProps := &awscloudfront.DistributionProps{
		Comment:           jsii.String(props.Comment),
		Enabled:           jsii.Bool(props.Enabled),
		DefaultRootObject: configureDefaultRootObject(props.DefaultRootObject),
		DomainNames:       convertToStringPointers(props.DomainNames),
		PriceClass:        configurePriceClass(props.PriceClass),
		HttpVersion:       configureHttpVersion(props.HttpVersion),
		EnableIpv6:        jsii.Bool(props.EnableIPv6),
	}

	// Configure default behavior and capture the origin
	defaultBehavior, defaultOrigin := configureDefaultBehavior(scope, props)
	distributionProps.DefaultBehavior = defaultBehavior

	configureSSLSettings(scope, distributionProps, props)
	configureSecurity(distributionProps, props)
	configureErrorPages(distributionProps, props)
	configureLogging(scope, distributionProps, props)
	configureMonitoring(distributionProps, props)

	// Pass defaultOrigin to additional behaviors
	configureAdditionalBehaviors(scope, distributionProps, props, defaultOrigin)

	distribution := awscloudfront.NewDistribution(scope, jsii.String(id), distributionProps)
	return distribution
}

// configureDefaultBehavior sets up the default cache behavior
func configureDefaultBehavior(scope constructs.Construct, props CloudFrontProperties) (*awscloudfront.BehaviorOptions, awscloudfront.IOrigin) {
	origin := createOrigin(scope, "DefaultOrigin", props)

	behaviorOptions := &awscloudfront.BehaviorOptions{
		Origin:                origin,
		ViewerProtocolPolicy:  configureViewerProtocolPolicy(props.ViewerProtocolPolicy),
		AllowedMethods:        configureAllowedMethods(props.AllowedMethods),
		CachedMethods:         configureCachedMethods(props.CachedMethods),
		Compress:              jsii.Bool(props.CompressResponse),
		CachePolicy:           configureCachePolicy(scope, props),
		OriginRequestPolicy:   configureOriginRequestPolicy(props.OriginRequestPolicy),
		ResponseHeadersPolicy: configureResponseHeadersPolicy(props.ResponseHeadersPolicy),
	}

	// Configure gRPC support
	if props.EnableGRPC {
		behaviorOptions.EnableGrpc = jsii.Bool(true)
	}

	// Configure smooth streaming
	if props.SmoothStreaming {
		behaviorOptions.SmoothStreaming = jsii.Bool(true)
	}

	// Configure trusted signers
	if len(props.TrustedSigners) > 0 {
		// Note: Trusted signers require specific implementation based on your key management
		// This is a placeholder for the concept
	}

	// Configure edge functions
	configureEdgeFunctions(scope, behaviorOptions, props)

	return behaviorOptions, origin
}

// createOrigin creates the appropriate origin based on the configuration
func createOrigin(scope constructs.Construct, idPrefix string, props CloudFrontProperties) awscloudfront.IOrigin {
	switch props.OriginType {
	case "S3":
		// 🛑 CORRECCIÓN: Pasar el idPrefix.
		return createS3Origin(scope, idPrefix, props)
	case "S3_WEBSITE":
		// 🛑 CORRECCIÓN: Pasar el idPrefix.
		return createS3WebsiteOrigin(scope, idPrefix, props)
	case "HTTP", "HTTPS":
		return createHttpOrigin(props)
	case "LOAD_BALANCER":
		return createLoadBalancerOrigin(props)
	default:
		// Default to HTTP origin
		return createHttpOrigin(props)
	}
}

func createS3Origin(scope constructs.Construct, idPrefix string, props CloudFrontProperties) awscloudfront.IOrigin {
	var bucket awss3.IBucket

	// Use direct bucket reference if provided, otherwise import by name
	if props.S3Bucket != nil {
		bucket = props.S3Bucket
	} else if props.S3BucketName != "" {
		originId := idPrefix + "-" + props.S3BucketName
		bucket = awss3.Bucket_FromBucketName(scope, jsii.String(originId), jsii.String(props.S3BucketName))
	} else {
		panic("Either S3Bucket or S3BucketName must be provided for S3 origin")
	}

	s3OriginProps := &awscloudfrontorigins.S3BucketOriginWithOACProps{
		OriginPath: jsii.String(props.OriginPath),
	}

	if props.OriginShield && props.OriginShieldRegion != "" {
		s3OriginProps.OriginShieldEnabled = jsii.Bool(true)
		s3OriginProps.OriginShieldRegion = jsii.String(props.OriginShieldRegion)
	}

	return awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, s3OriginProps)
}

// createS3WebsiteOrigin creates an S3 website origin
func createS3WebsiteOrigin(scope constructs.Construct, idPrefix string, props CloudFrontProperties) awscloudfront.IOrigin {
	var bucket awss3.IBucket

	// Use direct bucket reference if provided, otherwise import by name
	if props.S3Bucket != nil {
		bucket = props.S3Bucket
	} else if props.S3BucketName != "" {
		originId := idPrefix + "-" + props.S3BucketName
		bucket = awss3.Bucket_FromBucketName(scope, jsii.String(originId), jsii.String(props.S3BucketName))
	} else {
		panic("Either S3Bucket or S3BucketName must be provided for S3 website origin")
	}

	s3WebsiteOriginProps := &awscloudfrontorigins.S3StaticWebsiteOriginProps{
		OriginPath: jsii.String(props.OriginPath),
	}

	if props.OriginShield && props.OriginShieldRegion != "" {
		s3WebsiteOriginProps.OriginShieldEnabled = jsii.Bool(true)
		s3WebsiteOriginProps.OriginShieldRegion = jsii.String(props.OriginShieldRegion)
	}

	return awscloudfrontorigins.NewS3StaticWebsiteOrigin(bucket, s3WebsiteOriginProps)
}

// createHttpOrigin creates a custom HTTP/HTTPS origin
func createHttpOrigin(props CloudFrontProperties) awscloudfront.IOrigin {
	httpOriginProps := &awscloudfrontorigins.HttpOriginProps{
		OriginPath:       jsii.String(props.OriginPath),
		ProtocolPolicy:   configureOriginProtocolPolicy(props.OriginProtocolPolicy),
		HttpPort:         jsii.Number(getPortOrDefault(props.OriginPort, 80)),
		HttpsPort:        jsii.Number(getPortOrDefault(props.OriginPort, 443)),
		ReadTimeout:      awscdk.Duration_Seconds(jsii.Number(getTimeoutOrDefault(props.OriginReadTimeout, 30))),
		KeepaliveTimeout: awscdk.Duration_Seconds(jsii.Number(getTimeoutOrDefault(props.OriginKeepaliveTimeout, 5))),
	}

	// Configure SSL protocols for HTTPS origins
	if len(props.OriginSSLProtocols) > 0 {
		httpOriginProps.OriginSslProtocols = convertToOriginSslProtocols(props.OriginSSLProtocols)
	}

	// Configure Origin Shield if enabled
	if props.OriginShield && props.OriginShieldRegion != "" {
		httpOriginProps.OriginShieldEnabled = jsii.Bool(true)
		httpOriginProps.OriginShieldRegion = jsii.String(props.OriginShieldRegion)
	}

	return awscloudfrontorigins.NewHttpOrigin(jsii.String(props.OriginDomainName), httpOriginProps)
}

// createLoadBalancerOrigin creates an Application Load Balancer origin
func createLoadBalancerOrigin(props CloudFrontProperties) awscloudfront.IOrigin {
	// For ALB origins, we use HttpOrigin with the ALB DNS name
	return createHttpOrigin(props)
}

// configureSSLSettings configures SSL/TLS settings for the distribution
func configureSSLSettings(scope constructs.Construct, distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.CertificateArn != "" {
		// Import existing certificate
		certificate := awscertificatemanager.Certificate_FromCertificateArn(
			scope, jsii.String("Certificate"), jsii.String(props.CertificateArn))
		distributionProps.Certificate = certificate

		// Set minimum protocol version
		distributionProps.MinimumProtocolVersion = configureMinimumProtocolVersion(props.MinimumProtocolVersion)

		// Set SSL support method
		distributionProps.SslSupportMethod = configureSSLSupportMethod(props.SSLSupportMethod)
	}
}

// configureSecurity configures security settings including WAF and geo restrictions
func configureSecurity(distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	// Configure AWS WAF WebACL
	if props.WebAclArn != "" {
		distributionProps.WebAclId = jsii.String(props.WebAclArn)
	}

	// Configure geographic restrictions
	if props.GeoRestrictionType != "NONE" && len(props.GeoRestrictionCountries) > 0 {
		distributionProps.GeoRestriction = configureGeoRestriction(
			props.GeoRestrictionType, props.GeoRestrictionCountries)
	}
}

// configureErrorPages sets up custom error page configurations
func configureErrorPages(distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.EnableErrorPages && len(props.ErrorPageConfigs) > 0 {
		errorResponses := make([]*awscloudfront.ErrorResponse, 0, len(props.ErrorPageConfigs))

		for _, config := range props.ErrorPageConfigs {
			errorResponse := &awscloudfront.ErrorResponse{
				HttpStatus:         jsii.Number(config.ErrorCode),
				ResponseHttpStatus: jsii.Number(config.ResponseCode),
				ResponsePagePath:   jsii.String(config.ResponsePagePath),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(config.ErrorCachingMinTTL)),
			}
			errorResponses = append(errorResponses, errorResponse)
		}

		distributionProps.ErrorResponses = &errorResponses
	}
}

// configureLogging sets up access logging configuration
func configureLogging(scope constructs.Construct, distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.EnableAccessLogging {
		distributionProps.EnableLogging = jsii.Bool(true)

		if props.LoggingBucket != "" {
			logBucket := awss3.Bucket_FromBucketName(scope, jsii.String("LogBucket"), jsii.String(props.LoggingBucket))
			distributionProps.LogBucket = logBucket
		}

		if props.LoggingPrefix != "" {
			distributionProps.LogFilePrefix = jsii.String(props.LoggingPrefix)
		}

		distributionProps.LogIncludesCookies = jsii.Bool(props.LoggingIncludeCookies)
	}
}

// configureMonitoring sets up monitoring and metrics configuration
func configureMonitoring(distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties) {
	if props.EnableAdditionalMetrics {
		distributionProps.PublishAdditionalMetrics = jsii.Bool(true)
	}
}

// configureAdditionalBehaviors sets up additional cache behaviors for path-based routing
func configureAdditionalBehaviors(scope constructs.Construct, distributionProps *awscloudfront.DistributionProps, props CloudFrontProperties, defaultOrigin awscloudfront.IOrigin) {
	if len(props.AdditionalBehaviors) > 0 {
		additionalBehaviors := make(map[string]*awscloudfront.BehaviorOptions)

		for _, behaviorConfig := range props.AdditionalBehaviors {
			behaviorOptions := createBehaviorFromConfig(scope, behaviorConfig, props, defaultOrigin) // ← Pasar defaultOrigin
			additionalBehaviors[behaviorConfig.PathPattern] = behaviorOptions
		}

		distributionProps.AdditionalBehaviors = &additionalBehaviors
	}
}

// createBehaviorFromConfig creates a behavior configuration from BehaviorConfig
func createBehaviorFromConfig(scope constructs.Construct, config BehaviorConfig, defaultProps CloudFrontProperties, defaultOrigin awscloudfront.IOrigin) *awscloudfront.BehaviorOptions {
	var origin awscloudfront.IOrigin
	// Create a temporary props object with overrides
	tempProps := defaultProps
	// Reutilizar origen por defecto si está especificado
	if config.UseDefaultOrigin {
		origin = defaultOrigin
	} else {
		// Crear nuevo origen solo si es necesario
		tempProps := defaultProps
		if config.OriginType != "" {
			tempProps.OriginType = config.OriginType
		}
		if config.OriginDomainName != "" {
			tempProps.OriginDomainName = config.OriginDomainName
		}

		behaviorOriginId := "BehaviorOrigin-" + sanitizeID(config.PathPattern)
		origin = createOrigin(scope, behaviorOriginId, tempProps)
	}

	behaviorOptions := &awscloudfront.BehaviorOptions{
		Origin:               origin,
		ViewerProtocolPolicy: configureViewerProtocolPolicy(getStringOrDefault(config.ViewerProtocolPolicy, defaultProps.ViewerProtocolPolicy)),
		AllowedMethods:       configureAllowedMethods(getStringSliceOrDefault(config.AllowedMethods, defaultProps.AllowedMethods)),
		CachedMethods:        configureCachedMethods(getStringSliceOrDefault(config.CachedMethods, defaultProps.CachedMethods)),
		Compress:             jsii.Bool(config.CompressResponse),
	}

	// Configure cache policy for this behavior
	if config.CachePolicy != "" {
		tempProps.CachePolicy = config.CachePolicy
		behaviorOptions.CachePolicy = configureCachePolicy(scope, tempProps)
	}

	// Configure origin request policy for this behavior
	if config.OriginRequestPolicy != "" {
		behaviorOptions.OriginRequestPolicy = configureOriginRequestPolicy(config.OriginRequestPolicy)
	}

	// Configure response headers policy for this behavior
	if config.ResponseHeadersPolicy != "" {
		behaviorOptions.ResponseHeadersPolicy = configureResponseHeadersPolicy(config.ResponseHeadersPolicy)
	}

	return behaviorOptions
}

// configureEdgeFunctions sets up CloudFront Functions and Lambda@Edge
func configureEdgeFunctions(scope constructs.Construct, behaviorOptions *awscloudfront.BehaviorOptions, props CloudFrontProperties) {
	if !props.EnableEdgeFunctions {
		return
	}

	// Configure CloudFront Functions
	if len(props.CloudFrontFunctions) > 0 {
		functionAssociations := make([]*awscloudfront.FunctionAssociation, 0, len(props.CloudFrontFunctions))

		for _, funcConfig := range props.CloudFrontFunctions {
			// Note: This is a conceptual implementation
			// You would need to create the CloudFront Function separately and reference it here
			functionAssociation := &awscloudfront.FunctionAssociation{
				EventType: configureFunctionEventType(funcConfig.EventType),
				// Function: ... // Reference to the actual CloudFront Function
			}
			functionAssociations = append(functionAssociations, functionAssociation)
		}

		behaviorOptions.FunctionAssociations = &functionAssociations
	}

	// Configure Lambda@Edge functions
	if len(props.LambdaEdgeFunctions) > 0 {
		edgeLambdas := make([]*awscloudfront.EdgeLambda, 0, len(props.LambdaEdgeFunctions))

		for _, lambdaConfig := range props.LambdaEdgeFunctions {
			// Create Lambda version from ARN
			functionVersion := awslambda.Version_FromVersionArn(scope, jsii.String("LambdaEdgeVersion"), jsii.String(lambdaConfig.FunctionArn))

			edgeLambda := &awscloudfront.EdgeLambda{
				EventType:       configureLambdaEventType(lambdaConfig.EventType),
				FunctionVersion: functionVersion,
				IncludeBody:     jsii.Bool(lambdaConfig.IncludeBody),
			}
			edgeLambdas = append(edgeLambdas, edgeLambda)
		}

		behaviorOptions.EdgeLambdas = &edgeLambdas
	}
}

// Helper functions for configuration conversion

func configureDefaultRootObject(rootObject string) *string {
	if rootObject == "" {
		return nil
	}
	return jsii.String(rootObject)
}

func configurePriceClass(priceClass string) awscloudfront.PriceClass {
	switch priceClass {
	case "ALL":
		return awscloudfront.PriceClass_PRICE_CLASS_ALL
	case "200":
		return awscloudfront.PriceClass_PRICE_CLASS_200
	case "100":
		return awscloudfront.PriceClass_PRICE_CLASS_100
	default:
		return awscloudfront.PriceClass_PRICE_CLASS_ALL
	}
}

func configureHttpVersion(httpVersion string) awscloudfront.HttpVersion {
	switch httpVersion {
	case "HTTP1_1":
		return awscloudfront.HttpVersion_HTTP1_1
	case "HTTP2":
		return awscloudfront.HttpVersion_HTTP2
	case "HTTP2_AND_3":
		return awscloudfront.HttpVersion_HTTP2_AND_3
	default:
		return awscloudfront.HttpVersion_HTTP2
	}
}

func configureViewerProtocolPolicy(policy string) awscloudfront.ViewerProtocolPolicy {
	switch policy {
	case "ALLOW_ALL":
		return awscloudfront.ViewerProtocolPolicy_ALLOW_ALL
	case "REDIRECT_TO_HTTPS":
		return awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS
	case "HTTPS_ONLY":
		return awscloudfront.ViewerProtocolPolicy_HTTPS_ONLY
	default:
		return awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS
	}
}

func configureAllowedMethods(methods []string) awscloudfront.AllowedMethods {
	if len(methods) == 0 {
		return awscloudfront.AllowedMethods_ALLOW_GET_HEAD()
	}

	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	// Check for all methods
	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] &&
		methodSet["PUT"] && methodSet["POST"] && methodSet["PATCH"] && methodSet["DELETE"] {
		return awscloudfront.AllowedMethods_ALLOW_ALL()
	}

	// Check for GET, HEAD, OPTIONS, PUT, POST, PATCH, DELETE
	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] &&
		methodSet["PUT"] && methodSet["POST"] && methodSet["PATCH"] && methodSet["DELETE"] {
		return awscloudfront.AllowedMethods_ALLOW_ALL()
	}

	// Check for GET, HEAD, OPTIONS
	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] {
		return awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS()
	}

	// Default to GET, HEAD
	return awscloudfront.AllowedMethods_ALLOW_GET_HEAD()
}

func configureCachedMethods(methods []string) awscloudfront.CachedMethods {
	if len(methods) == 0 {
		return awscloudfront.CachedMethods_CACHE_GET_HEAD()
	}

	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	if methodSet["GET"] && methodSet["HEAD"] && methodSet["OPTIONS"] {
		return awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS()
	}

	return awscloudfront.CachedMethods_CACHE_GET_HEAD()
}

func configureCachePolicy(scope constructs.Construct, props CloudFrontProperties) awscloudfront.ICachePolicy {
	switch props.CachePolicy {
	case "MANAGED_CACHING_OPTIMIZED":
		return awscloudfront.CachePolicy_CACHING_OPTIMIZED()
	case "MANAGED_CACHING_DISABLED":
		return awscloudfront.CachePolicy_CACHING_DISABLED()
	case "MANAGED_AMPLIFY":
		return awscloudfront.CachePolicy_AMPLIFY()
	case "CUSTOM":
		// Create custom cache policy
		return createCustomCachePolicy(scope, props)
	default:
		return awscloudfront.CachePolicy_CACHING_OPTIMIZED()
	}
}

func createCustomCachePolicy(scope constructs.Construct, props CloudFrontProperties) awscloudfront.ICachePolicy {
	// Create custom cache policy with user-defined parameters
	cachePolicyProps := &awscloudfront.CachePolicyProps{
		CachePolicyName: jsii.String(getStringOrDefault(props.CustomCachePolicyName, "CustomCachePolicy")),
		DefaultTtl:      awscdk.Duration_Seconds(jsii.Number(getInt32OrDefault(props.CustomCacheTTLDefault, 86400))), // 1 day default
		MinTtl:          awscdk.Duration_Seconds(jsii.Number(getInt32OrDefault(props.CustomCacheTTLMin, 0))),
		MaxTtl:          awscdk.Duration_Seconds(jsii.Number(getInt32OrDefault(props.CustomCacheTTLMax, 31536000))), // 1 year max
		Comment:         jsii.String("Custom cache policy created by CloudFront template"),
	}

	// Configure query string behavior
	if len(props.CustomCacheQueryStrings) > 0 {
		cachePolicyProps.QueryStringBehavior = awscloudfront.CacheQueryStringBehavior_AllowList(
			*jsii.Strings(props.CustomCacheQueryStrings...)...)
	} else {
		cachePolicyProps.QueryStringBehavior = awscloudfront.CacheQueryStringBehavior_None()
	}

	// Configure header behavior
	if len(props.CustomCacheHeaders) > 0 {
		cachePolicyProps.HeaderBehavior = awscloudfront.CacheHeaderBehavior_AllowList(
			*jsii.Strings(props.CustomCacheHeaders...)...)
	} else {
		cachePolicyProps.HeaderBehavior = awscloudfront.CacheHeaderBehavior_None()
	}

	// Configure cookie behavior
	if len(props.CustomCacheCookies) > 0 {
		cachePolicyProps.CookieBehavior = awscloudfront.CacheCookieBehavior_AllowList(
			*jsii.Strings(props.CustomCacheCookies...)...)
	} else {
		cachePolicyProps.CookieBehavior = awscloudfront.CacheCookieBehavior_None()
	}

	// Enable origin compression by default for custom policies
	cachePolicyProps.EnableAcceptEncodingGzip = jsii.Bool(true)
	cachePolicyProps.EnableAcceptEncodingBrotli = jsii.Bool(true)

	return awscloudfront.NewCachePolicy(scope, jsii.String("CustomCachePolicy"), cachePolicyProps)
}

// Helper function for int32 defaults
func getInt32OrDefault(value int32, defaultValue int32) int32 {
	if value == 0 {
		return defaultValue
	}
	return value
}

func configureOriginRequestPolicy(policy string) awscloudfront.IOriginRequestPolicy {
	switch policy {
	case "MANAGED_ALL_VIEWER":
		return awscloudfront.OriginRequestPolicy_ALL_VIEWER()
	case "MANAGED_CORS_S3":
		return awscloudfront.OriginRequestPolicy_CORS_S3_ORIGIN()
	case "MANAGED_ELEMENT_CAPTURE":
		return awscloudfront.OriginRequestPolicy_ELEMENTAL_MEDIA_TAILOR()
	case "CUSTOM":
		// Custom origin request policy would need to be created separately
		return awscloudfront.OriginRequestPolicy_ALL_VIEWER()
	default:
		return nil // Use default behavior
	}
}

func configureResponseHeadersPolicy(policy string) awscloudfront.IResponseHeadersPolicy {
	switch policy {
	case "MANAGED_CORS_ALLOW_ALL":
		return awscloudfront.ResponseHeadersPolicy_CORS_ALLOW_ALL_ORIGINS()
	case "MANAGED_SECURITY_HEADERS":
		return awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS()
	case "CUSTOM":
		// Custom response headers policy would need to be created separately
		return awscloudfront.ResponseHeadersPolicy_SECURITY_HEADERS()
	default:
		return nil // Use default behavior
	}
}

func configureOriginProtocolPolicy(policy string) awscloudfront.OriginProtocolPolicy {
	switch policy {
	case "HTTP_ONLY":
		return awscloudfront.OriginProtocolPolicy_HTTP_ONLY
	case "HTTPS_ONLY":
		return awscloudfront.OriginProtocolPolicy_HTTPS_ONLY
	case "MATCH_VIEWER":
		return awscloudfront.OriginProtocolPolicy_MATCH_VIEWER
	default:
		return awscloudfront.OriginProtocolPolicy_HTTPS_ONLY
	}
}

func configureMinimumProtocolVersion(version string) awscloudfront.SecurityPolicyProtocol {
	switch version {
	case "TLS_V1_2016":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2016
	case "TLS_V1_1_2016":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_1_2016
	case "TLS_V1_2_2019":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2019
	case "TLS_V1_2_2021":
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
	default:
		return awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021
	}
}

func configureSSLSupportMethod(method string) awscloudfront.SSLMethod {
	switch method {
	case "SNI_ONLY":
		return awscloudfront.SSLMethod_SNI
	case "VIP":
		return awscloudfront.SSLMethod_VIP
	default:
		return awscloudfront.SSLMethod_SNI
	}
}

func configureGeoRestriction(restrictionType string, countries []string) awscloudfront.GeoRestriction {
	converted := jsii.Strings(countries...) // tipo *([]*string)

	switch restrictionType {
	case "ALLOW":
		return awscloudfront.GeoRestriction_Allowlist((*converted)...)
	case "DENY":
		return awscloudfront.GeoRestriction_Denylist((*converted)...)
	default:
		return awscloudfront.GeoRestriction_Allowlist((*converted)...)
	}
}

func configureFunctionEventType(eventType string) awscloudfront.FunctionEventType {
	switch eventType {
	case "VIEWER_REQUEST":
		return awscloudfront.FunctionEventType_VIEWER_REQUEST
	case "VIEWER_RESPONSE":
		return awscloudfront.FunctionEventType_VIEWER_RESPONSE
	default:
		return awscloudfront.FunctionEventType_VIEWER_REQUEST
	}
}

func configureLambdaEventType(eventType string) awscloudfront.LambdaEdgeEventType {
	switch eventType {
	case "ORIGIN_REQUEST":
		return awscloudfront.LambdaEdgeEventType_ORIGIN_REQUEST
	case "ORIGIN_RESPONSE":
		return awscloudfront.LambdaEdgeEventType_ORIGIN_RESPONSE
	case "VIEWER_REQUEST":
		return awscloudfront.LambdaEdgeEventType_VIEWER_REQUEST
	case "VIEWER_RESPONSE":
		return awscloudfront.LambdaEdgeEventType_VIEWER_RESPONSE
	default:
		return awscloudfront.LambdaEdgeEventType_VIEWER_REQUEST
	}
}

func convertToOriginSslProtocols(protocols []string) *[]awscloudfront.OriginSslPolicy {
	sslPolicies := make([]awscloudfront.OriginSslPolicy, 0, len(protocols))

	for _, protocol := range protocols {
		switch protocol {
		case "SSLv3":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_SSL_V3)
		case "TLSv1":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_TLS_V1)
		case "TLSv1.1":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_TLS_V1_1)
		case "TLSv1.2":
			sslPolicies = append(sslPolicies, awscloudfront.OriginSslPolicy_TLS_V1_2)
		}
	}

	return &sslPolicies
}

// Helper utility functions

func convertToStringPointers(strings []string) *[]*string {
	if len(strings) == 0 {
		return nil
	}
	pointers := make([]*string, len(strings))
	for i, s := range strings {
		pointers[i] = jsii.String(s)
	}
	return &pointers
}

func getPortOrDefault(port int32, defaultPort int32) int32 {
	if port == 0 {
		return defaultPort
	}
	return port
}

func getTimeoutOrDefault(timeout int32, defaultTimeout int32) int32 {
	if timeout == 0 {
		return defaultTimeout
	}
	return timeout
}

func getStringOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func getStringSliceOrDefault(value, defaultValue []string) []string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// Default property factories for common use cases

// DefaultS3StaticWebsiteProps returns defaults for an S3 static website (public website endpoint)
func DefaultS3StaticWebsiteProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:           "S3 Static Website Distribution",
		Enabled:           true,
		DefaultRootObject: "index.html",
		PriceClass:        "ALL",
		HttpVersion:       "HTTP2_AND_3",
		EnableIPv6:        true,

		OriginType:   "S3_WEBSITE",
		S3BucketName: "", // completar nombre del bucket

		ViewerProtocolPolicy: "REDIRECT_TO_HTTPS",
		AllowedMethods:       []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:        []string{"GET", "HEAD"},
		CompressResponse:     true,

		CachePolicy:           "MANAGED_CACHING_OPTIMIZED",
		OriginRequestPolicy:   "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",

		EnableErrorPages: true,
		ErrorPageConfigs: []ErrorPageConfig{
			{ErrorCode: 403, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
			{ErrorCode: 404, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
		},
	}
}

// DefaultS3PrivateOACProps returns defaults for serving private S3 content via OAC
func DefaultS3PrivateOACProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:     "S3 Private Content with OAC",
		Enabled:     true,
		PriceClass:  "ALL",
		HttpVersion: "HTTP2_AND_3",
		EnableIPv6:  true,

		OriginType:            "S3",
		S3BucketName:          "", // completar nombre del bucket
		DefaultRootObject:     "index.html",
		ViewerProtocolPolicy:  "REDIRECT_TO_HTTPS",
		AllowedMethods:        []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:         []string{"GET", "HEAD"},
		CompressResponse:      true,
		CachePolicy:           "MANAGED_CACHING_OPTIMIZED",
		OriginRequestPolicy:   "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",
		EnableErrorPages:      true,

		ErrorPageConfigs: []ErrorPageConfig{
			// OAC a menudo devuelve 403 para objetos que faltan. Debemos capturarlo.
			{ErrorCode: 403, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
			{ErrorCode: 404, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
		},
	}
}

// DefaultHttpApiProps returns defaults for a public HTTP(s) API on a custom origin
func DefaultHttpApiProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:     "HTTP API behind CloudFront",
		Enabled:     true,
		PriceClass:  "ALL",
		HttpVersion: "HTTP2_AND_3",
		EnableIPv6:  true,

		OriginType:           "HTTP",
		OriginDomainName:     "", // completar dominio del API o custom origin
		OriginProtocolPolicy: "HTTPS_ONLY",
		OriginPort:           443,

		ViewerProtocolPolicy:  "REDIRECT_TO_HTTPS",
		AllowedMethods:        []string{"GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"},
		CachedMethods:         []string{"GET", "HEAD"},
		CompressResponse:      true,
		CachePolicy:           "MANAGED_CACHING_DISABLED", // APIs usualmente no se cachean por defecto
		OriginRequestPolicy:   "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",
	}
}

// DefaultAlbApiProps returns defaults for an API served by an ALB
func DefaultAlbApiProps() CloudFrontProperties {
	props := DefaultHttpApiProps()
	props.Comment = "ALB API behind CloudFront"
	props.OriginType = "LOAD_BALANCER"
	// OriginDomainName debe ser el DNS del ALB
	return props
}

// DefaultSpaProps returns defaults for Single Page Applications (SPA) on S3 website hosting
func DefaultSpaProps() CloudFrontProperties {
	props := DefaultS3StaticWebsiteProps()
	props.Comment = "SPA on S3 Website with SPA fallbacks"
	props.DefaultRootObject = "index.html"
	props.EnableErrorPages = true
	props.ErrorPageConfigs = []ErrorPageConfig{
		{ErrorCode: 403, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
		{ErrorCode: 404, ResponseCode: 200, ResponsePagePath: "/index.html", ErrorCachingMinTTL: 0},
	}
	return props
}

// DefaultMediaStreamingProps returns defaults optimized for media streaming workloads
func DefaultMediaStreamingProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:         "Media streaming optimized distribution",
		Enabled:         true,
		PriceClass:      "ALL",
		HttpVersion:     "HTTP2_AND_3",
		EnableIPv6:      true,
		SmoothStreaming: true,

		OriginType:       "HTTP",
		OriginDomainName: "", // completar origen del media server/packager

		ViewerProtocolPolicy:  "REDIRECT_TO_HTTPS",
		AllowedMethods:        []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:         []string{"GET", "HEAD"},
		CompressResponse:      true,
		CachePolicy:           "MANAGED_CACHING_OPTIMIZED",
		OriginRequestPolicy:   "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",
	}
}

// DefaultPrivateSignedContentProps returns defaults for private content using signed URLs/cookies
func DefaultPrivateSignedContentProps() CloudFrontProperties {
	return CloudFrontProperties{
		Comment:     "Private content with signed URLs/cookies",
		Enabled:     true,
		PriceClass:  "ALL",
		HttpVersion: "HTTP2_AND_3",
		EnableIPv6:  true,

		OriginType:            "S3",
		S3BucketName:          "", // completar nombre del bucket
		ViewerProtocolPolicy:  "REDIRECT_TO_HTTPS",
		AllowedMethods:        []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:         []string{"GET", "HEAD"},
		CompressResponse:      true,
		CachePolicy:           "MANAGED_CACHING_OPTIMIZED",
		OriginRequestPolicy:   "MANAGED_ALL_VIEWER",
		ResponseHeadersPolicy: "MANAGED_SECURITY_HEADERS",

		// El consumidor debe configurar TrustedSigners o TrustedKeyGroups
		TrustedSigners:   []string{},
		TrustedKeyGroups: []string{},
	}
}

// sanitizeID limpia una cadena para usarla como ID de Constructo (reemplaza caracteres no permitidos)
func sanitizeID(s string) string {
	s = strings.ReplaceAll(s, "/*", "All")
	s = strings.ReplaceAll(s, "*", "Any")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}
