output "deployment_summary" {
  description = "Summary of deployed resources"
  value = {
    region      = var.aws_region
    environment = var.environment
    timestamp   = timestamp()
  }
}

output "estimated_monthly_costs" {
  description = "Estimated monthly costs for all resources"
  value = {
    ec2_instances = {
      count = var.instance_count
      cost  = "$${var.instance_count * 560} (${var.instance_count}x m5.4xlarge)"
    }
    rds_instances = {
      count = var.rds_count
      cost  = var.enable_multi_az ? "$${var.rds_count * 735} (${var.rds_count}x db.r5.2xlarge Multi-AZ)" : "$${var.rds_count * 368} (${var.rds_count}x db.r5.2xlarge Single-AZ)"
    }
    networking = {
      nat_gateways     = "~$90 (2x NAT Gateway)"
      load_balancers   = "~$36 (2x ALB)"
    }
    storage = {
      ebs_volumes = "~$200 (High IOPS volumes)"
      s3_buckets  = "~$50 (2x S3 buckets with versioning)"
    }
    total_estimated = var.enable_multi_az ? 
      "$${var.instance_count * 560 + var.rds_count * 735 + 376}/month" : 
      "$${var.instance_count * 560 + var.rds_count * 368 + 376}/month"
  }
}

output "resource_details" {
  description = "Details of created resources for Yukti platform"
  value = {
    ec2_instances = {
      arns          = aws_instance.expensive_compute[*].arn
      instance_ids  = aws_instance.expensive_compute[*].id
      instance_type = "m5.4xlarge"
      count         = length(aws_instance.expensive_compute)
    }
    rds_instances = {
      arns        = aws_db_instance.expensive_database[*].arn
      identifiers = aws_db_instance.expensive_database[*].identifier
      engine      = "postgres"
      multi_az    = var.enable_multi_az
      count       = length(aws_db_instance.expensive_database)
    }
    s3_buckets = {
      arns   = aws_s3_bucket.expensive_storage[*].arn
      names  = aws_s3_bucket.expensive_storage[*].id
      count  = length(aws_s3_bucket.expensive_storage)
    }
    load_balancers = {
      arns  = aws_lb.expensive_alb[*].arn
      names = aws_lb.expensive_alb[*].name
      count = length(aws_lb.expensive_alb)
    }
    nat_gateways = {
      ids   = aws_nat_gateway.expensive_nat[*].id
      count = length(aws_nat_gateway.expensive_nat)
    }
  }
}

output "yukti_integration_info" {
  description = "Information needed for Yukti FinOps platform integration"
  value = {
    aws_account_id = data.aws_caller_identity.current.account_id
    region         = var.aws_region
    resource_count = {
      ec2 = length(aws_instance.expensive_compute)
      rds = length(aws_db_instance.expensive_database)
      s3  = length(aws_s3_bucket.expensive_storage)
      alb = length(aws_lb.expensive_alb)
      nat = length(aws_nat_gateway.expensive_nat)
    }
    tags_for_filtering = {
      Environment = var.environment
      CostCenter  = var.cost_center
      Purpose     = "cost-optimization-testing"
    }
  }
}

output "cost_optimization_opportunities" {
  description = "Expected cost optimization findings from Yukti"
  value = {
    ec2_rightsizing = {
      current_cost     = "$${var.instance_count * 560}/month"
      optimized_cost   = "$${var.instance_count * 280}/month (m5.2xlarge)"
      potential_saving = "$${var.instance_count * 280}/month (50% reduction)"
    }
    rds_optimization = var.enable_multi_az ? {
      multi_az_review = {
        current_cost     = "$${var.rds_count * 735}/month"
        single_az_cost   = "$${var.rds_count * 368}/month"
        potential_saving = "$${var.rds_count * 367}/month (50% reduction)"
      }
    } : {}
    reserved_instances = {
      ec2_ri_savings = "$${var.instance_count * 168}/month (30% with 1-year RI)"
      rds_ri_savings = var.enable_multi_az ? "$${var.rds_count * 220}/month (30% with 1-year RI)" : "$${var.rds_count * 110}/month"
    }
    storage_optimization = {
      ebs_type_change    = "~$120/month (io2 → gp3)"
      s3_lifecycle       = "~$30/month (intelligent tiering)"
    }
    networking = {
      nat_instance       = "~$72/month (NAT Gateway → NAT Instance)"
      alb_consolidation  = "~$18/month (2 ALBs → 1 ALB)"
    }
  }
}

# Data source to get current AWS account ID
data "aws_caller_identity" "current" {}