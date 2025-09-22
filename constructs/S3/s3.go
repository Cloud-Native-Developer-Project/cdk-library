package s3

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// S3Properties defines all configurable properties for S3 bucket creation
// Includes security, lifecycle, monitoring, and performance configurations
type S3Properties struct {

	// Basic Configuration
	BucketName        string // Name of the S3 bucket (must be globally unique)
	RemovalPolicy     string // Policy for bucket deletion: "retain", "destroy", or "retain_on_update_or_delete"
	AutoDeleteObjects bool   // Whether to delete all objects when the bucket is deleted

	// Security Configuration
	PublicAccess      bool    // Whether to allow public access (for static websites)
	Encryption        string  // Encryption type: "S3_MANAGED", "KMS", "DSSE"
	BucketKeyEnabled  bool    // Use S3 bucket keys to reduce KMS API calls and costs
	EnforceSSL        bool    // Force HTTPS-only access (recommended for security)
	MinimumTLSVersion float64 // Minimum TLS version (1.2 or 1.3) - requires EnforceSSL

	// Versioning & Object Lock (for compliance and data protection)
	Versioned                      bool   // Enable versioning for object rollback capability
	ObjectLockEnabled              bool   // Enable Object Lock for write-once-read-many compliance
	ObjectLockDefaultRetentionMode string // "GOVERNANCE" or "COMPLIANCE" retention mode
	ObjectLockDefaultRetentionDays int32  // Default retention period in days

	// Lifecycle Management (for cost optimization)
	EnableIntelligentTiering bool                   // Automatically move objects to cost-effective storage classes
	LifecycleRules           []*awss3.LifecycleRule // Custom lifecycle rules for object transitions
	TransitionMinimumSize    string                 // Minimum object size for lifecycle transitions

	// Cross-Region Replication (for disaster recovery)
	ReplicationEnabled      bool   // Enable cross-region replication
	ReplicationDestination  string // Target bucket ARN for replication
	ReplicationStorageClass string // Storage class for replicated objects

	// Logging & Monitoring
	EnableAccessLogs   bool   // Enable server access logging
	AccessLogsBucket   string // Destination bucket for access logs (optional)
	AccessLogsPrefix   string // Prefix for access log files
	EventBridgeEnabled bool   // Send S3 events to EventBridge
	EnableInventory    bool   // Enable S3 inventory reports

	// Performance & Network Optimization
	TransferAcceleration bool     // Enable S3 Transfer Acceleration for faster uploads
	EnableCORS           bool     // Enable Cross-Origin Resource Sharing
	CORSAllowedOrigins   []string // Allowed origins for CORS requests
	CORSAllowedMethods   []string // Allowed HTTP methods for CORS
	CORSAllowedHeaders   []string // Allowed headers for CORS

	// Static Website Hosting
	WebsiteEnabled       bool   // Enable static website hosting
	WebsiteIndexDocument string // Index document for website (e.g., "index.html")
	WebsiteErrorDocument string // Error document for website (e.g., "error.html")

	// Metrics Configuration
	EnableMetrics     bool              // Enable CloudWatch request metrics
	MetricsId         string            // Custom metrics configuration ID
	MetricsPrefix     string            // Monitor only objects with this prefix
	MetricsTagFilters map[string]string // Monitor only objects with these tags
}

// NewBucket creates a new S3 bucket with comprehensive configuration options
// This function applies AWS best practices for security, cost optimization, and performance
func NewBucket(scope constructs.Construct, id string, props S3Properties) awss3.Bucket {
	// Initialize bucket properties with basic configuration
	bucketProps := &awss3.BucketProps{
		BucketName:        jsii.String(props.BucketName),
		RemovalPolicy:     configureRemovalPolicy(props.RemovalPolicy),
		AutoDeleteObjects: jsii.Bool(props.AutoDeleteObjects),
		BlockPublicAccess: configureBlockPublicAccess(props.PublicAccess),
	}

	// Configure security settings
	configureSecurity(bucketProps, props)

	// Configure versioning and object lock
	configureVersioningAndObjectLock(bucketProps, props)

	// Configure lifecycle management
	configureLifecycleManagement(bucketProps, props)

	// Configure logging and monitoring
	configureLoggingAndMonitoring(bucketProps, props)

	// Configure performance and network settings
	configurePerformanceAndNetwork(bucketProps, props)

	// Configure website hosting if enabled
	configureWebsiteHosting(bucketProps, props)

	// Create and return the bucket
	bucket := awss3.NewBucket(scope, jsii.String(id), bucketProps)

	return bucket
}

