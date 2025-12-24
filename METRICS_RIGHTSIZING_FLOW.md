# Metrics-Based Right-Sizing - Complete Data Flow

**Question**: Do we need to store metrics data (CloudWatch) in InfluxDB/ClickHouse for analysis before recommending scale in/out?

**Answer**: **YES** - We must store CloudWatch metrics in ClickHouse to analyze utilization patterns over time (7-90 days) before making right-sizing recommendations.

---

## COMPLETE DATA FLOW

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         YUKTI FINOPS DATA FLOW                               │
└─────────────────────────────────────────────────────────────────────────────┘

Step 1: COLLECT METRICS (Every 5 minutes)
┌──────────────────────────────────────────────────────────────┐
│  AWS CloudWatch                                               │
│  ├─ EC2 Instance i-0a046ebb489ff3cd7                         │
│  │  ├─ CPUUtilization: 8.5% (timestamp: 2024-01-15 10:00)   │
│  │  ├─ NetworkIn: 1.2 MB (timestamp: 2024-01-15 10:00)      │
│  │  ├─ NetworkOut: 0.8 MB (timestamp: 2024-01-15 10:00)     │
│  │  └─ DiskReadOps: 50 (timestamp: 2024-01-15 10:00)        │
│  └─ RDS Instance db-mysql-prod                               │
│     ├─ CPUUtilization: 45% (timestamp: 2024-01-15 10:00)    │
│     ├─ DatabaseConnections: 120 (timestamp: 2024-01-15 10:00)│
│     └─ ReadIOPS: 500 (timestamp: 2024-01-15 10:00)          │
└──────────────────────────────────────────────────────────────┘
                              ↓
Step 2: STORE IN CLICKHOUSE (Time-series database)
┌──────────────────────────────────────────────────────────────┐
│  ClickHouse Table: cloudwatch_metrics                        │
│  ┌────────────┬──────────────┬──────────┬───────┬──────┐   │
│  │ tenant_id  │ resource_id  │ metric   │ value │ time │   │
│  ├────────────┼──────────────┼──────────┼───────┼──────┤   │
│  │ 27         │ i-0a046ebb.. │ CPU      │ 8.5   │ 10:00│   │
│  │ 27         │ i-0a046ebb.. │ CPU      │ 7.2   │ 10:05│   │
│  │ 27         │ i-0a046ebb.. │ CPU      │ 9.1   │ 10:10│   │
│  │ ...        │ ...          │ ...      │ ...   │ ...  │   │
│  │ 27         │ i-0a046ebb.. │ CPU      │ 8.8   │ 17:00│   │
│  └────────────┴──────────────┴──────────┴───────┴──────┘   │
│  Retention: 90 days (2,592,000 data points per resource)    │
└──────────────────────────────────────────────────────────────┘
                              ↓
Step 3: ANALYZE UTILIZATION (Every 24 hours)
┌──────────────────────────────────────────────────────────────┐
│  Right-Sizing Analyzer (Go service)                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Query ClickHouse:                                      │ │
│  │ SELECT                                                 │ │
│  │   resource_id,                                         │ │
│  │   avg(value) as avg_cpu,                              │ │
│  │   max(value) as max_cpu,                              │ │
│  │   min(value) as min_cpu,                              │ │
│  │   percentile(value, 0.95) as p95_cpu                  │ │
│  │ FROM cloudwatch_metrics                                │ │
│  │ WHERE tenant_id = 27                                   │ │
│  │   AND metric_name = 'CPUUtilization'                  │ │
│  │   AND timestamp >= now() - INTERVAL 7 DAY             │ │
│  │ GROUP BY resource_id;                                  │ │
│  │                                                        │ │
│  │ Result:                                                │ │
│  │ i-0a046ebb489ff3cd7: avg=8.5%, max=15%, p95=12%      │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
                              ↓
Step 4: GENERATE RECOMMENDATIONS
┌──────────────────────────────────────────────────────────────┐
│  Recommendation Engine                                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ IF avg_cpu < 10% AND max_cpu < 20% FOR 7 days:       │ │
│  │   → Recommendation: Scale DOWN (t3.large → t3.small)  │ │
│  │   → Savings: $31/month                                 │ │
│  │   → Confidence: 95%                                    │ │
│  │                                                        │ │
│  │ IF avg_cpu > 70% AND p95_cpu > 85% FOR 7 days:       │ │
│  │   → Recommendation: Scale UP (t3.large → t3.xlarge)   │ │
│  │   → Risk: Performance degradation                      │ │
│  │   → Confidence: 90%                                    │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
                              ↓
