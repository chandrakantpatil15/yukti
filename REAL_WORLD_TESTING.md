# Yukti FinOps - Real World Testing Guide

## 🎯 **Testing Scenario**
Create realistic workload patterns with **20 spot instances** running stress tests to validate the Yukti FinOps platform with real CloudWatch data.

## 📋 **Test Setup**

### **Workload Pattern Design**
```
Stress Cycle (Repeating):
├── 10 minutes: High CPU/Memory stress (80-90% utilization)
├── 5 minutes: Idle period (0-5% utilization)  
└── Repeat continuously
```

**Expected Classifications:**
- **Batch/Intermittent**: High spikes → idle periods
- **Variable Utilization**: 80-90% → 0-5% cycles
- **Bursty Pattern**: High variance in CPU usage
- **Cost Optimization**: Spot instances = 70% savings

## 🚀 **Quick Start**

### **1. Launch Test Infrastructure**
```bash
# Setup 20 spot instances with stress testing
make test-setup
```

**What this does:**
- ✅ Launches 20 t3.medium spot instances
- ✅ Installs stress testing tools
- ✅ Configures CloudWatch detailed monitoring
- ✅ Sets up automated stress cycles (10min stress → 5min idle)
- ✅ Tags instances for identification

### **2. Monitor Real-Time Data**
```bash
# Start real-time monitoring
make test-monitor
```

**Monitoring Features:**
- ✅ Collects CloudWatch metrics every 5 minutes
- ✅ Stores data in PostgreSQL
- ✅ Classifies workload patterns in real-time
- ✅ Shows pattern distribution summary

### **3. Run FinOps Analysis**
```bash
# Wait 30-60 minutes for data, then analyze
make sync-all-data
make assess-daily
```

### **4. Cleanup Resources**
```bash
# Terminate all test instances
make test-cleanup
```

## 📊 **Expected Results**

### **Workload Patterns**
After 1 hour of testing, expect to see:

| Pattern | Count | Avg CPU | Max CPU | Optimization Score |
|---------|-------|---------|---------|-------------------|
| batch   | 12-15 | 25-35%  | 85-95%  | 0.75-0.85        |
| bursty  | 3-5   | 40-50%  | 80-90%  | 0.65-0.75        |
| idle    | 2-3   | 2-8%    | 10-15%  | 0.90-0.95        |

### **Cost Analysis**
- **Spot Instances**: ~$0.01-0.02/hour each
- **Total Test Cost**: ~$4-8 for 2-hour test
- **Potential Savings Identified**: 60-80% through rightsizing

### **Assessment Accuracy**
- **Classification Accuracy**: >90% for clear patterns
- **Optimization Recommendations**: Downsize, spot, terminate
- **Cost Savings Potential**: $50-200/month per instance

## 🔍 **Validation Checklist**

### **Data Collection** ✅
- [ ] 20 instances launched successfully
- [ ] CloudWatch metrics flowing (5-minute intervals)
- [ ] Stress patterns executing (10min/5min cycles)
- [ ] Database storing metrics correctly

### **Pattern Recognition** ✅
- [ ] Batch patterns detected (high → low cycles)
- [ ] Bursty patterns identified (high variance)
- [ ] Idle periods classified correctly
- [ ] Optimization scores calculated

### **Cost Optimization** ✅
- [ ] Pricing data integrated
- [ ] Monthly cost calculations accurate
- [ ] Savings recommendations generated
- [ ] Spot instance benefits quantified

### **Enterprise Features** ✅
- [ ] Multi-tenant data isolation
- [ ] API endpoints responding
- [ ] Timeline queries performing well
- [ ] Assessment history tracking

## 📈 **Monitoring Dashboard**

Real-time monitoring shows:
```
📊 MONITORING SUMMARY
------------------------------
Pattern      Count    Avg CPU    Max CPU
batch        14       28.5       89.2
bursty       4        45.1       82.7
idle         2        4.2        12.1

📈 Total metrics collected (last hour): 240
```

## 🎯 **Success Criteria**

**✅ Platform Validation:**
- Real CloudWatch data processed correctly
- Workload patterns classified with >90% accuracy
- Cost optimization recommendations generated
- Assessment engine handles 20+ concurrent instances

**✅ Enterprise Readiness:**
- Multi-tenant isolation working
- API performance acceptable (<1s response)
- Database queries optimized
- Scalable architecture validated

## 💡 **Next Steps After Testing**

1. **Scale Testing**: Increase to 50-100 instances
2. **Pattern Refinement**: Tune classification algorithms
3. **ML Integration**: Add Lambda-based ML models
4. **Customer Onboarding**: Deploy for first enterprise customer

## 🧹 **Cost Management**

**Estimated Costs:**
- **2-hour test**: ~$4-8 total
- **24-hour test**: ~$48-96 total
- **Cleanup**: Automatic termination prevents runaway costs

**Always run cleanup after testing to avoid unexpected charges!**

```bash
make test-cleanup  # Terminates all test instances
```