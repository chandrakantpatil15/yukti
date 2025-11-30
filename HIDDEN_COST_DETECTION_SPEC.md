# Hidden Cost Detection - Technical Specification

## Overview
Premium feature detecting 50+ AWS cost traps that standard cost tools miss. Positioned as $10M competitive moat.

---

## Detection Categories

### 1. Data Transfer Costs (15 patterns)

#### 1.1 Cross-AZ Data Transfer
```go
// Detection Logic
func DetectCrossAZTransfer(resources []Resource) []Finding {
    // Check ELB/ALB with targets in multiple AZs
    // Check RDS Multi-AZ without read replicas in same AZ as app
    // Check ElastiCache clusters with clients in different AZs
    // Estimate: $0.01/GB * monthly transfer volume
}
```
**Impact**: $500-$5K/month per misconfigured service  
**Recommendation**: Co-locate resources in same AZ, use VPC endpoints

#### 1.2 NAT Gateway Data Processing
```go
func DetectNATGatewayWaste(natGateways []NATGateway) []Finding {
    // Check data processed >1TB/month
    // Compare cost vs VPC endpoints ($0.01/GB vs $0.01/hour)
    // Identify S3/DynamoDB traffic that could use gateway endpoints (free)
}
```
**Impact**: $45/TB vs $7.20/month for VPC endpoint  
**Recommendation**: Use VPC gateway endpoints for S3/DynamoDB (saves 84%)

#### 1.3 CloudFront Origin Fetches
```go
func DetectCloudFrontOriginWaste(distributions []CloudFrontDistribution) []Finding {
    // Check cache hit ratio <80%
    // Check TTL settings <1 hour for static content
    // Estimate wasted origin fetches * $0.085/GB
}
```
**Impact**: $200-$2K/month per distribution  
**Recommendation**: Increase TTL, enable compression, optimize cache policies

#### 1.4 Inter-Region Replication
```go
func DetectInterRegionReplication(resources []Resource) []Finding {
    // S3 cross-region replication: $0.02/GB
    // RDS cross-region snapshots: $0.02/GB
    // DynamoDB global tables: $0.02/GB write replication
    // Check if replication is actually needed (compliance, DR)
}
```
**Impact**: $1K-$10K/month for high-volume replication  
**Recommendation**: Use same-region replication, compress before transfer

#### 1.5 VPC Peering Data Transfer
```go
func DetectVPCPeeringCosts(peerings []VPCPeering) []Finding {
    // Intra-region peering: $0.01/GB
    // Inter-region peering: $0.02/GB
    // Check if Transit Gateway would be cheaper for hub-spoke (>5 VPCs)
}
```
**Impact**: $500-$3K/month  
**Recommendation**: Consolidate VPCs, use Transit Gateway for complex topologies

### 2. Storage Lifecycle Waste (12 patterns)

#### 2.1 EBS Snapshots Never Deleted
```go
func DetectOrphanedSnapshots(snapshots []EBSSnapshot) []Finding {
    // Check snapshots >90 days old
    // Check if source volume still exists
    // Check if snapshot used in any AMI
    // Cost: $0.05/GB-month
}
```
**Impact**: $500-$5K/month for large environments  
**Recommendation**: Implement 90-day retention policy, delete orphaned snapshots

#### 2.2 S3 Intelligent-Tiering Overhead
```go
func DetectIntelligentTieringWaste(buckets []S3Bucket) []Finding {
    // Monitoring fee: $0.0025 per 1,000 objects
    // Check if bucket has <128KB objects (not cost-effective)
    // Check if access patterns are predictable (use lifecycle rules instead)
}
```
**Impact**: $100-$1K/month for buckets with millions of small objects  
**Recommendation**: Use lifecycle rules for predictable patterns, standard storage for <128KB

#### 2.3 RDS Automated Backups Beyond Retention
```go
func DetectRDSBackupWaste(instances []RDSInstance) []Finding {
    // Free backup storage = DB size
    // Additional backups: $0.095/GB-month
    // Check if retention >7 days is needed
    // Check for manual snapshots duplicating automated backups
}
```
**Impact**: $200-$2K/month per large database  
**Recommendation**: Reduce retention to 7 days, delete manual snapshots, use S3 for long-term

