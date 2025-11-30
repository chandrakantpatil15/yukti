# UI Testing Guide - Tenant Isolation Verification

## Overview
Step-by-step guide to verify tenant isolation and security fixes in the UI.

## Prerequisites
```bash
# Ensure all services are running
docker-compose ps

# Should see:
# - yukti-backend (port 8081)
# - yukti-frontend (port 3000)
# - yukti-postgres (port 5432)
```

---

## Test 1: Basic Login & Navigation

### Steps
1. Open browser: http://localhost:3000
2. Login with test credentials:
   - Email: `yourname123@example.com`
   - Password: `Chandra!@#$143`
3. Verify redirect to Dashboard
4. Navigate to each page:
   - Dashboard
   - Hidden Costs
   - Resources
   - Whitelists
   - Onboarding

### Expected Results
- ✅ Login successful
- ✅ Dashboard shows data for tenant 18
- ✅ All pages load without errors
- ✅ Navigation works smoothly

---

## Test 2: Tenant Isolation - Query Parameter Manipulation

### Steps
1. Login as tenant 18
2. Open Dashboard (http://localhost:3000/dashboard)
3. Open DevTools → Network tab
4. Find API request to `/api/customers/dashboard`
5. Check request URL and headers

### Expected Results
- ✅ **NO** `tenant_id` in query parameters
- ✅ **NO** `X-Tenant-ID` in request headers
- ✅ Only `Authorization: Bearer <token>` header present
- ✅ Backend extracts tenant_id from JWT

### Attack Attempt (Should Fail)
1. Try manually adding `?tenant_id=99` to URL
2. Refresh page
3. Verify data still shows tenant 18 (not tenant 99)

**Result**: ✅ Query parameter ignored, JWT tenant_id used

---

## Test 3: Tenant Isolation - LocalStorage Manipulation

### Steps
1. Login as tenant 18
2. Open DevTools → Application → Local Storage
3. Find `yukti_user` entry
4. Modify `tenant_id` from `18` to `99`
5. Refresh Dashboard page

### Expected Results
- ✅ Dashboard still shows tenant 18 data
- ✅ LocalStorage modification has no effect
- ✅ Backend uses JWT tenant_id (not localStorage)

---

## Test 4: Token Expiration Handling

### Steps
1. Login as tenant 18
2. Open DevTools → Application → Local Storage
3. Find `yukti_auth_token`
4. Copy token value
5. Decode at https://jwt.io
6. Note expiration time (`exp` field)
7. Manually set `exp` to past timestamp
8. Refresh page

### Expected Results
- ✅ Automatic redirect to login page
- ✅ Token cleared from localStorage
- ✅ User cleared from localStorage
- ✅ No error messages shown

---

## Test 5: Invalid Token Handling

### Steps
1. Login as tenant 18
2. Open DevTools → Application → Local Storage
3. Find `yukti_auth_token`
4. Modify token (change last few characters)
5. Refresh page

### Expected Results
- ✅ Automatic redirect to login page
- ✅ Token cleared from localStorage
- ✅ 401 Unauthorized response triggers logout

---

## Test 6: Dashboard Data Verification

### Steps
1. Login as tenant 18
2. View Dashboard
3. Note the metrics:
   - Total Savings
   - Findings Count
   - Budget Usage
   - RI Savings

### Expected Results
- ✅ Total Savings: $425.60
- ✅ Findings Count: 7
- ✅ Data matches tenant 18 seed data
- ✅ No data from other tenants visible

---

## Test 7: Hidden Costs Filtering

### Steps
1. Login as tenant 18
2. Navigate to Hidden Costs
3. Verify 7 findings displayed
4. Apply filters:
   - Category: "Data Transfer"
   - Severity: "High"
5. Clear filters
6. Open DevTools → Network tab
7. Check API requests

### Expected Results
- ✅ Filters work correctly
- ✅ API requests contain filter params only
- ✅ **NO** `tenant_id` in query parameters
- ✅ Backend uses JWT tenant_id

---

## Test 8: Onboarding Flow

### Steps
1. Login as tenant 18
2. Navigate to Onboarding
3. Fill in AWS details:
   - Account ID: `123456789012`
   - Role ARN: `arn:aws:iam::123456789012:role/YuktiRole`
4. Submit form
5. Open DevTools → Network tab
6. Check POST request to `/api/onboarding/aws-connection`

### Expected Results
- ✅ Request body contains: `account_id`, `role_arn`, `external_id`, `regions`
- ✅ **NO** `tenant_id` in request body
- ✅ Backend extracts tenant_id from JWT
- ✅ Success message displayed

---

## Test 9: Admin Impersonation (If Admin Access)

### Steps
1. Login as admin user
2. Navigate to Admin Dashboard
3. Click "View" on a customer
4. Open DevTools → Network tab
5. Check POST request to `/api/admin/impersonate`
6. Verify response contains new JWT token
7. Check localStorage for updated token

### Expected Results
- ✅ Backend returns new JWT with impersonated tenant_id
- ✅ New token stored in localStorage
- ✅ Redirect to customer's dashboard
- ✅ Dashboard shows impersonated tenant's data

---

## Test 10: Logout & Session Cleanup

### Steps
1. Login as tenant 18
2. Navigate to Dashboard
3. Click Logout button
4. Open DevTools → Application → Local Storage
5. Check for remaining data

### Expected Results
- ✅ Redirect to login page
- ✅ `yukti_auth_token` removed
- ✅ `yukti_user` removed
- ✅ All session data cleared

---

## Test 11: Cross-Browser Testing

### Browsers to Test
- Chrome
- Firefox
- Safari
- Edge

### Steps
1. Repeat Tests 1-10 in each browser
2. Verify consistent behavior
3. Check for browser-specific issues

### Expected Results
- ✅ All tests pass in all browsers
- ✅ No browser-specific bugs
- ✅ Consistent security behavior

---

## Test 12: Network Inspection

### Steps
1. Login as tenant 18
2. Navigate through all pages
3. Open DevTools → Network tab
4. Filter by "Fetch/XHR"
5. Inspect all API requests

### Expected Results
For **EVERY** API request:
- ✅ `Authorization: Bearer <token>` header present
- ✅ **NO** `X-Tenant-ID` header
- ✅ **NO** `tenant_id` in query parameters
- ✅ **NO** `tenant_id` in request body (except admin endpoints)

---

## Security Checklist

### ✅ Tenant Isolation
- [ ] User A cannot access User B's data
- [ ] Query parameter manipulation has no effect
- [ ] LocalStorage manipulation has no effect
- [ ] Header manipulation has no effect
- [ ] All data filtered by JWT tenant_id

### ✅ Session Management
- [ ] Expired tokens trigger automatic logout
- [ ] Invalid tokens trigger automatic logout
- [ ] 401 responses trigger automatic logout
- [ ] Token expiration checked on page load
- [ ] Logout clears all session data

### ✅ API Security
- [ ] No X-Tenant-ID header sent
- [ ] No tenant_id in query params
- [ ] No tenant_id in request bodies (user endpoints)
- [ ] All requests include Authorization header
- [ ] Backend validates JWT on every request

### ✅ UI/UX
- [ ] All pages load correctly
- [ ] Navigation works smoothly
- [ ] Filters work as expected
- [ ] Error messages are user-friendly
- [ ] Loading states displayed

---

## Known Issues (To Be Fixed)

### Backend Updates Needed
1. **Onboarding Endpoints**: Remove tenant_id parameter requirement
   - `GET /api/onboarding/status`
   - `GET /api/onboarding/external-id`
   - `POST /api/onboarding/aws-connection`

2. **Admin Impersonate**: Return new JWT token
   - `POST /api/admin/impersonate` should return `{ token: "..." }`

### Frontend Improvements
1. Add loading spinner during API calls
2. Add error boundary for API failures
3. Add retry logic for failed requests
4. Add toast notifications for success/error

---

## Troubleshooting

### Issue: "Session expired" on every page load
**Cause**: Token expiration time too short
**Fix**: Check JWT_EXPIRATION in backend config

### Issue: 401 Unauthorized on all requests
**Cause**: JWT secret mismatch or invalid token
**Fix**: Clear localStorage and login again

### Issue: Dashboard shows no data
**Cause**: Tenant 18 has no seed data
**Fix**: Run seed script: `make seed`

### Issue: Onboarding fails with 400 error
**Cause**: Backend still expects tenant_id parameter
**Fix**: Update backend onboarding handlers

---

## Success Criteria

All tests must pass with these results:
- ✅ 0 tenant isolation vulnerabilities
- ✅ 0 session management issues
- ✅ 0 API security issues
- ✅ 100% JWT-based authentication
- ✅ Automatic logout on token expiration
- ✅ No client-side tenant_id manipulation possible

---

**Last Updated**: Session 15 - UI security testing guide
**Status**: Ready for comprehensive testing