// configureSecurity applies security-related settings to the bucket
func configureSecurity(bucketProps *awss3.BucketProps, props S3Properties) {
	// Set encryption configuration
	if props.Encryption != "" {
		bucketProps.Encryption = configureEncryption(props.Encryption)
	}

	// Enable S3 bucket keys to reduce KMS costs
	if props.BucketKeyEnabled {
		bucketProps.BucketKeyEnabled = jsii.Bool(true)
	}

	// Enforce SSL for all requests (security best practice)
	if props.EnforceSSL {
		bucketProps.EnforceSSL = jsii.Bool(true)

		// Set minimum TLS version if specified
		if props.MinimumTLSVersion > 0 {
			bucketProps.MinimumTLSVersion = jsii.Number(props.MinimumTLSVersion)
		}
	}

}

// configureVersioningAndObjectLock sets up versioning and compliance features
func configureVersioningAndObjectLock(bucketProps *awss3.BucketProps, props S3Properties) {
	// Enable versioning for object rollback capability
	if props.Versioned {
		bucketProps.Versioned = jsii.Bool(true)
	}

	// Configure Object Lock for compliance requirements
	if props.ObjectLockEnabled {
		bucketProps.ObjectLockEnabled = jsii.Bool(true)

		// Set default retention if specified
		if props.ObjectLockDefaultRetentionMode != "" && props.ObjectLockDefaultRetentionDays > 0 {
			bucketProps.ObjectLockDefaultRetention = configureObjectLockRetention(
				props.ObjectLockDefaultRetentionMode,
				props.ObjectLockDefaultRetentionDays,
			)
		}
	}
}

// configureLifecycleManagement sets up cost optimization through lifecycle policies
func configureLifecycleManagement(bucketProps *awss3.BucketProps, props S3Properties) {
	// Enable Intelligent Tiering for automatic cost optimization
	if props.EnableIntelligentTiering {
		bucketProps.IntelligentTieringConfigurations = &[]*awss3.IntelligentTieringConfiguration{
			{
				Name:                      jsii.String("EntireBucket"),
				ArchiveAccessTierTime:     awscdk.Duration_Days(jsii.Number(90)),
				DeepArchiveAccessTierTime: awscdk.Duration_Days(jsii.Number(180)),
			},
		}
	}

	// Apply custom lifecycle rules if provided
	if len(props.LifecycleRules) > 0 {
		bucketProps.LifecycleRules = &props.LifecycleRules
	}

	// Set transition minimum object size if specified
	if props.TransitionMinimumSize != "" {
		bucketProps.TransitionDefaultMinimumObjectSize = configureTransitionMinimumSize(props.TransitionMinimumSize)
	}
}

