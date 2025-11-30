# Stripe Billing Implementation - Complete Summary

## Overview
This document summarizes the complete Stripe billing integration, trial enforcement, and feature gating implementation.

## Database Changes

### Migration: `scripts/013_create_billing_tables.sql`

**Tables Created:**
1. `yt_billing_customers` - Links tenants to Stripe customers
2. `yt_stripe_events` - Webhook event log
3. `yt_billing_invoices` - Invoice records
4. `yt_billing_products_prices` - Stripe price mappings
5. `yt_billing_subscriptions` - Active subscription tracking

**Key Features:**
- Foreign keys to `yt_tenants`
- Indexes for performance
- Auto-update triggers for `updated_at`
- Default price mappings (placeholders - update with actual Stripe price IDs)

**To Apply:**
```bash
psql $DATABASE_URL -f scripts/013_create_billing_tables.sql
```

## Backend Implementation

### New Files

1. **`internal/models/billing.go`**
   - GORM models for all billing tables
   - Helper functions: `GetBillingCustomerByTenant`, `GetBillingSubscriptionByTenant`, `GetBillingProductPriceByPlan`

2. **`internal/billing/stripe_client.go`**
   - Stripe API client wrapper
   - Functions:
     - `CreateCustomer()` - Creates Stripe customer for tenant
     - `CreateCheckoutSession()` - Creates checkout session
     - `VerifyWebhookSignature()` - Verifies webhook authenticity
     - `StoreStripeEvent()` - Stores webhook events
     - `ProcessInvoicePaid()` - Handles invoice.paid events
     - `ProcessSubscriptionUpdated()` - Handles subscription.updated events
     - `ProcessSubscriptionCanceled()` - Handles subscription.deleted events

3. **`internal/api/handlers/billing.go`**
   - `CreateCheckoutSession()` - POST /api/v1/billing/checkout-session
   - `HandleStripeWebhook()` - POST /api/v1/webhooks/stripe
   - `GetBillingInfo()` - GET /api/v1/billing/info

4. **`internal/api/middleware/trial_enforcement.go`**
   - `RequireActiveSubscription()` - Middleware that enforces trial expiration
   - Returns 402 Payment Required if trial expired and no subscription
   - `CheckTrialStatus()` - Helper to check trial status

### Modified Files

1. **`internal/feature/feature_gate.go`**
   - Updated `IsEnabled()` to check trial expiration
   - Checks for active subscription before allowing features
   - Returns proper error messages with upgrade URLs
   - Added `RequireFeatureMiddleware()` for route protection

2. **`internal/api/routes/routes.go`**
   - Added billing routes:
     - POST /api/v1/billing/checkout-session (JWT protected)
     - GET /api/v1/billing/info (JWT protected)
     - POST /api/v1/webhooks/stripe (public, signature verified)

3. **`internal/database/database.go`**
   - Added billing models to AutoMigrate:
     - `BillingCustomer`
     - `StripeEvent`
     - `BillingInvoice`
     - `BillingProductPrice`
     - `BillingSubscription`

## Environment Variables Required

```bash
# Stripe API keys (required)
STRIPE_SECRET_KEY=sk_test_...  # or sk_live_... for production
STRIPE_WEBHOOK_SECRET=whsec_...  # From Stripe dashboard webhook settings

# Frontend URL for checkout redirects
FRONTEND_URL=http://localhost:3000  # or https://app.yukti.io for production
```

## Stripe Setup Instructions

1. **Create Stripe Account** (if not already done)
   - Go to https://dashboard.stripe.com
   - Get API keys from Developers > API keys

2. **Create Products and Prices**
   - Go to Products in Stripe dashboard
   - Create products:
     - Professional Plan ($99/month)
     - Enterprise Plan ($499/month)
     - Financial Plan ($1999/month)
   - Create prices for each product (monthly recurring)
   - Note the `price_xxx` IDs

3. **Update Price Mappings**
   ```sql
   UPDATE yt_billing_products_prices
   SET stripe_price_id = 'price_xxxxx'
   WHERE plan_slug = 'professional' AND currency = 'USD';
   
   -- Repeat for enterprise and financial
   ```

