# ClickHouse Migration - Implementation Plan

**Trigger**: 50 customers OR 500K resources OR 10M metric data points  
**Timeline**: 5 weeks (35 days)  
**Team**: 2 backend engineers + 1 DevOps engineer  
**Risk Level**: Medium (dual-write strategy minimizes downtime)

---

## WEEK 1: INFRASTRUCTURE SETUP (Days 1-7)

### Day 1-2: ClickHouse Cluster Setup

**Tasks**:
1. Provision 3 ClickHouse nodes (HA cluster)
2. Configure replication and sharding
3. Set up monitoring (Prometheus + Grafana)
4. Configure backups (S3)

**Infrastructure** (AWS):
```yaml
# 3-node ClickHouse cluster
- Instance Type: c5.2xlarge (8 vCPU, 16GB RAM)
- Storage: 500GB EBS gp3 per node
- Network: VPC with private subnets
- Cost: ~$600/month
```

**ClickHouse Config**:
```xml
<!-- /etc/clickhouse-server/config.xml -->
<clickhouse>
    <listen_host>0.0.0.0</listen_host>
    <http_port>8123</http_port>
    <tcp_port>9000</tcp_port>
    
    <remote_servers>
        <yukti_cluster>
            <shard>
                <replica>
                    <host>clickhouse-1</host>
                    <port>9000</port>
                </replica>
                <replica>
                    <host>clickhouse-2</host>
                    <port>9000</port>
                </replica>
            </shard>
        </yukti_cluster>
    </remote_servers>
    
    <zookeeper>
        <node>
            <host>zookeeper-1</host>
            <port>2181</port>
        </node>
    </zookeeper>
</clickhouse>
```

**Deliverables**:
- ✅ 3-node ClickHouse cluster running
- ✅ Replication configured
- ✅ Monitoring dashboards
- ✅ Backup automation

---

### Day 3-4: Schema Design & Creation

**Create Tables**:

```sql
-- 1. CloudWatch Metrics
CREATE TABLE cloudwatch_metrics ON CLUSTER yukti_cluster (
    tenant_id UInt32,
    resource_id String,
    metric_name LowCardinality(String),
    timestamp DateTime,
    value Float64,
    unit LowCardinality(String),
    region LowCardinality(String),
    date Date MATERIALIZED toDate(timestamp)
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/cloudwatch_metrics', '{replica}')
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, resource_id, metric_name, timestamp)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- 2. AWS Cost Data
CREATE TABLE aws_cost_data ON CLUSTER yukti_cluster (
    tenant_id UInt32,
    date Date,
    hour DateTime,
    service LowCardinality(String),
    region LowCardinality(String),
    resource_id String,
    usage_type LowCardinality(String),
    cost Decimal(18, 6),
    usage_quantity Decimal(18, 6),
    pricing_unit LowCardinality(String)
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/aws_cost_data', '{replica}')
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, date, service, region)
TTL date + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- 3. Pricing History
CREATE TABLE pricing_history ON CLUSTER yukti_cluster (
    service LowCardinality(String),
    region LowCardinality(String),
    instance_type LowCardinality(String),
    pricing_model LowCardinality(String),
    price_per_hour Decimal(10, 6),
    effective_date Date
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/pricing_history', '{replica}', effective_date)
ORDER BY (service, region, instance_type, pricing_model);

-- 4. Usage Metrics
CREATE TABLE usage_metrics ON CLUSTER yukti_cluster (
    tenant_id UInt32,
    date Date,
    metric_type LowCardinality(String),
    source LowCardinality(String),
    destination LowCardinality(String),
    quantity Decimal(18, 6),
    unit LowCardinality(String)
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/usage_metrics', '{replica}')
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, date, metric_type)
TTL date + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
```

**Deliverables**:
- ✅ 4 tables created with replication
- ✅ Partitioning configured
- ✅ TTL policies set
- ✅ Indexes optimized

---

### Day 5-7: Go Client Integration

**Install ClickHouse Driver**:
```bash
go get github.com/ClickHouse/clickhouse-go/v2
```

