# CloudWatch & CloudTrail Integration Guide

## Overview
Integration guide for connecting AWS CloudWatch (utilization metrics) and CloudTrail (audit logs) with ClickHouse, including resource name mapping and utilization analysis.

## Architecture Flow

```
┌─────────────────┐         ┌──────────────────┐
│  AWS CloudWatch │────────▶│ CloudWatch       │
│  (Utilization)  │         │ Collector        │
└─────────────────┘         └────────┬─────────┘
                                     │
┌─────────────────┐         ┌────────▼─────────┐
│  AWS CloudTrail │────────▶│ CloudTrail       │
│  (Audit Logs)   │         │ Collector        │
└─────────────────┘         └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │ Resource Mapper  │
                            │ (Name Matching)  │
                            └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │   ClickHouse     │
                            │  (Storage)       │
                            └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │ Utilization      │
                            │ Analyzer         │
                            └────────┬─────────┐
                                     │         │
                            ┌────────▼───┐    │
                            │   Redis    │    │
                            │   (Cache)  │    │
                            └────────────┘    │
                                     │        │
                            ┌────────▼────────▼─┐
                            │   Frontend/API    │
                            │  (Decisions)      │
                            └───────────────────┘
```

## CloudWatch Integration

### Purpose
- Collect resource utilization metrics (CPU, Memory, Network, Disk)
- Map CloudWatch resources to yukti resources by name
- Store metrics in ClickHouse for analysis

### Metrics Collected

#### EC2 Instances
- `CPUUtilization` (Average, Maximum)
- `NetworkIn`, `NetworkOut` (Bytes)
- `DiskReadOps`, `DiskWriteOps` (Count)
- `DiskReadBytes`, `DiskWriteBytes` (Bytes)

#### RDS Instances
- `CPUUtilization` (Average, Maximum)
- `DatabaseConnections` (Count)
- `ReadLatency`, `WriteLatency` (Milliseconds)
- `FreeStorageSpace` (Bytes)

#### Lambda Functions
- `Invocations` (Count)
- `Duration` (Milliseconds)
- `Errors` (Count)
- `Throttles` (Count)
- `ConcurrentExecutions` (Count)

### Resource Name Mapping

CloudWatch dimensions contain resource identifiers:
- EC2: `InstanceId` dimension
- RDS: `DBInstanceIdentifier` dimension
- Lambda: `FunctionName` dimension

Mapping strategy:
1. Extract resource identifier from CloudWatch dimensions
2. Query PostgreSQL `yt_tenant_resources` table
3. Match by `resource_id` or `resource_name`
4. Store mapping in ClickHouse `yt_resource_mapping` table
5. Cache mappings in Redis for performance

### Collection Frequency
- Default: Every 5 minutes
- Configurable via `AWS_CLOUDWATCH_METRICS_INTERVAL`
- Supports multiple AWS accounts and regions

## CloudTrail Integration

### Purpose
- Collect audit logs for security and compliance
- Track resource access and modifications
- Map CloudTrail events to yukti resources by name

### Event Types Collected

#### Resource Operations
- `RunInstances`, `TerminateInstances` (EC2)
- `CreateDBInstance`, `DeleteDBInstance` (RDS)
- `CreateFunction`, `DeleteFunction` (Lambda)
- `CreateBucket`, `DeleteBucket` (S3)
- And other resource-specific operations

#### Security Events
- `AssumeRole`, `AssumeRoleWithSAML`
- `ConsoleLogin`
- `CreateAccessKey`, `DeleteAccessKey`
- Failed API calls (error_code present)

### Resource Name Extraction

CloudTrail events contain resource information in:
1. `resources` array (resourceName, resourceType)
2. `requestParameters` JSON (varies by API)
3. Resource ARN format

Extraction strategy:
```go
// Example: RunInstances event
resourceName := requestParameters["instanceId"] // or tag Name
resourceType := "ec2-instance"
region := event.awsRegion
accountId := event.recipientAccountId
```

### Log Processing

