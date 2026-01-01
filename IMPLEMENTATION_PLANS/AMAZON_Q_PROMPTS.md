# Amazon Q Implementation Prompts

## Prompt 1: ClickHouse Client Implementation

```
You are implementing a ClickHouse client for a Go application. 

Create a new file: internal/metrics/clickhouse_client.go

Requirements:
1. Use github.com/ClickHouse/clickhouse-go/v2 client library
2. Create a ClickHouseClient struct with connection pooling
3. Environment variables:
   - CLICKHOUSE_HOST (default: localhost)
   - CLICKHOUSE_PORT (default: 9000)
   - CLICKHOUSE_DATABASE (default: yukti_metrics)
   - CLICKHOUSE_USER (default: default)
   - CLICKHOUSE_PASSWORD (default: empty)
4. Methods needed:
   - NewClickHouseClient() - constructor with connection setup
   - InsertMetric(tenantID, resourceID, metricName, value, tags) error
   - InsertCostMetric(tenantID, serviceName, resourceID, amount, currency) error
   - InsertUtilizationMetric(tenantID, resourceID, cpu, memory, network, disk) error
   - QueryMetrics(tenantID, resourceID, metricName, startTime, endTime) ([]MetricPoint, error)
   - QueryCostMetrics(tenantID, startDate, endDate, serviceName) ([]CostPoint, error)
   - Close() error
5. Ensure tenant isolation in all queries (always filter by tenant_id)
6. Use context.Context for timeouts
7. Handle connection errors gracefully
8. Log errors using log.Printf with [METRICS] prefix

Follow existing code patterns from internal/billing/stripe_client.go for structure.
```

## Prompt 2: Redis Cache Implementation

```
You are implementing a Redis cache layer for a Go application.

Create a new file: internal/metrics/redis_cache.go

Requirements:
1. Use github.com/redis/go-redis/v9 client library
2. Create a RedisCache struct with connection management
3. Environment variables:
   - REDIS_HOST (default: localhost)
   - REDIS_PORT (default: 6379)
   - REDIS_PASSWORD (default: empty)
   - REDIS_DB (default: 0)
4. Methods needed:
   - NewRedisCache() - constructor with connection setup
   - Get(key string) ([]byte, error) - get cached value
   - Set(key string, value []byte, ttl time.Duration) error - set cached value
   - Delete(key string) error - delete cached value
   - DeletePattern(pattern string) error - delete keys matching pattern
   - GenerateKey(parts ...string) string - generate cache key with ":" separator
   - Close() error
5. Cache key patterns:
   - metrics:tenant:{id}:resource:{id}:{metric}:{range}
   - cost:tenant:{id}:service:{service}:{date}
   - utilization:tenant:{id}:resource:{id}:{range}
   - dashboard:tenant:{id}:summary
6. Use JSON encoding for complex values
7. Handle connection errors and retries
8. Log errors using log.Printf with [CACHE] prefix

Follow existing code patterns from internal/billing/stripe_client.go for structure.
```

## Prompt 3: Metrics Service Layer

```
You are implementing a metrics service that orchestrates ClickHouse and Redis.

Create a new file: internal/metrics/metrics_service.go

Requirements:
1. Create a MetricsService struct with:
   - clickhouseClient *ClickHouseClient
   - redisCache *RedisCache
2. Methods needed:
   - GetResourceMetrics(tenantID, resourceID, metricName, timeRange) ([]MetricPoint, error)
     - Check Redis cache first
     - If cache miss, query ClickHouse
     - Cache results with appropriate TTL
   - GetCostMetrics(tenantID, startDate, endDate, serviceName) ([]CostPoint, error)
   - GetUtilizationMetrics(tenantID, resourceID, timeRange) ([]UtilizationPoint, error)
   - GetDashboardSummary(tenantID) (*DashboardSummary, error)
   - InvalidateCache(tenantID, pattern) error - invalidate cache for tenant
3. Cache TTL strategy:
   - Real-time (< 1 hour): 60 seconds
   - Hourly aggregations: 3600 seconds
   - Daily aggregations: 86400 seconds
   - Dashboard summaries: 300 seconds
4. Always enforce tenant isolation
5. Handle errors from both ClickHouse and Redis gracefully
6. Use context.Context for all operations

Follow existing service patterns from internal/services/invitation_service.go.
```

## Prompt 4: Metrics API Handlers (Updated with Utilization & Audit)