// configureLoggingAndMonitoring sets up observability and compliance logging
func configureLoggingAndMonitoring(bucketProps *awss3.BucketProps, props S3Properties) {
	// Configure server access logs
	if props.EnableAccessLogs {
		if props.AccessLogsBucket != "" {
			// Use external bucket for access logs
			bucketProps.ServerAccessLogsBucket = awss3.Bucket_FromBucketName(
				nil, jsii.String("AccessLogsBucket"), jsii.String(props.AccessLogsBucket),
			)
		}
		if props.AccessLogsPrefix != "" {
			bucketProps.ServerAccessLogsPrefix = jsii.String(props.AccessLogsPrefix)
		}
	}

	// Enable EventBridge for event-driven architectures
	if props.EventBridgeEnabled {
		bucketProps.EventBridgeEnabled = jsii.Bool(true)
	}

	// Configure S3 inventory for object reporting
	if props.EnableInventory {
		bucketProps.Inventories = &[]*awss3.Inventory{
			{
				Enabled:               jsii.Bool(true),
				IncludeObjectVersions: awss3.InventoryObjectVersion_CURRENT,
				Frequency:             awss3.InventoryFrequency_DAILY,
			},
		}
	}

	// Configure CloudWatch request metrics
	if props.EnableMetrics {
		metricsConfig := &awss3.BucketMetrics{
			Id: jsii.String(getMetricsId(props.MetricsId)), // Use custom ID or default
		}

		// Add prefix filter if specified
		if props.MetricsPrefix != "" {
			metricsConfig.Prefix = jsii.String(props.MetricsPrefix)
		}

		// Add tag filters if specified
		if len(props.MetricsTagFilters) > 0 {
			tagFilters := make(map[string]interface{})
			for key, value := range props.MetricsTagFilters {
				tagFilters[key] = value
			}
			metricsConfig.TagFilters = &tagFilters
		}

		bucketProps.Metrics = &[]*awss3.BucketMetrics{metricsConfig}
	}
}

// getMetricsId returns a metrics ID, using custom ID if provided or default
func getMetricsId(customId string) string {
	if customId != "" {
		return customId
	}
	return "EntireBucket" // Default ID for monitoring all objects
}

// configurePerformanceAndNetwork sets up performance and network optimizations
func configurePerformanceAndNetwork(bucketProps *awss3.BucketProps, props S3Properties) {
	// Enable Transfer Acceleration for faster uploads
	if props.TransferAcceleration {
		bucketProps.TransferAcceleration = jsii.Bool(true)
	}

	// Configure CORS for web applications
	if props.EnableCORS {
		corsRules := []*awss3.CorsRule{}

		// Create CORS rule with specified origins or default
		origins := props.CORSAllowedOrigins
		if len(origins) == 0 {
			origins = []string{"*"} // Default to all origins if none specified
		}

		methods := props.CORSAllowedMethods
		if len(methods) == 0 {
			methods = []string{"GET", "POST", "PUT", "DELETE", "HEAD"} // Default methods
		}

		headers := props.CORSAllowedHeaders
		if len(headers) == 0 {
			headers = []string{"*"} // Default to all headers
		}

		corsRule := &awss3.CorsRule{
			AllowedOrigins: jsii.Strings(origins...),
			AllowedMethods: convertToHttpMethods(methods),
			AllowedHeaders: jsii.Strings(headers...),
			MaxAge:         jsii.Number(3000), // Cache preflight response for 50 minutes
		}

		corsRules = append(corsRules, corsRule)
		bucketProps.Cors = &corsRules
	}
}

// configureWebsiteHosting sets up static website hosting
func configureWebsiteHosting(bucketProps *awss3.BucketProps, props S3Properties) {
	if props.WebsiteEnabled {
		// Set index document (required for website hosting)
		if props.WebsiteIndexDocument != "" {
			bucketProps.WebsiteIndexDocument = jsii.String(props.WebsiteIndexDocument)
		}

		// Set error document (optional but recommended)
		if props.WebsiteErrorDocument != "" {
			bucketProps.WebsiteErrorDocument = jsii.String(props.WebsiteErrorDocument)
		}
	}
}

// configureRemovalPolicy converts string policy to CDK RemovalPolicy enum
func configureRemovalPolicy(policy string) awscdk.RemovalPolicy {
	switch policy {
	case "retain":
		return awscdk.RemovalPolicy_RETAIN_ON_UPDATE_OR_DELETE
	case "destroy":
		return awscdk.RemovalPolicy_DESTROY
	case "retain_on_update_or_delete":
		return awscdk.RemovalPolicy_RETAIN_ON_UPDATE_OR_DELETE
	default:
		return awscdk.RemovalPolicy_RETAIN // Safest default
	}
}

