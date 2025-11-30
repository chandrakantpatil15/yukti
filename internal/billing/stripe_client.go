package billing

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/event"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeClient wraps Stripe API operations
type StripeClient struct {
	apiKey        string
	webhookSecret string
	db            *sql.DB
}

// NewStripeClient creates a new Stripe client
func NewStripeClient(db *sql.DB) (*StripeClient, error) {
	apiKey := os.Getenv("STRIPE_SECRET_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY environment variable not set")
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Printf("[WARN] STRIPE_WEBHOOK_SECRET not set - webhook verification will fail")
	}

	stripe.Key = apiKey

	return &StripeClient{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		db:            db,
	}, nil
}

// CreateCustomer creates a Stripe customer for a tenant
func (c *StripeClient) CreateCustomer(tenantID int, email, companyName string) (string, error) {
	// Check if customer already exists
	var existingCustomerID string
	err := c.db.QueryRow(`
		SELECT stripe_customer_id FROM yt_billing_customers WHERE tenant_id = $1
	`, tenantID).Scan(&existingCustomerID)
	if err == nil {
		return existingCustomerID, nil
	}

	// Create Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(companyName),
		Metadata: map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenantID),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Store in database
	_, err = c.db.Exec(`
		INSERT INTO yt_billing_customers (tenant_id, stripe_customer_id, currency, billing_email)
		VALUES ($1, $2, 'USD', $3)
		ON CONFLICT (tenant_id) DO UPDATE
		SET stripe_customer_id = EXCLUDED.stripe_customer_id,
		    billing_email = EXCLUDED.billing_email,
		    updated_at = NOW()
	`, tenantID, cust.ID, email)
	if err != nil {
		// Rollback: delete Stripe customer if DB insert fails
		customer.Del(cust.ID, nil)
		return "", fmt.Errorf("failed to store customer in database: %w", err)
	}

	return cust.ID, nil
}

