# Yukti Platform Features - Implementation Plan

**Goal**: Add billing, subscriptions, revenue analytics, and infrastructure visualization  
**Timeline**: 7-8 weeks  
**Status**: Ready to implement

---

## 🎯 Overview

Adding **7 new features** to Yukti FinOps platform:
1. Billing Dashboard
2. Payment Gateways
3. Subscriptions Management
4. Revenue Analytics
5. Infrastructure Visualization
6. System Architecture Diagram
7. Invoice Generation

---

## 📅 Implementation Phases

### **Phase 1: Core Billing** (3 weeks) - HIGH PRIORITY

#### Week 1: Billing Dashboard
**Goal**: Manage customer billing and invoices

**Frontend**:
- `frontend/src/pages/Admin/AdminBilling.tsx`
- Table: Tenant, Plan, Amount, Status, Due Date, Actions
- Export buttons (CSV, PDF, XLSX)
- Payment status indicators
- Filter by status (paid, pending, overdue)

**Backend**:
- `internal/api/handlers/billing.go`
- `internal/models/billing.go`
- `internal/services/billing_service.go`

**Database**:
```sql
-- migrations/011_billing_system.sql
CREATE TABLE yt_billing (
    id SERIAL PRIMARY KEY,
    tenant_id INT REFERENCES yt_customers(tenant_id),
    plan VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    due_date DATE NOT NULL,
    paid_date DATE,
    invoice_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_billing_tenant ON yt_billing(tenant_id);
CREATE INDEX idx_billing_status ON yt_billing(status);
CREATE INDEX idx_billing_due_date ON yt_billing(due_date);
```

**API Endpoints**:
```
GET    /api/admin/billing              # List all billing records
GET    /api/admin/billing/:id          # Get billing details
POST   /api/admin/billing              # Create billing record
PUT    /api/admin/billing/:id          # Update billing record
DELETE /api/admin/billing/:id          # Delete billing record
GET    /api/admin/billing/export       # Export billing data (CSV/PDF/XLSX)
POST   /api/admin/billing/:id/pay      # Mark as paid
```

---

#### Week 2: Subscriptions Management
**Goal**: Manage customer subscriptions

**Frontend**:
- `frontend/src/pages/Admin/AdminSubscriptions.tsx`
- Table: Tenant, Plan, Start Date, End Date, Status, Actions
- Upgrade/downgrade modal
- Extend subscription modal
- Cancel subscription confirmation

**Backend**:
- `internal/api/handlers/subscriptions.go`
- `internal/models/subscription.go`
- `internal/services/subscription_service.go`

**Database**:
```sql
-- migrations/012_subscriptions.sql
CREATE TABLE yt_subscriptions (
    id SERIAL PRIMARY KEY,
    tenant_id INT REFERENCES yt_customers(tenant_id) UNIQUE,
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    auto_renew BOOLEAN DEFAULT true,
    trial_end_date DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE yt_subscription_history (
    id SERIAL PRIMARY KEY,
    subscription_id INT REFERENCES yt_subscriptions(id),
    action VARCHAR(50) NOT NULL,
    old_plan VARCHAR(50),
    new_plan VARCHAR(50),
    performed_by INT REFERENCES yt_admin_users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_tenant ON yt_subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_status ON yt_subscriptions(status);
CREATE INDEX idx_subscriptions_end_date ON yt_subscriptions(end_date);
```

**API Endpoints**:
```
GET    /api/admin/subscriptions                    # List all subscriptions
GET    /api/admin/subscriptions/:id                # Get subscription details
POST   /api/admin/subscriptions                    # Create subscription
PUT    /api/admin/subscriptions/:id/upgrade        # Upgrade plan
PUT    /api/admin/subscriptions/:id/downgrade      # Downgrade plan
PUT    /api/admin/subscriptions/:id/extend         # Extend subscription
DELETE /api/admin/subscriptions/:id                # Cancel subscription
GET    /api/admin/subscriptions/:id/history        # Get subscription history
```

---

#### Week 3: Payment Gateways
**Goal**: Configure payment providers

**Frontend**:
- `frontend/src/pages/Admin/AdminPaymentGateways.tsx`
- Add gateway form (Name, Provider, API Key, API URL, Test/Live toggle)
- List of configured gateways
- Test connection button
- Enable/disable toggle