```
You are implementing API handlers for metrics, utilization, and audit endpoints.

Create a new file: internal/api/handlers/metrics.go

Requirements:
1. Create a MetricsHandler struct with:
   - metricsService *MetricsService
   - utilizationAnalyzer *UtilizationAnalyzer
   - resourceMapper *ResourceMapper
2. Metrics endpoints:
   - GET /api/v1/metrics/resource/:id
     Query params: metric_name, start_time, end_time, time_range
     Response: { success: true, data: []MetricPoint }
   - GET /api/v1/metrics/cost
     Query params: start_date, end_date, service_name, group_by
     Response: { success: true, data: []CostPoint, total: float64 }
   - GET /api/v1/metrics/utilization/:id
     Query params: start_time, end_time, time_range
     Response: { success: true, data: []UtilizationPoint }
   - GET /api/v1/metrics/dashboard
     Response: { success: true, data: DashboardSummary }
3. Utilization analysis endpoints:
   - GET /api/v1/utilization/recommendations
     Query params: time_range (1d, 7d, 30d)
     Response: { success: true, recommendations: []AnalysisResult }
   - GET /api/v1/utilization/under-utilized
     Query params: time_range, resource_type
     Response: { success: true, resources: []AnalysisResult }
   - GET /api/v1/utilization/over-utilized
     Query params: time_range, resource_type
     Response: { success: true, resources: []AnalysisResult }
   - GET /api/v1/utilization/analysis/:resource_id
     Query params: time_range
     Response: { success: true, analysis: AnalysisResult }
   - POST /api/v1/utilization/analyze
     Body: { resource_id, time_range }
     Response: { success: true, analysis_id: string }
4. Audit log endpoints:
   - GET /api/v1/audit/logs
     Query params: start_time, end_time, resource_name, event_name, user_identity
     Response: { success: true, logs: []AuditLog, total: int }
   - GET /api/v1/audit/logs/:resource_id
     Query params: start_time, end_time
     Response: { success: true, logs: []AuditLog }
5. Use JWT middleware for authentication
6. Extract tenant_id from JWT context
7. Validate query parameters
8. Handle errors appropriately (400, 500)
9. Return JSON responses
10. Support pagination for audit logs

Follow existing handler patterns from internal/api/handlers/billing.go.
Reference: internal/metrics/utilization_analyzer.go for analysis logic.
```

## Prompt 5: CloudWatch Metrics Collector

```
You are implementing an AWS CloudWatch metrics collector.

Create a new file: internal/metrics/cloudwatch_collector.go

Requirements:
1. Use AWS SDK v2: github.com/aws/aws-sdk-go-v2/service/cloudwatch
2. Create CloudWatchCollector struct with:
   - clickhouseClient *ClickHouseClient
   - awsConfig aws.Config
   - regions []string
   - namespaces []string (e.g., AWS/EC2, AWS/RDS, AWS/Lambda)
   - interval time.Duration
   - batchSize int
3. Methods needed:
   - NewCloudWatchCollector(config, clickhouseClient) *CloudWatchCollector
   - CollectMetrics(ctx context.Context, tenantID, accountID) error
   - collectEC2Metrics(ctx, tenantID, accountID, region) error
   - collectRDSMetrics(ctx, tenantID, accountID, region) error
   - collectLambdaMetrics(ctx, tenantID, accountID, region) error
   - mapResourceName(cloudwatchResource, resourceType) (yuktiResourceID, error)
4. Metrics to collect:
   - CPUUtilization (for EC2, RDS)
   - MemoryUtilization (for Lambda)
   - NetworkIn, NetworkOut
   - DiskReadOps, DiskWriteOps
   - RequestCount, ErrorRate (for Lambda)
5. Resource name mapping:
   - Extract resource name from CloudWatch dimensions
   - Map to yukti resources using resource_mapper service
   - Handle cases where mapping doesn't exist (log warning)
6. Batch inserts to ClickHouse:
   - Batch metrics before inserting (default: 1000 per batch)
   - Use INSERT INTO yt_utilization_metrics
   - Include: timestamp, tenant_id, resource_id, resource_name, metrics, region, account_id
7. Error handling:
   - Handle AWS API rate limits (exponential backoff)
   - Handle missing resources gracefully
   - Log errors with [CLOUDWATCH] prefix
8. Configuration from environment:
   - AWS_CLOUDWATCH_ENABLED
   - AWS_CLOUDWATCH_REGIONS (comma-separated)
   - AWS_CLOUDWATCH_METRICS_INTERVAL (seconds)
   - AWS_CLOUDWATCH_NAMESPACES (comma-separated)

Follow AWS SDK patterns from existing codebase if available.
Reference: internal/metrics/resource_mapper.go for resource mapping.
```