1. **S3 Log Files**: CloudTrail stores logs in S3
2. **Download**: Fetch new log files since last check
3. **Parse**: Extract JSON events from compressed files
4. **Filter**: Filter by event names and regions
5. **Map**: Map resource names to yukti resources
6. **Store**: Insert into ClickHouse `yt_audit_logs` table

### Collection Frequency
- Default: Every 15 minutes
- Configurable via CloudTrail collector interval
- Tracks processed files to avoid duplicates

## Resource Mapping Service

### Mapping Strategy

#### 1. Exact Match (Primary)
```
CloudWatch: i-1234567890abcdef0
Yukti DB: resource_id = 'i-1234567890abcdef0'
Result: Direct match
```

#### 2. Name-based Match
```
CloudWatch: my-production-server (from tags or Name tag)
Yukti DB: resource_name = 'my-production-server'
Result: Match by name
```

#### 3. ARN Parsing
```
CloudTrail ARN: arn:aws:ec2:us-east-1:123456789:instance/i-1234567890abcdef0
Extracted: instance-id = 'i-1234567890abcdef0'
Yukti DB: resource_id = 'i-1234567890abcdef0'
Result: Match after ARN parsing
```

#### 4. Fuzzy Matching (Fallback)
```
CloudWatch: my-prod-server-01
Yukti DB: resource_name = 'my-production-server-01'
Levenshtein distance: 3
Threshold: 5
Result: Match (within threshold)
```

### Mapping Cache

- **Redis Key**: `mapping:tenant:{id}:name:{name}:type:{type}`
- **Value**: `yukti_resource_id`
- **TTL**: 24 hours
- **Purpose**: Fast lookup for real-time mapping

### Mapping Storage

- **ClickHouse Table**: `yt_resource_mapping`
- **Engine**: ReplacingMergeTree (keeps latest mapping)
- **Purpose**: Historical tracking and batch operations

### Background Sync Job

- **Frequency**: Every 1 hour
- **Actions**:
  1. Query all yukti resources from PostgreSQL
  2. Update ClickHouse mapping table
  3. Refresh Redis cache
  4. Handle new resources

## Utilization Analysis

### Analysis Process

1. **Query Metrics**: Query `yt_utilization_metrics` from ClickHouse
2. **Calculate Statistics**: 
   - Average CPU/Memory usage
   - Maximum CPU/Memory usage
   - Percentiles (p50, p95, p99)
3. **Compare Thresholds**:
   - Under-utilized: avg < low_threshold AND max < low_threshold
   - Over-utilized: avg > high_threshold OR max > high_threshold
   - Optimal: between thresholds
4. **Generate Recommendations**:
   - Under-utilized → "downsize" or "right-size to smaller instance"
   - Over-utilized → "upsize" or "scale horizontally"
   - Optimal → "maintain current size"
5. **Calculate Confidence**:
   - Data points count (more = higher)
   - Variance (lower = higher)
   - Time range coverage
6. **Store Results**: Insert into `yt_utilization_analysis` table

### Thresholds (Default)

- **CPU Low**: 20% (resources using < 20% are under-utilized)
- **CPU High**: 80% (resources using > 80% are over-utilized)
- **Memory Low**: 20%
- **Memory High**: 80%

Configurable via environment variables.

### Analysis Time Ranges

- **1 Day**: Recent utilization trends
- **7 Days**: Weekly patterns
- **30 Days**: Monthly trends
- **Custom**: User-defined ranges

### Recommendations Examples

#### Under-Utilized
```json
{
  "resource_id": "i-1234567890abcdef0",
  "resource_name": "web-server-01",
  "utilization_status": "under_utilized",
  "avg_cpu_usage": 12.5,
  "avg_memory_usage": 18.3,
  "recommendation": "Consider downsizing from t3.large to t3.medium",
  "confidence_score": 0.92,
  "estimated_savings": 45.50
}
```