// configureBlockPublicAccess sets up public access controls based on use case
func configureBlockPublicAccess(publicAccess bool) awss3.BlockPublicAccess {
	if publicAccess {
		// Allow public access through bucket policies (for static websites)
		// Still blocks public ACLs for security
		return awss3.BlockPublicAccess_BLOCK_ACLS_ONLY()
	}

	// Most secure option: block all public access
	return awss3.BlockPublicAccess_BLOCK_ALL()
}

// configureEncryption converts string encryption type to CDK BucketEncryption enum
func configureEncryption(encType string) awss3.BucketEncryption {
	switch encType {
	case "KMS":
		return awss3.BucketEncryption_KMS_MANAGED
	case "DSSE":
		return awss3.BucketEncryption_DSSE_MANAGED
	case "S3_MANAGED":
		return awss3.BucketEncryption_S3_MANAGED
	default:
		return awss3.BucketEncryption_S3_MANAGED // Default to S3-managed encryption
	}
}

// configureObjectLockRetention sets up Object Lock retention policy
func configureObjectLockRetention(mode string, days int32) awss3.ObjectLockRetention {
	duration := awscdk.Duration_Days(jsii.Number(days))

	switch mode {
	case "COMPLIANCE":
		// Compliance mode: objects cannot be deleted by any user during retention period
		return awss3.ObjectLockRetention_Compliance(duration)
	case "GOVERNANCE":
		// Governance mode: objects can be deleted with special permissions
		return awss3.ObjectLockRetention_Governance(duration)
	default:
		// Default to governance mode (less restrictive)
		return awss3.ObjectLockRetention_Governance(duration)
	}
}

// configureTransitionMinimumSize sets the minimum object size for lifecycle transitions
func configureTransitionMinimumSize(size string) awss3.TransitionDefaultMinimumObjectSize {
	switch size {
	case "ALL_STORAGE_CLASSES_128_K":
		return awss3.TransitionDefaultMinimumObjectSize_ALL_STORAGE_CLASSES_128_K
	case "VARIES_BY_STORAGE_CLASS":
		return awss3.TransitionDefaultMinimumObjectSize_VARIES_BY_STORAGE_CLASS
	default:
		return awss3.TransitionDefaultMinimumObjectSize_ALL_STORAGE_CLASSES_128_K
	}
}

// convertToHttpMethods converts string slice to CDK HttpMethods slice
func convertToHttpMethods(methods []string) *[]awss3.HttpMethods {
	httpMethods := make([]awss3.HttpMethods, 0, len(methods))

	for _, method := range methods {
		switch method {
		case "GET":
			httpMethods = append(httpMethods, awss3.HttpMethods_GET)
		case "POST":
			httpMethods = append(httpMethods, awss3.HttpMethods_POST)
		case "PUT":
			httpMethods = append(httpMethods, awss3.HttpMethods_PUT)
		case "DELETE":
			httpMethods = append(httpMethods, awss3.HttpMethods_DELETE)
		case "HEAD":
			httpMethods = append(httpMethods, awss3.HttpMethods_HEAD)
		}
	}

	return &httpMethods
}

// GetDefaultProperties returns a S3Properties struct with recommended default values
// Use this as a starting point and customize based on your specific requirements
func GetDefaultProperties() S3Properties {
	return S3Properties{
		// Basic Configuration - customize these
		BucketName:        "", // Must be provided by caller
		RemovalPolicy:     "retain",
		AutoDeleteObjects: false,

		// Security - recommended secure defaults
		PublicAccess:      false,
		Encryption:        "S3_MANAGED",
		BucketKeyEnabled:  true,
		EnforceSSL:        true,
		MinimumTLSVersion: 1.2,

		// Versioning - enabled for data protection
		Versioned:         true,
		ObjectLockEnabled: false, // Enable only if compliance is required

		// Lifecycle Management - enabled for cost optimization
		EnableIntelligentTiering: true,
		TransitionMinimumSize:    "ALL_STORAGE_CLASSES_128_K",

		// Replication - disabled by default
		ReplicationEnabled: false,

		// Logging & Monitoring - basic monitoring enabled
		EnableAccessLogs:   false, // Enable in production
		EventBridgeEnabled: false,
		EnableInventory:    false,
		EnableMetrics:      false,

		// Performance - basic settings
		TransferAcceleration: false,
		EnableCORS:           false,

		// Website Hosting - disabled by default
		WebsiteEnabled: false,
	}
}

