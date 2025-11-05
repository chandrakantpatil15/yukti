# Yukti FinOps - Health Monitoring & Kill Switch

## 🏥 **Health Monitoring System**

### **Real-Time Health Check**
```bash
# Start health monitor (runs on port 8081)
make health-monitor

# Check system health
make health-check
```

**Health Check Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-11-04T10:30:00Z",
  "version": "1.0.0",
  "components": {
    "database": "healthy",
    "aws": "healthy"
  },
  "test_instances": 5,
  "estimated_hourly_cost": 0.052,
  "uptime": "2h15m30s"
}
```

### **Health Status Levels**
- **🟢 healthy**: All systems operational
- **🟡 degraded**: Some components have issues
- **🔴 unhealthy**: Critical systems down
- **🔒 kill-switch-enabled**: Emergency mode active

## 🚨 **Kill Switch System**

### **Manual Kill Switch Control**
```bash
# Enable kill switch
make kill-switch-enable reason="Cost limit exceeded"

# Disable kill switch
make kill-switch-disable

# Check kill switch status
./scripts/health-check-endpoints.sh kill-status
```

### **Emergency Stop**
```bash
# Immediate termination of all test instances
make emergency-stop
```

**Emergency Stop Actions:**
1. ✅ Terminates all test instances immediately
2. ✅ Enables kill switch automatically
3. ✅ Logs emergency action with timestamp
4. ✅ Prevents new instance launches

## 💰 **Cost Guard Protection**

### **Automatic Cost Monitoring**
```bash
# Start cost guard (continuous monitoring)
make cost-guard
```

**Default Protection Limits:**
- **Max Hourly Cost**: $1.00
- **Max Daily Cost**: $20.00
- **Check Interval**: 5 minutes
- **Auto-shutdown**: After 4 hours runtime

### **Custom Cost Limits**
```bash
# Set custom limits via environment variables
export MAX_HOURLY_COST=0.50
export MAX_DAILY_COST=10.00
export CHECK_INTERVAL=2m
make cost-guard
```

### **Cost Guard Actions**
When limits are exceeded:
1. 🚨 **Immediate Alert**: Logs cost violation
2. 🛑 **Auto-Termination**: Stops all test instances
3. 🔒 **Kill Switch**: Enables emergency mode
4. 📊 **Cost Report**: Shows final costs and instances

## 🛡️ **Safety Features**

### **Multi-Layer Protection**
1. **Health Monitor**: Real-time system status
2. **Cost Guard**: Automatic cost limits
3. **Kill Switch**: Manual emergency control
4. **Auto-Cleanup**: Scheduled termination
5. **Instance Tagging**: Easy identification

### **Fail-Safe Mechanisms**
- **Default Limits**: Conservative cost thresholds
- **Auto-Termination**: Long-running instance detection
- **Emergency Override**: Manual kill switch
- **Cleanup Scripts**: Automated resource cleanup

## 📊 **Monitoring Dashboard**

### **Continuous Monitoring**
```bash
# Start continuous health monitoring
./scripts/health-check-endpoints.sh monitor
```

**Sample Output:**
```
2024-11-04 10:30:00 - Health Check
----------------------------------------
{
  "status": "healthy",
  "test_instances": 5,
  "estimated_hourly_cost": 0.052,
  "components": {
    "database": "healthy",
    "aws": "healthy"
  }
}
```

### **Cost Monitoring Output**
```
💰 YUKTI FINOPS - COST GUARD
===========================
🛡️  Cost Protection Limits:
   Max Hourly Cost: $1.00
   Max Daily Cost:  $20.00
   Check Interval:  5m0s

⏰ 10:30:05 - Cost Check
   Running Instances: 5
   Current Hourly Cost: $0.0520
   Projected Daily Cost: $1.25
   ✅ Within cost limits
```

## 🚨 **Emergency Procedures**

### **Cost Overrun Response**
1. **Automatic**: Cost Guard triggers emergency stop
2. **Manual**: Use `make emergency-stop`
3. **Verification**: Check with `make health-check`

### **System Failure Response**
1. **Health Check**: Identify failed components
2. **Kill Switch**: Enable if necessary
3. **Manual Cleanup**: Use cleanup scripts
4. **Restart**: Disable kill switch when ready

### **Testing Emergency Procedures**
```bash
# Test emergency stop (safe - only affects test instances)
make emergency-stop

# Verify cleanup
make health-check

# Reset system
make kill-switch-disable
```

## 🎯 **Best Practices**

### **Before Testing**
1. ✅ Start health monitor: `make health-monitor`
2. ✅ Start cost guard: `make cost-guard`
3. ✅ Set appropriate cost limits
4. ✅ Verify emergency procedures

### **During Testing**
1. 📊 Monitor health regularly
2. 💰 Watch cost accumulation
3. ⏰ Set testing time limits
4. 🔍 Check for long-running instances

### **After Testing**
1. 🧹 Always run cleanup: `make test-cleanup`
2. ✅ Verify termination: `make health-check`
3. 📊 Review cost reports
4. 🔒 Disable kill switch if enabled

## 🎉 **Safety Summary**

**✅ Complete Protection:**
- Real-time health monitoring
- Automatic cost limits ($1/hour, $20/day)
- Manual kill switch control
- Emergency stop capability
- Long-running instance detection
- Automated cleanup procedures

**The Yukti FinOps platform now has enterprise-grade safety controls to prevent cost overruns and ensure reliable testing!**