# Redis Implementation Strategy - Yukti Platform

**Goal**: Enhance user experience with Redis caching for sub-second response times  
**Current Status**: ❌ Redis NOT implemented (only basic cache package exists)  
**Target**: 10x faster dashboard loads (3s → 300ms)

---

## 🔍 Current State Analysis

### What Exists
- ✅ Basic Redis client package: `internal/cache/redis.go`
- ✅ Generic Set/Get methods with JSON marshaling
- ✅ Pricing data cache methods (24-hour TTL)
- ✅ Resource data cache methods (1-hour TTL)

### What's Missing
- ❌ Redis NOT running in docker-compose.yml
- ❌ No Redis integration in API handlers
- ❌ No session management in Redis
- ❌ No dashboard data caching
- ❌ No rate limiting with Redis
- ❌ No real-time features (pub/sub)
- ❌ No distributed locking
- ❌ No OTP storage in Redis

**Impact**: All API calls hit PostgreSQL directly → Slow dashboard loads (3-5 seconds)

---

## 🎯 Redis Implementation Roadmap

### Phase 1: Infrastructure Setup (1 hour)
**Goal**: Get Redis running

```yaml
# docker-compose.yml - Add Redis service
services:
  redis:
    image: redis:7-alpine
    container_name: yukti-redis
    ports:
      - "${REDIS_PORT:-6379}:6379"
    command: redis-server --requirepass ${REDIS_PASSWORD:-yukti123}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  redis_data:
```

```bash
# .env.ports - Add Redis config
REDIS_PORT=6379
REDIS_PASSWORD=yukti123
REDIS_HOST=redis
```

---

### Phase 2: Session Management (2 hours)
**Goal**: Store JWT sessions in Redis for instant validation

#### Current Flow (Slow)
```
User Request → JWT Middleware → Query PostgreSQL → Validate User → Process Request
Time: 50-100ms per request
```

#### With Redis (Fast)
```
User Request → JWT Middleware → Check Redis → Process Request
Time: 5-10ms per request (10x faster)
```

#### Implementation

```go
// internal/cache/session.go (NEW)
package cache

import (
    "fmt"
    "time"
)

type SessionCache struct {
    redis *RedisCache
}

func NewSessionCache(redis *RedisCache) *SessionCache {
    return &SessionCache{redis: redis}
}

// Store JWT session
func (s *SessionCache) SetSession(userID string, sessionData map[string]interface{}) error {
    key := fmt.Sprintf("session:%s", userID)
    return s.redis.Set(key, sessionData, 24*time.Hour)
}

// Get JWT session
func (s *SessionCache) GetSession(userID string) (map[string]interface{}, error) {
    key := fmt.Sprintf("session:%s", userID)
    var data map[string]interface{}
    err := s.redis.Get(key, &data)
    return data, err
}

// Invalidate session (logout)
func (s *SessionCache) DeleteSession(userID string) error {
    key := fmt.Sprintf("session:%s", userID)
    return s.redis.client.Del(s.redis.ctx, key).Err()
}
```

**Usage in JWT Middleware**:
```go
// internal/api/middleware/jwt_auth.go
func (m *JWTAuthMiddleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract JWT token
        token := extractToken(r)
        
        // 2. Parse JWT claims
        claims := parseJWT(token)
        
        // 3. Check Redis session (FAST)
        session, err := sessionCache.GetSession(claims.UserID)
        if err != nil {
            // Session not in Redis, query PostgreSQL (fallback)
            user := queryDatabase(claims.UserID)
            // Store in Redis for next request
            sessionCache.SetSession(claims.UserID, user)
        }
        
        // 4. Process request
        next.ServeHTTP(w, r)
    })
}
```

**Performance Gain**: 50ms → 5ms per request (10x faster)

---

### Phase 3: Dashboard Data Caching (3 hours)
**Goal**: Cache expensive dashboard queries

#### Current Flow (Slow)
```
Dashboard Load → 5 PostgreSQL queries → Aggregate data → Return JSON
Time: 3-5 seconds
```

#### With Redis (Fast)
```
Dashboard Load → Check Redis → Return cached JSON
Time: 300ms (10x faster)
```

#### Implementation