#### 2.4 Glacier Retrieval Costs
```go
func DetectGlacierRetrievalRisk(buckets []S3Bucket) []Finding {
    // Expedited: $0.03/GB + $0.01 per request
    // Standard: $0.01/GB + $0.05 per 1,000 requests
    // Check if lifecycle moved data to Glacier that's frequently accessed
    // Check if Glacier Deep Archive used for data needed within 12 hours
}
```
**Impact**: $1K-$10K for accidental bulk retrieval  
**Recommendation**: Use S3 Glacier Instant Retrieval for frequent access, test restore procedures

#### 2.5 EFS Infrequent Access Not Enabled
```go
func DetectEFSLifecycleWaste(filesystems []EFS) []Finding {
    // Standard: $0.30/GB-month
    // Infrequent Access: $0.025/GB-month (92% savings)
    // Check if lifecycle policy enabled
    // Estimate savings based on file access patterns
}
```
**Impact**: $500-$5K/month for large file systems  
**Recommendation**: Enable lifecycle policy (30-day transition to IA)

### 3. Compute Waste (10 patterns)

#### 3.1 EC2 Detailed Monitoring
```go
func DetectDetailedMonitoringWaste(instances []EC2Instance) []Finding {
    // Cost: $2.10/instance/month for 1-min metrics
    // Check if detailed monitoring actually used in alarms/dashboards
    // 99% of instances don't need 1-min granularity
}
```
**Impact**: $200-$2K/month for 100-1000 instances  
**Recommendation**: Disable detailed monitoring, use 5-min free metrics

#### 3.2 Lambda Memory Over-Provisioning
```go
func DetectLambdaMemoryWaste(functions []LambdaFunction) []Finding {
    // Check CloudWatch Logs for actual memory usage
    // If max memory used <60% of allocated, over-provisioned
    // Cost scales linearly with memory (128MB to 10GB)
}
```
**Impact**: $100-$1K/month for high-volume functions  
**Recommendation**: Right-size memory to 1.2x max observed usage

#### 3.3 ECS/Fargate Over-Provisioned Tasks
```go
func DetectFargateWaste(tasks []ECSTask) []Finding {
    // Check CPU/memory utilization from Container Insights
    // If avg utilization <40%, over-provisioned
    // Fargate pricing: $0.04048/vCPU/hour + $0.004445/GB/hour
}
```
**Impact**: $500-$5K/month for large clusters  
**Recommendation**: Right-size tasks, use Fargate Spot for fault-tolerant workloads (70% savings)

#### 3.4 Elastic Beanstalk Hidden Costs
```go
func DetectBeanstalkWaste(environments []BeanstalkEnvironment) []Finding {
    // No charge for Beanstalk, but underlying resources:
    // - ELB: $0.0225/hour ($16.20/month)
    // - Auto Scaling: Free, but instances charged
    // - CloudWatch Logs: $0.50/GB ingested
    // Check if environment is dev/test (use single instance instead of HA)
}
```
**Impact**: $200-$1K/month per environment  
**Recommendation**: Use single-instance for non-prod, delete unused environments

### 4. Database Inefficiencies (8 patterns)

#### 4.1 RDS Multi-AZ for Non-Production
```go
func DetectRDSMultiAZWaste(instances []RDSInstance) []Finding {
    // Multi-AZ doubles cost (2x instance + 2x storage)
    // Check tags for env=dev/test/staging
    // Check if instance has <95% uptime requirement
}
```
**Impact**: $500-$5K/month per database  
**Recommendation**: Disable Multi-AZ for non-prod, use snapshots for recovery

#### 4.2 DynamoDB On-Demand for Predictable Workloads
```go
func DetectDynamoDBPricingWaste(tables []DynamoDBTable) []Finding {
    // On-Demand: $1.25/million writes, $0.25/million reads
    // Provisioned: $0.00065/WCU/hour, $0.00013/RCU/hour
    // Check if traffic is predictable (CV <30%)
    // Calculate breakeven: On-Demand cheaper if <20% utilization
}
```
**Impact**: $200-$2K/month per high-traffic table  
**Recommendation**: Switch to provisioned capacity with auto-scaling for predictable loads

#### 4.3 RDS Storage Auto-Scaling Runaway
```go
func DetectRDSStorageRunaway(instances []RDSInstance) []Finding {
    // Auto-scaling increases storage but never decreases
    // Check if storage increased >50% in 30 days
    // Check if actual usage <60% of allocated
    // Cost: $0.115/GB-month (gp3)
}
```
**Impact**: $100-$1K/month per database  
**Recommendation**: Disable auto-scaling, manually resize, or restore from snapshot to smaller volume

