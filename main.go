package main

import (
	stacks "cdk-library/stacks/website"
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
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
		EnableWAF:   false,
	})

	app.Synth(nil)
}