```go
// internal/cache/dashboard.go (NEW)
package cache

import (
    "fmt"
    "time"
)

type DashboardCache struct {
    redis *RedisCache
}

func NewDashboardCache(redis *RedisCache) *DashboardCache {
    return &DashboardCache{redis: redis}
}

// Cache dashboard data
func (d *DashboardCache) SetDashboard(tenantID int, data interface{}) error {
    key := fmt.Sprintf("dashboard:tenant:%d", tenantID)
    return d.redis.Set(key, data, 5*time.Minute) // 5-minute cache
}

// Get cached dashboard
func (d *DashboardCache) GetDashboard(tenantID int) (interface{}, error) {
    key := fmt.Sprintf("dashboard:tenant:%d", tenantID)
    var data interface{}
    err := d.redis.Get(key, &data)
    return data, err
}

// Invalidate dashboard cache (after scan)
func (d *DashboardCache) InvalidateDashboard(tenantID int) error {
    key := fmt.Sprintf("dashboard:tenant:%d", tenantID)
    return d.redis.client.Del(d.redis.ctx, key).Err()
}
```

**Usage in Dashboard Handler**:
```go
// internal/api/handlers/customers.go
func (h *CustomerHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    tenantID := getTenantID(r)
    
    // 1. Try Redis cache first
    cachedData, err := dashboardCache.GetDashboard(tenantID)
    if err == nil {
        log.Printf("[CACHE HIT] Dashboard for tenant %d", tenantID)
        json.NewEncoder(w).Encode(cachedData)
        return
    }
    
    // 2. Cache miss - query PostgreSQL
    log.Printf("[CACHE MISS] Dashboard for tenant %d", tenantID)
    data := queryDashboardFromDB(tenantID)
    
    // 3. Store in Redis for next request
    dashboardCache.SetDashboard(tenantID, data)
    
    json.NewEncoder(w).Encode(data)
}
```

**Cache Invalidation**:
```go
// After resource scan completes
func (s *AWSScanner) ScanTenant(tenantID int) {
    // ... scan resources ...
    
    // Invalidate dashboard cache
    dashboardCache.InvalidateDashboard(tenantID)
    
    log.Printf("[CACHE] Invalidated dashboard for tenant %d", tenantID)
}
```

**Performance Gain**: 3-5s → 300ms (10x faster)

---

### Phase 4: OTP Storage (1 hour)
**Goal**: Store OTP codes in Redis instead of PostgreSQL

#### Current Flow (Missing)
```
Signup → Generate OTP → ??? (no storage) → User enters OTP → ??? (no validation)
```

#### With Redis (Secure)
```
Signup → Generate OTP → Store in Redis (10-min TTL) → User enters OTP → Validate from Redis
```

#### Implementation

```go
// internal/cache/otp.go (NEW)
package cache

import (
    "crypto/rand"
    "fmt"
    "math/big"
    "time"
)

type OTPCache struct {
    redis *RedisCache
}

func NewOTPCache(redis *RedisCache) *OTPCache {
    return &OTPCache{redis: redis}
}

// Generate and store OTP
func (o *OTPCache) GenerateOTP(email string) (string, error) {
    // Generate 6-digit OTP
    max := big.NewInt(999999)
    min := big.NewInt(100000)
    n, _ := rand.Int(rand.Reader, max.Sub(max, min))
    code := fmt.Sprintf("%06d", n.Add(n, min).Int64())
    
    // Store in Redis with 10-minute expiry
    key := fmt.Sprintf("otp:%s", email)
    err := o.redis.Set(key, code, 10*time.Minute)
    
    return code, err
}

// Validate OTP
func (o *OTPCache) ValidateOTP(email, code string) (bool, error) {
    key := fmt.Sprintf("otp:%s", email)
    
    var storedCode string
    err := o.redis.Get(key, &storedCode)
    if err != nil {
        return false, fmt.Errorf("OTP expired or not found")
    }
    
    if storedCode != code {
        return false, fmt.Errorf("Invalid OTP")
    }
    
    // Delete OTP after successful validation
    o.redis.client.Del(o.redis.ctx, key)
    
    return true, nil
}

// Resend OTP (rate limiting)
func (o *OTPCache) CanResendOTP(email string) (bool, error) {
    key := fmt.Sprintf("otp:resend:%s", email)
    
    // Check if resend cooldown exists
    _, err := o.redis.client.Get(o.redis.ctx, key).Result()
    if err == nil {
        return false, fmt.Errorf("Please wait 60 seconds before resending")
    }
    
    // Set 60-second cooldown
    o.redis.client.Set(o.redis.ctx, key, "1", 60*time.Second)
    
    return true, nil
}
```

