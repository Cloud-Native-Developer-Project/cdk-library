package waf

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// WAFWebApplicationStrategy implements WAF Web ACL optimized for web applications
// This is the RECOMMENDED profile for static websites, SPAs, and general web applications
//
// Security Model:
// - OWASP Top 10 protection via AWS Managed Rules
// - Known bad inputs blocking
// - IP reputation filtering
// - Rate limiting per IP
// - Optional geo-blocking
//
// Use Cases:
// - Static websites (React, Vue, Angular)
// - Single Page Applications (SPAs)
// - JAMstack sites
// - General web applications without complex APIs
//
// Cost Estimate:
// - Web ACL: $5/month
// - Core Rule Set: $1/month
// - Known Bad Inputs: $1/month
// - IP Reputation: $1/month
// - Rate Limit Rule: $1/month (if enabled)
// - Requests: $0.60 per 1M requests
// Total: ~$9-10/month + $0.60/1M requests
type WAFWebApplicationStrategy struct{}

// Build creates a WAF Web ACL configured for web application protection
func (s *WAFWebApplicationStrategy) Build(scope constructs.Construct, id string, props WAFFactoryProps) awswafv2.CfnWebACL {

	// =============================================================================
	// WEB ACL CONFIGURATION - Web Application Protection
	// =============================================================================

	// Determine scope (CLOUDFRONT or REGIONAL)
	wafScope := "CLOUDFRONT"
	if props.Scope == ScopeRegional {
		wafScope = "REGIONAL"
	}

	// Default name
	webACLName := props.Name
	if webACLName == "" {
		webACLName = id + "-WebACL"
	}

	// Build rules array
	rules := make([]interface{}, 0)
	priority := int64(0)

	// =============================================================================
	// RULE 1: Rate Limiting (if specified)
	// Blocks IPs that exceed request threshold in 5-minute window
	// =============================================================================
	if props.RateLimitRequests != nil && *props.RateLimitRequests > 0 {
		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("RateLimitRule"),
			Priority: jsii.Number(priority),
			Statement: &awswafv2.CfnWebACL_StatementProperty{
				RateBasedStatement: &awswafv2.CfnWebACL_RateBasedStatementProperty{
					Limit:              jsii.Number(*props.RateLimitRequests),
					AggregateKeyType:   jsii.String("IP"),
					ScopeDownStatement: nil, // Apply to all requests
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
				MetricName:               jsii.String("RateLimitRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// RULE 2: Geo Blocking (if specified)
	// Blocks requests from specific countries
	// =============================================================================
	if len(props.GeoBlockCountries) > 0 {
		countryCodes := make([]*string, len(props.GeoBlockCountries))
		for i, code := range props.GeoBlockCountries {
			countryCodes[i] = jsii.String(code)
		}

		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("GeoBlockingRule"),
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
				MetricName:               jsii.String("GeoBlockingRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// RULE 3: IP Blocklist (if specified)
	// Blocks specific IP addresses or CIDR ranges
	// =============================================================================
	if len(props.BlockedIPs) > 0 {
		// Create IP Set
		ipSet := awswafv2.NewCfnIPSet(scope, jsii.String(id+"BlockedIPSet"), &awswafv2.CfnIPSetProps{
			Name:             jsii.String(webACLName + "-BlockedIPs"),
			Scope:            jsii.String(wafScope),
			IpAddressVersion: jsii.String("IPV4"),
			Addresses:        jsii.Strings(props.BlockedIPs...),
			Description:      jsii.String("Blocked IP addresses"),
		})

		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("IPBlocklistRule"),
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
				MetricName:               jsii.String("IPBlocklistRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// RULE 4: IP Allowlist (if specified)
	// Allows specific IP addresses to bypass all other rules
	// =============================================================================
	if len(props.AllowedIPs) > 0 {
		// Create IP Set
		ipSet := awswafv2.NewCfnIPSet(scope, jsii.String(id+"AllowedIPSet"), &awswafv2.CfnIPSetProps{
			Name:             jsii.String(webACLName + "-AllowedIPs"),
			Scope:            jsii.String(wafScope),
			IpAddressVersion: jsii.String("IPV4"),
			Addresses:        jsii.Strings(props.AllowedIPs...),
			Description:      jsii.String("Allowed IP addresses (whitelist)"),
		})

		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("IPAllowlistRule"),
			Priority: jsii.Number(priority),
			Statement: &awswafv2.CfnWebACL_StatementProperty{
				IpSetReferenceStatement: &awswafv2.CfnWebACL_IPSetReferenceStatementProperty{
					Arn: ipSet.AttrArn(),
				},
			},
			Action: &awswafv2.CfnWebACL_RuleActionProperty{
				Allow: &awswafv2.CfnWebACL_AllowActionProperty{},
			},
			VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
				SampledRequestsEnabled:   jsii.Bool(true),
				CloudWatchMetricsEnabled: jsii.Bool(true),
				MetricName:               jsii.String("IPAllowlistRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// AWS MANAGED RULE GROUPS
	// =============================================================================

	// RULE 5: AWS Managed Rules - Core Rule Set (OWASP Top 10)
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesCommonRuleSet"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesCommonRuleSet"),
				// ExcludedRules can be added here if needed
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

	// RULE 6: AWS Managed Rules - Known Bad Inputs
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

	// RULE 7: AWS Managed Rules - Amazon IP Reputation List
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
	priority++

	// RULE 8: AWS Managed Rules - Anonymous IP List (blocks VPNs, Tor, proxies)
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesAnonymousIpList"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesAnonymousIpList"),
			},
		},
		OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
			None: map[string]interface{}{},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("AWSManagedRulesAnonymousIpList"),
		},
	})

	// =============================================================================
	// CREATE WEB ACL
	// =============================================================================

	webACL := awswafv2.NewCfnWebACL(scope, jsii.String(id), &awswafv2.CfnWebACLProps{
		Name:  jsii.String(webACLName),
		Scope: jsii.String(wafScope),

		// Default action: Allow (unless blocked by rules above)
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

		// Optional: Description
		Description: jsii.String("Web Application Firewall for " + webACLName + " - OWASP Top 10 Protection"),
	})

	return webACL
}
