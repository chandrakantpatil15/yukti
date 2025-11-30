package budget

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Budget struct {
	ID             string
	TenantID       string
	Name           string
	Amount         float64
	Period         string
	StartDate      time.Time
	EndDate        *time.Time
	AlertThreshold float64
	CurrentSpend   float64
	Status         string
	CreatedAt      time.Time
}

type BudgetAlert struct {
	ID           string
	BudgetID     string
	TenantID     string
	AlertType    string
	Threshold    float64
	CurrentSpend float64
	Message      string
	Triggered    bool
	CreatedAt    time.Time
}

func (s *Service) CreateBudget(ctx context.Context, budget *Budget) error {
	budget.ID = uuid.New().String()
	budget.CreatedAt = time.Now()
	budget.Status = "active"

	query := `
		INSERT INTO yt_budgets (id, tenant_id, name, amount, period, start_date, end_date, alert_threshold, current_spend, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := s.db.ExecContext(ctx, query,
		budget.ID, budget.TenantID, budget.Name, budget.Amount, budget.Period,
		budget.StartDate, budget.EndDate, budget.AlertThreshold, budget.CurrentSpend,
		budget.Status, budget.CreatedAt,
	)
	return err
}

func (s *Service) UpdateBudgetSpend(ctx context.Context, budgetID string, currentSpend float64) error {
	query := `UPDATE yt_budgets SET current_spend = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, currentSpend, budgetID)
	return err
}

func (s *Service) CheckBudgetAlerts(ctx context.Context, tenantID string) ([]BudgetAlert, error) {
	query := `
		SELECT id, name, amount, alert_threshold, current_spend
		FROM yt_budgets
		WHERE tenant_id = $1 AND status = 'active'
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []BudgetAlert
	for rows.Next() {
		var budget Budget
		if err := rows.Scan(&budget.ID, &budget.Name, &budget.Amount, &budget.AlertThreshold, &budget.CurrentSpend); err != nil {
			continue
		}

		percentUsed := (budget.CurrentSpend / budget.Amount) * 100
		if percentUsed >= budget.AlertThreshold {
			alert := BudgetAlert{
				ID:           uuid.New().String(),
				BudgetID:     budget.ID,
				TenantID:     tenantID,
				AlertType:    "threshold_exceeded",
				Threshold:    budget.AlertThreshold,
				CurrentSpend: budget.CurrentSpend,
				Message:      "Budget threshold exceeded",
				Triggered:    true,
				CreatedAt:    time.Now(),
			}
			alerts = append(alerts, alert)
			s.createAlert(ctx, &alert)
		}
	}

	return alerts, nil
}

func (s *Service) createAlert(ctx context.Context, alert *BudgetAlert) error {
	query := `
		INSERT INTO yt_budget_alerts (id, budget_id, tenant_id, alert_type, threshold, current_spend, message, triggered, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.ExecContext(ctx, query,
		alert.ID, alert.BudgetID, alert.TenantID, alert.AlertType,
		alert.Threshold, alert.CurrentSpend, alert.Message, alert.Triggered, alert.CreatedAt,
	)
	return err
}

func (s *Service) GetBudgets(ctx context.Context, tenantID string) ([]Budget, error) {
	query := `
		SELECT id, tenant_id, name, amount, period, start_date, end_date, alert_threshold, current_spend, status, created_at
		FROM yt_budgets
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		if err := rows.Scan(
			&budget.ID, &budget.TenantID, &budget.Name, &budget.Amount, &budget.Period,
			&budget.StartDate, &budget.EndDate, &budget.AlertThreshold, &budget.CurrentSpend,
			&budget.Status, &budget.CreatedAt,
		); err != nil {
			continue
		}
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

func (s *Service) GetAlerts(ctx context.Context, tenantID string) ([]BudgetAlert, error) {
	query := `
		SELECT id, budget_id, tenant_id, alert_type, threshold, current_spend, message, triggered, created_at
		FROM yt_budget_alerts
		WHERE tenant_id = $1 AND triggered = true
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []BudgetAlert
	for rows.Next() {
		var alert BudgetAlert
		if err := rows.Scan(
			&alert.ID, &alert.BudgetID, &alert.TenantID, &alert.AlertType,
			&alert.Threshold, &alert.CurrentSpend, &alert.Message, &alert.Triggered, &alert.CreatedAt,
		); err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}
