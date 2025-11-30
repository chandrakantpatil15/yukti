variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
  
  validation {
    condition = contains([
      "us-east-1", "us-west-1", "us-west-2", 
      "eu-west-1", "eu-central-1", "ap-southeast-1"
    ], var.aws_region)
    error_message = "Region must be a valid AWS region."
  }
}

variable "environment" {
  description = "Environment name for resource tagging"
  type        = string
  default     = "yukti-test"
  
  validation {
    condition     = length(var.environment) > 0 && length(var.environment) <= 20
    error_message = "Environment name must be between 1 and 20 characters."
  }
}

variable "instance_count" {
  description = "Number of expensive EC2 instances to create"
  type        = number
  default     = 3
  
  validation {
    condition     = var.instance_count >= 1 && var.instance_count <= 5
    error_message = "Instance count must be between 1 and 5."
  }
}

variable "rds_count" {
  description = "Number of expensive RDS instances to create"
  type        = number
  default     = 2
  
  validation {
    condition     = var.rds_count >= 1 && var.rds_count <= 3
    error_message = "RDS count must be between 1 and 3."
  }
}

variable "enable_multi_az" {
  description = "Enable Multi-AZ for RDS instances (doubles cost)"
  type        = bool
  default     = true
}

variable "enable_detailed_monitoring" {
  description = "Enable detailed monitoring for EC2 instances"
  type        = bool
  default     = true
}

variable "cost_center" {
  description = "Cost center for billing allocation"
  type        = string
  default     = "finops-testing"
}

variable "owner_email" {
  description = "Owner email for resource tagging"
  type        = string
  default     = "admin@company.com"
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.owner_email))
    error_message = "Owner email must be a valid email address."
  }
}