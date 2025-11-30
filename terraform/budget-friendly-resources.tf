# Budget-Friendly AWS Resources for Yukti FinOps Testing
# Designed for $100 AWS credit - will last 2-3 weeks with optimization opportunities

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

# 1. MODERATE EC2 INSTANCES (~$50/month total)
resource "aws_instance" "test_compute" {
  count         = 2
  ami           = "ami-0c02fb55956c7d316" # Amazon Linux 2
  instance_type = "t3.large"             # $0.0832/hour = $60/month
  
  # Moderate EBS volume with optimization opportunities
  root_block_device {
    volume_type = "gp2"      # Can optimize to gp3
    volume_size = 100        # Can right-size
    encrypted   = false      # Security finding
  }
  
  # Additional volume for optimization
  ebs_block_device {
    device_name = "/dev/sdf"
    volume_type = "io1"      # Expensive - can optimize to gp3
    volume_size = 50
    iops        = 1000       # Over-provisioned
  }
  
  # Detailed monitoring (optimization opportunity)
  monitoring = true
  
  tags = {
    Name        = "test-compute-${count.index + 1}"
    Environment = "development"  # Can use spot instances
    CostCenter  = "testing"
    Owner       = "finops-team"
    Purpose     = "cost-optimization-demo"
  }
}

# 2. SMALL RDS INSTANCE (~$25/month)
resource "aws_db_instance" "test_database" {
  identifier = "test-db-1"
  
  instance_class = "db.t3.micro"  # $0.017/hour = $12/month
  engine         = "postgres"
  engine_version = "15.4"
  
  allocated_storage = 100        # Can optimize storage
  storage_type      = "gp2"      # Can upgrade to gp3
  
  # Multi-AZ disabled but can be optimization finding
  multi_az = false
  
  # Long backup retention (cost optimization opportunity)
  backup_retention_period = 14
  backup_window          = "03:00-04:00"
  
  db_name  = "testdb"
  username = "admin"
  password = "ChangeMe123!"
  
  skip_final_snapshot = true
  deletion_protection = false
  
  tags = {
    Name        = "test-database"
    Environment = "development"
    CostCenter  = "testing"
  }
}

# 3. S3 BUCKETS with optimization opportunities
resource "aws_s3_bucket" "test_storage" {
  count         = 3
  bucket        = "yukti-test-storage-${count.index + 1}-${random_id.bucket_suffix.hex}"
  force_destroy = true  # Ensures complete cleanup
  
  tags = {
    Name        = "test-storage-${count.index + 1}"
    Environment = "development"
    CostCenter  = "testing"
  }
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# S3 versioning (optimization opportunity)
resource "aws_s3_bucket_versioning" "test_versioning" {
  count  = length(aws_s3_bucket.test_storage)
  bucket = aws_s3_bucket.test_storage[count.index].id
  
  versioning_configuration {
    status = "Enabled"  # Can add lifecycle policies
  }
}

# 4. LOAD BALANCER (~$18/month)
resource "aws_lb" "test_alb" {
  name               = "test-alb"
  internal           = false
  load_balancer_type = "application"
  subnets            = [aws_subnet.public_1.id, aws_subnet.public_2.id]
  
  tags = {
    Name        = "test-alb"
    Environment = "development"
  }
}

# 5. NAT GATEWAY (~$45/month) - Major optimization opportunity
resource "aws_nat_gateway" "test_nat" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = aws_subnet.public_1.id
  
  tags = {
    Name = "test-nat-gateway"
  }
}

resource "aws_eip" "nat_eip" {
  domain = "vpc"
  
  tags = {
    Name = "test-nat-eip"
  }
}

# 6. VPC and Networking
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

resource "aws_subnet" "private_1" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = "${var.aws_region}a"
  
  tags = {
    Name = "private-subnet-1"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  
  tags = {
    Name = "main-igw"
  }
}

# Route table for private subnet (uses NAT Gateway)
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.test_nat.id
  }
  
  tags = {
    Name = "private-route-table"
  }
}

resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private_1.id
  route_table_id = aws_route_table.private.id
}

# 7. UNUSED RESOURCES (optimization opportunities)
resource "aws_ebs_volume" "unused_volume" {
  availability_zone = "${var.aws_region}a"
  size              = 50
  type              = "gp2"
  
  tags = {
    Name = "unused-volume"  # Will be detected as unused
    AutoDelete = "true"     # Cleanup marker
  }
}

# Lifecycle rule to prevent accidental long-running costs
resource "aws_s3_bucket_lifecycle_configuration" "cleanup" {
  count  = length(aws_s3_bucket.test_storage)
  bucket = aws_s3_bucket.test_storage[count.index].id

  rule {
    id     = "cleanup_rule"
    status = "Enabled"

    expiration {
      days = 30  # Auto-delete after 30 days
    }

    noncurrent_version_expiration {
      noncurrent_days = 7  # Delete old versions after 7 days
    }
  }
}

# Output budget-friendly costs
output "estimated_monthly_costs" {
  value = {
    ec2_instances    = "~$120 (2x t3.large)"
    rds_instance     = "~$12 (1x db.t3.micro)"
    nat_gateway      = "~$45 (1x NAT Gateway)"
    load_balancer    = "~$18 (1x ALB)"
    storage          = "~$15 (EBS + S3)"
    total_estimated  = "~$210/month"
    daily_cost       = "~$7/day"
    credit_duration  = "~14 days with $100 credit"
  }
}

output "optimization_opportunities" {
  value = {
    ec2_rightsizing     = "$60/month (t3.large → t3.medium)"
    spot_instances      = "$84/month (70% savings for dev)"
    nat_alternative     = "$36/month (NAT Gateway → NAT Instance)"
    storage_optimization = "$8/month (io1 → gp3, gp2 → gp3)"
    unused_resources    = "$5/month (delete unused EBS volume)"
    reserved_instances  = "$36/month (30% savings with RI)"
    total_savings       = "$229/month (potential 109% cost reduction)"
  }
}

output "resource_arns" {
  value = {
    ec2_instances = aws_instance.test_compute[*].arn
    rds_instance  = [aws_db_instance.test_database.arn]
    s3_buckets    = aws_s3_bucket.test_storage[*].arn
    load_balancer = [aws_lb.test_alb.arn]
    nat_gateway   = [aws_nat_gateway.test_nat.id]
  }
}