package waf

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// WAFBotControlStrategy implements WAF Web ACL with advanced bot detection and mitigation
// This is the PREMIUM profile with AWS Bot Control Managed Rule Group
//
// Security Model:
// - Advanced bot detection using machine learning
// - Verified bot allowlist (Google, Bing, etc.)
// - CAPTCHA challenges for suspected bots
// - Rate limiting per bot category
// - All baseline protections (OWASP, SQL, etc.)
//
// Use Cases:
// - E-commerce sites (prevent inventory hoarding)
// - High-value web applications
// - Sites with bot scraping problems
// - Applications requiring human verification
// - Ticket sales, limited inventory items
//
// Cost Estimate (PREMIUM):
// - Web ACL: $5/month
// - Bot Control: $10/month
// - Core Rule Set: $1/month
// - SQL Database: $1/month
// - Known Bad Inputs: $1/month
// - IP Reputation: $1/month
// - Anonymous IP List: $1/month
// - Requests: $0.60 per 1M requests
// - Bot Control Requests: $1.00 per 1M requests
// Total: ~$20/month + $1.60/1M requests
//
// WARNING: Bot Control is significantly more expensive than other profiles.
// Use only when bot protection is critical to business operations.
type WAFBotControlStrategy struct{}

// Build creates a WAF Web ACL configured with advanced bot control
func (s *WAFBotControlStrategy) Build(scope constructs.Construct, id string, props WAFFactoryProps) awswafv2.CfnWebACL {

	// =============================================================================
	// WEB ACL CONFIGURATION - Bot Control Protection
	// =============================================================================

	// Determine scope (CLOUDFRONT or REGIONAL)
	wafScope := "CLOUDFRONT"
	if props.Scope == ScopeRegional {
		wafScope = "REGIONAL"
	}

	// Default name
	webACLName := props.Name
	if webACLName == "" {
		webACLName = id + "-BotControl-WebACL"
	}

	// Build rules array
	rules := make([]interface{}, 0)
	priority := int64(0)

	// =============================================================================
	// RULE 1: Strict Rate Limiting for Bot Protection
	// Lower threshold to catch aggressive bots
	// =============================================================================
	rateLimitValue := int64(500) // Stricter default for bot control
	if props.RateLimitRequests != nil && *props.RateLimitRequests > 0 {
		rateLimitValue = *props.RateLimitRequests
	}

	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("BotControlRateLimitRule"),
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
					ResponseCode: jsii.Number(429),
				},
			},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("BotControlRateLimitRule"),
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
			Name:     jsii.String("BotControlGeoBlockingRule"),
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
				MetricName:               jsii.String("BotControlGeoBlockingRule"),
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
			Description:      jsii.String("Bot Control blocked IP addresses"),
		})

		rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
			Name:     jsii.String("BotControlIPBlocklistRule"),
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
				MetricName:               jsii.String("BotControlIPBlocklistRule"),
			},
		})
		priority++
	}

	// =============================================================================
	// AWS MANAGED RULE GROUPS - Bot Control Stack
	// =============================================================================

	// RULE 4: AWS Managed Rules - Bot Control (PREMIUM)
	// This is the main bot detection engine
	rules = append(rules, &awswafv2.CfnWebACL_RuleProperty{
		Name:     jsii.String("AWSManagedRulesBotControlRuleSet"),
		Priority: jsii.Number(priority),
		Statement: &awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: &awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName: jsii.String("AWS"),
				Name:       jsii.String("AWSManagedRulesBotControlRuleSet"),
				// ManagedRuleGroupConfigs can be used to configure bot control behavior
				// Example: Configure to allow verified bots (Google, Bing, etc.)
				ManagedRuleGroupConfigs: &[]*awswafv2.CfnWebACL_ManagedRuleGroupConfigProperty{
					{
						// AWS Bot Control configuration
						// InspectionLevel can be "COMMON" or "TARGETED"
						AwsManagedRulesBotControlRuleSet: &awswafv2.CfnWebACL_AWSManagedRulesBotControlRuleSetProperty{
							InspectionLevel: jsii.String("COMMON"), // COMMON is less expensive than TARGETED
						},
					},
				},
			},
		},
		OverrideAction: &awswafv2.CfnWebACL_OverrideActionProperty{
			None: map[string]interface{}{},
		},
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String("AWSManagedRulesBotControlRuleSet"),
		},
	})
	priority++

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

	// RULE 6: AWS Managed Rules - SQL Database Protection
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
	priority++

	// RULE 9: AWS Managed Rules - Anonymous IP List (blocks VPNs, Tor, proxies)
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
	// CREATE WEB ACL WITH CAPTCHA CONFIGURATION
	// =============================================================================

	webACL := awswafv2.NewCfnWebACL(scope, jsii.String(id), &awswafv2.CfnWebACLProps{
		Name:  jsii.String(webACLName),
		Scope: jsii.String(wafScope),

		// Default action: Allow (bots are challenged/blocked by Bot Control rule)
		DefaultAction: &awswafv2.CfnWebACL_DefaultActionProperty{
			Allow: &awswafv2.CfnWebACL_AllowActionProperty{},
		},

		// Rules array
		Rules: &rules,

		// CAPTCHA configuration (optional, for additional bot verification)
		CaptchaConfig: &awswafv2.CfnWebACL_CaptchaConfigProperty{
			ImmunityTimeProperty: &awswafv2.CfnWebACL_ImmunityTimePropertyProperty{
				ImmunityTime: jsii.Number(300), // 5 minutes immunity after solving CAPTCHA
			},
		},

		// Challenge configuration (similar to CAPTCHA but less intrusive)
		ChallengeConfig: &awswafv2.CfnWebACL_ChallengeConfigProperty{
			ImmunityTimeProperty: &awswafv2.CfnWebACL_ImmunityTimePropertyProperty{
				ImmunityTime: jsii.Number(300), // 5 minutes immunity
			},
		},

		// Visibility configuration
		VisibilityConfig: &awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(true),
			CloudWatchMetricsEnabled: jsii.Bool(true),
			MetricName:               jsii.String(webACLName + "-Metrics"),
		},

		// Description
		Description: jsii.String("Bot Control WAF for " + webACLName + " - Advanced Bot Detection with ML, CAPTCHA, and OWASP Protection"),
	})

	return webACL
}
