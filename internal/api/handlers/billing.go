package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"yukti/internal/billing"
	"yukti/internal/models"
)

type BillingHandler struct {
	db           *sql.DB
	stripeClient *billing.StripeClient
}

func NewBillingHandler(db *sql.DB) (*BillingHandler, error) {
	client, err := billing.NewStripeClient(db)
	if err != nil {
		return nil, err
	}
	return &BillingHandler{
		db:           db,
		stripeClient: client,
	}, nil
}

// CreateCheckoutSessionRequest represents checkout session creation request
type CreateCheckoutSessionRequest struct {
	PlanSlug string `json:"plan_slug" validate:"required,oneof=professional enterprise financial"`
	Currency string `json:"currency,omitempty" validate:"omitempty,oneof=USD EUR GBP INR"`
}

// CreateCheckoutSessionResponse represents checkout session response
type CreateCheckoutSessionResponse struct {
	Success   bool   `json:"success"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CreateCheckoutSession creates a Stripe Checkout session
// POST /api/v1/billing/checkout-session
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] CreateCheckoutSession request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get tenant_id from context (set by JWT middleware)
	tenantID, ok := r.Context().Value("tenant_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(CreateCheckoutSessionResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	var req CreateCheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateCheckoutSessionResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.PlanSlug == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateCheckoutSessionResponse{
			Success: false,
			Error:   "plan_slug is required",
		})
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Get price for plan
	price, err := models.GetBillingProductPriceByPlan(h.db, req.PlanSlug, req.Currency)
	if err != nil {
		log.Printf("[ERROR] Failed to get price for plan %s: %v", req.PlanSlug, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateCheckoutSessionResponse{
			Success: false,
			Error:   "Invalid plan or currency",
		})
		return
	}

	// Build success and cancel URLs
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	successURL := baseURL + "/billing/success?session_id={CHECKOUT_SESSION_ID}"
	cancelURL := baseURL + "/billing"

	// Create checkout session
	checkoutURL, err := h.stripeClient.CreateCheckoutSession(
		tenantID,
		price.StripePriceID,
		req.Currency,
		successURL,
		cancelURL,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create checkout session: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateCheckoutSessionResponse{
			Success: false,
			Error:   "Failed to create checkout session",
		})
		return
	}

	log.Printf("[INFO] Checkout session created for tenant %d, plan %s", tenantID, req.PlanSlug)
	json.NewEncoder(w).Encode(CreateCheckoutSessionResponse{
		Success:    true,
		CheckoutURL: checkoutURL,
	})
}

// HandleStripeWebhook processes Stripe webhook events
// POST /api/v1/webhooks/stripe
func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Stripe webhook received from IP: %s", r.RemoteAddr)

	// Read raw body for signature verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read webhook body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get signature from header
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		log.Printf("[WARN] Missing Stripe-Signature header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	event, err := h.stripeClient.VerifyWebhookSignature(body, signature)
	if err != nil {
		log.Printf("[ERROR] Webhook signature verification failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Store event in database
	err = h.stripeClient.StoreStripeEvent(event.ID, event.Type, body)
	if err != nil {
		log.Printf("[ERROR] Failed to store webhook event: %v", err)
		// Continue processing even if storage fails
	}

	// Process event
	var processErr error
	switch event.Type {
	case "invoice.paid":
		invoice := &event.Data.Object
		// Parse invoice from event data
		invoiceBytes, _ := json.Marshal(invoice)
		var invoiceObj map[string]interface{}
		json.Unmarshal(invoiceBytes, &invoiceObj)
		// Note: In production, use stripe.Invoice type properly
		log.Printf("[INFO] Processing invoice.paid event: %s", event.ID)
		// processErr = h.stripeClient.ProcessInvoicePaid(invoice)
		// TODO: Properly parse invoice from event

	case "customer.subscription.updated":
		log.Printf("[INFO] Processing subscription.updated event: %s", event.ID)
		// processErr = h.stripeClient.ProcessSubscriptionUpdated(sub)

	case "customer.subscription.deleted":
		log.Printf("[INFO] Processing subscription.deleted event: %s", event.ID)
		// processErr = h.stripeClient.ProcessSubscriptionCanceled(sub)

	default:
		log.Printf("[INFO] Unhandled event type: %s", event.Type)
	}

	// Mark event as processed
	h.stripeClient.MarkEventProcessed(event.ID, processErr)
	if processErr != nil {
		log.Printf("[ERROR] Failed to process event %s: %v", event.ID, processErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"event_id": event.ID,
	})
}

// GetBillingInfo returns current billing information for tenant
// GET /api/v1/billing/info
func (h *BillingHandler) GetBillingInfo(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetBillingInfo request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	tenantID, ok := r.Context().Value("tenant_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Get tenant info
	var tier string
	var trialEndsAt *string
	err := h.db.QueryRow(`
		SELECT subscription_tier, trial_ends_at::text
		FROM yt_tenants WHERE id = $1
	`, tenantID).Scan(&tier, &trialEndsAt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to get tenant info",
		})
		return
	}

	// Get subscription if exists
	sub, _ := models.GetBillingSubscriptionByTenant(h.db, tenantID)

	// Get recent invoices
	rows, _ := h.db.Query(`
		SELECT id, stripe_invoice_id, amount_cents, tax_cents, currency,
		       paid, paid_at, pdf_url, invoice_number, created_at
		FROM yt_billing_invoices
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`, tenantID)
	defer rows.Close()

	invoices := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var invoiceID, currency, pdfURL, invoiceNumber string
		var amountCents, taxCents int
		var paid bool
		var paidAt, createdAt *string
		rows.Scan(&id, &invoiceID, &amountCents, &taxCents, &currency,
			&paid, &paidAt, &pdfURL, &invoiceNumber, &createdAt)
		invoices = append(invoices, map[string]interface{}{
			"id":            id,
			"invoice_id":    invoiceID,
			"amount_cents":  amountCents,
			"tax_cents":     taxCents,
			"currency":      currency,
			"paid":          paid,
			"paid_at":       paidAt,
			"pdf_url":       pdfURL,
			"invoice_number": invoiceNumber,
			"created_at":    createdAt,
		})
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"subscription_tier": tier,
			"trial_ends_at":     trialEndsAt,
			"subscription":      sub,
			"invoices":          invoices,
		},
	}

	json.NewEncoder(w).Encode(response)
}