**Backend**:
- `internal/api/handlers/payment_gateways.go`
- `internal/models/payment_gateway.go`
- `internal/services/payment_service.go`
- `internal/integrations/stripe.go`
- `internal/integrations/razorpay.go`

**Database**:
```sql
-- migrations/013_payment_gateways.sql
CREATE TABLE yt_payment_gateways (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    api_secret_encrypted TEXT,
    webhook_secret_encrypted TEXT,
    mode VARCHAR(10) DEFAULT 'test',
    enabled BOOLEAN DEFAULT true,
    config JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE yt_payment_transactions (
    id SERIAL PRIMARY KEY,
    billing_id INT REFERENCES yt_billing(id),
    gateway_id INT REFERENCES yt_payment_gateways(id),
    transaction_id VARCHAR(255) UNIQUE,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    gateway_response JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_transactions_billing ON yt_payment_transactions(billing_id);
CREATE INDEX idx_transactions_status ON yt_payment_transactions(status);
```

**API Endpoints**:
```
GET    /api/admin/payment-gateways              # List gateways
GET    /api/admin/payment-gateways/:id          # Get gateway details
POST   /api/admin/payment-gateways              # Add gateway
PUT    /api/admin/payment-gateways/:id          # Update gateway
DELETE /api/admin/payment-gateways/:id          # Remove gateway
POST   /api/admin/payment-gateways/:id/test     # Test connection
PUT    /api/admin/payment-gateways/:id/toggle   # Enable/disable
```

---

### **Phase 2: Analytics** (1 week) - MEDIUM PRIORITY

#### Week 4: Revenue Analytics
**Goal**: Track revenue metrics (MRR, ARPU, churn)

**Frontend**:
- `frontend/src/pages/Admin/AdminRevenue.tsx`
- Revenue trend chart (line chart)
- Subscription growth chart (bar chart)
- Key metrics cards (MRR, ARPU, Churn, LTV)
- Subscription status pie chart

**Backend**:
- `internal/api/handlers/revenue.go`
- `internal/services/revenue_service.go`

**Database**:
```sql
-- migrations/014_revenue_metrics.sql
CREATE TABLE yt_revenue_metrics (
    id SERIAL PRIMARY KEY,
    metric_date DATE NOT NULL UNIQUE,
    mrr DECIMAL(10,2) DEFAULT 0,
    arr DECIMAL(10,2) DEFAULT 0,
    arpu DECIMAL(10,2) DEFAULT 0,
    churn_rate DECIMAL(5,2) DEFAULT 0,
    new_customers INT DEFAULT 0,
    churned_customers INT DEFAULT 0,
    total_customers INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_revenue_metrics_date ON yt_revenue_metrics(metric_date);
```

**API Endpoints**:
```
GET /api/admin/revenue/metrics          # Get current metrics (MRR, ARPU, churn)
GET /api/admin/revenue/trend            # Get revenue trend (last 12 months)
GET /api/admin/revenue/subscriptions    # Get subscription breakdown
GET /api/admin/revenue/growth           # Get growth metrics
```

---

### **Phase 3: Visualization** (1 week) - LOW PRIORITY

#### Week 5: Infrastructure Visualization
**Goal**: Visualize VPC, subnets, services

**Frontend**:
- `frontend/src/pages/Admin/AdminInfrastructure.tsx`
- Two-column layout (Private/Public subnets)
- Service cards with status indicators
- Interactive diagram (click for details)

**Backend**:
- `internal/api/handlers/infrastructure.go`
- `internal/services/infrastructure_service.go`

**Database**:
```sql
-- migrations/015_infrastructure.sql
CREATE TABLE yt_infrastructure_services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    subnet_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'running',
    ip_address VARCHAR(50),
    port INT,
    health_check_url TEXT,
    last_health_check TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_infra_services_type ON yt_infrastructure_services(type);
CREATE INDEX idx_infra_services_status ON yt_infrastructure_services(status);
```

**API Endpoints**:
```
GET /api/admin/infrastructure              # Get all services
GET /api/admin/infrastructure/:id          # Get service details
GET /api/admin/infrastructure/health       # Health check all services
PUT /api/admin/infrastructure/:id/restart  # Restart service
```

---

