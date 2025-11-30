package models

import (
	"database/sql"
	"time"
)

// BillingCustomer represents a Stripe customer linked to a tenant
type BillingCustomer struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	TenantID        int       `gorm:"not null;uniqueIndex" json:"tenant_id"`
	StripeCustomerID string  `gorm:"type:varchar(255);not null;uniqueIndex" json:"stripe_customer_id"`
	Currency        string    `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	TaxID           string    `gorm:"type:varchar(100)" json:"tax_id,omitempty"`
	BillingEmail    string    `gorm:"type:text" json:"billing_email,omitempty"`
	CreatedAt       time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"default:now()" json:"updated_at"`
}

func (BillingCustomer) TableName() string {
	return "yt_billing_customers"
}

// StripeEvent represents a Stripe webhook event
type StripeEvent struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	StripeEventID string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"stripe_event_id"`
	EventType     string    `gorm:"type:varchar(100);not null" json:"event_type"`
	Payload       string    `gorm:"type:jsonb;not null" json:"payload"` // JSONB stored as string in GORM
	Processed     bool      `gorm:"default:false;not null" json:"processed"`
	ProcessedAt   *time.Time `gorm:"default:null" json:"processed_at,omitempty"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time `gorm:"default:now()" json:"created_at"`
}

func (StripeEvent) TableName() string {
	return "yt_stripe_events"
}

// BillingInvoice represents a Stripe invoice
type BillingInvoice struct {
	ID               int       `gorm:"primaryKey" json:"id"`
	TenantID         int       `gorm:"not null;index" json:"tenant_id"`
	StripeInvoiceID  string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"stripe_invoice_id"`
	StripeCustomerID string    `gorm:"type:varchar(255);not null" json:"stripe_customer_id"`
	AmountCents      int       `gorm:"not null" json:"amount_cents"`
	TaxCents         int       `gorm:"default:0;not null" json:"tax_cents"`
	Currency         string    `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	Paid             bool      `gorm:"default:false;not null" json:"paid"`
	PaidAt           *time.Time `gorm:"default:null" json:"paid_at,omitempty"`
	PDFURL           string    `gorm:"type:text" json:"pdf_url,omitempty"`
	InvoiceNumber    string    `gorm:"type:varchar(100)" json:"invoice_number,omitempty"`
	PeriodStart      *time.Time `gorm:"default:null" json:"period_start,omitempty"`
	PeriodEnd        *time.Time `gorm:"default:null" json:"period_end,omitempty"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
}

func (BillingInvoice) TableName() string {
	return "yt_billing_invoices"
}

// BillingProductPrice represents a Stripe price mapping
type BillingProductPrice struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	PlanSlug          string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"plan_slug"`
	StripePriceID     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"stripe_price_id"`
	StripeProductID   string    `gorm:"type:varchar(255)" json:"stripe_product_id"`
	Currency          string    `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	AmountCents       int       `gorm:"not null" json:"amount_cents"`
	RecurringInterval string    `gorm:"type:varchar(20);not null" json:"recurring_interval"`
	IsActive          bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt         time.Time `gorm:"default:now()" json:"updated_at"`
}

func (BillingProductPrice) TableName() string {
	return "yt_billing_products_prices"
}

// BillingSubscription represents an active subscription
type BillingSubscription struct {
	ID                  int       `gorm:"primaryKey" json:"id"`
	TenantID            int       `gorm:"not null;uniqueIndex" json:"tenant_id"`
	StripeSubscriptionID string   `gorm:"type:varchar(255);not null;uniqueIndex" json:"stripe_subscription_id"`
	StripeCustomerID    string    `gorm:"type:varchar(255);not null" json:"stripe_customer_id"`
	StripePriceID       string    `gorm:"type:varchar(255);not null" json:"stripe_price_id"`
	PlanSlug            string    `gorm:"type:varchar(50);not null" json:"plan_slug"`
	Status              string    `gorm:"type:varchar(50);not null" json:"status"`
	CurrentPeriodStart  *time.Time `gorm:"default:null" json:"current_period_start,omitempty"`
	CurrentPeriodEnd    *time.Time `gorm:"default:null" json:"current_period_end,omitempty"`
	CancelAtPeriodEnd   bool      `gorm:"default:false;not null" json:"cancel_at_period_end"`
	CanceledAt          *time.Time `gorm:"default:null" json:"canceled_at,omitempty"`
	CreatedAt           time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt           time.Time `gorm:"default:now()" json:"updated_at"`
}

func (BillingSubscription) TableName() string {
	return "yt_billing_subscriptions"
}

// GetBillingCustomerByTenant retrieves billing customer for a tenant
func GetBillingCustomerByTenant(db *sql.DB, tenantID int) (*BillingCustomer, error) {
	var customer BillingCustomer
	err := db.QueryRow(`
		SELECT id, tenant_id, stripe_customer_id, currency, tax_id, billing_email, created_at, updated_at
		FROM yt_billing_customers
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&customer.ID, &customer.TenantID, &customer.StripeCustomerID,
		&customer.Currency, &customer.TaxID, &customer.BillingEmail,
		&customer.CreatedAt, &customer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetBillingSubscriptionByTenant retrieves active subscription for a tenant
func GetBillingSubscriptionByTenant(db *sql.DB, tenantID int) (*BillingSubscription, error) {
	var sub BillingSubscription
	err := db.QueryRow(`
		SELECT id, tenant_id, stripe_subscription_id, stripe_customer_id, stripe_price_id,
		       plan_slug, status, current_period_start, current_period_end,
		       cancel_at_period_end, canceled_at, created_at, updated_at
		FROM yt_billing_subscriptions
		WHERE tenant_id = $1 AND status IN ('active', 'trialing')
	`, tenantID).Scan(
		&sub.ID, &sub.TenantID, &sub.StripeSubscriptionID, &sub.StripeCustomerID,
		&sub.StripePriceID, &sub.PlanSlug, &sub.Status, &sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetBillingProductPriceByPlan retrieves price for a plan
func GetBillingProductPriceByPlan(db *sql.DB, planSlug, currency string) (*BillingProductPrice, error) {
	var price BillingProductPrice
	err := db.QueryRow(`
		SELECT id, plan_slug, stripe_price_id, stripe_product_id, currency,
		       amount_cents, recurring_interval, is_active, created_at, updated_at
		FROM yt_billing_products_prices
		WHERE plan_slug = $1 AND currency = $2 AND is_active = true
	`, planSlug, currency).Scan(
		&price.ID, &price.PlanSlug, &price.StripePriceID, &price.StripeProductID,
		&price.Currency, &price.AmountCents, &price.RecurringInterval,
		&price.IsActive, &price.CreatedAt, &price.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

