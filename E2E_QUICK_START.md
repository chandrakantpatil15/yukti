# 🚀 E2E Test - Quick Start

## ✅ Database Cleaned - Ready to Test!

```
✅ All user data cleared
✅ Fresh start for onboarding
✅ Platform running on http://localhost:3000
```

---

## 📝 Test Credentials (Create New)

**Email**: `test@yukti.com` (or your email)  
**Password**: `Test@123456`  
**AWS Account**: `424851482219`  
**IAM Role**: `YuktiFinOpsRole`

---

## ⚡ Quick Test Flow (30 min)

### 1. Signup (5 min)
```
http://localhost:3000/signup
→ Enter email + password
→ Get OTP from: docker-compose logs backend | grep "OTP"
→ Verify email
```

### 2. Login (2 min)
```
http://localhost:3000/login
→ Enter credentials
→ Redirected to Dashboard
```

### 3. Onboarding (10 min)
```
Dashboard → Configure AWS
→ Account ID: 424851482219
→ Role Name: YuktiFinOpsRole
→ Copy External ID
→ Create IAM role in AWS Console
→ Verify Connection
```

### 4. First Scan (5 min)
```
Dashboard → Scan Resources
→ Wait 30-60 seconds
→ Check Resources page
→ Verify EC2/RDS/S3 counts
```

### 5. Check Results (5 min)
```
Resources → View discovered resources
Hidden Costs → Check findings (may be 0 first time)
Profile → Verify AWS connection
```

---

## 🎯 Expected Results

### First Scan (Today):
- ✅ Resources discovered
- ✅ Metadata stored
- ⚠️ 0 findings (need 24h metrics)

### Second Scan (Tomorrow):
- ✅ CloudWatch metrics collected
- ✅ 5-10 findings generated
- ✅ Savings recommendations

---

## 🐛 Quick Troubleshooting

**SES Permission Error?**
```bash
# Add SES permissions to IAM user
./scripts/add-ses-permissions.sh
docker-compose restart backend
```

**OTP not showing?**
```bash
docker-compose logs backend | grep "OTP"
```

**AWS connection failed?**
```bash
docker-compose logs backend | grep "ERROR"
```

**No resources found?**
- Check IAM role permissions
- Verify resources exist in AWS
- Check configured regions

**Backend crashed?**
```bash
docker-compose restart backend
```

---

## 🔄 Reset Database

To start over:
```bash
psql -U yukti -d yukti_finops -f scripts/cleanup-for-e2e-test.sql
```

---

## 📚 Full Guide

See `E2E_TEST_GUIDE.md` for detailed step-by-step instructions.

---

**Ready to test! Start at: http://localhost:3000** 🚀