#### Week 6: System Architecture Diagram
**Goal**: Show system architecture flowchart

**Frontend**:
- `frontend/src/pages/Admin/AdminArchitecture.tsx`
- React Flow diagram
- Zoom in/out controls
- Export as PNG/PDF

**Backend**:
- Static page (no backend needed)

**API Endpoints**:
- None (static diagram)

---

### **Phase 4: Advanced** (2 weeks) - OPTIONAL

#### Week 7-8: Invoice Generation
**Goal**: Generate and manage invoices

**Frontend**:
- `frontend/src/pages/Admin/AdminInvoices.tsx`
- Invoice list table
- Generate invoice modal
- Invoice preview
- Download PDF button
- Send email button

**Backend**:
- `internal/api/handlers/invoices.go`
- `internal/services/invoice_service.go`
- `internal/services/pdf_service.go`

**Database**:
```sql
-- migrations/016_invoices.sql
CREATE TABLE yt_invoices (
    id SERIAL PRIMARY KEY,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    billing_id INT REFERENCES yt_billing(id),
    tenant_id INT REFERENCES yt_customers(tenant_id),
    amount DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    pdf_url TEXT,
    sent_at TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_invoices_tenant ON yt_invoices(tenant_id);
CREATE INDEX idx_invoices_status ON yt_invoices(status);
CREATE INDEX idx_invoices_number ON yt_invoices(invoice_number);
```

**API Endpoints**:
```
GET    /api/admin/invoices                # List invoices
GET    /api/admin/invoices/:id            # Get invoice details
POST   /api/admin/invoices/generate       # Generate invoice
GET    /api/admin/invoices/:id/pdf        # Download PDF
POST   /api/admin/invoices/:id/send       # Email invoice
PUT    /api/admin/invoices/:id/mark-paid  # Mark as paid
```

---

## 🗺️ Complete Route Structure

### **Frontend Routes** (Add to `App.tsx`)

```typescript
// Admin Routes (Protected)
<Route path="/admin" element={<AdminLayout />}>
  {/* Existing */}
  <Route path="login" element={<AdminLogin />} />
  <Route path="dashboard" element={<AdminDashboard />} />
  <Route path="tenants" element={<AdminTenants />} />
  <Route path="users" element={<AdminUsers />} />
  <Route path="analytics" element={<AdminAnalytics />} />
  
  {/* NEW - Phase 1 */}
  <Route path="billing" element={<AdminBilling />} />
  <Route path="subscriptions" element={<AdminSubscriptions />} />
  <Route path="payment-gateways" element={<AdminPaymentGateways />} />
  
  {/* NEW - Phase 2 */}
  <Route path="revenue" element={<AdminRevenue />} />
  
  {/* NEW - Phase 3 */}
  <Route path="infrastructure" element={<AdminInfrastructure />} />
  <Route path="architecture" element={<AdminArchitecture />} />
  
  {/* NEW - Phase 4 */}
  <Route path="invoices" element={<AdminInvoices />} />
</Route>
```

### **Backend Routes** (Add to `routes.go`)

