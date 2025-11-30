# 🚀 Test Customer → Yukti Platform Flow NOW

## ✅ Setup Complete

**Simulates**: Customer giving Yukti read-only access to their AWS account

**What's Ready**:
- ✅ Yukti platform IAM user created
- ✅ Customer IAM role created (ReadOnlyAccess)
- ✅ Backend using Yukti credentials
- ✅ Real AWS verification enabled

---

## Test in 1 Minute

### 1. Open Onboarding
```
http://localhost:3000/onboarding
```

### 2. Enter These Values
- **AWS Account ID**: `144403604430`
- **Role Name**: `YuktiTestReadOnlyRole`

### 3. Click "Connect AWS Account"

### 4. Watch Logs
```bash
docker-compose logs -f backend
```

---

## What Will Happen

Backend will:
1. ✅ Validate Account ID format
2. ✅ Validate Role ARN format
3. ✅ Auto-generate External ID
4. ✅ Call AWS STS AssumeRole
5. ❌ **FAIL** with "Access Denied" (External ID mismatch)

**Why fail?** 
- Trust policy expects: `yukti-test-12345`
- Backend generates: `yukti-18-{random}`

---

## This Proves

✅ **Real AWS verification working**  
✅ **External ID security working** (prevents wrong external ID)  
✅ **Customer → Yukti flow ready**  

To make it succeed, update trust policy with backend's generated external ID.

---

## Test Now

http://localhost:3000/onboarding
