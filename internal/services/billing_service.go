package services

import (
	"database/sql"
	"fmt"
	"time"
	"yukti/internal/models"
)

// BillingService handles billing business logic
type BillingService struct {
	db *sql.DB
}

// NewBillingService creates a new billing service
func NewBillingService(db *sql.DB) *BillingService {
	return &BillingService{db: db}
}

// ListBillings returns paginated list of billing records
func (s *BillingService) ListBillings(page, limit int, status, tenantID string) ([]models.Billing, int, error) {
	offset := (page - 1) * limit
	
	query := `
		SELECT 
			b.id, b.tenant_id, b.plan, b.amount, b.currency, b.status,
			b.due_date, b.paid_date, b.invoice_url, b.notes,
			b.created_at, b.updated_at,
			c.company_name, c.email
		FROM yt_billing b
		LEFT JOIN yt_customers c ON b.tenant_id = c.tenant_id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1
	
	if status != "" {
		query += fmt.Sprintf(" AND b.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}
	
	if tenantID != "" {
		query += fmt.Sprintf(" AND b.tenant_id = $%d", argCount)
		args = append(args, tenantID)
		argCount++
	}
	
	query += " ORDER BY b.due_date DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	billings := []models.Billing{}
	for rows.Next() {
		var b models.Billing
		err := rows.Scan(
			&b.ID, &b.TenantID, &b.Plan, &b.Amount, &b.Currency, &b.Status,
			&b.DueDate, &b.PaidDate, &b.InvoiceURL, &b.Notes,
			&b.CreatedAt, &b.UpdatedAt,
			&b.CompanyName, &b.Email,
		)
		if err != nil {
			return nil, 0, err
		}
		billings = append(billings, b)
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM yt_billing WHERE 1=1"
	if status != "" {
		countQuery += " AND status = '" + status + "'"
	}
	if tenantID != "" {
		countQuery += " AND tenant_id = " + tenantID
	}
	
	var totalCount int
	err = s.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	return billings, totalCount, nil
}

// GetBilling returns a single billing record by ID
func (s *BillingService) GetBilling(id int) (*models.Billing, error) {
	query := `
		SELECT 
			b.id, b.tenant_id, b.plan, b.amount, b.currency, b.status,
			b.due_date, b.paid_date, b.invoice_url, b.notes,
			b.created_at, b.updated_at,
			c.company_name, c.email
		FROM yt_billing b
		LEFT JOIN yt_customers c ON b.tenant_id = c.tenant_id
		WHERE b.id = $1
	`
	
	var b models.Billing
	err := s.db.QueryRow(query, id).Scan(
		&b.ID, &b.TenantID, &b.Plan, &b.Amount, &b.Currency, &b.Status,
		&b.DueDate, &b.PaidDate, &b.InvoiceURL, &b.Notes,
		&b.CreatedAt, &b.UpdatedAt,
		&b.CompanyName, &b.Email,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("billing record not found")
	}
	if err != nil {
		return nil, err
	}
	
	return &b, nil
}

// CreateBilling creates a new billing record
func (s *BillingService) CreateBilling(req *models.CreateBillingRequest) (*models.Billing, error) {
	query := `
		INSERT INTO yt_billing (tenant_id, plan, amount, currency, due_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, plan, amount, currency, status, due_date, paid_date, invoice_url, notes, created_at, updated_at
	`
	
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	
	var b models.Billing
	err := s.db.QueryRow(
		query,
		req.TenantID, req.Plan, req.Amount, currency, req.DueDate, req.Notes,
	).Scan(
		&b.ID, &b.TenantID, &b.Plan, &b.Amount, &b.Currency, &b.Status,
		&b.DueDate, &b.PaidDate, &b.InvoiceURL, &b.Notes,
		&b.CreatedAt, &b.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &b, nil
}

// UpdateBilling updates an existing billing record
func (s *BillingService) UpdateBilling(id int, req *models.UpdateBillingRequest) (*models.Billing, error) {
	query := `
		UPDATE yt_billing
		SET plan = COALESCE(NULLIF($1, ''), plan),
		    amount = COALESCE(NULLIF($2, 0), amount),
		    status = COALESCE(NULLIF($3, ''), status),
		    due_date = COALESCE($4, due_date),
		    paid_date = $5,
		    invoice_url = COALESCE(NULLIF($6, ''), invoice_url),
		    notes = COALESCE(NULLIF($7, ''), notes)
		WHERE id = $8
		RETURNING id, tenant_id, plan, amount, currency, status, due_date, paid_date, invoice_url, notes, created_at, updated_at
	`
	
	var b models.Billing
	err := s.db.QueryRow(
		query,
		req.Plan, req.Amount, req.Status, req.DueDate, req.PaidDate, req.InvoiceURL, req.Notes, id,
	).Scan(
		&b.ID, &b.TenantID, &b.Plan, &b.Amount, &b.Currency, &b.Status,
		&b.DueDate, &b.PaidDate, &b.InvoiceURL, &b.Notes,
		&b.CreatedAt, &b.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("billing record not found")
	}
	if err != nil {
		return nil, err
	}
	
	return &b, nil
}

// DeleteBilling deletes a billing record
func (s *BillingService) DeleteBilling(id int) error {
	query := "DELETE FROM yt_billing WHERE id = $1"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return fmt.Errorf("billing record not found")
	}
	
	return nil
}

// MarkAsPaid marks a billing record as paid
func (s *BillingService) MarkAsPaid(id int) (*models.Billing, error) {
	now := time.Now()
	req := &models.UpdateBillingRequest{
		Status:   models.BillingStatusPaid,
		PaidDate: &now,
	}
	return s.UpdateBilling(id, req)
}

// GetBillingStats returns billing statistics
func (s *BillingService) GetBillingStats() (*models.BillingStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0) as pending_amount,
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN amount ELSE 0 END), 0) as overdue_amount,
			COALESCE(SUM(CASE WHEN status = 'paid' AND EXTRACT(MONTH FROM paid_date) = EXTRACT(MONTH FROM NOW()) THEN amount ELSE 0 END), 0) as paid_this_month,
			COUNT(*) as total_billings,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_billings,
			COUNT(CASE WHEN status = 'overdue' THEN 1 END) as overdue_billings
		FROM yt_billing
	`
	
	var stats models.BillingStats
	err := s.db.QueryRow(query).Scan(
		&stats.TotalRevenue,
		&stats.PendingAmount,
		&stats.OverdueAmount,
		&stats.PaidThisMonth,
		&stats.TotalBillings,
		&stats.PendingBillings,
		&stats.OverdueBillings,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}
