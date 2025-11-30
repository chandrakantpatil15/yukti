package remediation

import (
	"fmt"
	"strings"
	"time"
)

type TerragruntGenerator struct {
	customerID string
}

func NewTerragruntGenerator(customerID string) *TerragruntGenerator {
	return &TerragruntGenerator{customerID: customerID}
}

func (t *TerragruntGenerator) GenerateRightSizingTerragrunt(instanceID, currentType, recommendedType string, monthlySavings float64) (string, string) {
	// Generate terragrunt.hcl
	terragruntHCL := t.generateTerragruntConfig(instanceID, "rightsizing")
	
	// Generate main.tf
	mainTF := t.generateRightSizingTerraform(instanceID, currentType, recommendedType, monthlySavings)
	
	return terragruntHCL, mainTF
}

func (t *TerragruntGenerator) generateTerragruntConfig(resourceID, actionType string) string {
	template := `# Yukti FinOps - Terragrunt Configuration
# Customer: %s
# Resource: %s
# Action: %s
# Generated: %s

terraform {
  source = "."
}

# Remote state configuration
remote_state {
  backend = "s3"
  config = {
    bucket         = "yukti-terraform-state-${local.customer_id}"
    key            = "finops/${local.environment}/${local.resource_id}/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "yukti-terraform-locks"
  }
}

# Local variables
locals {
  customer_id  = "%s"
  resource_id  = "%s"
  action_type  = "%s"
  environment  = get_env("ENVIRONMENT", "dev")
  
  # Common tags
  common_tags = {
    Customer     = local.customer_id
    Environment  = local.environment
    ManagedBy    = "Yukti-FinOps"
    ActionType   = local.action_type
    ResourceId   = local.resource_id
    CreatedDate  = "%s"
  }
}

# Input variables
inputs = {
  customer_id    = local.customer_id
  resource_id    = local.resource_id
  environment    = local.environment
  common_tags    = local.common_tags
  
  # Action-specific inputs
  instance_id    = local.resource_id
}

# Dependencies (if any)
dependencies {
  paths = []
}

# Hooks for validation and safety
terraform {
  before_hook "validate_environment" {
    commands = ["plan", "apply"]
    execute  = ["echo", "Validating environment: ${local.environment}"]
  }
  
  before_hook "cost_estimation" {
    commands = ["plan"]
    execute  = ["echo", "Estimated monthly savings: $%.2f"]
  }
  
  after_hook "notify_completion" {
    commands = ["apply"]
    execute  = ["echo", "Yukti FinOps optimization completed for ${local.resource_id}"]
  }
}

# Prevent accidental destruction
prevent_destroy = true
`

	return fmt.Sprintf(template,
		t.customerID,
		resourceID,
		actionType,
		time.Now().Format("2006-01-02 15:04:05"),
		t.customerID,
		resourceID,
		actionType,
		time.Now().Format("2006-01-02"),
		monthlySavings,
	)
}

func (t *TerragruntGenerator) generateRightSizingTerraform(instanceID, currentType, recommendedType string, monthlySavings float64) string {
	template := `# Yukti FinOps - EC2 Right-sizing Terraform Module
# This file works with terragrunt.hcl for enterprise deployment

variable "customer_id" {
  description = "Customer identifier"
  type        = string
}

variable "resource_id" {
  description = "Resource identifier (instance ID)"
  type        = string
}

variable "environment" {
  description = "Environment (dev/staging/prod)"
  type        = string
  default     = "dev"
}

variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

variable "instance_id" {
  description = "EC2 instance ID to optimize"
  type        = string
}

# Data source for current instance
data "aws_instance" "current" {
  instance_id = var.instance_id
}

# Create optimized instance
resource "aws_instance" "optimized" {
  instance_type = "%s"
  
  # Copy configuration from existing instance
  ami                    = data.aws_instance.current.ami
  key_name              = data.aws_instance.current.key_name
  vpc_security_group_ids = data.aws_instance.current.vpc_security_group_ids
  subnet_id             = data.aws_instance.current.subnet_id
  
  # Enhanced tagging with customer context
  tags = merge(
    var.common_tags,
    data.aws_instance.current.tags,
    {
      Name                = "${data.aws_instance.current.tags.Name}-optimized"
      OriginalInstanceId  = var.instance_id
      OriginalType        = "%s"
      OptimizedType       = "%s"
      EstimatedSavings    = "%.2f"
      OptimizationDate    = "%s"
    }
  )
  
  lifecycle {
    create_before_destroy = true
    
    # Prevent accidental changes to critical attributes
    ignore_changes = [
      ami,
      key_name,
    ]
  }
}

# Create backup of original instance (optional)
resource "aws_ami" "backup" {
  count       = var.environment == "prod" ? 1 : 0
  name        = "yukti-backup-${var.instance_id}-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
  source_instance_id = var.instance_id
  
  tags = merge(
    var.common_tags,
    {
      Name        = "Yukti Backup - ${var.instance_id}"
      Purpose     = "Pre-optimization backup"
      SourceInstance = var.instance_id
    }
  )
}

# Outputs
output "optimized_instance_id" {
  description = "ID of the optimized instance"
  value       = aws_instance.optimized.id
}

output "optimized_instance_type" {
  description = "Instance type of optimized instance"
  value       = aws_instance.optimized.instance_type
}

output "original_instance_id" {
  description = "ID of the original instance"
  value       = var.instance_id
}

output "estimated_monthly_savings" {
  description = "Estimated monthly cost savings"
  value       = "%.2f"
}

output "backup_ami_id" {
  description = "AMI ID of backup (if created)"
  value       = var.environment == "prod" ? aws_ami.backup[0].id : null
}

output "optimization_summary" {
  description = "Summary of the optimization"
  value = {
    customer_id           = var.customer_id
    original_instance_id  = var.instance_id
    optimized_instance_id = aws_instance.optimized.id
    original_type         = "%s"
    optimized_type        = "%s"
    estimated_savings     = "%.2f"
    environment          = var.environment
    optimization_date    = "%s"
  }
}
`

	return fmt.Sprintf(template,
		recommendedType,
		currentType,
		recommendedType,
		monthlySavings,
		time.Now().Format("2006-01-02"),
		monthlySavings,
		currentType,
		recommendedType,
		monthlySavings,
		time.Now().Format("2006-01-02"),
	)
}