**Create ClickHouse Client**:
```go
// internal/storage/clickhouse/client.go
package clickhouse

import (
    "context"
    "fmt"
    "github.com/ClickHouse/clickhouse-go/v2"
    "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var conn driver.Conn

func InitClickHouse() error {
    var err error
    conn, err = clickhouse.Open(&clickhouse.Options{
        Addr: []string{"clickhouse-1:9000", "clickhouse-2:9000", "clickhouse-3:9000"},
        Auth: clickhouse.Auth{
            Database: "yukti",
            Username: "default",
            Password: os.Getenv("CLICKHOUSE_PASSWORD"),
        },
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
        Compression: &clickhouse.Compression{
            Method: clickhouse.CompressionLZ4,
        },
    })
    return err
}

func GetConnection() driver.Conn {
    return conn
}
```

**Deliverables**:
- ✅ ClickHouse Go client configured
- ✅ Connection pooling
- ✅ Error handling
- ✅ Health checks

---

## WEEK 2: DUAL-WRITE IMPLEMENTATION (Days 8-14)

### Day 8-10: Metrics Writer

**Create Dual-Write Logic**:
```go
// internal/storage/metrics_writer.go
package storage

import (
    "context"
    "log"
    "yukti/internal/storage/clickhouse"
    "yukti/internal/storage/postgres"
)

type MetricsWriter struct {
    pgConn driver.Conn
    chConn driver.Conn
}

func (w *MetricsWriter) WriteMetrics(ctx context.Context, metrics []CloudWatchMetric) error {
    // Write to PostgreSQL (existing)
    if err := postgres.InsertMetrics(ctx, metrics); err != nil {
        return fmt.Errorf("postgres write failed: %w", err)
    }
    
    // Write to ClickHouse (new) - non-blocking
    go func() {
        if err := clickhouse.InsertMetrics(context.Background(), metrics); err != nil {
            log.Printf("[WARN] ClickHouse write failed: %v", err)
            // Don't fail the request
        }
    }()
    
    return nil
}
```

**ClickHouse Insert**:
```go
// internal/storage/clickhouse/metrics.go
package clickhouse

func InsertMetrics(ctx context.Context, metrics []CloudWatchMetric) error {
    batch, err := conn.PrepareBatch(ctx, "INSERT INTO cloudwatch_metrics")
    if err != nil {
        return err
    }
    
    for _, m := range metrics {
        err := batch.Append(
            m.TenantID,
            m.ResourceID,
            m.MetricName,
            m.Timestamp,
            m.Value,
            m.Unit,
            m.Region,
        )
        if err != nil {
            return err
        }
    }
    
    return batch.Send()
}
```

**Deliverables**:
- ✅ Dual-write for CloudWatch metrics
- ✅ Dual-write for cost data
- ✅ Dual-write for usage metrics
- ✅ Error handling (non-blocking)

---

### Day 11-14: Cost Data Writer

**Cost Data Dual-Write**:
```go
// internal/storage/cost_writer.go
package storage

func (w *CostWriter) WriteCostData(ctx context.Context, costs []AWSCostData) error {
    // Write to PostgreSQL
    if err := postgres.InsertCostData(ctx, costs); err != nil {
        return err
    }
    
    // Write to ClickHouse (async)
    go func() {
        if err := clickhouse.InsertCostData(context.Background(), costs); err != nil {
            log.Printf("[WARN] ClickHouse cost write failed: %v", err)
        }
    }()
    
    return nil
}
```

**Deliverables**:
- ✅ Cost data dual-write
- ✅ Pricing history dual-write
- ✅ Monitoring (write success rate)
- ✅ Alerting (if ClickHouse write fails >10%)

---

## WEEK 3: DATA VALIDATION (Days 15-21)

### Day 15-17: Historical Data Migration

**Backfill Script**:
```go
// scripts/backfill_clickhouse.go
package main

func main() {
    // Migrate last 90 days of metrics
    startDate := time.Now().AddDate(0, 0, -90)
    endDate := time.Now()
    
    for date := startDate; date.Before(endDate); date = date.AddDate(0, 0, 1) {
        log.Printf("Migrating data for %s", date.Format("2006-01-02"))
        
        // Fetch from PostgreSQL
        metrics, err := postgres.GetMetrics(date)
        if err != nil {
            log.Fatal(err)
        }
        
        // Insert to ClickHouse
        if err := clickhouse.InsertMetrics(context.Background(), metrics); err != nil {
            log.Fatal(err)
        }
        
        log.Printf("Migrated %d metrics", len(metrics))
    }
}
```

**Run Migration**:
```bash
# Migrate in batches (1 day at a time)
go run scripts/backfill_clickhouse.go
```