4. **Configure Webhook**
   - Go to Developers > Webhooks
   - Add endpoint: `https://api.yukti.io/api/v1/webhooks/stripe`
   - Select events:
     - `invoice.paid`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`
   - Copy webhook signing secret to `STRIPE_WEBHOOK_SECRET`

## API Endpoints

### POST /api/v1/billing/checkout-session
Creates a Stripe Checkout session.

**Request:**
```json
{
  "plan_slug": "professional",
  "currency": "USD"
}
```

**Response:**
```json
{
  "success": true,
  "checkout_url": "https://checkout.stripe.com/c/pay/..."
}
```

### GET /api/v1/billing/info
Returns current billing information.

**Response:**
```json
{
  "success": true,
  "data": {
    "subscription_tier": "professional",
    "trial_ends_at": "2024-02-01T00:00:00Z",
    "subscription": {
      "id": 1,
      "stripe_subscription_id": "sub_xxx",
      "plan_slug": "professional",
      "status": "active",
      "current_period_end": "2024-03-01T00:00:00Z"
    },
    "invoices": [...]
  }
}
```

### POST /api/v1/webhooks/stripe
Stripe webhook endpoint (no auth, signature verified).

**Headers:**
- `Stripe-Signature: t=...,v1=...`

**Events Handled:**
- `invoice.paid` - Creates invoice record, updates tenant tier
- `customer.subscription.updated` - Updates subscription status
- `customer.subscription.deleted` - Cancels subscription, downgrades tenant

## Trial Enforcement

### How It Works

1. **Trial Period**: New tenants get 30-day trial (set in `yt_tenants.trial_ends_at`)
2. **Trial Check**: Middleware checks if trial expired
3. **Subscription Check**: If trial expired, checks for active subscription
4. **Access Control**: Returns 402 Payment Required if no subscription

### Usage

```go
// In routes.go
trialEnforcement := middleware.NewTrialEnforcementMiddleware(db)
router.Handle("/api/v1/protected-endpoint",
    jwtAuthMw.RequireAuth(
        trialEnforcement.RequireActiveSubscription(
            http.HandlerFunc(handler.ProtectedEndpoint)
        )
    )
).Methods("GET")
```

## Feature Gating

### Updated Feature Matrix

| Feature | FREE | Professional | Enterprise | Financial |
|---------|------|--------------|------------|-----------|
| Budget Tracking | ✅ | ✅ | ✅ | ✅ |
| Hidden Cost Detection | ❌ | ✅ | ✅ | ✅ |
| IaC Generation | ❌ | ✅ | ✅ | ✅ |
| ML Forecasting | ❌ | ✅ | ✅ | ✅ |
| Multi-Account | ✅ | ✅ | ✅ | ✅ |
| Whitelisting | ✅ | ✅ | ✅ | ✅ |
| API Keys | ❌ | ❌ | ✅ | ✅ |

### Usage

```go
// In routes.go
router.Handle("/api/v1/iac/generate",
    jwtAuthMw.RequireAuth(
        feature.RequireFeatureMiddleware(db, feature.FeatureIaCGeneration)(
            http.HandlerFunc(handler.GenerateIaC)
        )
    )
).Methods("POST")
```

## Frontend Implementation (TODO)

Frontend billing pages need to be created:
- `/billing` - Current plan, trial status, invoices
- `/billing/upgrade` - Plan selection and checkout
- `/billing/success` - Post-checkout success page

See next section for frontend implementation.

## Testing

### Manual Testing

1. **Create Checkout Session:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/billing/checkout-session \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"plan_slug":"professional","currency":"USD"}'
   ```

2. **Get Billing Info:**
   ```bash
   curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/billing/info
   ```

3. **Test Webhook** (use Stripe CLI):
   ```bash
   stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe
   stripe trigger invoice.paid
   ```

### Unit Tests (TODO)

- `internal/api/handlers/billing_test.go`
- `internal/billing/stripe_client_test.go`
- Mock Stripe events for webhook testing

## Rollout Plan

### Phase 1: Staging
1. Run migration on staging DB
2. Set Stripe test mode keys
3. Create test products/prices in Stripe test mode
4. Update price mappings
5. Test checkout flow
6. Test webhook processing
7. Verify trial enforcement

### Phase 2: Production
1. Run migration on production DB
2. Set Stripe live mode keys
3. Create live products/prices
4. Update price mappings
5. Configure production webhook endpoint
6. Monitor webhook processing
7. Enable trial enforcement on protected endpoints

## Security Considerations

- ✅ Webhook signature verification
- ✅ No card details stored (Stripe handles PCI)
- ✅ JWT required for billing endpoints
- ✅ Trial enforcement middleware
- ⚠️ Update Stripe price IDs in database (not hardcoded)
- ⚠️ Use HTTPS in production
- ⚠️ Store webhook secret securely

## Known Limitations

1. **Webhook Event Parsing**: Currently uses generic event parsing. Should use proper Stripe types.
2. **Email Notifications**: Not implemented. Should send emails on:
   - Trial ending soon
   - Payment successful
   - Subscription canceled
3. **Multi-Currency**: Price mappings support multiple currencies but UI doesn't show currency selector
4. **Tax Calculation**: Uses Stripe Tax but doesn't display tax breakdown in UI

## Next Steps

1. Create frontend billing pages
2. Add email notifications
3. Implement proper webhook event parsing
4. Add unit tests
5. Add integration tests
6. Create admin billing dashboard
7. Add subscription management (cancel, upgrade, downgrade)

---

**Status**: ✅ Backend implementation complete. Frontend pages pending.