// GetEnterpriseDataProperties returns a configuration for a secure enterprise data bucket
// Optimized for: Financial data, PII, regulated industries, compliance requirements
func GetEnterpriseDataProperties() S3Properties {
	return S3Properties{
		// Basic Configuration
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "retain",
		AutoDeleteObjects: false,

		// Enhanced Security Configuration
		PublicAccess:      false, // Never allow public access
		Encryption:        "KMS", // Use KMS for maximum control
		BucketKeyEnabled:  true,  // Reduce KMS costs
		EnforceSSL:        true,  // Force HTTPS
		MinimumTLSVersion: 1.3,   // Highest TLS version

		// Data Protection & Compliance
		Versioned:                      true,
		ObjectLockEnabled:              true,
		ObjectLockDefaultRetentionMode: "COMPLIANCE", // Cannot be bypassed
		ObjectLockDefaultRetentionDays: 2555,         // 7 years retention

		// Cost Optimization
		EnableIntelligentTiering: true,
		TransitionMinimumSize:    "ALL_STORAGE_CLASSES_128_K",

		// Comprehensive Monitoring & Auditing
		EnableAccessLogs:   true,
		EnableInventory:    true,
		EnableMetrics:      true,
		EventBridgeEnabled: true, // For compliance automation

		// Performance - Not needed for enterprise data
		TransferAcceleration: false,
		EnableCORS:           false,

		// Lifecycle Management for Compliance
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id: jsii.String("ComplianceArchival"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)), // 1 year to Glacier
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(1095)), // 3 years to Deep Archive
					},
				},
			},
		},
	}
}

// GetStaticWebsiteProperties returns a configuration for a public static website hosting bucket
// NOTE: For production websites, use CloudFront + S3 with OAC instead of direct S3 website hosting
/*
func GetStaticWebsiteProperties() S3Properties {
	return S3Properties{
		// Basic Configuration
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "destroy",
		AutoDeleteObjects: true,

		// Website Hosting Configuration
		WebsiteEnabled:       true,
		WebsiteIndexDocument: "index.html",
		WebsiteErrorDocument: "404.html",

		// Security - Minimal for website hosting
		PublicAccess:     true,         // Required for website hosting
		Encryption:       "S3_MANAGED", // Basic encryption
		EnforceSSL:       false,        // Cannot enforce SSL with website hosting
		BucketKeyEnabled: true,

		// Data Protection
		Versioned:         true,  // Keep for deployment rollbacks
		ObjectLockEnabled: false, // Not needed for websites

		// Cost Optimization
		EnableIntelligentTiering: false, // Not cost-effective for websites

		// Lifecycle for cleanup
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id:                          jsii.String("WebsiteCleanup"),
				NoncurrentVersionExpiration: awscdk.Duration_Minutes(jsii.Number(30)),
			},
		},

		// Performance & CORS
		TransferAcceleration: false, // Use CloudFront instead
		EnableCORS:           true,
		CORSAllowedOrigins:   []string{"*"},           // Allow all origins for website
		CORSAllowedMethods:   []string{"GET", "HEAD"}, // Only read operations
		CORSAllowedHeaders:   []string{"*"},

		// Monitoring
		EnableAccessLogs:   false, // CloudFront logs are better
		EventBridgeEnabled: false,
		EnableInventory:    false,
		EnableMetrics:      false,
	}
}
*/