#### 4.4 ElastiCache Reserved Nodes Not Used
```go
func DetectElastiCacheReservationWaste(reservations []ElastiCacheReservation) []Finding {
    // Check if reserved nodes match running nodes (type, count, region)
    // Unused reservations still charged (no refunds)
    // Cost: $500-$5K/year per unused reservation
}
```
**Impact**: $500-$5K/year per unused reservation  
**Recommendation**: Modify reservations to match actual usage, sell on Reserved Instance Marketplace

### 5. Networking Costs (5 patterns)

#### 5.1 Load Balancer Idle Charges
```go
func DetectIdleLoadBalancers(loadBalancers []LoadBalancer) []Finding {
    // ALB: $0.0225/hour ($16.20/month) + $0.008/LCU-hour
    // NLB: $0.0225/hour ($16.20/month) + $0.006/NLCU-hour
    // Check if <100 requests/hour (effectively idle)
    // Check if targets are healthy
}
```
**Impact**: $200-$2K/month for 10-100 idle LBs  
**Recommendation**: Delete idle LBs, consolidate low-traffic apps behind single ALB

#### 5.2 Elastic IP Not Attached
```go
func DetectUnattachedEIPs(eips []ElasticIP) []Finding {
    // Attached: Free
    // Unattached: $0.005/hour ($3.60/month)
    // Additional EIPs on running instance: $0.005/hour each
}
```
**Impact**: $50-$500/month for 10-100 unattached EIPs  
**Recommendation**: Release unattached EIPs, use single EIP per instance

#### 5.3 VPN Connection Idle Charges
```go
func DetectIdleVPNConnections(vpns []VPNConnection) []Finding {
    // Cost: $0.05/hour ($36/month) per VPN connection
    // Check if data transfer <1GB/month (effectively idle)
    // Check if VPN used for dev/test (use temporary connections)
}
```
**Impact**: $100-$1K/month for 3-30 idle VPNs  
**Recommendation**: Delete idle VPNs, use AWS Client VPN for temporary access

### 6. Managed Service Premiums (5 patterns)

#### 6.1 AWS Managed Prometheus/Grafana
```go
func DetectManagedObservabilityWaste(services []ManagedService) []Finding {
    // AMP: $0.30/metric/month + $0.10/GB ingested
    // AMG: $9/active user/month
    // Compare vs self-hosted: EC2 t3.large ($60/month) + EBS ($20/month)
    // Breakeven: >10 users or >300 metrics
}
```
**Impact**: $500-$5K/month for large deployments  
**Recommendation**: Self-host for <10 users, use managed for scale/compliance

#### 6.2 AWS Transfer Family
```go
func DetectTransferFamilyWaste(servers []TransferServer) []Finding {
    // Cost: $0.30/hour ($216/month) per endpoint
    // Data transfer: $0.04/GB
    // Check if used <10 hours/month (use Lambda + S3 instead)
}
```
**Impact**: $200-$2K/month per server  
**Recommendation**: Use Lambda + S3 for infrequent transfers, API Gateway for frequent

#### 6.3 AWS Backup vs Native Snapshots
```go
func DetectAWSBackupPremium(backupPlans []BackupPlan) []Finding {
    // AWS Backup: $0.05/GB-month + $0.02/GB restore
    // EBS snapshots: $0.05/GB-month + free restore
    // RDS snapshots: $0.095/GB-month + free restore
    // Premium for centralized management: 0-40% depending on service
}
```
**Impact**: $100-$1K/month  
**Recommendation**: Use native snapshots for simple use cases, AWS Backup for compliance/cross-region

### 7. Serverless Inefficiencies (3 patterns)

#### 7.1 API Gateway REST vs HTTP API
```go
func DetectAPIGatewayWaste(apis []APIGateway) []Finding {
    // REST API: $3.50/million requests
    // HTTP API: $1.00/million requests (71% cheaper)
    // Check if REST-only features used (API keys, request validation, SDK generation)
}
```
**Impact**: $100-$1K/month for high-traffic APIs  
**Recommendation**: Migrate to HTTP API unless REST-specific features needed

#### 7.2 Step Functions Express vs Standard
```go
func DetectStepFunctionsWaste(stateMachines []StateMachine) []Finding {
    // Standard: $0.025 per 1,000 state transitions
    // Express: $1.00 per 1 million requests + $0.00001667/GB-second
    // Express cheaper for >40 transitions per execution
}
```
**Impact**: $50-$500/month for high-volume workflows  
**Recommendation**: Use Express for high-volume, short-duration (<5 min) workflows

### 8. Container Waste (2 patterns)

