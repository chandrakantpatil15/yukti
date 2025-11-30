package aws

// AWSServiceRegistry contains all AWS services organized by category
type AWSServiceRegistry struct {
	Services map[string][]string `json:"services"`
	Total    int                 `json:"total"`
}

// NewAWSServiceRegistry creates a comprehensive registry of all AWS services
func NewAWSServiceRegistry() *AWSServiceRegistry {
	services := map[string][]string{
		"compute": {
			"ec2", "lambda", "ecs", "eks", "fargate", "batch", "lightsail",
			"elastic-beanstalk", "app-runner", "serverless-application-repository",
			"outposts", "wavelength", "local-zones", "nitro-enclaves",
		},
		"storage": {
			"s3", "ebs", "efs", "fsx", "s3-glacier", "storage-gateway",
			"backup", "datasync", "transfer-family", "snow-family",
		},
		"database": {
			"rds", "dynamodb", "redshift", "elasticache", "neptune", "timestream",
			"documentdb", "keyspaces", "qldb", "memorydb", "rds-proxy",
		},
		"networking": {
			"vpc", "cloudfront", "route53", "api-gateway", "direct-connect",
			"elb", "global-accelerator", "transit-gateway", "privatelink",
			"client-vpn", "site-to-site-vpn", "cloud-wan", "verified-access",
		},
		"security": {
			"iam", "cognito", "secrets-manager", "kms", "acm", "waf", "shield",
			"inspector", "guardduty", "macie", "security-hub", "detective",
			"access-analyzer", "cloudhsm", "certificate-manager", "artifact",
			"audit-manager", "network-firewall", "firewall-manager",
		},
		"analytics": {
			"emr", "kinesis", "glue", "athena", "quicksight", "opensearch",
			"msk", "data-pipeline", "lake-formation", "kinesis-analytics",
			"redshift-spectrum", "clean-rooms", "finspace", "healthlake",
		},
		"machine-learning": {
			"sagemaker", "comprehend", "lex", "polly", "rekognition", "translate",
			"transcribe", "textract", "personalize", "forecast", "fraud-detector",
			"kendra", "augmented-ai", "codewhisperer", "bedrock", "lookout",
		},
		"developer-tools": {
			"codecommit", "codebuild", "codedeploy", "codepipeline", "codestar",
			"cloud9", "x-ray", "codeartifact", "codeguru", "fault-injection-simulator",
			"application-composer", "codecatalyst", "migration-hub-refactor-spaces",
		},
		"management": {
			"cloudwatch", "cloudformation", "cloudtrail", "config", "systems-manager",
			"trusted-advisor", "personal-health-dashboard", "service-catalog",
			"well-architected-tool", "control-tower", "organizations", "resource-groups",
			"tag-editor", "resource-access-manager", "license-manager", "service-quotas",
			"compute-optimizer", "chatbot", "launch-wizard", "resilience-hub",
			"application-insights", "ops-center", "incident-manager", "proton",
		},
		"integration": {
			"sns", "sqs", "eventbridge", "step-functions", "swf", "mq", "managed-workflows",
			"appflow", "api-gateway", "app-mesh", "cloud-map", "app-sync",
		},
		"containers": {
			"ecs", "eks", "fargate", "ecr", "app-runner", "copilot", "red-hat-openshift",
		},
		"migration": {
			"migration-hub", "application-migration-service", "database-migration-service",
			"datasync", "transfer-family", "snow-family", "application-discovery-service",
			"migration-evaluator", "mainframe-modernization",
		},
		"media": {
			"elemental-mediaconnect", "elemental-medialive", "elemental-mediapackage",
			"elemental-mediastore", "elemental-mediatailor", "kinesis-video-streams",
			"interactive-video-service", "nimble-studio", "elastic-transcoder",
		},
		"business-applications": {
			"workspaces", "appstream", "workdocs", "workmail", "chime", "connect",
			"pinpoint", "simple-email-service", "worklink", "alexa-for-business",
			"honeycode", "wickr", "supply-chain", "verified-permissions",
		},
		"end-user-computing": {
			"workspaces", "appstream", "workspaces-web", "workspaces-core",
		},
		"iot": {
			"iot-core", "iot-device-management", "iot-device-defender", "iot-analytics",
			"iot-events", "iot-greengrass", "iot-sitewise", "iot-things-graph",
			"iot-1-click", "iot-button", "freertos", "iot-roborunner", "iot-twinmaker",
			"iot-fleetwise", "iot-expresslink",
		},
		"robotics": {
			"robomaker", "iot-roborunner",
		},
		"blockchain": {
			"managed-blockchain", "quantum-ledger-database",
		},
		"satellite": {
			"ground-station",
		},
		"quantum": {
			"braket",
		},
		"ar-vr": {
			"sumerian",
		},
		"game-development": {
			"gamelift", "lumberyard",
		},
		"cost-management": {
			"cost-explorer", "budgets", "cost-and-usage-report", "savings-plans",
			"reserved-instances", "billing-conductor", "application-cost-profiler",
		},
		"front-end-web-mobile": {
			"amplify", "device-farm", "location-service", "pinpoint",
		},
		"customer-engagement": {
			"connect", "pinpoint", "simple-email-service", "workdocs",
		},
	}

	total := 0
	for _, serviceList := range services {
		total += len(serviceList)
	}

	return &AWSServiceRegistry{
		Services: services,
		Total:    total,
	}
}

// GetServicesByCategory returns services for a specific category
func (r *AWSServiceRegistry) GetServicesByCategory(category string) []string {
	return r.Services[category]
}

// GetAllCategories returns all service categories
func (r *AWSServiceRegistry) GetAllCategories() []string {
	categories := make([]string, 0, len(r.Services))
	for category := range r.Services {
		categories = append(categories, category)
	}
	return categories
}

// IsServiceSupported checks if a service is in the registry
func (r *AWSServiceRegistry) IsServiceSupported(serviceName string) bool {
	for _, serviceList := range r.Services {
		for _, service := range serviceList {
			if service == serviceName {
				return true
			}
		}
	}
	return false
}

// GetServiceCategory returns the category for a given service
func (r *AWSServiceRegistry) GetServiceCategory(serviceName string) string {
	for category, serviceList := range r.Services {
		for _, service := range serviceList {
			if service == serviceName {
				return category
			}
		}
	}
	return "unknown"
}