Step 5: STORE RECOMMENDATIONS (PostgreSQL)
┌──────────────────────────────────────────────────────────────┐
│  PostgreSQL Table: yt_right_sizing_recommendations           │
│  ┌────────┬──────────┬─────────┬────────┬─────────┬──────┐ │
│  │ tenant │ resource │ current │ recom. │ savings │ conf │ │
│  ├────────┼──────────┼─────────┼────────┼─────────┼──────┤ │
│  │ 27     │ i-0a04.. │ t3.large│t3.small│ $31/mo  │ 95%  │ │
│  │ 27     │ i-0b15.. │ t3.xlrg │t3.2xlrg│ -$62/mo │ 90%  │ │
│  └────────┴──────────┴─────────┴────────┴─────────┴──────┘ │
└──────────────────────────────────────────────────────────────┘
                              ↓
Step 6: DISPLAY IN UI
┌──────────────────────────────────────────────────────────────┐
│  Yukti Dashboard - Resource Utilization Page                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Right-Sizing Recommendations                           │ │
│  │ ┌──────────────────────────────────────────────────┐  │ │
│  │ │ i-0a046ebb489ff3cd7                              │  │ │
│  │ │ Current: t3.large → Recommended: t3.small        │  │ │
│  │ │ Reason: Avg CPU 8.5%, Max CPU 15% (7 days)      │  │ │
│  │ │ Savings: $31/month                                │  │ │
│  │ │ [Apply] [Dismiss] [View Details]                 │  │ │
│  │ └──────────────────────────────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

---

## WHY WE NEED CLICKHOUSE FOR METRICS

### Problem with PostgreSQL

**Scenario**: Analyze CPU utilization for 1,000 EC2 instances over 7 days

**Data Volume**:
- 1,000 instances × 7 days × 24 hours × 12 samples/hour = **2,016,000 data points**
- Each data point: 50 bytes
- Total: **100 MB per query**

**PostgreSQL**:
```sql
-- Query: Get average CPU for last 7 days
SELECT 
    resource_id,
    AVG(cpu_utilization) as avg_cpu
FROM cloudwatch_metrics
WHERE tenant_id = 27
  AND timestamp >= NOW() - INTERVAL '7 days'
GROUP BY resource_id;

-- Execution time: 30-60 seconds (row-based scan)
-- Problem: Too slow for real-time recommendations
```

**ClickHouse**:
```sql
-- Same query
SELECT 
    resource_id,
    avg(value) as avg_cpu
FROM cloudwatch_metrics
WHERE tenant_id = 27
  AND timestamp >= now() - INTERVAL 7 DAY
  AND metric_name = 'CPUUtilization'
GROUP BY resource_id;

-- Execution time: 0.5-1 second (columnar scan)
-- Solution: Fast enough for real-time recommendations
```

**Verdict**: ClickHouse is **30-60x faster** for metrics analysis.

---

## DATA STORAGE STRATEGY

### What Goes Where

| **Data Type** | **Database** | **Reason** | **Retention** |
|---------------|--------------|------------|---------------|
| **CloudWatch Metrics** | ClickHouse | High volume, time-series, aggregations | 90 days |
| **AWS Cost Data** | ClickHouse | High volume, time-series, analytics | 2 years |
| **Resource Inventory** | PostgreSQL | Relational, low volume, CRUD | Forever |
| **Recommendations** | PostgreSQL | Relational, low volume, CRUD | Forever |
| **Users/Tenants** | PostgreSQL | Relational, ACID, low volume | Forever |

### Why This Split?

**ClickHouse** (Time-Series Data):
- ✅ CloudWatch metrics (CPU, memory, network) - **millions of data points**
- ✅ AWS cost data (hourly costs) - **millions of records**
- ✅ Fast aggregations (AVG, MAX, MIN, percentiles)
- ✅ 90-day retention (auto-delete old data)

**PostgreSQL** (Relational Data):
- ✅ Resource inventory (EC2, RDS, S3) - **thousands of records**
- ✅ Recommendations (right-sizing, RI/SP) - **hundreds of records**
- ✅ Users, tenants, roles - **hundreds of records**
- ✅ Complex joins (resources + recommendations + users)

---

## RIGHT-SIZING ALGORITHM

### Step 1: Collect Metrics (Every 5 minutes)