// GetCloudFrontOriginProperties returns optimized configuration for S3 bucket as CloudFront origin
// This is the RECOMMENDED approach for static websites instead of S3 website hosting
func GetCloudFrontOriginProperties() S3Properties {
	return S3Properties{
		// Basic Configuration
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "destroy",
		AutoDeleteObjects: true,

		// Security - Keep bucket private, access via CloudFront only
		PublicAccess:      false, // CloudFront uses OAC for access
		Encryption:        "S3_MANAGED",
		BucketKeyEnabled:  true,
		EnforceSSL:        true, // CloudFront handles SSL termination
		MinimumTLSVersion: 1.2,

		// Data Protection
		Versioned:         true, // For deployment rollbacks
		ObjectLockEnabled: false,

		// Performance Optimization
		TransferAcceleration: false, // CloudFront provides acceleration
		EnableCORS:           false, // CloudFront handles CORS

		// Cost Optimization
		EnableIntelligentTiering: false, // Not cost-effective for frequently accessed content

		// Lifecycle Management
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id:                          jsii.String("WebContentCleanup"),
				NoncurrentVersionExpiration: awscdk.Duration_Minutes(jsii.Number(30)),
			},
		},

		// Monitoring
		EnableAccessLogs:   false, // CloudFront access logs are sufficient
		EventBridgeEnabled: true,  // For automated deployments
		EnableInventory:    false,
		EnableMetrics:      false,

		// Website Hosting - Disabled (CloudFront handles this)
		WebsiteEnabled: false,
	}
}

// GetDataLakeProperties returns a configuration for a data lake analytics bucket
// Optimized for: Big data analytics, data science, batch processing workloads
func GetDataLakeProperties() S3Properties {
	return S3Properties{
		// Basic Configuration
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "retain",
		AutoDeleteObjects: false,

		// Security
		PublicAccess:      false,
		Encryption:        "KMS", // Better for analytics compliance
		BucketKeyEnabled:  true,
		EnforceSSL:        true,
		MinimumTLSVersion: 1.2,

		// Data Protection
		Versioned:         true,
		ObjectLockEnabled: false, // Usually not needed for analytics data

		// Cost Optimization - Critical for data lakes
		EnableIntelligentTiering: true,
		TransitionMinimumSize:    "ALL_STORAGE_CLASSES_128_K",

		// Comprehensive Lifecycle Management
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id:     jsii.String("DataLakeLifecycle"),
				Prefix: jsii.String("raw-data/"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)),
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)),
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),
					},
				},
			},
			{
				Id:     jsii.String("ProcessedDataLifecycle"),
				Prefix: jsii.String("processed-data/"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(7)), // Faster transition for processed data
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)),
					},
				},
			},
		},

		// Enhanced Monitoring & Analytics
		EnableAccessLogs:   true,
		AccessLogsPrefix:   "access-logs/",
		EnableInventory:    true,
		EnableMetrics:      true,
		MetricsPrefix:      "analytics/",
		EventBridgeEnabled: true, // For data pipeline automation

		// Performance
		TransferAcceleration: false, // Usually not needed for batch processing
		EnableCORS:           false,
	}
}

// GetBackupProperties returns a configuration for a backup and disaster recovery bucket
// Optimized for: Database backups, application backups, disaster recovery
func GetBackupProperties() S3Properties {
	return S3Properties{
		// Basic Configuration
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "retain",
		AutoDeleteObjects: false,

		// Enhanced Security for Backups
		PublicAccess:      false,
		Encryption:        "KMS", // Enhanced security for backups
		BucketKeyEnabled:  true,
		EnforceSSL:        true,
		MinimumTLSVersion: 1.2,

		// Data Protection & Compliance
		Versioned:                      true,
		ObjectLockEnabled:              true,
		ObjectLockDefaultRetentionMode: "GOVERNANCE", // Allows administrative overrides
		ObjectLockDefaultRetentionDays: 90,           // 3 months minimum retention

		// Aggressive Cost Optimization for Backups
		EnableIntelligentTiering: true,

		// Lifecycle Management for Backup Retention
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id: jsii.String("BackupRetention"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)), // Move to IA after 1 month
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)), // Archive after 3 months
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)), // Deep archive after 1 year
					},
				},
				// Backup retention policy
				Expiration: awscdk.Duration_Days(jsii.Number(2555)),
			},
		},

		// Cross-Region Replication for DR
		ReplicationEnabled:      false, // Enable manually with specific destination
		ReplicationDestination:  "",    // Set to target region bucket ARN
		ReplicationStorageClass: "GLACIER",

		// Comprehensive Monitoring
		EnableAccessLogs:   true,
		EnableInventory:    true,
		EnableMetrics:      true,
		EventBridgeEnabled: true, // For backup automation workflows

		// Performance
		TransferAcceleration: false, // Not typically needed for backups
		EnableCORS:           false,
	}
}