// CreateCheckoutSession creates a Stripe Checkout session
func (c *StripeClient) CreateCheckoutSession(tenantID int, priceID, currency, successURL, cancelURL string) (string, error) {
	// Get or create Stripe customer
	var stripeCustomerID string
	err := c.db.QueryRow(`
		SELECT stripe_customer_id FROM yt_billing_customers WHERE tenant_id = $1
	`, tenantID).Scan(&stripeCustomerID)
	if err != nil {
		// Customer doesn't exist - we'll create it during checkout
		// For now, we'll create it here
		var email, companyName string
		c.db.QueryRow(`
			SELECT email, company_name FROM yt_tenants t
			JOIN yt_users u ON u.tenant_id = t.id
			WHERE t.id = $1 AND u.role = 'admin'
			LIMIT 1
		`, tenantID).Scan(&email, &companyName)
		
		stripeCustomerID, err = c.CreateCustomer(tenantID, email, companyName)
		if err != nil {
			return "", fmt.Errorf("failed to create customer: %w", err)
		}
	}

	params := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(stripeCustomerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenantID),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// VerifyWebhookSignature verifies a Stripe webhook signature
func (c *StripeClient) VerifyWebhookSignature(payload []byte, signature string) (*stripe.Event, error) {
	if c.webhookSecret == "" {
		return nil, fmt.Errorf("webhook secret not configured")
	}

	event, err := webhook.ConstructEvent(payload, signature, c.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	return &event, nil
}

// StoreStripeEvent stores a webhook event in the database
func (c *StripeClient) StoreStripeEvent(eventID, eventType string, payload []byte) error {
	_, err := c.db.Exec(`
		INSERT INTO yt_stripe_events (stripe_event_id, event_type, payload, processed)
		VALUES ($1, $2, $3::jsonb, false)
		ON CONFLICT (stripe_event_id) DO NOTHING
	`, eventID, eventType, string(payload))
	return err
}

// MarkEventProcessed marks an event as processed
func (c *StripeClient) MarkEventProcessed(eventID string, err error) error {
	if err != nil {
		_, dbErr := c.db.Exec(`
			UPDATE yt_stripe_events
			SET processed = true, processed_at = NOW(), error_message = $1
			WHERE stripe_event_id = $2
		`, err.Error(), eventID)
		return dbErr
	}

	_, dbErr := c.db.Exec(`
		UPDATE yt_stripe_events
		SET processed = true, processed_at = NOW()
		WHERE stripe_event_id = $1
	`, eventID)
	return dbErr
}

// ProcessInvoicePaid processes invoice.paid event
func (c *StripeClient) ProcessInvoicePaid(invoice *stripe.Invoice) error {
	// Extract tenant_id from customer metadata or lookup
	var tenantID int
	err := c.db.QueryRow(`
		SELECT tenant_id FROM yt_billing_customers
		WHERE stripe_customer_id = $1
	`, invoice.Customer.ID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Store invoice
	_, err = c.db.Exec(`
		INSERT INTO yt_billing_invoices (
			tenant_id, stripe_invoice_id, stripe_customer_id,
			amount_cents, tax_cents, currency, paid, paid_at,
			pdf_url, invoice_number, period_start, period_end
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (stripe_invoice_id) DO UPDATE
		SET paid = EXCLUDED.paid,
		    paid_at = EXCLUDED.paid_at,
		    pdf_url = EXCLUDED.pdf_url
	`,
		tenantID, invoice.ID, invoice.Customer.ID,
		invoice.AmountPaid, invoice.Tax, invoice.Currency,
		invoice.Paid, time.Unix(invoice.StatusTransitions.PaidAt, 0),
		invoice.InvoicePDF, invoice.Number,
		time.Unix(invoice.PeriodStart, 0),
		time.Unix(invoice.PeriodEnd, 0),
	)
	if err != nil {
		return fmt.Errorf("failed to store invoice: %w", err)
	}

	// Update tenant subscription tier if this is first payment
	// Get subscription to determine plan
	var planSlug string
	err = c.db.QueryRow(`
		SELECT plan_slug FROM yt_billing_subscriptions
		WHERE tenant_id = $1 AND status = 'active'
		LIMIT 1
	`, tenantID).Scan(&planSlug)
	if err == nil && planSlug != "" {
		// Update tenant tier
		_, err = c.db.Exec(`
			UPDATE yt_tenants
			SET subscription_tier = $1
			WHERE id = $2
		`, planSlug, tenantID)
		if err != nil {
			log.Printf("[ERROR] Failed to update tenant tier: %v", err)
		}
	}

	return nil
}

// ProcessSubscriptionUpdated processes subscription.updated event
func (c *StripeClient) ProcessSubscriptionUpdated(sub *stripe.Subscription) error {
	// Get tenant_id from customer
	var tenantID int
	err := c.db.QueryRow(`
		SELECT tenant_id FROM yt_billing_customers
		WHERE stripe_customer_id = $1
	`, sub.Customer.ID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Get plan_slug from price_id
	var planSlug string
	err = c.db.QueryRow(`
		SELECT plan_slug FROM yt_billing_products_prices
		WHERE stripe_price_id = $1
	`, sub.Items.Data[0].Price.ID).Scan(&planSlug)
	if err != nil {
		return fmt.Errorf("price not found: %w", err)
	}

	// Upsert subscription
	_, err = c.db.Exec(`
		INSERT INTO yt_billing_subscriptions (
			tenant_id, stripe_subscription_id, stripe_customer_id,
			stripe_price_id, plan_slug, status,
			current_period_start, current_period_end,
			cancel_at_period_end, canceled_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id) DO UPDATE
		SET stripe_subscription_id = EXCLUDED.stripe_subscription_id,
		    stripe_price_id = EXCLUDED.stripe_price_id,
		    plan_slug = EXCLUDED.plan_slug,
		    status = EXCLUDED.status,
		    current_period_start = EXCLUDED.current_period_start,
		    current_period_end = EXCLUDED.current_period_end,
		    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
		    canceled_at = EXCLUDED.canceled_at,
		    updated_at = NOW()
	`,
		tenantID, sub.ID, sub.Customer.ID,
		sub.Items.Data[0].Price.ID, planSlug, string(sub.Status),
		time.Unix(sub.CurrentPeriodStart, 0),
		time.Unix(sub.CurrentPeriodEnd, 0),
		sub.CancelAtPeriodEnd,
		func() *time.Time {
			if sub.CanceledAt > 0 {
				t := time.Unix(sub.CanceledAt, 0)
				return &t
			}
			return nil
		}(),
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update tenant tier
	_, err = c.db.Exec(`
		UPDATE yt_tenants
		SET subscription_tier = $1
		WHERE id = $2
	`, planSlug, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to update tenant tier: %v", err)
	}

	return nil
}

// ProcessSubscriptionCanceled processes subscription.deleted/canceled event
func (c *StripeClient) ProcessSubscriptionCanceled(sub *stripe.Subscription) error {
	// Get tenant_id
	var tenantID int
	err := c.db.QueryRow(`
		SELECT tenant_id FROM yt_billing_customers
		WHERE stripe_customer_id = $1
	`, sub.Customer.ID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Update subscription status
	_, err = c.db.Exec(`
		UPDATE yt_billing_subscriptions
		SET status = 'canceled',
		    canceled_at = NOW(),
		    updated_at = NOW()
		WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Downgrade tenant to FREE tier
	_, err = c.db.Exec(`
		UPDATE yt_tenants
		SET subscription_tier = 'FREE'
		WHERE id = $1
	`, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to downgrade tenant: %v", err)
	}

	return nil
}

// GetEvent retrieves a Stripe event by ID
func (c *StripeClient) GetEvent(eventID string) (*stripe.Event, error) {
	return event.Get(eventID, nil)
}