#### Over-Utilized
```json
{
  "resource_id": "i-0987654321fedcba0",
  "resource_name": "db-server-01",
  "utilization_status": "over_utilized",
  "avg_cpu_usage": 87.2,
  "max_cpu_usage": 98.5,
  "recommendation": "Consider upsizing to larger instance or add read replicas",
  "confidence_score": 0.88
}
```

## Decision Making

### Under-Utilized Resources
- **Decision**: Right-size or downsize
- **Action**: Recommend smaller instance type
- **Savings**: Calculate cost difference
- **Risk**: Low (if max usage is also low)

### Over-Utilized Resources
- **Decision**: Scale up or out
- **Action**: Recommend larger instance or horizontal scaling
- **Performance**: Prevent throttling/errors
- **Cost**: May increase, but necessary for performance

### Optimal Resources
- **Decision**: Maintain current size
- **Action**: No changes needed
- **Monitoring**: Continue tracking

## API Endpoints

### Utilization Analysis
- `GET /api/v1/utilization/recommendations` - All recommendations
- `GET /api/v1/utilization/under-utilized` - Under-utilized resources
- `GET /api/v1/utilization/over-utilized` - Over-utilized resources
- `GET /api/v1/utilization/analysis/:resource_id` - Specific resource analysis
- `POST /api/v1/utilization/analyze` - Trigger new analysis

### Audit Logs
- `GET /api/v1/audit/logs` - All audit logs (filtered)
- `GET /api/v1/audit/logs/:resource_id` - Resource-specific logs

## Implementation Checklist

### Phase 1: CloudWatch Collector
- [ ] AWS SDK setup
- [ ] CloudWatch client initialization
- [ ] Metrics collection logic
- [ ] Resource name extraction
- [ ] Batch insert to ClickHouse
- [ ] Error handling and retries

### Phase 2: CloudTrail Collector
- [ ] S3 client setup
- [ ] CloudTrail log file download
- [ ] Event parsing
- [ ] Resource name extraction
- [ ] Batch insert to ClickHouse
- [ ] Duplicate prevention

### Phase 3: Resource Mapper
- [ ] PostgreSQL query logic
- [ ] Exact matching
- [ ] Name-based matching
- [ ] ARN parsing
- [ ] Fuzzy matching (optional)
- [ ] Redis caching
- [ ] ClickHouse storage

### Phase 4: Utilization Analyzer
- [ ] ClickHouse query logic
- [ ] Statistics calculation
- [ ] Threshold comparison
- [ ] Recommendation generation
- [ ] Confidence scoring
- [ ] Results storage

### Phase 5: API & Frontend
- [ ] API endpoints
- [ ] Frontend API methods
- [ ] UI components
- [ ] Visualization charts
- [ ] Recommendations display

## Troubleshooting

### CloudWatch Issues
- **No metrics**: Check AWS permissions, namespace, dimensions
- **Wrong resources**: Verify resource mapping
- **Rate limits**: Implement exponential backoff

### CloudTrail Issues
- **No logs**: Check S3 bucket access, log prefix
- **Missing events**: Verify CloudTrail is enabled for region
- **Parsing errors**: Handle malformed JSON gracefully

### Mapping Issues
- **No matches**: Check resource names in database
- **Wrong matches**: Improve fuzzy matching threshold
- **Cache stale**: Reduce cache TTL or force refresh

### Analysis Issues
- **Incorrect recommendations**: Adjust thresholds
- **Low confidence**: Increase time range or data points
- **Missing data**: Check CloudWatch collection status

## Security Considerations

1. **AWS Credentials**: Use IAM roles or environment variables
2. **S3 Access**: Restrict CloudTrail S3 bucket access
3. **Tenant Isolation**: Always filter by tenant_id in queries
4. **Audit Logs**: Store sensitive fields securely
5. **Resource Mapping**: Validate tenant ownership

## Performance Optimization

1. **Batch Operations**: Batch inserts to ClickHouse (1000+ records)
2. **Caching**: Cache mappings in Redis
3. **Materialized Views**: Use for aggregations
4. **Indexes**: Proper indexes on resource_name, tenant_id
5. **Parallel Collection**: Collect from multiple regions in parallel