// GetMediaStreamingProperties returns a configuration for a media streaming application bucket
// Optimized for: Video/audio streaming, CDN origin, high-throughput content delivery
func GetMediaStreamingProperties() S3Properties {
	return S3Properties{
		// Basic Configuration
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "retain",
		AutoDeleteObjects: false,

		// Security - Balanced for content delivery
		PublicAccess:      false,        // Use CloudFront with OAC instead
		Encryption:        "S3_MANAGED", // KMS adds latency for streaming
		BucketKeyEnabled:  true,
		EnforceSSL:        true,
		MinimumTLSVersion: 1.2,

		// Data Protection
		Versioned:         false, // Media files are typically immutable
		ObjectLockEnabled: false,

		// Cost Optimization for Media
		EnableIntelligentTiering: true,

		// Lifecycle Management for Media Content
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id:     jsii.String("MediaContentLifecycle"),
				Prefix: jsii.String("videos/"),
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)), // Popular content stays hot longer
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)), // Archive old content
					},
				},
			},
		},

		// Performance Optimization for Streaming
		TransferAcceleration: false, // Use CloudFront instead
		EnableCORS:           true,
		CORSAllowedOrigins: []string{
			"https://player.example.com", // Specific player domains
			"https://*.cdn.example.com",  // CDN subdomains
		},
		CORSAllowedMethods: []string{"GET", "HEAD"}, // Only read operations for streaming
		CORSAllowedHeaders: []string{"Range", "Authorization", "Content-Type"},

		// Monitoring for Performance
		EnableAccessLogs:   false, // Use CloudFront logs instead
		EnableInventory:    false,
		EnableMetrics:      true,
		MetricsPrefix:      "videos/", // Monitor video performance
		EventBridgeEnabled: true,      // For content processing workflows

		// Website Hosting - Disabled (use CloudFront)
		WebsiteEnabled: false,
	}
}

// GetDevelopmentProperties returns a configuration for development/testing environments
// Optimized for: Cost efficiency, easy cleanup, development workflows
func GetDevelopmentProperties() S3Properties {
	return S3Properties{
		// Basic Configuration - Easy cleanup
		BucketName:        "", // Caller must provide a unique name
		RemovalPolicy:     "destroy",
		AutoDeleteObjects: true,

		// Minimal Security for Development
		PublicAccess:      false,
		Encryption:        "S3_MANAGED",
		BucketKeyEnabled:  false, // Reduce complexity
		EnforceSSL:        false, // Reduce complexity for dev
		MinimumTLSVersion: 1.2,

		// Basic Data Protection
		Versioned:         false, // Reduce costs in dev
		ObjectLockEnabled: false,

		// Minimal Cost Optimization
		EnableIntelligentTiering: false, // Not cost-effective for short-lived dev data

		// Simple Lifecycle for Cost Control
		LifecycleRules: []*awss3.LifecycleRule{
			{
				Id:         jsii.String("DevCleanup"),
				Expiration: awscdk.Duration_Days(jsii.Number(30)), // Auto-cleanup after 30 days
			},
		},

		// Minimal Monitoring
		EnableAccessLogs:   false,
		EnableInventory:    false,
		EnableMetrics:      false,
		EventBridgeEnabled: false,

		// Development Features
		TransferAcceleration: false,
		EnableCORS:           true,          // Often needed for web development
		CORSAllowedOrigins:   []string{"*"}, // Permissive for development
		CORSAllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		CORSAllowedHeaders:   []string{"*"},

		// Website Hosting - Enabled for dev testing
		WebsiteEnabled: false, // Enable only if needed
	}
}
