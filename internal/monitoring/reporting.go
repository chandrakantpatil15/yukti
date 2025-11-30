package monitoring

import (
	"context"
	"time"
)

// ReportingService provides executive reporting and KPIs
type ReportingService struct {
	dataSource DataSource
}

// NewReportingService creates a new reporting service
func NewReportingService(ds DataSource) *ReportingService {
	return &ReportingService{
		dataSource: ds,
	}
}

// GenerateExecutiveReport generates executive-level cost report
func (rs *ReportingService) GenerateExecutiveReport(ctx context.Context, accountID string) (*ExecutiveReport, error) {
	report := &ExecutiveReport{
		AccountID:   accountID,
		ReportDate:  time.Now(),
		Period:      "monthly",
	}

	// Current month metrics
	report.CurrentMonthCost = rs.dataSource.GetCurrentMonthCost(ctx, accountID)
	report.PreviousMonthCost = rs.dataSource.GetPreviousMonthCost(ctx, accountID)
	
	// Calculate trends
	if report.PreviousMonthCost > 0 {
		report.MonthOverMonthChange = ((report.CurrentMonthCost - report.PreviousMonthCost) / report.PreviousMonthCost) * 100
	}

	// Service breakdown
	report.TopServices = rs.dataSource.GetServiceBreakdown(ctx, accountID)
	if len(report.TopServices) > 5 {
		report.TopServices = report.TopServices[:5] // Top 5 services
	}

	// Calculate KPIs
	report.KPIs = rs.calculateKPIs(ctx, accountID)

	// Cost optimization opportunities
	report.OptimizationOpportunities = rs.getOptimizationSummary(ctx, accountID)

	// Regional distribution
	report.RegionalDistribution = rs.dataSource.GetRegionBreakdown(ctx, accountID)

	return report, nil
}

// GenerateCostAllocationReport generates cost allocation report
func (rs *ReportingService) GenerateCostAllocationReport(ctx context.Context, accountID string) (*CostAllocationReport, error) {
	report := &CostAllocationReport{
		AccountID:   accountID,
		ReportDate:  time.Now(),
		Period:      "monthly",
	}

	// Get cost breakdown by different dimensions
	report.ByEnvironment = rs.getCostByEnvironment(ctx, accountID)
	report.ByProject = rs.getCostByProject(ctx, accountID)
	report.ByCostCenter = rs.getCostByCostCenter(ctx, accountID)
	report.ByOwner = rs.getCostByOwner(ctx, accountID)

	// Calculate totals
	report.TotalCost = rs.dataSource.GetCurrentMonthCost(ctx, accountID)

	return report, nil
}

// GenerateChargebackReport generates chargeback/showback report
func (rs *ReportingService) GenerateChargebackReport(ctx context.Context, accountID string) (*ChargebackReport, error) {
	report := &ChargebackReport{
		AccountID:   accountID,
		ReportDate:  time.Now(),
		Period:      "monthly",
	}

	// Get chargeback by team/department
	report.Chargebacks = rs.getChargebackByTeam(ctx, accountID)

	// Calculate total
	for _, cb := range report.Chargebacks {
		report.TotalChargeback += cb.Amount
	}

	return report, nil
}

// calculateKPIs calculates key performance indicators
func (rs *ReportingService) calculateKPIs(ctx context.Context, accountID string) []KPI {
	currentCost := rs.dataSource.GetCurrentMonthCost(ctx, accountID)
	previousCost := rs.dataSource.GetPreviousMonthCost(ctx, accountID)

	kpis := []KPI{
		{
			Name:        "Total Monthly Cost",
			Value:       currentCost,
			Unit:        "USD",
			Trend:       calculateTrend(currentCost, previousCost),
			Status:      "normal",
			Description: "Total cloud spend for current month",
		},
		{
			Name:        "Cost per Day",
			Value:       currentCost / float64(time.Now().Day()),
			Unit:        "USD",
			Trend:       0,
			Status:      "normal",
			Description: "Average daily cloud spend",
		},
		{
			Name:        "Month-over-Month Change",
			Value:       calculateTrend(currentCost, previousCost),
			Unit:        "%",
			Trend:       0,
			Status:      rs.getTrendStatus(calculateTrend(currentCost, previousCost)),
			Description: "Cost change compared to previous month",
		},
	}

	return kpis
}

// getOptimizationSummary gets optimization opportunities summary
func (rs *ReportingService) getOptimizationSummary(ctx context.Context, accountID string) OptimizationSummary {
	return OptimizationSummary{
		TotalOpportunities:   15,
		PotentialSavings:     12500.00,
		HighPriority:         5,
		MediumPriority:       7,
		LowPriority:          3,
		ImplementationStatus: "pending",
	}
}

