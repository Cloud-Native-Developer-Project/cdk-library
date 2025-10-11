package waf

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// WAFAPIProtectionStrategy implements WAF Web ACL optimized for API endpoints
// This profile focuses on API-specific threats like SQL injection, high request rates, and body inspection
//
// Security Model:
// - SQL injection protection
// - OWASP Top 10 API protection
// - Higher rate limits (APIs handle more traffic)
// - Request body inspection enabled by default
// - Size constraints on request body
//
// Use Cases:
// - REST APIs (API Gateway, ALB)
// - GraphQL APIs
// - Backend APIs for SPAs
// - Microservices endpoints
//
// Cost Estimate:
// - Web ACL: $5/month
// - Core Rule Set: $1/month
// - SQL Database: $1/month
// - Known Bad Inputs: $1/month
// - IP Reputation: $1/month
// - Rate Limit Rule: $1/month
// - Requests: $0.60 per 1M requests
// Total: ~$10/month + $0.60/1M requests
type WAFAPIProtectionStrategy struct{}

// Build creates a WAF Web ACL configured for API protection
func (s *WAFAPIProtectionStrategy) Build(scope constructs.Construct, id string, props WAFFactoryProps) awswafv2.CfnWebACL {

	// =============================================================================
	// WEB ACL CONFIGURATION - API Protection
	// =============================================================================

	// Determine scope (CLOUDFRONT or REGIONAL)
	wafScope := "CLOUDFRONT"
	if props.Scope == ScopeRegional {
		wafScope = "REGIONAL"
	}

	// Default name
	webACLName := props.Name
	if webACLName == "" {
		webACLName = id + "-API-WebACL"
	}

	// Build rules array
	rules := make([]interface{}, 0)
	priority := int64(0)

	// =============================================================================
	// RULE 1: Rate Limiting (default 10,000 req/5min for APIs)
	// Higher threshold than web apps since APIs handle more legitimate traffic
	// =============================================================================
	rateLimitValue := int64(10000) // Default for APIs
	if props.RateLimitRequests != nil && *props.RateLimitRequests > 0 {
		rateLimitValue = *props.RateLimitRequests
	}

	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("APIRateLimitRule"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			RateBasedStatement: &awswafv2.CfnWebACL_RateBasedStatementProperty{
				Limit:            jsii.Number(float64(rateLimitValue)),
				AggregateKeyType: jsii.String("IP"),
			},
		},
		Action: &awswafv2.CfnWebACL_RuleActionProperty{
			Block: &awswafv2.CfnWebACL_BlockActionProperty{
				CustomResponse: &awswafv2.CfnWebACL_CustomResponseProperty{
					ResponseCode: jsii.Number(429), // Too Many Requests
				},
			},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("APIRateLimitRule"),
		},
	})
	priority++

	// =============================================================================
	// RULE 2: Geo Blocking (if specified)
	// =============================================================================
	if len(props.GeoBlockCountries) > 0 {
		countryCodes := make([]*string, len(props.GeoBlockCountries))
		for i, code := range props.GeoBlockCountries {
			countryCodes[i] = jsii.String(code)
		}

		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("APIGeoBlockingRule"),
			Priority: jsii.Number(priority),
			Statement: &awswafv2.CfnWebACL_StatementProperty{
				GeoMatchStatement: &awswafv2.CfnWebACL_GeoMatchStatementProperty{
					CountryCodes: &countryCodes,
				},
			},
			Action: &awswafv2.CfnWebACL_RuleActionProperty{
				Block: &awswafv2.CfnWebACL_BlockActionProperty{},
			},
			VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
				SampledRequestsEnabled:   jsii.Bool(true),
				CloudWatchMetricsEnabled: jsii.Bool(true),
				MetricName:               jsii.String("APIGeoBlockingRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// RULE 3: IP Blocklist (if specified)
	// =============================================================================
	if len(props.BlockedIPs) > 0 {
		ipSet := awswafv2.NewCfnIPSet(scope, jsii.String(id+"BlockedIPSet"), &awswafv2.CfnIPSetProps{
			Name:             jsii.String(webACLName + "-BlockedIPs"),
			Scope:            jsii.String(wafScope),
			IpAddressVersion: jsii.String("IPV4"),
			Addresses:        jsii.Strings(props.BlockedIPs...),
			Description:      jsii.String("API blocked IP addresses"),
		})

		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("APIIPBlocklistRule"),
			Priority: jsii.Number(priority),
			Statement: &awswafv2.CfnWebACL_StatementProperty{
				IpSetReferenceStatement: &awswafv2.CfnWebACL_IPSetReferenceStatementProperty{
					Arn: ipSet.AttrArn(),
				},
			},
			Action: &awswafv2.CfnWebACL_RuleActionProperty{
				Block: &awswafv2.CfnWebACL_BlockActionProperty{},
			},
			VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
				SampledRequestsEnabled:   jsii.Bool(true),
				CloudWatchMetricsEnabled: jsii.Bool(true),
				MetricName:               jsii.String("APIIPBlocklistRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// RULE 4: Request Size Constraint (protect against large payloads)
	// Blocks requests with body > 8KB (typical API threshold)
	// =============================================================================
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("APISizeConstraintRule"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			SizeConstraintStatement: &awswafv2.CfnWebACL_SizeConstraintStatementProperty{
				FieldToMatch: &awswafv2.CfnWebACL_FieldToMatchProperty{
					Body: &awswafv2.CfnWebACL_BodyProperty{
						OversizeHandling: jsii.String("CONTINUE"),
					},
				},
				ComparisonOperator: jsii.String("GT"),
				Size:               jsii.Number(8192), // 8KB limit
				TextTransformations: &[]*awswafv2.CfnWebACL_TextTransformationProperty{
					{
						Priority: jsii.Number(0),
						Type:     jsii.String("NONE"),
					},
				},
			},
		},
		Action: &awswafv2.CfnWebACL_RuleActionProperty{
			Block: &awswafv2.CfnWebACL_BlockActionProperty{
				CustomResponse: &awswafv2.CfnWebACL_CustomResponseProperty{
					ResponseCode: jsii.Number(413), // Payload Too Large
				},
			},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("APISizeConstraintRule"),
		},
	})
	priority++

	// =============================================================================
	// AWS MANAGED RULE GROUPS - API Focused
	// =============================================================================

	// RULE 5: AWS Managed Rules - Core Rule Set (OWASP Top 10)
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesCommonRuleSet"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesCommonRuleSet"),
			},
		},
		OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
			None: map[string]interface{}{},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("AWSManagedRulesCommonRuleSet"),
		},
	})
	priority++

	// RULE 6: AWS Managed Rules - SQL Database (Critical for APIs)
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesSQLiRuleSet"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesSQLiRuleSet"),
			},
		},
		OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
			None: map[string]interface{}{},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("AWSManagedRulesSQLiRuleSet"),
		},
	})
	priority++

	// RULE 7: AWS Managed Rules - Known Bad Inputs
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesKnownBadInputsRuleSet"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesKnownBadInputsRuleSet"),
			},
		},
		OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
			None: map[string]interface{}{},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("AWSManagedRulesKnownBadInputsRuleSet"),
		},
	})
	priority++

	// RULE 8: AWS Managed Rules - Amazon IP Reputation List
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesAmazonIpReputationList"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesAmazonIpReputationList"),
			},
		},
		OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
			None: map[string]interface{}{},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("AWSManagedRulesAmazonIpReputationList"),
		},
	})

	// =============================================================================
	// CREATE WEB ACL
	// =============================================================================

	webACL := awswafv2.NewCfnWebACL(scope, jsii.String(id), &awswafv2.CfnWebACLProps{
		Name:  jsii.String(webACLName),
		Scope: jsii.String(wafScope),

		// Default action: Allow
		DefaultAction: &awswafv2.CfnWebACL_DefaultActionProperty{
			Allow: &awswafv2.CfnWebACL_AllowActionProperty{},
		},

		// Rules array
		Rules: &rules,

		// Visibility configuration
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String(webACLName + "-Metrics"),
		},

		// Description
		Description: jsii.String("API Protection WAF for " + webACLName + " - SQL Injection, Rate Limiting, Body Inspection"),
	})

	return webACL
}