**Deliverables**:
- ✅ 90 days of metrics migrated
- ✅ 2 years of cost data migrated
- ✅ Pricing history migrated
- ✅ Data integrity verified

---

### Day 18-21: Query Comparison

**Validation Script**:
```go
// scripts/validate_data.go
package main

func validateMetrics(tenantID int, date time.Time) error {
    // Query PostgreSQL
    pgMetrics, err := postgres.GetMetrics(tenantID, date)
    if err != nil {
        return err
    }
    
    // Query ClickHouse
    chMetrics, err := clickhouse.GetMetrics(tenantID, date)
    if err != nil {
        return err
    }
    
    // Compare counts
    if len(pgMetrics) != len(chMetrics) {
        return fmt.Errorf("count mismatch: pg=%d ch=%d", len(pgMetrics), len(chMetrics))
    }
    
    // Compare aggregates
    pgAvg := calculateAvg(pgMetrics)
    chAvg := calculateAvg(chMetrics)
    
    if math.Abs(pgAvg-chAvg) > 0.01 {
        return fmt.Errorf("avg mismatch: pg=%.2f ch=%.2f", pgAvg, chAvg)
    }
    
    return nil
}
```

**Run Validation**:
```bash
# Validate all tenants
go run scripts/validate_data.go
```

**Deliverables**:
- ✅ Data consistency verified (100% match)
- ✅ Query results match (PostgreSQL vs ClickHouse)
- ✅ Performance benchmarks (10x faster confirmed)
- ✅ Compression ratio measured (50-100x)

---

## WEEK 4: SWITCH READS (Days 22-28)

### Day 22-24: Read Path Migration

**Feature Flag**:
```go
// internal/config/config.go
package config

var UseClickHouse = os.Getenv("USE_CLICKHOUSE") == "true"
```

**Query Router**:
```go
// internal/storage/metrics_reader.go
package storage

func GetMetrics(ctx context.Context, tenantID int, startDate, endDate time.Time) ([]CloudWatchMetric, error) {
    if config.UseClickHouse {
        return clickhouse.GetMetrics(ctx, tenantID, startDate, endDate)
    }
    return postgres.GetMetrics(ctx, tenantID, startDate, endDate)
}
```

**ClickHouse Query**:
```go
// internal/storage/clickhouse/metrics_reader.go
package clickhouse

func GetMetrics(ctx context.Context, tenantID int, startDate, endDate time.Time) ([]CloudWatchMetric, error) {
    query := `
        SELECT 
            tenant_id,
            resource_id,
            metric_name,
            timestamp,
            value,
            unit,
            region
        FROM cloudwatch_metrics
        WHERE tenant_id = ?
          AND date >= ?
          AND date <= ?
        ORDER BY timestamp DESC
    `
    
    rows, err := conn.Query(ctx, query, tenantID, startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var metrics []CloudWatchMetric
    for rows.Next() {
        var m CloudWatchMetric
        if err := rows.Scan(&m.TenantID, &m.ResourceID, &m.MetricName, &m.Timestamp, &m.Value, &m.Unit, &m.Region); err != nil {
            return nil, err
        }
        metrics = append(metrics, m)
    }
    
    return metrics, nil
}
```

**Deliverables**:
- ✅ Feature flag implemented
- ✅ All read queries migrated
- ✅ Fallback to PostgreSQL on error
- ✅ Monitoring (query latency, error rate)

---

### Day 25-28: Gradual Rollout

**Rollout Plan**:
```bash
# Day 25: Enable for 10% of tenants
USE_CLICKHOUSE=true CLICKHOUSE_ROLLOUT_PCT=10

# Day 26: Enable for 50% of tenants
USE_CLICKHOUSE=true CLICKHOUSE_ROLLOUT_PCT=50

# Day 27: Enable for 100% of tenants
USE_CLICKHOUSE=true CLICKHOUSE_ROLLOUT_PCT=100

# Day 28: Monitor and fix issues
```

**Rollout Logic**:
```go
func shouldUseClickHouse(tenantID int) bool {
    if !config.UseClickHouse {
        return false
    }
    
    rolloutPct := config.ClickHouseRolloutPct
    return (tenantID % 100) < rolloutPct
}
```

**Deliverables**:
- ✅ 10% rollout (Day 25)
- ✅ 50% rollout (Day 26)
- ✅ 100% rollout (Day 27)
- ✅ Zero downtime
- ✅ Performance improvement confirmed

