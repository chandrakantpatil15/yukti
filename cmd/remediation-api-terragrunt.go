package main

import (
	"fmt"
	"strings"
	"time"
)

func (api *RemediationAPI) generateTerragruntRightSizing(req RemediationRequest) (string, string) {
	currentType := req.Parameters["current_type"]
	recommendedType := req.Parameters["recommended_type"]
	if currentType == "" {
		currentType = "m5.large"
	}
	if recommendedType == "" {
		recommendedType = "m5.medium"
	}

	// Generate terragrunt.hcl
	terragruntHCL := fmt.Sprintf(`# Yukti FinOps - Terragrunt Configuration
# Customer: %s
# Resource: %s
# Environment: %s

terraform {
  source = "."
}

include "root" {
  path = find_in_parent_folders()
}

locals {
  customer_id = "%s"
  resource_id = "%s"
  environment = "%s"
  action_type = "rightsizing"
}

inputs = {
  customer_id    = local.customer_id
  resource_id    = local.resource_id
  environment    = local.environment
  instance_id    = local.resource_id
  
  current_type      = "%s"
  recommended_type  = "%s"
  estimated_savings = 120.50
  
  tags = {
    Customer      = local.customer_id
    Environment   = local.environment
    ManagedBy     = "Yukti-FinOps"
    ActionType    = local.action_type
    OptimizedDate = "%s"
  }
}

# Environment-specific settings
terraform {
  before_hook "validate" {
    commands = ["plan", "apply"]
    execute  = ["echo", "Validating %s environment optimization"]
  }
  
  after_hook "notify" {
    commands = ["apply"]
    execute  = ["echo", "Optimization completed for %s in %s"]
  }
}
`, req.CustomerID, req.ResourceID, req.Environment, req.CustomerID, req.ResourceID, req.Environment,
		currentType, recommendedType, time.Now().Format("2006-01-02"), req.Environment, req.ResourceID, req.Environment)

	// Generate main.tf
	mainTF := fmt.Sprintf(`variable "customer_id" {
  description = "Customer identifier"
  type        = string
}

variable "resource_id" {
  description = "Resource identifier"
  type        = string
}

variable "environment" {
  description = "Environment"
  type        = string
}

variable "instance_id" {
  description = "Instance ID to optimize"
  type        = string
}

variable "current_type" {
  description = "Current instance type"
  type        = string
}

variable "recommended_type" {
  description = "Recommended instance type"
  type        = string
}

variable "estimated_savings" {
  description = "Estimated monthly savings"
  type        = number
}

variable "tags" {
  description = "Resource tags"
  type        = map(string)
  default     = {}
}

data "aws_instance" "current" {
  instance_id = var.instance_id
}

resource "aws_instance" "optimized" {
  instance_type = var.recommended_type
  
  ami                    = data.aws_instance.current.ami
  key_name              = data.aws_instance.current.key_name
  vpc_security_group_ids = data.aws_instance.current.vpc_security_group_ids
  subnet_id             = data.aws_instance.current.subnet_id
  
  tags = merge(
    var.tags,
    data.aws_instance.current.tags,
    {
      Name             = "$${data.aws_instance.current.tags.Name}-optimized"
      OriginalType     = var.current_type
      OptimizedType    = var.recommended_type
      EstimatedSavings = "$${var.estimated_savings}"
    }
  )
  
  lifecycle {
    create_before_destroy = true
  }
}

output "optimized_instance_id" {
  value = aws_instance.optimized.id
}

output "estimated_monthly_savings" {
  value = var.estimated_savings
}
`)

	return terragruntHCL, mainTF
}

func (api *RemediationAPI) generateTerragruntStructure(req RemediationRequest) map[string]string {
	files := make(map[string]string)

	// Root terragrunt.hcl
	files["terragrunt.hcl"] = fmt.Sprintf(`remote_state {
  backend = "s3"
  config = {
    bucket         = "yukti-terraform-state-%s"
    key            = "$${path_relative_to_include()}/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "yukti-terraform-locks-%s"
  }
}

locals {
  customer_id = "%s"
  
  common_tags = {
    Customer    = local.customer_id
    ManagedBy   = "Yukti-FinOps"
    Terraform   = "true"
  }
}
`, req.CustomerID, req.CustomerID, req.CustomerID)

	// Environment-specific config
	files[fmt.Sprintf("environments/%s/terragrunt.hcl", req.Environment)] = fmt.Sprintf(`include "root" {
  path = find_in_parent_folders()
}

locals {
  environment = "%s"
}

inputs = {
  environment = local.environment
  
  # Environment-specific settings
  backup_enabled = %s
  monitoring_enabled = %s
  
  tags = merge(
    local.common_tags,
    {
      Environment = local.environment
    }
  )
}
`, req.Environment,
		map[string]string{"prod": "true", "staging": "true", "dev": "false"}[req.Environment],
		map[string]string{"prod": "true", "staging": "true", "dev": "false"}[req.Environment])

	return files
}

func (api *RemediationAPI) getTerragruntInstructions(templateFormat, environment string) string {
	if templateFormat != "terragrunt" {
		return "1. Review the configuration\n2. Test in non-production first\n3. Apply during maintenance window"
	}

	instructions := fmt.Sprintf(`# Terragrunt Deployment Instructions

## Prerequisites
1. Install Terragrunt: https://terragrunt.gruntwork.io/docs/getting-started/install/
2. Configure AWS credentials
3. Ensure S3 bucket exists: yukti-terraform-state-<customer-id>
4. Ensure DynamoDB table exists: yukti-terraform-locks-<customer-id>

## Deployment Steps

### 1. Initialize Terragrunt
`)

	if environment == "prod" {
		instructions += `
# PRODUCTION DEPLOYMENT - Extra Safety Steps
terragrunt plan --terragrunt-working-dir environments/prod
# Review the plan carefully
# Get approval from team lead
terragrunt apply --terragrunt-working-dir environments/prod
`
	} else {
		instructions += fmt.Sprintf(`
terragrunt plan --terragrunt-working-dir environments/%s
terragrunt apply --terragrunt-working-dir environments/%s
`, environment, environment)
	}

	instructions += `
### 2. Verify Deployment
terragrunt output --terragrunt-working-dir environments/` + environment + `

### 3. Monitor Results
- Check AWS Console for new instance
- Verify cost savings in CloudWatch
- Monitor application performance

## Rollback (if needed)
terragrunt destroy --terragrunt-working-dir environments/` + environment + `

## Multi-Environment Promotion
# Deploy to dev first
terragrunt apply --terragrunt-working-dir environments/dev

# Then staging
terragrunt apply --terragrunt-working-dir environments/staging

# Finally production (with approval)
terragrunt apply --terragrunt-working-dir environments/prod
`

	return instructions
}