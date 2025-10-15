package main

import (
	"fmt"
	"os"

	addistack "cdk-library/stacks/addi"

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

	// =============================================================================
	// ADDI - S3 TO SFTP PIPELINE
	// =============================================================================
	addistack.NewAddiS3ToSFTPStack(app, "AddiStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String(region),
		},
		StackName:   jsii.String("addi-s3-to-sftp-pipeline"),
		Description: jsii.String("Addi CSV processing pipeline: S3 → EventBridge → Lambda → Webhook → SFTP"),
		Tags: &map[string]*string{
			"Environment": jsii.String("Production"),
			"Project":     jsii.String("Addi"),
			"ManagedBy":   jsii.String("CDK"),
			"CostCenter":  jsii.String("Operations"),
		},
	})

	app.Synth(nil)
}