## Prompt 6: CloudTrail Audit Log Collector

```
You are implementing an AWS CloudTrail audit log collector.

Create a new file: internal/metrics/cloudtrail_collector.go

Requirements:
1. Use AWS SDK v2: github.com/aws/aws-sdk-go-v2/service/cloudtrail and service/s3
2. Create CloudTrailCollector struct with:
   - clickhouseClient *ClickHouseClient
   - s3Client *s3.Client
   - cloudtrailClient *cloudtrail.Client
   - s3Bucket string
   - logPrefix string
   - regions []string
   - interval time.Duration
3. Methods needed:
   - NewCloudTrailCollector(config, clickhouseClient) *CloudTrailCollector
   - CollectLogs(ctx context.Context, tenantID, accountID) error
   - listLogFiles(ctx, startTime, endTime) ([]LogFile, error)
   - downloadLogFile(ctx, logFile) ([]Event, error)
   - parseCloudTrailEvent(event) (*AuditLog, error)
   - extractResourceName(event) (resourceName, resourceType, error)
   - mapResourceName(cloudtrailResource) (yuktiResourceID, error)
4. CloudTrail event processing:
   - Download logs from S3 bucket
   - Parse JSON events from log files
   - Extract: eventName, eventSource, userIdentity, sourceIP, resourceName
   - Map resource names to yukti resources
   - Insert into yt_audit_logs table
5. Resource name extraction:
   - Extract from resourceName field
   - Extract from requestParameters (JSON parsing)
   - Extract from ARN format (parse ARN)
   - Handle multiple resource types (EC2, RDS, S3, Lambda, etc.)
6. Batch inserts to ClickHouse:
   - Batch events before inserting (default: 1000 per batch)
   - Use INSERT INTO yt_audit_logs
   - Include all CloudTrail fields
7. Log file tracking:
   - Track processed log files (use Redis or database)
   - Avoid re-processing same files
   - Handle incremental collection (only new files)
8. Error handling:
   - Handle S3 access errors
   - Handle malformed JSON gracefully
   - Log errors with [CLOUDTRAIL] prefix
9. Configuration from environment:
   - AWS_CLOUDTRAIL_ENABLED
   - AWS_CLOUDTRAIL_REGIONS
   - AWS_CLOUDTRAIL_S3_BUCKET
   - AWS_CLOUDTRAIL_LOG_PREFIX

Follow AWS SDK patterns from existing codebase.
Reference: internal/metrics/resource_mapper.go for resource mapping.
```

## Prompt 7: Resource Mapper Service

```
You are implementing a service to map CloudWatch/CloudTrail resource names to yukti resources.

Create a new file: internal/metrics/resource_mapper.go

Requirements:
1. Create ResourceMapper struct with:
   - db *sql.DB (PostgreSQL - yukti resources)
   - clickhouseClient *ClickHouseClient
   - redisCache *RedisCache
2. Methods needed:
   - NewResourceMapper(db, clickhouse, redis) *ResourceMapper
   - MapResourceName(tenantID, cloudwatchName, resourceType) (yuktiResourceID, error)
   - MapResourceFromCloudTrail(tenantID, cloudtrailResource, resourceType) (yuktiResourceID, error)
   - UpdateResourceMapping(tenantID, yuktiResourceID, cloudwatchName, cloudtrailName) error
   - GetResourceMapping(tenantID, resourceName) (*ResourceMapping, error)
   - SyncMappings(ctx context.Context) error - background sync job
3. Mapping strategy:
   - Exact match: resource_name = cloudwatch_name
   - Fuzzy match: Levenshtein distance < threshold
   - ARN parsing: Extract resource name from ARN
   - Cache mappings in Redis (TTL: 24 hours)
   - Store mappings in ClickHouse yt_resource_mapping table
4. Query yukti resources from PostgreSQL:
   - Query yt_tenant_resources table
   - Match by resource_id or resource_name
   - Consider tenant_id for isolation
5. Update ClickHouse mapping table:
   - INSERT INTO yt_resource_mapping
   - Handle conflicts (use ReplacingMergeTree)
   - Track mapping creation timestamp
6. Cache strategy:
   - Cache key: mapping:tenant:{id}:name:{name}:type:{type}
   - Cache value: yukti_resource_id
   - Cache TTL: 24 hours
7. Background sync job:
   - Periodically sync all resource mappings
   - Update ClickHouse table
   - Refresh Redis cache
   - Handle new resources added to yukti

Reference existing database models from internal/models/.
Follow service patterns from internal/services/invitation_service.go.
```

