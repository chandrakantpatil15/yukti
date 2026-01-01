# Metrics Storage with ClickHouse Implementation Plan

## Overview
Implement ClickHouse for time-series metrics storage and Redis for caching to improve performance and scalability.

## Architecture

### Data Flow
```
AWS CloudWatch → Metrics Collector → ClickHouse (Utilization Metrics)
AWS CloudTrail → Audit Collector → ClickHouse (Audit Logs)
                                     ↓
                                Redis (Cache Layer)
                                     ↓
                                Resource Mapper (Map by resource name)
                                     ↓
                                Decision Engine (Under/Over Utilization)
                                     ↓
                                Frontend/API
```

### Data Sources
1. **AWS CloudWatch**: Utilization metrics (CPU, Memory, Network, Disk)
2. **AWS CloudTrail**: Audit logs for security and compliance
3. **Resource Mapping**: Map CloudWatch/CloudTrail resources to yukti resources by name
4. **Analysis Engine**: Detect under/over utilization patterns

## Phase 1: ClickHouse Setup

### 1.1 Database Schema Design

#### Table: `yt_metrics`
**Purpose**: Store time-series resource metrics
```sql
CREATE TABLE yt_metrics (
    timestamp DateTime DEFAULT now(),
    tenant_id UInt32,
    resource_id String,
    resource_type String,
    metric_name String,
    metric_value Float64,
    tags Map(String, String),
    region String,
    account_id String,
    INDEX idx_tenant_id tenant_id TYPE minmax GRANULARITY 3,
    INDEX idx_resource_id resource_id TYPE bloom_filter GRANULARITY 3,
    INDEX idx_metric_name metric_name TYPE bloom_filter GRANULARITY 3
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, resource_id, metric_name, timestamp)
TTL timestamp + INTERVAL 90 DAY;
```

#### Table: `yt_cost_metrics`
**Purpose**: Store cost/time series data
```sql
CREATE TABLE yt_cost_metrics (
    timestamp DateTime DEFAULT now(),
    tenant_id UInt32,
    service_name String,
    resource_id String,
    cost_amount Float64,
    currency String DEFAULT 'USD',
    region String,
    account_id String,
    tags Map(String, String),
    INDEX idx_tenant_id tenant_id TYPE minmax GRANULARITY 3,
    INDEX idx_service service_name TYPE bloom_filter GRANULARITY 3
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, service_name, timestamp)
TTL timestamp + INTERVAL 365 DAY;
```

#### Table: `yt_utilization_metrics`
**Purpose**: Store resource utilization data from CloudWatch
```sql
CREATE TABLE yt_utilization_metrics (
    timestamp DateTime DEFAULT now(),
    tenant_id UInt32,
    resource_id String,
    resource_name String,
    resource_type String,
    region String,
    account_id String,
    cpu_usage Float64,
    memory_usage Float64,
    network_in Float64,
    network_out Float64,
    disk_read Float64,
    disk_write Float64,
    cloudwatch_namespace String,
    cloudwatch_metric_name String,
    tags Map(String, String),
    INDEX idx_tenant_resource (tenant_id, resource_id) TYPE minmax GRANULARITY 3,
    INDEX idx_resource_name resource_name TYPE bloom_filter GRANULARITY 3
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, resource_name, resource_id, timestamp)
TTL timestamp + INTERVAL 90 DAY;
```

#### Table: `yt_audit_logs`
**Purpose**: Store CloudTrail audit logs
```sql
CREATE TABLE yt_audit_logs (
    timestamp DateTime,
    tenant_id UInt32,
    resource_name String,
    resource_id String,
    resource_type String,
    region String,
    account_id String,
    event_name String,
    event_source String,
    user_identity_arn String,
    source_ip_address String,
    user_agent String,
    request_parameters String,
    response_elements String,
    error_code String,
    error_message String,
    event_type String,
    api_version String,
    request_id String,
    tags Map(String, String),
    INDEX idx_tenant_resource (tenant_id, resource_name) TYPE minmax GRANULARITY 3,
    INDEX idx_event_name event_name TYPE bloom_filter GRANULARITY 3,
    INDEX idx_user_identity user_identity_arn TYPE bloom_filter GRANULARITY 3
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, resource_name, timestamp, event_name)
TTL timestamp + INTERVAL 365 DAY;
```

#### Table: `yt_resource_mapping`
**Purpose**: Map CloudWatch/CloudTrail resources to yukti resources by name
```sql
CREATE TABLE yt_resource_mapping (
    tenant_id UInt32,
    yukti_resource_id String,
    cloudwatch_resource_name String,
    cloudtrail_resource_name String,
    resource_type String,
    region String,
    account_id String,
    mapped_at DateTime DEFAULT now(),
    is_active UInt8 DEFAULT 1,
    INDEX idx_yukti_resource yukti_resource_id TYPE bloom_filter GRANULARITY 3,
    INDEX idx_cloudwatch_name cloudwatch_resource_name TYPE bloom_filter GRANULARITY 3,
    INDEX idx_cloudtrail_name cloudtrail_resource_name TYPE bloom_filter GRANULARITY 3
) ENGINE = ReplacingMergeTree(mapped_at)
ORDER BY (tenant_id, yukti_resource_id, cloudwatch_resource_name)
SETTINGS index_granularity = 8192;
```