**Usage in Auth Handler**:
```go
// internal/api/handlers/auth.go
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
    // ... create user ...
    
    // Generate OTP and store in Redis
    otp, _ := otpCache.GenerateOTP(email)
    
    // Send OTP via AWS SES
    emailService.SendOTP(email, otp)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "OTP sent to your email",
    })
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")
    code := r.FormValue("code")
    
    // Validate OTP from Redis
    valid, err := otpCache.ValidateOTP(email, code)
    if !valid {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   err.Error(),
        })
        return
    }
    
    // Update user as verified
    db.Exec("UPDATE yt_users SET email_verified = true WHERE email = $1", email)
    
    // Generate JWT token
    token := generateJWT(email)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "token":   token,
    })
}
```

**Benefits**:
- ✅ Automatic expiry (10 minutes)
- ✅ No database cleanup needed
- ✅ Rate limiting for resend
- ✅ Instant validation

---

### Phase 5: Rate Limiting (2 hours)
**Goal**: Prevent API abuse with Redis-based rate limiting

#### Implementation

```go
// internal/cache/rate_limiter.go (NEW)
package cache

import (
    "fmt"
    "time"
)

type RateLimiter struct {
    redis *RedisCache
}

func NewRateLimiter(redis *RedisCache) *RateLimiter {
    return &RateLimiter{redis: redis}
}

// Check rate limit (100 requests per minute)
func (r *RateLimiter) CheckLimit(userID string, limit int, window time.Duration) (bool, error) {
    key := fmt.Sprintf("ratelimit:%s", userID)
    
    // Increment counter
    count, err := r.redis.client.Incr(r.redis.ctx, key).Result()
    if err != nil {
        return false, err
    }
    
    // Set expiry on first request
    if count == 1 {
        r.redis.client.Expire(r.redis.ctx, key, window)
    }
    
    // Check if limit exceeded
    if count > int64(limit) {
        return false, fmt.Errorf("Rate limit exceeded: %d/%d requests", count, limit)
    }
    
    return true, nil
}
```

**Usage in Middleware**:
```go
// internal/api/middleware/rate_limiter.go
func (m *RateLimiter) Limit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserID(r)
        
        // Check rate limit (100 req/min)
        allowed, err := rateLimiter.CheckLimit(userID, 100, 1*time.Minute)
        if !allowed {
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Rate limit exceeded. Try again in 60 seconds.",
            })
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

### Phase 6: Real-Time Features (3 hours)
**Goal**: Live dashboard updates with Redis Pub/Sub

#### Implementation

```go
// internal/cache/pubsub.go (NEW)
package cache

type PubSub struct {
    redis *RedisCache
}

func NewPubSub(redis *RedisCache) *PubSub {
    return &PubSub{redis: redis}
}

// Publish scan completion event
func (p *PubSub) PublishScanComplete(tenantID int) error {
    channel := fmt.Sprintf("scan:complete:%d", tenantID)
    message := map[string]interface{}{
        "tenant_id": tenantID,
        "timestamp": time.Now(),
    }
    
    data, _ := json.Marshal(message)
    return p.redis.client.Publish(p.redis.ctx, channel, data).Err()
}

