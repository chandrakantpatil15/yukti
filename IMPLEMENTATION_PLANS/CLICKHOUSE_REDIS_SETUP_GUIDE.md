# ClickHouse & Redis Setup Guide

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ for backend development
- PostgreSQL already running (existing setup)

## Step 1: Local Development Setup

### Using Docker Compose

1. **Create docker-compose.metrics.yml** (see prompts)
2. **Start services**:
   ```bash
   docker-compose -f docker-compose.metrics.yml up -d
   ```

3. **Verify services are running**:
   ```bash
   docker ps | grep -E "clickhouse|redis"
   ```

### Manual Installation (Alternative)

#### ClickHouse
```bash
# macOS
brew install clickhouse

# Linux
sudo apt-get install clickhouse-server clickhouse-client

# Start ClickHouse
sudo systemctl start clickhouse-server
```

#### Redis
```bash
# macOS
brew install redis

# Linux
sudo apt-get install redis-server

# Start Redis
sudo systemctl start redis-server
```

## Step 2: Database Initialization

### ClickHouse Setup

1. **Connect to ClickHouse**:
   ```bash
   clickhouse-client
   # Or via Docker:
   docker exec -it clickhouse clickhouse-client
   ```

2. **Create database**:
   ```sql
   CREATE DATABASE IF NOT EXISTS yukti_metrics;
   USE yukti_metrics;
   ```

3. **Run migration script**:
   ```bash
   clickhouse-client < scripts/014_setup_clickhouse.sql
   ```

4. **Verify tables created**:
   ```sql
   SHOW TABLES;
   ```

### Redis Setup

1. **Test Redis connection**:
   ```bash
   redis-cli ping
   # Should return: PONG
   ```

2. **Set password (optional, for production)**:
   ```bash
   redis-cli
   CONFIG SET requirepass "your-password"
   ```

## Step 3: Environment Configuration

Add to `.env` file:

```bash
# ClickHouse Configuration
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=yukti_metrics
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Step 4: Go Dependencies

Add to `go.mod` (run these commands):

```bash
go get github.com/ClickHouse/clickhouse-go/v2
go get github.com/redis/go-redis/v9
go mod tidy
```

## Step 5: Implementation Order

Follow these prompts in order (from AMAZON_Q_PROMPTS.md):

1. **Prompt 6**: Create ClickHouse database schema
2. **Prompt 7**: Create Docker Compose configuration
3. **Prompt 1**: Implement ClickHouse client
4. **Prompt 2**: Implement Redis cache
5. **Prompt 3**: Implement metrics service
6. **Prompt 4**: Implement API handlers
7. **Prompt 5**: Implement metrics collector
8. **Prompt 8**: Update environment variables
9. **Prompt 9**: Register routes
10. **Prompt 10**: Frontend API integration

## Step 6: Testing

### ClickHouse Tests

1. **Insert test data**:
   ```sql
   INSERT INTO yt_metrics (tenant_id, resource_id, metric_name, metric_value, timestamp)
   VALUES (1, 'test-resource-1', 'cpu_usage', 45.5, now());
   ```

2. **Query test data**:
   ```sql
   SELECT * FROM yt_metrics WHERE tenant_id = 1 LIMIT 10;
   ```

### Redis Tests

1. **Test cache operations**:
   ```bash
   redis-cli
   SET test:key "test value"
   GET test:key
   TTL test:key
   ```

2. **Test with TTL**:
   ```bash
   SETEX test:ttl 60 "expires in 60 seconds"
   TTL test:ttl
   ```

## Step 7: Performance Tuning

### ClickHouse Optimization

1. **Check table sizes**:
   ```sql
   SELECT 
       table,
       formatReadableSize(sum(bytes)) as size,
       sum(rows) as rows
   FROM system.parts
   WHERE database = 'yukti_metrics'
   GROUP BY table;
   ```

2. **Monitor queries**:
   ```sql
   SELECT * FROM system.query_log 
   WHERE type = 'QueryFinish' 
   ORDER BY query_duration_ms DESC 
   LIMIT 10;
   ```

3. **Optimize merges**:
   ```sql
   OPTIMIZE TABLE yt_metrics FINAL;
   ```

### Redis Optimization

1. **Monitor memory usage**:
   ```bash
   redis-cli INFO memory
   ```

2. **Check connected clients**:
   ```bash
   redis-cli CLIENT LIST
   ```

3. **Monitor commands**:
   ```bash
   redis-cli MONITOR
   ```

## Step 8: Production Considerations

### ClickHouse Production

1. **Replication Setup**:
   - Configure Zookeeper for replication
   - Use ReplicatedMergeTree engine
   - Setup multiple ClickHouse nodes

2. **Backup Strategy**:
   ```bash
   # Backup database
   clickhouse-client --query "BACKUP DATABASE yukti_metrics TO Disk('backups', 'backup_$(date +%Y%m%d)')"
   ```

3. **Monitoring**:
   - Use ClickHouse's system tables for monitoring
   - Integrate with Prometheus/Grafana
   - Set up alerts for disk space, query performance

### Redis Production

1. **Persistence**:
   - Enable AOF (Append Only File) for durability
   - Configure RDB snapshots
   - Setup replication for high availability

2. **Memory Management**:
   - Set maxmemory policy (eviction strategy)
   - Monitor memory usage
   - Plan for scaling (cluster mode)

3. **Security**:
   - Enable password authentication
   - Use SSL/TLS for connections
   - Restrict network access

## Troubleshooting

### ClickHouse Issues

**Connection refused**:
```bash
# Check if ClickHouse is running
sudo systemctl status clickhouse-server
# Check port
netstat -tulpn | grep 9000
```

**Out of memory errors**:
- Reduce max_memory_usage setting
- Optimize queries
- Increase server RAM

**Slow queries**:
- Check indexes
- Review ORDER BY clauses
- Use materialized views for aggregations

### Redis Issues

**Connection refused**:
```bash
# Check if Redis is running
redis-cli ping
# Check port
netstat -tulpn | grep 6379
```

**Memory errors**:
- Check maxmemory setting
- Review eviction policy
- Clear old keys

**Performance issues**:
- Monitor slow queries
- Use pipeline for bulk operations
- Consider Redis Cluster for scaling

## Useful Commands

### ClickHouse

```bash
# Start/Stop
sudo systemctl start clickhouse-server
sudo systemctl stop clickhouse-server

# View logs
sudo tail -f /var/log/clickhouse-server/clickhouse-server.log

# Client connection
clickhouse-client --host localhost --port 9000
```

### Redis

```bash
# Start/Stop
sudo systemctl start redis-server
sudo systemctl stop redis-server

# View logs
sudo tail -f /var/log/redis/redis-server.log

# Client connection
redis-cli -h localhost -p 6379

# Monitor commands in real-time
redis-cli MONITOR

# Get info
redis-cli INFO
redis-cli INFO stats
redis-cli INFO memory
```

## Next Steps

1. Complete implementation using Amazon Q prompts
2. Write unit tests for ClickHouse client
3. Write unit tests for Redis cache
4. Integration tests for metrics service
5. Load testing for performance validation
6. Documentation updates

