package waf

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/constructs-go/constructs/v10"
)

// WebApplicationFirewallStrategy defines the contract for WAF Web ACL creation strategies
// Each strategy implements a specific security profile (Web Application, API Protection, Bot Control, etc.)
type WebApplicationFirewallStrategy interface {
	Build(scope constructs.Construct, id string, props WAFFactoryProps) awswafv2.CfnWebACL
}