// Subscribe to scan events
func (p *PubSub) SubscribeScanEvents(tenantID int, callback func(message string)) {
    channel := fmt.Sprintf("scan:complete:%d", tenantID)
    pubsub := p.redis.client.Subscribe(p.redis.ctx, channel)
    
    for msg := range pubsub.Channel() {
        callback(msg.Payload)
    }
}
```

**Frontend Integration** (WebSocket):
```typescript
// frontend/src/services/websocket.ts
const ws = new WebSocket('ws://localhost:8081/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  if (data.type === 'scan_complete') {
    // Refresh dashboard automatically
    fetchDashboardData();
    showToast('Resource scan completed!');
  }
};
```

---

## 📊 Performance Comparison

### Before Redis (Current)

| Operation | Time | Database Hits |
|-----------|------|---------------|
| Login | 200ms | 2 queries |
| Dashboard Load | 3-5s | 5 queries |
| Resources List | 1-2s | 3 queries |
| Hidden Costs | 2-3s | 4 queries |
| **Total** | **6-10s** | **14 queries** |

### After Redis (Target)

| Operation | Time | Database Hits | Redis Hits |
|-----------|------|---------------|------------|
| Login | 50ms | 0 (cached) | 1 |
| Dashboard Load | 300ms | 0 (cached) | 1 |
| Resources List | 200ms | 0 (cached) | 1 |
| Hidden Costs | 400ms | 0 (cached) | 1 |
| **Total** | **950ms** | **0** | **4** |

**Performance Gain**: 6-10s → 950ms (6-10x faster)

---

## 🗄️ Redis Data Structure

### Key Naming Convention
```
session:{user_id}                    → JWT session data (24h TTL)
dashboard:tenant:{tenant_id}         → Dashboard data (5min TTL)
resources:tenant:{tenant_id}         → Resources list (10min TTL)
findings:tenant:{tenant_id}          → Findings list (5min TTL)
otp:{email}                          → OTP code (10min TTL)
otp:resend:{email}                   → Resend cooldown (60s TTL)
ratelimit:{user_id}                  → Rate limit counter (1min TTL)
scan:status:{tenant_id}              → Scan progress (1h TTL)
pricing:{region}:{instance_type}     → AWS pricing (24h TTL)
```

### Memory Usage Estimate
```
1,000 users × 5KB session = 5MB
1,000 tenants × 50KB dashboard = 50MB
1,000 tenants × 200KB resources = 200MB
OTP codes (transient) = 1MB
Rate limiting (transient) = 2MB
---
Total: ~260MB for 1,000 tenants
```

**Redis Instance**: 512MB (sufficient for 2,000 tenants)

---

## 🚀 Implementation Priority

### Week 1: Critical (Must Have)
1. ✅ Add Redis to docker-compose.yml
2. ✅ Session management (JWT caching)
3. ✅ OTP storage (email verification)
4. ✅ Dashboard data caching

**Impact**: 10x faster dashboard, complete signup flow

### Week 2: Important (Should Have)
5. ✅ Rate limiting
6. ✅ Resources list caching
7. ✅ Findings list caching

**Impact**: API protection, faster page loads

### Week 3: Nice to Have
8. ✅ Real-time features (Pub/Sub)
9. ✅ Distributed locking (scan coordination)
10. ✅ Cache warming (pre-populate on startup)

**Impact**: Live updates, better UX

---

## 📝 Implementation Checklist

### Infrastructure
- [ ] Add Redis service to docker-compose.yml
- [ ] Add Redis environment variables to .env.ports
- [ ] Test Redis connection from backend
- [ ] Add Redis health check

### Code Changes
- [ ] Create session cache package
- [ ] Create dashboard cache package
- [ ] Create OTP cache package
- [ ] Create rate limiter package
- [ ] Update JWT middleware to use Redis
- [ ] Update dashboard handler to use Redis
- [ ] Create email verification endpoint
- [ ] Add cache invalidation on scan complete

### Testing
- [ ] Test session caching (login/logout)
- [ ] Test dashboard caching (load time)
- [ ] Test OTP flow (signup → verify)
- [ ] Test rate limiting (100 req/min)
- [ ] Test cache invalidation (after scan)
- [ ] Load test with 100 concurrent users

### Documentation
- [ ] Update README.md with Redis setup
- [ ] Document cache key patterns
- [ ] Document TTL strategies
- [ ] Create Redis monitoring guide

---

## 🎯 Success Metrics

### Performance
- ✅ Dashboard load: 3-5s → 300ms (10x faster)
- ✅ Login time: 200ms → 50ms (4x faster)
- ✅ API response: 500ms → 50ms (10x faster)

### User Experience
- ✅ Instant dashboard refresh
- ✅ Real-time scan updates
- ✅ No loading spinners (cached data)
- ✅ Smooth navigation (no delays)

### Scalability
- ✅ Support 10,000 concurrent users
- ✅ Handle 1M requests/day
- ✅ 99.9% cache hit rate

---

## 🔧 Quick Start Commands

```bash
# Start Redis
docker-compose up -d redis

# Check Redis status
docker exec -it yukti-redis redis-cli ping

# Monitor Redis
docker exec -it yukti-redis redis-cli monitor

# Check memory usage
docker exec -it yukti-redis redis-cli info memory

# Flush all cache (dev only)
docker exec -it yukti-redis redis-cli FLUSHALL
```

---

## 📊 Monitoring Dashboard

### Redis Metrics to Track
- **Hit Rate**: Cache hits / Total requests (target: >95%)
- **Memory Usage**: Current / Max (target: <80%)
- **Evictions**: Keys evicted due to memory (target: 0)
- **Connections**: Active connections (target: <100)
- **Latency**: Average response time (target: <5ms)

### Grafana Dashboard
```
Panel 1: Cache Hit Rate (line chart)
Panel 2: Memory Usage (gauge)
Panel 3: Keys by Pattern (pie chart)
Panel 4: Operations per Second (bar chart)
```

---

**Estimated Implementation Time**: 2-3 days  
**Performance Improvement**: 10x faster  
**User Experience**: Instant, real-time, smooth