#### 8.1 ECR Image Scanning Costs
```go
func DetectECRScanningWaste(repositories []ECRRepository) []Finding {
    // Basic scanning: Free (Clair-based)
    // Enhanced scanning: $0.09/image scan
    // Check if enhanced scanning enabled for all images (including dev/test)
}
```
**Impact**: $100-$1K/month for large registries  
**Recommendation**: Use basic scanning for non-prod, enhanced for production only

#### 8.2 EKS Control Plane Costs
```go
func DetectEKSControlPlaneWaste(clusters []EKSCluster) []Finding {
    // Cost: $0.10/hour ($73/month) per cluster
    // Check if multiple clusters for dev/test (consolidate with namespaces)
    // Check if cluster has <10 nodes (overhead too high)
}
```
**Impact**: $200-$2K/month for 3-30 clusters  
**Recommendation**: Consolidate dev/test into single cluster, use namespaces for isolation

---

## Implementation Architecture

### Detection Engine
```go
type HiddenCostDetector struct {
    awsClient    *aws.Client
    mlService    *ml.Client
    cache        *redis.Client
    detectors    []Detector
}

type Detector interface {
    Name() string
    Category() string
    Detect(ctx context.Context, resources []Resource) ([]Finding, error)
    EstimateSavings(finding Finding) float64
}

type Finding struct {
    ID              string
    DetectorName    string
    Category        string
    Severity        string // Critical, High, Medium, Low
    Title           string
    Description     string
    ResourceARN     string
    EstimatedCost   float64 // Current monthly cost
    EstimatedSavings float64 // Potential monthly savings
    Confidence      float64 // 0.0-1.0
    Recommendation  string
    RemediationSteps []string
    IaCCode         string // Terraform/CloudFormation to fix
    DetectedAt      time.Time
}

func (d *HiddenCostDetector) RunDetection(ctx context.Context, tenantID string) ([]Finding, error) {
    // 1. Fetch resources from cache (6-hour refresh)
    resources := d.cache.GetResources(tenantID)
    
    // 2. Run all detectors in parallel
    findings := []Finding{}
    for _, detector := range d.detectors {
        detectorFindings, _ := detector.Detect(ctx, resources)
        findings = append(findings, detectorFindings...)
    }
    
    // 3. Deduplicate and prioritize
    findings = d.deduplicateFindings(findings)
    findings = d.prioritizeByROI(findings)
    
    // 4. Store in database
    d.storeFindings(tenantID, findings)
    
    return findings, nil
}
```

### ML-Enhanced Detection
```python
# ml-service/hidden_costs.py
class HiddenCostPredictor:
    def predict_data_transfer_costs(self, resource_topology):
        """Predict cross-AZ/region data transfer based on topology"""
        # Train on historical CloudWatch metrics + Cost Explorer data
        # Features: resource types, AZs, regions, network topology
        # Output: Predicted monthly data transfer cost
        
    def detect_anomalous_costs(self, cost_timeseries):
        """Detect unusual cost patterns that indicate hidden costs"""
        # Use Isolation Forest for anomaly detection
        # Flag sudden spikes, gradual increases, unexpected patterns
        
    def estimate_savings_confidence(self, finding):
        """Calculate confidence score for savings estimate"""
        # Based on: data quality, historical accuracy, resource stability
        # Output: 0.0-1.0 confidence score
```

### Caching Strategy
```go
// Cache resource metadata for 6 hours
// Cache detection results for 24 hours
// Invalidate cache on resource changes (CloudWatch Events)

type CacheKey struct {
    TenantID   string
    Detector   string
    ResourceID string
}

func (d *HiddenCostDetector) getCachedFindings(key CacheKey) ([]Finding, bool) {
    val, err := d.cache.Get(key.String())
    if err != nil {
        return nil, false
    }
    var findings []Finding
    json.Unmarshal([]byte(val), &findings)
    return findings, true
}
```

---

## API Endpoints