## Prompt 8: Utilization Analyzer Service

```
You are implementing a service to analyze resource utilization and detect under/over utilization.

Create a new file: internal/metrics/utilization_analyzer.go

Requirements:
1. Create UtilizationAnalyzer struct with:
   - clickhouseClient *ClickHouseClient
   - resourceMapper *ResourceMapper
   - cpuLowThreshold float64 (default: 20.0)
   - cpuHighThreshold float64 (default: 80.0)
   - memoryLowThreshold float64 (default: 20.0)
   - memoryHighThreshold float64 (default: 80.0)
2. Methods needed:
   - NewUtilizationAnalyzer(client, mapper, thresholds) *UtilizationAnalyzer
   - AnalyzeResource(ctx, tenantID, resourceID, startTime, endTime) (*AnalysisResult, error)
   - AnalyzeAllResources(ctx, tenantID, timeRange) ([]AnalysisResult, error)
   - GetUnderUtilizedResources(ctx, tenantID, timeRange) ([]AnalysisResult, error)
   - GetOverUtilizedResources(ctx, tenantID, timeRange) ([]AnalysisResult, error)
   - GenerateRecommendation(analysis) string
   - CalculateConfidenceScore(analysis) float64
3. Analysis logic:
   - Query yt_utilization_metrics from ClickHouse
   - Calculate: avg_cpu, avg_memory, max_cpu, max_memory over time range
   - Determine status:
     * Under-utilized: avg < low_threshold AND max < low_threshold
     * Over-utilized: avg > high_threshold OR max > high_threshold
     * Optimal: between thresholds
4. Recommendations:
   - Under-utilized: "downsize" or "consider smaller instance"
   - Over-utilized: "upsize" or "scale horizontally"
   - Optimal: "maintain current size"
   - Include specific suggestions based on metrics
5. Confidence score calculation:
   - Based on data points count (more = higher confidence)
   - Based on variance (low variance = higher confidence)
   - Range: 0.0 to 1.0
6. Store results:
   - INSERT INTO yt_utilization_analysis
   - Include all analysis fields
   - Link to resource_id and resource_name
7. Query optimization:
   - Use materialized views for aggregations
   - Cache analysis results in Redis (TTL: 1 hour)
   - Support different time ranges (1 day, 7 days, 30 days)
8. Configuration from environment:
   - UTILIZATION_ANALYSIS_ENABLED
   - UTILIZATION_CPU_LOW_THRESHOLD
   - UTILIZATION_CPU_HIGH_THRESHOLD
   - UTILIZATION_MEMORY_LOW_THRESHOLD
   - UTILIZATION_MEMORY_HIGH_THRESHOLD

Follow analysis patterns from existing recommendation logic if available.
Reference: internal/metrics/clickhouse_client.go for queries.
```

## Prompt 9: Metrics Collector Orchestrator

```
You are implementing a background job orchestrator to coordinate CloudWatch and CloudTrail collectors.

Update: internal/metrics/collector.go

Requirements:
1. Create MetricsCollector struct with:
   - cloudwatchCollector *CloudWatchCollector
   - cloudtrailCollector *CloudTrailCollector
   - resourceMapper *ResourceMapper
   - utilizationAnalyzer *UtilizationAnalyzer
   - clickhouseClient *ClickHouseClient
   - cloudwatchInterval time.Duration
   - cloudtrailInterval time.Duration
   - analysisInterval time.Duration
2. Methods needed:
   - NewMetricsCollector(config) *MetricsCollector
   - Start(ctx context.Context) error - start all collectors
   - Stop() error - stop all collectors
   - runCloudWatchCollector(ctx) error - goroutine for CloudWatch
   - runCloudTrailCollector(ctx) error - goroutine for CloudTrail
   - runResourceMapperSync(ctx) error - goroutine for mapping sync
   - runUtilizationAnalysis(ctx) error - goroutine for analysis
3. Orchestration:
   - Run CloudWatch collector every N minutes (default: 5)
   - Run CloudTrail collector every N minutes (default: 15)
   - Run resource mapper sync every N hours (default: 1)
   - Run utilization analysis every N hours (default: 1)
   - Use separate goroutines for each task
   - Use context for cancellation
4. Error handling:
   - Log errors but continue running
   - Retry failed operations with exponential backoff
   - Handle AWS rate limits gracefully
   - Track collector health
5. Initialization:
   - Initialize all sub-collectors
   - Validate AWS credentials
   - Test connections to ClickHouse/Redis
   - Start all goroutines

Initialize in cmd/server/main.go.
Follow background job patterns from existing codebase.
```