func (t *TerragruntGenerator) GenerateEnvironmentStructure(customerID string) map[string]string {
	files := make(map[string]string)
	
	// Root terragrunt.hcl
	files["terragrunt.hcl"] = t.generateRootTerragrunt(customerID)
	
	// Environment-specific configurations
	environments := []string{"dev", "staging", "prod"}
	
	for _, env := range environments {
		files[fmt.Sprintf("environments/%s/terragrunt.hcl", env)] = t.generateEnvironmentConfig(customerID, env)
	}
	
	// Common variables
	files["common/variables.hcl"] = t.generateCommonVariables(customerID)
	
	return files
}

func (t *TerragruntGenerator) generateRootTerragrunt(customerID string) string {
	return fmt.Sprintf(`# Yukti FinOps - Root Terragrunt Configuration
# Customer: %s

# Configure Terragrunt to automatically store tfstate files in S3
remote_state {
  backend = "s3"
  config = {
    bucket         = "yukti-terraform-state-%s"
    key            = "${path_relative_to_include()}/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "yukti-terraform-locks-%s"
  }
}

# Configure root level variables that all resources can inherit
locals {
  customer_id = "%s"
  
  # Common tags applied to all resources
  common_tags = {
    Customer    = local.customer_id
    ManagedBy   = "Yukti-FinOps"
    Terraform   = "true"
    Repository  = "yukti-finops"
  }
}
`, customerID, customerID, customerID, customerID)
}

func (t *TerragruntGenerator) generateEnvironmentConfig(customerID, environment string) string {
	return fmt.Sprintf(`# Yukti FinOps - %s Environment Configuration
# Customer: %s

# Include root configuration
include "root" {
  path = find_in_parent_folders()
}

# Environment-specific locals
locals {
  environment = "%s"
  
  # Environment-specific tags
  env_tags = {
    Environment = local.environment
  }
}

# Environment-specific inputs
inputs = {
  environment = local.environment
  
  # Environment-specific settings
  instance_monitoring = %s
  backup_retention    = %d
  
  # Merge common and environment tags
  tags = merge(
    local.common_tags,
    local.env_tags
  )
}
`, strings.Title(environment), customerID, environment, 
   map[string]string{"prod": "true", "staging": "true", "dev": "false"}[environment],
   map[string]int{"prod": 30, "staging": 7, "dev": 1}[environment])
}

func (t *TerragruntGenerator) generateCommonVariables(customerID string) string {
	return fmt.Sprintf(`# Yukti FinOps - Common Variables
# Customer: %s

locals {
  # AWS regions for multi-region deployment
  aws_regions = {
    primary   = "us-east-1"
    secondary = "us-west-2"
  }
  
  # Instance type mappings for optimization
  instance_type_mappings = {
    "t3.micro"  = "t3.nano"
    "t3.small"  = "t3.micro"
    "t3.medium" = "t3.small"
    "m5.large"  = "m5.medium"
    "m5.xlarge" = "m5.large"
  }
  
  # Cost optimization thresholds
  optimization_thresholds = {
    cpu_utilization_max    = 80
    cpu_utilization_min    = 20
    memory_utilization_max = 85
    memory_utilization_min = 30
  }
  
  # Backup and retention policies
  backup_policies = {
    dev = {
      retention_days = 1
      frequency     = "daily"
    }
    staging = {
      retention_days = 7
      frequency     = "daily"
    }
    prod = {
      retention_days = 30
      frequency     = "daily"
    }
  }
}
`, customerID)
}