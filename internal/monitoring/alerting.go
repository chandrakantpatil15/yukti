package monitoring

import (
	"context"
	"fmt"
	"time"
)

// AlertingService provides custom cost alerting capabilities
type AlertingService struct {
	rules      map[string]*AlertRule
	notifier   NotificationService
	dataSource DataSource
}

// NewAlertingService creates a new alerting service
func NewAlertingService(notifier NotificationService, ds DataSource) *AlertingService {
	return &AlertingService{
		rules:      make(map[string]*AlertRule),
		notifier:   notifier,
		dataSource: ds,
	}
}

// CreateAlertRule creates a new alert rule
func (as *AlertingService) CreateAlertRule(rule *AlertRule) error {
	rule.ID = fmt.Sprintf("rule-%d", time.Now().Unix())
	rule.CreatedAt = time.Now()
	rule.Enabled = true
	as.rules[rule.ID] = rule
	return nil
}

// EvaluateAlerts evaluates all alert rules
func (as *AlertingService) EvaluateAlerts(ctx context.Context, accountID string) ([]*Alert, error) {
	var triggeredAlerts []*Alert

	for _, rule := range as.rules {
		if !rule.Enabled || rule.AccountID != accountID {
			continue
		}

		alert := as.evaluateRule(ctx, rule)
		if alert != nil {
			triggeredAlerts = append(triggeredAlerts, alert)
			as.notifier.SendAlert(ctx, alert)
		}
	}

	return triggeredAlerts, nil
}

// evaluateRule evaluates a single alert rule
func (as *AlertingService) evaluateRule(ctx context.Context, rule *AlertRule) *Alert {
	var currentValue float64

	switch rule.Type {
	case "daily_cost":
		currentValue = as.dataSource.GetCurrentMonthCost(ctx, rule.AccountID) / float64(time.Now().Day())
	case "monthly_cost":
		currentValue = as.dataSource.GetCurrentMonthCost(ctx, rule.AccountID)
	case "service_cost":
		services := as.dataSource.GetServiceBreakdown(ctx, rule.AccountID)
		for _, svc := range services {
			if svc.ServiceName == rule.ServiceName {
				currentValue = svc.Cost
				break
			}
		}
	case "cost_spike":
		current := as.dataSource.GetCurrentMonthCost(ctx, rule.AccountID)
		previous := as.dataSource.GetPreviousMonthCost(ctx, rule.AccountID)
		if previous > 0 {
			currentValue = ((current - previous) / previous) * 100
		}
	}

	// Check if threshold is exceeded
	if as.checkThreshold(currentValue, rule.Threshold, rule.Operator) {
		return &Alert{
			ID:           fmt.Sprintf("alert-%d", time.Now().Unix()),
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			AccountID:    rule.AccountID,
			Severity:     rule.Severity,
			CurrentValue: currentValue,
			Threshold:    rule.Threshold,
			Message:      fmt.Sprintf("%s: Current value %.2f exceeds threshold %.2f", rule.Name, currentValue, rule.Threshold),
			TriggeredAt:  time.Now(),
			Status:       "active",
		}
	}

	return nil
}

// checkThreshold checks if value exceeds threshold based on operator
func (as *AlertingService) checkThreshold(value, threshold float64, operator string) bool {
	switch operator {
	case "greater_than":
		return value > threshold
	case "less_than":
		return value < threshold
	case "equals":
		return value == threshold
	default:
		return value > threshold
	}
}

// AlertRule represents a cost alert rule
type AlertRule struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // daily_cost, monthly_cost, service_cost, cost_spike
	Threshold   float64   `json:"threshold"`
	Operator    string    `json:"operator"` // greater_than, less_than, equals
	ServiceName string    `json:"service_name,omitempty"`
	Severity    string    `json:"severity"` // info, warning, critical
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID           string    `json:"id"`
	RuleID       string    `json:"rule_id"`
	RuleName     string    `json:"rule_name"`
	AccountID    string    `json:"account_id"`
	Severity     string    `json:"severity"`
	CurrentValue float64   `json:"current_value"`
	Threshold    float64   `json:"threshold"`
	Message      string    `json:"message"`
	TriggeredAt  time.Time `json:"triggered_at"`
	Status       string    `json:"status"` // active, acknowledged, resolved
}

// NotificationService interface for sending alerts
type NotificationService interface {
	SendAlert(ctx context.Context, alert *Alert) error
}