```go
// internal/scanner/cloudwatch_collector.go
func CollectMetrics(resourceID string) error {
    // Fetch from CloudWatch
    metrics, err := cloudwatch.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
        Namespace:  aws.String("AWS/EC2"),
        MetricName: aws.String("CPUUtilization"),
        Dimensions: []*cloudwatch.Dimension{{
            Name:  aws.String("InstanceId"),
            Value: aws.String(resourceID),
        }},
        StartTime:  aws.Time(time.Now().Add(-5 * time.Minute)),
        EndTime:    aws.Time(time.Now()),
        Period:     aws.Int64(300), // 5 minutes
        Statistics: []*string{aws.String("Average")},
    })
    
    // Store in ClickHouse
    for _, datapoint := range metrics.Datapoints {
        clickhouse.InsertMetric(MetricData{
            TenantID:   tenantID,
            ResourceID: resourceID,
            MetricName: "CPUUtilization",
            Value:      *datapoint.Average,
            Timestamp:  *datapoint.Timestamp,
        })
    }
    
    return nil
}
```

### Step 2: Analyze Utilization (Every 24 hours)

```go
// internal/analyzer/right_sizing.go
func AnalyzeUtilization(tenantID int, resourceID string) (*Recommendation, error) {
    // Query ClickHouse for 7-day metrics
    query := `
        SELECT 
            avg(value) as avg_cpu,
            max(value) as max_cpu,
            min(value) as min_cpu,
            quantile(0.95)(value) as p95_cpu,
            quantile(0.99)(value) as p99_cpu
        FROM cloudwatch_metrics
        WHERE tenant_id = ?
          AND resource_id = ?
          AND metric_name = 'CPUUtilization'
          AND timestamp >= now() - INTERVAL 7 DAY
    `
    
    var stats struct {
        AvgCPU float64
        MaxCPU float64
        MinCPU float64
        P95CPU float64
        P99CPU float64
    }
    
    err := clickhouse.QueryRow(query, tenantID, resourceID).Scan(
        &stats.AvgCPU, &stats.MaxCPU, &stats.MinCPU, &stats.P95CPU, &stats.P99CPU,
    )
    
    // Generate recommendation
    return generateRecommendation(resourceID, stats)
}
```

### Step 3: Generate Recommendation

```go
func generateRecommendation(resourceID string, stats UtilizationStats) (*Recommendation, error) {
    currentType := getCurrentInstanceType(resourceID) // e.g., "t3.large"
    
    // Rule 1: Scale DOWN if underutilized
    if stats.AvgCPU < 10 && stats.MaxCPU < 20 {
        recommendedType := scaleDown(currentType) // "t3.small"
        savings := calculateSavings(currentType, recommendedType)
        
        return &Recommendation{
            ResourceID:      resourceID,
            CurrentType:     currentType,
            RecommendedType: recommendedType,
            Reason:          fmt.Sprintf("Avg CPU %.1f%%, Max CPU %.1f%% (7 days)", stats.AvgCPU, stats.MaxCPU),
            MonthlySavings:  savings,
            Confidence:      95,
            Action:          "scale_down",
        }, nil
    }
    
    // Rule 2: Scale UP if overutilized
    if stats.AvgCPU > 70 && stats.P95CPU > 85 {
        recommendedType := scaleUp(currentType) // "t3.xlarge"
        additionalCost := calculateSavings(currentType, recommendedType) // negative
        
        return &Recommendation{
            ResourceID:      resourceID,
            CurrentType:     currentType,
            RecommendedType: recommendedType,
            Reason:          fmt.Sprintf("Avg CPU %.1f%%, P95 CPU %.1f%% (7 days)", stats.AvgCPU, stats.P95CPU),
            MonthlySavings:  additionalCost,
            Confidence:      90,
            Action:          "scale_up",
        }, nil
    }
    
    // Rule 3: No change needed
    return &Recommendation{
        ResourceID:      resourceID,
        CurrentType:     currentType,
        RecommendedType: currentType,
        Reason:          "Utilization is optimal",
        MonthlySavings:  0,
        Confidence:      100,
        Action:          "no_change",
    }, nil
}
```

### Step 4: Store Recommendation (PostgreSQL)

