package waf

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/constructs-go/constructs/v10"
)

// WAFProfileType defines the type of WAF protection profile to create
type WAFProfileType string

const (
	// ProfileTypeWebApplication creates WAF optimized for web applications (OWASP Top 10 protection)
	ProfileTypeWebApplication WAFProfileType = "WEB_APPLICATION"

	// ProfileTypeAPIProtection creates WAF optimized for API endpoints (rate limiting, SQL injection)
	ProfileTypeAPIProtection WAFProfileType = "API_PROTECTION"

	// ProfileTypeBotControl creates WAF with advanced bot detection and mitigation
	ProfileTypeBotControl WAFProfileType = "BOT_CONTROL"

	// TODO: Add more profile types as we implement them
	// ProfileTypeWordPress     WAFProfileType = "WORDPRESS"
	// ProfileTypeCustom        WAFProfileType = "CUSTOM"
)

// WAFScope defines whether the WAF is for CloudFront (global) or regional resources
type WAFScope string

const (
	// ScopeCloudFront is for CloudFront distributions (must be in us-east-1)
	ScopeCloudFront WAFScope = "CLOUDFRONT"

	// ScopeRegional is for regional resources (ALB, API Gateway, etc.)
	ScopeRegional WAFScope = "REGIONAL"
)

// WAFFactoryProps defines properties for creating a WAF Web ACL via Factory
type WAFFactoryProps struct {
	// Required: WAF scope (CloudFront or Regional)
	Scope WAFScope

	// Required: Security profile type
	ProfileType WAFProfileType

	// Optional: Custom name for the Web ACL
	Name string

	// Optional: Rate limiting (requests per 5 minutes per IP)
	RateLimitRequests *int64

	// Optional: Countries to block (ISO 3166-1 alpha-2 codes)
	GeoBlockCountries []string

	// Optional: Countries to allow (all others blocked)
	GeoAllowCountries []string

	// Optional: IP addresses to block (CIDR notation)
	BlockedIPs []string

	// Optional: IP addresses to always allow (whitelist)
	AllowedIPs []string

	// Optional: Enable request body inspection (increases costs)
	InspectRequestBody *bool

	// Optional: Enable CloudWatch metrics
	EnableMetrics *bool

	// Optional: Enable sampled request logging
	EnableSampledRequests *bool
}

// NewWebApplicationFirewallFactory creates a WAF Web ACL using the Factory + Strategy pattern
//
// This factory selects the appropriate strategy based on ProfileType and delegates
// Web ACL creation to the specialized strategy implementation.
//
// Example usage:
//
//	webACL := waf.NewWebApplicationFirewallFactory(stack, "WebsiteWAF",
//	    waf.WAFFactoryProps{
//	        Scope: waf.ScopeCloudFront,
//	        ProfileType: waf.ProfileTypeWebApplication,
//	        RateLimitRequests: jsii.Int64(2000),
//	    })
func NewWebApplicationFirewallFactory(scope constructs.Construct, id string, props WAFFactoryProps) awswafv2.CfnWebACL {
	var strategy WebApplicationFirewallStrategy

	// Select strategy based on profile type
	switch props.ProfileType {
	case ProfileTypeWebApplication:
		strategy = &WAFWebApplicationStrategy{}

	case ProfileTypeAPIProtection:
		strategy = &WAFAPIProtectionStrategy{}

	case ProfileTypeBotControl:
		strategy = &WAFBotControlStrategy{}

	// TODO: Implement additional strategies
	// case ProfileTypeWordPress:
	//     strategy = &WAFWordPressStrategy{}
	// case ProfileTypeCustom:
	//     strategy = &WAFCustomStrategy{}

	default:
		panic(fmt.Sprintf("Unsupported WAF ProfileType: %s", props.ProfileType))
	}

	// Delegate Web ACL creation to selected strategy
	return strategy.Build(scope, id, props)
}