### List Hidden Costs
```http
GET /api/v1/hidden-costs
Authorization: Bearer <jwt_token>

Query Parameters:
- category: string (optional) - Filter by category
- severity: string (optional) - Filter by severity
- min_savings: float (optional) - Minimum monthly savings
- sort: string (optional) - Sort by: savings, confidence, detected_at

Response:
{
  "findings": [
    {
      "id": "hc_abc123",
      "detector": "cross_az_data_transfer",
      "category": "Data Transfer Costs",
      "severity": "High",
      "title": "RDS instance transferring 500GB/month across AZs",
      "description": "Your RDS instance in us-east-1a is serving traffic from EC2 instances in us-east-1b, incurring $10/GB cross-AZ charges.",
      "resource_arn": "arn:aws:rds:us-east-1:123456789012:db:prod-db",
      "estimated_cost": 500.00,
      "estimated_savings": 450.00,
      "confidence": 0.95,
      "recommendation": "Move RDS read replica to us-east-1b or move EC2 instances to us-east-1a",
      "remediation_steps": [
        "Create RDS read replica in us-east-1b",
        "Update application to use read replica for queries",
        "Monitor cross-AZ traffic reduction"
      ],
      "iac_code": "resource \"aws_db_instance\" \"read_replica\" { ... }",
      "detected_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total_findings": 47,
  "total_estimated_savings": 12450.00,
  "categories": {
    "Data Transfer Costs": 15,
    "Storage Lifecycle Waste": 12,
    "Compute Waste": 10,
    "Database Inefficiencies": 8,
    "Networking Costs": 2
  }
}
```

### Get Hidden Cost Details
```http
GET /api/v1/hidden-costs/{id}
Authorization: Bearer <jwt_token>

Response:
{
  "finding": { ... },
  "historical_cost": [
    {"month": "2024-11", "cost": 480.00},
    {"month": "2024-12", "cost": 495.00},
    {"month": "2025-01", "cost": 500.00}
  ],
  "similar_findings": [
    {"id": "hc_def456", "title": "Another RDS cross-AZ issue", "savings": 320.00}
  ],
  "remediation_impact": {
    "effort": "Medium (2-4 hours)",
    "risk": "Low (read replica creation is safe)",
    "downtime": "None (zero-downtime migration)"
  }
}
```

### Suppress Hidden Cost
```http
POST /api/v1/hidden-costs/{id}/suppress
Authorization: Bearer <jwt_token>

Request:
{
  "reason": "Business requirement to keep RDS in us-east-1a for compliance",
  "suppressed_until": "2025-12-31T23:59:59Z"
}

Response:
{
  "success": true,
  "message": "Hidden cost suppressed until 2025-12-31"
}
```

---

## UI Components

### Hidden Costs Dashboard
```jsx
// frontend/src/pages/HiddenCosts.js
function HiddenCostsDashboard() {
  return (
    <div>
      <SavingsSummary totalSavings={12450} findingsCount={47} />
      <CategoryBreakdown categories={categories} />
      <FindingsTable findings={findings} onSuppress={handleSuppress} />
      <SavingsTrend data={historicalSavings} />
    </div>
  );
}
```

### Finding Detail Panel
```jsx
function FindingDetailPanel({ finding }) {
  return (
    <SlideOutPanel>
      <FindingHeader finding={finding} />
      <CostImpact current={500} potential={50} savings={450} />
      <RemediationSteps steps={finding.remediation_steps} />
      <IaCCodeBlock code={finding.iac_code} language="terraform" />
      <ActionButtons onApply={apply} onSuppress={suppress} onIgnore={ignore} />
    </SlideOutPanel>
  );
}
```

---

## Pricing Strategy

### Feature Availability by Tier
- **FREE**: 0 hidden cost detectors (teaser only)
- **PROFESSIONAL**: 25 detectors (Data Transfer, Storage, Compute)
- **ENTERPRISE**: 50 detectors (All categories)
- **FINANCIAL**: 50 detectors + custom detectors + priority support

### Upsell Messaging
```
"We detected $12,450 in hidden costs, but you're only seeing 25 of 47 findings.
Upgrade to ENTERPRISE to unlock all 50 detectors and save an additional $5,200/month.
ROI: 10.4x (pay $499/month, save $5,200/month)"
```

---

## Success Metrics

### Product Metrics
- **Detection Accuracy**: >90% of findings result in actual savings
- **False Positive Rate**: <10% of findings are invalid
- **Savings Realization**: 60% of findings are acted upon within 30 days
- **Time to Detect**: New hidden costs detected within 24 hours

### Business Metrics
- **Upsell Rate**: 30% of PROFESSIONAL users upgrade to ENTERPRISE for hidden costs
- **Customer Savings**: Average $8,500/month in hidden costs detected per customer
- **Competitive Moat**: No competitor offers >20 hidden cost detectors

---

**Last Updated**: January 2025  
**Owner**: Product & Engineering Teams  
**Status**: Phase 1 (25 detectors) in production, Phase 2 (50 detectors) in development