```go
// Store in PostgreSQL (relational data)
func StoreRecommendation(rec *Recommendation) error {
    query := `
        INSERT INTO yt_right_sizing_recommendations 
        (tenant_id, resource_id, current_type, recommended_type, reason, monthly_savings, confidence, action, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
        ON CONFLICT (tenant_id, resource_id) 
        DO UPDATE SET 
            current_type = EXCLUDED.current_type,
            recommended_type = EXCLUDED.recommended_type,
            reason = EXCLUDED.reason,
            monthly_savings = EXCLUDED.monthly_savings,
            confidence = EXCLUDED.confidence,
            action = EXCLUDED.action,
            updated_at = NOW()
    `
    
    _, err := postgres.Exec(query, 
        rec.TenantID, rec.ResourceID, rec.CurrentType, rec.RecommendedType, 
        rec.Reason, rec.MonthlySavings, rec.Confidence, rec.Action,
    )
    
    return err
}
```

---

## METRICS COLLECTION FREQUENCY

### Recommended Schedule

| **Metric Type** | **Collection Frequency** | **Retention** | **Storage** |
|-----------------|-------------------------|---------------|-------------|
| **CPU Utilization** | Every 5 minutes | 90 days | ClickHouse |
| **Memory Utilization** | Every 5 minutes | 90 days | ClickHouse |
| **Network I/O** | Every 5 minutes | 90 days | ClickHouse |
| **Disk I/O** | Every 5 minutes | 90 days | ClickHouse |
| **Cost Data** | Every 24 hours | 2 years | ClickHouse |
| **Resource Inventory** | Every 24 hours | Forever | PostgreSQL |
| **Recommendations** | Every 24 hours | Forever | PostgreSQL |

### Data Volume Calculation

**Scenario**: 1,000 EC2 instances, 4 metrics per instance

**Per Day**:
- 1,000 instances × 4 metrics × 24 hours × 12 samples/hour = **1,152,000 data points/day**
- Each data point: 50 bytes
- Total: **57.6 MB/day**

**Per 90 Days** (retention period):
- 57.6 MB/day × 90 days = **5.2 GB**
- With 100:1 compression: **52 MB** (ClickHouse compressed)

**Verdict**: ClickHouse can easily handle this volume.

---

## RECOMMENDATION CONFIDENCE LEVELS

### How We Calculate Confidence

```go
func calculateConfidence(stats UtilizationStats, days int) int {
    confidence := 50 // Base confidence
    
    // More data = higher confidence
    if days >= 30 {
        confidence += 30
    } else if days >= 14 {
        confidence += 20
    } else if days >= 7 {
        confidence += 10
    }
    
    // Consistent pattern = higher confidence
    variance := calculateVariance(stats)
    if variance < 5 {
        confidence += 20 // Very consistent
    } else if variance < 10 {
        confidence += 10 // Somewhat consistent
    }
    
    // Clear signal = higher confidence
    if stats.AvgCPU < 5 || stats.AvgCPU > 80 {
        confidence += 10 // Very clear signal
    }
    
    return min(confidence, 100)
}
```

### Confidence Levels

| **Confidence** | **Meaning** | **Action** |
|----------------|-------------|------------|
| **90-100%** | Very high confidence | Auto-apply (with approval) |
| **70-89%** | High confidence | Recommend (manual review) |
| **50-69%** | Medium confidence | Suggest (needs validation) |
| **<50%** | Low confidence | Don't show (insufficient data) |

---

## FINAL ANSWER

### Do We Need ClickHouse for Metrics?

**YES** - We must store CloudWatch metrics in ClickHouse because:

1. **Volume**: Millions of data points (5-minute intervals × 90 days)
2. **Speed**: 30-60x faster queries (0.5s vs 30s)
3. **Retention**: 90 days of granular data (vs 30 days in PostgreSQL)
4. **Aggregations**: AVG, MAX, MIN, percentiles (optimized for time-series)
5. **Compression**: 100:1 ratio (5.2GB → 52MB)

### Data Flow Summary

1. **Collect** CloudWatch metrics every 5 minutes → Store in **ClickHouse**
2. **Analyze** utilization patterns over 7-90 days → Query **ClickHouse**
3. **Generate** right-sizing recommendations → Store in **PostgreSQL**
4. **Display** recommendations in UI → Fetch from **PostgreSQL**

### Why Not PostgreSQL for Metrics?

- ❌ Too slow (30-60 seconds per query)
- ❌ Too expensive (no compression)
- ❌ Not optimized for time-series data
- ❌ Can't handle millions of data points

### Why ClickHouse?

- ✅ Fast (0.5-1 second per query)
- ✅ Cheap (100:1 compression)
- ✅ Optimized for time-series data
- ✅ Handles billions of data points

**Conclusion**: ClickHouse is **essential** for storing CloudWatch metrics to enable accurate right-sizing recommendations.

---

**END OF DOCUMENT**