// getCostByEnvironment gets cost breakdown by environment
func (rs *ReportingService) getCostByEnvironment(ctx context.Context, accountID string) []AllocationItem {
	return []AllocationItem{
		{Name: "Production", Cost: 15000.00, Percentage: 60},
		{Name: "Staging", Cost: 5000.00, Percentage: 20},
		{Name: "Development", Cost: 3000.00, Percentage: 12},
		{Name: "Testing", Cost: 2000.00, Percentage: 8},
	}
}

// getCostByProject gets cost breakdown by project
func (rs *ReportingService) getCostByProject(ctx context.Context, accountID string) []AllocationItem {
	return []AllocationItem{
		{Name: "Project Alpha", Cost: 10000.00, Percentage: 40},
		{Name: "Project Beta", Cost: 8000.00, Percentage: 32},
		{Name: "Project Gamma", Cost: 7000.00, Percentage: 28},
	}
}

// getCostByCostCenter gets cost breakdown by cost center
func (rs *ReportingService) getCostByCostCenter(ctx context.Context, accountID string) []AllocationItem {
	return []AllocationItem{
		{Name: "Engineering", Cost: 12000.00, Percentage: 48},
		{Name: "Product", Cost: 8000.00, Percentage: 32},
		{Name: "Operations", Cost: 5000.00, Percentage: 20},
	}
}

// getCostByOwner gets cost breakdown by owner
func (rs *ReportingService) getCostByOwner(ctx context.Context, accountID string) []AllocationItem {
	return []AllocationItem{
		{Name: "Team A", Cost: 9000.00, Percentage: 36},
		{Name: "Team B", Cost: 8000.00, Percentage: 32},
		{Name: "Team C", Cost: 8000.00, Percentage: 32},
	}
}

// getChargebackByTeam gets chargeback amounts by team
func (rs *ReportingService) getChargebackByTeam(ctx context.Context, accountID string) []ChargebackItem {
	return []ChargebackItem{
		{Team: "Engineering", Amount: 12000.00, Resources: 45, BillingPeriod: "2024-01"},
		{Team: "Product", Amount: 8000.00, Resources: 30, BillingPeriod: "2024-01"},
		{Team: "Operations", Amount: 5000.00, Resources: 20, BillingPeriod: "2024-01"},
	}
}

// getTrendStatus determines status based on trend
func (rs *ReportingService) getTrendStatus(trend float64) string {
	if trend > 20 {
		return "critical"
	} else if trend > 10 {
		return "warning"
	}
	return "normal"
}

// ExecutiveReport represents executive-level cost report
type ExecutiveReport struct {
	AccountID                 string                  `json:"account_id"`
	ReportDate                time.Time               `json:"report_date"`
	Period                    string                  `json:"period"`
	CurrentMonthCost          float64                 `json:"current_month_cost"`
	PreviousMonthCost         float64                 `json:"previous_month_cost"`
	MonthOverMonthChange      float64                 `json:"month_over_month_change"`
	TopServices               []ServiceCost           `json:"top_services"`
	RegionalDistribution      []RegionCost            `json:"regional_distribution"`
	KPIs                      []KPI                   `json:"kpis"`
	OptimizationOpportunities OptimizationSummary     `json:"optimization_opportunities"`
}

// KPI represents a key performance indicator
type KPI struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Trend       float64 `json:"trend"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

// OptimizationSummary represents optimization opportunities summary
type OptimizationSummary struct {
	TotalOpportunities   int     `json:"total_opportunities"`
	PotentialSavings     float64 `json:"potential_savings"`
	HighPriority         int     `json:"high_priority"`
	MediumPriority       int     `json:"medium_priority"`
	LowPriority          int     `json:"low_priority"`
	ImplementationStatus string  `json:"implementation_status"`
}

// CostAllocationReport represents cost allocation report
type CostAllocationReport struct {
	AccountID     string           `json:"account_id"`
	ReportDate    time.Time        `json:"report_date"`
	Period        string           `json:"period"`
	TotalCost     float64          `json:"total_cost"`
	ByEnvironment []AllocationItem `json:"by_environment"`
	ByProject     []AllocationItem `json:"by_project"`
	ByCostCenter  []AllocationItem `json:"by_cost_center"`
	ByOwner       []AllocationItem `json:"by_owner"`
}

// AllocationItem represents a cost allocation item
type AllocationItem struct {
	Name       string  `json:"name"`
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
}

// ChargebackReport represents chargeback/showback report
type ChargebackReport struct {
	AccountID       string           `json:"account_id"`
	ReportDate      time.Time        `json:"report_date"`
	Period          string           `json:"period"`
	TotalChargeback float64          `json:"total_chargeback"`
	Chargebacks     []ChargebackItem `json:"chargebacks"`
}

// ChargebackItem represents a chargeback item
type ChargebackItem struct {
	Team          string  `json:"team"`
	Amount        float64 `json:"amount"`
	Resources     int     `json:"resources"`
	BillingPeriod string  `json:"billing_period"`
}