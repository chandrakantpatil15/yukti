# Redis Deployment Guide - Quick Start

## 🚀 Deploy Redis (5 minutes)

### Step 1: Start Redis
```bash
cd /Users/chandrakantpatil/workspace/yukti

# Start Redis service
docker-compose up -d redis

# Wait for Redis to be ready
sleep 3

# Check Redis status
docker ps | grep redis
```

### Step 2: Verify Redis Connection
```bash
# Test connection
docker exec -it yukti-redis redis-cli -a yukti123 ping
# Expected output: PONG

# Check Redis info
docker exec -it yukti-redis redis-cli -a yukti123 INFO server
```

### Step 3: Run Test Script
```bash
# Run comprehensive tests
./test-redis.sh

# Expected: All tests pass ✅
```

### Step 4: Restart Backend with Redis
```bash
# Rebuild backend with Redis support
docker-compose up -d --build backend

# Check backend logs
docker logs yukti-backend | grep Redis
# Expected: "[INFO] Redis connection established"
```

### Step 5: Verify Integration
```bash
# Check backend can connect to Redis
docker logs yukti-backend | tail -20

# Should see:
# [INFO] Connecting to Redis...
# [INFO] Redis connection established at redis:6379
# [INFO] Initializing cache services...
# [INFO] Cache services initialized
```

---

## 📊 Monitor Redis

### Real-time Monitoring
```bash
# Watch all Redis commands
docker exec -it yukti-redis redis-cli -a yukti123 monitor

# In another terminal, make API calls and watch Redis activity
```

### Check Cache Keys
```bash
# List all keys
docker exec -it yukti-redis redis-cli -a yukti123 KEYS "*"

# List session keys
docker exec -it yukti-redis redis-cli -a yukti123 KEYS "session:*"

# List dashboard keys
docker exec -it yukti-redis redis-cli -a yukti123 KEYS "dashboard:*"

# List OTP keys
docker exec -it yukti-redis redis-cli -a yukti123 KEYS "otp:*"
```

### Check Memory Usage
```bash
# Memory stats
docker exec -it yukti-redis redis-cli -a yukti123 INFO memory

# Key count
docker exec -it yukti-redis redis-cli -a yukti123 DBSIZE
```

---

## 🧪 Test Cache Operations

### Test Session Cache
```bash
# Set session
docker exec -it yukti-redis redis-cli -a yukti123 SET "session:test-user" '{"user_id":"123","email":"test@yukti.com"}' EX 86400

# Get session
docker exec -it yukti-redis redis-cli -a yukti123 GET "session:test-user"

# Check TTL
docker exec -it yukti-redis redis-cli -a yukti123 TTL "session:test-user"
```

### Test OTP Cache
```bash
# Set OTP
docker exec -it yukti-redis redis-cli -a yukti123 SET "otp:test@yukti.com" "123456" EX 600

# Get OTP
docker exec -it yukti-redis redis-cli -a yukti123 GET "otp:test@yukti.com"

# Wait 10 minutes and check (should be expired)
docker exec -it yukti-redis redis-cli -a yukti123 GET "otp:test@yukti.com"
```

### Test Dashboard Cache
```bash
# Set dashboard data
docker exec -it yukti-redis redis-cli -a yukti123 SET "dashboard:tenant:18" '{"total_cost":12450,"savings":425.60}' EX 300

# Get dashboard data
docker exec -it yukti-redis redis-cli -a yukti123 GET "dashboard:tenant:18"

# Invalidate cache
docker exec -it yukti-redis redis-cli -a yukti123 DEL "dashboard:tenant:18"
```

---

## 🔧 Troubleshooting

### Redis not starting
```bash
# Check logs
docker logs yukti-redis

# Common issues:
# 1. Port 6379 already in use
# 2. Permission issues with volume
# 3. Memory limit reached

# Solution: Stop other Redis instances
docker stop $(docker ps -q --filter ancestor=redis)
```

### Connection refused
```bash
# Check if Redis is running
docker ps | grep redis

# Check network
docker network inspect yukti_default

# Restart Redis
docker-compose restart redis
```

### Authentication failed
```bash
# Check password in .env.ports
cat .env.ports | grep REDIS_PASSWORD

# Test with correct password
docker exec -it yukti-redis redis-cli -a yukti123 ping
```

### Backend can't connect to Redis
```bash
# Check backend environment variables
docker exec yukti-backend env | grep REDIS

# Should show:
# REDIS_HOST=redis
# REDIS_PORT=6379
# REDIS_PASSWORD=yukti123

# If missing, rebuild backend
docker-compose up -d --build backend
```

---

## 🎯 Performance Testing

### Before Redis (Baseline)
```bash
# Test dashboard load time
time curl -X GET http://localhost:8081/api/customers/dashboard \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected: 3-5 seconds
```

### After Redis (Target)
```bash
# First request (cache miss)
time curl -X GET http://localhost:8081/api/customers/dashboard \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
# Expected: 3-5 seconds

# Second request (cache hit)
time curl -X GET http://localhost:8081/api/customers/dashboard \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
# Expected: 300ms (10x faster!)
```

---

## 📈 Success Metrics

### Cache Hit Rate
```bash
# Check Redis stats
docker exec -it yukti-redis redis-cli -a yukti123 INFO stats | grep keyspace

# Target: >95% hit rate after warm-up
```

### Response Times
- Login: 200ms → 50ms (4x faster)
- Dashboard: 3-5s → 300ms (10x faster)
- Resources: 1-2s → 200ms (5-10x faster)

### Memory Usage
```bash
# Check memory
docker exec -it yukti-redis redis-cli -a yukti123 INFO memory | grep used_memory_human

# Expected: <100MB for 1,000 users
```

---

## 🔄 Next Steps

1. **Integrate Session Cache** (JWT middleware)
2. **Integrate OTP Cache** (Auth handler)
3. **Integrate Dashboard Cache** (Customer handler)
4. **Add Cache Invalidation** (Scanner)
5. **Load Testing** (100 concurrent users)

---

## 📝 Quick Commands Reference

```bash
# Start Redis
docker-compose up -d redis

# Stop Redis
docker-compose stop redis

# Restart Redis
docker-compose restart redis

# View logs
docker logs -f yukti-redis

# Connect to Redis CLI
docker exec -it yukti-redis redis-cli -a yukti123

# Flush all data (dev only!)
docker exec -it yukti-redis redis-cli -a yukti123 FLUSHALL

# Check health
docker exec -it yukti-redis redis-cli -a yukti123 ping

# Monitor commands
docker exec -it yukti-redis redis-cli -a yukti123 monitor

# Get all keys
docker exec -it yukti-redis redis-cli -a yukti123 KEYS "*"
```

---

**Status**: Redis Infrastructure Ready ✅  
**Performance**: 10x faster (target)  
**Next**: API Handler Integration
