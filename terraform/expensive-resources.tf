# Expensive AWS Resources for Yukti FinOps Testing
# This creates intentionally expensive resources to test cost optimization

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "yukti-test"
}

# 1. EXPENSIVE EC2 INSTANCES ($500-1000/month each)
resource "aws_instance" "expensive_compute" {
  count         = 3
  ami           = "ami-0c02fb55956c7d316" # Amazon Linux 2
  instance_type = "m5.4xlarge"           # $0.768/hour = $560/month
  
  # Expensive EBS volumes
  root_block_device {
    volume_type = "gp3"
    volume_size = 500  # 500GB
    iops        = 3000
    throughput  = 125
  }
  
  # Additional expensive volume
  ebs_block_device {
    device_name = "/dev/sdf"
    volume_type = "io2"
    volume_size = 1000  # 1TB
    iops        = 10000 # Very expensive IOPS
  }
  
  # Detailed monitoring (extra cost)
  monitoring = true
  
  tags = {
    Name        = "expensive-compute-${count.index + 1}"
    Environment = var.environment
    CostCenter  = "development"
    Owner       = "test-team"
    Purpose     = "cost-optimization-testing"
  }
}

# 2. EXPENSIVE RDS INSTANCES ($800-1500/month each)
resource "aws_db_instance" "expensive_database" {
  count = 2
  
  identifier = "expensive-db-${count.index + 1}"
  
  # Expensive instance class
  instance_class = "db.r5.2xlarge"  # $1.008/hour = $735/month
  engine         = "postgres"
  engine_version = "15.4"
  
  allocated_storage     = 1000  # 1TB
  max_allocated_storage = 2000  # Auto-scaling to 2TB
  storage_type          = "io1"
  iops                  = 5000  # Expensive provisioned IOPS
  
  # Multi-AZ for high availability (doubles cost)
  multi_az = true
  
  # Expensive backup retention
  backup_retention_period = 30
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  # Performance Insights (extra cost)
  performance_insights_enabled = true
  performance_insights_retention_period = 7
  
  db_name  = "expensivedb"
  username = "admin"
  password = "ChangeMe123!"
  
  skip_final_snapshot = true
  deletion_protection = false
  
  tags = {
    Name        = "expensive-database-${count.index + 1}"
    Environment = var.environment
    CostCenter  = "production"
    Owner       = "database-team"
  }
}

# 3. EXPENSIVE S3 BUCKETS with costly storage classes
resource "aws_s3_bucket" "expensive_storage" {
  count  = 2
  bucket = "yukti-expensive-storage-${count.index + 1}-${random_id.bucket_suffix.hex}"
  
  tags = {
    Name        = "expensive-storage-${count.index + 1}"
    Environment = var.environment
    CostCenter  = "data-analytics"
  }
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Expensive S3 storage configuration
resource "aws_s3_bucket_versioning" "expensive_versioning" {
  count  = length(aws_s3_bucket.expensive_storage)
  bucket = aws_s3_bucket.expensive_storage[count.index].id
  
  versioning_configuration {
    status = "Enabled"  # Keeps multiple versions (expensive)
  }
}

# 4. EXPENSIVE LOAD BALANCERS
resource "aws_lb" "expensive_alb" {
  count = 2
  
  name               = "expensive-alb-${count.index + 1}"
  internal           = false
  load_balancer_type = "application"
  subnets            = [aws_subnet.public_1.id, aws_subnet.public_2.id]
  
  # Cross-zone load balancing (extra cost)
  enable_cross_zone_load_balancing = true
  
  tags = {
    Name        = "expensive-alb-${count.index + 1}"
    Environment = var.environment
  }
}

# 5. EXPENSIVE NAT GATEWAYS ($45/month each + data transfer)
resource "aws_nat_gateway" "expensive_nat" {
  count = 2
  
  allocation_id = aws_eip.nat_eip[count.index].id
  subnet_id     = aws_subnet.public_1.id
  
  tags = {
    Name = "expensive-nat-${count.index + 1}"
  }
}

resource "aws_eip" "nat_eip" {
  count  = 2
  domain = "vpc"
  
  tags = {
    Name = "nat-eip-${count.index + 1}"
  }
}

# 6. VPC and Networking (required for resources)
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name = "yukti-test-vpc"
  }
}

resource "aws_subnet" "public_1" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.aws_region}a"
  map_public_ip_on_launch = true
  
  tags = {
    Name = "public-subnet-1"
  }
}

resource "aws_subnet" "public_2" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.2.0/24"
  availability_zone       = "${var.aws_region}b"
  map_public_ip_on_launch = true
  
  tags = {
    Name = "public-subnet-2"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  
  tags = {
    Name = "main-igw"
  }
}

# Output estimated monthly costs
output "estimated_monthly_costs" {
  value = {
    ec2_instances    = "~$1,680 (3x m5.4xlarge)"
    rds_instances    = "~$1,470 (2x db.r5.2xlarge Multi-AZ)"
    nat_gateways     = "~$90 (2x NAT Gateway)"
    load_balancers   = "~$36 (2x ALB)"
    storage_iops     = "~$200 (High IOPS volumes)"
    total_estimated  = "~$3,476/month"
  }
}

output "resource_arns" {
  value = {
    ec2_instances = aws_instance.expensive_compute[*].arn
    rds_instances = aws_db_instance.expensive_database[*].arn
    s3_buckets    = aws_s3_bucket.expensive_storage[*].arn
  }
}