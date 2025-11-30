# Expensive AWS Resources for Yukti FinOps Testing

This Terraform configuration creates intentionally expensive AWS resources to test the Yukti FinOps platform's cost optimization capabilities.

## 💰 **Estimated Monthly Cost: ~$3,476**

### Resources Created:
- **3x EC2 m5.4xlarge instances** (~$1,680/month)
- **2x RDS db.r5.2xlarge Multi-AZ** (~$1,470/month) 
- **2x NAT Gateways** (~$90/month)
- **2x Application Load Balancers** (~$36/month)
- **High IOPS EBS volumes** (~$200/month)

## 🚀 **Quick Deploy**

```bash
# 1. Initialize Terraform
cd terraform
terraform init

# 2. Plan deployment
terraform plan

# 3. Deploy resources
terraform apply

# 4. Get resource ARNs
terraform output resource_arns
```

## 🎯 **Cost Optimization Opportunities**

Once deployed, Yukti will detect these optimization opportunities:

### **EC2 Optimizations:**
- **Right-sizing**: m5.4xlarge → m5.2xlarge (50% savings)
- **Reserved Instances**: 1-year term (30% savings)
- **Spot Instances**: Development workloads (70% savings)
- **Detailed Monitoring**: Disable if not needed ($2.10/month per instance)

### **RDS Optimizations:**
- **Multi-AZ**: Disable for non-production (50% savings)
- **Storage Type**: io1 → gp3 (40% savings on IOPS)
- **Backup Retention**: 30 days → 7 days (reduce storage cost)
- **Performance Insights**: Disable if not used

### **Storage Optimizations:**
- **S3 Versioning**: Lifecycle policies for old versions
- **EBS Volume Types**: io2 → gp3 (significant IOPS savings)
- **Unused Volumes**: Identify and delete orphaned volumes

### **Network Optimizations:**
- **NAT Gateway**: Replace with NAT Instance (80% savings)
- **Load Balancer**: Consolidate or use Network LB
- **Data Transfer**: Optimize cross-AZ traffic

## 📊 **Expected Yukti Findings**

After deployment and scan, expect these findings:

1. **EC2 Right-sizing**: $840/month savings
2. **RDS Multi-AZ Review**: $735/month potential savings  
3. **Storage Optimization**: $120/month savings
4. **NAT Gateway Alternative**: $72/month savings
5. **Reserved Instance Recommendations**: $504/month savings

**Total Potential Savings: ~$2,271/month (65% reduction)**

## 🔧 **Testing Workflow**

1. **Deploy Resources**: `terraform apply`
2. **Connect Yukti**: Add AWS account to platform
3. **Trigger Scan**: Platform will discover resources
4. **View Findings**: Dashboard shows optimization opportunities
5. **Implement Fixes**: Use Yukti's recommendations
6. **Monitor Savings**: Track cost reduction over time

## 🧹 **Cleanup**

```bash
# Destroy all resources to avoid charges
terraform destroy
```

## ⚠️ **Cost Warning**

These resources will incur **real AWS charges**. Monitor your AWS billing and destroy resources when testing is complete.

## 🎯 **Yukti Platform Integration**

1. **Resource Discovery**: Platform scans and catalogs all resources
2. **Cost Analysis**: 77 detectors analyze each resource
3. **Optimization Recommendations**: Specific savings opportunities
4. **Implementation Guidance**: Step-by-step cost reduction
5. **Ongoing Monitoring**: Continuous optimization tracking