#### Table: `yt_utilization_analysis`
**Purpose**: Store utilization analysis results (under/over utilization decisions)
```sql
CREATE TABLE yt_utilization_analysis (
    analysis_id String,
    timestamp DateTime DEFAULT now(),
    tenant_id UInt32,
    resource_id String,
    resource_name String,
    resource_type String,
    analysis_period_start DateTime,
    analysis_period_end DateTime,
    avg_cpu_usage Float64,
    avg_memory_usage Float64,
    max_cpu_usage Float64,
    max_memory_usage Float64,
    utilization_status String, -- 'under_utilized', 'optimal', 'over_utilized'
    recommendation String, -- 'downsize', 'maintain', 'upsize', 'optimize'
    confidence_score Float64, -- 0.0 to 1.0
    threshold_cpu_low Float64 DEFAULT 20.0,
    threshold_cpu_high Float64 DEFAULT 80.0,
    threshold_memory_low Float64 DEFAULT 20.0,
    threshold_memory_high Float64 DEFAULT 80.0,
    INDEX idx_tenant_resource (tenant_id, resource_id) TYPE minmax GRANULARITY 3,
    INDEX idx_status utilization_status TYPE bloom_filter GRANULARITY 3
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, resource_id, timestamp)
TTL timestamp + INTERVAL 365 DAY;
```

### 1.2 Materialized Views for Aggregations

#### Daily Aggregations
```sql
CREATE MATERIALIZED VIEW yt_metrics_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, resource_id, metric_name, date)
AS SELECT
    tenant_id,
    resource_id,
    metric_name,
    toDate(timestamp) as date,
    sum(metric_value) as total_value,
    avg(metric_value) as avg_value,
    max(metric_value) as max_value,
    min(metric_value) as min_value
FROM yt_metrics
GROUP BY tenant_id, resource_id, metric_name, date;
```

### 1.3 AWS Integration Tables

#### CloudWatch Metrics Mapping
- Map CloudWatch metrics to resource names
- Support multiple namespaces (EC2, RDS, Lambda, etc.)
- Handle dimension-based resource identification

#### CloudTrail Event Mapping
- Map CloudTrail events to resource names
- Support resource ARN parsing
- Handle resource name extraction from request parameters

### 1.4 Dependencies
- ClickHouse server (Docker or managed service)
- ClickHouse Go client: `github.com/ClickHouse/clickhouse-go/v2`
- AWS SDK for Go: `github.com/aws/aws-sdk-go-v2`
- CloudWatch client: `github.com/aws/aws-sdk-go-v2/service/cloudwatch`
- CloudTrail client: `github.com/aws/aws-sdk-go-v2/service/cloudtrail`

## Phase 2: Redis Cache Layer

### 2.1 Cache Strategy

#### Cache Keys Pattern
```
metrics:tenant:{tenant_id}:resource:{resource_id}:{metric_name}:{time_range}
cost:tenant:{tenant_id}:service:{service}:{date}
utilization:tenant:{tenant_id}:resource:{resource_id}:{time_range}
dashboard:tenant:{tenant_id}:summary
```

#### TTL Strategy
- Real-time metrics: 60 seconds
- Hourly aggregations: 3600 seconds (1 hour)
- Daily aggregations: 86400 seconds (24 hours)
- Dashboard summaries: 300 seconds (5 minutes)

### 2.2 Dependencies
- Redis server (Docker or managed service)
- Redis Go client: `github.com/redis/go-redis/v9`

## Phase 3: Implementation Files Needed

### Backend Files

1. **`internal/metrics/clickhouse_client.go`**
   - ClickHouse connection management
   - Insert methods for metrics
   - Query methods with tenant isolation

2. **`internal/metrics/redis_cache.go`**
   - Redis connection management
   - Cache get/set/delete methods
   - Cache key generation utilities

3. **`internal/metrics/metrics_service.go`**
   - Service layer for metrics operations
   - Orchestrates ClickHouse + Redis
   - Handles cache invalidation

4. **`internal/metrics/cloudwatch_collector.go`**
   - AWS CloudWatch metrics collection
   - Fetch utilization metrics (CPU, Memory, Network, Disk)
   - Map CloudWatch resources to yukti resources by name
   - Support multiple AWS accounts/regions
   - Handle pagination and rate limiting

5. **`internal/metrics/cloudtrail_collector.go`**
   - AWS CloudTrail log collection
   - Parse CloudTrail events
   - Extract resource names from events
   - Map CloudTrail resources to yukti resources
   - Handle S3 log file processing

6. **`internal/metrics/resource_mapper.go`**
   - Map CloudWatch/CloudTrail resource names to yukti resources
   - Support fuzzy matching and exact matching
   - Handle resource name normalization
   - Cache mappings in Redis
   - Background job to update mappings