```go
// Admin routes (existing)
adminRoutes := router.Group("/api/admin")
adminRoutes.Use(adminAuthMw.RequireAdmin())
{
    // Existing
    adminRoutes.POST("/login", adminAuthHandler.Login)
    adminRoutes.GET("/stats", adminAnalyticsHandler.GetPlatformStats)
    adminRoutes.GET("/tenants", adminTenantsHandler.ListTenants)
    adminRoutes.GET("/users", adminImpersonationHandler.ListUsers)
    adminRoutes.GET("/analytics", adminAnalyticsHandler.GetAnalytics)
    
    // NEW - Phase 1: Billing
    adminRoutes.GET("/billing", billingHandler.ListBilling)
    adminRoutes.GET("/billing/:id", billingHandler.GetBilling)
    adminRoutes.POST("/billing", billingHandler.CreateBilling)
    adminRoutes.PUT("/billing/:id", billingHandler.UpdateBilling)
    adminRoutes.DELETE("/billing/:id", billingHandler.DeleteBilling)
    adminRoutes.GET("/billing/export", billingHandler.ExportBilling)
    adminRoutes.POST("/billing/:id/pay", billingHandler.MarkAsPaid)
    
    // NEW - Phase 1: Subscriptions
    adminRoutes.GET("/subscriptions", subscriptionsHandler.ListSubscriptions)
    adminRoutes.GET("/subscriptions/:id", subscriptionsHandler.GetSubscription)
    adminRoutes.POST("/subscriptions", subscriptionsHandler.CreateSubscription)
    adminRoutes.PUT("/subscriptions/:id/upgrade", subscriptionsHandler.UpgradePlan)
    adminRoutes.PUT("/subscriptions/:id/downgrade", subscriptionsHandler.DowngradePlan)
    adminRoutes.PUT("/subscriptions/:id/extend", subscriptionsHandler.ExtendSubscription)
    adminRoutes.DELETE("/subscriptions/:id", subscriptionsHandler.CancelSubscription)
    adminRoutes.GET("/subscriptions/:id/history", subscriptionsHandler.GetHistory)
    
    // NEW - Phase 1: Payment Gateways
    adminRoutes.GET("/payment-gateways", paymentGatewaysHandler.ListGateways)
    adminRoutes.GET("/payment-gateways/:id", paymentGatewaysHandler.GetGateway)
    adminRoutes.POST("/payment-gateways", paymentGatewaysHandler.CreateGateway)
    adminRoutes.PUT("/payment-gateways/:id", paymentGatewaysHandler.UpdateGateway)
    adminRoutes.DELETE("/payment-gateways/:id", paymentGatewaysHandler.DeleteGateway)
    adminRoutes.POST("/payment-gateways/:id/test", paymentGatewaysHandler.TestConnection)
    adminRoutes.PUT("/payment-gateways/:id/toggle", paymentGatewaysHandler.ToggleGateway)
    
    // NEW - Phase 2: Revenue
    adminRoutes.GET("/revenue/metrics", revenueHandler.GetMetrics)
    adminRoutes.GET("/revenue/trend", revenueHandler.GetTrend)
    adminRoutes.GET("/revenue/subscriptions", revenueHandler.GetSubscriptionBreakdown)
    adminRoutes.GET("/revenue/growth", revenueHandler.GetGrowth)
    
    // NEW - Phase 3: Infrastructure
    adminRoutes.GET("/infrastructure", infrastructureHandler.ListServices)
    adminRoutes.GET("/infrastructure/:id", infrastructureHandler.GetService)
    adminRoutes.GET("/infrastructure/health", infrastructureHandler.HealthCheck)
    adminRoutes.PUT("/infrastructure/:id/restart", infrastructureHandler.RestartService)
    
    // NEW - Phase 4: Invoices
    adminRoutes.GET("/invoices", invoicesHandler.ListInvoices)
    adminRoutes.GET("/invoices/:id", invoicesHandler.GetInvoice)
    adminRoutes.POST("/invoices/generate", invoicesHandler.GenerateInvoice)
    adminRoutes.GET("/invoices/:id/pdf", invoicesHandler.DownloadPDF)
    adminRoutes.POST("/invoices/:id/send", invoicesHandler.SendEmail)
    adminRoutes.PUT("/invoices/:id/mark-paid", invoicesHandler.MarkAsPaid)
}
```

---

## 📊 Progress Tracking

| Phase | Feature | Status | ETA |
|-------|---------|--------|-----|
| 1 | Billing Dashboard | ⏳ Pending | Week 1 |
| 1 | Subscriptions | ⏳ Pending | Week 2 |
| 1 | Payment Gateways | ⏳ Pending | Week 3 |
| 2 | Revenue Analytics | ⏳ Pending | Week 4 |
| 3 | Infrastructure Viz | ⏳ Pending | Week 5 |
| 3 | Architecture Diagram | ⏳ Pending | Week 6 |
| 4 | Invoice Generation | ⏳ Pending | Week 7-8 |

---

## 🚀 Next Steps

1. **Start with Phase 1, Week 1**: Billing Dashboard
2. **Create database migration**: `migrations/011_billing_system.sql`
3. **Build backend**: Models, handlers, services
4. **Build frontend**: AdminBilling.tsx component
5. **Test end-to-end**: Create billing record, export data
6. **Move to Week 2**: Subscriptions Management

---

**Ready to start Week 1: Billing Dashboard?** 🚀

**Last Updated**: January 31, 2025  
**Status**: Planning complete - ready for implementation