---

## WEEK 5: DEPRECATE POSTGRESQL (Days 29-35)

### Day 29-31: Stop Dual-Write

**Remove Dual-Write**:
```go
// internal/storage/metrics_writer.go
func (w *MetricsWriter) WriteMetrics(ctx context.Context, metrics []CloudWatchMetric) error {
    // Only write to ClickHouse now
    return clickhouse.InsertMetrics(ctx, metrics)
}
```

**Deliverables**:
- ✅ Dual-write removed
- ✅ Only ClickHouse writes
- ✅ PostgreSQL writes stopped

---

### Day 32-34: Archive Old Data

**Archive Script**:
```bash
# Export PostgreSQL data to S3
pg_dump -t yt_cost_data yukti_finops | gzip > cost_data_archive.sql.gz
aws s3 cp cost_data_archive.sql.gz s3://yukti-archives/postgres/

# Drop old tables
DROP TABLE yt_cost_data;
DROP TABLE yt_cloudwatch_metrics;
```

**Deliverables**:
- ✅ Old data archived to S3
- ✅ Old tables dropped
- ✅ PostgreSQL storage freed (500GB → 100GB)

---

### Day 35: Documentation & Handoff

**Update Documentation**:
- Architecture diagrams
- Query examples
- Monitoring runbooks
- Rollback procedures

**Deliverables**:
- ✅ Migration complete
- ✅ Documentation updated
- ✅ Team trained
- ✅ Monitoring dashboards

---

## ROLLBACK PLAN

**If ClickHouse Fails**:
```bash
# Step 1: Disable ClickHouse reads
export USE_CLICKHOUSE=false

# Step 2: All queries go to PostgreSQL
# (Dual-write still active, so PostgreSQL has latest data)

# Step 3: Fix ClickHouse issue

# Step 4: Re-enable ClickHouse
export USE_CLICKHOUSE=true
```

**Rollback Window**: 2 weeks (while dual-write is active)

---

## MONITORING & ALERTS

**Metrics to Track**:
```yaml
# ClickHouse Health
- clickhouse_up (1 = healthy, 0 = down)
- clickhouse_query_latency_p95
- clickhouse_insert_errors_total
- clickhouse_disk_usage_pct

# Migration Progress
- dual_write_success_rate (target: >99%)
- data_consistency_check (target: 100%)
- query_performance_improvement (target: >5x)

# Business Metrics
- cost_per_query (target: <$0.001)
- storage_cost_savings (target: >50%)
```

**Alerts**:
```yaml
# Critical
- ClickHouse cluster down (page on-call)
- Dual-write success rate <95% (page on-call)
- Data consistency <99% (page on-call)

# Warning
- Query latency >1s (Slack alert)
- Disk usage >80% (Slack alert)
- Insert errors >1% (Slack alert)
```

---

## COST ANALYSIS

### Before Migration (PostgreSQL Only)
- **Storage**: 1TB = $100/month (RDS)
- **IOPS**: 10K IOPS = $650/month
- **Total**: $750/month

### After Migration (PostgreSQL + ClickHouse)
- **PostgreSQL**: 100GB = $10/month (relational data only)
- **ClickHouse**: 3x c5.2xlarge = $600/month
- **ClickHouse Storage**: 1TB compressed = $50/month
- **Total**: $660/month

**Savings**: $90/month (12%) + 10x faster queries

### At Scale (200 customers)
- **Before**: $15,000/month (PostgreSQL only)
- **After**: $2,200/month (PostgreSQL + ClickHouse + S3)
- **Savings**: $12,800/month (85%)

---

## SUCCESS CRITERIA

✅ Zero downtime during migration  
✅ 100% data consistency (PostgreSQL vs ClickHouse)  
✅ 10x query performance improvement  
✅ 50%+ storage cost reduction  
✅ <1% error rate during dual-write  
✅ All 50 customers migrated successfully  
✅ Rollback plan tested and documented  

---

## TEAM RESPONSIBILITIES

**Backend Engineer 1**:
- Dual-write implementation
- Query migration
- Data validation

**Backend Engineer 2**:
- ClickHouse schema design
- Go client integration
- Performance optimization

**DevOps Engineer**:
- ClickHouse cluster setup
- Monitoring and alerting
- Backup automation

---

**END OF PLAN**