7. **`internal/metrics/utilization_analyzer.go`**
   - Analyze utilization patterns
   - Detect under-utilized resources (CPU/Memory < threshold_low)
   - Detect over-utilized resources (CPU/Memory > threshold_high)
   - Generate recommendations (downsize, maintain, upsize)
   - Calculate confidence scores
   - Store analysis results in ClickHouse

8. **`internal/api/handlers/metrics.go`**
   - API endpoints for metrics
   - GET /api/v1/metrics/resource/:id
   - GET /api/v1/metrics/cost
   - GET /api/v1/metrics/utilization
   - GET /api/v1/metrics/analysis/:resource_id (utilization analysis)
   - GET /api/v1/audit/logs (CloudTrail audit logs)

9. **`internal/api/handlers/utilization.go`**
   - GET /api/v1/utilization/recommendations
   - GET /api/v1/utilization/under-utilized
   - GET /api/v1/utilization/over-utilized
   - POST /api/v1/utilization/analyze (trigger analysis)

10. **`internal/metrics/collector.go`**
    - Background job orchestrator
    - Coordinate CloudWatch and CloudTrail collectors
    - Handle scheduling and error recovery
    - Batch insert to ClickHouse
    - Trigger resource mapping updates

11. **`scripts/014_setup_clickhouse.sql`**
    - Database initialization script
    - Tables creation (including audit_logs, resource_mapping, utilization_analysis)
    - Materialized views setup

12. **`docker-compose.metrics.yml`**
    - ClickHouse service configuration
    - Redis service configuration
    - Networking setup

### Configuration Files

8. **`.env.example`** (additions)
   ```
   CLICKHOUSE_HOST=localhost
   CLICKHOUSE_PORT=9000
   CLICKHOUSE_DATABASE=yukti_metrics
   CLICKHOUSE_USER=default
   CLICKHOUSE_PASSWORD=
   
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0
   
   # AWS CloudWatch Configuration
   AWS_CLOUDWATCH_ENABLED=true
   AWS_CLOUDWATCH_REGIONS=us-east-1,us-west-2
   AWS_CLOUDWATCH_METRICS_INTERVAL=300  # 5 minutes
   AWS_CLOUDWATCH_NAMESPACES=AWS/EC2,AWS/RDS,AWS/Lambda
   
   # AWS CloudTrail Configuration
   AWS_CLOUDTRAIL_ENABLED=true
   AWS_CLOUDTRAIL_REGIONS=us-east-1,us-west-2
   AWS_CLOUDTRAIL_S3_BUCKET=cloudtrail-logs-bucket
   AWS_CLOUDTRAIL_LOG_PREFIX=CloudTrail
   
   # Utilization Analysis Configuration
   UTILIZATION_ANALYSIS_ENABLED=true
   UTILIZATION_ANALYSIS_INTERVAL=3600  # 1 hour
   UTILIZATION_CPU_LOW_THRESHOLD=20.0
   UTILIZATION_CPU_HIGH_THRESHOLD=80.0
   UTILIZATION_MEMORY_LOW_THRESHOLD=20.0
   UTILIZATION_MEMORY_HIGH_THRESHOLD=80.0
   ```

## Phase 4: Migration Strategy

1. **Setup Phase** (Week 1)
   - Install ClickHouse and Redis
   - Create database schema (including audit_logs, resource_mapping, utilization_analysis)
   - Test connectivity
   - Configure AWS credentials for CloudWatch/CloudTrail access

2. **Implementation Phase** (Week 2-3)
   - Implement metrics service layer
   - Implement CloudWatch collector
   - Implement CloudTrail collector
   - Implement resource mapper
   - Implement utilization analyzer
   - Add API endpoints
   - Implement cache layer
   - Add background collector orchestrator

3. **Integration Phase** (Week 4)
   - Test CloudWatch metric collection
   - Test CloudTrail log collection
   - Test resource name mapping
   - Test utilization analysis engine
   - Validate recommendations accuracy

4. **Testing Phase** (Week 5)
   - Load testing
   - Performance benchmarking
   - Cache hit rate monitoring
   - Utilization analysis accuracy testing
   - End-to-end integration testing

5. **Deployment Phase** (Week 6)
   - Deploy to staging
   - Gradual rollout
   - Monitor performance
   - Monitor analysis accuracy

## Performance Targets

- **Query Response Time**: < 200ms for cached queries
- **Cache Hit Rate**: > 80% for dashboard queries
- **ClickHouse Write Throughput**: > 10k metrics/second
- **Storage Efficiency**: 90-day retention with compression

## Monitoring & Observability

- ClickHouse query performance metrics
- Redis cache hit/miss rates
- Memory usage for both systems
- Write latency metrics
- Query latency percentiles (p50, p95, p99)

## Security Considerations

- Tenant isolation in ClickHouse queries
- Redis password authentication
- ClickHouse user-based access control
- Network encryption for both services
- Regular backup strategy for ClickHouse

