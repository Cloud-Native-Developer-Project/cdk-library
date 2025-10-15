package addi

import (
	eventbridgeintegrations "cdk-library/constructs/EventBridgeIntegrations"
	guardduty "cdk-library/constructs/GuardDuty"
	golambda "cdk-library/constructs/Lambda"
	s3 "cdk-library/constructs/S3"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewAddiS3ToSFTPStack creates the complete S3 → Lambda → Webhook → SFTP pipeline
//
// Architecture:
//
//	Client → S3 Bucket → EventBridge → Lambda → Webhook (on-premise) → SFTP
//	                                      ↓
//	                                    SQS DLQ
//
// Components:
// - S3 Bucket (Enterprise-grade: KMS, Object Lock, Versioning)
// - EventBridge Rule (filters uploads/ prefix)
// - Lambda Function (Go, ARM64, generates Presigned URLs)
// - Secrets Manager (webhook credentials)
// - SQS DLQ (failed webhook invocations)
// - GuardDuty (optional: S3 protection for anomaly detection)
func NewAddiS3ToSFTPStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// ========== 1. S3 Landing Zone (Enterprise Strategy) ==========
	bucket := s3.NewSimpleStorageServiceFactory(stack, "LandingZone",
		s3.SimpleStorageServiceFactoryProps{
			BucketType: s3.BucketTypeDevelopment,
			BucketName: "addi-landing-zone-dev",
		})

	// ========== 2. Secrets Manager (Webhook Credentials) ==========
	webhookSecret := awssecretsmanager.NewSecret(stack, jsii.String("WebhookCredentials"), &awssecretsmanager.SecretProps{
		SecretName:  jsii.String("addi/webhook-credentials"),
		Description: jsii.String("Webhook endpoint and HMAC secret for on-premise server"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String(`{
				"webhookUrl": "https://360b6fbd06bc.ngrok-free.app/webhook/addi-csv",
				"apiKey": "addi_prod_ak_placeholder"
			}`),
			GenerateStringKey: jsii.String("hmacSecret"),
			PasswordLength:    jsii.Number(64),
		},
	})

	// ========== 3. Lambda Function (Webhook Notifier) ==========
	// Using Lambda construct with optimized defaults (ARM64, 512MB, 30s timeout, X-Ray tracing)
	lambdaFunction := golambda.NewGoLambda(stack, "WebhookNotifier", golambda.GoLambdaProps{
		FunctionName: "addi-webhook-notifier",
		CodePath:     "stacks/addi/lambda/webhook-notifier",
		Description:  jsii.String("Generates S3 Presigned URLs and sends webhook to on-premise server"),
		Environment: &map[string]*string{
			"BUCKET_NAME":            bucket.BucketName(),
			"WEBHOOK_SECRET_ARN":     webhookSecret.SecretArn(),
			"PRESIGNED_URL_EXPIRES":  jsii.String("900"), // 15 minutes
			"MAX_RETRY_ATTEMPTS":     jsii.String("4"),
			"RETRY_EXPONENTIAL_BASE": jsii.String("2"),

			// 🔧 DEVELOPMENT MODE: Uncomment and update with your ngrok URL
			// This bypasses Secrets Manager and sends webhooks directly to your local backend
			// Get your URL by running: ./stacks/addi/backend/get-ngrok-url.sh
			"WEBHOOK_URL_OVERRIDE": jsii.String("https://35b57cefe2cc.ngrok-free.app/webhook/addi-csv"),
		},
	})

	// Grant Lambda permissions to:
	// - Read objects from S3 (to generate Presigned URLs)
	// - Read webhook credentials from Secrets Manager
	bucket.GrantRead(lambdaFunction, jsii.String("uploads/*"))
	webhookSecret.GrantRead(lambdaFunction, nil)

	// ========== 4. EventBridge Integration (S3 → Lambda) ==========
	eventbridgeintegrations.NewEventBridgeIntegrationFactory(
		stack,
		"S3ToLambdaIntegration",
		eventbridgeintegrations.EventBridgeIntegrationFactoryProps{
			IntegrationType: eventbridgeintegrations.IntegrationTypeS3ToLambda,
			S3ToLambdaConfig: &eventbridgeintegrations.S3ToLambdaConfig{
				SourceBucket:     bucket,
				TargetLambda:     lambdaFunction,
				ObjectKeyPrefix:  jsii.String("uploads/"),
				EventTypes:       []string{"Object Created"},
				MaxRetryAttempts: jsii.Number(4),
				MaxEventAge:      awscdk.Duration_Minutes(jsii.Number(15)),
				EnableDLQ:        jsii.Bool(true),
			},
		})

	// ========== 5. GuardDuty (Data Protection Strategy) ==========
	// Monitors S3, EKS, RDS, Lambda, and EBS without runtime agents
	// Cost: ~$15-50/month | Ideal for S3-centric and serverless workloads
	guardduty.NewGuardDutyDetector(stack, "SecurityMonitor",
		guardduty.GuardDutyFactoryProps{
			DetectorType:               guardduty.GuardDutyTypeDataProtection,
			FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
		})

	// ========== Outputs ==========
	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 Landing Zone bucket name"),
		ExportName:  jsii.String("AddiLandingZoneBucket"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LambdaFunctionArn"), &awscdk.CfnOutputProps{
		Value:       lambdaFunction.FunctionArn(),
		Description: jsii.String("Webhook Notifier Lambda ARN"),
		ExportName:  jsii.String("AddiWebhookNotifierArn"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebhookSecretArn"), &awscdk.CfnOutputProps{
		Value:       webhookSecret.SecretArn(),
		Description: jsii.String("Secrets Manager ARN for webhook credentials"),
		ExportName:  jsii.String("AddiWebhookSecretArn"),
	})

	return stack
}