## Prompt 10: ClickHouse Database Schema (Updated)

```
You are creating a SQL migration script for ClickHouse database setup.

Create a new file: scripts/014_setup_clickhouse.sql

Requirements:
1. Create database: yukti_metrics
2. Create tables:
   - yt_metrics (timestamp, tenant_id, resource_id, metric_name, metric_value, tags, region, account_id)
   - yt_cost_metrics (timestamp, tenant_id, service_name, resource_id, cost_amount, currency, region, account_id, tags)
   - yt_utilization_metrics (timestamp, tenant_id, resource_id, resource_type, cpu_usage, memory_usage, network_in, network_out, disk_read, disk_write)
3. Use MergeTree engine
4. Partition by month (toYYYYMM(timestamp))
5. Order by (tenant_id, resource_id, metric_name, timestamp)
6. Add TTL: 90 days for metrics/utilization, 365 days for cost
7. Add indexes for tenant_id, resource_id, metric_name
8. Create materialized view for daily aggregations:
   - yt_metrics_daily (SummingMergeTree)
   - Aggregates: sum, avg, max, min by day

Reference IMPLEMENTATION_PLANS/METRICS_CLICKHOUSE_IMPLEMENTATION.md for detailed schema.
```

## Prompt 11: Docker Compose Configuration

```
You are creating a Docker Compose configuration for ClickHouse and Redis services.

Create a new file: docker-compose.metrics.yml

Requirements:
1. ClickHouse service:
   - Image: clickhouse/clickhouse-server:latest
   - Ports: 8123 (HTTP), 9000 (Native)
   - Volumes: ./clickhouse-data:/var/lib/clickhouse
   - Environment: CLICKHOUSE_DB=yukti_metrics
   - Health check
2. Redis service:
   - Image: redis:7-alpine
   - Port: 6379
   - Volumes: ./redis-data:/data
   - Command: redis-server --appendonly yes
   - Health check
3. Network: create metrics_network
4. Both services should be on same network
5. Add restart policies
6. Set resource limits

Reference docker-compose.yml for existing patterns.
```

## Prompt 12: Environment Variables Update (AWS Integration)

```
Update the .env.example file to include ClickHouse and Redis configuration.

Add these variables:
- CLICKHOUSE_HOST=localhost
- CLICKHOUSE_PORT=9000
- CLICKHOUSE_DATABASE=yukti_metrics
- CLICKHOUSE_USER=default
- CLICKHOUSE_PASSWORD=
- REDIS_HOST=localhost
- REDIS_PORT=6379
- REDIS_PASSWORD=
- REDIS_DB=0

Add to existing .env.example file.
Keep existing variables intact.
```

## Prompt 13: Routes Registration (Updated with Utilization APIs)

```
Add metrics routes to the router configuration.

Update: internal/api/routes/routes.go

Requirements:
1. Initialize MetricsHandler with metricsService
2. Add routes (all protected with JWT middleware):
   - GET /api/v1/metrics/resource/:id - MetricsHandler.GetResourceMetrics
   - GET /api/v1/metrics/cost - MetricsHandler.GetCostMetrics
   - GET /api/v1/metrics/utilization/:id - MetricsHandler.GetUtilizationMetrics
   - GET /api/v1/metrics/dashboard - MetricsHandler.GetDashboardSummary
3. Add rate limiting middleware
4. Use tenant isolation middleware if available

Follow existing route patterns from billing routes.
```

## Prompt 14: Frontend API Integration (Updated)

```
Add metrics API methods to the frontend API service.

Update: frontend/src/services/api.ts

Requirements:
1. Add methods:
   - getResourceMetrics(resourceId, metricName, timeRange) - GET /api/v1/metrics/resource/:id
   - getCostMetrics(startDate, endDate, serviceName, groupBy) - GET /api/v1/metrics/cost
   - getUtilizationMetrics(resourceId, timeRange) - GET /api/v1/metrics/utilization/:id
   - getDashboardMetrics() - GET /api/v1/metrics/dashboard
2. Handle query parameters properly
3. Return typed responses
4. Include error handling

Follow existing API method patterns in api.ts.
```

