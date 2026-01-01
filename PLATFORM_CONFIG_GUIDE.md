# Platform Configuration Management

## 🎯 CONFIGURABLE SETTINGS

All platform settings are stored in `yt_platform_config` table and can be changed anytime without code deployment.

---

## 📊 AVAILABLE SETTINGS

### **trial_days** (Default: 30)
- **Description**: Number of days for free trial
- **Usage**: New signups get this many days free
- **Change**: Update anytime to affect NEW signups

### **grace_days** (Default: 7)
- **Description**: Grace period after subscription expires
- **Usage**: Users can still access platform for this many days after expiry
- **Change**: Affects all expired subscriptions

### **max_users_per_tenant** (Default: 10)
- **Description**: Maximum team members on free/trial plan
- **Usage**: Limits team size for non-paying customers

### **max_aws_accounts** (Default: 1)
- **Description**: Maximum AWS accounts on trial
- **Usage**: Trial users can connect only 1 AWS account

---

## 🔧 HOW TO CHANGE SETTINGS

### **Option 1: Direct Database Update**
```sql
-- Change trial period to 14 days
UPDATE yt_platform_config 
SET config_value = '14', updated_at = NOW() 
WHERE config_key = 'trial_days';

-- Change grace period to 3 days
UPDATE yt_platform_config 
SET config_value = '3', updated_at = NOW() 
WHERE config_key = 'grace_days';

-- View all settings
SELECT * FROM yt_platform_config ORDER BY config_key;
```

### **Option 2: Admin API** (Future)
```bash
# Update via API
curl -X PUT http://localhost:8081/api/admin/config \
  -H "Authorization: Bearer admin_token" \
  -d '{"trial_days": "14"}'
```

### **Option 3: Environment Variables** (Override)
```bash
# .env file
TRIAL_DAYS=14
GRACE_DAYS=3

# Backend reads from env first, then database
```

---

## 📋 CURRENT SETTINGS

To view current settings:
```sql
SELECT 
    config_key,
    config_value,
    description,
    updated_at
FROM yt_platform_config
ORDER BY config_key;
```

**Output**:
```
config_key            | config_value | description
----------------------|--------------|----------------------------------
grace_days            | 7            | Grace period after expiry
max_aws_accounts      | 1            | Max AWS accounts on trial
max_users_per_tenant  | 10           | Max users per tenant on free plan
trial_days            | 30           | Default trial period in days
```

---

## 🎯 USE CASES

### **Scenario 1: Black Friday Promotion**
```sql
-- Extend trial to 60 days for Black Friday
UPDATE yt_platform_config 
SET config_value = '60' 
WHERE config_key = 'trial_days';

-- After promotion ends, revert to 30 days
UPDATE yt_platform_config 
SET config_value = '30' 
WHERE config_key = 'trial_days';
```

### **Scenario 2: Reduce Trial for Abuse**
```sql
-- If seeing too many trial abusers, reduce to 7 days
UPDATE yt_platform_config 
SET config_value = '7' 
WHERE config_key = 'trial_days';
```

### **Scenario 3: Enterprise Customer**
```sql
-- Give specific tenant 90-day trial
UPDATE yt_subscriptions 
SET trial_days = 90,
    current_period_end = NOW() + INTERVAL '90 days'
WHERE tenant_id = 123;
```

---

## 🔐 SECURITY

**Who can change settings?**
- ✅ Database admin (direct SQL access)
- ✅ Platform admin (via admin API - future)
- ❌ Regular users (no access)
- ❌ Tenant admins (no access)

**Audit Trail**:
- All changes logged with `updated_at` timestamp
- Future: Add `updated_by` column for admin tracking

---

## 📝 ADDING NEW SETTINGS

To add a new configurable setting:

```sql
-- Add new setting
INSERT INTO yt_platform_config (config_key, config_value, description)
VALUES ('new_setting', 'default_value', 'Description of what it does');

-- Use in code
SELECT config_value FROM yt_platform_config WHERE config_key = 'new_setting';
```

---

## ✅ BENEFITS

1. **No Code Deployment**: Change settings instantly
2. **A/B Testing**: Test different trial periods
3. **Promotions**: Easy to run limited-time offers
4. **Per-Tenant Override**: Can customize for specific customers
5. **Audit Trail**: Track when settings changed

---

**Status**: Implemented in migration 012_add_subscriptions.sql
**Location**: `yt_platform_config` table
**Access**: Database admin or future admin API
