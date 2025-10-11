package main

import (
	stacks "cdk-library/stacks/website"
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	// Importa WAF si deseas habilitar protección
	// waf "cdk-library/constructs/WAF"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Get AWS Account and Region from environment
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	// Validate account is set
	if account == "" {
		fmt.Println("Warning: CDK_DEFAULT_ACCOUNT not set, using default AWS credentials")
		account = "default"
	}
	if region == "" {
		fmt.Println("Warning: CDK_DEFAULT_REGION not set, defaulting to us-east-1")
		region = "us-east-1"
	}

	// Create unique but readable bucket name
	// S3 bucket names must be globally unique and lowercase

	// =============================================================================
	// DEVELOPMENT ENVIRONMENT
	// =============================================================================
	stacks.NewStaticWebsiteStack(app, "DevStaticWebsite", &stacks.StaticWebsiteStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(account),
				Region:  jsii.String(region),
			},
			StackName:   jsii.String("dev-static-website"),
			Description: jsii.String("Development static website with S3 + CloudFront + OAC"),
			Tags: &map[string]*string{
				"Environment": jsii.String("Development"),
				"Project":     jsii.String("StaticWebsite"),
				"ManagedBy":   jsii.String("CDK"),
				"CostCenter":  jsii.String("Engineering"),
			},
		},
		BucketName:  *jsii.String(fmt.Sprintf("dev-static-website-%s", account)),
		WebsiteName: "my-website-dev",
		SourcePath:  "stacks/website/dist",
		PriceClass:  "100",

		// WAF Configuration (Optional)
		EnableWAF: true, // Set to true to enable WAF protection

		// WAF Profile Options (uncomment waf import above):
		// WafProfileType: waf.ProfileTypeWebApplication, // OWASP Top 10, XSS, SQL injection, rate limit 2000 req/5min
		// WafProfileType: waf.ProfileTypeAPIProtection,  // Rate limit 10000 req/5min, SQL injection protection
		// WafProfileType: waf.ProfileTypeBotControl,     // Advanced bot detection, rate limit 5000 req/5min

		// See stacks/website/USAGE_EXAMPLES.md for detailed WAF configuration examples
	})

	app.Synth(nil)
}
