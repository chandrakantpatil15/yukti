package models

import "time"

// Billing represents a customer billing record
type Billing struct {
	ID         int       `json:"id" db:"id"`
	TenantID   int       `json:"tenant_id" db:"tenant_id"`
	Plan       string    `json:"plan" db:"plan"`
	Amount     float64   `json:"amount" db:"amount"`
	Currency   string    `json:"currency" db:"currency"`
	Status     string    `json:"status" db:"status"`
	DueDate    time.Time `json:"due_date" db:"due_date"`
	PaidDate   *time.Time `json:"paid_date,omitempty" db:"paid_date"`
	InvoiceURL string    `json:"invoice_url,omitempty" db:"invoice_url"`
	Notes      string    `json:"notes,omitempty" db:"notes"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	CompanyName string `json:"company_name,omitempty" db:"company_name"`
	Email       string `json:"email,omitempty" db:"email"`
}

// BillingStatus constants
const (
	BillingStatusPending   = "pending"
	BillingStatusPaid      = "paid"
	BillingStatusOverdue   = "overdue"
	BillingStatusCancelled = "cancelled"
)

// BillingPlan constants
const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// CreateBillingRequest represents request to create billing record
type CreateBillingRequest struct {
	TenantID int       `json:"tenant_id" binding:"required"`
	Plan     string    `json:"plan" binding:"required"`
	Amount   float64   `json:"amount" binding:"required"`
	Currency string    `json:"currency"`
	DueDate  time.Time `json:"due_date" binding:"required"`
	Notes    string    `json:"notes"`
}

// UpdateBillingRequest represents request to update billing record
type UpdateBillingRequest struct {
	Plan       string     `json:"plan"`
	Amount     float64    `json:"amount"`
	Status     string     `json:"status"`
	DueDate    *time.Time `json:"due_date"`
	PaidDate   *time.Time `json:"paid_date"`
	InvoiceURL string     `json:"invoice_url"`
	Notes      string     `json:"notes"`
}

// BillingListResponse represents paginated billing list
type BillingListResponse struct {
	Billings   []Billing `json:"billings"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
}

// BillingStats represents billing statistics
type BillingStats struct {
	TotalRevenue    float64 `json:"total_revenue"`
	PendingAmount   float64 `json:"pending_amount"`
	OverdueAmount   float64 `json:"overdue_amount"`
	PaidThisMonth   float64 `json:"paid_this_month"`
	TotalBillings   int     `json:"total_billings"`
	PendingBillings int     `json:"pending_billings"`
	OverdueBillings int     `json:"overdue_billings"`
}
