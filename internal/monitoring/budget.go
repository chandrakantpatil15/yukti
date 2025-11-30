package monitoring

import (
	"context"
	"fmt"
	"time"
)

// BudgetService provides budget tracking and forecasting
type BudgetService struct {
	budgets    map[string]*Budget
	dataSource DataSource
}

// NewBudgetService creates a new budget service
func NewBudgetService(ds DataSource) *BudgetService {
	return &BudgetService{
		budgets:    make(map[string]*Budget),
		dataSource: ds,
	}
}

// CreateBudget creates a new budget
func (bs *BudgetService) CreateBudget(budget *Budget) error {
	budget.ID = fmt.Sprintf("budget-%d", time.Now().Unix())
	budget.CreatedAt = time.Now()
	budget.Status = "active"
	bs.budgets[budget.ID] = budget
	return nil
}

// GetBudgetStatus returns current budget status
func (bs *BudgetService) GetBudgetStatus(ctx context.Context, budgetID string) (*BudgetStatus, error) {
	budget, exists := bs.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget not found")
	}

	currentSpend := bs.dataSource.GetCurrentMonthCost(ctx, budget.AccountID)
	
	status := &BudgetStatus{
		BudgetID:      budget.ID,
		BudgetName:    budget.Name,
		BudgetAmount:  budget.Amount,
		CurrentSpend:  currentSpend,
		RemainingBudget: budget.Amount - currentSpend,
		PercentUsed:   (currentSpend / budget.Amount) * 100,
		Period:        budget.Period,
		Status:        bs.calculateBudgetStatus(currentSpend, budget.Amount),
		UpdatedAt:     time.Now(),
	}

	// Calculate forecast
	daysElapsed := float64(time.Now().Day())
	daysInMonth := 30.0
	dailyAverage := currentSpend / daysElapsed
	status.ForecastedSpend = dailyAverage * daysInMonth
	status.ForecastedOverrun = status.ForecastedSpend - budget.Amount

	// Check if alerts should be triggered
	if status.PercentUsed >= 80 {
		status.AlertLevel = "critical"
	} else if status.PercentUsed >= 60 {
		status.AlertLevel = "warning"
	} else {
		status.AlertLevel = "normal"
	}

	return status, nil
}

// GetAllBudgets returns all budgets for an account
func (bs *BudgetService) GetAllBudgets(ctx context.Context, accountID string) ([]*BudgetStatus, error) {
	var statuses []*BudgetStatus

	for _, budget := range bs.budgets {
		if budget.AccountID == accountID {
			status, err := bs.GetBudgetStatus(ctx, budget.ID)
			if err != nil {
				continue
			}
			statuses = append(statuses, status)
		}
	}

	return statuses, nil
}

// GetBudgetForecast returns budget forecast for next N days
func (bs *BudgetService) GetBudgetForecast(ctx context.Context, budgetID string, days int) (*BudgetForecast, error) {
	budget, exists := bs.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget not found")
	}

	currentSpend := bs.dataSource.GetCurrentMonthCost(ctx, budget.AccountID)
	daysElapsed := float64(time.Now().Day())
	dailyAverage := currentSpend / daysElapsed

	forecast := &BudgetForecast{
		BudgetID:     budget.ID,
		BudgetAmount: budget.Amount,
		CurrentSpend: currentSpend,
		DailyAverage: dailyAverage,
		ForecastDays: days,
		Projections:  make([]ForecastPoint, days),
		GeneratedAt:  time.Now(),
	}

	// Generate daily projections
	for i := 0; i < days; i++ {
		projectedSpend := currentSpend + (dailyAverage * float64(i+1))
		forecast.Projections[i] = ForecastPoint{
			Day:            i + 1,
			ProjectedSpend: projectedSpend,
			BudgetRemaining: budget.Amount - projectedSpend,
			OverBudget:     projectedSpend > budget.Amount,
		}
	}

	return forecast, nil
}

// calculateBudgetStatus determines budget status
func (bs *BudgetService) calculateBudgetStatus(currentSpend, budgetAmount float64) string {
	percentUsed := (currentSpend / budgetAmount) * 100

	if percentUsed >= 100 {
		return "exceeded"
	} else if percentUsed >= 80 {
		return "critical"
	} else if percentUsed >= 60 {
		return "warning"
	}
	return "healthy"
}

// Budget represents a cost budget
type Budget struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	Period      string    `json:"period"` // monthly, quarterly, yearly
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// BudgetStatus represents current budget status
type BudgetStatus struct {
	BudgetID        string    `json:"budget_id"`
	BudgetName      string    `json:"budget_name"`
	BudgetAmount    float64   `json:"budget_amount"`
	CurrentSpend    float64   `json:"current_spend"`
	RemainingBudget float64   `json:"remaining_budget"`
	PercentUsed     float64   `json:"percent_used"`
	ForecastedSpend float64   `json:"forecasted_spend"`
	ForecastedOverrun float64 `json:"forecasted_overrun"`
	Period          string    `json:"period"`
	Status          string    `json:"status"`
	AlertLevel      string    `json:"alert_level"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BudgetForecast represents budget forecast
type BudgetForecast struct {
	BudgetID     string          `json:"budget_id"`
	BudgetAmount float64         `json:"budget_amount"`
	CurrentSpend float64         `json:"current_spend"`
	DailyAverage float64         `json:"daily_average"`
	ForecastDays int             `json:"forecast_days"`
	Projections  []ForecastPoint `json:"projections"`
	GeneratedAt  time.Time       `json:"generated_at"`
}

// ForecastPoint represents a single forecast data point
type ForecastPoint struct {
	Day             int     `json:"day"`
	ProjectedSpend  float64 `json:"projected_spend"`
	BudgetRemaining float64 `json:"budget_remaining"`
	OverBudget      bool    `json:"over_budget"`
}